package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/Pekly/NeoPanel/internal/database"
)

const nodeWakeBodyLimit = 128 * 1024

// Applied means configuration was persisted on the host, not that every game
// listener bound successfully. Remote status is the last acknowledged delivery.
type NodeWakeView struct {
	Config  database.NodeWakeConfig `json:"config"`
	Applied bool                    `json:"applied"`
	Error   string                  `json:"error"`
}

type nodeWakeDelivery struct {
	Fingerprint string `json:"fingerprint"`
	Applied     bool   `json:"applied"`
	Error       string `json:"error"`
}

// Require every exact field once, rejecting null, aliases, duplicate keys and
// trailing JSON. This is a complete replacement API, not an implicit patch.
func decodeWakeObject(body []byte, fields map[string]any) error {
	if len(body) > nodeWakeBodyLimit || !utf8.Valid(body) {
		return fmt.Errorf("invalid or oversized JSON")
	}
	d := json.NewDecoder(bytes.NewReader(body))
	token, err := d.Token()
	if err != nil || token != json.Delim('{') {
		return fmt.Errorf("expected a JSON object")
	}
	seen := make(map[string]bool)
	for d.More() {
		token, err := d.Token()
		if err != nil {
			return fmt.Errorf("invalid JSON field")
		}
		key, ok := token.(string)
		dest, exists := fields[key]
		if !ok || !exists || seen[key] {
			return fmt.Errorf("unknown or duplicate field %q", key)
		}
		seen[key] = true
		var raw json.RawMessage
		if err := d.Decode(&raw); err != nil {
			return fmt.Errorf("invalid %s", key)
		}
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("%s must not be null", key)
		}
		if _, isString := dest.(*string); isString && !wakeJSONUnicodeValid(raw) {
			return fmt.Errorf("%s contains an invalid Unicode escape", key)
		}
		if err := json.Unmarshal(raw, dest); err != nil {
			return fmt.Errorf("invalid %s", key)
		}
	}
	if token, err = d.Token(); err != nil || token != json.Delim('}') {
		return fmt.Errorf("invalid JSON object")
	}
	if len(seen) != len(fields) {
		return fmt.Errorf("all configuration fields are required")
	}
	if d.Decode(new(any)) != io.EOF {
		return fmt.Errorf("expected exactly one JSON object")
	}
	return nil
}

// encoding/json replaces unpaired UTF-16 surrogates with U+FFFD. Reject those
// rather than silently persisting different text; valid escaped emoji work.
func wakeJSONUnicodeValid(raw []byte) bool {
	for i := 0; i < len(raw); i++ {
		if raw[i] != '\\' {
			continue
		}
		i++
		if i >= len(raw) || raw[i] != 'u' {
			continue
		}
		if i+4 >= len(raw) {
			return false
		}
		r, err := strconv.ParseUint(string(raw[i+1:i+5]), 16, 16)
		if err != nil {
			return false
		}
		i += 4
		if r >= 0xDC00 && r <= 0xDFFF {
			return false
		}
		if r < 0xD800 || r > 0xDBFF {
			continue
		}
		if i+6 >= len(raw) || raw[i+1] != '\\' || raw[i+2] != 'u' {
			return false
		}
		low, err := strconv.ParseUint(string(raw[i+3:i+7]), 16, 16)
		if err != nil || low < 0xDC00 || low > 0xDFFF {
			return false
		}
		i += 6
	}
	return true
}

func decodeWakeConfig(body []byte) (database.NodeWakeConfig, error) {
	var c database.NodeWakeConfig
	err := decodeWakeObject(body, map[string]any{
		"enabled": &c.Enabled, "bind_address": &c.BindAddress,
		"offline_motd": &c.OfflineMOTD, "version_label": &c.VersionLabel,
		"max_players": &c.MaxPlayers, "icon_data_url": &c.IconDataURL,
		"use_server_icon": &c.UseServerIcon, "starting_message": &c.StartingMessage,
		"unavailable_message": &c.UnavailableMessage, "cooldown_seconds": &c.CooldownSeconds,
	})
	if err == nil {
		err = database.ValidateNodeWakeConfig(c)
	}
	return c, err
}

func wakeAdmin(w http.ResponseWriter, r *http.Request) bool {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil || claims.Role != "admin" {
		writeError(w, http.StatusForbidden, "Administrator privileges required")
		return false
	}
	return true
}

func (h *NodeHandler) wakeLock(id string) *sync.Mutex {
	lock, _ := h.wakeLocks.LoadOrStore(id, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func (h *NodeHandler) wakeNode(w http.ResponseWriter, id string) (*database.ClusterNode, bool) {
	if id == "local" {
		return nil, true
	}
	node, err := h.db.GetNode(id)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, 404, "Node not found")
		return nil, false
	}
	if err != nil {
		writeError(w, 500, "Failed to load node")
		return nil, false
	}
	return node, true
}

func wakeFingerprint(c database.NodeWakeConfig, node *database.ClusterNode) string {
	data, _ := json.Marshal(struct {
		Config     database.NodeWakeConfig
		URL, Token string
	}{c, node.BaseURL, node.APIToken})
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func (h *NodeHandler) saveWakeDelivery(id string, state nodeWakeDelivery) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return h.db.SetPanelSetting("node_wake_delivery:"+id, string(data))
}

