package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/downloader"
	"github.com/Pekly/NeoPanel/internal/proxyconfig"
	"github.com/Pekly/NeoPanel/internal/rooted"
	"github.com/Pekly/NeoPanel/internal/runner"
)

func (h *ServerHandler) requirePortAccess(w http.ResponseWriter, r *http.Request, id, permission string) bool {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if status := authorizeServer(h.db, claims, id, permission); status != 0 {
		writeError(w, status, "Port settings access denied")
		return false
	}
	return true
}
func portAllocationMessage(primary bool) string {
	if primary {
		return "Primary configuration saved for the next manual start. No process was started or restarted; external reachability is not verified."
	}
	return "Additional port assigned to this server. Multiple ports can be used simultaneously when the server or its plugins listen on each port with the selected protocol. External firewall and router access is not verified."
}
func writePortError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, runner.ErrProxyConfigRunning) || errors.Is(err, runner.ErrServerOperation) || errors.Is(err, runner.ErrServerProcessAlive) || errors.Is(err, runner.ErrRestoreRecovery) {
		status = http.StatusConflict
	}
	writeError(w, status, err.Error())
}

// GetServerPortSummary is additive: the existing GET /ports remains an array.
func (h *ServerHandler) GetServerPortSummary(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if !h.requirePortAccess(w, r, id, "settings.view") {
		return
	}
	s, err := h.db.GetServer(id)
	if err != nil {
		writeError(w, 404, "Server not found")
		return
	}
	writeJSON(w, 200, map[string]int{"used": s.PortsUsed, "allowed": s.PortsAllowed})
}
func (h *ServerHandler) SetServerPortAllowance(w http.ResponseWriter, r *http.Request) {
	id := getURLParam(r, "id")
	if !h.requirePortAccess(w, r, id, "settings.edit") {
		return
	}
	if u := getUserFromContext(r); u == nil || u.Role != "admin" {
		writeError(w, 403, "Only administrators can change the port allowance")
		return
	}
	var req struct {
		Allowed int `json:"allowed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}
	if err := h.db.SetServerPortAllowance(id, req.Allowed); err != nil {
		writePortError(w, err)
		return
	}
	logAudit(h.db, r, "port_allowance", fmt.Sprintf("Set total port allowance for %s to %d", id, req.Allowed))
	h.GetServerPortSummary(w, r)
}
func (h *ServerHandler) GetPortPool(w http.ResponseWriter, r *http.Request) {
	if u := getUserFromContext(r); u == nil || u.Role != "admin" {
		writeError(w, 403, "Administrator access required")
		return
	}
	pool, err := h.db.GetPortPool()
	if err != nil {
		writeError(w, 409, err.Error())
		return
	}
	writeJSON(w, 200, pool)
}
func (h *ServerHandler) SetPortPool(w http.ResponseWriter, r *http.Request) {
	if u := getUserFromContext(r); u == nil || u.Role != "admin" {
		writeError(w, 403, "Administrator access required")
		return
	}
	var pool database.PortPool
	if err := json.NewDecoder(r.Body).Decode(&pool); err != nil {
		writeError(w, 400, "Invalid request body")
		return
	}
	if err := h.db.SetPortPool(pool); err != nil {
		writePortError(w, err)
		return
	}
	logAudit(h.db, r, "port_pool", fmt.Sprintf("Set member port allocation pool to %d..%d", pool.Start, pool.End))
	writeJSON(w, 200, pool)
}

type portConfigSnapshot struct {
	name   string
	data   []byte
	mode   os.FileMode
	exists bool
}

func snapshotPortConfig(root *rooted.Root, name string) (portConfigSnapshot, error) {
	s := portConfigSnapshot{name: name}
	if err := root.Check(name); err != nil {
		return s, err
	}
	st, err := root.Lstat(name)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	if !st.Mode().IsRegular() {
		return s, fmt.Errorf("port configuration is not a regular file")
	}
	s.data, err = root.ReadFile(name)
	s.mode = st.Mode().Perm()
	s.exists = true
	return s, err
}
func (s portConfigSnapshot) restore(root *rooted.Root) error {
	if err := root.Check(s.name); err != nil {
		return err
	}
	if s.exists {
		return root.WriteFile(s.name, s.data, s.mode)
	}
	err := root.Remove(s.name)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// The type-agnostic proxy gate checks stopped status, cancels retries/wake,
// takes the process run lock and excludes power/restore/import AND managed
// API/SFTP mutations. Retain it through snapshot, DB commit, exact rollback
// and in-memory refresh; otherwise rollback could discard an editor's write.
func (h *ServerHandler) withPrimaryPortChange(r *http.Request, id string, change func(database.PortConfigApply) error) error {
	srv, err := h.db.GetServer(id)
	if err != nil {
		return err
	}
	if srv.NodeID != "" && srv.NodeID != "local" && (h.cfg == nil || h.cfg.Server.Role != "node") {
		return fmt.Errorf("change primary ports on the hosting worker, not remote metadata")
	}
	release, err := h.manager.BeginProxyConfigOperation(r.Context(), id)
	if err != nil {
		return err
	}
	releaseGate := true
	defer func() {
		if releaseGate {
			release()
		}
	}()
	// Re-read after taking the gate: an earlier settings/primary change may have won.
	srv, err = h.db.GetServer(id)
	if err != nil {
		return err
	}
	root, err := rooted.Open(h.manager.GetWorkingDir(id))
	if err != nil {
		return err
	}
	defer root.Close()
	name := "server.properties"
	if downloader.IsNativeSoftware(srv.Type) {
		name = filepath.Join("config", "config.toml")
	} else if proxyconfig.Supported(srv.Type) {
		name = proxyconfig.FileName(srv.Type)
	}
	snap, err := snapshotPortConfig(root, name)
	if err != nil {
		return err
	}
	snapshots := []portConfigSnapshot{snap}
	// Ensure can create this secret only when initializing a missing Velocity file.
	if srv.Type == "velocity" && !snap.exists {
		secret, err := snapshotPortConfig(root, "forwarding.secret")
		if err != nil {
			return err
		}
		snapshots = append(snapshots, secret)
	}
	attempted := false
	err = change(func(port int) error { attempted = true; return h.updatePrimaryPortConfig(srv, port) })
	if err != nil {
		if attempted {
			for _, s := range snapshots {
				if restoreErr := s.restore(root); restoreErr != nil {
					// Do not permit a start with mismatched configuration. Operator repair is
					// required; intentionally keep the operation/run locks until daemon exit.
					releaseGate = false
					err = errors.Join(err, fmt.Errorf("port configuration rollback failed; server locked pending operator repair: %w", restoreErr))
				}
			}
		}
		return err
	}
	srv, err = h.db.GetServer(id)
	if err != nil {
		return err
	}
	h.manager.UpdateInstanceServer(srv)
	return nil
}
