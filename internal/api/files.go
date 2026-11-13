package api

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
)

type FileHandler struct {
	db      *database.DB
	manager *runner.ServerManager
}

func NewFileHandler(db *database.DB, manager *runner.ServerManager) *FileHandler {
	return &FileHandler{db: db, manager: manager}
}

type FileItem struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	IsDir    bool      `json:"is_dir"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

// safePath ensures requested path cannot escape server working directory
func (h *FileHandler) safePath(serverID, reqPath string) (string, error) {
	baseDir, err := filepath.Abs(h.manager.GetWorkingDir(serverID))
	if err != nil {
		return "", err
	}
	if resolved, e := filepath.EvalSymlinks(baseDir); e == nil {
		baseDir = resolved
	}
	cleanRel := filepath.Clean("/" + strings.ReplaceAll(reqPath, "\\", "/"))
	fullPath := filepath.Join(baseDir, cleanRel)
	rel, err := filepath.Rel(baseDir, fullPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("access denied: path escapes server directory")
	}
	// Do not follow user-controlled symlinks, including symlinked parent folders.
	for p := fullPath; p != baseDir; p = filepath.Dir(p) {
		info, e := os.Lstat(p)
		if e != nil && !os.IsNotExist(e) {
			return "", e
		}
		if e == nil && info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("access denied: symbolic links are not supported")
		}
	}
	return fullPath, nil
}

func (h *FileHandler) destructivePath(serverID, reqPath string) (string, error) {
	fullPath, err := h.safePath(serverID, reqPath)
	if err != nil {
		return "", err
	}
	root, err := h.safePath(serverID, "")
	if err != nil {
		return "", err
	}
	if fullPath == root {
		return "", fmt.Errorf("cannot modify server root directory")
	}
	return fullPath, nil
}

func (h *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	reqPath := r.URL.Query().Get("path")

	fullPath, err := h.safePath(serverID, reqPath)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to read directory"})
		return
	}

	files := make([]FileItem, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		itemPath := filepath.ToSlash(filepath.Join(reqPath, entry.Name()))
		files = append(files, FileItem{
			Name:     entry.Name(),
			Path:     itemPath,
			IsDir:    entry.IsDir(),
			Size:     info.Size(),
			Modified: info.ModTime(),
		})
	}

	// Default order: folders first, then files alphabetically (case-insensitive)
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	writeJSON(w, http.StatusOK, files)
}

func (h *FileHandler) GetContent(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	reqPath := r.URL.Query().Get("path")

	fullPath, err := h.safePath(serverID, reqPath)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "File not found or is a directory"})
		return
	}

	if info.Size() > 5*1024*1024 { // 5MB limit for text editor
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "File exceeds 5MB text editor limit"})
		return
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to read file"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"path":    reqPath,
		"content": string(content),
	})
}

func (h *FileHandler) SaveContent(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	fullPath, err := h.safePath(serverID, req.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := os.WriteFile(fullPath, []byte(req.Content), 0644); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save file"})
		return
	}

	logAudit(h.db, r, "file_save", fmt.Sprintf("Saved file '%s' on %s (%s)", req.Path, auditServerLabel(h.db, serverID), formatBytesAudit(int64(len(req.Content)))))

	writeJSON(w, http.StatusOK, map[string]string{"message": "File saved successfully"})
}

func (h *FileHandler) Create(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	var req struct {
		Path  string `json:"path"`
		IsDir bool   `json:"is_dir"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	fullPath, err := h.safePath(serverID, req.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if req.IsDir {
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create directory"})
			return
		}
	} else {
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create parent directory"})
			return
		}
		f, err := os.Create(fullPath)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create file"})
			return
		}
		_ = f.Close()
	}

	kind := "file"
	if req.IsDir {
		kind = "directory"
	}
	logAudit(h.db, r, "file_create", fmt.Sprintf("Created %s '%s' on %s", kind, req.Path, auditServerLabel(h.db, serverID)))

	writeJSON(w, http.StatusOK, map[string]string{"message": "Created successfully"})
}

func (h *FileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	reqPath := r.URL.Query().Get("path")
	if reqPath == "" || reqPath == "/" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Cannot delete root directory"})
		return
	}

	fullPath, err := h.destructivePath(serverID, reqPath)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := os.RemoveAll(fullPath); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete file/folder"})
		return
	}

	logAudit(h.db, r, "file_delete", fmt.Sprintf("Deleted '%s' on %s", reqPath, auditServerLabel(h.db, serverID)))

	writeJSON(w, http.StatusOK, map[string]string{"message": "Deleted successfully"})
}

