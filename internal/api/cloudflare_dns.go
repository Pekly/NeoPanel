package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultCloudflareAPIBase = "https://api.cloudflare.com/client/v4"
	settingCloudflareToken   = "subdomain_cloudflare_api_token"
	settingCloudflareZoneID  = "subdomain_cloudflare_zone_id"
)

type CloudflareClient struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func NewCloudflareClient(token string) *CloudflareClient {
	return &CloudflareClient{
		BaseURL: defaultCloudflareAPIBase,
		Token:   strings.TrimSpace(token),
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type cloudflareResponse struct {
	Success  bool            `json:"success"`
	Errors   []cfError       `json:"errors"`
	Messages []string        `json:"messages"`
	Result   json.RawMessage `json:"result"`
}

type cfError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type CloudflareZone struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type CloudflareDNSRecord struct {
	ID       string                 `json:"id,omitempty"`
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	Content  string                 `json:"content,omitempty"`
	Proxied  bool                   `json:"proxied"`
	TTL      int                    `json:"ttl"`
	Comment  string                 `json:"comment,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

type CloudflareSyncResult struct {
	ZoneID      string `json:"zone_id"`
	ZoneName    string `json:"zone_name"`
	RecordA     string `json:"record_a_id"`
	RecordSRV   string `json:"record_srv_id"`
	FQDN        string `json:"fqdn"`
	TargetHost  string `json:"target_host"`
	Port        int    `json:"port"`
	Summary     string `json:"summary"`
	SyncedAt    string `json:"synced_at"`
}

func (c *CloudflareClient) doRequest(ctx context.Context, method, endpoint string, body interface{}) (*cloudflareResponse, error) {
	if c.Token == "" {
		return nil, fmt.Errorf("Cloudflare API token is required")
	}

	reqURL := c.BaseURL + endpoint
	var bodyReader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "NeoPanel-Cloudflare-DNS/1.0")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Cloudflare request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var cfResp cloudflareResponse
	if err := json.Unmarshal(respBytes, &cfResp); err != nil {
		return nil, fmt.Errorf("parsing Cloudflare JSON response (HTTP %d): %w", resp.StatusCode, err)
	}

	if !cfResp.Success {
		var errMsgs []string
		for _, e := range cfResp.Errors {
			errMsgs = append(errMsgs, fmt.Sprintf("[%d] %s", e.Code, e.Message))
		}
		if len(errMsgs) == 0 {
			errMsgs = append(errMsgs, fmt.Sprintf("HTTP %d", resp.StatusCode))
		}
		return nil, fmt.Errorf("Cloudflare API error: %s", strings.Join(errMsgs, ", "))
	}

	return &cfResp, nil
}

func (c *CloudflareClient) VerifyToken(ctx context.Context) (string, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/user/tokens/verify", nil)
	if err != nil {
		return "", err
	}
	var res struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	_ = json.Unmarshal(resp.Result, &res)
	if res.Status != "active" && res.Status != "" {
		return res.Status, fmt.Errorf("token status is %s (expected active)", res.Status)
	}
	return res.Status, nil
}

func (c *CloudflareClient) ListZones(ctx context.Context, nameFilter string) ([]CloudflareZone, error) {
	endpoint := "/zones?status=active&per_page=50"
	if nameFilter != "" {
		endpoint += "&name=" + url.QueryEscape(nameFilter)
	}
	resp, err := c.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	var zones []CloudflareZone
	if err := json.Unmarshal(resp.Result, &zones); err != nil {
		return nil, err
	}
	return zones, nil
}

func (c *CloudflareClient) FindZoneForFQDN(ctx context.Context, fqdn string) (*CloudflareZone, error) {
	clean := strings.ToLower(strings.Trim(fqdn, "."))
	zones, err := c.ListZones(ctx, "")
	if err != nil {
		return nil, err
	}

	var bestZone *CloudflareZone
	for i := range zones {
		z := &zones[i]
		zName := strings.ToLower(z.Name)
		if clean == zName || strings.HasSuffix(clean, "."+zName) {
			if bestZone == nil || len(zName) > len(bestZone.Name) {
				bestZone = z
			}
		}
	}

	if bestZone == nil {
		return nil, fmt.Errorf("no active Cloudflare zone found matching %q among %d accessible zones", fqdn, len(zones))
	}
	return bestZone, nil
}

func (c *CloudflareClient) FindRecords(ctx context.Context, zoneID, name, recType string) ([]CloudflareDNSRecord, error) {
	endpoint := fmt.Sprintf("/zones/%s/dns_records?per_page=100", url.PathEscape(zoneID))
	if name != "" {
		endpoint += "&name=" + url.QueryEscape(name)
	}
	if recType != "" {
		endpoint += "&type=" + url.QueryEscape(recType)
	}
	resp, err := c.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	var records []CloudflareDNSRecord
	if err := json.Unmarshal(resp.Result, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func (c *CloudflareClient) CreateRecord(ctx context.Context, zoneID string, rec CloudflareDNSRecord) (*CloudflareDNSRecord, error) {
	endpoint := fmt.Sprintf("/zones/%s/dns_records", url.PathEscape(zoneID))
	resp, err := c.doRequest(ctx, http.MethodPost, endpoint, rec)
	if err != nil {
		return nil, err
	}
	var created CloudflareDNSRecord
	if err := json.Unmarshal(resp.Result, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

func (c *CloudflareClient) UpdateRecord(ctx context.Context, zoneID, recordID string, rec CloudflareDNSRecord) (*CloudflareDNSRecord, error) {
	endpoint := fmt.Sprintf("/zones/%s/dns_records/%s", url.PathEscape(zoneID), url.PathEscape(recordID))
	resp, err := c.doRequest(ctx, http.MethodPatch, endpoint, rec)
	if err != nil {
		return nil, err
	}
	var updated CloudflareDNSRecord
	if err := json.Unmarshal(resp.Result, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (c *CloudflareClient) DeleteRecord(ctx context.Context, zoneID, recordID string) error {
	endpoint := fmt.Sprintf("/zones/%s/dns_records/%s", url.PathEscape(zoneID), url.PathEscape(recordID))
	_, err := c.doRequest(ctx, http.MethodDelete, endpoint, nil)
	return err
}

func (c *CloudflareClient) SyncServerDNS(ctx context.Context, zoneID, fqdn, targetHost string, port int) (*CloudflareSyncResult, error) {
	fqdn = strings.ToLower(strings.Trim(fqdn, "."))
	cleanTarget := strings.Trim(strings.TrimSpace(targetHost), "[]")
	if fqdn == "" {
		return nil, fmt.Errorf("FQDN is required for DNS sync")
	}
	if cleanTarget == "" {
		return nil, fmt.Errorf("target host or IP is required for DNS sync")
	}
	if port <= 0 {
		port = 25565
	}

	var zone *CloudflareZone
	if zoneID != "" {
		zone = &CloudflareZone{ID: zoneID, Name: ""}
	} else {
		z, err := c.FindZoneForFQDN(ctx, fqdn)
		if err != nil {
			return nil, err
		}
		zone = z
		zoneID = z.ID
	}

	// 1. Determine address record type (A, AAAA, or CNAME)
	recType := "CNAME"
	if ip := net.ParseIP(cleanTarget); ip != nil {
		if ip.To4() != nil {
			recType = "A"
		} else {
			recType = "AAAA"
		}
	}

	// 2. Sync Address Record (A/AAAA/CNAME) with proxied=false
	addrRecs, err := c.FindRecords(ctx, zoneID, fqdn, "")
	if err != nil {
		return nil, fmt.Errorf("finding existing address records: %w", err)
	}

	addrRecord := CloudflareDNSRecord{
		Type:    recType,
		Name:    fqdn,
		Content: cleanTarget,
		Proxied: false, // Minecraft TCP traffic cannot be proxied through Cloudflare CDN
		TTL:     1,     // Auto TTL
		Comment: "Managed automatically by NeoPanel",
	}

	var finalAddrID string
	var existingAddr *CloudflareDNSRecord
	for i := range addrRecs {
		r := &addrRecs[i]
		if strings.EqualFold(r.Name, fqdn) && (r.Type == "A" || r.Type == "AAAA" || r.Type == "CNAME") {
			existingAddr = r
			break
		}
	}

	if existingAddr != nil {
		updated, err := c.UpdateRecord(ctx, zoneID, existingAddr.ID, addrRecord)
		if err != nil {
			return nil, fmt.Errorf("updating address record (%s): %w", recType, err)
		}
		finalAddrID = updated.ID
	} else {
		created, err := c.CreateRecord(ctx, zoneID, addrRecord)
		if err != nil {
			return nil, fmt.Errorf("creating address record (%s): %w", recType, err)
		}
		finalAddrID = created.ID
	}

	// 3. Sync SRV Record (_minecraft._tcp.<fqdn>)
	srvName := "_minecraft._tcp." + fqdn
	srvRecs, err := c.FindRecords(ctx, zoneID, srvName, "SRV")
	if err != nil {
		return nil, fmt.Errorf("finding existing SRV records: %w", err)
	}

	srvRecord := CloudflareDNSRecord{
		Type: "SRV",
		Name: srvName,
		Data: map[string]interface{}{
			"service":  "_minecraft",
			"proto":    "_tcp",
			"name":     fqdn,
			"priority": 0,
			"weight":   5,
			"port":     port,
			"target":   fqdn,
		},
		TTL:     1,
		Comment: "Managed automatically by NeoPanel",
	}

	var finalSRVID string
	if len(srvRecs) > 0 {
		updated, err := c.UpdateRecord(ctx, zoneID, srvRecs[0].ID, srvRecord)
		if err != nil {
			return nil, fmt.Errorf("updating SRV record: %w", err)
		}
		finalSRVID = updated.ID
	} else {
		created, err := c.CreateRecord(ctx, zoneID, srvRecord)
		if err != nil {
			return nil, fmt.Errorf("creating SRV record: %w", err)
		}
		finalSRVID = created.ID
	}

	zName := zone.Name
	if zName == "" {
		zName = zoneID
	}

	return &CloudflareSyncResult{
		ZoneID:     zoneID,
		ZoneName:   zName,
		RecordA:    finalAddrID,
		RecordSRV:  finalSRVID,
		FQDN:       fqdn,
		TargetHost: cleanTarget,
		Port:       port,
		Summary:    fmt.Sprintf("Synchronized %s (%s -> %s) and SRV record (port %d) to Cloudflare zone %s", fqdn, recType, cleanTarget, port, zName),
		SyncedAt:   time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// Handlers for SubdomainHandler integration

func (h *SubdomainHandler) getEffectiveCloudflareClient(serverID string) (*CloudflareClient, string, error) {
	token := ""
	zoneID := ""

	// Check global panel setting
	if h.db != nil {
		if t, err := h.db.GetPanelSetting(settingCloudflareToken); err == nil {
			token = strings.TrimSpace(t)
		}
		if z, err := h.db.GetPanelSetting(settingCloudflareZoneID); err == nil {
			zoneID = strings.TrimSpace(z)
		}
	}

	if token == "" {
		return nil, "", fmt.Errorf("Cloudflare API token is not configured. Add it in Admin Settings -> Subdomain / DNS.")
	}

	return NewCloudflareClient(token), zoneID, nil
}

func (h *SubdomainHandler) SyncCloudflareDNS(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	srv, err := h.db.GetServer(serverID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	var req struct {
		Token      string `json:"token,omitempty"`
		ZoneID     string `json:"zone_id,omitempty"`
		FQDN       string `json:"fqdn,omitempty"`
		TargetHost string `json:"target_host,omitempty"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	cfg, defaultTarget, _ := h.loadConfig(srv, r.Host)

	fqdn := sanitizeDomainHost(req.FQDN)
	if fqdn == "" {
		fqdn = sanitizeDomainHost(cfg.FQDN)
	}
	if fqdn == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Server subdomain FQDN is not configured"})
		return
	}

	target := strings.TrimSpace(req.TargetHost)
	if target == "" {
		target = strings.TrimSpace(cfg.TargetHost)
	}
	if target == "" {
		target = defaultTarget
	}

	var client *CloudflareClient
	zoneID := strings.TrimSpace(req.ZoneID)
	if req.Token != "" {
		client = NewCloudflareClient(req.Token)
	} else {
		c, savedZone, err := h.getEffectiveCloudflareClient(serverID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		client = c
		if zoneID == "" {
			zoneID = savedZone
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := client.SyncServerDNS(ctx, zoneID, fqdn, target, srv.Port)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	// Update saved subdomain config with sync status
	cfg.FQDN = fqdn
	cfg.TargetHost = target
	cfg.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	raw, _ := json.Marshal(cfg)
	_ = h.db.SetPanelSetting(panelSettingSubdomainPrefix+srv.ID, string(raw))

	// Record sync receipt in panel settings
	syncReceiptKey := "subdomain_cf_sync_" + srv.ID
	receiptRaw, _ := json.Marshal(result)
	_ = h.db.SetPanelSetting(syncReceiptKey, string(receiptRaw))

	logAudit(h.db, r, "subdomain_cloudflare_sync", fmt.Sprintf("Synced Cloudflare DNS for %s (A: %s, SRV: %s) on zone %s", fqdn, result.RecordA, result.RecordSRV, result.ZoneID))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": result.Summary,
		"result":  result,
	})
}

func (h *SubdomainHandler) DeleteCloudflareDNS(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	srv, err := h.db.GetServer(serverID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	syncReceiptKey := "subdomain_cf_sync_" + srv.ID
	raw, err := h.db.GetPanelSetting(syncReceiptKey)
	if err != nil || strings.TrimSpace(raw) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No recorded Cloudflare DNS synchronization found for this server"})
		return
	}

	var syncData CloudflareSyncResult
	if err := json.Unmarshal([]byte(raw), &syncData); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Corrupt Cloudflare sync receipt"})
		return
	}

	client, _, err := h.getEffectiveCloudflareClient(serverID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var deleted []string
	if syncData.RecordA != "" {
		_ = client.DeleteRecord(ctx, syncData.ZoneID, syncData.RecordA)
		deleted = append(deleted, "Address record")
	}
	if syncData.RecordSRV != "" {
		_ = client.DeleteRecord(ctx, syncData.ZoneID, syncData.RecordSRV)
		deleted = append(deleted, "SRV record")
	}

	_ = h.db.DeletePanelSetting(syncReceiptKey)
	logAudit(h.db, r, "subdomain_cloudflare_delete", fmt.Sprintf("Removed Cloudflare DNS records (%s) for %s", strings.Join(deleted, ", "), syncData.FQDN))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Removed Cloudflare DNS records for %s", syncData.FQDN),
		"deleted": deleted,
	})
}

