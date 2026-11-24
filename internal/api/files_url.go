package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/rooted"
	"github.com/Pekly/NeoPanel/internal/runner"
)

type UploadFromURLRequest struct {
	URL       string `json:"url"`
	Filename  string `json:"filename,omitempty"`
	Path      string `json:"path,omitempty"`
	Overwrite bool   `json:"overwrite,omitempty"`
}

var urlDownloadClient = &http.Client{
	Timeout: 15 * time.Minute,
}

func extractFilenameFromURL(rawURL string, resp *http.Response) string {
	if resp != nil {
		cd := resp.Header.Get("Content-Disposition")
		if cd != "" {
			lower := strings.ToLower(cd)
			if idx := strings.Index(lower, "filename*="); idx != -1 {
				val := strings.TrimSpace(cd[idx+len("filename*="):])
				if semi := strings.Index(val, ";"); semi != -1 {
					val = val[:semi]
				}
				val = strings.Trim(val, `"'`)
				if parts := strings.SplitN(val, "''", 2); len(parts) == 2 {
					if unesc, err := url.PathUnescape(parts[1]); err == nil && unesc != "" {
						base := filepath.Base(filepath.Clean(unesc))
						if rooted.Basename(base) {
							return base
						}
					}
				}
			}
			if idx := strings.Index(lower, "filename="); idx != -1 {
				val := strings.TrimSpace(cd[idx+len("filename="):])
				if semi := strings.Index(val, ";"); semi != -1 {
					val = val[:semi]
				}
				val = strings.Trim(val, `"'`)
				if val != "" {
					base := filepath.Base(filepath.Clean(val))
					if rooted.Basename(base) {
						return base
					}
				}
			}
		}
	}
	if u, err := url.Parse(rawURL); err == nil {
		cleanPath := path.Clean(u.Path)
		base := path.Base(cleanPath)
		if unesc, err := url.PathUnescape(base); err == nil && unesc != "" && unesc != "." && unesc != "/" {
			baseClean := filepath.Base(filepath.Clean(unesc))
			if rooted.Basename(baseClean) {
				return baseClean
			}
		}
	}
	return "downloaded_file"
}

func (h *FileHandler) UploadFromURL(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	if serverID == "" {
		writeError(w, http.StatusBadRequest, "Server ID is required")
		return
	}

	var req UploadFromURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request JSON payload")
		return
	}

	rawURL := strings.TrimSpace(req.URL)
	if rawURL == "" {
		writeError(w, http.StatusBadRequest, "Download URL is required")
		return
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
		writeError(w, http.StatusBadRequest, "Invalid URL: must be a valid HTTP or HTTPS link")
		return
	}

	rawPath := strings.ReplaceAll(req.Path, "\\", "/")
	for _, part := range strings.Split(rawPath, "/") {
		if part == ".." || strings.ContainsRune(part, '\x00') {
			writeError(w, http.StatusBadRequest, "Invalid destination directory")
			return
		}
	}
	targetDir := strings.Trim(rawPath, "/")
	if targetDir == "." {
		targetDir = ""
	}

	releaseMutation, err := rooted.TryMutation(h.manager.GetWorkingDir(serverID))
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	defer releaseMutation()

	root, err := rooted.Open(h.manager.GetWorkingDir(serverID))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer root.Close()

	if targetDir != "" {
		if st, err := root.Lstat(targetDir); err != nil || !st.IsDir() {
			writeError(w, http.StatusBadRequest, "Target directory does not exist")
			return
		}
	}

	httpReq, err := http.NewRequestWithContext(r.Context(), http.MethodGet, rawURL, nil)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to construct request: %v", err))
		return
	}
	httpReq.Header.Set("User-Agent", "NeoPanel/1.0 (+https://github.com/Pekly/NeoPanel)")

	resp, err := urlDownloadClient.Do(httpReq)
	if err != nil {
		writeError(w, http.StatusBadGateway, fmt.Sprintf("Failed to download from URL: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		writeError(w, http.StatusBadGateway, fmt.Sprintf("Remote server responded with HTTP %d (%s)", resp.StatusCode, resp.Status))
		return
	}

	filename := strings.TrimSpace(req.Filename)
	if filename != "" {
		filename = filepath.Base(filepath.Clean(filename))
	} else {
		filename = extractFilenameFromURL(rawURL, resp)
	}

	if !rooted.Basename(filename) {
		writeError(w, http.StatusBadRequest, "Invalid target filename derived from URL")
		return
	}

	destRel := filepath.Join(targetDir, filename)
	fullPath, err := root.Path(destRel)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if !req.Overwrite {
		if _, err := root.Lstat(fullPath); err == nil {
			writeError(w, http.StatusConflict, fmt.Sprintf("File %q already exists; check 'Overwrite' to replace it", filename))
			return
		}
	}

	// Enforce server disk limit
	srv, srvErr := h.db.GetServer(serverID)
	maxAllowedBytes := int64(4096 << 20) // 4 GiB max file download
	if srvErr == nil && srv != nil && srv.DiskLimitMB > 0 {
		curBytes := runner.CalculateDirSizeMB(h.manager.GetWorkingDir(serverID)) * (1 << 20)
		remBytes := srv.DiskLimitMB*(1<<20) - curBytes
		if remBytes <= 0 {
			writeError(w, http.StatusInsufficientStorage, fmt.Sprintf("Download rejected: server disk limit exceeded (%d MB limit)", srv.DiskLimitMB))
			return
		}
		if remBytes < maxAllowedBytes {
			maxAllowedBytes = remBytes
		}
	}

	if resp.ContentLength > 0 && resp.ContentLength > maxAllowedBytes {
		writeError(w, http.StatusInsufficientStorage, fmt.Sprintf("Download rejected: file size (%s) exceeds available server disk space", formatBytesAudit(resp.ContentLength)))
		return
	}

	written, err := root.Copy(r.Context(), fullPath, resp.Body, maxAllowedBytes, 0644)
	if err != nil {
		if os.IsExist(err) {
			writeError(w, http.StatusConflict, "Destination file already exists")
		} else {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to write downloaded file: %v", err))
		}
		return
	}

	logAudit(h.db, r, "file_upload_url", fmt.Sprintf("Downloaded %q (%s) from %s into %s", filename, formatBytesAudit(written), rawURL, auditServerLabel(h.db, serverID)))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":  fmt.Sprintf("File %s downloaded successfully", filename),
		"filename": filename,
		"path":     destRel,
		"size":     written,
	})
}
