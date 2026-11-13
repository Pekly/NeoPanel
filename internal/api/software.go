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
	h.manager.UpdateInstanceServer(srv)

	logAudit(h.db, r, "software_install", fmt.Sprintf("Installed %s %s (%s) on %s (%s)", req.Software, req.Version, jarName, srv.Name, serverID))

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

	logAudit(h.db, r, "properties_save", fmt.Sprintf("Saved server.properties (%d settings) on %s", len(newProps), auditServerLabel(h.db, serverID)))

	writeJSON(w, http.StatusOK, map[string]string{"message": "server.properties saved successfully"})
}

type ReinstallRequest struct {
	Software        string `json:"software"` // paper, purpur, leaf, folia, fabric, vanilla, velocity
	Version         string `json:"version"`  // 1.21.4, latest, etc.
	PreserveWorld   bool   `json:"preserve_world"`
	PreserveConfig  bool   `json:"preserve_config"`
	PreservePlugins bool   `json:"preserve_plugins"`
	WipeFiles       bool   `json:"wipe_files"` // true = wipe all files before reinstall (incompatible egg change)
}

func (h *SoftwareHandler) ReinstallServer(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	srv, err := h.db.GetServer(serverID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	var req ReinstallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid reinstall request"})
		return
	}

	req.Software = strings.ToLower(strings.TrimSpace(req.Software))
	if req.Software == "" {
		req.Software = srv.Type
		if req.Software == "" {
			req.Software = "paper"
		}
	}
	req.Version = strings.TrimSpace(req.Version)
	if req.Version == "" {
		req.Version = srv.Version
		if req.Version == "" {
			req.Version = "latest"
		}
	}

	// 1. Stop server if active
	_ = h.manager.Stop(srv.ID, false)

	serverDir := h.manager.GetWorkingDir(serverID)

	// 2. Perform selective cleanup based on preservation flags
	entries, err := os.ReadDir(serverDir)
	if err == nil {
		for _, entry := range entries {
			name := entry.Name()
			lower := strings.ToLower(name)

			// Keep worlds?
			if req.PreserveWorld && (lower == "world" || lower == "world_nether" || lower == "world_the_end") {
				continue
			}

			// Keep plugins or mods?
			if req.PreservePlugins && (lower == "plugins" || lower == "mods") {
				continue
			}

			// Keep config?
			if req.PreserveConfig && (lower == "server.properties" || lower == "config" || lower == "bukkit.yml" || lower == "spigot.yml" || lower == "paper.yml") {
				continue
			}

			// Remove all other runtime files, old jars, cache directories
			_ = os.RemoveAll(filepath.Join(serverDir, name))
		}
	}

	// 3. Download fresh server JAR
	jarName := "server.jar"
	if err := downloader.DownloadServerJar(req.Software, req.Version, serverDir, jarName); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Reinstall failed to download jar: " + err.Error()})
		return
	}

	// 4. Ensure server.properties exists with port
	propsPath := filepath.Join(serverDir, "server.properties")
	if _, err := os.Stat(propsPath); os.IsNotExist(err) {
		propsContent := fmt.Sprintf("# Minecraft server properties generated by NeoPanel\nserver-port=%d\nquery.port=%d\nmotd=%s\nenable-rcon=false\n", srv.Port, srv.Port, srv.Name)
		_ = os.WriteFile(propsPath, []byte(propsContent), 0644)
	}

	// 5. Update database record
	srv.Type = req.Software
	srv.Version = req.Version
	srv.JarFile = jarName
	if err := h.db.UpdateServer(srv); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update server record: " + err.Error()})
		return
	}
	h.manager.UpdateInstanceServer(srv)

	logAudit(h.db, r, "server_reinstall", fmt.Sprintf("Reinstalled %s %s on %s (%s)", req.Software, req.Version, srv.Name, serverID))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":  fmt.Sprintf("Server reinstalled successfully to %s %s", req.Software, req.Version),
		"software": req.Software,
		"version":  req.Version,
		"jar_file": jarName,
	})
}