func (h *SubdomainHandler) GetCloudflareStatus(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	if _, err := h.db.GetServer(serverID); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	tokenSet := false
	if t, err := h.db.GetPanelSetting(settingCloudflareToken); err == nil && strings.TrimSpace(t) != "" {
		tokenSet = true
	}

	var syncReceipt *CloudflareSyncResult
	if raw, err := h.db.GetPanelSetting("subdomain_cf_sync_" + serverID); err == nil && strings.TrimSpace(raw) != "" {
		var res CloudflareSyncResult
		if json.Unmarshal([]byte(raw), &res) == nil {
			syncReceipt = &res
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"configured": tokenSet,
		"sync":       syncReceipt,
	})
}

func (h *SubdomainHandler) GetAdminCloudflareConfig(w http.ResponseWriter, r *http.Request) {
	token := ""
	zoneID := ""
	if t, err := h.db.GetPanelSetting(settingCloudflareToken); err == nil {
		token = strings.TrimSpace(t)
	}
	if z, err := h.db.GetPanelSetting(settingCloudflareZoneID); err == nil {
		zoneID = strings.TrimSpace(z)
	}

	maskedToken := ""
	if len(token) > 8 {
		maskedToken = token[:4] + strings.Repeat("•", len(token)-8) + token[len(token)-4:]
	} else if token != "" {
		maskedToken = "••••••••"
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"configured":   token != "",
		"masked_token": maskedToken,
		"zone_id":      zoneID,
	})
}