const (
	maxBulkDeletePaths = 1000
	maxBulkDeleteBody  = 1 << 20 // 1 MiB, including JSON and whitespace
)

type bulkDeleteFailure struct {
	Path  string `json:"path"`
	Error string `json:"error"`
}

// DeleteBulk validates the whole batch before removing anything. RemoveAll is
// idempotent, so duplicates, missing paths and overlapping directories are safe.
func (h *FileHandler) DeleteBulk(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	var req struct {
		Paths []string `json:"paths"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBulkDeleteBody)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request (maximum body size 1 MiB): " + err.Error()})
		return
	}
	if err := decoder.Decode(new(interface{})); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Request must contain one JSON object and fit within 1 MiB"})
		return
	}
	if len(req.Paths) == 0 || len(req.Paths) > maxBulkDeletePaths {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "paths must contain between 1 and 1000 file or directory paths"})
		return
	}

	fullPaths := make([]string, len(req.Paths))
	for i, reqPath := range req.Paths {
		// Reject traversal above the virtual root rather than silently clamping
		// it to another deletion target. Leading slashes remain server-relative.
		depth := 0
		for _, part := range strings.Split(strings.ReplaceAll(reqPath, "\\", "/"), "/") {
			switch part {
			case "", ".":
			case "..":
				depth--
			default:
				depth++
			}
			if depth < 0 {
				break
			}
		}
		if depth < 0 || strings.ContainsRune(reqPath, '\x00') {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid path %q: traversal outside the server or NUL characters are not allowed", reqPath)})
			return
		}
		fullPath, err := h.destructivePath(serverID, reqPath)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Invalid path %q: %s", reqPath, err)})
			return
		}
		fullPaths[i] = fullPath
	}

	result := struct {
		Deleted []string            `json:"deleted"`
		Failed  []bulkDeleteFailure `json:"failed"`
	}{Deleted: make([]string, 0, len(req.Paths)), Failed: make([]bulkDeleteFailure, 0)}
	for i, fullPath := range fullPaths {
		if err := os.RemoveAll(fullPath); err != nil {
			result.Failed = append(result.Failed, bulkDeleteFailure{Path: req.Paths[i], Error: err.Error()})
			continue
		}
		result.Deleted = append(result.Deleted, req.Paths[i])
	}
	logAudit(h.db, r, "file_delete_bulk", fmt.Sprintf("Bulk delete on %s: %d requested, %d deleted, %d failed", auditServerLabel(h.db, serverID), len(req.Paths), len(result.Deleted), len(result.Failed)))
	writeJSON(w, http.StatusOK, result)
}

func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	targetDir := r.URL.Query().Get("path")

	// Limit upload to 200MB
	_ = r.ParseMultipartForm(200 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Failed to retrieve uploaded file"})
		return
	}
	defer file.Close()

	destRel := filepath.Join(targetDir, handler.Filename)
	fullPath, err := h.safePath(serverID, destRel)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create destination file"})
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to write file"})
		return
	}

	logAudit(h.db, r, "file_upload", fmt.Sprintf("Uploaded '%s' (%s) to %s", filepath.ToSlash(destRel), formatBytesAudit(written), auditServerLabel(h.db, serverID)))

	writeJSON(w, http.StatusOK, map[string]string{"message": "File uploaded successfully", "name": handler.Filename})
}

func (h *FileHandler) Download(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	reqPath := r.URL.Query().Get("path")
	fullPath, err := h.safePath(serverID, reqPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		http.Error(w, "File not found or is a directory", http.StatusNotFound)
		return
	}
	logAudit(h.db, r, "file_download", fmt.Sprintf("Downloaded '%s' (%s) from %s", reqPath, formatBytesAudit(info.Size()), auditServerLabel(h.db, serverID)))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(fullPath)))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, fullPath)
}

func (h *FileHandler) Rename(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	var req struct {
		OldPath string `json:"old_path"`
		NewPath string `json:"new_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}
	oldFull, err := h.destructivePath(serverID, req.OldPath)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	newFull, err := h.destructivePath(serverID, req.NewPath)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := os.Rename(oldFull, newFull); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to rename: " + err.Error()})
		return
	}
	logAudit(h.db, r, "file_rename", fmt.Sprintf("Renamed '%s' -> '%s' on %s", req.OldPath, req.NewPath, auditServerLabel(h.db, serverID)))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Renamed successfully"})
}

