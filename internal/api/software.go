package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/downloader"
	"github.com/Pekly/NeoPanel/internal/runner"
)

type SoftwareHandler struct {
	db      *database.DB
	manager *runner.ServerManager
}

func NewSoftwareHandler(db *database.DB, manager *runner.ServerManager) *SoftwareHandler {
	return &SoftwareHandler{
		db:      db,
		manager: manager,
	}
}

func (h *SoftwareHandler) ChangeSoftware(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	srv, err := h.db.GetServer(serverID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	var req struct {
		Software string `json:"software"` // paper, purpur, fabric, vanilla
		Version  string `json:"version"`  // 1.20.4, 1.21.1, etc.
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
		return
	}

	req.Software = strings.ToLower(strings.TrimSpace(req.Software))
	req.Version = strings.TrimSpace(req.Version)
	if req.Software == "" || req.Version == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Software and version are required"})
		return
	}

	serverDir := h.manager.GetWorkingDir(serverID)
	jarName := fmt.Sprintf("%s-%s.jar", req.Software, req.Version)

	// Download jar
	if err := downloader.DownloadServerJar(req.Software, req.Version, serverDir, jarName); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Download failed: " + err.Error()})
		return
	}

	// Update DB record
	srv.Type = req.Software
	srv.Version = req.Version
	srv.JarFile = jarName
	if err := h.db.UpdateServer(srv); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update server record: " + err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":  fmt.Sprintf("Installed %s %s successfully", req.Software, req.Version),
		"software": req.Software,
		"version":  req.Version,
		"jar_file": jarName,
	})
}

func (h *SoftwareHandler) GetProperties(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	propPath := filepath.Join(h.manager.GetWorkingDir(serverID), "server.properties")

	props := make(map[string]string)
	if file, err := os.Open(propPath); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				props[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	writeJSON(w, http.StatusOK, props)
}

func (h *SoftwareHandler) SaveProperties(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	propPath := filepath.Join(h.manager.GetWorkingDir(serverID), "server.properties")

	var newProps map[string]string
	if err := json.NewDecoder(r.Body).Decode(&newProps); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid properties object"})
		return
	}

	var sb strings.Builder
	sb.WriteString("# Minecraft server properties saved by NeoPanel\n")
	for k, v := range newProps {
		sb.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}

	if err := os.WriteFile(propPath, []byte(sb.String()), 0644); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save server.properties: " + err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "server.properties saved successfully"})
}
