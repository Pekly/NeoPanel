package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/Pekly/NeoPanel/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db        *database.DB
	jwtSecret []byte
	expHours  int
}

type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthHandler(db *database.DB, jwtSecret string, expHours int) *AuthHandler {
	return &AuthHandler{
		db:        db,
		jwtSecret: []byte(jwtSecret),
		expHours:  expHours,
	}
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

func (h *AuthHandler) ValidateToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return h.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

func (h *AuthHandler) Status(w http.ResponseWriter, r *http.Request) {
	count, err := h.db.GetUserCount()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to query database"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"needs_setup": count == 0,
		"version":     "1.0.0",
	})
}

func (h *AuthHandler) Setup(w http.ResponseWriter, r *http.Request) {
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
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid input"})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if len(req.Username) < 3 || len(req.Password) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Username >= 3 chars, Password >= 6 chars required"})
		return
	}

	hash, err := h.HashPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
		return
	}

	user, err := h.db.CreateUser(req.Username, hash, "admin")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create admin user"})
		return
	}

	token, _ := h.GenerateToken(user)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Admin user created successfully",
		"token":   token,
		"user":    user,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	user, err := h.db.GetUserByUsername(req.Username)
	if err != nil || !h.CheckPasswordHash(req.Password, user.PasswordHash) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid username or password"})
		return
	}

	token, err := h.GenerateToken(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
		return
	}

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

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(contextUserKey).(*JWTClaims)
	writeJSON(w, http.StatusOK, user)
}

