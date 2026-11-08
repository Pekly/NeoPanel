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
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/runner"
)

type FileHandler struct {
	manager *runner.ServerManager
}

func NewFileHandler(manager *runner.ServerManager) *FileHandler {
	return &FileHandler{manager: manager}
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
	baseDir := h.manager.GetWorkingDir(serverID)
	cleanRel := filepath.Clean("/" + reqPath)
	fullPath := filepath.Join(baseDir, cleanRel)

	rel, err := filepath.Rel(baseDir, fullPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("access denied: path escapes server directory")
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

	writeJSON(w, http.StatusOK, map[string]string{"message": "Created successfully"})
}

func (h *FileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	reqPath := r.URL.Query().Get("path")
	if reqPath == "" || reqPath == "/" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Cannot delete root directory"})
		return
	}

	fullPath, err := h.safePath(serverID, reqPath)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := os.RemoveAll(fullPath); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete file/folder"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Deleted successfully"})
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

	if _, err := io.Copy(dst, file); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to write file"})
		return
	}

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
	oldFull, err := h.safePath(serverID, req.OldPath)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	newFull, err := h.safePath(serverID, req.NewPath)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := os.Rename(oldFull, newFull); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to rename: " + err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Renamed successfully"})
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


