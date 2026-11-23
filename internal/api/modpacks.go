package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/downloader"
	"github.com/Pekly/NeoPanel/internal/runner"
)

type ModpackHandler struct {
	db      *database.DB
	manager *runner.ServerManager
}

func NewModpackHandler(db *database.DB, manager *runner.ServerManager) *ModpackHandler {
	return &ModpackHandler{
		db:      db,
		manager: manager,
	}
}

func (h *ModpackHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("query")
	loader := r.URL.Query().Get("loader")
	version := r.URL.Query().Get("version")
	sortBy := r.URL.Query().Get("sort")
	limit, offset, err := pluginSearchPage(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	res, err := downloader.SearchModrinthPluginsPage(r.Context(), q, limit, offset, loader, version, sortBy, "modpack")
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, res)
}

func (h *ModpackHandler) GetInstalled(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	if _, err := h.db.GetServer(serverID); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	workingDir := h.manager.GetWorkingDir(serverID)
	metaPath := filepath.Join(workingDir, ".neopanel_modpack.json")

	raw, err := os.ReadFile(metaPath)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"installed": false,
		})
		return
	}

	var meta downloader.InstalledModpackMetadata
	if err := json.Unmarshal(raw, &meta); err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"installed": false,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"installed": true,
		"metadata":  meta,
	})
}

func (h *ModpackHandler) Install(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	srv, err := h.db.GetServer(serverID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	var req struct {
		ProjectID string `json:"project_id"`
		Slug      string `json:"slug"`
		VersionID string `json:"version_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON body"})
		return
	}

	identifier := strings.TrimSpace(req.Slug)
	if identifier == "" {
		identifier = strings.TrimSpace(req.ProjectID)
	}
	if identifier == "" || !downloader.ValidModrinthIdentifier(identifier) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Valid modpack identifier is required"})
		return
	}

	// 1. Check server status - server MUST be offline to install modpack safely
	inst, ok := h.manager.GetInstance(serverID)
	if ok && inst.GetStatus() != "offline" {
		writeJSON(w, http.StatusConflict, map[string]string{
			"error": "Server must be stopped completely before installing a modpack. Stop the server and try again.",
		})
		return
	}

	// 2. Fetch project versions to find target .mrpack
	versions, err := downloader.GetModrinthProjectVersions(r.Context(), identifier)
	if err != nil || len(versions) == 0 {
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"error": "Failed to retrieve modpack versions from Modrinth: " + err.Error(),
		})
		return
	}

	var targetVer *downloader.ModrinthVersion
	if req.VersionID != "" {
		for i := range versions {
			if versions[i].ID == req.VersionID || versions[i].VersionNumber == req.VersionID {
				targetVer = &versions[i]
				break
			}
		}
	}
	if targetVer == nil {
		targetVer = &versions[0]
	}

	// Find the .mrpack file in targetVer.Files
	mrpackURL := ""
	for _, f := range targetVer.Files {
		if strings.HasSuffix(strings.ToLower(f.Filename), ".mrpack") || f.Primary {
			mrpackURL = f.URL
			break
		}
	}
	if mrpackURL == "" && len(targetVer.Files) > 0 {
		mrpackURL = targetVer.Files[0].URL
	}
	if mrpackURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Selected modpack version does not contain an installable .mrpack archive"})
		return
	}

	// 3. Download the .mrpack to a temp file
	tmpDir := os.TempDir()
	tmpMrpack := filepath.Join(tmpDir, fmt.Sprintf("neopanel_modpack_%s_%d.mrpack", serverID, time.Now().UnixNano()))
	defer os.Remove(tmpMrpack)

	dlReq, err := http.NewRequestWithContext(r.Context(), http.MethodGet, mrpackURL, nil)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	dlReq.Header.Set("User-Agent", "NeoPanel-Modpack-Installer/1.0")

	client := &http.Client{Timeout: 90 * time.Second}
	dlResp, err := client.Do(dlReq)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Failed to download modpack archive: " + err.Error()})
		return
	}
	defer dlResp.Body.Close()

	if dlResp.StatusCode != http.StatusOK {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": fmt.Sprintf("Upstream returned HTTP %d while downloading modpack", dlResp.StatusCode)})
		return
	}

	tmpFile, err := os.Create(tmpMrpack)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	_, copyErr := io.Copy(tmpFile, io.LimitReader(dlResp.Body, 500<<20)) // 500 MB max archive size
	tmpFile.Close()
	if copyErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save modpack archive: " + copyErr.Error()})
		return
	}

	// 4. Install into server directory
	workingDir := h.manager.GetWorkingDir(serverID)
	result, err := downloader.InstallModpackArchive(r.Context(), tmpMrpack, workingDir, client)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Installation failed: " + err.Error()})
		return
	}

	// 5. Update server software type in DB if different (e.g. set to fabric or forge or neoforge)
	if result.Loader != "" && srv.Type != result.Loader {
		srv.Type = result.Loader
		_ = h.db.UpdateServer(srv)
	}

	logAudit(h.db, r, "modpack_install", fmt.Sprintf("Installed modpack '%s' (%s, %d mods) on %s", result.Name, result.Loader, result.ModsInstalled, auditServerLabel(h.db, srv.ID)))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": result.Summary,
		"result":  result,
	})
}

func (h *ModpackHandler) DeleteInstalled(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	if _, err := h.db.GetServer(serverID); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	workingDir := h.manager.GetWorkingDir(serverID)
	metaPath := filepath.Join(workingDir, ".neopanel_modpack.json")
	_ = os.Remove(metaPath)

	logAudit(h.db, r, "modpack_unlink", fmt.Sprintf("Unlinked modpack metadata on %s", auditServerLabel(h.db, serverID)))

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Modpack link removed",
	})
}
