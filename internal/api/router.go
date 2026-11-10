package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

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
	backups        *BackupHandler
	players        *PlayerHandler
	software       *SoftwareHandler
	branding       *BrandingHandler
	mods           *ModsHandler
	databases      *DatabaseHandler
	nodes          *NodeHandler
	subdomains     *SubdomainHandler
	subusers       *SubuserHandler
	schedules      *ScheduleHandler
	clientAPI      *ClientAPIHandler
	globalLimiter  *IPRateLimiter
	authLimiter    *IPRateLimiter
	embeddedStatic fs.FS
}

func NewRouter(cfg *config.Config, db *database.DB, manager *runner.ServerManager, embeddedStatic fs.FS) *Router {
	authHandler := NewAuthHandler(db, cfg)
	accountHandler := NewAccountHandler(db, authHandler)
	serverHandler := NewServerHandler(cfg, db, manager)
	consoleHandler := NewConsoleHandler(db, authHandler, manager)
	fileHandler := NewFileHandler(db, manager)
	pluginHandler := NewPluginHandler(db, manager)
	backupHandler := NewBackupHandler(cfg, db, manager)
	playerHandler := NewPlayerHandler(db, manager)
	softwareHandler := NewSoftwareHandler(db, manager)
	brandingHandler := NewBrandingHandler(db, cfg)
	modsHandler := NewModsHandler(db, cfg)
	databaseHandler := NewDatabaseHandler(cfg, db)
	nodeHandler := NewNodeHandler(cfg, db, authHandler, manager)
	subdomainHandler := NewSubdomainHandler(db, nodeHandler)
	subuserHandler := NewSubuserHandler(db)
	scheduleHandler := NewScheduleHandler(db, manager)
	clientAPIHandler := NewClientAPIHandler(cfg, db, manager)

	r := &Router{
		mux:            http.NewServeMux(),
		cfg:            cfg,
		auth:           authHandler,
		account:        accountHandler,
		servers:        serverHandler,
		console:        consoleHandler,
		files:          fileHandler,
		plugins:        pluginHandler,
		backups:        backupHandler,
		players:        playerHandler,
		software:       softwareHandler,
		branding:       brandingHandler,
		mods:           modsHandler,
		databases:      databaseHandler,
		nodes:          nodeHandler,
		subdomains:     subdomainHandler,
		subusers:       subuserHandler,
		schedules:      scheduleHandler,
		clientAPI:      clientAPIHandler,
		globalLimiter:  NewIPRateLimiter(60, 100),
		authLimiter:    NewIPRateLimiter(1, 15),
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

	if r.globalLimiter != nil {
		r.globalLimiter.Middleware(r.mux).ServeHTTP(w, req)
		return
	}
	r.mux.ServeHTTP(w, req)
}

func (r *Router) registerRoutes(db *database.DB) {
	// Public auth, branding & mods runtime routes
	r.mux.HandleFunc("GET /api/auth/status", r.auth.Status)
	r.mux.HandleFunc("POST /api/auth/setup", r.auth.Setup)
	if r.authLimiter != nil {
		r.mux.Handle("POST /api/auth/login", r.authLimiter.Middleware(http.HandlerFunc(r.auth.Login)))
	} else {
		r.mux.HandleFunc("POST /api/auth/login", r.auth.Login)
	}
	r.mux.HandleFunc("GET /api/branding", r.branding.GetPublicBranding)
	r.mux.HandleFunc("GET /branding/{filename}", r.branding.ServeBrandingAsset)
	r.mux.HandleFunc("GET /api/mods", r.mods.ListRuntimeMods)
	r.mux.HandleFunc("GET /mods/{id}/{path...}", r.mods.ServeModAsset)

	// Automated Client Area Provisioning API (WHMCS, Blesta, Billing portals)
	r.mux.HandleFunc("POST /api/client-api/v1/provision", r.clientAPI.Provision)
	r.mux.HandleFunc("GET /api/client-api/v1/code-samples", r.requireAuth(r.clientAPI.GetCodeSamples))

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

	// Server Icon routes
	r.mux.HandleFunc("POST /api/servers/{id}/icon", r.requireAuth(r.servers.UploadServerIcon))
	r.mux.HandleFunc("GET /api/servers/{id}/icon", r.servers.GetServerIcon)
	r.mux.HandleFunc("DELETE /api/servers/{id}/icon", r.requireAuth(r.servers.DeleteServerIcon))

	// Server Quick Configs & Optimization Presets
	r.mux.HandleFunc("GET /api/servers/{id}/configs", r.requireAuth(r.servers.GetQuickConfigs))
	r.mux.HandleFunc("POST /api/servers/{id}/configs", r.requireAuth(r.servers.UpdateQuickConfigs))

	// Subusers & Permission Management
	r.mux.HandleFunc("GET /api/servers/{id}/subusers", r.requireAuth(r.subusers.ListSubusers))
	r.mux.HandleFunc("POST /api/servers/{id}/subusers", r.requireAuth(r.subusers.AddSubuser))
	r.mux.HandleFunc("PUT /api/servers/{id}/subusers/{sub_id}", r.requireAuth(r.subusers.UpdateSubuser))
	r.mux.HandleFunc("DELETE /api/servers/{id}/subusers/{sub_id}", r.requireAuth(r.subusers.DeleteSubuser))
	r.mux.HandleFunc("GET /api/servers/{id}/permissions", r.requireAuth(r.subusers.GetMyPermissions))

	// Schedules & Automation
	r.mux.HandleFunc("GET /api/servers/{id}/schedules", r.requireAuth(r.schedules.ListSchedules))
	r.mux.HandleFunc("POST /api/servers/{id}/schedules", r.requireAuth(r.schedules.CreateSchedule))
	r.mux.HandleFunc("PUT /api/servers/{id}/schedules/{schedId}", r.requireAuth(r.schedules.UpdateSchedule))
	r.mux.HandleFunc("DELETE /api/servers/{id}/schedules/{schedId}", r.requireAuth(r.schedules.DeleteSchedule))
	r.mux.HandleFunc("POST /api/servers/{id}/schedules/{schedId}/toggle", r.requireAuth(r.schedules.ToggleSchedule))
	r.mux.HandleFunc("POST /api/servers/{id}/schedules/{schedId}/run", r.requireAuth(r.schedules.RunScheduleNow))

	// Port allocations routes
	r.mux.HandleFunc("GET /api/servers/{id}/ports", r.requireAuth(r.servers.ListServerPorts))
	r.mux.HandleFunc("POST /api/servers/{id}/ports", r.requireAuth(r.servers.AddServerPort))
	r.mux.HandleFunc("DELETE /api/servers/{id}/ports/{portId}", r.requireAuth(r.servers.DeleteServerPort))
	r.mux.HandleFunc("POST /api/servers/{id}/ports/{portId}/primary", r.requireAuth(r.servers.SetPrimaryPort))

	// WebSocket console route (handles token auth inside)
	r.mux.HandleFunc("GET /api/servers/{id}/ws", r.console.HandleWebSocket)

	// File manager routes
	r.mux.HandleFunc("GET /api/servers/{id}/files", r.requireAuth(r.files.ListFiles))
	r.mux.HandleFunc("GET /api/servers/{id}/files/content", r.requireAuth(r.files.GetContent))
	r.mux.HandleFunc("POST /api/servers/{id}/files/content", r.requireAuth(r.files.SaveContent))
	r.mux.HandleFunc("POST /api/servers/{id}/files/create", r.requireAuth(r.files.Create))
	r.mux.HandleFunc("DELETE /api/servers/{id}/files", r.requireAuth(r.files.Delete))
	r.mux.HandleFunc("POST /api/servers/{id}/files/upload", r.requireAuth(r.files.Upload))
	r.mux.HandleFunc("GET /api/servers/{id}/files/download", r.requireAuth(r.files.Download))
	r.mux.HandleFunc("POST /api/servers/{id}/files/rename", r.requireAuth(r.files.Rename))
	r.mux.HandleFunc("POST /api/servers/{id}/files/move", r.requireAuth(r.files.Move))
	r.mux.HandleFunc("POST /api/servers/{id}/files/compress", r.requireAuth(r.files.Compress))
	r.mux.HandleFunc("POST /api/servers/{id}/files/archive", r.requireAuth(r.files.Compress))
	r.mux.HandleFunc("POST /api/servers/{id}/files/unzip", r.requireAuth(r.files.Unzip))
	r.mux.HandleFunc("POST /api/servers/{id}/files/extract", r.requireAuth(r.files.Unzip))

	// Modrinth Plugins & Addons routes
	r.mux.HandleFunc("GET /api/plugins/search", r.requireAuth(r.plugins.Search))
	r.mux.HandleFunc("GET /api/servers/{id}/plugins", r.requireAuth(r.plugins.ListInstalled))
	r.mux.HandleFunc("POST /api/servers/{id}/plugins/install", r.requireAuth(r.plugins.Install))
	r.mux.HandleFunc("DELETE /api/servers/{id}/plugins", r.requireAuth(r.plugins.Delete))

	// Backups routes
	r.mux.HandleFunc("GET /api/servers/{id}/backups", r.requireAuth(r.backups.List))
	r.mux.HandleFunc("POST /api/servers/{id}/backups", r.requireAuth(r.backups.Create))
	r.mux.HandleFunc("DELETE /api/servers/{id}/backups", r.requireAuth(r.backups.Delete))
	r.mux.HandleFunc("GET /api/servers/{id}/backups/download", r.requireAuth(r.backups.Download))

	// Players routes
	r.mux.HandleFunc("GET /api/servers/{id}/players", r.requireAuth(r.players.GetPlayers))
	r.mux.HandleFunc("POST /api/servers/{id}/players/action", r.requireAuth(r.players.Action))

	// Software, Minecraft Versions & Properties routes
	r.mux.HandleFunc("GET /api/versions", r.requireAuth(HandleGetMinecraftVersions))
	r.mux.HandleFunc("POST /api/servers/{id}/software", r.requireAuth(r.software.ChangeSoftware))
	r.mux.HandleFunc("POST /api/servers/{id}/reinstall", r.requireAuth(r.software.ReinstallServer))
	r.mux.HandleFunc("GET /api/servers/{id}/properties", r.requireAuth(r.software.GetProperties))
	r.mux.HandleFunc("POST /api/servers/{id}/properties", r.requireAuth(r.software.SaveProperties))

	// Per-server SQLite Databases routes
	r.mux.HandleFunc("GET /api/servers/{id}/databases", r.requireAuth(r.databases.List))
	r.mux.HandleFunc("POST /api/servers/{id}/databases", r.requireAuth(r.databases.Create))
	r.mux.HandleFunc("DELETE /api/servers/{id}/databases/{dbId}", r.requireAuth(r.databases.Delete))
	r.mux.HandleFunc("POST /api/servers/{id}/databases/{dbId}/rotate", r.requireAuth(r.databases.RotatePassword))
	r.mux.HandleFunc("GET /api/servers/{id}/databases/{dbId}/tables", r.requireAuth(r.databases.Tables))
	r.mux.HandleFunc("POST /api/servers/{id}/databases/{dbId}/query", r.requireAuth(r.databases.Query))
	r.mux.HandleFunc("GET /api/servers/{id}/databases/{dbId}/export", r.requireAuth(r.databases.Export))

	// Subdomain & DNS/SRV management routes
	r.mux.HandleFunc("GET /api/servers/{id}/subdomain", r.requireAuth(r.subdomains.GetSubdomain))
	r.mux.HandleFunc("PUT /api/servers/{id}/subdomain", r.requireAuth(r.subdomains.UpdateSubdomain))
	r.mux.HandleFunc("POST /api/servers/{id}/subdomain/verify", r.requireAuth(r.subdomains.VerifySubdomainDNS))

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
		host := r.nodes.resolveLocalFQDN(req.Host)
		portStr := fmt.Sprintf("%d", r.cfg.SFTP.Port)
		sftpUser := fmt.Sprintf("%s.%s", username, srv.ID)
		listening := false
		if r.cfg.SFTP.Enabled && r.cfg.SFTP.Port > 0 {
			if conn, dialErr := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", portStr), 250*time.Millisecond); dialErr == nil {
				listening = true
				_ = conn.Close()
			}
		}
		launchURI := (&url.URL{
			Scheme: "sftp",
			User:   url.User(sftpUser),
			Host:   net.JoinHostPort(host, portStr),
			Path:   "/",
		}).String()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"enabled":       r.cfg.SFTP.Enabled,
			"listening":     listening,
			"port":          r.cfg.SFTP.Port,
			"host":          host,
			"username":      sftpUser,
			"launch_uri":    launchURI,
			"password_hint": "Use your NeoPanel account password (or add an SSH Public Key in Account Settings).",
			"server_id":     srv.ID,
			"server_name":   srv.Name,
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

	// System info & Admin tools
	r.mux.HandleFunc("GET /api/system", r.requireAuth(HandleSystemInfo))
	r.mux.HandleFunc("GET /api/system/java", r.requireAuth(HandleGetJavaRuntimes))
	r.mux.HandleFunc("POST /api/system/java/install", r.requireAuth(HandleInstallJavaRuntime(db)))
	r.mux.HandleFunc("POST /api/admin/gc", r.requireAuth(func(w http.ResponseWriter, req *http.Request) {
		logAudit(db, req, "admin_gc", "Triggered manual Go runtime garbage collection")
		HandleRunGC(w, req)
	}))
	r.mux.HandleFunc("GET /api/admin/turnstile", r.requireAuth(r.auth.GetTurnstileConfig))
	r.mux.HandleFunc("POST /api/admin/turnstile", r.requireAuth(r.auth.UpdateTurnstileConfig))
	r.mux.HandleFunc("GET /api/admin/branding", r.requireAuth(r.branding.GetAdminBranding))
	r.mux.HandleFunc("PUT /api/admin/branding", r.requireAuth(r.branding.UpdateAdminBranding))
	r.mux.HandleFunc("POST /api/admin/branding/upload", r.requireAuth(r.branding.UploadBrandingAsset))
	r.mux.HandleFunc("POST /api/admin/branding/reset", r.requireAuth(r.branding.ResetAdminBranding))
	r.mux.HandleFunc("GET /api/admin/mods", r.requireAuth(r.mods.ListAdminMods))
	r.mux.HandleFunc("POST /api/admin/mods", r.requireAuth(r.mods.SaveAdminMod))
	r.mux.HandleFunc("POST /api/admin/mods/upload", r.requireAuth(r.mods.UploadAdminMod))
	r.mux.HandleFunc("PUT /api/admin/mods/{id}", r.requireAuth(r.mods.SaveAdminMod))
	r.mux.HandleFunc("DELETE /api/admin/mods/{id}", r.requireAuth(r.mods.DeleteAdminMod))
	r.mux.HandleFunc("GET /api/admin/users", r.requireAuth(r.auth.ListUsers))
	r.mux.HandleFunc("POST /api/admin/users", r.requireAuth(r.auth.AdminCreateUser))
	r.mux.HandleFunc("DELETE /api/admin/users/{id}", r.requireAuth(r.auth.AdminDeleteUser))

	// Multi-node cluster & proxy routes
	r.mux.HandleFunc("GET /api/cluster/servers", r.requireAuth(r.nodes.ListClusterServers))
	r.mux.HandleFunc("GET /api/nodes", r.requireAuth(r.nodes.ListNodes))
	r.mux.HandleFunc("POST /api/nodes", r.requireAuth(r.nodes.CreateNode))
	r.mux.HandleFunc("POST /api/nodes/test", r.requireAuth(r.nodes.TestNodePayload))
	r.mux.HandleFunc("POST /api/nodes/{id}/test", r.requireAuth(r.nodes.TestNodeByID))
	r.mux.HandleFunc("PUT /api/nodes/{id}", r.requireAuth(r.nodes.UpdateNode))
	r.mux.HandleFunc("DELETE /api/nodes/{id}", r.requireAuth(r.nodes.DeleteNode))
	for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"} {
		r.mux.HandleFunc(method+" /api/nodes/{id}/proxy/{rest...}", r.requireAuth(r.nodes.Proxy))
	}
	r.mux.HandleFunc("GET /api/nodes/{id}/ws/{serverId}", r.nodes.ProxyWS)

	// Embedded Static & Multi-Page Panel routes (active in panel mode; omitted in node daemons)
	if r.embeddedStatic != nil && !(r.cfg != nil && r.cfg.IsNode()) {
		fileServer := http.FileServer(http.FS(r.embeddedStatic))
		assetExts := map[string]bool{
			".css": true, ".js": true, ".map": true, ".svg": true,
			".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
			".webp": true, ".ico": true, ".woff": true, ".woff2": true,
			".ttf": true, ".json": true, ".txt": true, ".xml": true,
		}
		r.mux.HandleFunc("GET /", func(w http.ResponseWriter, req *http.Request) {
			reqPath := req.URL.Path
			if strings.HasPrefix(reqPath, "/api/") {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "API endpoint not found"})
				return
			}
			clean := strings.TrimPrefix(path.Clean("/"+reqPath), "/")
			if clean == "" || clean == "." {
				req.URL.Path = "/"
				fileServer.ServeHTTP(w, req)
				return
			}
			if info, err := fs.Stat(r.embeddedStatic, clean); err == nil && !info.IsDir() {
				fileServer.ServeHTTP(w, req)
				return
			}
			ext := strings.ToLower(path.Ext(clean))
			if !assetExts[ext] {
				req.URL.Path = "/"
			}
			fileServer.ServeHTTP(w, req)
		})
	} else {
		r.mux.HandleFunc("GET /", func(w http.ResponseWriter, req *http.Request) {
			if strings.HasPrefix(req.URL.Path, "/api/") {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "API endpoint not found"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"service": "neopanel",
				"role":    "node",
				"status":  "online",
				"version": "1.0.0",
			})
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

		// 3. Check ?token= query fallback (used for browser downloads across proxied nodes)
		if tokenStr == "" {
			if qt := strings.TrimSpace(req.URL.Query().Get("token")); qt != "" {
				tokenStr = qt
			}
		}

		if tokenStr == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized: missing token"})
			return
		}

		claims, err := r.auth.ValidateAuthToken(tokenStr)
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
