package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/Pekly/NeoPanel/internal/database"
)

// CanCreateServer must be called before a master proxies a create request.
// It uses current database permissions, not the role embedded in a session JWT.
func (h *AuthHandler) CanCreateServer(r *http.Request) (bool, error) {
	return canCreateServer(h.db, r)
}

func canCreateServer(db *database.DB, r *http.Request) (bool, error) {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil {
		return false, nil
	}
	// A synchronized worker identity does not carry the master's custom roles.
	// Only trusted forwarding middleware may set this non-serialized claim after
	// authenticating the master and its explicit authorized-create header.
	if r.Header.Get("X-NeoPanel-User-ID") != "" {
		return claims.UserID > 0 && claims.AuthorizedServerCreate, nil
	}
	// Configured native node tokens are authenticated machine identities, not users.
	if claims.UserID == 0 && claims.Role == "admin" {
		return true, nil
	}
	if claims.UserID <= 0 {
		return false, nil
	}
	return db.UserCanCreateServer(claims.UserID)
}

func (h *AuthHandler) requirePanelRoleAdmin(w http.ResponseWriter, r *http.Request) bool {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return false
	}
	if claims.Role != "admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Administrator access required"})
		return false
	}
	if claims.UserID == 0 {
		return true
	} // authenticated node machine identity
	user, err := h.db.GetUserByID(claims.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unable to resolve administrator identity"})
		return false
	}
	if user.Role != "admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Administrator access required"})
		return false
	}
	return true
}

func panelRoleError(w http.ResponseWriter, err error) {
	status, message := http.StatusInternalServerError, "Panel role storage error"
	switch {
	case errors.Is(err, database.ErrInvalidPanelRole):
		status, message = http.StatusBadRequest, err.Error()
	case errors.Is(err, database.ErrPanelRoleConflict), errors.Is(err, database.ErrPanelRoleAssigned):
		status, message = http.StatusConflict, err.Error()
	case errors.Is(err, sql.ErrNoRows):
		status, message = http.StatusNotFound, "User or panel role not found"
	}
	writeJSON(w, status, map[string]string{"error": message})
}

// Strict decoding prevents built-in role, identity, and permission injection.
func decodePanelRoleBody(w http.ResponseWriter, r *http.Request, target interface{}) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return false
	}
	if err := decoder.Decode(new(interface{})); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Expected one JSON object"})
		return false
	}
	return true
}

func (h *AuthHandler) ListPanelRoles(w http.ResponseWriter, r *http.Request) {
	if !h.requirePanelRoleAdmin(w, r) {
		return
	}
	roles, err := h.db.ListPanelRoles()
	if err != nil {
		panelRoleError(w, err)
		return
	}
	// Built-in roles are informational; their IDs cannot be assigned or mutated
	// through this API. The users.role column remains their sole authority.
	roles = append([]*database.PanelRole{
		{ID: "user", Name: "user", Description: "Standard panel account; no server creation by default", Permissions: []string{}, Builtin: true},
		{ID: "admin", Name: "admin", Description: "Panel administrator", Permissions: []string{database.PermissionServersCreate}, Builtin: true},
	}, roles...)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"roles":       roles,
		"permissions": []map[string]string{{"id": database.PermissionServersCreate, "label": "Create servers", "description": "Create new servers without granting administrator or access to other users' servers."}},
	})
}

func readPanelRole(w http.ResponseWriter, r *http.Request) (*database.PanelRole, bool) {
	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Color       string   `json:"color"`
		Permissions []string `json:"permissions"`
	}
	if !decodePanelRoleBody(w, r, &req) {
		return nil, false
	}
	role := &database.PanelRole{Name: req.Name, Description: req.Description, Color: req.Color, Permissions: req.Permissions}
	if err := database.ValidatePanelRole(role); err != nil {
		panelRoleError(w, err)
		return nil, false
	}
	return role, true
}

func (h *AuthHandler) CreatePanelRole(w http.ResponseWriter, r *http.Request) {
	if !h.requirePanelRoleAdmin(w, r) {
		return
	}
	role, ok := readPanelRole(w, r)
	if !ok {
		return
	}
	if err := h.db.CreatePanelRole(role); err != nil {
		panelRoleError(w, err)
		return
	}
	logAudit(h.db, r, "admin_panel_role_create", "Created panel role "+role.ID)
	writeJSON(w, http.StatusCreated, role)
}

func (h *AuthHandler) UpdatePanelRole(w http.ResponseWriter, r *http.Request) {
	if !h.requirePanelRoleAdmin(w, r) {
		return
	}
	role, ok := readPanelRole(w, r)
	if !ok {
		return
	}
	role.ID = getURLParam(r, "roleId")
	if err := h.db.UpdatePanelRole(role); err != nil {
		panelRoleError(w, err)
		return
	}
	logAudit(h.db, r, "admin_panel_role_update", "Updated panel role "+role.ID)
	writeJSON(w, http.StatusOK, role)
}

func (h *AuthHandler) DeletePanelRole(w http.ResponseWriter, r *http.Request) {
	if !h.requirePanelRoleAdmin(w, r) {
		return
	}
	id := getURLParam(r, "roleId")
	if err := h.db.DeletePanelRole(id); err != nil {
		panelRoleError(w, err)
		return
	}
	logAudit(h.db, r, "admin_panel_role_delete", "Deleted panel role "+id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) AssignPanelRole(w http.ResponseWriter, r *http.Request) {
	if !h.requirePanelRoleAdmin(w, r) {
		return
	}
	id, err := strconv.Atoi(getURLParam(r, "id"))
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
		return
	}
	var req struct {
		RoleID *string `json:"role_id"`
	}
	if !decodePanelRoleBody(w, r, &req) {
		return
	}
	if req.RoleID == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "role_id must be a string; use an empty string to remove the assignment"})
		return
	}
	if err := h.db.AssignPanelRole(id, *req.RoleID); err != nil {
		panelRoleError(w, err)
		return
	}
	user, err := h.db.GetUserByID(id)
	if err != nil {
		panelRoleError(w, err)
		return
	}
	logAudit(h.db, r, "admin_panel_role_assign", "Updated panel role for user "+strconv.Itoa(id))
	writeJSON(w, http.StatusOK, user)
}
