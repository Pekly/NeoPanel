package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Pekly/NeoPanel/internal/database"
)

type SubuserHandler struct {
	db *database.DB
}

func NewSubuserHandler(db *database.DB) *SubuserHandler {
	return &SubuserHandler{db: db}
}

// ListSubusers returns all subusers for a server
func (h *SubuserHandler) ListSubusers(w http.ResponseWriter, r *http.Request) {
	serverID := extractIDFromPath(r.URL.Path, "/api/servers/", "/subusers")
	if serverID == "" {
		writeError(w, http.StatusBadRequest, "Missing server ID")
		return
	}

	subusers, err := h.db.ListServerSubusers(serverID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list subusers: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"subusers": subusers,
	})
}

// AddSubuser creates a new subuser association for a server
func (h *SubuserHandler) AddSubuser(w http.ResponseWriter, r *http.Request) {
	serverID := extractIDFromPath(r.URL.Path, "/api/servers/", "/subusers")
	if serverID == "" {
		writeError(w, http.StatusBadRequest, "Missing server ID")
		return
	}

	var req struct {
		UserEmail   string   `json:"user_email"`
		Username    string   `json:"username"`
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	lookup := strings.TrimSpace(req.UserEmail)
	if lookup == "" {
		lookup = strings.TrimSpace(req.Username)
	}
	if lookup == "" {
		writeError(w, http.StatusBadRequest, "User email or username is required")
		return
	}

	user, err := h.db.GetUserByEmailOrUsername(lookup)
	if err != nil || user == nil {
		writeError(w, http.StatusNotFound, "User not found with the given email or username")
		return
	}

	// Verify server exists
	srv, err := h.db.GetServer(serverID)
	if err != nil || srv == nil {
		writeError(w, http.StatusNotFound, "Server not found")
		return
	}

	if srv.OwnerID == user.ID {
		writeError(w, http.StatusBadRequest, "Cannot add server owner as a subuser")
		return
	}

	subuser, err := h.db.AddServerSubuser(serverID, user.ID, req.Permissions)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to add subuser: "+err.Error())
		return
	}

	subuser.Username = user.Username

	_ = h.db.CreateAuditLog(user.Username, "subuser.add", "Added subuser "+user.Username+" to server "+srv.Name, getClientIP(r))

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"status":  "success",
		"subuser": subuser,
	})
}

// UpdateSubuser updates permissions for an existing subuser
func (h *SubuserHandler) UpdateSubuser(w http.ResponseWriter, r *http.Request) {
	serverID := extractIDFromPath(r.URL.Path, "/api/servers/", "/subusers/")
	parts := strings.Split(r.URL.Path, "/subusers/")
	if len(parts) != 2 || parts[1] == "" {
		writeError(w, http.StatusBadRequest, "Missing subuser ID")
		return
	}
	subuserID := parts[1]
	subuserIntID, err := strconv.Atoi(subuserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid subuser ID")
		return
	}

	var req struct {
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.db.UpdateServerSubuser(serverID, subuserIntID, req.Permissions); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Subuser not found")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to update subuser: "+err.Error())
		}
		return
	}

	_ = h.db.CreateAuditLog("admin", "subuser.update", "Updated permissions for subuser "+subuserID+" on server "+serverID, getClientIP(r))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
	})
}

// DeleteSubuser removes a subuser association
func (h *SubuserHandler) DeleteSubuser(w http.ResponseWriter, r *http.Request) {
	serverID := extractIDFromPath(r.URL.Path, "/api/servers/", "/subusers/")
	parts := strings.Split(r.URL.Path, "/subusers/")
	if len(parts) != 2 || parts[1] == "" {
		writeError(w, http.StatusBadRequest, "Missing subuser ID")
		return
	}
	subuserID := parts[1]
	subuserIntID, err := strconv.Atoi(subuserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid subuser ID")
		return
	}

	if err := h.db.DeleteServerSubuser(serverID, subuserIntID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Subuser not found")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to delete subuser: "+err.Error())
		}
		return
	}

	_ = h.db.CreateAuditLog("admin", "subuser.delete", "Removed subuser "+subuserID, getClientIP(r))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
	})
}

// GetMyPermissions returns current user's effective permissions for a server
func (h *SubuserHandler) GetMyPermissions(w http.ResponseWriter, r *http.Request) {
	serverID := extractIDFromPath(r.URL.Path, "/api/servers/", "/permissions")
	if serverID == "" {
		writeError(w, http.StatusBadRequest, "Missing server ID")
		return
	}

	user := getUserFromContext(r)
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Administrators have all permissions
	if user.Role == "admin" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"is_admin": true,
			"is_owner": true,
			"permissions": []string{
				"console.view", "console.command", "server.power",
				"files.view", "files.edit", "files.delete",
				"settings.view", "settings.edit", "settings.ram",
				"database.manage", "backup.manage", "subusers.manage", "schedules.manage",
			},
		})
		return
	}

	srv, err := h.db.GetServer(serverID)
	if err != nil || srv == nil {
		writeError(w, http.StatusNotFound, "Server not found")
		return
	}

	if srv.OwnerID == user.ID {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"is_admin": false,
			"is_owner": true,
			"permissions": []string{
				"console.view", "console.command", "server.power",
				"files.view", "files.edit", "files.delete",
				"settings.view", "settings.edit", "settings.ram",
				"database.manage", "backup.manage", "subusers.manage", "schedules.manage",
			},
		})
		return
	}

	perms, err := h.db.GetUserServerPermissions(serverID, user.ID)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"is_admin":    false,
			"is_owner":    false,
			"permissions": []string{},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"is_admin":    false,
		"is_owner":    false,
		"permissions": perms,
	})
}
