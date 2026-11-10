package api

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
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
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func (h *BackupHandler) List(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	dir := h.getBackupsDir(serverID)

	entries, err := os.ReadDir(dir)
	if err != nil {
		writeJSON(w, http.StatusOK, []BackupItem{})
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
		baseName = fmt.Sprintf("backup-%s-%s", srv.Name, time.Now().Format("2006-01-02-150405"))
	}
	// Sanitize name
	baseName = strings.ReplaceAll(baseName, " ", "-")
	baseName = strings.ReplaceAll(baseName, "/", "_")
	baseName = strings.ReplaceAll(baseName, "\\", "_")
	if !strings.HasSuffix(baseName, ".tar.gz") {
		baseName += ".tar.gz"
	}

	backupsDir := h.getBackupsDir(serverID)
	destPath := filepath.Join(backupsDir, baseName)
	serverDir := h.manager.GetWorkingDir(serverID)

	// Create tar.gz archive
	outFile, err := os.Create(destPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create backup file: " + err.Error()})
		return
	}
	defer outFile.Close()

	gw := gzip.NewWriter(outFile)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	err = filepath.Walk(serverDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(serverDir, path)
		if err != nil || rel == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return nil
		}
		header.Name = filepath.ToSlash(rel)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return nil
			}
			defer f.Close()
			_, _ = io.Copy(tw, f)
		}
		return nil
	})

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Archive creation error: " + err.Error()})
		return
	}

	fi, _ := outFile.Stat()
	size := int64(0)
	if fi != nil {
		size = fi.Size()
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
