package api

import (
	"encoding/json"
	"fmt"
	"github.com/Pekly/NeoPanel/internal/rooted"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/downloader"
	"github.com/Pekly/NeoPanel/internal/runner"
)

type PluginHandler struct {
	db      *database.DB
	manager *runner.ServerManager
	mu      sync.Mutex
}

func NewPluginHandler(db *database.DB, manager *runner.ServerManager) *PluginHandler {
	return &PluginHandler{
		db:      db,
		manager: manager,
	}
}

type InstalledPlugin struct {
	Filename string    `json:"filename"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
	Enabled  bool      `json:"enabled"`
}

func (h *PluginHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("query")
	loader := r.URL.Query().Get("loader")
	version := r.URL.Query().Get("version")
	sortBy := r.URL.Query().Get("sort")
	projType := r.URL.Query().Get("type")
	limitStr := r.URL.Query().Get("limit")
	limit := 24
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	res, err := downloader.SearchModrinthPlugins(q, limit, loader, version, sortBy, projType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, res)
}

func (h *PluginHandler) ListInstalled(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	srv, err := h.db.GetServer(serverID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}
	if rejectSteelPlugins(srv, w) {
		return
	}

	isMod := downloader.IsModdedSoftware(srv.Type)
	folderName := "plugins"
	kind := "plugin"
	if isMod {
		folderName = "mods"
		kind = "mod"
	}

	pluginsDir, err := h.pluginDir(serverID, folderName)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	defer pluginsDir.Close()

	entries, err := pluginsDir.ReadDir(".")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to read %s directory", folderName)})
		return
	}

	var list []InstalledPlugin
	for _, entry := range entries {
		if !entry.IsDir() && entry.Type()&os.ModeSymlink == 0 && validInstalledPluginName(entry.Name()) {
			info, err := entry.Info()
			if err == nil {
				list = append(list, InstalledPlugin{
					Filename: entry.Name(),
					Size:     info.Size(),
					Modified: info.ModTime(),
					Enabled:  !strings.HasSuffix(strings.ToLower(entry.Name()), ".disabled"),
				})
			}
		}
	}
	if list == nil {
		list = []InstalledPlugin{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"folder":  folderName,
		"kind":    kind,
		"plugins": list,
	})
}

func (h *PluginHandler) Install(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	srv, err := h.db.GetServer(serverID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}
	if rejectSteelPlugins(srv, w) {
		return
	}

	var req struct {
		Slug      string `json:"slug"`
		ProjectID string `json:"project_id"`
		ID        string `json:"id"`
		VersionID string `json:"version_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = strings.TrimSpace(req.ProjectID)
	}
	if slug == "" {
		slug = strings.TrimSpace(req.ID)
	}
	if slug == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Valid project slug is required"})
		return
	}

	isMod := downloader.IsModdedSoftware(srv.Type)
	folderName := "plugins"
	kind := "plugin"
	noun := "Plugin"
	auditEvent := "plugin_install"
	if isMod {
		folderName = "mods"
		kind = "mod"
		noun = "Mod"
		auditEvent = "mod_install"
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	pluginsDir, err := h.pluginDir(serverID, folderName)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	defer pluginsDir.Close()
	filename, err := downloader.DownloadModrinthPluginRoot(slug, pluginsDir, srv.Type, srv.Version, req.VersionID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	logAudit(h.db, r, auditEvent, fmt.Sprintf("Installed %s '%s' (%s) on %s (%s)", kind, filename, slug, srv.Name, serverID))

	writeJSON(w, http.StatusOK, map[string]string{
		"message":  noun + " installed successfully. Restart the server to apply.",
		"filename": filename,
		"folder":   folderName,
		"kind":     kind,
	})
}

func (h *PluginHandler) Delete(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	srv, err := h.db.GetServer(serverID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	filename := r.URL.Query().Get("filename")
	if !validInstalledPluginName(filename) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid filename"})
		return
	}

	isMod := downloader.IsModdedSoftware(srv.Type)
	folderName := "plugins"
	kind := "plugin"
	noun := "Plugin"
	auditEvent := "plugin_delete"
	if isMod {
		folderName = "mods"
		kind = "mod"
		noun = "Mod"
		auditEvent = "mod_delete"
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	dir, err := h.pluginDir(serverID, folderName)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	defer dir.Close()
	if err := dir.Remove(filename); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to remove %s: %s", kind, err.Error())})
		return
	}

	logAudit(h.db, r, auditEvent, fmt.Sprintf("Uninstalled %s '%s' from %s (%s)", kind, filename, srv.Name, serverID))

	writeJSON(w, http.StatusOK, map[string]string{
		"message": noun + " removed successfully",
		"folder":  folderName,
		"kind":    kind,
	})
}

func validInstalledPluginName(name string) bool {
	low := strings.ToLower(name)
	return name != "" && !strings.ContainsAny(name, "/\\:") && !strings.Contains(name, "..") && (strings.HasSuffix(low, ".jar") || strings.HasSuffix(low, ".jar.disabled"))
}

func (h *PluginHandler) pluginDir(serverID, folder string) (*rooted.Root, error) {
	root, err := rooted.Open(h.manager.GetWorkingDir(serverID))
	if err != nil {
		return nil, err
	}
	defer root.Close()
	if err := root.Check(folder); err != nil {
		return nil, err
	}
	if err := root.MkdirAll(folder, 0755); err != nil {
		return nil, err
	}
	return root.Sub(folder)
}

func setPluginEnabled(dir, filename string, enabled bool) (string, error) {
	root, err := rooted.Open(dir)
	if err != nil {
		return "", err
	}
	defer root.Close()
	return setPluginEnabledRoot(root, filename, enabled)
}

func setPluginEnabledRoot(dir *rooted.Root, filename string, enabled bool) (string, error) {
	if !validInstalledPluginName(filename) {
		return "", fmt.Errorf("invalid plugin filename")
	}
	src := filename
	info, err := dir.Lstat(src)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("plugin must be a regular file")
	}
	wasEnabled := !strings.HasSuffix(strings.ToLower(filename), ".disabled")
	if wasEnabled == enabled {
		return filename, nil
	}
	destName := filename + ".disabled"
	if enabled {
		destName = filename[:len(filename)-len(".disabled")]
	}
	dest := destName
	if err := dir.Link(src, dest); err != nil {
		return "", fmt.Errorf("cannot change plugin state (destination may exist): %w", err)
	}
	if err := dir.Remove(src); err != nil {
		_ = dir.Remove(dest)
		return "", err
	}
	return destName, nil
}

