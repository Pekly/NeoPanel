package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

type featureView struct {
	Policy    database.FeaturePolicy `json:"policy"`
	Global    database.FeaturePolicy `json:"global"`
	Effective database.FeaturePolicy `json:"effective"`
	Nodes     map[string]bool        `json:"nodes,omitempty"`
}

func featureForRoute(method, route string) string {
	parts := strings.Split(strings.Trim(route, "/"), "/")
	if len(parts) < 4 || parts[0] != "api" || parts[1] != "servers" {
		return ""
	}
	suffix := strings.Join(parts[3:], "/")
	// Reading results, cancellation and cleanup remain available after a disable.
	if strings.HasPrefix(suffix, "world-conversion") || strings.HasPrefix(suffix, "files/import-sftp") {
		if method == http.MethodGet || method == http.MethodDelete {
			return ""
		}
		if strings.HasPrefix(suffix, "world-conversion") {
			return "world_conversion"
		}
		return "sftp_import"
	}
	if suffix == "power/countdown" {
		if method != http.MethodPost {
			return ""
		}
		return "countdowns"
	}
	if suffix == "ports" || strings.HasPrefix(suffix, "ports/") {
		if method == http.MethodPost && suffix == "ports" {
			return "extra_ports"
		}
		return ""
	}
	switch parts[3] {
	case "backups":
		if method == http.MethodGet || method == http.MethodDelete || suffix == "backups/restore" {
			return ""
		}
		return "backups"
	case "schedules":
		return "schedules"
	case "sftp":
		return "sftp"
	case "plugins":
		return "plugins"
	case "databases":
		return "databases"
	case "players":
		return "players"
	case "proxy-config":
		return "proxy_settings"
	case "software", "reinstall":
		return "software"
	case "subdomain":
		return "subdomains"
	}
	return ""
}

func (r *Router) featureAllowed(w http.ResponseWriter, req *http.Request) bool {
	key := featureForRoute(req.Method, req.URL.Path)
	if key == "" {
		return true
	}
	if !r.auth.db.FeatureEnabled(req.PathValue("id"), key) {
		writeError(w, http.StatusForbidden, "This feature is disabled by an administrator")
		return false
	}
	return true
}

func (r *Router) getFeatures(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	p, err := r.auth.db.GetFeaturePolicy(id)
	if err != nil {
		writeError(w, 503, "Feature settings are unavailable")
		return
	}
	g, err := r.auth.db.GetFeaturePolicy("")
	if err != nil {
		writeError(w, 503, "Feature settings are unavailable")
		return
	}
	e, err := r.auth.db.EffectiveFeatures(id)
	if err != nil {
		writeError(w, 503, "Feature settings are unavailable")
		return
	}
	view := featureView{Policy: p, Global: g, Effective: e}
	if id == "" && getUserFromContext(req) != nil && getUserFromContext(req).Role == "admin" {
		view.Nodes = r.featureNodeStatus(g)
	}
	writeJSON(w, 200, view)
}

func decodeFeatures(w http.ResponseWriter, req *http.Request) (database.FeaturePolicy, bool) {
	var raw map[string]json.RawMessage
	dec := json.NewDecoder(http.MaxBytesReader(w, req.Body, 8192))
	if err := dec.Decode(&raw); err != nil || raw == nil {
		writeError(w, 400, "Expected feature settings")
		return nil, false
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		writeError(w, 400, "Expected one settings object")
		return nil, false
	}
	p := database.FeaturePolicy{}
	for key, value := range raw {
		if !database.ValidFeature(key) || (string(value) != "true" && string(value) != "false") {
			writeError(w, 400, "Unknown feature or invalid switch value")
			return nil, false
		}
		p[key] = string(value) == "true"
	}
	return p, true
}

