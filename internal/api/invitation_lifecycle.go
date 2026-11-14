package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

// Durable receipts are safe to reconcile even after the server itself is gone.
func (r *Router) invitationRemovalReceipts(w http.ResponseWriter, req *http.Request) {
	server := req.URL.Query().Get("server_id")
	if !invitationID.MatchString(server) {
		writeError(w, 400, "Invalid server ID")
		return
	}
	ids, err := r.auth.db.RevokedInvitationGrants(server)
	if err != nil {
		writeError(w, 500, "Could not read invitation removal receipts")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-NeoPanel-Invitation-Cleanup", "1")
	writeJSON(w, 200, map[string]any{"revoked": ids})
}

func (r *Router) reconcileRemoteInvitationRemovals(req *http.Request, node *database.ClusterNode, server string) error {
	remote, err := http.NewRequestWithContext(req.Context(), http.MethodGet, strings.TrimRight(node.BaseURL, "/")+"/api/cluster/invitation-removals?server_id="+url.QueryEscape(server), nil)
	if err != nil {
		return err
	}
	remote.Header.Set("Authorization", "Bearer "+node.APIToken)
	client := *r.nodes.httpClient
	client.Timeout = 15 * time.Second
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(remote)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 || resp.Header.Get("X-NeoPanel-Invitation-Cleanup") != "1" {
		return fmt.Errorf("node lacks invitation cleanup receipts")
	}
	var result struct {
		Revoked []string `json:"revoked"`
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
		return err
	}
	return r.auth.db.ReconcileInvitationRemovals(node.ID, server, result.Revoked)
}

// Retry endpoint for direct worker deletions, old workers and uncertain responses.
// This reconciles receipts only; it cannot authorize global account deletion.
func (r *Router) reconcileInvitationRemovals(w http.ResponseWriter, req *http.Request) {
	server := req.URL.Query().Get("server_id")
	if !invitationID.MatchString(server) {
		writeError(w, 400, "Invalid server ID")
		return
	}
	node, err := r.auth.db.GetNode(getURLParam(req, "id"))
	if err != nil {
		writeError(w, 404, "Node not found")
		return
	}
	if err = r.reconcileRemoteInvitationRemovals(req, node, server); err != nil {
		writeError(w, 502, "Reconciliation unresolved; accounts retained")
		return
	}
	writeJSON(w, 200, map[string]string{"status": "reconciled", "accounts": "retained: global associations cannot be proven absent"})
}

func removedRemoteServer(method, rest string) string {
	if method != http.MethodDelete {
		return ""
	}
	parts := strings.Split(rest, "/")
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "servers" && invitationID.MatchString(parts[2]) &&
		(len(parts) == 3 || (len(parts) == 5 && parts[3] == "subusers")) {
		return parts[2]
	}
	return ""
}
