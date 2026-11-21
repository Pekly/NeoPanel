package api

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
)

const nodeWakeStatusLimit = 1 << 20

func (h *NodeHandler) GetNodeWakeStatus(w http.ResponseWriter, r *http.Request) {
	if !wakeAdmin(w, r) {
		return
	}
	id := r.PathValue("id")
	node, ok := h.wakeNode(w, id)
	if !ok {
		return
	}
	h.getWakeStatus(w, r, id, node)
}

// GetClusterWakeStatus is intentionally separate from configuration delivery.
// requireAdmin accepts authenticated native machine tokens as well as admins.
func (h *NodeHandler) GetClusterWakeStatus(w http.ResponseWriter, r *http.Request) {
	if !wakeAdmin(w, r) {
		return
	}
	h.getWakeStatus(w, r, "local", nil)
}

func (h *NodeHandler) getWakeStatus(w http.ResponseWriter, r *http.Request, id string, node *database.ClusterNode) {
	w.Header().Set("Cache-Control", "no-store")
	var report runner.NodeWakeStatus
	if node == nil {
		if h.manager == nil {
			writeError(w, 503, "Wake runtime status is unavailable")
			return
		}
		report = h.manager.GetNodeWakeStatus()
	} else {
		var err error
		report, err = fetchNodeWakeStatus(r.Context(), node)
		if err != nil {
			writeError(w, 502, err.Error())
			return
		}
	}
	// Read desired configuration AFTER the remote round trip, not from a delivery
	// acknowledgement or a pre-request copy that a concurrent save can supersede.
	desired, err := h.db.GetNodeWakeConfig(id)
	if err != nil {
		writeError(w, 500, "Unable to read desired wake configuration")
		return
	}
	report.ConfigMatches = report.ConfigMatches && report.ConfigHash == runner.NodeWakeConfigHash(desired) && report.Enabled == desired.Enabled
	if !report.ConfigMatches {
		if report.Error != "" {
			report.Error += "; "
		}
		report.Error += "Saved configuration is pending reconciliation on this node"
		for i := range report.Servers {
			report.Servers[i].State, report.Servers[i].Detail = "pending", "Waiting for reconciliation of the saved configuration"
		}
	}
	writeJSON(w, 200, report)
}

func fetchNodeWakeStatus(ctx context.Context, node *database.ClusterNode) (runner.NodeWakeStatus, error) {
	var report runner.NodeWakeStatus
	u, err := url.Parse(node.BaseURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return report, fmt.Errorf("Invalid node base URL; listener status unknown")
	}
	if strings.TrimSpace(node.APIToken) == "" {
		return report, fmt.Errorf("Node API token is missing; listener status unknown")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(node.BaseURL, "/")+"/api/cluster/wake/status", nil)
	if err != nil {
		return report, fmt.Errorf("Unable to request node listener status")
	}
	req.Header.Set("Authorization", "Bearer "+node.APIToken)
	client := &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		return report, fmt.Errorf("Cannot reach node listener status; status unknown")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 || resp.StatusCode == 405 {
			return report, fmt.Errorf("Node does not support listener status (HTTP %d); update the worker", resp.StatusCode)
		}
		return report, fmt.Errorf("Node listener status request failed (HTTP %d); status unknown", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, nodeWakeStatusLimit+1))
	if err != nil || len(body) > nodeWakeStatusLimit {
		return report, fmt.Errorf("Unable to read bounded node listener status")
	}
	report, err = decodeNodeWakeStatus(body, time.Now())
	if err != nil {
		return runner.NodeWakeStatus{}, fmt.Errorf("Invalid or stale node listener status; status unknown")
	}
	// Worker text is not trusted: never relay raw error strings, paths or tokens.
	for i := range report.Servers {
		report.Servers[i].Detail = wakeRemoteDetails[report.Servers[i].State]
	}
	if report.Error != "" {
		truncated := wakeStatusTruncation.FindString(report.Error)
		report.Error = "Node reports a listener status warning; check node logs"
		if truncated != "" {
			report.Error += "; " + truncated
		}
	}
	return report, nil
}

var wakeRemoteDetails = map[string]string{
	"listening":          "Listener bound and waiting for a player join",
	"cooldown":           "Listener bound; startup cooldown is active",
	"online":             "Server is already running",
	"starting":           "Server startup is queued or in progress",
	"busy":               "Another server operation is in progress",
	"blocked":            "Automatic startup is blocked; administrator action required",
	"conflict":           "Multiple local servers use this game port",
	"bind_failed":        "Port unavailable or bind address not available; check node logs",
	"pending":            "Waiting for listener reconciliation",
	"disabled":           "Join-to-start is disabled",
	"ineligible":         "Unsupported server type, server ID or game port",
	"node_idle_disabled": "Node idle shutdown is disabled",
	"user_disabled":      "Server join-to-start preference is disabled",
	"not_idle_stopped":   "Server was not stopped by the node idle policy",
}
var wakeStatusTruncation = regexp.MustCompile(`Server details truncated: [0-9]{1,10} additional local servers omitted`)