func (h *NodeHandler) GetNodeWake(w http.ResponseWriter, r *http.Request) {
	if !wakeAdmin(w, r) {
		return
	}
	id := r.PathValue("id")
	node, ok := h.wakeNode(w, id)
	if !ok {
		return
	}
	lock := h.wakeLock(id)
	lock.Lock()
	defer lock.Unlock()
	c, err := h.db.GetNodeWakeConfig(id)
	if err != nil {
		writeError(w, 500, "Failed to load wake configuration")
		return
	}
	view := NodeWakeView{Config: c, Applied: node == nil}
	if node != nil {
		view.Error = "Configuration has not been confirmed on this node; save to retry delivery"
		raw, err := h.db.GetPanelSetting("node_wake_delivery:" + id)
		var state nodeWakeDelivery
		if err == nil && json.Unmarshal([]byte(raw), &state) == nil && state.Fingerprint == wakeFingerprint(c, node) {
			view.Applied, view.Error = state.Applied, state.Error
			if !view.Applied && view.Error == "" {
				view.Error = "Configuration delivery pending"
			}
		} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
			view.Error = "Unable to read configuration delivery status"
		}
	}
	writeJSON(w, http.StatusOK, view)
}

func (h *NodeHandler) PutNodeWake(w http.ResponseWriter, r *http.Request) {
	h.putNodeWake(w, r, r.PathValue("id"))
}

// ApplyClusterWake deliberately only writes the worker's local configuration.
func (h *NodeHandler) ApplyClusterWake(w http.ResponseWriter, r *http.Request) {
	h.putNodeWake(w, r, "local")
}

func (h *NodeHandler) putNodeWake(w http.ResponseWriter, r *http.Request, id string) {
	if !wakeAdmin(w, r) {
		return
	}
	node, ok := h.wakeNode(w, id)
	if !ok {
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, nodeWakeBodyLimit))
	if err != nil {
		writeError(w, 400, "Wake configuration exceeds request limit")
		return
	}
	c, err := decodeWakeConfig(body)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	// Serialize persistence AND delivery, including disables. An older request
	// can never acknowledge or overwrite a newer desired configuration.
	lock := h.wakeLock(id)
	lock.Lock()
	defer lock.Unlock()
	state := nodeWakeDelivery{Error: "Configuration delivery pending"}
	if node != nil {
		state.Fingerprint = wakeFingerprint(c, node)
		// Invalidate the previous acknowledgement first: a crash cannot falsely
		// report the newly desired configuration as applied (even an ABA update).
		if err := h.saveWakeDelivery(id, state); err != nil {
			writeError(w, 500, "Failed to persist pending delivery")
			return
		}
	}
	if err := h.db.SetNodeWakeConfig(id, c); err != nil {
		writeError(w, 500, "Failed to persist wake configuration")
		return
	}
	view := NodeWakeView{Config: c, Applied: true}
	status := http.StatusOK
	if node != nil {
		err := h.deliverNodeWake(r.Context(), node, c)
		state.Applied, state.Error = err == nil, ""
		if err != nil {
			state.Error = err.Error()
			status = http.StatusBadGateway
		}
		if err := h.saveWakeDelivery(id, state); err != nil {
			state.Applied, state.Error = false, "Unable to persist delivery acknowledgement; save to retry"
			status = http.StatusInternalServerError
		}
		view.Applied, view.Error = state.Applied, state.Error
	}
	writeJSON(w, status, view)
}

func (h *NodeHandler) deliverNodeWake(ctx context.Context, node *database.ClusterNode, c database.NodeWakeConfig) error {
	u, err := url.Parse(node.BaseURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("Invalid node base URL; configuration delivery pending")
	}
	if strings.TrimSpace(node.APIToken) == "" {
		return fmt.Errorf("Node API token is missing; configuration delivery pending")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	data, _ := json.Marshal(c)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, strings.TrimRight(node.BaseURL, "/")+"/api/cluster/wake", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("Unable to construct wake delivery request")
	}
	req.Header.Set("Authorization", "Bearer "+node.APIToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	if h.httpClient != nil {
		client.Transport = h.httpClient.Transport
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Node unreachable or request timed out; configuration delivery pending")
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		return fmt.Errorf("Node returned HTTP %d for PUT /api/cluster/wake. Settings are saved on this panel but worker application is unconfirmed. Update and restart the worker to a join-to-start capable version (updating only the panel is not enough). Check that the node URL targets that daemon and its reverse proxy forwards PUT /api/cluster/wake, then Retry saved settings", resp.StatusCode)
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("Node returned HTTP %d for PUT /api/cluster/wake. Settings are saved on this panel but worker application is unconfirmed. Check the node's saved Node Token or administrator API key and any proxy access rules, then Retry saved settings", resp.StatusCode)
	}
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		return fmt.Errorf("Node returned HTTP %d redirect. Settings delivery does not follow redirects, to protect the node token. Set the node URL to the daemon's final HTTP/HTTPS address, then Retry saved settings", resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Node returned HTTP %d; configuration delivery pending", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, nodeWakeBodyLimit+1))
	if err != nil {
		return fmt.Errorf("Unable to read node acknowledgement; configuration delivery pending")
	}
	var raw json.RawMessage
	var applied bool
	var remoteError string
	err = decodeWakeObject(body, map[string]any{"config": &raw, "applied": &applied, "error": &remoteError})
	if err != nil {
		return fmt.Errorf("Invalid node acknowledgement; configuration delivery pending")
	}
	actual, err := decodeWakeConfig(raw)
	if err != nil || !applied || remoteError != "" || actual != c {
		return fmt.Errorf("Node did not confirm the requested configuration; configuration delivery pending")
	}
	return nil
}
