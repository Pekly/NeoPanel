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
	limitStr := r.URL.Query().Get("limit")
	limit := 24
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	res, err := downloader.SearchModrinthPlugins(q, limit, loader, version, sortBy)
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

	folderName := "plugins"
	if strings.ToLower(srv.Type) == "fabric" {
		folderName = "mods"
	}

	pluginsDir := filepath.Join(h.manager.GetWorkingDir(serverID), folderName)
	_ = os.MkdirAll(pluginsDir, 0755)

	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to read plugins directory"})
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
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Valid plugin slug is required"})
		return
	}

	folderName := "plugins"
	if strings.ToLower(srv.Type) == "fabric" {
		folderName = "mods"
	}
	pluginsDir := filepath.Join(h.manager.GetWorkingDir(serverID), folderName)

	filename, err := downloader.DownloadModrinthPlugin(slug, pluginsDir, srv.Type)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	logAudit(h.db, r, "plugin_install", fmt.Sprintf("Installed plugin '%s' (%s) on %s (%s)", filename, slug, srv.Name, serverID))

	writeJSON(w, http.StatusOK, map[string]string{
		"message":  "Plugin installed successfully",
		"filename": filename,
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

	folderName := "plugins"
	if strings.ToLower(srv.Type) == "fabric" {
		folderName = "mods"
	}

	filePath := filepath.Join(h.manager.GetWorkingDir(serverID), folderName, filename)
	if err := os.Remove(filePath); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to remove plugin: " + err.Error()})
		return
	}

	logAudit(h.db, r, "plugin_delete", fmt.Sprintf("Uninstalled plugin '%s' from %s (%s)", filename, srv.Name, serverID))

	writeJSON(w, http.StatusOK, map[string]string{"message": "Plugin removed successfully"})
}
