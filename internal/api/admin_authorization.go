package api

import (
	"net/http"

	"github.com/Pekly/NeoPanel/internal/database"
)

// CanCreateServer must be called before a master proxies a create request.
// It uses the current built-in database role, not a stale session or forwarded grant.
func (h *AuthHandler) CanCreateServer(r *http.Request) (bool, error) {
	return canCreateServer(h.db, r)
}

func canCreateServer(db *database.DB, r *http.Request) (bool, error) {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil || claims.Role != "admin" {
		return false, nil
	}
	// Configured native node tokens are authenticated machine identities, not users.
	if claims.UserID == 0 && claims.Role == "admin" {
		return true, nil
	}
	if claims.UserID <= 0 {
		return false, nil
	}
	user, err := db.GetUserByID(claims.UserID)
	if err != nil {
		return false, err
	}
	return user.Role == "admin", nil
}

func (h *AuthHandler) requireCurrentAdmin(w http.ResponseWriter, r *http.Request) bool {
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
