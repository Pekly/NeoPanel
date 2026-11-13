package api

import "github.com/Pekly/NeoPanel/internal/database"

func permissionsContainAll(held, requested []string) bool {
	allowed := make(map[string]bool, len(held))
	for _, permission := range held {
		allowed[permission] = true
	}
	for _, permission := range requested {
		if !allowed[permission] {
			return false
		}
	}
	return true
}

// Team managers can delegate only their own permissions. Owners and built-in
// administrators retain full control; global panel roles do not confer team access.
func canGrantServerPermissions(db *database.DB, claims *JWTClaims, serverID string, requested []string) bool {
	if claims == nil {
		return false
	}
	server, err := db.GetServer(serverID)
	if err != nil || server == nil {
		return false
	}
	if claims.Role == "admin" || server.OwnerID == claims.UserID {
		return true
	}
	held, err := db.GetUserServerPermissions(serverID, claims.UserID)
	return err == nil && permissionsContainAll(held, append([]string{"subusers.manage"}, requested...))
}
