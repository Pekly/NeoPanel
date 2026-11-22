package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/proxyconfig"
	"github.com/Pekly/NeoPanel/internal/rooted"
	"github.com/Pekly/NeoPanel/internal/runner"
	"github.com/pelletier/go-toml/v2"
)

type proxyConfigView struct {
	proxyconfig.Snapshot
	Port          int  `json:"port"`
	CanEdit       bool `json:"can_edit"`
	CanViewSecret bool `json:"can_view_secret"`
	Running       bool `json:"running"`
	Ready         bool `json:"ready"`
}

func (h *ServerHandler) proxyServer(w http.ResponseWriter, r *http.Request) *database.Server {
	srv, err := h.db.GetServer(r.PathValue("id"))
	if err != nil || srv == nil {
		writeError(w, 404, "Server not found")
		return nil
	}
	if !proxyconfig.Supported(srv.Type) {
		writeError(w, 400, "These settings are for Velocity, BungeeCord and Waterfall proxies")
		return nil
	}
	if srv.NodeID != "" && srv.NodeID != "local" {
		writeError(w, 409, "Open this proxy's own Settings page to configure it")
		return nil
	}
	return srv
}

func (h *ServerHandler) proxyView(r *http.Request, srv *database.Server, snapshot proxyconfig.Snapshot) proxyConfigView {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if snapshot.Config.Backends == nil {
		snapshot.Config.Backends = []proxyconfig.Backend{}
	}
	if snapshot.Config.Try == nil {
		snapshot.Config.Try = []string{}
	}
	if snapshot.Config.ForcedHosts == nil {
		snapshot.Config.ForcedHosts = []proxyconfig.ForcedHost{}
	}
	return proxyConfigView{Snapshot: snapshot, Port: srv.Port,
		CanEdit:       authorizeServer(h.db, claims, srv.ID, "settings.edit") == 0,
		CanViewSecret: strings.EqualFold(srv.Type, "velocity") && authorizeServer(h.db, claims, srv.ID, "settings.edit+files.view") == 0,
		Running:       h.manager.GetStatus(srv.ID) != "offline",
		Ready:         snapshot.Exists && proxyconfig.Validate(snapshot.Type, snapshot.Config) == nil,
	}
}

func writeProxyConfigError(w http.ResponseWriter, r *http.Request, err error) {
	status, message := http.StatusBadRequest, err.Error()
	code := proxyconfig.ErrorCode(err)
	switch {
	case errors.Is(err, proxyconfig.ErrConflict):
		code = "revision_conflict"
		status, message = 409, "The config changed since you opened it. Refresh before saving."
	case errors.Is(err, proxyconfig.ErrMissing):
		code = "config_missing"
		status, message = 409, "Create the proxy config first."
	case errors.Is(err, runner.ErrProxyConfigRunning), errors.Is(err, runner.ErrServerProcessAlive):
		code = "server_running"
		status, message = 409, "Stop the proxy before saving its settings."
	case errors.Is(err, runner.ErrServerOperation):
		code = "operation_busy"
		status, message = 409, "Another server action is in progress. Try again when it finishes."
	case code == "file_unavailable":
		status, message = 409, "The config file cannot be accessed safely. Ask the administrator to check file permissions, size and server path links, then retry."
	default:
		var pathError *os.PathError
		if errors.As(err, &pathError) {
			code = "file_unavailable"
			status = http.StatusConflict
			if user := getUserFromContext(r); user == nil || user.Role != "admin" {
				message = "The configuration file is unavailable. Ask the administrator to check access and retry."
			}
		}
	}
	writeJSON(w, status, map[string]string{"error": message, "code": code})
}

func decodeProxyConfig(w http.ResponseWriter, r *http.Request, dst any) bool {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 256<<10))
	if err != nil {
		writeError(w, 413, "Proxy settings exceed the request limit")
		return false
	}
	if string(bytes.TrimSpace(body)) == "null" {
		writeError(w, 400, "Proxy settings must be an object")
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeError(w, 400, "Invalid proxy settings")
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		writeError(w, 400, "Expected one proxy settings object")
		return false
	}
	return true
}

func (h *ServerHandler) GetProxyConfig(w http.ResponseWriter, r *http.Request) {
	srv := h.proxyServer(w, r)
	if srv == nil {
		return
	}
	snapshot, err := proxyconfig.Read(h.manager.GetWorkingDir(srv.ID), srv.Type, srv.Port)
	if err != nil {
		writeProxyConfigError(w, r, err)
		return
	}
	writeJSON(w, 200, h.proxyView(r, srv, snapshot))
}

