package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Pekly/NeoPanel/internal/database"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
)

type AccountHandler struct {
	db   *database.DB
	auth *AuthHandler
}

func NewAccountHandler(db *database.DB, auth *AuthHandler) *AccountHandler {
	return &AccountHandler{
		db:   db,
		auth: auth,
	}
}

func (h *AccountHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextUserKey).(*JWTClaims)
	if !ok || claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	user, err := h.db.GetUserByUsername(claims.Username)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	sshKeys, _ := h.db.ListSSHKeys(user.ID)
	apiKeys, _ := h.db.ListAPIKeys(user.ID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":            user.ID,
		"username":      user.Username,
		"role":          user.Role,
		"created_at":    user.CreatedAt,
		"ssh_key_count": len(sshKeys),
		"api_key_count": len(apiKeys),
	})
}

func (h *AccountHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextUserKey).(*JWTClaims)
	if !ok || claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
		return
	}

	if len(req.NewPassword) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "New password must be at least 6 characters"})
		return
	}

	user, err := h.db.GetUserByUsername(claims.Username)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Current password incorrect"})
		return
	}

	newHash, err := h.auth.HashPassword(req.NewPassword)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
		return
	}

	if err := h.db.UpdateUserPassword(user.ID, newHash); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update password"})
		return
	}

	_ = h.db.LogActivity(user.ID, "password_change", "Updated master account password", extractClientIP(r))

	writeJSON(w, http.StatusOK, map[string]string{"message": "Password updated successfully"})
}

// --- SSH Keys ---

func (h *AccountHandler) ListSSHKeys(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextUserKey).(*JWTClaims)
	if !ok || claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	user, err := h.db.GetUserByUsername(claims.Username)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	keys, err := h.db.ListSSHKeys(user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to list SSH keys"})
		return
	}
	if keys == nil {
		keys = []*database.SSHKey{}
	}

	writeJSON(w, http.StatusOK, keys)
}

func (h *AccountHandler) AddSSHKey(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextUserKey).(*JWTClaims)
	if !ok || claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	var req struct {
		Name      string `json:"name"`
		PublicKey string `json:"public_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.PublicKey = strings.TrimSpace(req.PublicKey)

	if req.Name == "" || req.PublicKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Key name and public key content are required"})
		return
	}

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(req.PublicKey))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid OpenSSH public key format"})
		return
	}

	fingerprint := ssh.FingerprintSHA256(pubKey)

	user, err := h.db.GetUserByUsername(claims.Username)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	key, err := h.db.AddSSHKey(user.ID, req.Name, fingerprint, req.PublicKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save SSH key: " + err.Error()})
		return
	}

	_ = h.db.LogActivity(user.ID, "ssh_key_add", fmt.Sprintf("Added SSH key '%s' (%s)", req.Name, fingerprint), extractClientIP(r))

	writeJSON(w, http.StatusOK, key)
}

func (h *AccountHandler) DeleteSSHKey(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextUserKey).(*JWTClaims)
	if !ok || claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	keyIDStr := getURLParam(r, "id")
	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid key ID"})
		return
	}

	user, err := h.db.GetUserByUsername(claims.Username)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	if err := h.db.DeleteSSHKey(user.ID, keyID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete SSH key"})
		return
	}

	_ = h.db.LogActivity(user.ID, "ssh_key_delete", fmt.Sprintf("Deleted SSH key ID %d", keyID), extractClientIP(r))

	writeJSON(w, http.StatusOK, map[string]string{"message": "SSH key removed"})
}

// --- API Keys ---

func (h *AccountHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextUserKey).(*JWTClaims)
	if !ok || claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	user, err := h.db.GetUserByUsername(claims.Username)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	keys, err := h.db.ListAPIKeys(user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to list API keys"})
		return
	}
	if keys == nil {
		keys = []*database.APIKey{}
	}

	writeJSON(w, http.StatusOK, keys)
}

func (h *AccountHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextUserKey).(*JWTClaims)
	if !ok || claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Key name is required"})
		return
	}

	// Generate 32 bytes random key
	tokenBytes := make([]byte, 32)
	_, _ = rand.Read(tokenBytes)
	rawToken := "neo_" + hex.EncodeToString(tokenBytes)

	// Hash key for storage
	hash := sha256.Sum256([]byte(rawToken))
	keyHash := hex.EncodeToString(hash[:])
	keyPrefix := rawToken[:12] + "..."

	user, err := h.db.GetUserByUsername(claims.Username)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	apiKey, err := h.db.AddAPIKey(user.ID, strings.TrimSpace(req.Name), keyHash, keyPrefix)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create API key"})
		return
	}

	_ = h.db.LogActivity(user.ID, "api_key_create", fmt.Sprintf("Created API key '%s' (%s)", req.Name, keyPrefix), extractClientIP(r))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":         apiKey.ID,
		"name":       apiKey.Name,
		"key_prefix": apiKey.KeyPrefix,
		"created_at": apiKey.CreatedAt,
		"token":      rawToken, // Provided once
	})
}

func (h *AccountHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextUserKey).(*JWTClaims)
	if !ok || claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	keyIDStr := getURLParam(r, "id")
	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid key ID"})
		return
	}

	user, err := h.db.GetUserByUsername(claims.Username)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	if err := h.db.DeleteAPIKey(user.ID, keyID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete API key"})
		return
	}

	_ = h.db.LogActivity(user.ID, "api_key_delete", fmt.Sprintf("Deleted API key ID %d", keyID), extractClientIP(r))

	writeJSON(w, http.StatusOK, map[string]string{"message": "API key deleted"})
}

// --- Activity Logs ---

func (h *AccountHandler) ListActivity(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextUserKey).(*JWTClaims)
	if !ok || claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	user, err := h.db.GetUserByUsername(claims.Username)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	limit := 200
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	includeAll := strings.EqualFold(user.Role, "admin") && r.URL.Query().Get("scope") != "mine"
	logs, err := h.db.ListActivity(user.ID, includeAll, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to load activity logs"})
		return
	}
	if logs == nil {
		logs = []*database.ActivityLog{}
	}

	writeJSON(w, http.StatusOK, logs)
}

