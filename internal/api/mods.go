package api

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
)

var (
	safeModIDRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,47}$`)
	modAssetCT  = map[string]string{
		".js":   "application/javascript; charset=utf-8",
		".mjs":  "application/javascript; charset=utf-8",
		".css":  "text/css; charset=utf-8",
		".json": "application/json; charset=utf-8",
		".svg":  "image/svg+xml; charset=utf-8",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".webp": "image/webp",
		".gif":  "image/gif",
		".ico":  "image/x-icon",
	}
)

type ModManifest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	EntryJS     string `json:"entry_js"`
	EntryCSS    string `json:"entry_css"`
	UpdatedAt   string `json:"updated_at"`
}

type ModSummary struct {
	ModManifest
	HasJS   bool   `json:"has_js"`
	HasCSS  bool   `json:"has_css"`
	JSURL   string `json:"js_url,omitempty"`
	CSSURL  string `json:"css_url,omitempty"`
	JSCode  string `json:"js_code,omitempty"`
	CSSCode string `json:"css_code,omitempty"`
}

type ModsHandler struct {
	db  *database.DB
	cfg *config.Config
	mu  sync.RWMutex
}

func NewModsHandler(db *database.DB, cfg *config.Config) *ModsHandler {
	return &ModsHandler{db: db, cfg: cfg}
}

func (h *ModsHandler) modsDir() string {
	dataDir := "./data"
	if h.cfg != nil && strings.TrimSpace(h.cfg.Paths.DataDir) != "" {
		dataDir = strings.TrimSpace(h.cfg.Paths.DataDir)
	}
	return filepath.Join(dataDir, "mods")
}

func normalizeModID(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	var b strings.Builder
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else if r == ' ' || r == '.' {
			b.WriteByte('-')
		}
	}
	id := strings.Trim(b.String(), "-_")
	if len(id) > 48 {
		id = id[:48]
		id = strings.Trim(id, "-_")
	}
	return id
}

func isValidModID(id string) bool {
	return safeModIDRe.MatchString(id)
}

func cleanModRelPath(rel, fallback string) string {
	rel = strings.ReplaceAll(strings.TrimSpace(rel), "\\", "/")
	rel = strings.TrimPrefix(path.Clean("/"+rel), "/")
	if rel == "" || rel == "." || strings.Contains(rel, "..") {
		return fallback
	}
	return rel
}

func (h *ModsHandler) safeModFilePath(modID, rel string) (string, error) {
	if !isValidModID(modID) {
		return "", fmt.Errorf("invalid mod id")
	}
	modDir := filepath.Join(h.modsDir(), modID)
	cleaned := cleanModRelPath(rel, "")
	if cleaned == "" {
		return "", fmt.Errorf("invalid mod asset path")
	}
	fullPath := filepath.Join(modDir, filepath.FromSlash(cleaned))
	relCheck, err := filepath.Rel(modDir, fullPath)
	if err != nil || relCheck == "." || strings.HasPrefix(relCheck, "..") {
		return "", fmt.Errorf("mod asset path escapes mod folder")
	}
	return fullPath, nil
}

func (h *ModsHandler) readModSummary(modID string, includeCode bool) (*ModSummary, error) {
	if !isValidModID(modID) {
		return nil, fmt.Errorf("invalid mod id")
	}
	modDir := filepath.Join(h.modsDir(), modID)
	manifestPath := filepath.Join(modDir, "manifest.json")

	manifest := ModManifest{
		ID:        modID,
		Name:      modID,
		Version:   "1.0.0",
		Author:    "Custom",
		Enabled:   true,
		EntryJS:   "mod.js",
		EntryCSS:  "mod.css",
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if data, err := os.ReadFile(manifestPath); err == nil {
		var parsed struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Version     string `json:"version"`
			Author      string `json:"author"`
			Description string `json:"description"`
			Enabled     *bool  `json:"enabled"`
			EntryJS     string `json:"entry_js"`
			EntryCSS    string `json:"entry_css"`
			UpdatedAt   string `json:"updated_at"`
		}
		if jsonErr := json.Unmarshal(data, &parsed); jsonErr == nil {
			if parsed.Name != "" {
				manifest.Name = parsed.Name
			}
			if parsed.Version != "" {
				manifest.Version = parsed.Version
			}
			if parsed.Author != "" {
				manifest.Author = parsed.Author
			}
			manifest.Description = parsed.Description
			if parsed.Enabled != nil {
				manifest.Enabled = *parsed.Enabled
			}
			if parsed.EntryJS != "" {
				manifest.EntryJS = cleanModRelPath(parsed.EntryJS, "mod.js")
			}
			if parsed.EntryCSS != "" {
				manifest.EntryCSS = cleanModRelPath(parsed.EntryCSS, "mod.css")
			}
			if parsed.UpdatedAt != "" {
				manifest.UpdatedAt = parsed.UpdatedAt
			}
		}
	}
	manifest.ID = modID

	summary := &ModSummary{
		ModManifest: manifest,
	}

	if jsPath, err := h.safeModFilePath(modID, manifest.EntryJS); err == nil {
		if data, readErr := os.ReadFile(jsPath); readErr == nil && len(bytes.TrimSpace(data)) > 0 {
			summary.HasJS = true
			summary.JSURL = fmt.Sprintf("/mods/%s/%s", modID, manifest.EntryJS)
			if includeCode {
				summary.JSCode = string(data)
			}
		}
	}
	if cssPath, err := h.safeModFilePath(modID, manifest.EntryCSS); err == nil {
		if data, readErr := os.ReadFile(cssPath); readErr == nil && len(bytes.TrimSpace(data)) > 0 {
			summary.HasCSS = true
			summary.CSSURL = fmt.Sprintf("/mods/%s/%s", modID, manifest.EntryCSS)
			if includeCode {
				summary.CSSCode = string(data)
			}
		}
	}

	return summary, nil
}

func (h *ModsHandler) listAllMods(includeCode bool) ([]ModSummary, error) {
	dir := h.modsDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := make([]ModSummary, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		modID := e.Name()
		if !isValidModID(modID) {
			continue
		}
		item, err := h.readModSummary(modID, includeCode)
		if err != nil || item == nil {
			continue
		}
		out = append(out, *item)
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}

// ListRuntimeMods returns enabled mods metadata + URLs for browser execution.
func (h *ModsHandler) ListRuntimeMods(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	all, err := h.listAllMods(false)
	h.mu.RUnlock()
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"mods": []ModSummary{}})
		return
	}
	enabled := make([]ModSummary, 0, len(all))
	for _, m := range all {
		if m.Enabled && (m.HasJS || m.HasCSS) {
			enabled = append(enabled, m)
		}
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]interface{}{"mods": enabled})
}

// ServeModAsset serves JS/CSS/media files belonging to an installed mod.
func (h *ModsHandler) ServeModAsset(w http.ResponseWriter, r *http.Request) {
	modID := strings.TrimSpace(getURLParam(r, "id"))
	assetPath := strings.TrimSpace(getURLParam(r, "path"))
	fullPath, err := h.safeModFilePath(modID, assetPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}
	ext := strings.ToLower(filepath.Ext(fullPath))
	ct, ok := modAssetCT[ext]
	if !ok {
		ct = "application/octet-stream"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "no-store")
	http.ServeFile(w, r, fullPath)
}

// ListAdminMods returns all installed mods including inline source code for the Admin Mods IDE.
func (h *ModsHandler) ListAdminMods(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil || claims.Role != "admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Admin privileges are required"})
		return
	}
	h.mu.RLock()
	mods, err := h.listAllMods(true)
	h.mu.RUnlock()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to list mods directory"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"mods":     mods,
		"mods_dir": h.modsDir(),
	})
}

// SaveAdminMod creates or updates a mod manifest and its JS/CSS files.
func (h *ModsHandler) SaveAdminMod(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil || claims.Role != "admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Admin privileges are required"})
		return
	}

	var req struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Version     string  `json:"version"`
		Author      string  `json:"author"`
		Description string  `json:"description"`
		Enabled     *bool   `json:"enabled"`
		EntryJS     string  `json:"entry_js"`
		EntryCSS    string  `json:"entry_css"`
		JSCode      *string `json:"js_code"`
		CSSCode     *string `json:"css_code"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 4<<20)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON mod payload"})
		return
	}

	pathID := strings.TrimSpace(getURLParam(r, "id"))
	rawID := req.ID
	if pathID != "" {
		rawID = pathID
	}
	modID := normalizeModID(rawID)
	if !isValidModID(modID) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Mod ID must be 2-48 lowercase letters, numbers, dashes, or underscores"})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	modDir := filepath.Join(h.modsDir(), modID)
	if err := os.MkdirAll(modDir, 0o755); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create mod folder"})
		return
	}

	existing, _ := h.readModSummary(modID, true)
	manifest := ModManifest{
		ID:        modID,
		Name:      modID,
		Version:   "1.0.0",
		Author:    claims.Username,
		Enabled:   true,
		EntryJS:   "mod.js",
		EntryCSS:  "mod.css",
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if existing != nil {
		manifest = existing.ModManifest
		manifest.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	if name := strings.TrimSpace(req.Name); name != "" {
		manifest.Name = name
	}
	if ver := strings.TrimSpace(req.Version); ver != "" {
		manifest.Version = ver
	}
	if auth := strings.TrimSpace(req.Author); auth != "" {
		manifest.Author = auth
	}
	manifest.Description = strings.TrimSpace(req.Description)
	if req.Enabled != nil {
		manifest.Enabled = *req.Enabled
	}
	if ejs := strings.TrimSpace(req.EntryJS); ejs != "" {
		manifest.EntryJS = cleanModRelPath(ejs, "mod.js")
	}
	if ecss := strings.TrimSpace(req.EntryCSS); ecss != "" {
		manifest.EntryCSS = cleanModRelPath(ecss, "mod.css")
	}

	if req.JSCode != nil {
		jsFullPath, err := h.safeModFilePath(modID, manifest.EntryJS)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid entry_js path"})
			return
		}
		if strings.TrimSpace(*req.JSCode) == "" {
			_ = os.Remove(jsFullPath)
		} else {
			if err := os.MkdirAll(filepath.Dir(jsFullPath), 0o755); err == nil {
				_ = os.WriteFile(jsFullPath, []byte(*req.JSCode), 0o644)
			}
		}
	}

	if req.CSSCode != nil {
		cssFullPath, err := h.safeModFilePath(modID, manifest.EntryCSS)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid entry_css path"})
			return
		}
		if strings.TrimSpace(*req.CSSCode) == "" {
			_ = os.Remove(cssFullPath)
		} else {
			if err := os.MkdirAll(filepath.Dir(cssFullPath), 0o755); err == nil {
				_ = os.WriteFile(cssFullPath, []byte(*req.CSSCode), 0o644)
			}
		}
	}

	mBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to encode manifest.json"})
		return
	}
	if err := os.WriteFile(filepath.Join(modDir, "manifest.json"), mBytes, 0o644); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save manifest.json"})
		return
	}

	updated, err := h.readModSummary(modID, true)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Saved mod, but failed to reload summary"})
		return
	}

	logAudit(h.db, r, "admin_mod_save", fmt.Sprintf("Saved panel mod %s (%s)", modID, manifest.Name))
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Mod saved successfully",
		"mod":     updated,
	})
}