func (h *FileHandler) Move(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	baseDir := h.manager.GetWorkingDir(serverID)

	var req struct {
		Paths       []string `json:"paths"`
		Files       []string `json:"files"`
		Destination string   `json:"destination"`
		From        string   `json:"from"`
		To          string   `json:"to"`
		Overwrite   bool     `json:"overwrite"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	sources := req.Paths
	if len(sources) == 0 {
		sources = req.Files
	}

	// Single from -> to move fallback
	if len(sources) == 0 && strings.TrimSpace(req.From) != "" {
		fromRel := strings.TrimSpace(req.From)
		if fromRel == "" || fromRel == "/" || fromRel == "." {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Cannot move server root directory"})
			return
		}
		srcFull, err := h.destructivePath(serverID, fromRel)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if srcFull == baseDir {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Cannot move server root directory"})
			return
		}
		srcInfo, err := os.Stat(srcFull)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Source path not found"})
			return
		}

		if strings.TrimSpace(req.Destination) != "" {
			sources = []string{fromRel}
		} else {
			toRel := strings.TrimSpace(req.To)
			if toRel == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Destination path is required"})
				return
			}
			dstFull, err := h.safePath(serverID, toRel)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			if dstInfo, statErr := os.Stat(dstFull); statErr == nil && dstInfo.IsDir() {
				dstFull = filepath.Join(dstFull, filepath.Base(srcFull))
			} else if strings.HasSuffix(toRel, "/") || strings.HasSuffix(toRel, "\\") {
				if err := os.MkdirAll(dstFull, 0755); err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create destination folder: " + err.Error()})
					return
				}
				dstFull = filepath.Join(dstFull, filepath.Base(srcFull))
			}

			sep := string(filepath.Separator)
			if srcInfo.IsDir() && (dstFull == srcFull || strings.HasPrefix(dstFull+sep, srcFull+sep)) {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Cannot move a folder into itself or its subdirectory"})
				return
			}
			if dstFull == srcFull {
				writeJSON(w, http.StatusOK, map[string]interface{}{"message": "Item is already at destination", "count": 0})
				return
			}
			if _, statErr := os.Stat(dstFull); statErr == nil {
				if !req.Overwrite {
					writeJSON(w, http.StatusConflict, map[string]string{"error": "Destination already exists"})
					return
				}
				_ = os.RemoveAll(dstFull)
			}
			if err := os.MkdirAll(filepath.Dir(dstFull), 0755); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create parent folder: " + err.Error()})
				return
			}
			if err := moveFileOrDir(srcFull, dstFull); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to move item: " + err.Error()})
				return
			}
			logAudit(h.db, r, "file_move", fmt.Sprintf("Moved '%s' -> '%s' on %s", fromRel, toRel, auditServerLabel(h.db, serverID)))
			writeJSON(w, http.StatusOK, map[string]interface{}{"message": "Moved successfully", "count": 1})
			return
		}
	}

	if len(sources) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "At least one file or folder path is required"})
		return
	}

	destDir := strings.TrimSpace(strings.ReplaceAll(req.Destination, "\\", "/"))
	if destDir == "/" || destDir == "." {
		destDir = ""
	}
	destFull, err := h.safePath(serverID, destDir)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid destination folder: " + err.Error()})
		return
	}
	if st, err := os.Stat(destFull); err == nil && !st.IsDir() {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Destination path is an existing file, not a folder"})
		return
	}

	type moveOp struct {
		srcRel  string
		srcFull string
		dstFull string
	}
	seen := make(map[string]bool)
	ops := make([]moveOp, 0, len(sources))
	sep := string(filepath.Separator)

	for _, raw := range sources {
		srcRel := strings.TrimSpace(strings.ReplaceAll(raw, "\\", "/"))
		if srcRel == "" || srcRel == "/" || srcRel == "." {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Cannot move server root directory"})
			return
		}
		if seen[srcRel] {
			continue
		}
		seen[srcRel] = true

		srcFull, err := h.destructivePath(serverID, srcRel)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if srcFull == baseDir {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Cannot move server root directory"})
			return
		}
		srcInfo, err := os.Stat(srcFull)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Item '%s' not found", srcRel)})
			return
		}

		baseName := filepath.Base(srcFull)
		dstFull := filepath.Join(destFull, baseName)

		if srcInfo.IsDir() && (dstFull == srcFull || strings.HasPrefix(destFull+sep, srcFull+sep)) {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("Cannot move folder '%s' into itself or a subdirectory", baseName),
			})
			return
		}

		if srcFull == dstFull {
			continue
		}

		if _, err := os.Stat(dstFull); err == nil && !req.Overwrite {
			writeJSON(w, http.StatusConflict, map[string]string{
				"error": fmt.Sprintf("'%s' already exists in destination folder", baseName),
			})
			return
		}

		ops = append(ops, moveOp{
			srcRel:  srcRel,
			srcFull: srcFull,
			dstFull: dstFull,
		})
	}

	if len(ops) == 0 {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"message":     "Selected items are already in the destination folder",
			"count":       0,
			"destination": strings.Trim(destDir, "/"),
		})
		return
	}

	if err := os.MkdirAll(destFull, 0755); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create destination folder: " + err.Error()})
		return
	}

	for _, op := range ops {
		if req.Overwrite {
			_ = os.RemoveAll(op.dstFull)
		}
		if err := moveFileOrDir(op.srcFull, op.dstFull); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to move '%s': %v", filepath.Base(op.srcFull), err),
			})
			return
		}
	}

	cleanDest := strings.Trim(destDir, "/")
	destLabel := "/"
	if cleanDest != "" {
		destLabel = "/" + cleanDest
	}
	logAudit(h.db, r, "file_move", fmt.Sprintf("Moved %d item(s) to '%s' on %s", len(ops), destLabel, auditServerLabel(h.db, serverID)))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":     fmt.Sprintf("Moved %d item(s) to %s", len(ops), destLabel),
		"count":       len(ops),
		"destination": cleanDest,
	})
}

func (h *FileHandler) Compress(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	baseDir := h.manager.GetWorkingDir(serverID)

	var req struct {
		Paths       []string `json:"paths"`
		Files       []string `json:"files"`
		Destination string   `json:"destination"`
		Filename    string   `json:"filename"`
		Name        string   `json:"name"`
		Format      string   `json:"format"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	sources := req.Paths
	if len(sources) == 0 {
		sources = req.Files
	}
	if len(sources) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Select at least one file or folder to archive"})
		return
	}

	format := strings.ToLower(strings.TrimSpace(req.Format))
	if format == "tgz" || format == "targz" || format == "tar.gz" {
		format = "tar.gz"
	} else if format == "" || format == "zip" {
		format = "zip"
	} else {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Unsupported archive format (use zip or tar.gz)"})
		return
	}

	destDir := strings.TrimSpace(strings.ReplaceAll(req.Destination, "\\", "/"))
	if destDir == "/" || destDir == "." {
		destDir = ""
	}
	if destDir == "" && len(sources) > 0 {
		firstParent := filepath.ToSlash(filepath.Dir(strings.TrimSpace(strings.ReplaceAll(sources[0], "\\", "/"))))
		if firstParent != "." && firstParent != "/" {
			destDir = strings.Trim(firstParent, "/")
		}
	}

	destFull, err := h.safePath(serverID, destDir)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid destination directory: " + err.Error()})
		return
	}
	if err := os.MkdirAll(destFull, 0755); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to prepare destination directory: " + err.Error()})
		return
	}

	rawName := strings.TrimSpace(req.Filename)
	if rawName == "" {
		rawName = strings.TrimSpace(req.Name)
	}
	rawName = strings.ReplaceAll(rawName, "\\", "/")
	if idx := strings.LastIndex(rawName, "/"); idx >= 0 {
		rawName = rawName[idx+1:]
	}
	rawName = strings.TrimSpace(rawName)
	if rawName == "" || rawName == "." || rawName == ".." {
		rawName = fmt.Sprintf("archive-%s", time.Now().Format("20060102-150405"))
	}

	lowerName := strings.ToLower(rawName)
	if format == "tar.gz" {
		if !strings.HasSuffix(lowerName, ".tar.gz") && !strings.HasSuffix(lowerName, ".tgz") {
			rawName += ".tar.gz"
		}
	} else {
		if !strings.HasSuffix(lowerName, ".zip") {
			rawName += ".zip"
		}
	}

	outRel := strings.Trim(filepath.ToSlash(filepath.Join(destDir, rawName)), "/")
	outFull, err := h.safePath(serverID, outRel)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid archive filename: " + err.Error()})
		return
	}
	if _, err := os.Stat(outFull); err == nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": fmt.Sprintf("Archive '%s' already exists", rawName)})
		return
	}

	items := make([]archiveSource, 0, len(sources))
	seen := make(map[string]bool)
	for _, raw := range sources {
		srcRel := strings.TrimSpace(strings.ReplaceAll(raw, "\\", "/"))
		if srcRel == "" || srcRel == "/" || srcRel == "." {
			continue
		}
		if seen[srcRel] {
			continue
		}
		seen[srcRel] = true

		srcFull, err := h.safePath(serverID, srcRel)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if srcFull == baseDir {
			continue
		}
		if _, err := os.Stat(srcFull); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Source '%s' not found", srcRel)})
			return
		}

		entryRoot := filepath.Base(srcFull)
		if rel, relErr := filepath.Rel(destFull, srcFull); relErr == nil && !strings.HasPrefix(rel, "..") && rel != "." {
			entryRoot = filepath.ToSlash(rel)
		}
		items = append(items, archiveSource{
			full:      srcFull,
			entryRoot: entryRoot,
		})
	}

	if len(items) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No valid files or folders selected"})
		return
	}

	tmpFull := outFull + ".neopanel-tmp"
	_ = os.Remove(tmpFull)

	var fileCount int
	if format == "tar.gz" {
		fileCount, err = buildTarGzArchive(tmpFull, items, map[string]bool{outFull: true, tmpFull: true})
	} else {
		fileCount, err = buildZipArchive(tmpFull, items, map[string]bool{outFull: true, tmpFull: true})
	}
	if err != nil {
		_ = os.Remove(tmpFull)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create archive: " + err.Error()})
		return
	}

	if err := os.Rename(tmpFull, outFull); err != nil {
		_ = os.Remove(tmpFull)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to finalize archive: " + err.Error()})
		return
	}

	var outSize int64
	if st, statErr := os.Stat(outFull); statErr == nil {
		outSize = st.Size()
	}

	logAudit(h.db, r, "file_compress", fmt.Sprintf("Archived %d item(s) into '%s' (%s) on %s", len(items), outRel, formatBytesAudit(outSize), auditServerLabel(h.db, serverID)))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("Created %s (%d file(s), %s)", rawName, fileCount, formatBytesAudit(outSize)),
		"file":    rawName,
		"path":    outRel,
		"size":    outSize,
		"count":   fileCount,
	})
}