func (h *PluginHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	srv, err := h.db.GetServer(serverID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}
	var req struct {
		Filename string `json:"filename"`
		Enabled  *bool  `json:"enabled"`
	}
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.Enabled == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "filename and enabled are required"})
		return
	}
	folder := "plugins"
	if downloader.IsModdedSoftware(srv.Type) {
		folder = "mods"
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	dir, err := h.pluginDir(serverID, folder)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	defer dir.Close()
	filename, err := setPluginEnabledRoot(dir, req.Filename, *req.Enabled)
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	logAudit(h.db, r, "plugin_toggle", fmt.Sprintf("Set %s enabled=%t on %s", filename, *req.Enabled, serverID))
	writeJSON(w, http.StatusOK, map[string]interface{}{"filename": filename, "enabled": *req.Enabled, "restart_required": true})
}

func (h *PluginHandler) Versions(w http.ResponseWriter, r *http.Request) {
	srv, err := h.db.GetServer(getURLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}
	if rejectSteelPlugins(srv, w) {
		return
	}
	versions, err := downloader.CompatibleModrinthVersions(r.URL.Query().Get("slug"), srv.Type, srv.Version)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"versions": versions, "minecraft_version": srv.Version, "loader": srv.Type})
}

func rejectSteelPlugins(srv *database.Server, w http.ResponseWriter) bool {
	if downloader.IsNativeSoftware(srv.Type) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Steel does not support Java plugins or mods from this catalog."})
		return true
	}
	return false
}
