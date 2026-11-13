package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db        *database.DB
	cfg       *config.Config
	mu        sync.RWMutex
	jwtSecret []byte
	expHours  int
}

type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Scope    string `json:"scope,omitempty"`
	// Set only by trusted node-forwarding middleware, never parsed from JWTs.
	AuthorizedServerCreate bool `json:"-"`
	jwt.RegisteredClaims
}

func NewAuthHandler(db *database.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		db:        db,
		cfg:       cfg,
		jwtSecret: []byte(cfg.Security.JWTSecret),
		expHours:  cfg.Security.TokenExpiration,
	}
}

func (h *AuthHandler) isTurnstileActive() (bool, string, string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.cfg == nil {
		return false, "", ""
	}
	t := h.cfg.Security.Turnstile
	active := t.Enabled && strings.TrimSpace(t.SiteKey) != "" && strings.TrimSpace(t.SecretKey) != ""
	return active, strings.TrimSpace(t.SiteKey), strings.TrimSpace(t.SecretKey)
}

func (h *AuthHandler) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func (h *AuthHandler) CheckPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (h *AuthHandler) GenerateToken(user *database.User) (string, error) {
	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(h.expHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}

func (h *AuthHandler) Generate2FATempToken(user *database.User) (string, error) {
	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		Scope:    "2fa_pending",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}

func (h *AuthHandler) Validate2FATempToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return h.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		if claims.Scope != "2fa_pending" {
			return nil, jwt.ErrTokenInvalidClaims
		}
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

func (h *AuthHandler) ValidateToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return h.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		if claims.Scope == "2fa_pending" {
			return nil, jwt.ErrTokenInvalidClaims
		}
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

// ValidateAuthToken validates a configured node token, a "neo_..." API key, or a signed JWT session token.
func (h *AuthHandler) ValidateAuthToken(tokenStr string) (*JWTClaims, error) {
	tokenStr = strings.TrimSpace(tokenStr)
	if tokenStr == "" {
		return nil, jwt.ErrTokenMalformed
	}
	if nodeTok := h.getNodeToken(); nodeTok != "" && len(tokenStr) == len(nodeTok) &&
		subtle.ConstantTimeCompare([]byte(tokenStr), []byte(nodeTok)) == 1 {
		return &JWTClaims{
			UserID:   0,
			Username: "node-daemon",
			Role:     "admin",
		}, nil
	}
	if strings.HasPrefix(tokenStr, "neo_") {
		sum := sha256.Sum256([]byte(tokenStr))
		keyHash := hex.EncodeToString(sum[:])
		user, err := h.db.GetUserByAPIKeyHash(keyHash)
		if err != nil || user == nil {
			return nil, jwt.ErrSignatureInvalid
		}
		return &JWTClaims{
			UserID:   user.ID,
			Username: user.Username,
			Role:     user.Role,
		}, nil
	}
	claims, err := h.ValidateToken(tokenStr)
	if err != nil {
		return nil, err
	}
	user, err := h.db.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		return nil, jwt.ErrSignatureInvalid
	}
	claims.UserID, claims.Username, claims.Role = user.ID, user.Username, user.Role
	return claims, nil
}

func (h *AuthHandler) getNodeToken() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.cfg == nil {
		return ""
	}
	return strings.TrimSpace(h.cfg.Security.NodeToken)
}

func (h *AuthHandler) Status(w http.ResponseWriter, r *http.Request) {
	if h.cfg != nil && h.cfg.IsNode() {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"needs_setup":        false,
			"version":            "1.0.0",
			"role":               "node",
			"node_mode":          true,
			"turnstile_enabled":  false,
			"turnstile_site_key": "",
		})
		return
	}

	count, err := h.db.GetUserCount()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to query database"})
		return
	}

	tsActive, tsSiteKey, _ := h.isTurnstileActive()
	if !tsActive {
		tsSiteKey = ""
	}
	branding := LoadBrandingConfig(h.db)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"needs_setup":         count == 0,
		"version":             "1.0.0",
		"role":                "panel",
		"node_mode":           false,
		"turnstile_enabled":   tsActive,
		"turnstile_site_key":  tsSiteKey,
		"maintenance_enabled": branding.MaintenanceEnabled,
		"maintenance_title":   branding.MaintenanceTitle,
		"maintenance_message": branding.MaintenanceMessage,
		"maintenance_eta":     branding.MaintenanceETA,
	})
}

