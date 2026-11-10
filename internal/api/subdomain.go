package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

const (
	panelSettingSubdomainPrefix     = "subdomain_srv_"
	panelSettingSubdomainBaseDomain = "subdomain_default_base_domain"
)

var (
	domainLabelRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)*$`)
	domainHostRe  = regexp.MustCompile(`^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$`)
)

type SubdomainHandler struct {
	db    *database.DB
	nodes *NodeHandler
}

func NewSubdomainHandler(db *database.DB, nodes *NodeHandler) *SubdomainHandler {
	return &SubdomainHandler{
		db:    db,
		nodes: nodes,
	}
}

type SubdomainConfig struct {
	ServerID   string `json:"server_id"`
	Enabled    bool   `json:"enabled"`
	Label      string `json:"label"`
	BaseDomain string `json:"base_domain"`
	FQDN       string `json:"fqdn"`
	TargetHost string `json:"target_host"`
	Port       int    `json:"port"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

type DNSRecordView struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Value    string `json:"value"`
	Priority *int   `json:"priority,omitempty"`
	Weight   *int   `json:"weight,omitempty"`
	Port     *int   `json:"port,omitempty"`
	TTL      string `json:"ttl"`
	Note     string `json:"note"`
}

func sanitizeSubdomainLabel(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.Trim(s, ".")
	return s
}

func sanitizeDomainHost(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	if idx := strings.IndexAny(s, "/?#"); idx != -1 {
		s = s[:idx]
	}
	if h, _, err := net.SplitHostPort(s); err == nil {
		s = h
	}
	s = strings.Trim(strings.TrimSpace(s), ".")
	return s
}

func (h *SubdomainHandler) loadConfig(srv *database.Server, reqHost string) (SubdomainConfig, string, string) {
	defaultTarget := ""
	if h.nodes != nil {
		defaultTarget = h.nodes.resolveLocalFQDN(reqHost)
	}
	defaultBase := ""
	if h.db != nil {
		if b, err := h.db.GetPanelSetting(panelSettingSubdomainBaseDomain); err == nil {
			defaultBase = sanitizeDomainHost(b)
		}
	}

	cfg := SubdomainConfig{
		ServerID:   srv.ID,
		Enabled:    false,
		BaseDomain: defaultBase,
		TargetHost: defaultTarget,
		Port:       srv.Port,
	}

	if h.db != nil {
		if raw, err := h.db.GetPanelSetting(panelSettingSubdomainPrefix + srv.ID); err == nil && strings.TrimSpace(raw) != "" {
			var saved SubdomainConfig
			if json.Unmarshal([]byte(raw), &saved) == nil {
				saved.ServerID = srv.ID
				if saved.Port <= 0 {
					saved.Port = srv.Port
				}
				if strings.TrimSpace(saved.TargetHost) == "" {
					saved.TargetHost = defaultTarget
				}
				if strings.TrimSpace(saved.BaseDomain) == "" {
					saved.BaseDomain = defaultBase
				}
				if strings.TrimSpace(saved.FQDN) == "" && saved.Label != "" && saved.BaseDomain != "" {
					saved.FQDN = saved.Label + "." + saved.BaseDomain
				}
				cfg = saved
			}
		}
	}

	return cfg, defaultTarget, defaultBase
}

func buildSubdomainDNSRecords(cfg SubdomainConfig, srvPort int) []DNSRecordView {
	fqdn := sanitizeDomainHost(cfg.FQDN)
	if fqdn == "" {
		lbl := sanitizeSubdomainLabel(cfg.Label)
		base := sanitizeDomainHost(cfg.BaseDomain)
		if lbl != "" && base != "" {
			fqdn = lbl + "." + base
		} else {
			fqdn = base
		}
	}
	if fqdn == "" {
		return []DNSRecordView{}
	}

	lbl := sanitizeSubdomainLabel(cfg.Label)
	if lbl == "" {
		lbl = fqdn
	}
	target := strings.TrimSpace(cfg.TargetHost)
	if target == "" {
		target = "YOUR_SERVER_IP"
	}
	port := srvPort
	if port <= 0 {
		port = 25565
	}

	recType := "CNAME"
	if ip := net.ParseIP(strings.Trim(target, "[]")); ip != nil {
		if ip.To4() != nil {
			recType = "A"
		} else {
			recType = "AAAA"
		}
	}

	prio := 0
	weight := 5
	pVal := port
	srvName := "_minecraft._tcp." + lbl
	if lbl == fqdn {
		srvName = "_minecraft._tcp." + fqdn
	}

	return []DNSRecordView{
		{
			Type:     recType,
			Name:     lbl,
			FullName: fqdn,
			Value:    target,
			TTL:      "Auto",
			Note:     "Set proxy status to DNS Only (grey cloud) on Cloudflare — proxied records break Minecraft TCP connections.",
		},
		{
			Type:     "SRV",
			Name:     srvName,
			FullName: "_minecraft._tcp." + fqdn,
			Value:    fqdn,
			Priority: &prio,
			Weight:   &weight,
			Port:     &pVal,
			TTL:      "Auto",
			Note:     fmt.Sprintf("Lets Java players connect to %s cleanly without typing :%d (service: _minecraft, protocol: _tcp).", fqdn, port),
		},
	}
}

