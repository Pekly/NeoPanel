package api

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/paneldomain"
	"golang.org/x/crypto/bcrypt"
)

type fakeDomainCaddy struct {
	available       bool
	calls           int
	capabilityCalls int
	onCapability    func(int)
	err             error
}

func (c *fakeDomainCaddy) Capability(context.Context) paneldomain.Capability {
	c.capabilityCalls++
	if c.onCapability != nil {
		c.onCapability(c.capabilityCalls)
	}
	return paneldomain.Capability{Available: c.available, Reason: "Configure your external proxy manually", Hosts: []string{"old.example"}, Upstream: "127.0.0.1:8080"}
}
func (c *fakeDomainCaddy) Provision(_ context.Context, host, email string) (paneldomain.ProvisionResult, error) {
	c.calls++
	return paneldomain.ProvisionResult{Changed: true, Hostname: host, Message: "HTTPS requested; certificate issuance is pending"}, c.err
}

type domainFixture struct {
	*reliabilityFixture
	h         *PanelDomainHandler
	caddy     *fakeDomainCaddy
	mode      atomic.Int32
	dialCount atomic.Int32
	server    *httptest.Server
}

func newDomainFixture(t *testing.T) *domainFixture {
	t.Helper()
	f := newReliabilityFixture(t)
	d := &domainFixture{reliabilityFixture: f, h: f.router.panelDomain, caddy: &fakeDomainCaddy{}}
	f.cfg.Server.PublicURL = "https://old.example"
	d.h.caddy = d.caddy
	d.h.checkInterval = 0
	d.h.setupInterval = 0
	d.server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
			t.Error("probe leaked authentication")
		}
		if d.mode.Load() == 1 {
			http.Redirect(w, r, "https://example.org", 302)
			return
		}
		if d.mode.Load() == 2 {
			writeJSON(w, 200, map[string]string{"proof": "wrong"})
			return
		}
		f.router.ServeHTTP(w, r)
	}))
	t.Cleanup(d.server.Close)
	d.h.roots = x509.NewCertPool()
	d.h.roots.AddCert(d.server.Certificate())
	d.h.lookup = func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("1.1.1.1")}, nil
	}
	d.h.dial = func(ctx context.Context, network, address string) (net.Conn, error) {
		d.dialCount.Add(1)
		if address != "1.1.1.1:443" {
			t.Errorf("not dialing pinned DNS address: %s", address)
		}
		return (&net.Dialer{}).DialContext(ctx, "tcp", d.server.Listener.Addr().String())
	}
	return d
}
func (d *domainFixture) check(t *testing.T) panelDomainReport {
	t.Helper()
	w := d.request("POST", "/api/admin/panel-url/check", `{"public_url":"example.com"}`, "admin")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var report panelDomainReport
	if err := json.Unmarshal(w.Body.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	return report
}
func (d *domainFixture) apply(report panelDomainReport) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]string{"public_url": report.PublicURL, "check_token": report.CheckToken})
	return d.request("POST", "/api/admin/panel-url/apply", string(body), "admin")
}
func domainCheckStatus(report panelDomainReport, id string) string {
	for _, check := range report.Checks {
		if check.ID == id {
			return check.Status
		}
	}
	return ""
}