func (h *AuthHandler) Setup(w http.ResponseWriter, r *http.Request) {
	if h.cfg != nil && h.cfg.IsNode() {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Setup is managed from the Master Panel on dedicated node instances; authenticate using an API key or node_token"})
		return
	}
	count, err := h.db.GetUserCount()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}
	if count > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Setup has already been completed"})
		return
	}

	var req struct {
		Username       string `json:"username"`
		Email          string `json:"email"`
		Password       string `json:"password"`
		TurnstileResp  string `json:"cf-turnstile-response"`
		TurnstileToken string `json:"turnstile_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid input"})
		return
	}

	if tsActive, _, tsSecret := h.isTurnstileActive(); tsActive {
		token := req.TurnstileResp
		if token == "" {
			token = req.TurnstileToken
		}
		if err := VerifyTurnstileToken(tsSecret, token, r); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	if len(req.Username) < 3 || len(req.Password) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Username >= 3 chars, Password >= 6 chars required"})
		return
	}
	if req.Email == "" || !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "A valid email address is required"})
		return
	}

	hash, err := h.HashPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
		return
	}

	user, err := h.db.CreateUserWithDetails(req.Username, req.Email, hash, "admin", "")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create admin user"})
		return
	}

	logAuditUser(h.db, user.ID, r, "auth_setup", fmt.Sprintf("Initialized NeoPanel admin account '%s' (%s)", user.Username, user.Email))

	token, _ := h.GenerateToken(user)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Admin user created successfully",
		"token":   token,
		"user":    user,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if h.cfg != nil && h.cfg.IsNode() {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Direct login is managed from the Master Panel on dedicated node instances; authenticate using an API key or node_token"})
		return
	}
	var req struct {
		Username       string `json:"username"`
		Password       string `json:"password"`
		TurnstileResp  string `json:"cf-turnstile-response"`
		TurnstileToken string `json:"turnstile_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if tsActive, _, tsSecret := h.isTurnstileActive(); tsActive {
		token := req.TurnstileResp
		if token == "" {
			token = req.TurnstileToken
		}
		if err := VerifyTurnstileToken(tsSecret, token, r); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	user, err := h.db.GetUserByEmailOrUsername(req.Username)
	if err != nil || !h.CheckPasswordHash(req.Password, user.PasswordHash) {
		logAuditUser(h.db, 0, r, "auth_login_failed", fmt.Sprintf("Failed sign-in attempt for '%s'", truncateAuditText(strings.TrimSpace(req.Username), 64)))
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid username or password"})
		return
	}

	h.completePrimaryLogin(w, r, user)
}

// completePrimaryLogin applies the account's second-factor policy after password verification.
func (h *AuthHandler) completePrimaryLogin(w http.ResponseWriter, r *http.Request, user *database.User) {
	if user.TwoFactorEnabled {
		tempToken, err := h.Generate2FATempToken(user)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to initiate 2FA verification"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"two_factor_required": true,
			"temp_token":          tempToken,
			"username":            user.Username,
		})
		return
	}
	h.completeVerifiedLogin(w, r, user)
}

