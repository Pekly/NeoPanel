package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
	"github.com/gorilla/websocket"
)

type NodeHandler struct {
	cfg        *config.Config
	db         *database.DB
	auth       *AuthHandler
	manager    *runner.ServerManager
	httpClient *http.Client
}

func NewNodeHandler(cfg *config.Config, db *database.DB, auth *AuthHandler, manager *runner.ServerManager) *NodeHandler {
	return &NodeHandler{
		cfg:     cfg,
		db:      db,
		auth:    auth,
		manager: manager,
		httpClient: &http.Client{
			Timeout: 55 * time.Second,
		},
	}
}

type NodeStatusView struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	BaseURL         string      `json:"base_url"`
	FQDN            string      `json:"fqdn"`
	Location        string      `json:"location"`
	IsLocal         bool        `json:"is_local"`
	Online          bool        `json:"online"`
	Error           string      `json:"error,omitempty"`
	LatencyMS       int64       `json:"latency_ms"`
	ServerCount     int         `json:"server_count"`
	System          *SystemInfo `json:"system,omitempty"`
	HasToken        bool        `json:"has_token"`
	TokenPrefix     string      `json:"token_prefix"`
	IsFree          bool        `json:"is_free"`
	IdleTimeoutMin  int         `json:"idle_timeout_min"`
	AllowUserRAM    bool        `json:"allow_user_ram"`
	AllowUserPorts  bool        `json:"allow_user_ports"`
	AllowUserJava   bool        `json:"allow_user_java"`
	IsolationEngine string      `json:"isolation_engine"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

func normalizeNodeBaseURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("base_url is required")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid base_url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("base_url must begin with http:// or https://")
	}
	if strings.TrimSpace(u.Host) == "" {
		return "", fmt.Errorf("base_url must include a host")
	}
	u.RawQuery = ""
	u.Fragment = ""
	cleaned := strings.TrimRight(u.String(), "/")
	return cleaned, nil
}

func maskNodeToken(tok string) (bool, string) {
	tok = strings.TrimSpace(tok)
	if tok == "" {
		return false, ""
	}
	if len(tok) <= 10 {
		return true, "****"
	}
	return true, tok[:8] + "..."
}

func (h *NodeHandler) resolveLocalFQDN(reqHost string) string {
	if h.db != nil {
		if raw, err := h.db.GetPanelSetting(panelSettingDaemonKey); err == nil && strings.TrimSpace(raw) != "" {
			var d DaemonSettingsConfig
			if json.Unmarshal([]byte(raw), &d) == nil && strings.TrimSpace(d.FQDN) != "" {
				return strings.TrimSpace(d.FQDN)
			}
		}
	}
	host := strings.TrimSpace(reqHost)
	if host != "" {
		if hPart, _, err := net.SplitHostPort(host); err == nil {
			host = hPart
		} else {
			host = strings.Trim(host, "[]")
		}
	}
	if host == "" {
		return "127.0.0.1"
	}
	return host
}

func (h *NodeHandler) buildLocalNodeView(reqHost string) NodeStatusView {
	sys := collectSystemInfo()
	srvs, _ := h.db.ListServers()
	localNode, _ := h.db.GetNode("local")
	isFree := false
	idleMin := 0
	allowRAM := true
	allowPorts := true
	allowJava := true
	isoEngine := "Bubblewrap + Systemd Slices"
	if localNode != nil {
		isFree = localNode.IsFree
		idleMin = localNode.IdleTimeoutMin
		allowRAM = localNode.AllowUserRAM
		allowPorts = localNode.AllowUserPorts
		allowJava = localNode.AllowUserJava
		if localNode.IsolationEngine != "" {
			isoEngine = localNode.IsolationEngine
		}
	}
	return NodeStatusView{
		ID:              "local",
		Name:            "Local Master Node",
		BaseURL:         "",
		FQDN:            h.resolveLocalFQDN(reqHost),
		Location:        "Primary Host",
		IsLocal:         true,
		Online:          true,
		LatencyMS:       0,
		ServerCount:     len(srvs),
		System:          &sys,
		HasToken:        true,
		TokenPrefix:     "local",
		IsFree:          isFree,
		IdleTimeoutMin:  idleMin,
		AllowUserRAM:    allowRAM,
		AllowUserPorts:  allowPorts,
		AllowUserJava:   allowJava,
		IsolationEngine: isoEngine,
	}
}

func (h *NodeHandler) probeRemote(ctx context.Context, baseURL, apiToken string) (bool, *SystemInfo, int, int64, string) {
	probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 5 * time.Second}
	started := time.Now()

	sysReq, err := http.NewRequestWithContext(probeCtx, http.MethodGet, baseURL+"/api/system", nil)
	if err != nil {
		return false, nil, 0, 0, err.Error()
	}
	sysReq.Header.Set("Authorization", "Bearer "+apiToken)
	sysReq.Header.Set("Accept", "application/json")

	sysResp, err := client.Do(sysReq)
	if err != nil {
		return false, nil, 0, 0, err.Error()
	}
	defer sysResp.Body.Close()

	latency := time.Since(started).Milliseconds()
	if sysResp.StatusCode == http.StatusUnauthorized || sysResp.StatusCode == http.StatusForbidden {
		return false, nil, 0, latency, fmt.Sprintf("Authentication failed (HTTP %d) — check node API token", sysResp.StatusCode)
	}
	if sysResp.StatusCode != http.StatusOK {
		return false, nil, 0, latency, fmt.Sprintf("Node returned HTTP %d", sysResp.StatusCode)
	}

	var sysInfo SystemInfo
	if err := json.NewDecoder(sysResp.Body).Decode(&sysInfo); err != nil {
		return false, nil, 0, latency, "Invalid /api/system JSON response"
	}

	srvCount := 0
	srvReq, err := http.NewRequestWithContext(probeCtx, http.MethodGet, baseURL+"/api/servers", nil)
	if err == nil {
		srvReq.Header.Set("Authorization", "Bearer "+apiToken)
		srvReq.Header.Set("Accept", "application/json")
		if srvResp, err := client.Do(srvReq); err == nil {
			defer srvResp.Body.Close()
			if srvResp.StatusCode == http.StatusOK {
				var arr []json.RawMessage
				if json.NewDecoder(srvResp.Body).Decode(&arr) == nil {
					srvCount = len(arr)
				}
			}
		}
	}

	go func() {
		_ = h.SyncAuthToNode(context.Background(), baseURL, apiToken)
	}()

	return true, &sysInfo, srvCount, latency, ""
}

func (h *NodeHandler) ListNodes(w http.ResponseWriter, r *http.Request) {
	localView := h.buildLocalNodeView(r.Host)

	remoteNodes, err := h.db.ListNodes()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	views := make([]NodeStatusView, 1+len(remoteNodes))
	views[0] = localView

	var wg sync.WaitGroup
	for i, n := range remoteNodes {
		hasTok, tokPrefix := maskNodeToken(n.APIToken)
		views[i+1] = NodeStatusView{
			ID:              n.ID,
			Name:            n.Name,
			BaseURL:         n.BaseURL,
			FQDN:            n.FQDN,
			Location:        n.Location,
			IsLocal:         false,
			HasToken:        hasTok,
			TokenPrefix:     tokPrefix,
			IsFree:          n.IsFree,
			IdleTimeoutMin:  n.IdleTimeoutMin,
			AllowUserRAM:    n.AllowUserRAM,
			AllowUserPorts:  n.AllowUserPorts,
			AllowUserJava:   n.AllowUserJava,
			IsolationEngine: n.IsolationEngine,
			CreatedAt:       n.CreatedAt,
			UpdatedAt:       n.UpdatedAt,
		}
		wg.Add(1)
		go func(idx int, node *database.ClusterNode) {
			defer wg.Done()
			online, sys, cnt, latency, probeErr := h.probeRemote(r.Context(), node.BaseURL, node.APIToken)
			views[idx].Online = online
			views[idx].System = sys
			views[idx].ServerCount = cnt
			views[idx].LatencyMS = latency
			views[idx].Error = probeErr
		}(i+1, n)
	}
	wg.Wait()

	writeJSON(w, http.StatusOK, views)
}

func (h *NodeHandler) ListClusterServers(w http.ResponseWriter, r *http.Request) {
	localFQDN := h.resolveLocalFQDN(r.Host)

	localServers, err := h.db.ListServers()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	type ClusterServerItem struct {
		*database.Server
		Stats    database.ServerStats `json:"stats"`
		MOTD     string               `json:"motd,omitempty"`
		NodeID   string               `json:"node_id"`
		NodeName string               `json:"node_name"`
		NodeFQDN string               `json:"node_fqdn"`
	}

	allServers := make([]interface{}, 0, len(localServers)+8)
	for _, s := range localServers {
		allServers = append(allServers, ClusterServerItem{
			Server:   s,
			Stats:    h.manager.GetStats(s),
			MOTD:     readServerMOTD(h.manager.GetWorkingDir(s.ID)),
			NodeID:   "local",
			NodeName: "Local Master Node",
			NodeFQDN: localFQDN,
		})
	}

	remoteNodes, _ := h.db.ListNodes()
	nodeErrors := map[string]string{}

	if len(remoteNodes) > 0 {
		var mu sync.Mutex
		var wg sync.WaitGroup
		client := &http.Client{Timeout: 5 * time.Second}

		for _, n := range remoteNodes {
			wg.Add(1)
			go func(node *database.ClusterNode) {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
				defer cancel()

				req, err := http.NewRequestWithContext(ctx, http.MethodGet, node.BaseURL+"/api/servers", nil)
				if err != nil {
					mu.Lock()
					nodeErrors[node.ID] = err.Error()
					mu.Unlock()
					return
				}
				req.Header.Set("Authorization", "Bearer "+node.APIToken)
				req.Header.Set("Accept", "application/json")

				resp, err := client.Do(req)
				if err != nil {
					mu.Lock()
					nodeErrors[node.ID] = err.Error()
					mu.Unlock()
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					mu.Lock()
					nodeErrors[node.ID] = fmt.Sprintf("HTTP %d", resp.StatusCode)
					mu.Unlock()
					return
				}

				var rawList []map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&rawList); err != nil {
					mu.Lock()
					nodeErrors[node.ID] = "Invalid JSON response"
					mu.Unlock()
					return
				}

				nodeFQDN := strings.TrimSpace(node.FQDN)
				if nodeFQDN == "" {
					if u, err := url.Parse(node.BaseURL); err == nil {
						nodeFQDN = u.Hostname()
					}
				}

				mu.Lock()
				for _, item := range rawList {
					item["node_id"] = node.ID
					item["node_name"] = node.Name
					item["node_fqdn"] = nodeFQDN
					allServers = append(allServers, item)
				}
				mu.Unlock()
			}(n)
		}
		wg.Wait()
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"servers":     allServers,
		"node_errors": nodeErrors,
	})
}

func (h *NodeHandler) TestNodePayload(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil || claims.Role != "admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Admin privileges are required"})
		return
	}

	var req struct {
		ID       string `json:"id"`
		BaseURL  string `json:"base_url"`
		APIToken string `json:"api_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body"})
		return
	}

	baseURL, err := normalizeNodeBaseURL(req.BaseURL)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	apiToken := strings.TrimSpace(req.APIToken)
	if apiToken == "" && req.ID != "" && req.ID != "local" {
		if existing, err := h.db.GetNode(req.ID); err == nil && existing != nil {
			apiToken = existing.APIToken
		}
	}
	if apiToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "API token (neo_...) is required"})
		return
	}

	online, sys, cnt, latency, probeErr := h.probeRemote(r.Context(), baseURL, apiToken)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"online":       online,
		"latency_ms":   latency,
		"server_count": cnt,
		"system":       sys,
		"error":        probeErr,
	})
}