func TestPanelDomainAdminOnly(t *testing.T) {
	d := newDomainFixture(t)
	for _, endpoint := range []string{"check", "apply", "https"} {
		for _, user := range []string{"", "owner", "subuser"} {
			w := d.request("POST", "/api/admin/panel-url/"+endpoint, `{"public_url":"example.com"}`, user)
			if w.Code != 401 && w.Code != 403 {
				t.Fatal(endpoint, user, w.Code)
			}
		}
	}
	if d.dialCount.Load() != 0 {
		t.Fatal("unauthorized probe")
	}
	if w := d.request("GET", "/api/panel-domain-proof/unknown", "", ""); w.Code != 404 {
		t.Fatal(w.Code)
	}
}
func TestPanelDomainChecksAndApplySameOrigin(t *testing.T) {
	d := newDomainFixture(t)
	report := d.check(t)
	if !report.Ready || !report.DNSReady || report.Hostname != "example.com" || report.CheckToken == "" || len(report.Certificates) != 1 {
		t.Fatalf("%+v", report)
	}
	if report.Certificates[0].LetsEncrypt {
		t.Fatal("test certificate falsely called Let's Encrypt")
	}
	if len(d.h.challenges) != 0 {
		t.Fatal("proofs retained after check")
	}
	if w := d.apply(report); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	settings, err := d.router.auth.panelURL()
	if err != nil || settings.PublicURL != "https://example.com" {
		t.Fatal(settings, err)
	}
	rp, err := d.db.PasskeyRPID()
	if err != nil || rp != "example.com" {
		t.Fatal(rp, err)
	}
	if d.dialCount.Load() != 2 {
		t.Fatal("apply did not perform a fresh verification")
	}
	if w := d.apply(report); w.Code != 409 {
		t.Fatal("consumed check reused", w.Code)
	}
}
func TestPanelDomainWrongSiteAndRedirectStayBlocked(t *testing.T) {
	for _, mode := range []int32{1, 2} {
		t.Run(string(rune('0'+mode)), func(t *testing.T) {
			d := newDomainFixture(t)
			d.mode.Store(mode)
			report := d.check(t)
			if report.Ready || domainCheckStatus(report, "panel") != "failed" {
				t.Fatal(report)
			}
			if domainCheckStatus(report, "https") != "passed" {
				t.Fatal("trusted TLS was confused with panel identity", report)
			}
			if w := d.apply(report); w.Code != 409 {
				t.Fatal(w.Code)
			}
			settings, _ := d.router.auth.panelURL()
			if settings.PublicURL != "https://old.example" {
				t.Fatal(settings)
			}
		})
	}
}
func TestPanelDomainUntrustedAndWrongHostnameCertificates(t *testing.T) {
	d := newDomainFixture(t)
	d.h.roots = x509.NewCertPool()
	report := d.check(t)
	if report.Ready || domainCheckStatus(report, "https") != "failed" {
		t.Fatal(report)
	}
	d.h.roots.AddCert(d.server.Certificate())
	w := d.request("POST", "/api/admin/panel-url/check", `{"public_url":"wrong.example"}`, "admin")
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	json.Unmarshal(w.Body.Bytes(), &report)
	if report.Ready || domainCheckStatus(report, "https") != "failed" {
		t.Fatal(report)
	}
}
func TestPanelDomainPrivateAndMixedDNSNeverDial(t *testing.T) {
	for _, addresses := range [][]string{{"127.0.0.1"}, {"10.0.0.1"}, {"169.254.169.254"}, {"::1"}, {"1.1.1.1", "192.168.0.1"}, {"100.64.0.1"}} {
		d := newDomainFixture(t)
		d.h.lookup = func(context.Context, string, string) ([]netip.Addr, error) {
			ips := []netip.Addr{}
			for _, raw := range addresses {
				ips = append(ips, netip.MustParseAddr(raw))
			}
			return ips, nil
		}
		report := d.check(t)
		if report.Ready || report.DNSReady || d.dialCount.Load() != 0 {
			t.Fatal(addresses, report)
		}
	}
}
func TestPanelDomainEveryResolvedAddressMustReachPanel(t *testing.T) {
	d := newDomainFixture(t)
	d.h.lookup = func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("1.1.1.1"), netip.MustParseAddr("8.8.8.8")}, nil
	}
	original := d.h.dial
	d.h.dial = func(ctx context.Context, network, address string) (net.Conn, error) {
		if address == "8.8.8.8:443" {
			return nil, errors.New("not reachable")
		}
		return original(ctx, network, address)
	}
	report := d.check(t)
	if report.Ready || domainCheckStatus(report, "panel") != "failed" {
		t.Fatal(report)
	}
}
func TestPanelDomainTicketBindsUserOriginExpiryAndSavedURL(t *testing.T) {
	d := newDomainFixture(t)
	report := d.check(t)
	body, _ := json.Marshal(map[string]string{"public_url": "other.example", "check_token": report.CheckToken})
	if w := d.request("POST", "/api/admin/panel-url/apply", string(body), "admin"); w.Code != 409 {
		t.Fatal(w.Code)
	}
	other, err := d.db.CreateUser("other-admin", "unused", "admin")
	if err != nil {
		t.Fatal(err)
	}
	d.tokens["other-admin"], _ = d.router.auth.GenerateToken(other)
	body, _ = json.Marshal(map[string]string{"public_url": report.PublicURL, "check_token": report.CheckToken})
	if w := d.request("POST", "/api/admin/panel-url/apply", string(body), "other-admin"); w.Code != 409 {
		t.Fatal(w.Code)
	}
	now := time.Now()
	d.h.now = func() time.Time { return now.Add(6 * time.Minute) }
	if w := d.apply(report); w.Code != 409 {
		t.Fatal("expired check accepted", w.Code)
	}
	d.h.now = time.Now
	report = d.check(t)
	if w := savePanelURL(d.reliabilityFixture, "https://changed.example"); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w := d.apply(report); w.Code != 409 {
		t.Fatal("overwrote a newer setting", w.Code)
	}
}
func TestPanelDomainApplyRechecksNetwork(t *testing.T) {
	d := newDomainFixture(t)
	report := d.check(t)
	d.mode.Store(2)
	if w := d.apply(report); w.Code != 409 || !strings.Contains(w.Body.String(), "checks changed") {
		t.Fatal(w.Code, w.Body.String())
	}
	settings, _ := d.router.auth.panelURL()
	if settings.PublicURL != "https://old.example" {
		t.Fatal(settings)
	}
}
func TestPanelDomainListsBlockingAccountsWithoutCredentialMaterial(t *testing.T) {
	d := newDomainFixture(t)
	owner := d.users["owner"]
	if _, err := d.db.AddPasskey(owner.ID, owner.PasswordHash, "old.example", "private label", []byte("secret-credential-id"), "secret-key-data"); err != nil {
		t.Fatal(err)
	}
	report := d.check(t)
	if report.Ready || len(report.PasskeyAccounts) != 1 || report.PasskeyAccounts[0].Username != "owner" {
		t.Fatal(report)
	}
	raw, _ := json.Marshal(report)
	if strings.Contains(string(raw), "secret-") || strings.Contains(string(raw), "private label") {
		t.Fatal("credential material leaked")
	}
}
func TestPanelDomainPasskeyRegistrationRaceStillBlocked(t *testing.T) {
	d := newDomainFixture(t)
	report := d.check(t)
	d.caddy.onCapability = func(call int) {
		if call == 2 {
			owner := d.users["owner"]
			if _, err := d.db.AddPasskey(owner.ID, owner.PasswordHash, "old.example", "New", []byte("racing-key"), "{}"); err != nil {
				t.Error(err)
			}
		}
	}
	if w := d.apply(report); w.Code != 409 || !strings.Contains(w.Body.String(), "passkeys") {
		t.Fatal(w.Code, w.Body.String())
	}
}
func TestPanelDomainMalformedInputAndRateLimit(t *testing.T) {
	d := newDomainFixture(t)
	for _, body := range []string{`{}`, `null`, `{"public_url":"example.com/path"}`, `{"public_url":"example.com","unknown":true}`, `{"public_url":"example.com"} {}`, strings.Repeat("x", 5000)} {
		if w := d.request("POST", "/api/admin/panel-url/check", body, "admin"); w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if d.dialCount.Load() != 0 {
		t.Fatal("invalid input dialed network")
	}
	d.h.checkInterval = time.Minute
	d.check(t)
	if w := d.request("POST", "/api/admin/panel-url/check", `{"public_url":"example.com"}`, "admin"); w.Code != 429 {
		t.Fatal(w.Code)
	}
}
func TestPanelDomainSetupRequiresPasswordConfirmationAndMatchingDNS(t *testing.T) {
	d := newDomainFixture(t)
	d.caddy.available = true
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	admin, err := d.db.CreateUser("certificate-admin", string(hash), "admin")
	if err != nil {
		t.Fatal(err)
	}
	d.users["admin"] = admin
	d.tokens["admin"], _ = d.router.auth.GenerateToken(admin)
	report := d.check(t)
	request := func(password string, confirm bool) *httptest.ResponseRecorder {
		body, _ := json.Marshal(map[string]any{"public_url": report.PublicURL, "check_token": report.CheckToken, "email": "admin@example.com", "password": password, "confirm": confirm})
		return d.request("POST", "/api/admin/panel-url/https", string(body), "admin")
	}
	if w := request("correct-password", false); w.Code != 400 {
		t.Fatal(w.Code)
	}
	if w := request("wrong", true); w.Code != 403 {
		t.Fatal(w.Code)
	}
	if d.caddy.calls != 0 {
		t.Fatal("unauthorized provisioning")
	}
	lookup := d.h.lookup
	d.h.lookup = func(ctx context.Context, network, host string) ([]netip.Addr, error) {
		if host == "old.example" {
			return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
		}
		return lookup(ctx, network, host)
	}
	if w := request("correct-password", true); w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	d.h.lookup = lookup
	if w := request("correct-password", true); w.Code != 200 || !strings.Contains(w.Body.String(), "needs_verification") {
		t.Fatal(w.Code, w.Body.String())
	}
	if d.caddy.calls != 1 {
		t.Fatal(d.caddy.calls)
	}
	settings, _ := d.router.auth.panelURL()
	if settings.PublicURL != "https://old.example" {
		t.Fatal("setup changed saved address")
	}
	if w := d.apply(report); w.Code != 409 {
		t.Fatal("setup left old verification usable", w.Code)
	}
}