// completeVerifiedLogin issues a session only after the caller has verified all
// required factors. A user-verified WebAuthn assertion is sufficient on its own.
func (h *AuthHandler) completeVerifiedLogin(w http.ResponseWriter, r *http.Request, user *database.User) {

	token, err := h.GenerateToken(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
		return
	}

	logAuditUser(h.db, user.ID, r, "auth_login", fmt.Sprintf("Signed in to NeoPanel (%s)", user.Role))

	// Set HTTP-only cookie as well for easy SPA access
	http.SetCookie(w, &http.Cookie{
		Name:     "neopanel_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(time.Duration(h.expHours) * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

func (h *AuthHandler) Login2FA(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TempToken string `json:"temp_token"`
		Code      string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid payload"})
		return
	}

	claims, err := h.Validate2FATempToken(req.TempToken)
	if err != nil || claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "2FA session expired or invalid. Please log in again."})
		return
	}

	user, err := h.db.GetUserByID(claims.UserID)
	if err != nil || user == nil || !user.TwoFactorEnabled {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "User not found or 2FA not enabled"})
		return
	}

	cleanCode := strings.TrimSpace(req.Code)
	isValid := ValidateTOTPCode(user.TwoFactorSecret, cleanCode)
	usedRecovery := false

	if !isValid {
		var remaining string
		isValid, remaining = CheckAndConsumeRecoveryCode(user.RecoveryCodes, cleanCode)
		if isValid {
			usedRecovery = true
			_ = h.db.UpdateUserTwoFactor(user.ID, true, user.TwoFactorSecret, remaining)
		}
	}

	if !isValid {
		logAuditUser(h.db, user.ID, r, "auth_2fa_failed", fmt.Sprintf("Invalid 2FA code entered for '%s'", user.Username))
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid 2FA authentication code"})
		return
	}

	token, err := h.GenerateToken(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to generate session token"})
		return
	}

	logMsg := "2FA verified successfully"
	if usedRecovery {
		logMsg = "2FA verified via emergency recovery code"
	}
	logAuditUser(h.db, user.ID, r, "auth_2fa_success", logMsg)

	http.SetCookie(w, &http.Cookie{
		Name:     "neopanel_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(time.Duration(h.expHours) * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

// --- Admin User Management Endpoints ---

func (h *AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if !h.requirePanelRoleAdmin(w, r) {
		return
	}
	users, err := h.db.ListUsers()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *AuthHandler) AdminCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		Role      string `json:"role"`
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
	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Username and password are required"})
		return
	}
	if req.Email == "" || !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "A valid email address is required for user accounts"})
		return
	}
	if req.Role != "admin" {
		req.Role = "user"
	}

	hash, err := h.HashPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
		return
	}

	user, err := h.db.CreateUserWithFullDetails(req.Username, req.FirstName, req.LastName, req.Email, hash, req.Role, req.AvatarURL)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Username or email already taken or invalid"})
		return
	}

	logAudit(h.db, r, "admin_user_create", fmt.Sprintf("Created panel user '%s' (%s, role: %s)", user.Username, user.Email, user.Role))

	writeJSON(w, http.StatusCreated, user)
}

func (h *AuthHandler) AdminUpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := getURLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Role      string `json:"role"`
		Password  string `json:"password"`
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
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Username is required"})
		return
	}
	if req.Email == "" || !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "A valid email address is required"})
		return
	}
	if req.Role != "admin" {
		req.Role = "user"
	}

	// Check if username taken by another user
	if existing, _ := h.db.GetUserByUsername(req.Username); existing != nil && existing.ID != id {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Username already taken by another account"})
		return
	}
	// Check if email taken by another user
	if existing, _ := h.db.GetUserByEmailOrUsername(req.Email); existing != nil && existing.ID != id {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Email already in use by another account"})
		return
	}

	var passwordHash string
	if strings.TrimSpace(req.Password) != "" {
		if len(req.Password) < 6 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Password must be at least 6 characters"})
			return
		}
		hash, err := h.HashPassword(req.Password)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
			return
		}
		passwordHash = hash
	}

	if err := h.db.AdminUpdateUser(id, req.Username, req.FirstName, req.LastName, req.Email, req.Role, passwordHash); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update user: " + err.Error()})
		return
	}

	updatedUser, _ := h.db.GetUserByID(id)
	logAudit(h.db, r, "admin_user_update", fmt.Sprintf("Updated panel user '%s' (ID %d, role: %s)", req.Username, id, req.Role))

	writeJSON(w, http.StatusOK, updatedUser)
}