// UploadAdminMod installs a mod from a .zip bundle or standalone .js/.css file.
func (h *ModsHandler) UploadAdminMod(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil || claims.Role != "admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Admin privileges are required"})
		return
	}

	const maxModUpload = 20 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxModUpload+4096)
	if err := r.ParseMultipartForm(maxModUpload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Package too large (max 20 MiB) or invalid upload"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing file in multipart field 'file'"})
		return
	}
	defer file.Close()

	rawBytes, err := io.ReadAll(io.LimitReader(file, maxModUpload+1))
	if err != nil || len(rawBytes) > maxModUpload {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Failed to read package or exceeded 20 MiB limit"})
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	stem := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	requestedID := normalizeModID(r.FormValue("id"))
	if requestedID == "" {
		requestedID = normalizeModID(stem)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	switch ext {
	case ".js", ".mjs", ".css":
		modID := requestedID
		if !isValidModID(modID) {
			modID = fmt.Sprintf("mod-%d", time.Now().Unix())
		}
		modDir := filepath.Join(h.modsDir(), modID)
		if err := os.MkdirAll(modDir, 0o755); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create mod folder"})
			return
		}
		manifest := ModManifest{
			ID:        modID,
			Name:      stem,
			Version:   "1.0.0",
			Author:    claims.Username,
			Enabled:   true,
			EntryJS:   "mod.js",
			EntryCSS:  "mod.css",
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		}
		if ext == ".css" {
			_ = os.WriteFile(filepath.Join(modDir, "mod.css"), rawBytes, 0o644)
		} else {
			_ = os.WriteFile(filepath.Join(modDir, "mod.js"), rawBytes, 0o644)
		}
		mBytes, _ := json.MarshalIndent(manifest, "", "  ")
		_ = os.WriteFile(filepath.Join(modDir, "manifest.json"), mBytes, 0o644)

		summary, _ := h.readModSummary(modID, true)
		logAudit(h.db, r, "admin_mod_upload", fmt.Sprintf("Installed single-file mod %s from %s", modID, header.Filename))
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"message": "Mod uploaded and installed",
			"mod":     summary,
		})
		return

	case ".zip":
		zr, err := zip.NewReader(bytes.NewReader(rawBytes), int64(len(rawBytes)))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid .zip archive: " + err.Error()})
			return
		}
		type zipEntry struct {
			rel  string
			data []byte
		}
		rawEntries := make([]zipEntry, 0, len(zr.File))
		topDirs := map[string]bool{}
		var totalBytes int64
		for _, f := range zr.File {
			if f.FileInfo().IsDir() || f.Mode()&os.ModeSymlink != 0 {
				continue
			}
			rel := strings.ReplaceAll(f.Name, "\\", "/")
			rel = strings.TrimPrefix(path.Clean("/"+rel), "/")
			if rel == "" || rel == "." || strings.Contains(rel, "..") ||
				strings.HasPrefix(rel, "__MACOSX/") || strings.HasSuffix(rel, ".DS_Store") {
				continue
			}
			rc, err := f.Open()
			if err != nil {
				continue
			}
			content, err := io.ReadAll(io.LimitReader(rc, (10<<20)+1))
			_ = rc.Close()
			if err != nil || len(content) > (10<<20) {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Zip entry exceeds 10 MiB limit"})
				return
			}
			totalBytes += int64(len(content))
			if totalBytes > (25<<20) || len(rawEntries) >= 150 {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Unpacked mod exceeds 25 MiB / 150 file safety limit"})
				return
			}
			if idx := strings.IndexByte(rel, '/'); idx > 0 {
				topDirs[rel[:idx]] = true
			} else {
				topDirs[""] = true
			}
			rawEntries = append(rawEntries, zipEntry{rel: rel, data: content})
		}
		if len(rawEntries) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Zip archive contains no usable mod files"})
			return
		}

		// Strip common top-level folder wrapper if all files live inside one folder
		stripPrefix := ""
		if len(topDirs) == 1 && !topDirs[""] {
			for k := range topDirs {
				stripPrefix = k + "/"
			}
		}

		type zipManifestRaw struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Version     string `json:"version"`
			Author      string `json:"author"`
			Description string `json:"description"`
			Enabled     *bool  `json:"enabled"`
			EntryJS     string `json:"entry_js"`
			EntryCSS    string `json:"entry_css"`
		}
		var manifestFromZip *zipManifestRaw
		var firstJS, firstCSS string
		cleanedEntries := make([]zipEntry, 0, len(rawEntries))
		for _, e := range rawEntries {
			rel := e.rel
			if stripPrefix != "" && strings.HasPrefix(rel, stripPrefix) {
				rel = strings.TrimPrefix(rel, stripPrefix)
			}
			rel = cleanModRelPath(rel, "")
			if rel == "" {
				continue
			}
			if rel == "manifest.json" {
				var parsed zipManifestRaw
				if json.Unmarshal(e.data, &parsed) == nil {
					manifestFromZip = &parsed
				}
			}
			lower := strings.ToLower(rel)
			if strings.HasSuffix(lower, ".js") || strings.HasSuffix(lower, ".mjs") {
				if firstJS == "" || lower == "mod.js" || lower == "index.js" {
					firstJS = rel
				}
			}
			if strings.HasSuffix(lower, ".css") {
				if firstCSS == "" || lower == "mod.css" || lower == "style.css" {
					firstCSS = rel
				}
			}
			cleanedEntries = append(cleanedEntries, zipEntry{rel: rel, data: e.data})
		}

		modID := requestedID
		if manifestFromZip != nil && normalizeModID(manifestFromZip.ID) != "" && r.FormValue("id") == "" {
			modID = normalizeModID(manifestFromZip.ID)
		}
		if !isValidModID(modID) {
			modID = fmt.Sprintf("mod-%d", time.Now().Unix())
		}

		modDir := filepath.Join(h.modsDir(), modID)
		_ = os.RemoveAll(modDir)
		if err := os.MkdirAll(modDir, 0o755); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to prepare mod destination"})
			return
		}
		for _, e := range cleanedEntries {
			dst, err := h.safeModFilePath(modID, e.rel)
			if err != nil {
				continue
			}
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err == nil {
				_ = os.WriteFile(dst, e.data, 0o644)
			}
		}

		manifest := ModManifest{
			ID:        modID,
			Name:      stem,
			Version:   "1.0.0",
			Author:    claims.Username,
			Enabled:   true,
			EntryJS:   "mod.js",
			EntryCSS:  "mod.css",
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		}
		if manifestFromZip != nil {
			if strings.TrimSpace(manifestFromZip.Name) != "" {
				manifest.Name = strings.TrimSpace(manifestFromZip.Name)
			}
			if strings.TrimSpace(manifestFromZip.Version) != "" {
				manifest.Version = strings.TrimSpace(manifestFromZip.Version)
			}
			if strings.TrimSpace(manifestFromZip.Author) != "" {
				manifest.Author = strings.TrimSpace(manifestFromZip.Author)
			}
			manifest.Description = strings.TrimSpace(manifestFromZip.Description)
			if manifestFromZip.Enabled != nil {
				manifest.Enabled = *manifestFromZip.Enabled
			}
			if strings.TrimSpace(manifestFromZip.EntryJS) != "" {
				manifest.EntryJS = cleanModRelPath(manifestFromZip.EntryJS, "mod.js")
			}
			if strings.TrimSpace(manifestFromZip.EntryCSS) != "" {
				manifest.EntryCSS = cleanModRelPath(manifestFromZip.EntryCSS, "mod.css")
			}
		} else {
			if firstJS != "" {
				manifest.EntryJS = firstJS
			}
			if firstCSS != "" {
				manifest.EntryCSS = firstCSS
			}
		}
		mBytes, _ := json.MarshalIndent(manifest, "", "  ")
		_ = os.WriteFile(filepath.Join(modDir, "manifest.json"), mBytes, 0o644)

		summary, _ := h.readModSummary(modID, true)
		logAudit(h.db, r, "admin_mod_upload", fmt.Sprintf("Installed zip mod %s from %s", modID, header.Filename))
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"message": "Mod bundle installed",
			"mod":     summary,
		})
		return

	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Unsupported package format. Upload a .zip bundle, .js file, or .css file."})
	}
}

// DeleteAdminMod removes a mod folder from disk.
func (h *ModsHandler) DeleteAdminMod(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil || claims.Role != "admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Admin privileges are required"})
		return
	}

	modID := normalizeModID(getURLParam(r, "id"))
	if !isValidModID(modID) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid mod id"})
		return
	}

	h.mu.Lock()
	modDir := filepath.Join(h.modsDir(), modID)
	err := os.RemoveAll(modDir)
	h.mu.Unlock()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete mod folder"})
		return
	}

	logAudit(h.db, r, "admin_mod_delete", fmt.Sprintf("Removed panel mod %s", modID))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Mod removed"})
}
