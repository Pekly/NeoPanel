package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
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

type URLDownloadJob struct {
	ID         string    `json:"id"`
	ServerID   string    `json:"server_id"`
	URL        string    `json:"url"`
	Filename   string    `json:"filename"`
	Path       string    `json:"path"`
	State      string    `json:"state"`
	BytesDone  int64     `json:"bytes_done"`
	BytesTotal int64     `json:"bytes_total"`
	SpeedBPS   int64     `json:"speed_bps"`
	Percent    int       `json:"percent"`
	Error      string    `json:"error,omitempty"`
	Message    string    `json:"message,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type urlDownloadRecord struct {
	job    URLDownloadJob
	cancel context.CancelFunc
}

var urlDownloadClient = &http.Client{
	Timeout: 0, // Context-controlled timeout for large downloads
	Transport: &http.Transport{
		ResponseHeaderTimeout: 30 * time.Second,
		TLSHandshakeTimeout:   15 * time.Second,
	},
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

func newURLJobID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return "url-" + hex.EncodeToString(b[:])
}

func (h *FileHandler) pruneURLJobsLocked() {
	if h.urlJobs == nil {
		h.urlJobs = make(map[string]*urlDownloadRecord)
		return
	}
	cutoff := time.Now().Add(-15 * time.Minute)
	for id, rec := range h.urlJobs {
		st := rec.job.State
		if (st == "completed" || st == "failed" || st == "cancelled") && rec.job.UpdatedAt.Before(cutoff) {
			delete(h.urlJobs, id)
		}
	}
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

	// Validate working directory & target directory exist
	workDir := h.manager.GetWorkingDir(serverID)
	root, err := rooted.Open(workDir)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if targetDir != "" {
		if st, err := root.Lstat(targetDir); err != nil || !st.IsDir() {
			root.Close()
			writeError(w, http.StatusBadRequest, "Target directory does not exist")
			return
		}
	}

	explicitFilename := strings.TrimSpace(req.Filename)
	if explicitFilename != "" {
		explicitFilename = filepath.Base(filepath.Clean(explicitFilename))
		if !rooted.Basename(explicitFilename) {
			root.Close()
			writeError(w, http.StatusBadRequest, "Invalid target filename")
			return
		}
		destRel := filepath.Join(targetDir, explicitFilename)
		if fullPath, err := root.Path(destRel); err != nil {
			root.Close()
			writeError(w, http.StatusBadRequest, err.Error())
			return
		} else if !req.Overwrite {
			if _, err := root.Lstat(fullPath); err == nil {
				root.Close()
				writeError(w, http.StatusConflict, fmt.Sprintf("File %q already exists; check 'Overwrite' to replace it", explicitFilename))
				return
			}
		}
	}
	root.Close()

	// Enforce server disk limit upfront
	srv, srvErr := h.db.GetServer(serverID)
	maxAllowedBytes := int64(4096 << 20) // 4 GiB max file download
	if srvErr == nil && srv != nil && srv.DiskLimitMB > 0 {
		curBytes := runner.CalculateDirSizeMB(workDir) * (1 << 20)
		remBytes := srv.DiskLimitMB*(1<<20) - curBytes
		if remBytes <= 0 {
			writeError(w, http.StatusInsufficientStorage, fmt.Sprintf("Download rejected: server disk limit exceeded (%d MB limit)", srv.DiskLimitMB))
			return
		}
		if remBytes < maxAllowedBytes {
			maxAllowedBytes = remBytes
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	now := time.Now()
	initialFilename := explicitFilename
	if initialFilename == "" {
		initialFilename = extractFilenameFromURL(rawURL, nil)
	}

	job := URLDownloadJob{
		ID:        newURLJobID(),
		ServerID:  serverID,
		URL:       rawURL,
		Filename:  initialFilename,
		Path:      targetDir,
		State:     "downloading",
		Message:   "Connecting to remote server...",
		CreatedAt: now,
		UpdatedAt: now,
	}

	h.urlJobsMu.Lock()
	h.pruneURLJobsLocked()
	h.urlJobs[job.ID] = &urlDownloadRecord{job: job, cancel: cancel}
	h.urlJobsMu.Unlock()

	var actor *JWTClaims
	if c, ok := r.Context().Value(contextUserKey).(*JWTClaims); ok && c != nil {
		cp := *c
		actor = &cp
	}
	clientIP := getClientIP(r)

	go h.runURLDownload(ctx, cancel, job.ID, serverID, workDir, rawURL, explicitFilename, targetDir, req.Overwrite, maxAllowedBytes, actor, clientIP)

	writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"job_id":  job.ID,
		"status":  "downloading",
		"message": fmt.Sprintf("Download started for %s", initialFilename),
		"job":     job,
	})
}

func (h *FileHandler) updateURLJob(jobID string, fn func(*URLDownloadJob)) {
	h.urlJobsMu.Lock()
	defer h.urlJobsMu.Unlock()
	if rec, ok := h.urlJobs[jobID]; ok {
		fn(&rec.job)
		rec.job.UpdatedAt = time.Now()
	}
}

func (h *FileHandler) failURLJob(jobID, msg string) {
	h.updateURLJob(jobID, func(j *URLDownloadJob) {
		j.State = "failed"
		j.SpeedBPS = 0
		j.Error = msg
		j.Message = msg
	})
}

type urlProgressReader struct {
	r          io.Reader
	onProgress func(bytesDone int64, speedBPS int64)
	done       int64
	lastBytes  int64
	lastTick   time.Time
}

func (p *urlProgressReader) Read(buf []byte) (int, error) {
	n, err := p.r.Read(buf)
	if n > 0 {
		p.done += int64(n)
		now := time.Now()
		elapsed := now.Sub(p.lastTick)
		if elapsed >= 250*time.Millisecond {
			delta := p.done - p.lastBytes
			bps := int64(float64(delta) / elapsed.Seconds())
			p.lastBytes = p.done
			p.lastTick = now
			if p.onProgress != nil {
				p.onProgress(p.done, bps)
			}
		}
	}
	return n, err
}

func (h *FileHandler) runURLDownload(
	ctx context.Context,
	cancel context.CancelFunc,
	jobID, serverID, workDir, rawURL, explicitFilename, targetDir string,
	overwrite bool,
	maxAllowedBytes int64,
	actor *JWTClaims,
	clientIP string,
) {
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		h.failURLJob(jobID, fmt.Sprintf("Failed to construct request: %v", err))
		return
	}
	httpReq.Header.Set("User-Agent", "Mozilla/5.0 (compatible; NeoPanel/1.0; +https://github.com/Pekly/NeoPanel)")
	httpReq.Header.Set("Accept", "*/*")

	resp, err := urlDownloadClient.Do(httpReq)
	if err != nil {
		if ctx.Err() == context.Canceled {
			h.updateURLJob(jobID, func(j *URLDownloadJob) {
				j.State = "cancelled"
				j.SpeedBPS = 0
				j.Error = "Download cancelled"
				j.Message = "Download cancelled"
			})
			return
		}
		h.failURLJob(jobID, fmt.Sprintf("Failed to connect to URL: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		h.failURLJob(jobID, fmt.Sprintf("Remote server returned HTTP %d (%s)", resp.StatusCode, resp.Status))
		return
	}

	filename := explicitFilename
	if filename == "" {
		filename = extractFilenameFromURL(rawURL, resp)
	}
	if !rooted.Basename(filename) {
		h.failURLJob(jobID, "Invalid target filename derived from URL")
		return
	}

	totalBytes := resp.ContentLength
	if totalBytes < 0 {
		totalBytes = 0
	}
	if totalBytes > 0 && totalBytes > maxAllowedBytes {
		h.failURLJob(jobID, fmt.Sprintf("Download rejected: file size (%s) exceeds available server disk space", formatBytesAudit(totalBytes)))
		return
	}

	h.updateURLJob(jobID, func(j *URLDownloadJob) {
		j.Filename = filename
		j.BytesTotal = totalBytes
		j.Message = fmt.Sprintf("Downloading %s...", filename)
	})

	releaseMutation, err := rooted.TryMutation(workDir)
	if err != nil {
		h.failURLJob(jobID, err.Error())
		return
	}
	defer releaseMutation()

	root, err := rooted.Open(workDir)
	if err != nil {
		h.failURLJob(jobID, err.Error())
		return
	}
	defer root.Close()

	destRel := filepath.Join(targetDir, filename)
	fullPath, err := root.Path(destRel)
	if err != nil {
		h.failURLJob(jobID, err.Error())
		return
	}

	if !overwrite {
		if _, err := root.Lstat(fullPath); err == nil {
			h.failURLJob(jobID, fmt.Sprintf("File %q already exists; check 'Overwrite' to replace it", filename))
			return
		}
	}

	pr := &urlProgressReader{
		r:        resp.Body,
		lastTick: time.Now(),
		onProgress: func(done, bps int64) {
			h.updateURLJob(jobID, func(j *URLDownloadJob) {
				j.BytesDone = done
				j.SpeedBPS = bps
				if j.BytesTotal > 0 {
					pct := int((done * 100) / j.BytesTotal)
					if pct > 99 {
						pct = 99
					}
					j.Percent = pct
				}
			})
		},
	}

	written, err := root.Copy(ctx, fullPath, pr, maxAllowedBytes, 0644)
	if err != nil {
		if ctx.Err() == context.Canceled {
			h.updateURLJob(jobID, func(j *URLDownloadJob) {
				j.State = "cancelled"
				j.SpeedBPS = 0
				j.Error = "Download cancelled"
				j.Message = "Download cancelled"
			})
			return
		}
		if os.IsExist(err) {
			h.failURLJob(jobID, "Destination file already exists")
		} else {
			h.failURLJob(jobID, fmt.Sprintf("Download failed: %v", err))
		}
		return
	}

	h.updateURLJob(jobID, func(j *URLDownloadJob) {
		j.State = "completed"
		j.BytesDone = written
		if j.BytesTotal == 0 {
			j.BytesTotal = written
		}
		j.Percent = 100
		j.SpeedBPS = 0
		j.Message = fmt.Sprintf("Downloaded %s (%s)", filename, formatBytesAudit(written))
	})

	if h.db != nil {
		uid := 0
		if actor != nil {
			uid = actor.UserID
		}
		_ = h.db.LogActivity(uid, "file_upload_url", fmt.Sprintf("Downloaded %q (%s) from %s into %s", filename, formatBytesAudit(written), rawURL, auditServerLabel(h.db, serverID)), clientIP)
	}
}

func (h *FileHandler) GetURLDownloads(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	if serverID == "" {
		writeError(w, http.StatusBadRequest, "Server ID is required")
		return
	}

	h.urlJobsMu.Lock()
	h.pruneURLJobsLocked()
	var list []URLDownloadJob
	for _, rec := range h.urlJobs {
		if rec.job.ServerID == serverID {
			list = append(list, rec.job)
		}
	}
	h.urlJobsMu.Unlock()

	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"jobs": list,
	})
}

func (h *FileHandler) CancelURLDownload(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	jobID := getURLParam(r, "jobId")
	if serverID == "" || jobID == "" {
		writeError(w, http.StatusBadRequest, "Server ID and Job ID are required")
		return
	}

	h.urlJobsMu.Lock()
	rec, ok := h.urlJobs[jobID]
	if !ok || rec.job.ServerID != serverID {
		h.urlJobsMu.Unlock()
		writeError(w, http.StatusNotFound, "Download job not found")
		return
	}
	cancel := rec.cancel
	h.urlJobsMu.Unlock()

	if cancel != nil {
		cancel()
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "cancelled",
		"message": "Download cancellation requested",
	})
}