func (h *SubdomainHandler) SaveAdminCloudflareConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token  string `json:"token"`
		ZoneID string `json:"zone_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	token := strings.TrimSpace(req.Token)
	zoneID := strings.TrimSpace(req.ZoneID)

	if token != "" && !strings.Contains(token, "•") {
		// Verify token with Cloudflare API
		client := NewCloudflareClient(token)
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		if _, err := client.VerifyToken(ctx); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Cloudflare API token verification failed: " + err.Error()})
			return
		}
		if err := h.db.SetPanelSetting(settingCloudflareToken, token); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	if zoneID != "" {
		_ = h.db.SetPanelSetting(settingCloudflareZoneID, zoneID)
	}

	logAudit(h.db, r, "admin_cloudflare_config", "Updated Cloudflare DNS automation credentials")
	writeJSON(w, http.StatusOK, map[string]string{"message": "Cloudflare DNS configuration saved successfully"})
}

func (h *SubdomainHandler) TestCloudflareToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	token := strings.TrimSpace(req.Token)
	if token == "" || strings.Contains(token, "•") {
		if t, err := h.db.GetPanelSetting(settingCloudflareToken); err == nil {
			token = strings.TrimSpace(t)
		}
	}

	if token == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Enter a Cloudflare API token to test"})
		return
	}

	client := NewCloudflareClient(token)
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	status, err := client.VerifyToken(ctx)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Verification failed: " + err.Error()})
		return
	}

	zones, _ := client.ListZones(ctx, "")
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"valid":       true,
		"status":      status,
		"zones_count": len(zones),
		"zones":       zones,
	})
}
