package api

import (
	"errors"
	"net/http"

	"github.com/Pekly/NeoPanel/internal/runner"
)

// ServerArtifact runs on the hosting worker, behind files.view authorization.
func (h *ServerHandler) ServerArtifact(w http.ResponseWriter, r *http.Request) {
	srv, err := h.db.GetServer(getURLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "Server not found")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	refresh := r.URL.Query().Get("refresh")
	if refresh != "" && refresh != "true" && refresh != "false" {
		writeError(w, http.StatusBadRequest, "Invalid artifact refresh mode")
		return
	}
	info, err := h.manager.InspectArtifact(r.Context(), srv, refresh == "true")
	if errors.Is(err, runner.ErrArtifactBusy) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error(), "code": "server_artifact_busy"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Server JAR identity could not be saved. Retry inspection.", "code": "server_artifact_unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, info)
}
