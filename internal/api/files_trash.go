package api

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/Pekly/NeoPanel/internal/rooted"
	"github.com/Pekly/NeoPanel/internal/trash"
)

func (h *FileHandler) deleteFile(serverID string, root *rooted.Root, name string, permanent bool) error {
	if permanent {
		return root.RemoveAll(name)
	}
	return h.manager.Trash().Put(serverID, root, name)
}

func (h *FileHandler) ListTrash(w http.ResponseWriter, r *http.Request) {
	items, err := h.manager.Trash().List(getURLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list trash")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

func trashError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	switch {
	case errors.Is(err, trash.ErrCollision):
		status = http.StatusConflict
	case errors.Is(err, trash.ErrExpired):
		status = http.StatusGone
	case os.IsNotExist(err):
		status = http.StatusNotFound
	}
	writeError(w, status, err.Error())
}

func (h *FileHandler) RestoreTrash(w http.ResponseWriter, r *http.Request) {
	serverID, id := getURLParam(r, "id"), getURLParam(r, "trashId")
	release, err := h.manager.BeginServerOperation(serverID)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	defer release()
	// The body has no path options: restoring always uses the recorded path.
	root, err := rooted.Open(h.manager.GetWorkingDir(serverID))
	if err != nil {
		trashError(w, err)
		return
	}
	defer root.Close()
	if err := h.manager.Trash().Restore(serverID, id, root); err != nil {
		trashError(w, err)
		return
	}
	logAudit(h.db, r, "file_trash_restore", fmt.Sprintf("Restored trash entry %s on %s", id, auditServerLabel(h.db, serverID)))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Restored successfully"})
}

func (h *FileHandler) PurgeTrash(w http.ResponseWriter, r *http.Request) {
	serverID, id := getURLParam(r, "id"), getURLParam(r, "trashId")
	if err := h.manager.Trash().Purge(serverID, id); err != nil {
		trashError(w, err)
		return
	}
	logAudit(h.db, r, "file_trash_purge", fmt.Sprintf("Purged trash entry %s on %s", id, auditServerLabel(h.db, serverID)))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Permanently deleted"})
}
