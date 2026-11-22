package api

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/downloader"
	"github.com/Pekly/NeoPanel/internal/proxyconfig"
	"github.com/Pekly/NeoPanel/internal/rooted"
	"github.com/Pekly/NeoPanel/internal/runner"
	"github.com/Pekly/NeoPanel/internal/serverjar"
)

type SoftwareHandler struct {
	db          *database.DB
	manager     *runner.ServerManager
	download    func(context.Context, string, string, string, string) (serverjar.Info, error)
	beginChange func(context.Context, string) (context.Context, func(), error)
}

func NewSoftwareHandler(db *database.DB, manager *runner.ServerManager) *SoftwareHandler {
	return &SoftwareHandler{
		db:       db,
		manager:  manager,
		download: downloader.DownloadServerJarWithInfo,
		// This stopped-process gate is engine-neutral, despite its legacy name.
		// It retains the canonical root lease/runlock and joins shutdown.
		beginChange: manager.BeginPluginUpdateOperation,
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
	ctx, release, err := h.beginChange(r.Context(), serverID)
	if err != nil {
		writeError(w, http.StatusConflict, "Stop the server and wait for active file or power operations before changing software")
		return
	}
	defer release()
	r = r.WithContext(ctx)
	srv, err = h.db.GetServer(serverID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Server not found")
		return
	}

	serverDir := h.manager.GetWorkingDir(serverID)
	jarName := fmt.Sprintf("%s-%s.jar", req.Software, req.Version)
	if downloader.IsNativeSoftware(req.Software) {
		jarName = downloader.ServerFileName(req.Software)
	}

	// The downloader validates and publishes only approved distribution files.
	info, err := h.download(ctx, req.Software, req.Version, serverDir, jarName)
	if err != nil {
		writeError(w, http.StatusBadGateway, "Server software download failed; retry after checking the selected release")
		return
	}
	if downloader.IsNativeSoftware(req.Software) {
		if err := runner.SyncSteelPort(serverDir, srv.Port); err != nil {
			writeError(w, http.StatusInternalServerError, "Software installed, but native server port configuration could not be updated")
			return
		}
	}
	if proxyconfig.Supported(req.Software) {
		if _, err := proxyconfig.Ensure(serverDir, req.Software, srv.Port); err != nil {
			writeError(w, http.StatusConflict, "Software installed, but proxy configuration could not be prepared safely")
			return
		}
		if err := proxyconfig.SyncPort(serverDir, req.Software, srv.Port); err != nil {
			writeError(w, http.StatusConflict, "Software installed, but proxy port configuration could not be updated safely")
			return
		}
	}

	// Update DB record
	oldNative := downloader.IsNativeSoftware(srv.Type)
	srv.Type = req.Software
	srv.Version = req.Version
	srv.JarFile = jarName
	if downloader.IsNativeSoftware(req.Software) {
		srv.MemoryMinMB, srv.JavaArgs, srv.JavaPath = 0, "", ""
	} else if oldNative {
		srv.JavaPath = "auto"
	}
	if err := h.db.UpdateServer(srv); err != nil {
		writeError(w, http.StatusInternalServerError, "Software files installed, but the server record could not be updated")
		return
	}
	h.manager.UpdateInstanceServer(srv)
	if err := h.manager.RecordArtifactInfo(srv, info); err != nil {
		writeError(w, http.StatusInternalServerError, "Software installed, but artifact metadata could not be verified and saved; refresh artifact information before starting")
		return
	}

	logAudit(h.db, r, "software_install", fmt.Sprintf("Installed %s %s (%s) on %s (%s)", req.Software, req.Version, jarName, srv.Name, serverID))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":   fmt.Sprintf("Installed %s %s successfully", req.Software, req.Version),
		"software":  req.Software,
		"version":   req.Version,
		"jar_file":  jarName,
		"java_path": srv.JavaPath,
		"artifact":  h.manager.ArtifactSummary(srv),
	})
}

