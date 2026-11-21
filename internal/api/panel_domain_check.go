package api

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/paneldomain"
)

type panelDomainCheck struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Status string `json:"status"`
	Detail string `json:"detail"`
	Action string `json:"action,omitempty"`
}
type panelDomainCertificate struct {
	Address     string    `json:"address"`
	Issuer      string    `json:"issuer"`
	ExpiresAt   time.Time `json:"expires_at"`
	LetsEncrypt bool      `json:"lets_encrypt"`
}
type panelDomainReport struct {
	PublicURL       string                               `json:"public_url"`
	Hostname        string                               `json:"hostname"`
	CurrentURL      string                               `json:"current_url"`
	CheckedAt       time.Time                            `json:"checked_at"`
	ExpiresAt       time.Time                            `json:"expires_at"`
	CheckToken      string                               `json:"check_token,omitempty"`
	Ready           bool                                 `json:"ready"`
	DNSReady        bool                                 `json:"dns_ready"`
	Addresses       []string                             `json:"addresses"`
	Checks          []panelDomainCheck                   `json:"checks"`
	Certificates    []panelDomainCertificate             `json:"certificates"`
	PasskeyAccounts []database.PanelDomainPasskeyAccount `json:"passkey_accounts"`
	HTTPSSetup      paneldomain.Capability               `json:"https_setup"`
}

func normalizeDomainOrigin(raw string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", errors.New("Enter the panel hostname first")
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	return canonicalPasskeyOrigin(raw)
}
func domainRandom() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
func domainLocal(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

// Resolve once, reject mixed/private results, then dial exactly these addresses.
// Localhost is an explicit development exception, not a private-network scanner.
func (h *PanelDomainHandler) resolve(ctx context.Context, host string) ([]netip.Addr, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	ips, err := h.lookup(ctx, "ip", host)
	if err != nil || len(ips) == 0 {
		return nil, fmt.Errorf("DNS did not resolve %s. Check its A/AAAA records and allow time for DNS changes to propagate", host)
	}
	if len(ips) > 16 {
		return nil, errors.New("DNS returned too many addresses to verify safely (maximum 16)")
	}
	found := make(map[netip.Addr]bool)
	out := make([]netip.Addr, 0, len(ips))
	for _, ip := range ips {
		ip = ip.Unmap()
		if (domainLocal(host) && !ip.IsLoopback()) || (!domainLocal(host) && !publicEmailLogoIP(ip)) {
			return nil, errors.New("This hostname resolves to a private, reserved, or local address. Public domain checks require public DNS; only explicit localhost development is supported")
		}
		if !found[ip] {
			found[ip] = true
			out = append(out, ip)
		}
	}
	return out, nil
}

type domainProbeResult struct {
	certificate *panelDomainCertificate
	tlsOK       bool
	panelOK     bool
	detail      string
}

func (h *PanelDomainHandler) probe(ctx context.Context, u *url.URL, ip netip.Addr, nonce, proof string) domainProbeResult {
	port := u.Port()
	if port == "" {
		port = "443"
		if u.Scheme == "http" {
			port = "80"
		}
	}
	transport := &http.Transport{
		Proxy: nil, DisableKeepAlives: true, MaxResponseHeaderBytes: 16 << 10, ResponseHeaderTimeout: 5 * time.Second, TLSHandshakeTimeout: 5 * time.Second,
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12, ServerName: u.Hostname(), RootCAs: h.roots},
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			return h.dial(ctx, "tcp", net.JoinHostPort(ip.String(), port))
		},
	}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 8 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	target := *u
	target.Path = "/api/panel-domain-proof/" + nonce
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return domainProbeResult{detail: "Could not build the panel check"}
	}
	req.Header.Set("Accept", "application/json")
	// No Authorization, cookies, environment proxy, or client-supplied headers.
	response, err := client.Do(req)
	if err != nil {
		var certErr *tls.CertificateVerificationError
		if errors.As(err, &certErr) {
			return domainProbeResult{detail: "The HTTPS certificate is untrusted, expired, or does not cover this hostname. Fix the certificate or use the supported Let's Encrypt setup"}
		}
		return domainProbeResult{detail: "Could not reach the panel at this address. Check DNS, the proxy's hostname, firewall, and HTTPS listener"}
	}
	defer response.Body.Close()
	result := domainProbeResult{tlsOK: u.Scheme == "http"}
	if u.Scheme == "https" && response.TLS != nil && len(response.TLS.VerifiedChains) > 0 && len(response.TLS.PeerCertificates) > 0 {
		leaf := response.TLS.PeerCertificates[0]
		issuer := strings.Join(leaf.Issuer.Organization, ", ")
		if issuer == "" {
			issuer = leaf.Issuer.CommonName
		}
		le := false
		for _, org := range leaf.Issuer.Organization {
			if strings.EqualFold(org, "Let's Encrypt") {
				le = true
			}
		}
		result.tlsOK = true
		result.certificate = &panelDomainCertificate{ip.String(), issuer, leaf.NotAfter, le}
	}
	if response.StatusCode != http.StatusOK {
		result.detail = fmt.Sprintf("The address returned HTTP %d instead of this panel's verification response. Route this hostname directly to this NeoPanel instance without a redirect or proxy login page", response.StatusCode)
		return result
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 4097))
	if err != nil || len(body) > 4096 {
		result.detail = "The address did not return a valid panel verification response"
		return result
	}
	var payload struct {
		Proof string `json:"proof"`
	}
	if json.Unmarshal(body, &payload) != nil || subtle.ConstantTimeCompare([]byte(payload.Proof), []byte(proof)) != 1 {
		result.detail = "This address does not reach this NeoPanel instance. Check the proxy's upstream and any caching or access-control rules"
		return result
	}
	result.panelOK = true
	return result
}

