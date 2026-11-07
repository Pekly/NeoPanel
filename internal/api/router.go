package api

import (
	"context"
	"encoding/json"
	"fmt"
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
	account        *AccountHandler
	servers        *ServerHandler
	console        *ConsoleHandler
	files          *FileHandler
	plugins        *PluginHandler
	embeddedStatic fs.FS
}

func NewRouter(cfg *config.Config, db *database.DB, manager *runner.ServerManager, embeddedStatic fs.FS) *Router {
	authHandler := NewAuthHandler(db, cfg.Security.JWTSecret, cfg.Security.TokenExpiration)
	accountHandler := NewAccountHandler(db, authHandler)
	serverHandler := NewServerHandler(db, manager)
	consoleHandler := NewConsoleHandler(authHandler, manager)
	fileHandler := NewFileHandler(manager)
	pluginHandler := NewPluginHandler(db, manager)

	r := &Router{
		mux:            http.NewServeMux(),
		cfg:            cfg,
		auth:           authHandler,
		account:        accountHandler,
		servers:        serverHandler,
		console:        consoleHandler,
		files:          fileHandler,
		plugins:        pluginHandler,
		embeddedStatic: embeddedStatic,
	}

	r.registerRoutes(db)
	return r
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Security & CORS headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	r.mux.ServeHTTP(w, req)
}

func (r *Router) registerRoutes(db *database.DB) {
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

	// Modrinth Plugins & Addons routes
	r.mux.HandleFunc("GET /api/plugins/search", r.requireAuth(r.plugins.Search))
	r.mux.HandleFunc("GET /api/servers/{id}/plugins", r.requireAuth(r.plugins.ListInstalled))
	r.mux.HandleFunc("POST /api/servers/{id}/plugins/install", r.requireAuth(r.plugins.Install))
	r.mux.HandleFunc("DELETE /api/servers/{id}/plugins", r.requireAuth(r.plugins.Delete))

	// SFTP details endpoint
	r.mux.HandleFunc("GET /api/servers/{id}/sftp", r.requireAuth(func(w http.ResponseWriter, req *http.Request) {
		serverID := getURLParam(req, "id")
		srv, err := db.GetServer(serverID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
			return
		}
		userClaims, _ := req.Context().Value(contextUserKey).(*JWTClaims)
		username := "admin"
		if userClaims != nil && userClaims.Username != "" {
			username = userClaims.Username
		}
		host := req.Host
		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"enabled":     r.cfg.SFTP.Enabled,
			"port":        r.cfg.SFTP.Port,
			"host":        host,
			"username":    fmt.Sprintf("%s.%s", username, srv.ID),
			"server_id":   srv.ID,
			"server_name": srv.Name,
		})
	}))

	// Account Settings routes
	r.mux.HandleFunc("GET /api/account/profile", r.requireAuth(r.account.GetProfile))
	r.mux.HandleFunc("POST /api/account/password", r.requireAuth(r.account.UpdatePassword))
	r.mux.HandleFunc("GET /api/account/ssh-keys", r.requireAuth(r.account.ListSSHKeys))
	r.mux.HandleFunc("POST /api/account/ssh-keys", r.requireAuth(r.account.AddSSHKey))
	r.mux.HandleFunc("DELETE /api/account/ssh-keys/{id}", r.requireAuth(r.account.DeleteSSHKey))
	r.mux.HandleFunc("GET /api/account/api-keys", r.requireAuth(r.account.ListAPIKeys))
	r.mux.HandleFunc("POST /api/account/api-keys", r.requireAuth(r.account.CreateAPIKey))
	r.mux.HandleFunc("DELETE /api/account/api-keys/{id}", r.requireAuth(r.account.DeleteAPIKey))
	r.mux.HandleFunc("GET /api/account/activity", r.requireAuth(r.account.ListActivity))

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