func (h *NodeHandler) TestNodeByID(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil || claims.Role != "admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Admin privileges are required"})
		return
	}

	nodeID := getURLParam(r, "id")
	if nodeID == "local" {
		lv := h.buildLocalNodeView(r.Host)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"online":       true,
			"latency_ms":   0,
			"server_count": lv.ServerCount,
			"system":       lv.System,
		})
		return
	}

	node, err := h.db.GetNode(nodeID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Node not found"})
		return
	}

	online, sys, cnt, latency, probeErr := h.probeRemote(r.Context(), node.BaseURL, node.APIToken)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"online":       online,
		"latency_ms":   latency,
		"server_count": cnt,
		"system":       sys,
		"error":        probeErr,
	})
}

func (h *NodeHandler) CreateNode(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil || claims.Role != "admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Admin privileges are required"})
		return
	}

	var req struct {
		Name            string `json:"name"`
		BaseURL         string `json:"base_url"`
		APIToken        string `json:"api_token"`
		FQDN            string `json:"fqdn"`
		Location        string `json:"location"`
		Upsert          bool   `json:"upsert"`
		IsFree          bool   `json:"is_free"`
		IdleTimeoutMin  int    `json:"idle_timeout_min"`
		AllowUserRAM    *bool  `json:"allow_user_ram"`
		AllowUserPorts  *bool  `json:"allow_user_ports"`
		AllowUserJava   *bool  `json:"allow_user_java"`
		IsolationEngine string `json:"isolation_engine"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Node name is required"})
		return
	}

	baseURL, err := normalizeNodeBaseURL(req.BaseURL)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	apiToken := strings.TrimSpace(req.APIToken)
	if apiToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "API token (neo_...) from the remote node is required"})
		return
	}

	fqdn := strings.TrimSpace(req.FQDN)
	if fqdn == "" {
		if u, err := url.Parse(baseURL); err == nil {
			fqdn = u.Hostname()
		}
	}

	if req.Upsert {
		if existingList, err := h.db.ListNodes(); err == nil {
			for _, ex := range existingList {
				if strings.EqualFold(strings.TrimRight(ex.BaseURL, "/"), baseURL) || strings.EqualFold(ex.Name, req.Name) {
					ex.Name = req.Name
					ex.BaseURL = baseURL
					ex.APIToken = apiToken
					ex.FQDN = fqdn
					ex.Location = strings.TrimSpace(req.Location)
					if err := h.db.UpdateNode(ex); err != nil {
						writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update existing node: " + err.Error()})
						return
					}
					online, sys, cnt, latency, probeErr := h.probeRemote(r.Context(), ex.BaseURL, ex.APIToken)
					hasTok, tokPrefix := maskNodeToken(ex.APIToken)
					logAudit(h.db, r, "node_update", fmt.Sprintf("Re-paired cluster node '%s' (%s) at %s", ex.Name, ex.ID, ex.BaseURL))
					writeJSON(w, http.StatusOK, NodeStatusView{
						ID:          ex.ID,
						Name:        ex.Name,
						BaseURL:     ex.BaseURL,
						FQDN:        ex.FQDN,
						Location:    ex.Location,
						IsLocal:     false,
						Online:      online,
						Error:       probeErr,
						LatencyMS:   latency,
						ServerCount: cnt,
						System:      sys,
						HasToken:    hasTok,
						TokenPrefix: tokPrefix,
						CreatedAt:   ex.CreatedAt,
						UpdatedAt:   ex.UpdatedAt,
					})
					return
				}
			}
		}
	}

	idBytes := make([]byte, 3)
	_, _ = rand.Read(idBytes)
	nodeID := "node-" + hex.EncodeToString(idBytes)

	allowRAM := true
	if req.AllowUserRAM != nil {
		allowRAM = *req.AllowUserRAM
	}
	allowPorts := true
	if req.AllowUserPorts != nil {
		allowPorts = *req.AllowUserPorts
	}
	allowJava := true
	if req.AllowUserJava != nil {
		allowJava = *req.AllowUserJava
	}
	isoEngine := "Bubblewrap + Systemd Slices"
	if req.IsolationEngine != "" {
		isoEngine = req.IsolationEngine
	}

	node := &database.ClusterNode{
		ID:              nodeID,
		Name:            req.Name,
		BaseURL:         baseURL,
		APIToken:        apiToken,
		FQDN:            fqdn,
		Location:        strings.TrimSpace(req.Location),
		IsFree:          req.IsFree,
		IdleTimeoutMin:  req.IdleTimeoutMin,
		AllowUserRAM:    allowRAM,
		AllowUserPorts:  allowPorts,
		AllowUserJava:   allowJava,
		IsolationEngine: isoEngine,
	}

	if err := h.db.CreateNode(node); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save node: " + err.Error()})
		return
	}

	online, sys, cnt, latency, probeErr := h.probeRemote(r.Context(), node.BaseURL, node.APIToken)
	hasTok, tokPrefix := maskNodeToken(node.APIToken)

	logAudit(h.db, r, "node_add", fmt.Sprintf("Connected cluster node '%s' (%s) at %s", node.Name, node.ID, node.BaseURL))

	writeJSON(w, http.StatusCreated, NodeStatusView{
		ID:              node.ID,
		Name:            node.Name,
		BaseURL:         node.BaseURL,
		FQDN:            node.FQDN,
		Location:        node.Location,
		IsLocal:         false,
		Online:          online,
		Error:           probeErr,
		LatencyMS:       latency,
		ServerCount:     cnt,
		System:          sys,
		HasToken:        hasTok,
		TokenPrefix:     tokPrefix,
		IsFree:          node.IsFree,
		IdleTimeoutMin:  node.IdleTimeoutMin,
		AllowUserRAM:    node.AllowUserRAM,
		AllowUserPorts:  node.AllowUserPorts,
		AllowUserJava:   node.AllowUserJava,
		IsolationEngine: node.IsolationEngine,
		CreatedAt:       node.CreatedAt,
		UpdatedAt:       node.UpdatedAt,
	})
}

func (h *NodeHandler) UpdateNode(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil || claims.Role != "admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Admin privileges are required"})
		return
	}

	nodeID := getURLParam(r, "id")
	if nodeID == "local" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Local master node settings are managed in Panel & Daemon Look"})
		return
	}

	existing, err := h.db.GetNode(nodeID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Node not found"})
		return
	}

	var req struct {
		Name            string `json:"name"`
		BaseURL         string `json:"base_url"`
		APIToken        string `json:"api_token"`
		FQDN            string `json:"fqdn"`
		Location        string `json:"location"`
		IsFree          bool   `json:"is_free"`
		IdleTimeoutMin  int    `json:"idle_timeout_min"`
		AllowUserRAM    *bool  `json:"allow_user_ram"`
		AllowUserPorts  *bool  `json:"allow_user_ports"`
		AllowUserJava   *bool  `json:"allow_user_java"`
		IsolationEngine string `json:"isolation_engine"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body"})
		return
	}

	if name := strings.TrimSpace(req.Name); name != "" {
		existing.Name = name
	}
	if strings.TrimSpace(req.BaseURL) != "" {
		baseURL, err := normalizeNodeBaseURL(req.BaseURL)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		existing.BaseURL = baseURL
	}
	if tok := strings.TrimSpace(req.APIToken); tok != "" {
		existing.APIToken = tok
	}
	existing.FQDN = strings.TrimSpace(req.FQDN)
	if existing.FQDN == "" {
		if u, err := url.Parse(existing.BaseURL); err == nil {
			existing.FQDN = u.Hostname()
		}
	}
	existing.Location = strings.TrimSpace(req.Location)
	existing.IsFree = req.IsFree
	existing.IdleTimeoutMin = req.IdleTimeoutMin
	if req.AllowUserRAM != nil {
		existing.AllowUserRAM = *req.AllowUserRAM
	}
	if req.AllowUserPorts != nil {
		existing.AllowUserPorts = *req.AllowUserPorts
	}
	if req.AllowUserJava != nil {
		existing.AllowUserJava = *req.AllowUserJava
	}
	if req.IsolationEngine != "" {
		existing.IsolationEngine = req.IsolationEngine
	}

	if err := h.db.UpdateNode(existing); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update node: " + err.Error()})
		return
	}

	online, sys, cnt, latency, probeErr := h.probeRemote(r.Context(), existing.BaseURL, existing.APIToken)
	hasTok, tokPrefix := maskNodeToken(existing.APIToken)

	logAudit(h.db, r, "node_update", fmt.Sprintf("Updated cluster node '%s' (%s)", existing.Name, existing.ID))

	writeJSON(w, http.StatusOK, NodeStatusView{
		ID:              existing.ID,
		Name:            existing.Name,
		BaseURL:         existing.BaseURL,
		FQDN:            existing.FQDN,
		Location:        existing.Location,
		IsLocal:         false,
		Online:          online,
		Error:           probeErr,
		LatencyMS:       latency,
		ServerCount:     cnt,
		System:          sys,
		HasToken:        hasTok,
		TokenPrefix:     tokPrefix,
		IsFree:          existing.IsFree,
		IdleTimeoutMin:  existing.IdleTimeoutMin,
		AllowUserRAM:    existing.AllowUserRAM,
		AllowUserPorts:  existing.AllowUserPorts,
		AllowUserJava:   existing.AllowUserJava,
		IsolationEngine: existing.IsolationEngine,
		CreatedAt:       existing.CreatedAt,
		UpdatedAt:       existing.UpdatedAt,
	})
}

