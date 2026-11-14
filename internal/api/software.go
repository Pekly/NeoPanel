package api

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Pekly/NeoPanel/internal/rooted"
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
	if err := downloader.ValidateSoftware(req.Software, req.Version); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	release, err := h.manager.BeginServerOperation(serverID)
	if err != nil {
		writeServerOperationError(w, err)
		return
	}
	defer release()

	serverDir := h.manager.GetWorkingDir(serverID)
	jarName := fmt.Sprintf("%s-%s.jar", req.Software, req.Version)
	if downloader.IsNativeSoftware(req.Software) {
		jarName = downloader.ServerFileName(req.Software)
	}

	// Download jar
	if err := downloader.DownloadServerJar(req.Software, req.Version, serverDir, jarName); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Download failed: " + err.Error()})
		return
	}
	if downloader.IsNativeSoftware(req.Software) {
		if err := runner.SyncSteelPort(serverDir, srv.Port); err != nil {
			writeServerOperationError(w, err)
			return
		}
	}

	// Update DB record
	srv.Type = req.Software
	srv.Version = req.Version
	srv.JarFile = jarName
	if downloader.IsNativeSoftware(req.Software) {
		srv.MemoryMinMB, srv.JavaArgs, srv.JavaPath = 0, "", ""
	}
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
	if rejectSteelJavaConfig(h.db, serverID, w) {
		return
	}
	root, err := rooted.Open(h.manager.GetWorkingDir(serverID))
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	defer root.Close()

	props := make(map[string]string)
	file, err := root.Open("server.properties")
	if err != nil && !os.IsNotExist(err) {
		writeError(w, 400, err.Error())
		return
	}
	if err == nil {
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
	if rejectSteelJavaConfig(h.db, serverID, w) {
		return
	}
	root, err := rooted.Open(h.manager.GetWorkingDir(serverID))
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	defer root.Close()

	var newProps map[string]string
	if err := json.NewDecoder(r.Body).Decode(&newProps); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid properties object"})
		return
	}

	var sb strings.Builder
	sb.WriteString("# Minecraft server properties saved by NeoPanel\n")
	for k, v := range newProps {
		if !rooted.PropertyText(k) || !rooted.PropertyText(v) {
			writeError(w, 400, "Property keys and values must not contain line controls")
			return
		}
		sb.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}

	if err := root.WriteFile("server.properties", []byte(sb.String()), 0644); err != nil {
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
	if err := downloader.ValidateSoftware(req.Software, req.Version); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Keep the same gate as restore through stop, cleanup, download and DB update.
	release, err := h.manager.BeginServerMaintenance(r.Context(), serverID)
	if err != nil {
		writeServerOperationError(w, err)
		return
	}
	defer release()

	serverDir := h.manager.GetWorkingDir(serverID)
	root, err := rooted.Open(serverDir)
	if err != nil {
		writeServerOperationError(w, err)
		return
	}
	defer root.Close()
	// Stage native downloads before destructive cleanup: bad tags, unsupported
	// hosts or network failures must leave the existing installation untouched.
	staged := ""
	if downloader.IsNativeSoftware(req.Software) {
		staged, err = os.MkdirTemp(filepath.Dir(serverDir), ".steel-reinstall-*")
		if err != nil {
			writeServerOperationError(w, err)
			return
		}
		defer os.RemoveAll(staged)
		if err := downloader.DownloadServerJar(req.Software, req.Version, staged, downloader.ServerFileName(req.Software)); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Steel reinstall download failed: " + err.Error()})
			return
		}
		if err := runner.SyncSteelPort(staged, srv.Port); err != nil {
			writeServerOperationError(w, err)
			return
		}
	}

	// 2. Perform selective cleanup based on preservation flags
	entries, err := root.ReadDir(".")
	if err != nil && !os.IsNotExist(err) {
		writeServerOperationError(w, err)
		return
	}
	if err == nil {
		for _, entry := range entries {
			name := entry.Name()
			lower := strings.ToLower(name)

			// Keep worlds?
			if req.PreserveWorld && (lower == "world" || lower == "world_nether" || lower == "world_the_end" || (downloader.IsNativeSoftware(srv.Type) && lower == "saves")) {
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
			if err := root.RemoveAll(name); err != nil {
				writeServerOperationError(w, err)
				return
			}
		}
	}

	// Install the staged native executable or download the Java distribution.
	jarName := downloader.ServerFileName(req.Software)
	if staged != "" {
		stagedRoot, err := rooted.Open(staged)
		if err != nil {
			writeServerOperationError(w, err)
			return
		}
		defer stagedRoot.Close()
		binary, err := stagedRoot.Open(jarName)
		if err != nil {
			writeServerOperationError(w, err)
			return
		}
		defer binary.Close()
		if _, err := root.Copy(context.Background(), jarName, binary, 1<<30, 0755); err != nil {
			writeServerOperationError(w, err)
			return
		}
		configPath := filepath.Join("config", "config.toml")
		if _, err := root.Lstat(configPath); os.IsNotExist(err) {
			if err := root.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
				writeServerOperationError(w, err)
				return
			}
			config, err := stagedRoot.ReadFile(configPath)
			if err != nil {
				writeServerOperationError(w, err)
				return
			}
			if err := root.WriteFile(configPath, config, 0644); err != nil {
				writeServerOperationError(w, err)
				return
			}
		}
		if err := runner.SyncSteelPort(serverDir, srv.Port); err != nil {
			writeServerOperationError(w, err)
			return
		}
	} else if err := downloader.DownloadServerJar(req.Software, req.Version, serverDir, jarName); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Reinstall failed to download server: " + err.Error()})
		return
	}

	// 4. Ensure server.properties exists with port
	if !downloader.IsNativeSoftware(req.Software) {
		if _, err := root.Lstat("server.properties"); os.IsNotExist(err) {
			propsContent := fmt.Sprintf("# Minecraft server properties generated by NeoPanel\nserver-port=%d\nquery.port=%d\nmotd=A NeoPanel Minecraft Server\nenable-rcon=false\n", srv.Port, srv.Port)
			if err := root.WriteFile("server.properties", []byte(propsContent), 0644); err != nil {
				writeServerOperationError(w, err)
				return
			}
		}
	}

	// 5. Update database record
	srv.Type = req.Software
	srv.Version = req.Version
	srv.JarFile = jarName
	if downloader.IsNativeSoftware(req.Software) {
		srv.MemoryMinMB, srv.JavaArgs, srv.JavaPath = 0, "", ""
	}
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

func writeServerOperationError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if errors.Is(err, runner.ErrServerOperation) || errors.Is(err, runner.ErrRestoreRecovery) || errors.Is(err, runner.ErrServerProcessAlive) {
		status = http.StatusConflict
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func HandleGetSteelVersions(w http.ResponseWriter, r *http.Request) {
	versions, err := downloader.GetSteelVersions()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, versions)
}

func rejectSteelJavaConfig(db *database.DB, serverID string, w http.ResponseWriter) bool {
	srv, err := db.GetServer(serverID)
	if err == nil && downloader.IsNativeSoftware(srv.Type) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Steel uses config/config.toml, not Java server.properties or eula.txt. Edit its configuration in Files."})
		return true
	}
	return false
}