func (h *AuthHandler) AdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := getURLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
		return
	}

	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims != nil && claims.UserID == id {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Cannot delete your own account"})
		return
	}

	targetUser, _ := h.db.GetUserByID(id)
	if err := h.db.DeleteUser(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete user"})
		return
	}

	targetDesc := fmt.Sprintf("ID %d", id)
	if targetUser != nil && targetUser.Username != "" {
		targetDesc = fmt.Sprintf("'%s' (ID %d)", targetUser.Username, id)
	}
	logAudit(h.db, r, "admin_user_delete", fmt.Sprintf("Deleted panel user %s", targetDesc))

	writeJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(contextUserKey).(*JWTClaims)
	if !ok || claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}
	var dbUser *database.User
	var err error
	if claims.UserID > 0 {
		dbUser, err = h.db.GetUserByID(claims.UserID)
	}
	if err == nil && dbUser != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"id":                 dbUser.ID,
			"user_id":            dbUser.ID,
			"username":           dbUser.Username,
			"first_name":         dbUser.FirstName,
			"last_name":          dbUser.LastName,
			"email":              dbUser.Email,
			"role":               dbUser.Role,
			"panel_role_id":      dbUser.PanelRoleID,
			"panel_role_name":    dbUser.PanelRoleName,
			"panel_permissions":  dbUser.PanelPermissions,
			"avatar_url":         dbUser.AvatarURL,
			"two_factor_enabled": dbUser.TwoFactorEnabled,
		})
		return
	}
	if claims.UserID == 0 && claims.Role == "admin" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"id": 0, "user_id": 0, "username": claims.Username, "role": claims.Role,
			"panel_role_id": "", "panel_role_name": "", "panel_permissions": []string{database.PermissionServersCreate},
		})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unable to resolve current user"})
}

func (h *AuthHandler) GetTurnstileConfig(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	t := h.cfg.Security.Turnstile
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"enabled":        t.Enabled,
		"site_key":       t.SiteKey,
		"has_secret_key": strings.TrimSpace(t.SecretKey) != "",
	})
}

func (h *AuthHandler) UpdateTurnstileConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled   bool   `json:"enabled"`
		SiteKey   string `json:"site_key"`
		SecretKey string `json:"secret_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}

	h.mu.Lock()
	h.cfg.Security.Turnstile.Enabled = req.Enabled
	h.cfg.Security.Turnstile.SiteKey = strings.TrimSpace(req.SiteKey)
	if strings.TrimSpace(req.SecretKey) != "" {
		h.cfg.Security.Turnstile.SecretKey = strings.TrimSpace(req.SecretKey)
	}
	if h.cfg.Security.Turnstile.Enabled && (h.cfg.Security.Turnstile.SiteKey == "" || h.cfg.Security.Turnstile.SecretKey == "") {
		h.mu.Unlock()
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Both Site Key and Secret Key are required when enabling Turnstile"})
		return
	}
	cfgPath := h.cfg.FilePath
	var saveErr error
	if cfgPath != "" {
		saveErr = config.SaveConfig(cfgPath, h.cfg)
	}
	enabled := h.cfg.Security.Turnstile.Enabled
	siteKey := h.cfg.Security.Turnstile.SiteKey
	hasSecret := strings.TrimSpace(h.cfg.Security.Turnstile.SecretKey) != ""
	h.mu.Unlock()

	if saveErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Updated in memory, but failed to write config.yaml: " + saveErr.Error()})
		return
	}

	tsState := "disabled"
	if enabled {
		tsState = "enabled"
	}
	logAudit(h.db, r, "admin_turnstile", fmt.Sprintf("Cloudflare Turnstile protection %s", tsState))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":        "Cloudflare Turnstile settings saved",
		"enabled":        enabled,
		"site_key":       siteKey,
		"has_secret_key": hasSecret,
	})
}