func moveFileOrDir(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		if err := copyDirRecursive(src, dst); err != nil {
			return err
		}
		return os.RemoveAll(src)
	}
	if err := copySingleFile(src, dst, info.Mode()); err != nil {
		return err
	}
	return os.Remove(src)
}

func copyDirRecursive(srcDir, dstDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dstDir, rel)
		if info.IsDir() {
			mode := info.Mode()
			if mode == 0 {
				mode = 0755
			}
			return os.MkdirAll(target, mode)
		}
		return copySingleFile(path, target, info.Mode())
	})
}

func copySingleFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	if mode == 0 {
		mode = 0644
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

type archiveSource struct {
	full      string
	entryRoot string
}

func buildZipArchive(outPath string, sources []archiveSource, skipPaths map[string]bool) (int, error) {
	outFile, err := os.Create(outPath)
	if err != nil {
		return 0, err
	}
	defer outFile.Close()

	zw := zip.NewWriter(outFile)
	defer zw.Close()

	count := 0
	for _, src := range sources {
		err := filepath.Walk(src.full, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if skipPaths[path] {
				return nil
			}
			if info.Mode()&os.ModeSymlink != 0 {
				return nil
			}
			rel, err := filepath.Rel(src.full, path)
			if err != nil {
				return err
			}
			archiveRel := src.entryRoot
			if rel != "." {
				archiveRel = filepath.ToSlash(filepath.Join(src.entryRoot, rel))
			} else {
				archiveRel = filepath.ToSlash(src.entryRoot)
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			if info.IsDir() {
				header.Name = strings.TrimSuffix(archiveRel, "/") + "/"
				_, err = zw.CreateHeader(header)
				return err
			}
			header.Name = archiveRel
			header.Method = zip.Deflate

			w, err := zw.CreateHeader(header)
			if err != nil {
				return err
			}
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(w, f); err != nil {
				return err
			}
			count++
			return nil
		})
		if err != nil {
			return count, err
		}
	}

	if err := zw.Close(); err != nil {
		return count, err
	}
	return count, outFile.Close()
}

func buildTarGzArchive(outPath string, sources []archiveSource, skipPaths map[string]bool) (int, error) {
	outFile, err := os.Create(outPath)
	if err != nil {
		return 0, err
	}
	defer outFile.Close()

	gw := gzip.NewWriter(outFile)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	count := 0
	for _, src := range sources {
		err := filepath.Walk(src.full, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if skipPaths[path] {
				return nil
			}
			if info.Mode()&os.ModeSymlink != 0 {
				return nil
			}
			rel, err := filepath.Rel(src.full, path)
			if err != nil {
				return err
			}
			archiveRel := src.entryRoot
			if rel != "." {
				archiveRel = filepath.ToSlash(filepath.Join(src.entryRoot, rel))
			} else {
				archiveRel = filepath.ToSlash(src.entryRoot)
			}

			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			header.Name = archiveRel
			if info.IsDir() && !strings.HasSuffix(header.Name, "/") {
				header.Name += "/"
			}
			if err := tw.WriteHeader(header); err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
			count++
			return nil
		})
		if err != nil {
			return count, err
		}
	}

	if err := tw.Close(); err != nil {
		return count, err
	}
	if err := gw.Close(); err != nil {
		return count, err
	}
	return count, outFile.Close()
}

