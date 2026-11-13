package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
)

type BackupHandler struct {
	cfg     *config.Config
	db      *database.DB
	manager *runner.ServerManager
}

func NewBackupHandler(cfg *config.Config, db *database.DB, manager *runner.ServerManager) *BackupHandler {
	manager.SetBackupDirectory(filepath.Join(cfg.Paths.DataDir, "backups"))
	return &BackupHandler{
		cfg:     cfg,
		db:      db,
		manager: manager,
	}
}

type BackupItem struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *BackupHandler) getBackupsDir(serverID string) string {
	dir := filepath.Join(h.cfg.Paths.DataDir, "backups", serverID)
	// Creation is handled by CreateBackup; reads must not hide filesystem errors.
	return dir
}

func (h *BackupHandler) List(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	dir := h.getBackupsDir(serverID)

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusOK, []BackupItem{})
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to list backups: "+err.Error())
		}
		return
	}

	var list []BackupItem
	for _, entry := range entries {
		if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".tar.gz") || strings.HasSuffix(entry.Name(), ".zip")) {
			info, err := entry.Info()
			if err == nil {
				list = append(list, BackupItem{
					Name:      entry.Name(),
					Size:      info.Size(),
					CreatedAt: info.ModTime(),
				})
			}
		}
	}

	if list == nil {
		list = []BackupItem{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *BackupHandler) Create(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	srv, err := h.db.GetServer(serverID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	baseName := strings.TrimSpace(req.Name)
	if baseName == "" {
		baseName = fmt.Sprintf("backup-%s-%s-%s", srv.Name, time.Now().UTC().Format("2006-01-02-150405"), randomPasswordHex(4))
	}
	// Sanitize name
	baseName = strings.ReplaceAll(baseName, " ", "-")
	baseName = strings.ReplaceAll(baseName, "/", "_")
	baseName = strings.ReplaceAll(baseName, "\\", "_")
	if !strings.HasSuffix(baseName, ".tar.gz") {
		baseName += ".tar.gz"
	}

	size, err := runner.CreateBackup(h.manager.GetWorkingDir(serverID), h.getBackupsDir(serverID), baseName)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, os.ErrExist) {
			status = http.StatusConflict
		}
		writeError(w, status, "Failed to create backup: "+err.Error())
		return
	}

	logAudit(h.db, r, "backup_create", fmt.Sprintf("Created backup '%s' (%s) for %s (%s)", baseName, formatBytesAudit(size), srv.Name, serverID))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Backup created successfully",
		"name":    baseName,
		"size":    size,
	})
}

func (h *BackupHandler) Delete(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	name := r.URL.Query().Get("name")
	if name == "" || strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid backup name"})
		return
	}

	filePath := filepath.Join(h.getBackupsDir(serverID), name)
	if err := os.Remove(filePath); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete backup: " + err.Error()})
		return
	}

	logAudit(h.db, r, "backup_delete", fmt.Sprintf("Deleted backup '%s' from %s", name, auditServerLabel(h.db, serverID)))

	writeJSON(w, http.StatusOK, map[string]string{"message": "Backup deleted"})
}

func (h *BackupHandler) Download(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	name := r.URL.Query().Get("name")
	if name == "" || strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		http.Error(w, "Invalid backup filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.getBackupsDir(serverID), name)
	fi, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		http.Error(w, "Backup not found", http.StatusNotFound)
		return
	}

	var sizeStr string
	if fi != nil {
		sizeStr = " (" + formatBytesAudit(fi.Size()) + ")"
	}
	logAudit(h.db, r, "backup_download", fmt.Sprintf("Downloaded backup '%s'%s from %s", name, sizeStr, auditServerLabel(h.db, serverID)))

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
	w.Header().Set("Content-Type", "application/gzip")
	http.ServeFile(w, r, filePath)
}
