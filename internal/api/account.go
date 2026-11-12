package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

	var user *database.User
	var err error
	if claims.UserID > 0 {
		user, err = h.db.GetUserByID(claims.UserID)
	}
	if user == nil || err != nil {
		user, err = h.db.GetUserByUsername(claims.Username)
	}
	if err != nil || user == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	sshKeys, _ := h.db.ListSSHKeys(user.ID)
	apiKeys, _ := h.db.ListAPIKeys(user.ID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":                 user.ID,
		"username":           user.Username,
		"first_name":         user.FirstName,
		"last_name":          user.LastName,
		"email":              user.Email,
		"role":               user.Role,
		"avatar_url":         user.AvatarURL,
		"two_factor_enabled": user.TwoFactorEnabled,
		"created_at":         user.CreatedAt,
		"ssh_key_count":      len(sshKeys),
		"api_key_count":      len(apiKeys),
	})
}

func (h *AccountHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextUserKey).(*JWTClaims)
	if !ok || claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}

	var user *database.User
	var err error
	if claims.UserID > 0 {
		user, err = h.db.GetUserByID(claims.UserID)
	}
	if user == nil || err != nil {
		user, err = h.db.GetUserByUsername(claims.Username)
	}
	if err != nil || user == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	var req struct {
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Email = strings.TrimSpace(req.Email)

	if req.Username == "" {
		req.Username = user.Username
	}
	if len(req.Username) < 2 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Account name must be at least 2 characters long"})
		return
	}

	if req.Email == "" || !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "A valid email address is required"})
		return
	}

	// Check if username taken by another user
	if existing, _ := h.db.GetUserByUsername(req.Username); existing != nil && existing.ID != user.ID {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Account name already taken by another user"})
		return
	}

	// Check if email taken by another user
	if existing, _ := h.db.GetUserByEmailOrUsername(req.Email); existing != nil && existing.ID != user.ID {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Email address already in use by another account"})
		return
	}

	if err := h.db.UpdateUserProfile(user.ID, req.Username, req.FirstName, req.LastName, req.Email); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update profile: " + err.Error()})
		return
	}

	user.Username = req.Username
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Email = req.Email

	if req.AvatarURL != "" {
		_ = h.db.UpdateUserAvatar(user.ID, req.AvatarURL)
		user.AvatarURL = req.AvatarURL
	}

	newToken, _ := h.auth.GenerateToken(user)
	if newToken != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     "jwt",
			Value:    newToken,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   r.TLS != nil,
			MaxAge:   86400 * 7,
		})
	}

	_ = h.db.LogActivity(user.ID, "profile_update", fmt.Sprintf("Updated account profile: %s (%s)", user.Username, user.Email), extractClientIP(r))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Profile updated successfully",
		"token":   newToken,
		"user":    user,
	})
}

func (h *AccountHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
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

	avatarURL, err := saveUploadedAvatarFile(r, user.ID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	_ = h.db.UpdateUserAvatar(user.ID, avatarURL)
	_ = h.db.LogActivity(user.ID, "avatar_upload", "Uploaded new profile avatar", extractClientIP(r))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "Avatar uploaded successfully",
		"avatar_url": avatarURL,
	})
}

func (h *AccountHandler) AdminUploadUserAvatar(w http.ResponseWriter, r *http.Request) {
	idStr := getURLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
		return
	}

	user, err := h.db.GetUserByID(id)
	if err != nil || user == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	avatarURL, err := saveUploadedAvatarFile(r, user.ID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	_ = h.db.UpdateUserAvatar(user.ID, avatarURL)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "User avatar updated",
		"avatar_url": avatarURL,
	})
}

func saveUploadedAvatarFile(r *http.Request, userID int) (string, error) {
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		return "", fmt.Errorf("avatar image must be under 5 MB")
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		return "", fmt.Errorf("avatar file field is required")
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" || (ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" && ext != ".gif" && ext != ".svg") {
		ext = ".png"
	}

	dir := filepath.Join("data", "avatars")
	_ = os.MkdirAll(dir, 0755)

	filename := fmt.Sprintf("avatar_%d%s", userID, ext)
	destPath := filepath.Join(dir, filename)

	dst, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to save avatar image")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to write avatar image")
	}

	return fmt.Sprintf("/api/avatars/%s?t=%d", filename, time.Now().Unix()), nil
}

func (h *AccountHandler) ServeAvatar(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/api/avatars/")
	if len(parts) < 2 || parts[1] == "" {
		http.NotFound(w, r)
		return
	}
	filename := filepath.Base(parts[1])
	filePath := filepath.Join("data", "avatars", filename)

	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeFile(w, r, filePath)
}

func (h *AccountHandler) Generate2FA(w http.ResponseWriter, r *http.Request) {
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

	secret, err := GenerateTOTPSecret()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to generate secret key"})
		return
	}

	otpURL := GenerateTOTPURL("NeoPanel", user.Username, secret)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"secret":      secret,
		"otpauth_url": otpURL,
		"username":    user.Username,
	})
}

func (h *AccountHandler) Enable2FA(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		Secret string `json:"secret"`
		Code   string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
		return
	}

	if !ValidateTOTPCode(req.Secret, req.Code) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid 6-digit verification code. Please ensure your authenticator app time is synchronized."})
		return
	}

	recoveryCodes := GenerateRecoveryCodes(8)
	recoveryCSV := strings.Join(recoveryCodes, ",")

	if err := h.db.UpdateUserTwoFactor(user.ID, true, req.Secret, recoveryCSV); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to enable 2FA: " + err.Error()})
		return
	}

	_ = h.db.LogActivity(user.ID, "2fa_enable", "Activated two-factor authentication", extractClientIP(r))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":        "Two-factor authentication successfully activated",
		"recovery_codes": recoveryCodes,
	})
}

func (h *AccountHandler) Disable2FA(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		Password string `json:"password"`
		Code     string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
		return
	}

	pwdValid := h.auth.CheckPasswordHash(req.Password, user.PasswordHash)
	codeValid := user.TwoFactorSecret != "" && ValidateTOTPCode(user.TwoFactorSecret, req.Code)

	if !pwdValid && !codeValid {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Current password or valid 2FA code is required to disable 2FA"})
		return
	}

	if err := h.db.UpdateUserTwoFactor(user.ID, false, "", ""); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to disable 2FA"})
		return
	}

	_ = h.db.LogActivity(user.ID, "2fa_disable", "Deactivated two-factor authentication", extractClientIP(r))

	writeJSON(w, http.StatusOK, map[string]string{"message": "Two-factor authentication has been disabled"})
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