func (h *FileHandler) Unzip(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	var req struct {
		Path        string `json:"path"`
		Destination string `json:"destination"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	if req.Path == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Archive path is required"})
		return
	}

	archiveFullPath, err := h.safePath(serverID, req.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	destDir := req.Destination
	if destDir == "" {
		destDir = filepath.ToSlash(filepath.Dir(req.Path))
		if destDir == "." {
			destDir = ""
		}
	}

	destFullPath, err := h.safePath(serverID, destDir)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid destination directory: " + err.Error()})
		return
	}

	if err := os.MkdirAll(destFullPath, 0755); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create destination folder: " + err.Error()})
		return
	}

	lower := strings.ToLower(req.Path)
	var count int
	if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") {
		count, err = extractTarGz(archiveFullPath, destFullPath)
	} else if strings.HasSuffix(lower, ".tar") {
		count, err = extractTar(archiveFullPath, destFullPath)
	} else {
		// Default to standard zip (handles .zip, .jar, etc.)
		count, err = extractZip(archiveFullPath, destFullPath)
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to extract archive: " + err.Error()})
		return
	}

	logAudit(h.db, r, "file_unzip", fmt.Sprintf("Extracted %d file(s) from '%s' on %s", count, req.Path, auditServerLabel(h.db, serverID)))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("Successfully extracted %d file(s)", count),
		"count":   count,
	})
}

func extractZip(zipPath, destDir string) (int, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	count := 0
	for _, f := range r.File {
		cleanName := filepath.Clean("/" + f.Name)
		target := filepath.Join(destDir, cleanName)
		rel, err := filepath.Rel(destDir, target)
		if err != nil || strings.HasPrefix(rel, "..") {
			continue // skip illegal paths (ZipSlip protection)
		}

		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(target, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return count, err
		}

		rc, err := f.Open()
		if err != nil {
			return count, err
		}

		mode := f.Mode()
		if mode == 0 {
			mode = 0644
		}
		outFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			rc.Close()
			return count, err
		}

		_, copyErr := io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if copyErr != nil {
			return count, copyErr
		}
		count++
	}
	return count, nil
}

func extractTarGz(tarGzPath, destDir string) (int, error) {
	f, err := os.Open(tarGzPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return 0, err
	}
	defer gzr.Close()

	return extractTarReader(tar.NewReader(gzr), destDir)
}

func extractTar(tarPath, destDir string) (int, error) {
	f, err := os.Open(tarPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return extractTarReader(tar.NewReader(f), destDir)
}

func extractTarReader(tr *tar.Reader, destDir string) (int, error) {
	count := 0
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, err
		}

		cleanName := filepath.Clean("/" + header.Name)
		target := filepath.Join(destDir, cleanName)
		rel, err := filepath.Rel(destDir, target)
		if err != nil || strings.HasPrefix(rel, "..") {
			continue // ZipSlip protection
		}

		switch header.Typeflag {
		case tar.TypeDir:
			_ = os.MkdirAll(target, 0755)
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return count, err
			}
			mode := header.FileInfo().Mode()
			if mode == 0 {
				mode = 0644
			}
			outFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
			if err != nil {
				return count, err
			}
			_, copyErr := io.Copy(outFile, tr)
			outFile.Close()
			if copyErr != nil {
				return count, copyErr
			}
			count++
		}
	}
	return count, nil
}
