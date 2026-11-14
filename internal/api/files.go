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
	"unicode/utf8"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/rooted"
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
	root, err := rooted.Open(h.manager.GetWorkingDir(serverID))
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	defer root.Close()
	reqPath := r.URL.Query().Get("path")

	fullPath, err := root.Path(reqPath)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	entries, err := root.ReadDir(fullPath)
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
	reqPath := r.URL.Query().Get("path")
	f, info, err := h.openPreviewFile(getURLParam(r, "id"), reqPath, os.O_RDONLY)
	if err != nil {
		status := http.StatusBadRequest
		if os.IsNotExist(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, "Cannot open text file: "+err.Error())
		return
	}
	defer f.Close()
	if info.Size() > fileEditorLimit {
		writeError(w, http.StatusBadRequest, "File exceeds 5 MiB text editor limit")
		return
	}
	content, err := readFileBounded(f, fileEditorLimit)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !editableFileText(info.Name(), content) {
		writeError(w, http.StatusUnsupportedMediaType, "Only UTF-8 text files can be edited")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"path": reqPath, "content": string(content)})
}

func (h *FileHandler) SaveContent(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	// JSON may encode each decoded byte as a six-byte Unicode escape.
	body, err := readFileBounded(r.Body, 6*fileEditorLimit+(64<<10))
	if err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "Request body exceeds text editor limit")
		return
	}
	if !utf8.Valid(body) || json.Unmarshal(body, &req) != nil {
		writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	if len(req.Content) > fileEditorLimit {
		writeError(w, http.StatusRequestEntityTooLarge, "Content exceeds 5 MiB text editor limit")
		return
	}
	if !editableFileText(req.Path, []byte(req.Content)) {
		writeError(w, http.StatusUnsupportedMediaType, "Only UTF-8 text files can be saved")
		return
	}
	root, err := rooted.Open(h.manager.GetWorkingDir(serverID))
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	defer root.Close()
	name, err := root.Path(req.Path)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	f, info, err := h.openPreviewFile(serverID, req.Path, os.O_RDWR)
	if os.IsNotExist(err) {
		f, info, err = h.openPreviewFile(serverID, req.Path, os.O_RDWR|os.O_CREATE|os.O_EXCL)
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "Cannot open text file: "+err.Error())
		return
	}
	defer f.Close()
	if info.Size() > fileEditorLimit {
		writeError(w, http.StatusRequestEntityTooLarge, "Existing file exceeds 5 MiB text editor limit")
		return
	}
	old, err := readFileBounded(f, fileEditorLimit)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Cannot validate existing file")
		return
	}
	if !editableFileText(info.Name(), old) {
		writeError(w, http.StatusUnsupportedMediaType, "Cannot overwrite a binary file with text")
		return
	}
	f.Close()
	err = root.WriteFile(name, []byte(req.Content), info.Mode().Perm())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save file")
		return
	}
	logAudit(h.db, r, "file_save", fmt.Sprintf("Saved file '%s' on %s (%s)", req.Path, auditServerLabel(h.db, serverID), formatBytesAudit(int64(len(req.Content)))))
	writeJSON(w, http.StatusOK, map[string]string{"message": "File saved successfully"})
}

