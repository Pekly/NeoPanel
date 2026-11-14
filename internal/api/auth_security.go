package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

const authBodyLimit int64 = 16 << 10
const passkeyBodyLimit int64 = 256 << 10

// Read the entire bounded document before decoding, including chunked bodies and
// trailing data. This prevents an early valid JSON value hiding an oversized body.
func decodeAuthJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, authBodyLimit)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, dst)
}

func (h *AuthHandler) validSessionVersion(claims *JWTClaims) bool {
	if h.db == nil || claims.UserID <= 0 {
		return false
	}
	user, err := h.db.GetUserByID(claims.UserID)
	return err == nil && user != nil && user.AuthVersion == claims.AuthVersion
}

// Restrict only auth JSON routes, never uploads. Passkey attestation needs a
// larger allowance than passwords, TOTP codes, and invitation credentials.
func limitAuthBody(w http.ResponseWriter, r *http.Request) bool {
	p := r.URL.Path
	limit := int64(0)
	if strings.HasPrefix(p, "/api/auth/") {
		limit = authBodyLimit
	}
	if strings.HasPrefix(p, "/api/auth/passkeys/") || strings.HasPrefix(p, "/api/account/passkeys/") {
		limit = passkeyBodyLimit
	}
	if limit == 0 {
		return true
	}
	if r.ContentLength > limit {
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "Request body too large"})
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	return true
}