func (h *ServerHandler) InitializeProxyConfig(w http.ResponseWriter, r *http.Request) {
	if h.proxyServer(w, r) == nil {
		return
	}
	if !decodeProxyConfig(w, r, &struct{}{}) {
		return
	}
	release, err := h.manager.BeginProxyConfigOperation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeProxyConfigError(w, r, err)
		return
	}
	defer release()
	srv := h.proxyServer(w, r)
	if srv == nil {
		return
	}
	created, err := proxyconfig.Ensure(h.manager.GetWorkingDir(srv.ID), srv.Type, srv.Port)
	if err != nil {
		writeProxyConfigError(w, r, err)
		return
	}
	snapshot, err := proxyconfig.Read(h.manager.GetWorkingDir(srv.ID), srv.Type, srv.Port)
	if err != nil {
		writeProxyConfigError(w, r, err)
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
		logAudit(h.db, r, "proxy_config_create", fmt.Sprintf("Created %s for %s", snapshot.File, srv.ID))
	}
	writeJSON(w, status, h.proxyView(r, srv, snapshot))
}

func (h *ServerHandler) SaveProxyConfig(w http.ResponseWriter, r *http.Request) {
	if h.proxyServer(w, r) == nil {
		return
	}
	var request struct {
		Revision       string                `json:"revision"`
		Config         *proxyconfig.Settings `json:"config"`
		ConfirmOffline bool                  `json:"confirm_offline"`
	}
	if !decodeProxyConfig(w, r, &request) {
		return
	}
	if request.Config == nil || request.Revision == "" {
		writeError(w, 400, "Config and revision are required")
		return
	}
	release, err := h.manager.BeginProxyConfigOperation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeProxyConfigError(w, r, err)
		return
	}
	defer release()
	srv := h.proxyServer(w, r)
	if srv == nil {
		return
	}
	dir := h.manager.GetWorkingDir(srv.ID)
	before, err := proxyconfig.Read(dir, srv.Type, srv.Port)
	if err != nil {
		writeProxyConfigError(w, r, err)
		return
	}
	if before.Config.OnlineMode && !request.Config.OnlineMode && !request.ConfirmOffline {
		writeError(w, 400, "Confirm the risk before disabling player authentication")
		return
	}
	snapshot, err := proxyconfig.Save(dir, srv.Type, srv.Port, request.Revision, *request.Config)
	if err != nil {
		writeProxyConfigError(w, r, err)
		return
	}
	logAudit(h.db, r, "proxy_config_save", fmt.Sprintf("Saved %s for %s (%d backend servers)", snapshot.File, srv.ID, len(snapshot.Config.Backends)))
	writeJSON(w, 200, h.proxyView(r, srv, snapshot))
}

func (h *ServerHandler) RevealProxySecret(w http.ResponseWriter, r *http.Request) {
	if !decodeProxyConfig(w, r, &struct{}{}) {
		return
	}
	srv := h.proxyServer(w, r)
	if srv == nil {
		return
	}
	secret, err := proxyconfig.ReadSecret(h.manager.GetWorkingDir(srv.ID), srv.Type)
	if err != nil {
		writeProxyConfigError(w, r, err)
		return
	}
	logAudit(h.db, r, "proxy_secret_view", fmt.Sprintf("Viewed forwarding secret for %s", srv.ID))
	w.Header().Set("Cache-Control", "private, no-store")
	writeJSON(w, 200, map[string]string{"secret": secret})
}

// A custom forwarding-secret file must survive a config-preserving reinstall too.
// Keep its top-level directory if the path is nested; never inspect outside root.
func proxyConfigPreservedNames(root *rooted.Root, velocity bool) (map[string]bool, error) {
	names := map[string]bool{"velocity.toml": true, "config.yml": true, "waterfall.yml": true, "forwarding.secret": true}
	if !velocity {
		return names, nil
	}
	f, err := root.Open("velocity.toml")
	if os.IsNotExist(err) {
		return names, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := rooted.ReadBounded(f, 1<<20)
	if err != nil {
		return nil, err
	}
	var document map[string]any
	if err := toml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("Cannot preserve an invalid velocity.toml; fix it in Files first")
	}
	value, exists := document["forwarding-secret-file"]
	if !exists {
		return names, nil
	}
	name, ok := value.(string)
	if !ok || !filepath.IsLocal(name) || strings.ContainsAny(name, "\\:\x00") || strings.Contains("/"+name+"/", "/../") {
		return nil, fmt.Errorf("The forwarding secret must use a relative path inside the server files")
	}
	if err := root.Check(name); err != nil {
		return nil, err
	}
	top := strings.Split(filepath.ToSlash(filepath.Clean(name)), "/")[0]
	names[top] = true
	names[strings.ToLower(top)] = true
	return names, nil
}
