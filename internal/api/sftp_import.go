package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
	"github.com/Pekly/NeoPanel/internal/sftpimport"
)

const sftpImportPermission = "files.view+files.edit+sftp.access"
const sftpImportExpiryHeader = "X-NeoPanel-Import-Expires"
const sftpImportBodyLimit = 128 << 10
const sftpImportMaxBytes int64 = 100 << 30
const sftpImportDefaultBytes int64 = 10 << 30
const sftpImportMaxJobs = 2
const sftpImportHistoryLimit = 128

// Credentials belong to the transfer goroutine, never to the retained job record.
type sftpImportJob struct {
	sftpimport.Progress
	ID          string    `json:"id"`
	ServerID    string    `json:"server_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Message     string    `json:"message"`
	UserMessage string    `json:"user_message,omitempty"`
	Error       string    `json:"error"`
	SafetyCopy  string    `json:"safety_copy"`
}

type sftpImportRecord struct {
	job    sftpImportJob
	userID int
	cancel context.CancelFunc
}

type sftpImportHandler struct {
	cfg      *config.Config
	db       *database.DB
	manager  *runner.ServerManager
	mu       sync.Mutex
	jobs     map[string]*sftpImportRecord
	probes   chan struct{}
	limiter  *IPRateLimiter
	transfer func(context.Context, sftpimport.Options, []string, string, func(sftpimport.Progress)) error
}

func newSFTPImportHandler(cfg *config.Config, db *database.DB, manager *runner.ServerManager) *sftpImportHandler {
	limiter := NewIPRateLimiter(12, 12)
	limiter.burstCapacity = 6 // Key lookup, preview, folder changes, then Start.
	return &sftpImportHandler{
		cfg: cfg, db: db, manager: manager,
		jobs: make(map[string]*sftpImportRecord), probes: make(chan struct{}, 2),
		limiter: limiter, transfer: sftpimport.Transfer,
	}
}

func (h *sftpImportHandler) allowAttempt(r *http.Request) bool {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil {
		return false
	}
	// Workers see the proxy's IP. Use the authenticated account, not that shared IP.
	ctx := context.WithValue(r.Context(), clientIPContextKey{}, fmt.Sprintf("import-user:%d", claims.UserID))
	return h.limiter.Allow(r.WithContext(ctx))
}

func importActive(state string) bool {
	return state == "queued" || state == "scanning" || state == "transferring" || state == "publishing"
}

func (h *sftpImportHandler) localServer(w http.ResponseWriter, r *http.Request) *database.Server {
	srv, err := h.db.GetServer(r.PathValue("id"))
	if err != nil || srv == nil {
		writeError(w, http.StatusNotFound, "Server not found")
		return nil
	}
	if srv.NodeID != "" && srv.NodeID != "local" {
		writeError(w, http.StatusConflict, "Open this server on its hosting node to import files")
		return nil
	}
	return srv
}

func importByteLimit(srv *database.Server) int64 {
	limit := sftpImportMaxBytes
	if srv.DiskLimitMB > 0 && srv.DiskLimitMB < limit/(1<<20) {
		limit = srv.DiskLimitMB * (1 << 20)
	}
	return limit
}

// Eligibility is informational; the runner repeats the checks under its operation gate.
func (h *sftpImportHandler) eligibility(id string) (bool, string) {
	if h.manager.GetStatus(id) != "offline" {
		return false, "Stop your server before importing files."
	}
	dir := h.manager.GetWorkingDir(id)
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return false, "Server files are unavailable. Contact support."
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return false, "Could not check the server directory."
	}
	defer root.Close()
	f, err := root.Open(".")
	if err != nil {
		return false, "Could not check the server directory."
	}
	defer f.Close()
	names, err := f.Readdirnames(1)
	if len(names) != 0 || err != io.EOF {
		return false, "Your server still has files. Back up and move or remove them before importing."
	}
	return true, "Ready to import into this empty, stopped server."
}

func (h *sftpImportHandler) pruneLocked() {
	for id, rec := range h.jobs {
		if !importActive(rec.job.State) && time.Since(rec.job.UpdatedAt) > time.Hour {
			delete(h.jobs, id)
		}
	}
}

func (h *sftpImportHandler) status(w http.ResponseWriter, r *http.Request) {
	srv := h.localServer(w, r)
	if srv == nil {
		return
	}
	eligible, reason := h.eligibility(srv.ID)
	var job *sftpImportJob
	h.mu.Lock()
	h.pruneLocked()
	if rec := h.jobs[srv.ID]; rec != nil {
		copy := importJobForViewer(r, rec.job)
		job = &copy
		if importActive(copy.State) {
			eligible, reason = false, "An import is already in progress."
		}
	}
	h.mu.Unlock()
	limit := importByteLimit(srv)
	writeJSON(w, http.StatusOK, map[string]any{"job": job, "eligible": eligible, "reason": reason, "limits": map[string]any{"max_bytes": limit, "default_bytes": min(sftpImportDefaultBytes, limit), "max_entries": 100000, "max_active": sftpImportMaxJobs}})
}

func decodeSFTPImport(w http.ResponseWriter, r *http.Request, dst any) bool {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, sftpImportBodyLimit))
	if err != nil {
		writeError(w, 413, "Import request exceeds 128 KiB")
		return false
	}
	if string(bytes.TrimSpace(body)) == "null" {
		writeError(w, 400, "An import options object is required")
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeError(w, 400, "Invalid import options")
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		writeError(w, 400, "Expected one import options object")
		return false
	}
	return true
}

func (h *sftpImportHandler) options(w http.ResponseWriter, r *http.Request, srv *database.Server) (sftpimport.Options, bool) {
	var opts sftpimport.Options
	if !decodeSFTPImport(w, r, &opts) {
		return opts, false
	}
	limit := importByteLimit(srv)
	if opts.MaxBytes == 0 {
		opts.MaxBytes = min(limit, sftpImportDefaultBytes)
	}
	if opts.MaxBytes > limit {
		writeError(w, 400, "Import size exceeds this server's disk allowance")
		return opts, false
	}
	opts, err := sftpimport.Validate(opts)
	if err != nil {
		writeError(w, 400, importErrorForViewer(r, err.Error()))
		return opts, false
	}
	return opts, true
}

func (h *sftpImportHandler) connectionSlot(w http.ResponseWriter, r *http.Request) (func(), bool) {
	if !h.allowAttempt(r) {
		w.Header().Set("Retry-After", "10")
		writeError(w, 429, "Too many SFTP connection attempts; wait before retrying")
		return nil, false
	}
	select {
	case h.probes <- struct{}{}:
		return func() { <-h.probes }, true
	default:
		writeError(w, 429, "SFTP connection checks are busy; retry shortly")
		return nil, false
	}
}

func (h *sftpImportHandler) hostKey(w http.ResponseWriter, r *http.Request) {
	if h.localServer(w, r) == nil {
		return
	}
	var req struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	if !decodeSFTPImport(w, r, &req) {
		return
	}
	release, ok := h.connectionSlot(w, r)
	if !ok {
		return
	}
	defer release()
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	info, err := sftpimport.HostKey(ctx, req.Host, req.Port, h.cfg.Security.SFTPImportAllowedHosts)
	if err != nil {
		writeError(w, 400, importErrorForViewer(r, err.Error()))
		return
	}
	writeJSON(w, 200, info)
}

func (h *sftpImportHandler) probe(w http.ResponseWriter, r *http.Request) {
	srv := h.localServer(w, r)
	if srv == nil {
		return
	}
	authority, err := h.captureAuthority(r)
	if err != nil {
		writeError(w, 401, "Could not verify import credentials")
		return
	}
	opts, ok := h.options(w, r, srv)
	if !ok {
		return
	}
	if !h.authorized(srv.ID, authority) {
		writeError(w, 403, "Import permission has changed; sign in again")
		return
	}
	release, ok := h.connectionSlot(w, r)
	if !ok {
		return
	}
	defer release()
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	preview, err := sftpimport.Probe(ctx, opts, h.cfg.Security.SFTPImportAllowedHosts)
	if err != nil {
		writeError(w, 400, importErrorForViewer(r, err.Error()))
		return
	}
	writeJSON(w, 200, preview)
}

type importAuthority struct {
	claims    JWTClaims
	keyHash   string
	hopHash   string
	hopIsNode bool
	expiresAt time.Time
}

func requestAuthToken(r *http.Request) string {
	if header := r.Header.Get("Authorization"); strings.HasPrefix(header, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	}
	if c, err := r.Cookie("neopanel_token"); err == nil {
		return c.Value
	}
	return ""
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (h *sftpImportHandler) captureAuthority(r *http.Request) (importAuthority, error) {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil {
		return importAuthority{}, errors.New("missing identity")
	}
	a := importAuthority{claims: *claims}
	if claims.ExpiresAt != nil {
		a.expiresAt = claims.ExpiresAt.Time
	}
	if r.Header.Get("X-NeoPanel-User-ID") != "" || claims.UserID == 0 {
		token := requestAuthToken(r)
		a.hopHash = tokenHash(token)
		a.hopIsNode = token != "" && token == strings.TrimSpace(h.cfg.Security.NodeToken)
		if claims.UserID > 0 {
			if values := r.Header.Values(sftpImportExpiryHeader); len(values) != 0 {
				if len(values) != 1 {
					return a, errors.New("invalid forwarded expiry")
				}
				seconds, err := strconv.ParseInt(values[0], 10, 64)
				if err != nil || seconds <= 0 || seconds > 253402300799 {
					return a, errors.New("invalid forwarded expiry")
				}
				a.expiresAt = time.Unix(seconds, 0)
			}
			user, err := h.db.GetUserByID(claims.UserID)
			if err != nil {
				if claims.Role == "admin" && a.hopIsNode {
					a.claims.AuthVersion = 0
				} else {
					return a, err
				}
			} else {
				a.claims.AuthVersion = user.AuthVersion
			}
		}
	} else if claims.APIKey {
		a.keyHash = tokenHash(requestAuthToken(r))
	}
	return a, nil
}

func (h *sftpImportHandler) authorized(id string, a importAuthority) bool {
	if a.hopHash != "" {
		if a.hopIsNode {
			if a.hopHash != tokenHash(strings.TrimSpace(h.cfg.Security.NodeToken)) {
				return false
			}
		} else {
			user, err := h.db.GetUserByAPIKeyHash(a.hopHash)
			if err != nil || user == nil || user.Role != "admin" {
				return false
			}
		}
	}
	fresh := a.claims
	if !a.expiresAt.IsZero() && !time.Now().Before(a.expiresAt) {
		return false
	}
	if fresh.UserID > 0 {
		user, err := h.db.GetUserByID(fresh.UserID)
		if err != nil || user == nil {
			return false
		}
		if a.keyHash != "" {
			keyUser, err := h.db.GetUserByAPIKeyHash(a.keyHash)
			if err != nil || keyUser == nil || keyUser.ID != fresh.UserID {
				return false
			}
		} else if user.AuthVersion != fresh.AuthVersion {
			return false
		}
		if user.Role != "admin" || fresh.Role != "admin" {
			fresh.Role = "user"
		}
	} else if fresh.UserID != 0 || fresh.Role != "admin" || a.hopHash == "" {
		return false
	}
	return authorizeServer(h.db, &fresh, id, sftpImportPermission) == 0
}

func (h *sftpImportHandler) start(w http.ResponseWriter, r *http.Request) {
	srv := h.localServer(w, r)
	if srv == nil {
		return
	}
	authority, err := h.captureAuthority(r)
	if err != nil {
		writeError(w, 401, "Could not verify import credentials")
		return
	}
	opts, ok := h.options(w, r, srv)
	if !ok {
		return
	}
	if !h.authorized(srv.ID, authority) {
		writeError(w, 403, "Import permission has changed; sign in again")
		return
	}
	if eligible, reason := h.eligibility(srv.ID); !eligible {
		writeError(w, 409, reason)
		return
	}
	if !h.allowAttempt(r) {
		w.Header().Set("Retry-After", "10")
		writeError(w, 429, "Too many import attempts; wait before retrying")
		return
	}
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		writeError(w, 503, "Could not create an import job")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(opts.TimeoutMinutes)*time.Minute)
	now := time.Now().UTC()
	rec := &sftpImportRecord{job: sftpImportJob{ID: hex.EncodeToString(raw[:]), ServerID: srv.ID, CreatedAt: now, UpdatedAt: now, Message: "Import queued; source files will not be modified."}, userID: authority.claims.UserID, cancel: cancel}
	rec.job.State = "queued"
	h.mu.Lock()
	h.pruneLocked()
	active := 0
	for _, old := range h.jobs {
		if importActive(old.job.State) {
			active++
			if old.job.ServerID == srv.ID || old.userID == rec.userID {
				h.mu.Unlock()
				cancel()
				writeError(w, 409, "An import is already running for this server or account")
				return
			}
		}
	}
	if active >= sftpImportMaxJobs {
		h.mu.Unlock()
		cancel()
		writeError(w, 429, "Imports are busy. Please try again shortly.")
		return
	}
	if len(h.jobs) >= sftpImportHistoryLimit && h.jobs[srv.ID] == nil {
		var oldest string
		for id, old := range h.jobs {
			if !importActive(old.job.State) && (oldest == "" || old.job.UpdatedAt.Before(h.jobs[oldest].job.UpdatedAt)) {
				oldest = id
			}
		}
		if oldest == "" {
			h.mu.Unlock()
			cancel()
			writeError(w, 429, "Import history is busy")
			return
		}
		delete(h.jobs, oldest)
	}
	h.jobs[srv.ID] = rec
	job := rec.job
	h.mu.Unlock()
	logAudit(h.db, r, "file_import_sftp", fmt.Sprintf("Started SFTP file import for server %s (job %s)", srv.ID, job.ID))
	allowed := append([]string(nil), h.cfg.Security.SFTPImportAllowedHosts...)
	go h.run(ctx, rec, srv, opts, allowed, authority)
	writeJSON(w, http.StatusAccepted, importJobForViewer(r, job))
}

func (h *sftpImportHandler) run(ctx context.Context, rec *sftpImportRecord, srv *database.Server, opts sftpimport.Options, allowed []string, authority importAuthority) {
	defer rec.cancel()
	result, err := h.manager.ImportServerFiles(ctx, srv, func(ctx context.Context, stage string) error {
		if !h.authorized(srv.ID, authority) {
			return errors.New("Import permission was revoked")
		}
		ctx, cancel := context.WithCancel(ctx)
		done := make(chan struct{})
		go func() {
			defer close(done)
			tick := time.NewTicker(time.Second)
			defer tick.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-tick.C:
					if !h.authorized(srv.ID, authority) {
						cancel()
						return
					}
				}
			}
		}()
		defer func() { cancel(); <-done }()
		err := h.transfer(ctx, opts, allowed, stage, func(progress sftpimport.Progress) {
			h.mu.Lock()
			defer h.mu.Unlock()
			rec.job.Progress = progress
			rec.job.UpdatedAt = time.Now().UTC()
		})
		if err != nil {
			return err
		}
		if !h.authorized(srv.ID, authority) {
			return errors.New("Import permission was revoked")
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		h.mu.Lock()
		rec.job.State = "publishing"
		rec.job.UpdatedAt = time.Now().UTC()
		h.mu.Unlock()
		return nil
	})
	h.mu.Lock()
	defer h.mu.Unlock()
	rec.job.UpdatedAt = time.Now().UTC()
	rec.job.CurrentPath = ""
	switch {
	case err == nil:
		rec.job.State = "completed"
		rec.job.SafetyCopy = result.SafetyCopy
		rec.job.Message = "Import complete. Server remains stopped. Check the startup file, runtime and assigned ports before starting."
	case errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded):
		rec.job.State = "cancelled"
		rec.job.Message = "Import cancelled or timed out; staged files were not published."
	default:
		rec.job.State = "failed"
		rec.job.Error = err.Error()
		rec.job.Message = "Import failed. Review the error before retrying."
	}
}

func (h *sftpImportHandler) cancelJob(w http.ResponseWriter, r *http.Request) {
	if h.localServer(w, r) == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	rec := h.jobs[r.PathValue("id")]
	if rec == nil || rec.job.ID != r.PathValue("jobId") {
		writeError(w, 404, "Import job not found; refresh its status")
		return
	}
	if importActive(rec.job.State) {
		rec.cancel()
		rec.job.Message = "Cancellation requested; waiting for the transfer to stop."
		rec.job.UpdatedAt = time.Now().UTC()
	}
	writeJSON(w, 200, importJobForViewer(r, rec.job))
}