func (h *FileHandler) Create(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	root, err := rooted.Open(h.manager.GetWorkingDir(serverID))
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	defer root.Close()
	var req struct {
		Path  string `json:"path"`
		IsDir bool   `json:"is_dir"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}

	fullPath, err := root.Path(req.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if req.IsDir {
		if err := root.MkdirAll(fullPath, 0755); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create directory"})
			return
		}
	} else {
		if err := root.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create parent directory"})
			return
		}
		if err := root.WriteFile(fullPath, nil, 0644); err != nil {
			writeError(w, 500, "Failed to create file: "+err.Error())
			return
		}
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
	root, err := rooted.Open(h.manager.GetWorkingDir(serverID))
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	defer root.Close()
	reqPath := r.URL.Query().Get("path")
	if reqPath == "" || reqPath == "/" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Cannot delete root directory"})
		return
	}

	fullPath, err := root.Destructive(reqPath)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := h.deleteFile(serverID, root, fullPath, r.URL.Query().Get("permanent") == "true"); err != nil {
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

// DeleteBulk validates the whole batch before removing anything. Missing paths,
// duplicates and overlapping directories remain idempotent in either mode.
func (h *FileHandler) DeleteBulk(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	root, err := rooted.Open(h.manager.GetWorkingDir(serverID))
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	defer root.Close()
	var req struct {
		Paths     []string `json:"paths"`
		Permanent bool     `json:"permanent"`
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
		fullPath, err := root.Destructive(reqPath)
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
		if err := h.deleteFile(serverID, root, fullPath, req.Permanent); err != nil {
			result.Failed = append(result.Failed, bulkDeleteFailure{Path: req.Paths[i], Error: err.Error()})
			continue
		}
		result.Deleted = append(result.Deleted, req.Paths[i])
	}
	logAudit(h.db, r, "file_delete_bulk", fmt.Sprintf("Bulk delete on %s: %d requested, %d deleted, %d failed", auditServerLabel(h.db, serverID), len(req.Paths), len(result.Deleted), len(result.Failed)))
	writeJSON(w, http.StatusOK, result)
}

func parseUploadForm(w http.ResponseWriter, r *http.Request, limit int64) error {
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	return r.ParseMultipartForm(8 << 20)
}

func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	root, err := rooted.Open(h.manager.GetWorkingDir(serverID))
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	defer root.Close()
	targetDir := r.URL.Query().Get("path")

	err = parseUploadForm(w, r, 200<<20)
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}
	if err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "Upload exceeds 200 MiB or invalid multipart body")
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Failed to retrieve uploaded file"})
		return
	}
	defer file.Close()
	if !rooted.Basename(handler.Filename) {
		writeError(w, 400, "Invalid filename")
		return
	}

	destRel := filepath.Join(targetDir, handler.Filename)
	fullPath, err := root.Path(destRel)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	written, err := root.Copy(r.Context(), fullPath, file, 200<<20, 0644)
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
	root, err := rooted.Open(h.manager.GetWorkingDir(serverID))
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	defer root.Close()
	fullPath, err := root.Path(reqPath)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	f, err := root.Open(fullPath)
	if err != nil {
		writeError(w, 404, err.Error())
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	logAudit(h.db, r, "file_download", fmt.Sprintf("Downloaded '%s' (%s) from %s", reqPath, formatBytesAudit(info.Size()), auditServerLabel(h.db, serverID)))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(fullPath)))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, info.Name(), info.ModTime(), f)
}

func (h *FileHandler) Rename(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	root, err := rooted.Open(h.manager.GetWorkingDir(serverID))
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	defer root.Close()
	var req struct {
		OldPath string `json:"old_path"`
		NewPath string `json:"new_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}
	oldFull, err := root.Destructive(req.OldPath)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	newFull, err := root.Destructive(req.NewPath)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := root.Rename(oldFull, newFull); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to rename: " + err.Error()})
		return
	}
	logAudit(h.db, r, "file_rename", fmt.Sprintf("Renamed '%s' -> '%s' on %s", req.OldPath, req.NewPath, auditServerLabel(h.db, serverID)))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Renamed successfully"})
}

