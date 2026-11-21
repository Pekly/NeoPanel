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
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

const nodeIdleBodyLimit = 4096

type NodeIdlePolicyView struct {
	Config  database.NodeIdlePolicy `json:"config"`
	Applied bool                    `json:"applied"`
	Error   string                  `json:"error"`
}

type nodeIdleDelivery struct {
	Fingerprint string `json:"fingerprint"`
	Applied     bool   `json:"applied"`
	Error       string `json:"error"`
}

func decodeIdlePolicy(body []byte) (database.NodeIdlePolicy, error) {
	var p database.NodeIdlePolicy
	if len(body) > nodeIdleBodyLimit {
		return p, fmt.Errorf("idle policy exceeds request limit")
	}
	err := decodeWakeObject(body, map[string]any{"enabled": &p.Enabled, "timeout_minutes": &p.TimeoutMinutes})
	if err == nil {
		err = database.ValidateNodeIdlePolicy(p)
	}
	return p, err
}

// Generic forms historically sent zero for an unset timeout. Omitted fields
// preserve policy during re-pairing; false always means disabled.
func mergeNodeIdlePolicy(node *database.ClusterNode, enabled *bool, minutes *int) error {
	p := database.DefaultNodeIdlePolicy()
	if node != nil {
		p = database.NodeIdlePolicy{Enabled: node.IsFree, TimeoutMinutes: node.IdleTimeoutMin}
	}
	if enabled != nil {
		p.Enabled = *enabled
	}
	if minutes != nil {
		p.TimeoutMinutes = *minutes
	}
	if p.TimeoutMinutes == 0 {
		p.TimeoutMinutes = 15
	}
	if err := database.ValidateNodeIdlePolicy(p); err != nil {
		return err
	}
	if node != nil {
		node.IsFree, node.IdleTimeoutMin = p.Enabled, p.TimeoutMinutes
	}
	return nil
}

