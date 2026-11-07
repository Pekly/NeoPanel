package api

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
)

type contextKey string

const contextUserKey contextKey = "user"

type Router struct {
	mux            *http.ServeMux
	cfg            *config.Config
	auth           *AuthHandler
	servers        *ServerHandler
	console        *ConsoleHandler
	files          *FileHandler
	embeddedStatic fs.FS
}

func NewRouter(cfg *config.Config, db *database.DB, manager *runner.ServerManager, embeddedStatic fs.FS) *Router {
	authHandler := NewAuthHandler(db, cfg.Security.JWTSecret, cfg.Security.TokenExpiration)
	serverHandler := NewServerHandler(db, manager)
	consoleHandler := NewConsoleHandler(authHandler, manager)
	fileHandler := NewFileHandler(manager)

	r := &Router{
		mux:            http.NewServeMux(),
		cfg:            cfg,
		auth:           authHandler,
		servers:        serverHandler,
		console:        consoleHandler,
		files:          fileHandler,
		embeddedStatic: embeddedStatic,
	}

	r.registerRoutes()
	return r
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Security & CORS headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	r.mux.ServeHTTP(w, req)
}

func (r *Router) registerRoutes() {
	// Public auth routes
	r.mux.HandleFunc("GET /api/auth/status", r.auth.Status)
	r.mux.HandleFunc("POST /api/auth/setup", r.auth.Setup)
	r.mux.HandleFunc("POST /api/auth/login", r.auth.Login)

	// Protected auth routes
	r.mux.HandleFunc("GET /api/auth/me", r.requireAuth(r.auth.Me))

	// Protected server routes
	r.mux.HandleFunc("GET /api/servers", r.requireAuth(r.servers.ListServers))
	r.mux.HandleFunc("POST /api/servers", r.requireAuth(r.servers.CreateServer))
	r.mux.HandleFunc("GET /api/servers/{id}", r.requireAuth(r.servers.GetServer))
	r.mux.HandleFunc("PUT /api/servers/{id}", r.requireAuth(r.servers.UpdateServer))
	r.mux.HandleFunc("DELETE /api/servers/{id}", r.requireAuth(r.servers.DeleteServer))
	r.mux.HandleFunc("POST /api/servers/{id}/power", r.requireAuth(r.servers.PowerAction))
	r.mux.HandleFunc("POST /api/servers/{id}/command", r.requireAuth(r.servers.SendCommand))

	// WebSocket console route (handles token auth inside)
	r.mux.HandleFunc("GET /api/servers/{id}/ws", r.console.HandleWebSocket)

	// File manager routes
	r.mux.HandleFunc("GET /api/servers/{id}/files", r.requireAuth(r.files.ListFiles))
	r.mux.HandleFunc("GET /api/servers/{id}/files/content", r.requireAuth(r.files.GetContent))
	r.mux.HandleFunc("POST /api/servers/{id}/files/content", r.requireAuth(r.files.SaveContent))
	r.mux.HandleFunc("POST /api/servers/{id}/files/create", r.requireAuth(r.files.Create))
	r.mux.HandleFunc("DELETE /api/servers/{id}/files", r.requireAuth(r.files.Delete))
	r.mux.HandleFunc("POST /api/servers/{id}/files/upload", r.requireAuth(r.files.Upload))

	// System info
	r.mux.HandleFunc("GET /api/system", r.requireAuth(HandleSystemInfo))

	// Embedded Static & SPA routes
	if r.embeddedStatic != nil {
		fileServer := http.FileServer(http.FS(r.embeddedStatic))
		r.mux.HandleFunc("GET /", func(w http.ResponseWriter, req *http.Request) {
			path := req.URL.Path
			// If request is not an API call and doesn't have an extension, serve index.html (SPA routing)
			if !strings.HasPrefix(path, "/api/") {
				if path == "/" || !strings.Contains(filepath.Base(path), ".") {
					req.URL.Path = "/"
				}
			}
			fileServer.ServeHTTP(w, req)
		})
	}
}

func (r *Router) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		tokenStr := ""

		// 1. Check Authorization Bearer header
		authHeader := req.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 2. Check Cookie fallback
		if tokenStr == "" {
			if cookie, err := req.Cookie("neopanel_token"); err == nil {
				tokenStr = cookie.Value
			}
		}

		if tokenStr == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized: missing token"})
			return
		}

		claims, err := r.auth.ValidateToken(tokenStr)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized: invalid or expired token"})
			return
		}

		ctx := context.WithValue(req.Context(), contextUserKey, claims)
		next(w, req.WithContext(ctx))
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func getURLParam(r *http.Request, key string) string {
	return r.PathValue(key)
}