func (h *FileHandler) Move(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	root, err := rooted.Open(h.manager.GetWorkingDir(serverID))
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	defer root.Close()
	baseDir := "."

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
		srcFull, err := root.Destructive(fromRel)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if srcFull == baseDir {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Cannot move server root directory"})
			return
		}
		srcInfo, err := root.Lstat(srcFull)
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
			dstFull, err := root.Path(toRel)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			if dstInfo, statErr := root.Lstat(dstFull); statErr == nil && dstInfo.IsDir() {
				dstFull = filepath.Join(dstFull, filepath.Base(srcFull))
			} else if strings.HasSuffix(toRel, "/") || strings.HasSuffix(toRel, "\\") {
				if err := root.MkdirAll(dstFull, 0755); err != nil {
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
			if _, statErr := root.Lstat(dstFull); statErr == nil {
				if !req.Overwrite {
					writeJSON(w, http.StatusConflict, map[string]string{"error": "Destination already exists"})
					return
				}
				if err := root.RemoveAll(dstFull); err != nil {
					writeError(w, 400, err.Error())
					return
				}
			}
			if err := root.MkdirAll(filepath.Dir(dstFull), 0755); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create parent folder: " + err.Error()})
				return
			}
			if err := root.Rename(srcFull, dstFull); err != nil {
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
	destFull, err := root.Path(destDir)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid destination folder: " + err.Error()})
		return
	}
	if st, err := root.Lstat(destFull); err == nil && !st.IsDir() {
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

		srcFull, err := root.Destructive(srcRel)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if srcFull == baseDir {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Cannot move server root directory"})
			return
		}
		srcInfo, err := root.Lstat(srcFull)
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

		if _, err := root.Lstat(dstFull); err == nil && !req.Overwrite {
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

	if err := root.MkdirAll(destFull, 0755); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create destination folder: " + err.Error()})
		return
	}

	for _, op := range ops {
		if req.Overwrite {
			if err := root.RemoveAll(op.dstFull); err != nil {
				writeError(w, 400, err.Error())
				return
			}
		}
		if err := root.Rename(op.srcFull, op.dstFull); err != nil {
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
	root, err := rooted.Open(h.manager.GetWorkingDir(serverID))
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	defer root.Close()
	baseDir := "."

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

	destFull, err := root.Path(destDir)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid destination directory: " + err.Error()})
		return
	}
	if err := root.MkdirAll(destFull, 0755); err != nil {
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
	outFull, err := root.Path(outRel)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid archive filename: " + err.Error()})
		return
	}
	if _, err := root.Lstat(outFull); err == nil {
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

		srcFull, err := root.Path(srcRel)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if srcFull == baseDir {
			continue
		}
		if _, err := root.Lstat(srcFull); err != nil {
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

	var fileCount int
	err = root.Publish(outFull, 0644, func(out *os.File) error {
		skip := map[string]bool{outFull: true}
		if format == "tar.gz" {
			fileCount, err = buildTarGzArchive(root, out, items, skip)
		} else {
			fileCount, err = buildZipArchive(root, out, items, skip)
		}
		return err
	})
	if err != nil {
		writeError(w, 500, "Failed to create archive: "+err.Error())
		return
	}

	var outSize int64
	if st, statErr := root.Lstat(outFull); statErr == nil {
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

type archiveSource struct {
	full      string
	entryRoot string
}

func buildZipArchive(root *rooted.Root, outFile io.Writer, sources []archiveSource, skipPaths map[string]bool) (int, error) {

	zw := zip.NewWriter(outFile)
	defer zw.Close()

	count := 0
	for _, src := range sources {
		err := root.Walk(src.full, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if skipPaths[path] || strings.HasPrefix(filepath.Base(path), ".neopanel-") {
				if info.IsDir() {
					return filepath.SkipDir
				}
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
			f, err := root.Open(path)
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
	return count, nil
}

func buildTarGzArchive(root *rooted.Root, outFile io.Writer, sources []archiveSource, skipPaths map[string]bool) (int, error) {

	gw := gzip.NewWriter(outFile)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	count := 0
	for _, src := range sources {
		err := root.Walk(src.full, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if skipPaths[path] || strings.HasPrefix(filepath.Base(path), ".neopanel-") {
				if info.IsDir() {
					return filepath.SkipDir
				}
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

			f, err := root.Open(path)
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
	return count, nil
}

func (h *FileHandler) Unzip(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	root, err := rooted.Open(h.manager.GetWorkingDir(serverID))
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	defer root.Close()
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

	archiveFullPath, err := root.Path(req.Path)
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

	destFullPath, err := root.Path(destDir)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid destination directory: " + err.Error()})
		return
	}

	if err := root.MkdirAll(destFullPath, 0755); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create destination folder: " + err.Error()})
		return
	}

	archive, err := root.Open(archiveFullPath)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	defer archive.Close()
	count, err := rooted.Extract(r.Context(), root, archive, req.Path, destFullPath, rooted.DefaultArchiveLimits)

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