func (h *PanelDomainHandler) inspect(ctx context.Context, origin, host string) (panelDomainReport, error) {
	now := h.now()
	report := panelDomainReport{PublicURL: origin, Hostname: host, CheckedAt: now, ExpiresAt: now.Add(5 * time.Minute), Addresses: []string{}, Checks: []panelDomainCheck{}, Certificates: []panelDomainCertificate{}}
	current, err := h.auth.configuredPanelURL()
	if err != nil {
		return report, err
	}
	report.CurrentURL = current.PublicURL
	conflicts, err := h.auth.db.PanelDomainPasskeyConflicts(host)
	if err != nil {
		return report, err
	}
	report.PasskeyAccounts = conflicts
	report.HTTPSSetup = h.caddy.Capability(ctx)
	if err := ctx.Err(); err != nil {
		return report, err
	}
	u, _ := url.Parse(origin)
	ips, resolveErr := h.resolve(ctx, host)
	dns := panelDomainCheck{"dns", "DNS", "passed", "DNS records resolved. All returned addresses are checked below.", ""}
	https := panelDomainCheck{"https", "HTTPS certificate", "waiting", "Waiting for DNS and the HTTPS connection.", ""}
	panel := panelDomainCheck{"panel", "This panel", "waiting", "Waiting for the same-panel verification response.", ""}
	if resolveErr != nil {
		dns.Status = "failed"
		dns.Detail = resolveErr.Error()
		dns.Action = "Update the domain's DNS records, then check again. No settings have changed."
	} else {
		report.DNSReady = true
		for _, ip := range ips {
			report.Addresses = append(report.Addresses, ip.String())
		}
		nonce, err := domainRandom()
		if err != nil {
			return report, err
		}
		proof, err := domainRandom()
		if err != nil {
			return report, err
		}
		h.mu.Lock()
		h.challenges[nonce] = domainChallenge{proof, h.now().Add(25 * time.Second)}
		h.mu.Unlock()
		defer func() { h.mu.Lock(); delete(h.challenges, nonce); h.mu.Unlock() }()
		results := make([]domainProbeResult, len(ips))
		var wg sync.WaitGroup
		for i, ip := range ips {
			wg.Add(1)
			go func(i int, ip netip.Addr) { defer wg.Done(); results[i] = h.probe(ctx, u, ip, nonce, proof) }(i, ip)
		}
		wg.Wait()
		https.Status = "passed"
		https.Detail = "HTTPS is trusted and the certificate covers this hostname on every resolved address."
		panel.Status = "passed"
		panel.Detail = "Every resolved address returned this panel's short-lived verification proof."
		for i, result := range results {
			if result.certificate != nil {
				report.Certificates = append(report.Certificates, *result.certificate)
			}
			if !result.tlsOK {
				https.Status = "failed"
				https.Detail = ips[i].String() + ": " + result.detail
				https.Action = "Fix the certificate/proxy or set up Let's Encrypt below, then check again."
			}
			if !result.panelOK {
				panel.Status = "failed"
				panel.Detail = ips[i].String() + ": " + result.detail
				panel.Action = "Keep the old address working; fix this hostname's proxy route before applying it."
			}
		}
		if u.Scheme == "http" {
			https.Status = "warning"
			https.Detail = "HTTP localhost development only. There is no TLS certificate; public panel addresses must use HTTPS."
		} else if https.Status == "passed" {
			for _, cert := range report.Certificates {
				if cert.ExpiresAt.Before(now.Add(14 * 24 * time.Hour)) {
					https.Status = "warning"
					https.Detail = "The certificate is valid but expires within 14 days. Check automatic renewal with the TLS provider."
					break
				}
			}
		}
	}
	keys := panelDomainCheck{"passkeys", "Passkey hostname", "passed", "Panel links and passkeys will use the same verified hostname: " + host, ""}
	if len(conflicts) > 0 {
		keys.Status = "blocked"
		keys.Detail = "Existing passkeys belong to a different hostname. Their accounts are listed below; no passkeys have been removed."
		keys.Action = "Confirm password sign-in, have the listed users remove old passkeys in Account → Security, then recheck and register new passkeys after the move."
	}
	report.Checks = []panelDomainCheck{dns, https, panel, keys}
	report.Ready = dns.Status == "passed" && (https.Status == "passed" || https.Status == "warning") && panel.Status == "passed" && keys.Status == "passed"
	return report, nil
}