// Require exact fields, types and a single object; missing booleans must not
// silently decode to plausible defaults. Reject duplicate keys and nulls too.
func decodeWakeStatusObject(body []byte, fields map[string]any) error {
	d := json.NewDecoder(bytes.NewReader(body))
	token, err := d.Token()
	if err != nil || token != json.Delim('{') {
		return fmt.Errorf("expected object")
	}
	seen := make(map[string]bool)
	for d.More() {
		token, err := d.Token()
		if err != nil {
			return err
		}
		key, ok := token.(string)
		dest, exists := fields[key]
		if !ok || !exists || seen[key] {
			return fmt.Errorf("invalid field")
		}
		seen[key] = true
		var raw json.RawMessage
		if err := d.Decode(&raw); err != nil {
			return err
		}
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("null field")
		}
		if err := json.Unmarshal(raw, dest); err != nil {
			return err
		}
	}
	if token, err = d.Token(); err != nil || token != json.Delim('}') || len(seen) != len(fields) {
		return fmt.Errorf("incomplete object")
	}
	if d.Decode(new(any)) != io.EOF {
		return fmt.Errorf("trailing data")
	}
	return nil
}

func decodeNodeWakeStatus(body []byte, now time.Time) (runner.NodeWakeStatus, error) {
	var v runner.NodeWakeStatus
	bad := fmt.Errorf("invalid listener status")
	if len(body) > nodeWakeStatusLimit || !utf8.Valid(body) {
		return v, bad
	}
	var rows []json.RawMessage
	err := decodeWakeStatusObject(body, map[string]any{"checked_at": &v.CheckedAt, "enabled": &v.Enabled, "coordinator_running": &v.CoordinatorRunning, "config_matches": &v.ConfigMatches, "config_hash": &v.ConfigHash, "listeners": &v.Listeners, "servers": &rows, "error": &v.Error})
	if err != nil || len(rows) > runner.WakeStatusMaxServers || v.Listeners < 0 || v.Listeners > runner.WakeStatusMaxServers || len(v.Error) > 4096 {
		return v, bad
	}
	if v.CheckedAt.After(now.Add(2*time.Second)) || (!v.CheckedAt.IsZero() && now.Sub(v.CheckedAt) > runner.WakeStatusMaxAge) {
		return v, bad
	}
	if v.CheckedAt.IsZero() && (v.CoordinatorRunning || v.ConfigHash != "" || v.Listeners != 0) {
		return v, bad
	}
	if v.ConfigHash != "" {
		hash, err := hex.DecodeString(v.ConfigHash)
		if err != nil || len(hash) != 32 || strings.ToLower(v.ConfigHash) != v.ConfigHash {
			return v, bad
		}
	} else if v.ConfigMatches || v.Error == "" {
		return v, bad
	}
	v.Servers = make([]runner.WakeServerStatus, 0, len(rows))
	seen := make(map[string]bool)
	bound := 0
	for _, raw := range rows {
		var row runner.WakeServerStatus
		if err := decodeWakeStatusObject(raw, map[string]any{"id": &row.ID, "name": &row.Name, "port": &row.Port, "state": &row.State, "detail": &row.Detail}); err != nil {
			return v, bad
		}
		if row.ID == "" || len(row.ID) > 256 || len(row.Name) > 512 || len(row.Detail) > 1024 || seen[row.ID] || wakeRemoteDetails[row.State] == "" {
			return v, bad
		}
		seen[row.ID] = true
		if row.State != "ineligible" && row.State != "disabled" && row.State != "pending" && (row.Port < 1024 || row.Port > 65535) {
			return v, bad
		}
		if row.State == "listening" || row.State == "cooldown" {
			bound++
			if !v.CoordinatorRunning || !v.ConfigMatches || !v.Enabled || v.CheckedAt.IsZero() {
				return v, bad
			}
		}
		v.Servers = append(v.Servers, row)
	}
	if bound > v.Listeners || v.Listeners > len(rows) || (!v.Enabled && v.Listeners != 0) || (!v.CoordinatorRunning && v.Listeners != 0 && v.Error == "") {
		return v, bad
	}
	return v, nil
}
