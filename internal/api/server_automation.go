package api

import (
	"errors"
	"io"
	"net/http"

	"github.com/Pekly/NeoPanel/internal/runner"
)

func (h *ServerHandler) GetServerAutomation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	id := r.PathValue("id")
	if status := authorizeServer(h.db, claims, id, ""); status != 0 {
		writeError(w, status, "Server membership required")
		return
	}
	view, err := h.manager.GetServerAutomation(id)
	if err != nil {
		writeError(w, 503, "Unable to read host automation policy")
		return
	}
	view.CanEditJoinToStart = authorizeServer(h.db, claims, id, "settings.edit+server.power") == 0
	writeJSON(w, http.StatusOK, view)
}

func (h *ServerHandler) PatchServerAutomation(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	id := r.PathValue("id")
	if status := authorizeServer(h.db, claims, id, "settings.edit+server.power"); status != 0 {
		writeError(w, status, "Settings edit and server power permissions required")
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 4096))
	if err != nil {
		writeError(w, 400, "Automation request exceeds limit")
		return
	}
	var enabled bool
	if err := decodeWakeObject(body, map[string]any{"join_to_start_enabled": &enabled}); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	if err := h.manager.SetServerWakePreference(id, enabled); err != nil {
		status := http.StatusServiceUnavailable
		if errors.Is(err, runner.ErrServerOperation) {
			status = http.StatusConflict
		}
		writeError(w, status, "Unable to change join-to-start preference")
		return
	}
	h.GetServerAutomation(w, r)
}