func (r *Router) putFeatures(w http.ResponseWriter, req *http.Request) {
	p, ok := decodeFeatures(w, req)
	if !ok {
		return
	}
	id := req.PathValue("id")
	r.featureMu.Lock()
	defer r.featureMu.Unlock()
	if id == "" && !strings.HasPrefix(req.URL.Path, "/api/cluster/") {
		nodes, err := r.auth.db.ListNodes()
		if err != nil {
			writeError(w, 503, "Unable to prepare node synchronization")
			return
		}
		for _, node := range nodes {
			if err := r.auth.db.DeletePanelSetting("features:delivery:" + node.ID); err != nil {
				writeError(w, 500, "Unable to prepare node synchronization")
				return
			}
		}
	}
	if err := r.auth.db.SetFeaturePolicy(id, p); err != nil {
		writeError(w, 500, "Unable to save feature settings")
		return
	}
	_ = r.auth.db.CreateAuditLog("admin", "features.update", "Updated optional features for "+id, getClientIP(req))
	if id == "" && !strings.HasPrefix(req.URL.Path, "/api/cluster/") {
		nodes, err := r.auth.db.ListNodes()
		if err != nil {
			writeError(w, 503, "Saved locally; node synchronization is pending")
			return
		}
		for _, node := range nodes {
			if node.ID != "local" && node.BaseURL != "" {
				_ = r.syncFeatureNode(req.Context(), node, p)
			}
		}
	}
	r.getFeatures(w, req)
}

func featureFingerprint(node *database.ClusterNode, p database.FeaturePolicy) string {
	data, _ := json.Marshal(struct {
		URL, Token string
		Policy     database.FeaturePolicy
	}{node.BaseURL, node.APIToken, p})
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
func (r *Router) featureNodeStatus(p database.FeaturePolicy) map[string]bool {
	result := map[string]bool{}
	nodes, err := r.auth.db.ListNodes()
	if err != nil {
		return result
	}
	for _, node := range nodes {
		if node.ID == "local" || node.BaseURL == "" {
			continue
		}
		value, _ := r.auth.db.GetPanelSetting("features:delivery:" + node.ID)
		result[node.ID] = value == featureFingerprint(node, p)
	}
	return result
}
func (r *Router) syncFeatureNode(ctx context.Context, node *database.ClusterNode, p database.FeaturePolicy) error {
	fingerprint := featureFingerprint(node, p)
	value, _ := r.auth.db.GetPanelSetting("features:delivery:" + node.ID)
	if value == fingerprint {
		return nil
	}
	u, err := url.Parse(node.BaseURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || strings.TrimSpace(node.APIToken) == "" {
		return fmt.Errorf("invalid node destination")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	data, _ := json.Marshal(p)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, strings.TrimRight(node.BaseURL, "/")+"/api/cluster/features", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+node.APIToken)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	if r.nodes.httpClient != nil {
		client.Transport = r.nodes.httpClient.Transport
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var view featureView
	if resp.StatusCode != 200 || json.NewDecoder(io.LimitReader(resp.Body, 16384)).Decode(&view) != nil {
		return fmt.Errorf("node did not accept feature settings")
	}
	for _, key := range database.FeatureKeys {
		want, set := p[key]
		if !set {
			want = true
		}
		got, present := view.Effective[key]
		if !present || got != want {
			return fmt.Errorf("node did not confirm feature settings")
		}
	}
	return r.auth.db.SetPanelSetting("features:delivery:"+node.ID, fingerprint)
}

func (r *Router) syncProxyFeatures(w http.ResponseWriter, req *http.Request, node *database.ClusterNode) bool {
	r.featureMu.Lock()
	defer r.featureMu.Unlock()
	p, err := r.auth.db.GetFeaturePolicy("")
	if err == nil && len(p) == 0 { // Preserve compatibility until feature controls are first configured.
		raw, _ := r.auth.db.GetPanelSetting("features:global")
		if raw == "" {
			return true
		}
	}
	if err == nil {
		err = r.syncFeatureNode(req.Context(), node, p)
	}
	if err != nil {
		writeError(w, 503, "Node feature settings are pending. Ask an administrator to retry synchronization.")
		return false
	}
	return true
}
