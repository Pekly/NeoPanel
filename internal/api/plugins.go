package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/downloader"
	"github.com/Pekly/NeoPanel/internal/runner"
)

type PluginHandler struct {
	db      *database.DB
	manager *runner.ServerManager
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

	isMod := downloader.IsModdedSoftware(srv.Type)
	folderName := "plugins"
	kind := "plugin"
	if isMod {
		folderName = "mods"
		kind = "mod"
	}

	pluginsDir := filepath.Join(h.manager.GetWorkingDir(serverID), folderName)
	_ = os.MkdirAll(pluginsDir, 0755)

	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to read %s directory", folderName)})
		return
	}

	var list []InstalledPlugin
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".jar") {
			info, err := entry.Info()
			if err == nil {
				list = append(list, InstalledPlugin{
					Filename: entry.Name(),
					Size:     info.Size(),
					Modified: info.ModTime(),
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

	var req struct {
		Slug      string `json:"slug"`
		ProjectID string `json:"project_id"`
		ID        string `json:"id"`
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
	pluginsDir := filepath.Join(h.manager.GetWorkingDir(serverID), folderName)

	filename, err := downloader.DownloadModrinthPlugin(slug, pluginsDir, srv.Type)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	logAudit(h.db, r, auditEvent, fmt.Sprintf("Installed %s '%s' (%s) on %s (%s)", kind, filename, slug, srv.Name, serverID))

	writeJSON(w, http.StatusOK, map[string]string{
		"message":  noun + " installed successfully",
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
	if filename == "" || strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
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

	filePath := filepath.Join(h.manager.GetWorkingDir(serverID), folderName, filename)
	if err := os.Remove(filePath); err != nil {
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