func (h *NodeHandler) DeleteNode(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil || claims.Role != "admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Admin privileges are required"})
		return
	}

	nodeID := getURLParam(r, "id")
	if nodeID == "local" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Cannot remove the local master node"})
		return
	}

	existing, err := h.db.GetNode(nodeID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Node not found"})
		return
	}

	if err := h.db.DeleteNode(nodeID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete node"})
		return
	}

	logAudit(h.db, r, "node_delete", fmt.Sprintf("Disconnected cluster node '%s' (%s)", existing.Name, existing.ID))

	writeJSON(w, http.StatusOK, map[string]string{"message": "Node removed from cluster"})
}

func (h *NodeHandler) Proxy(w http.ResponseWriter, r *http.Request) {
	nodeID := getURLParam(r, "id")
	if nodeID == "" || nodeID == "local" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid target node ID for proxy"})
		return
	}

	node, err := h.db.GetNode(nodeID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Target node not found"})
		return
	}

	rest := getURLParam(r, "rest")
	cleanRest := strings.TrimPrefix(path.Clean("/"+rest), "/")
	if !strings.HasPrefix(cleanRest, "api/") {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Proxied path must stay within /api/"})
		return
	}
	if cleanRest == "api/nodes" ||
		strings.HasPrefix(cleanRest, "api/nodes/") ||
		strings.HasPrefix(cleanRest, "api/cluster/") {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Recursive node proxying is not allowed"})
		return
	}

	targetURL := node.BaseURL + "/" + cleanRest
	if r.URL.RawQuery != "" {
		// Remove master session query token if present when forwarding to remote node via Bearer header
		q := r.URL.Query()
		q.Del("token")
		if encoded := q.Encode(); encoded != "" {
			targetURL += "?" + encoded
		}
	}

	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Failed to construct remote request: " + err.Error()})
		return
	}
	if ct := r.Header.Get("Content-Type"); ct != "" {
		proxyReq.Header.Set("Content-Type", ct)
	}
	if acc := r.Header.Get("Accept"); acc != "" {
		proxyReq.Header.Set("Accept", acc)
	}
	if r.ContentLength > 0 {
		proxyReq.ContentLength = r.ContentLength
	}
	proxyReq.Header.Set("Authorization", "Bearer "+node.APIToken)

	resp, err := h.httpClient.Do(proxyReq)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Failed to reach remote node '" + node.Name + "': " + err.Error()})
		return
	}
	defer resp.Body.Close()

	for _, hdr := range []string{"Content-Type", "Content-Disposition", "Content-Length", "Cache-Control"} {
		if v := resp.Header.Get(hdr); v != "" {
			w.Header().Set(hdr, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (h *NodeHandler) ProxyWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		if cookie, err := r.Cookie("neopanel_token"); err == nil {
			token = cookie.Value
		}
	}
	if token == "" || h.auth == nil {
		http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
		return
	}
	if _, err := h.auth.ValidateAuthToken(token); err != nil {
		http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
		return
	}

	nodeID := getURLParam(r, "id")
	serverID := getURLParam(r, "serverId")
	if nodeID == "" || nodeID == "local" || serverID == "" {
		http.Error(w, "Invalid node or server ID", http.StatusBadRequest)
		return
	}

	node, err := h.db.GetNode(nodeID)
	if err != nil {
		http.Error(w, "Target node not found", http.StatusNotFound)
		return
	}

	parsed, err := url.Parse(node.BaseURL)
	if err != nil {
		http.Error(w, "Invalid node URL", http.StatusBadGateway)
		return
	}
	wsScheme := "ws"
	if parsed.Scheme == "https" {
		wsScheme = "wss"
	}
	parsed.Scheme = wsScheme
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/api/servers/" + url.PathEscape(serverID) + "/ws"
	q := url.Values{}
	q.Set("token", node.APIToken)
	parsed.RawQuery = q.Encode()

	dialHeader := http.Header{}
	dialHeader.Set("Authorization", "Bearer "+node.APIToken)

	dialer := *websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	remoteConn, _, err := dialer.DialContext(r.Context(), parsed.String(), dialHeader)
	if err != nil {
		http.Error(w, "Failed to connect to remote node console: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer remoteConn.Close()

	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer clientConn.Close()

	errCh := make(chan struct{}, 2)

	// Client -> Remote
	go func() {
		defer func() { errCh <- struct{}{} }()
		for {
			msgType, payload, err := clientConn.ReadMessage()
			if err != nil {
				return
			}
			if err := remoteConn.WriteMessage(msgType, payload); err != nil {
				return
			}
		}
	}()

	// Remote -> Client
	go func() {
		defer func() { errCh <- struct{}{} }()
		for {
			msgType, payload, err := remoteConn.ReadMessage()
			if err != nil {
				return
			}
			if err := clientConn.WriteMessage(msgType, payload); err != nil {
				return
			}
		}
	}()

	<-errCh
}

type SyncAuthRequest struct {
	Users []SyncedUser `json:"users"`
	Keys  []SyncedKey  `json:"keys"`
}

type SyncedUser struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
	Role         string `json:"role"`
}

type SyncedKey struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

func (h *NodeHandler) SyncAuthHandler(w http.ResponseWriter, r *http.Request) {
	var req SyncAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid auth sync payload: " + err.Error()})
		return
	}
	if r.Header.Get("X-NeoPanel-Invitation-Provision") == "1" {
		if len(req.Users) != 1 || len(req.Keys) != 0 {
			writeError(w, 400, "Expected one invitation account")
			return
		}
		u := req.Users[0]
		if u.ID <= 0 || !invitationUsername.MatchString(u.Username) || (u.Role != "user" && u.Role != "admin") {
			writeError(w, 400, "Invalid invitation account")
			return
		}
		if err := h.db.ProvisionInvitationUser(&database.User{ID: u.ID, Username: u.Username, Email: u.Email, PasswordHash: u.PasswordHash, Role: u.Role}); err != nil {
			invitationError(w, err)
			return
		}
		w.Header().Set("X-NeoPanel-Invitation-Provision", "1")
		writeJSON(w, 200, map[string]string{"status": "success"})
		return
	}
	for _, u := range req.Users {
		if err := h.db.UpsertSyncedUser(u.ID, u.Username, u.Email, u.PasswordHash, u.Role); err != nil {
			writeError(w, 409, "User synchronization failed")
			return
		}
	}
	for _, k := range req.Keys {
		if err := h.db.UpsertSyncedSSHKey(k.ID, k.UserID, k.Name, k.PublicKey); err != nil {
			writeError(w, 409, "Key synchronization failed")
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":       "success",
		"synced_users": len(req.Users),
		"synced_keys":  len(req.Keys),
	})
}

func (h *NodeHandler) SyncAuthToNode(ctx context.Context, baseURL, apiToken string) error {
	users, err := h.db.ListUsers()
	if err != nil || len(users) == 0 {
		return err
	}
	var syncUsers []SyncedUser
	for _, u := range users {
		syncUsers = append(syncUsers, SyncedUser{
			ID:           u.ID,
			Username:     u.Username,
			Email:        u.Email,
			PasswordHash: u.PasswordHash,
			Role:         u.Role,
		})
	}
	keys, _ := h.db.GetAllSSHKeys()
	var syncKeys []SyncedKey
	for _, k := range keys {
		syncKeys = append(syncKeys, SyncedKey{
			ID:        k.ID,
			UserID:    k.UserID,
			Name:      k.Name,
			PublicKey: k.PublicKey,
		})
	}

	payload, err := json.Marshal(SyncAuthRequest{
		Users: syncUsers,
		Keys:  syncKeys,
	})
	if err != nil {
		return err
	}

	targetURL := strings.TrimRight(baseURL, "/") + "/api/cluster/sync-auth"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("node auth synchronization failed: HTTP %d", resp.StatusCode)
	}
	return nil
}
