package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Pekly/NeoPanel/internal/chunker"
	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
)

type worldConversionHandler struct {
	db        *database.DB
	converter *runner.WorldConverter
}

func newWorldConversionHandler(cfg *config.Config, db *database.DB, manager *runner.ServerManager) *worldConversionHandler {
	return &worldConversionHandler{db: db, converter: runner.NewWorldConverter(manager, filepath.Join(cfg.Paths.DataDir, "tools", "chunker"), func(id string) bool { return db.FeatureEnabled(id, "world_conversion") })}
}

// Permissions are enforced here as well as at the router. Module disable is
// intentionally NOT a blanket route middleware: status/cancel remain available.
func (h *worldConversionHandler) server(w http.ResponseWriter, r *http.Request, permission string) *database.Server {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if status := authorizeServer(h.db, claims, r.PathValue("id"), permission); status != 0 {
		writeError(w, status, "Not authorized for this world conversion action")
		return nil
	}
	srv, err := h.db.GetServer(r.PathValue("id"))
	if err != nil || srv == nil {
		writeError(w, 404, "Server not found")
		return nil
	}
	if srv.NodeID != "" && srv.NodeID != "local" {
		writeError(w, 409, "Open world conversion on this server's hosting node")
		return nil
	}
	return srv
}
func (h *worldConversionHandler) Status(w http.ResponseWriter, r *http.Request) {
	srv := h.server(w, r, "files.view")
	if srv == nil {
		return
	}
	jobs := h.converter.List(srv.ID)
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	canReplace := authorizeServer(h.db, claims, srv.ID, "files.view+files.edit+backup.manage") == 0
	for i := range jobs {
		if !canReplace {
			jobs[i].CanReplace = false
			jobs[i].ReplaceToken = ""
		}
		if claims == nil || claims.Role != "admin" {
			jobs[i].Log = "" // Converter diagnostics can contain internal host paths.
			if jobs[i].SafetyCopy != "" {
				jobs[i].SafetyCopy = "Original world preserved for administrator recovery"
			}
			if jobs[i].Error != "" {
				jobs[i].Error = "Conversion did not complete safely. Original data was not intentionally removed; ask an administrator to inspect the job details."
			}
		}
	}
	reason := runner.WorldConversionSupport(srv)
	if reason != "" && (claims == nil || claims.Role != "admin") {
		reason = "World conversion is not available on this host."
	}
	writeJSON(w, 200, map[string]any{"enabled": h.db.FeatureEnabled(srv.ID, "world_conversion"), "supported": reason == "", "unavailable_reason": reason, "formats": chunker.Formats(), "jobs": jobs, "warning": runner.WorldConversionWarning, "chunker_version": chunker.Version, "max_bytes": chunker.MaxBytes, "max_entries": chunker.MaxEntries, "timeout_seconds": 1800})
}
func (h *worldConversionHandler) Start(w http.ResponseWriter, r *http.Request) {
	srv := h.server(w, r, "files.view+files.edit")
	if srv == nil {
		return
	}
	if !h.db.FeatureEnabled(srv.ID, "world_conversion") {
		writeError(w, 403, "World conversion is disabled")
		return
	}
	var req runner.WorldConversionRequest
	if !decodeWorldConversion(w, r, &req) {
		return
	}
	job, err := h.converter.Start(srv, req)
	if err != nil {
		writeWorldConversionError(w, r, err)
		return
	}
	writeJSON(w, 202, job)
}
func (h *worldConversionHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	srv := h.server(w, r, "files.view+files.edit")
	if srv == nil {
		return
	}
	if err := h.converter.Cancel(srv.ID, r.PathValue("jobId")); err != nil {
		writeWorldConversionError(w, r, err)
		return
	}
	writeJSON(w, 202, map[string]string{"message": "Cancellation requested; the server stays stopped until cleanup finishes"})
}
func (h *worldConversionHandler) Replace(w http.ResponseWriter, r *http.Request) {
	srv := h.server(w, r, "files.view+files.edit+backup.manage")
	if srv == nil {
		return
	}
	if !h.db.FeatureEnabled(srv.ID, "world_conversion") {
		writeError(w, 403, "World conversion is disabled")
		return
	}
	var req runner.WorldConversionReplaceRequest
	if !decodeWorldConversion(w, r, &req) {
		return
	}
	job, err := h.converter.Replace(srv, r.PathValue("jobId"), req)
	if err != nil {
		writeWorldConversionError(w, r, err)
		return
	}
	writeJSON(w, 202, job)
}
func decodeWorldConversion(w http.ResponseWriter, r *http.Request, dst any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeError(w, 400, "Invalid world conversion request")
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		writeError(w, 400, "Expected one world conversion request")
		return false
	}
	return true
}
func writeWorldConversionError(w http.ResponseWriter, r *http.Request, err error) {
	status := 409
	message := err.Error()
	if errors.Is(err, os.ErrNotExist) {
		status = 404
		message = "World or conversion job not found"
	}
	var pathError *os.PathError
	if errors.As(err, &pathError) {
		claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
		if claims == nil || claims.Role != "admin" {
			message = "World files could not be accessed safely; ask an administrator to inspect storage"
		}
	}
	writeError(w, status, message)
}
