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