func (h *NodeHandler) idleLock(id string) *sync.Mutex {
	lock, _ := h.idleLocks.LoadOrStore(id, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func idleFingerprint(p database.NodeIdlePolicy, node *database.ClusterNode) string {
	data, _ := json.Marshal(struct {
		Config     database.NodeIdlePolicy
		URL, Token string
	}{p, node.BaseURL, node.APIToken})
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func (h *NodeHandler) saveIdleDelivery(id string, state nodeIdleDelivery) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return h.db.SetPanelSetting("node_idle_delivery:"+id, string(data))
}

// Called under idleLock before changing either policy or destination. Invalidates
// even an identical/ABA save, so a crash cannot retain an earlier acknowledgement.
func (h *NodeHandler) pendingIdleDelivery(node *database.ClusterNode) error {
	p := database.NodeIdlePolicy{Enabled: node.IsFree, TimeoutMinutes: node.IdleTimeoutMin}
	return h.saveIdleDelivery(node.ID, nodeIdleDelivery{Fingerprint: idleFingerprint(p, node), Error: "Idle policy delivery pending; save to retry"})
}

func (h *NodeHandler) syncNodeIdlePolicy(ctx context.Context, node *database.ClusterNode) NodeIdlePolicyView {
	p := database.NodeIdlePolicy{Enabled: node.IsFree, TimeoutMinutes: node.IdleTimeoutMin}
	view := NodeIdlePolicyView{Config: p, Applied: true}
	if err := h.deliverNodeIdlePolicy(ctx, node, p); err != nil {
		view.Applied, view.Error = false, err.Error()
	}
	state := nodeIdleDelivery{Fingerprint: idleFingerprint(p, node), Applied: view.Applied, Error: view.Error}
	if err := h.saveIdleDelivery(node.ID, state); err != nil {
		view.Applied, view.Error = false, "Unable to persist idle policy acknowledgement; save to retry"
	}
	return view
}

func (h *NodeHandler) GetNodeIdlePolicy(w http.ResponseWriter, r *http.Request) {
	if !wakeAdmin(w, r) {
		return
	}
	id := r.PathValue("id")
	lock := h.idleLock(id)
	lock.Lock()
	defer lock.Unlock()
	node, ok := h.wakeNode(w, id)
	if !ok {
		return
	}
	p, err := h.db.GetNodeIdlePolicy(id)
	view := NodeIdlePolicyView{Config: p, Applied: node == nil}
	if err != nil {
		view.Applied, view.Error = false, "Failed to load idle policy; policy is disabled until corrected"
		writeJSON(w, 500, view)
		return
	}
	if node != nil {
		view.Error = "Idle policy has not been confirmed on this node; save to retry"
		raw, err := h.db.GetPanelSetting("node_idle_delivery:" + id)
		var state nodeIdleDelivery
		if err == nil && json.Unmarshal([]byte(raw), &state) == nil && state.Fingerprint == idleFingerprint(p, node) {
			view.Applied, view.Error = state.Applied && state.Error == "", state.Error
			if !view.Applied && view.Error == "" {
				view.Error = "Idle policy delivery pending; save to retry"
			}
		}
	}
	writeJSON(w, 200, view)
}

func (h *NodeHandler) PutNodeIdlePolicy(w http.ResponseWriter, r *http.Request) {
	h.putNodeIdlePolicy(w, r, r.PathValue("id"))
}
func (h *NodeHandler) ApplyClusterIdlePolicy(w http.ResponseWriter, r *http.Request) {
	h.putNodeIdlePolicy(w, r, "local")
}

func (h *NodeHandler) putNodeIdlePolicy(w http.ResponseWriter, r *http.Request, id string) {
	if !wakeAdmin(w, r) {
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, nodeIdleBodyLimit))
	if err != nil {
		writeError(w, 400, "Idle policy exceeds request limit")
		return
	}
	p, err := decodeIdlePolicy(body)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	lock := h.idleLock(id)
	lock.Lock()
	defer lock.Unlock()
	// Refetch only after locking: a generic save may have changed the destination.
	node, ok := h.wakeNode(w, id)
	if !ok {
		return
	}
	if node != nil {
		node.IsFree, node.IdleTimeoutMin = p.Enabled, p.TimeoutMinutes
		if err := h.pendingIdleDelivery(node); err != nil {
			writeError(w, 500, "Failed to persist pending idle delivery")
			return
		}
	}
	if err := h.db.SetNodeIdlePolicy(id, p); err != nil {
		writeError(w, 500, "Failed to persist idle policy")
		return
	}
	view := NodeIdlePolicyView{Config: p, Applied: true}
	status := http.StatusOK
	if node != nil {
		view = h.syncNodeIdlePolicy(r.Context(), node)
		if !view.Applied {
			status = http.StatusBadGateway
		}
	}
	writeJSON(w, status, view)
}

func (h *NodeHandler) deliverNodeIdlePolicy(ctx context.Context, node *database.ClusterNode, p database.NodeIdlePolicy) error {
	u, err := url.Parse(node.BaseURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("Invalid node URL; idle policy delivery pending")
	}
	if strings.TrimSpace(node.APIToken) == "" {
		return fmt.Errorf("Node API token missing; idle policy delivery pending")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	data, _ := json.Marshal(p)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, strings.TrimRight(node.BaseURL, "/")+"/api/cluster/idle-policy", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("Unable to construct idle policy delivery request")
	}
	req.Header.Set("Authorization", "Bearer "+node.APIToken)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	if h.httpClient != nil {
		client.Transport = h.httpClient.Transport
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Node unreachable or request timed out; idle policy delivery pending")
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 || resp.StatusCode == 405 {
		return fmt.Errorf("Node returned HTTP %d for PUT /api/cluster/idle-policy; desired policy saved but unconfirmed. Update/restart the worker and check proxy routing, then retry", resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Node returned HTTP %d; idle policy delivery pending (redirects are not followed)", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, nodeIdleBodyLimit+1))
	if err != nil || len(body) > nodeIdleBodyLimit {
		return fmt.Errorf("Invalid idle policy acknowledgement; delivery pending")
	}
	var raw json.RawMessage
	var applied bool
	var remoteError string
	err = decodeWakeObject(body, map[string]any{"config": &raw, "applied": &applied, "error": &remoteError})
	if err != nil {
		return fmt.Errorf("Invalid idle policy acknowledgement; delivery pending")
	}
	actual, err := decodeIdlePolicy(raw)
	if err != nil || !applied || remoteError != "" || actual != p {
		return fmt.Errorf("Node did not confirm requested idle policy; delivery pending")
	}
	return nil
}