func (h *SoftwareHandler) GetProperties(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	if rejectUnsupportedJavaConfig(h.db, serverID, w) {
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
	if rejectUnsupportedJavaConfig(h.db, serverID, w) {
		return
	}
	release, err := rooted.TryMutation(h.manager.GetWorkingDir(serverID))
	if err != nil {
		writeError(w, http.StatusConflict, "Files are busy with another server operation; retry when it finishes")
		return
	}
	defer release()
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
		status := http.StatusInternalServerError
		if errors.Is(err, runner.ErrServerOperation) || errors.Is(err, runner.ErrRestoreRecovery) || errors.Is(err, runner.ErrServerProcessAlive) {
			status = http.StatusConflict
		}
		writeError(w, status, "Cannot enter reinstall maintenance; stop the server and wait for active operations before retrying")
		return
	}
	defer release()
	srv, err = h.db.GetServer(serverID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Server not found")
		return
	}

	serverDir := h.manager.GetWorkingDir(serverID)
	root, err := rooted.Open(serverDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Cannot open the server installation directory")
		return
	}
	defer root.Close()
	// Stage every distribution before cleanup. Failed downloads/validation leave
	// the existing installation untouched; publication itself is not crash-atomic.
	jarName := downloader.ServerFileName(req.Software)
	staged, stageRoot, err := root.TempDir()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Cannot prepare a private software staging directory")
		return
	}
	defer func() { stageRoot.Close(); _ = root.RemoveAll(staged) }()
	stageDir := filepath.Join(serverDir, staged)
	info, err := h.download(r.Context(), req.Software, req.Version, stageDir, jarName)
	if err != nil {
		writeError(w, http.StatusBadGateway, "Server software download failed; existing installation files were preserved")
		return
	}
	if downloader.IsNativeSoftware(req.Software) {
		if err := runner.SyncSteelPort(stageDir, srv.Port); err != nil {
			writeError(w, http.StatusInternalServerError, "Cannot prepare native server port configuration; existing files were preserved")
			return
		}
	}
	if err := validateSoftwareStage(r.Context(), stageRoot, stageDir, req.Software, jarName, info); err != nil {
		writeError(w, http.StatusBadGateway, "Downloaded distribution could not be validated; existing installation files were preserved")
		return
	}

	preservedProxyFiles := map[string]bool{}
	if req.PreserveConfig {
		preservedProxyFiles, err = proxyConfigPreservedNames(root, strings.EqualFold(srv.Type, "velocity") || req.Software == "velocity")
		if err != nil {
			writeError(w, http.StatusConflict, "Cannot safely determine which proxy configuration files to preserve; existing files were preserved")
			return
		}
	}

	if r.Context().Err() != nil {
		writeError(w, http.StatusRequestTimeout, "Reinstall canceled before cleanup; existing installation files were preserved")
		return
	}
	// 2. Perform selective cleanup based on preservation flags
	entries, err := root.ReadDir(".")
	if err != nil && !os.IsNotExist(err) {
		writeError(w, http.StatusInternalServerError, "Cannot list the server installation directory for cleanup")
		return
	}
	if err == nil {
		for _, entry := range entries {
			if r.Context().Err() != nil {
				writeError(w, http.StatusRequestTimeout, "Reinstall canceled during cleanup; installation may be incomplete")
				return
			}
			name := entry.Name()
			if name == staged {
				continue
			}
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
			if req.PreserveConfig && (preservedProxyFiles[name] || preservedProxyFiles[lower] || lower == "server.properties" || lower == "config" || lower == "bukkit.yml" || lower == "spigot.yml" || lower == "paper.yml") {
				continue
			}

			// Remove all other runtime files, old jars, cache directories
			if err := root.RemoveAll(name); err != nil {
				writeError(w, http.StatusInternalServerError, "Reinstall cleanup failed; installation may be incomplete and must be retried before starting")
				return
			}
		}
	}

	// Move only the validated, downloader-approved outputs; no provider scripts
	// or generic archive extraction is performed by the API.
	if err := publishSoftwareStage(r.Context(), root, stageRoot, staged, req.Software, jarName); err != nil {
		writeError(w, http.StatusInternalServerError, "Reinstall could not publish all software files; installation may be incomplete and must be retried before starting")
		return
	}
	if downloader.IsNativeSoftware(req.Software) {
		if err := runner.SyncSteelPort(serverDir, srv.Port); err != nil {
			writeError(w, http.StatusInternalServerError, "Software installed, but native server port configuration could not be updated")
			return
		}
	}

	// Proxies use their own config; world servers retain server.properties.
	if proxyconfig.Supported(req.Software) {
		if _, err := proxyconfig.Ensure(serverDir, req.Software, srv.Port); err != nil {
			writeError(w, http.StatusConflict, "Software installed, but proxy configuration could not be prepared safely")
			return
		}
		if err := proxyconfig.SyncPort(serverDir, req.Software, srv.Port); err != nil {
			writeError(w, http.StatusConflict, "Software installed, but proxy port configuration could not be updated safely")
			return
		}
	} else if !downloader.IsNativeSoftware(req.Software) {
		if _, err := root.Lstat("server.properties"); os.IsNotExist(err) {
			propsContent := fmt.Sprintf("# Minecraft server properties generated by NeoPanel\nserver-port=%d\nquery.port=%d\nmotd=A NeoPanel Minecraft Server\nenable-rcon=false\n", srv.Port, srv.Port)
			if err := root.WriteFile("server.properties", []byte(propsContent), 0644); err != nil {
				writeError(w, http.StatusInternalServerError, "Software installed, but server.properties could not be created")
				return
			}
		}
	}

	// 5. Update database record
	oldNative := downloader.IsNativeSoftware(srv.Type)
	srv.Type = req.Software
	srv.Version = req.Version
	srv.JarFile = jarName
	if downloader.IsNativeSoftware(req.Software) {
		srv.MemoryMinMB, srv.JavaArgs, srv.JavaPath = 0, "", ""
	} else if oldNative {
		srv.JavaPath = "auto"
	}
	if err := h.db.UpdateServer(srv); err != nil {
		writeError(w, http.StatusInternalServerError, "Software files installed, but the server record could not be updated")
		return
	}
	h.manager.UpdateInstanceServer(srv)
	if err := h.manager.RecordArtifactInfo(srv, info); err != nil {
		writeError(w, http.StatusInternalServerError, "Software installed, but artifact metadata could not be verified and saved; refresh artifact information before starting")
		return
	}

	logAudit(h.db, r, "server_reinstall", fmt.Sprintf("Reinstalled %s %s on %s (%s)", req.Software, req.Version, srv.Name, serverID))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":   fmt.Sprintf("Server reinstalled successfully to %s %s", req.Software, req.Version),
		"software":  req.Software,
		"version":   req.Version,
		"jar_file":  jarName,
		"java_path": srv.JavaPath,
		"artifact":  h.manager.ArtifactSummary(srv),
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

func rejectUnsupportedJavaConfig(db *database.DB, serverID string, w http.ResponseWriter) bool {
	srv, err := db.GetServer(serverID)
	if err == nil && proxyconfig.Supported(srv.Type) {
		writeError(w, http.StatusBadRequest, "This proxy uses "+proxyconfig.FileName(srv.Type)+". Open Proxy settings to edit it.")
		return true
	}
	if err == nil && downloader.IsNativeSoftware(srv.Type) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Steel uses config/config.toml, not Java server.properties or eula.txt. Edit its configuration in Files."})
		return true
	}
	return false
}
