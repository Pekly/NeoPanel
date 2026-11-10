package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Pekly/NeoPanel/internal/database"
)

func logAudit(db *database.DB, r *http.Request, action, details string) {
	if db == nil {
		return
	}
	userID := 0
	if r != nil {
		if claims, ok := r.Context().Value(contextUserKey).(*JWTClaims); ok && claims != nil {
			userID = claims.UserID
		}
	}
	logAuditUser(db, userID, r, action, details)
}

func logAuditUser(db *database.DB, userID int, r *http.Request, action, details string) {
	if db == nil {
		return
	}
	ip := ""
	if r != nil {
		ip = extractClientIP(r)
	}
	_ = db.LogActivity(userID, action, details, ip)
}

func auditServerLabel(db *database.DB, serverID string) string {
	if db != nil && serverID != "" {
		if srv, err := db.GetServer(serverID); err == nil && srv != nil && strings.TrimSpace(srv.Name) != "" {
			return fmt.Sprintf("%s (%s)", srv.Name, serverID)
		}
	}
	return serverID
}

func truncateAuditText(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func formatBytesAudit(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	units := []string{"KiB", "MiB", "GiB", "TiB"}
	val := float64(bytes) / 1024.0
	for _, u := range units {
		if val < 1024.0 || u == "TiB" {
			return fmt.Sprintf("%.1f %s", val, u)
		}
		val /= 1024.0
	}
	return fmt.Sprintf("%d B", bytes)
}