func (h *SubdomainHandler) GetSubdomain(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	srv, err := h.db.GetServer(serverID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	cfg, defaultTarget, defaultBase := h.loadConfig(srv, r.Host)
	records := buildSubdomainDNSRecords(cfg, srv.Port)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"config":              cfg,
		"default_target":      defaultTarget,
		"default_base_domain": defaultBase,
		"server_port":         srv.Port,
		"records":             records,
	})
}

func (h *SubdomainHandler) UpdateSubdomain(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	srv, err := h.db.GetServer(serverID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	var req struct {
		Enabled    bool   `json:"enabled"`
		Label      string `json:"label"`
		BaseDomain string `json:"base_domain"`
		FQDN       string `json:"fqdn"`
		TargetHost string `json:"target_host"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body"})
		return
	}

	label := sanitizeSubdomainLabel(req.Label)
	baseDomain := sanitizeDomainHost(req.BaseDomain)
	fqdn := sanitizeDomainHost(req.FQDN)
	targetHost := strings.TrimSpace(req.TargetHost)
	if targetHost == "" && h.nodes != nil {
		targetHost = h.nodes.resolveLocalFQDN(r.Host)
	}

	if label != "" && !domainLabelRe.MatchString(label) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Subdomain label must contain only lowercase letters, numbers, and hyphens"})
		return
	}
	if baseDomain != "" && !domainHostRe.MatchString(baseDomain) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Base domain must be a valid domain name (e.g. play.example.com or example.com)"})
		return
	}

	if label != "" && baseDomain != "" {
		fqdn = label + "." + baseDomain
	} else if fqdn == "" && baseDomain != "" {
		fqdn = baseDomain
	}

	if req.Enabled && fqdn == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Provide a subdomain label and base domain (or full FQDN) to enable"})
		return
	}
	if fqdn != "" && !domainHostRe.MatchString(fqdn) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Resulting address is not a valid domain name"})
		return
	}

	cfg := SubdomainConfig{
		ServerID:   srv.ID,
		Enabled:    req.Enabled,
		Label:      label,
		BaseDomain: baseDomain,
		FQDN:       fqdn,
		TargetHost: targetHost,
		Port:       srv.Port,
		UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	raw, err := json.Marshal(cfg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to serialize subdomain settings"})
		return
	}

	if err := h.db.SetPanelSetting(panelSettingSubdomainPrefix+srv.ID, string(raw)); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save subdomain configuration"})
		return
	}
	if baseDomain != "" {
		_ = h.db.SetPanelSetting(panelSettingSubdomainBaseDomain, baseDomain)
	}

	logAudit(h.db, r, "subdomain_update", fmt.Sprintf("Updated subdomain '%s' (%v) -> %s:%d on %s", fqdn, req.Enabled, targetHost, srv.Port, auditServerLabel(h.db, srv.ID)))

	defaultTarget := ""
	if h.nodes != nil {
		defaultTarget = h.nodes.resolveLocalFQDN(r.Host)
	}
	records := buildSubdomainDNSRecords(cfg, srv.Port)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":             "Subdomain settings saved successfully",
		"config":              cfg,
		"default_target":      defaultTarget,
		"default_base_domain": baseDomain,
		"server_port":         srv.Port,
		"records":             records,
	})
}

func (h *SubdomainHandler) VerifySubdomainDNS(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	srv, err := h.db.GetServer(serverID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	cfg, defaultTarget, _ := h.loadConfig(srv, r.Host)
	var req struct {
		FQDN       string `json:"fqdn"`
		TargetHost string `json:"target_host"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	fqdn := sanitizeDomainHost(req.FQDN)
	if fqdn == "" {
		fqdn = sanitizeDomainHost(cfg.FQDN)
	}
	if fqdn == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Configure or enter a domain address first before running DNS verification"})
		return
	}

	expectedTarget := strings.TrimSpace(req.TargetHost)
	if expectedTarget == "" {
		expectedTarget = strings.TrimSpace(cfg.TargetHost)
	}
	if expectedTarget == "" {
		expectedTarget = defaultTarget
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	resolver := net.DefaultResolver

	expectedIPs := map[string]bool{}
	cleanTarget := strings.Trim(expectedTarget, "[]")
	if ip := net.ParseIP(cleanTarget); ip != nil {
		expectedIPs[ip.String()] = true
	} else if cleanTarget != "" {
		if tAddrs, err := resolver.LookupHost(ctx, cleanTarget); err == nil {
			for _, ipStr := range tAddrs {
				expectedIPs[ipStr] = true
			}
		}
	}

	aRecords, aErr := resolver.LookupHost(ctx, fqdn)
	cname, _ := resolver.LookupCNAME(ctx, fqdn)
	cname = strings.TrimSuffix(cname, ".")

	aMatched := false
	for _, addr := range aRecords {
		if expectedIPs[addr] {
			aMatched = true
			break
		}
	}
	if !aMatched && cleanTarget != "" && strings.EqualFold(cname, cleanTarget) {
		aMatched = true
	}

	type srvView struct {
		Target   string `json:"target"`
		Port     uint16 `json:"port"`
		Priority uint16 `json:"priority"`
		Weight   uint16 `json:"weight"`
	}
	_, srvAddrs, srvErr := resolver.LookupSRV(ctx, "minecraft", "tcp", fqdn)
	var srvViews []srvView
	srvMatched := false
	for _, rec := range srvAddrs {
		tgt := strings.TrimSuffix(rec.Target, ".")
		srvViews = append(srvViews, srvView{
			Target:   tgt,
			Port:     rec.Port,
			Priority: rec.Priority,
			Weight:   rec.Weight,
		})
		if int(rec.Port) == srv.Port && (strings.EqualFold(tgt, fqdn) || strings.EqualFold(tgt, cleanTarget)) {
			srvMatched = true
		}
	}
	if srvViews == nil {
		srvViews = []srvView{}
	}
	if aRecords == nil {
		aRecords = []string{}
	}

	status := "pending"
	summary := "DNS records did not resolve to this node yet. Add the A/CNAME and SRV records at your DNS provider and allow propagation."
	if aMatched && (srvMatched || srv.Port == 25565) {
		status = "verified"
		if srvMatched {
			summary = fmt.Sprintf("DNS Verified! %s and _minecraft._tcp.%s route properly to port %d.", fqdn, fqdn, srv.Port)
		} else {
			summary = fmt.Sprintf("DNS A/CNAME resolves to this node! Standard port 25565 works directly (add SRV record for cleaner routing).")
		}
	} else if aMatched || srvMatched || len(aRecords) > 0 || len(srvViews) > 0 {
		status = "partial"
		if !aMatched && len(aRecords) > 0 {
			summary = fmt.Sprintf("%s resolved to %s (expected %s). Check if Cloudflare proxy (orange cloud) is turned ON or target IP differs.", fqdn, strings.Join(aRecords, ", "), expectedTarget)
		} else if aMatched && !srvMatched && srv.Port != 25565 {
			summary = fmt.Sprintf("A/CNAME resolves to your node, but _minecraft._tcp.%s SRV record for port %d was not detected yet.", fqdn, srv.Port)
		} else {
			summary = "Partial DNS detected; verify both A/CNAME and SRV records point to this server."
		}
	}

	var warnings []string
	if aErr != nil {
		warnings = append(warnings, "Host lookup: "+aErr.Error())
	}
	if srvErr != nil && srv.Port != 25565 {
		warnings = append(warnings, "SRV lookup (_minecraft._tcp): "+srvErr.Error())
	}
	if len(expectedIPs) > 0 && !aMatched && len(aRecords) > 0 {
		warnings = append(warnings, "Resolved IPs differ from node IP/FQDN. If using Cloudflare, make sure Proxy status is DNS Only (grey cloud).")
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"fqdn":            fqdn,
		"expected_target": expectedTarget,
		"expected_port":   srv.Port,
		"status":          status,
		"summary":         summary,
		"a_records":       aRecords,
		"cname":           cname,
		"a_matched":       aMatched,
		"srv_records":     srvViews,
		"srv_matched":     srvMatched,
		"warnings":        warnings,
	})
}
