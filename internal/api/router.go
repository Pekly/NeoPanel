package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
	"github.com/gorilla/websocket"
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
	email          *EmailHandler
	invitations    *InvitationHandler
	globalLimiter  *IPRateLimiter
	authLimiter    *IPRateLimiter
	clientIPs      *clientIPPolicy
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
	emailHandler := NewEmailHandler(db, cfg, authHandler)

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
		email:          emailHandler,
		globalLimiter:  NewIPRateLimiter(60, 100),
		authLimiter:    NewIPRateLimiter(1, 15),
		clientIPs:      newClientIPPolicy(cfg.Security.TrustedProxies),
		embeddedStatic: embeddedStatic,
	}
	r.invitations = NewInvitationHandler(r)

	r.registerRoutes(db)
	return r
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Security & CORS headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	if !limitAuthBody(w, req) {
		return
	}
	req = r.clientIPs.request(req)

	if r.globalLimiter != nil {
		r.globalLimiter.Middleware(r.mux).ServeHTTP(w, req)
		return
	}
	r.mux.ServeHTTP(w, req)
}

func (r *Router) registerRoutes(db *database.DB) {
	community := NewCommunityHandler(db)
	r.mux.HandleFunc("GET /api/community", community.GetPublic)
	r.mux.HandleFunc("GET /api/admin/community", r.requireAdmin(community.GetAdmin))
	r.mux.HandleFunc("PUT /api/admin/community", r.requireAdmin(community.UpdateAdmin))
	r.mux.HandleFunc("GET /api/admin/roles", r.requireAdmin(r.auth.ListPanelRoles))
	r.mux.HandleFunc("POST /api/admin/roles", r.requireAdmin(r.auth.CreatePanelRole))
	r.mux.HandleFunc("PUT /api/admin/roles/{roleId}", r.requireAdmin(r.auth.UpdatePanelRole))
	r.mux.HandleFunc("DELETE /api/admin/roles/{roleId}", r.requireAdmin(r.auth.DeletePanelRole))
	r.mux.HandleFunc("PUT /api/admin/users/{id}/panel-role", r.requireAdmin(r.auth.AssignPanelRole))
	// Public auth, branding & mods runtime routes
	passkeys := NewPasskeyHandler(r.auth)
	for route, handler := range map[string]http.HandlerFunc{
		"POST /api/auth/passkeys/login/begin":  passkeys.LoginBegin,
		"POST /api/auth/passkeys/login/finish": passkeys.LoginFinish,
	} {
		if r.authLimiter != nil {
			r.mux.Handle(route, r.authLimiter.Middleware(handler))
		} else {
			r.mux.HandleFunc(route, handler)
		}
	}
	r.mux.HandleFunc("GET /api/account/passkeys", r.requireAuth(passkeys.List))
	r.mux.HandleFunc("POST /api/account/passkeys/register/begin", r.requireAuth(passkeys.RegisterBegin))
	r.mux.HandleFunc("POST /api/account/passkeys/register/finish", r.requireAuth(passkeys.RegisterFinish))
	r.mux.HandleFunc("DELETE /api/account/passkeys/{passkeyId}", r.requireAuth(passkeys.Delete))
	r.mux.HandleFunc("GET /api/auth/status", r.auth.Status)
	r.mux.HandleFunc("POST /api/auth/setup", r.auth.Setup)
	if r.authLimiter != nil {
		r.mux.Handle("POST /api/auth/login", r.authLimiter.Middleware(http.HandlerFunc(r.auth.Login)))
		r.mux.Handle("POST /api/auth/login/2fa", r.authLimiter.Middleware(http.HandlerFunc(r.auth.Login2FA)))
		r.mux.Handle("POST /api/auth/password-reset/request", r.authLimiter.Middleware(http.HandlerFunc(r.email.RequestPasswordReset)))
		r.mux.Handle("POST /api/auth/password-reset/confirm", r.authLimiter.Middleware(http.HandlerFunc(r.email.ConfirmPasswordReset)))
		r.mux.Handle("POST /api/auth/invitations/inspect", r.authLimiter.Middleware(http.HandlerFunc(r.invitations.Inspect)))
		r.mux.Handle("POST /api/auth/invitations/accept", r.authLimiter.Middleware(http.HandlerFunc(r.invitations.Accept)))
	} else {
		r.mux.HandleFunc("POST /api/auth/login", r.auth.Login)
		r.mux.HandleFunc("POST /api/auth/login/2fa", r.auth.Login2FA)
		r.mux.HandleFunc("POST /api/auth/password-reset/request", r.email.RequestPasswordReset)
		r.mux.HandleFunc("POST /api/auth/password-reset/confirm", r.email.ConfirmPasswordReset)
		r.mux.HandleFunc("POST /api/auth/invitations/inspect", r.invitations.Inspect)
		r.mux.HandleFunc("POST /api/auth/invitations/accept", r.invitations.Accept)
	}
	r.mux.HandleFunc("GET /api/branding", r.branding.GetPublicBranding)
	r.mux.HandleFunc("GET /branding/{filename}", r.branding.ServeBrandingAsset)
	r.mux.HandleFunc("GET /api/avatars/{filename}", r.account.ServeAvatar)
	r.mux.HandleFunc("GET /api/mods", r.mods.ListRuntimeMods)
	r.mux.HandleFunc("GET /mods/{id}/{path...}", r.mods.ServeModAsset)

	// Automated Client Area Provisioning API (WHMCS, Blesta, Billing portals)
	r.mux.HandleFunc("POST /api/client-api/v1/provision", r.clientAPI.Provision)
	r.mux.HandleFunc("GET /api/client-api/v1/code-samples", r.requireAdmin(r.clientAPI.GetCodeSamples))

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
	r.mux.HandleFunc("GET /api/servers/{id}/icon", r.requireAuth(r.servers.GetServerIcon))
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
	// Invitations are master-owned, including invitations to remote servers.
	r.mux.HandleFunc("GET /api/team/invitations", r.requireAuth(r.invitations.List))
	r.mux.HandleFunc("POST /api/team/invitations", r.requireAuth(r.invitations.Create))
	r.mux.HandleFunc("POST /api/team/invitations/{inviteId}/resend", r.requireAuth(r.invitations.Resend))
	r.mux.HandleFunc("DELETE /api/team/invitations/{inviteId}", r.requireAuth(r.invitations.Revoke))

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
	r.mux.HandleFunc("GET /api/servers/{id}/files/preview", r.requireAuth(r.files.Preview))
	r.mux.HandleFunc("GET /api/servers/{id}/files/media", r.requireAuth(r.files.Media))
	r.mux.HandleFunc("GET /api/servers/{id}/files/zip", r.requireAuth(r.files.ZipPreview))
	r.mux.HandleFunc("GET /api/servers/{id}/files/zip-entry", r.requireAuth(r.files.ZipEntryPreview))
	r.mux.HandleFunc("POST /api/servers/{id}/files/content", r.requireAuth(r.files.SaveContent))
	r.mux.HandleFunc("POST /api/servers/{id}/files/create", r.requireAuth(r.files.Create))
	r.mux.HandleFunc("DELETE /api/servers/{id}/files", r.requireAuth(r.files.Delete))
	r.mux.HandleFunc("POST /api/servers/{id}/files/delete-bulk", r.requireAuth(r.files.DeleteBulk))
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
	r.mux.HandleFunc("POST /api/servers/{id}/plugins/toggle", r.requireAuth(r.plugins.Toggle))
	r.mux.HandleFunc("GET /api/servers/{id}/plugins/versions", r.requireAuth(r.plugins.Versions))

	// Backups routes
	r.mux.HandleFunc("GET /api/servers/{id}/backups", r.requireAuth(r.backups.List))
	r.mux.HandleFunc("POST /api/servers/{id}/backups", r.requireAuth(r.backups.Create))
	r.mux.HandleFunc("DELETE /api/servers/{id}/backups", r.requireAuth(r.backups.Delete))
	r.mux.HandleFunc("GET /api/servers/{id}/backups/download", r.requireAuth(r.backups.Download))
	r.mux.HandleFunc("POST /api/servers/{id}/backups/restore", r.requireAuth(r.backups.Restore))

	// Players routes
	r.mux.HandleFunc("GET /api/servers/{id}/players", r.requireAuth(r.players.GetPlayers))
	r.mux.HandleFunc("POST /api/servers/{id}/players/action", r.requireAuth(r.players.Action))

	// Software, Minecraft Versions & Properties routes
	r.mux.HandleFunc("GET /api/versions", r.requireAuth(HandleGetMinecraftVersions))
	r.mux.HandleFunc("GET /api/steel/versions", r.requireAuth(HandleGetSteelVersions))
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
		if srv.NodeID != "" && srv.NodeID != "local" {
			if node, _ := db.GetNode(srv.NodeID); node != nil {
				if strings.TrimSpace(node.FQDN) != "" {
					host = strings.TrimSpace(node.FQDN)
				} else if node.BaseURL != "" {
					if u, err := url.Parse(node.BaseURL); err == nil && u.Hostname() != "" {
						host = u.Hostname()
					}
				}
			}
		}
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
	r.mux.HandleFunc("PUT /api/account/profile", r.requireAuth(r.account.UpdateProfile))
	r.mux.HandleFunc("POST /api/account/avatar", r.requireAuth(r.account.UploadAvatar))
	r.mux.HandleFunc("POST /api/account/2fa/generate", r.requireAuth(r.account.Generate2FA))
	r.mux.HandleFunc("POST /api/account/2fa/enable", r.requireAuth(r.account.Enable2FA))
	r.mux.HandleFunc("POST /api/account/2fa/disable", r.requireAuth(r.account.Disable2FA))
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
	r.mux.HandleFunc("POST /api/system/java/install", r.requireAdmin(HandleInstallJavaRuntime(db)))
	r.mux.HandleFunc("POST /api/admin/gc", r.requireAdmin(func(w http.ResponseWriter, req *http.Request) {
		logAudit(db, req, "admin_gc", "Triggered manual Go runtime garbage collection")
		HandleRunGC(w, req)
	}))
	r.mux.HandleFunc("GET /api/admin/turnstile", r.requireAdmin(r.auth.GetTurnstileConfig))
	r.mux.HandleFunc("POST /api/admin/turnstile", r.requireAdmin(r.auth.UpdateTurnstileConfig))
	r.mux.HandleFunc("GET /api/admin/panel-url", r.requireAdmin(r.auth.GetPanelURL))
	r.mux.HandleFunc("PUT /api/admin/panel-url", r.requireAdmin(r.auth.UpdatePanelURL))
	r.mux.HandleFunc("GET /api/admin/branding", r.requireAdmin(r.branding.GetAdminBranding))
	r.mux.HandleFunc("PUT /api/admin/branding", r.requireAdmin(r.branding.UpdateAdminBranding))
	r.mux.HandleFunc("POST /api/admin/branding", r.requireAdmin(r.branding.UpdateAdminBranding))
	r.mux.HandleFunc("PUT /api/branding", r.requireAdmin(r.branding.UpdateAdminBranding))
	r.mux.HandleFunc("POST /api/branding", r.requireAdmin(r.branding.UpdateAdminBranding))
	r.mux.HandleFunc("POST /api/admin/branding/upload", r.requireAdmin(r.branding.UploadBrandingAsset))
	r.mux.HandleFunc("POST /api/admin/branding/reset", r.requireAdmin(r.branding.ResetAdminBranding))
	r.mux.HandleFunc("GET /api/admin/mods", r.requireAdmin(r.mods.ListAdminMods))
	r.mux.HandleFunc("POST /api/admin/mods", r.requireAdmin(r.mods.SaveAdminMod))
	r.mux.HandleFunc("POST /api/admin/mods/upload", r.requireAdmin(r.mods.UploadAdminMod))
	r.mux.HandleFunc("PUT /api/admin/mods/{id}", r.requireAdmin(r.mods.SaveAdminMod))
	r.mux.HandleFunc("DELETE /api/admin/mods/{id}", r.requireAdmin(r.mods.DeleteAdminMod))
	r.mux.HandleFunc("GET /api/admin/users", r.requireAdmin(r.auth.ListUsers))
	r.mux.HandleFunc("POST /api/admin/users", r.requireAdmin(r.auth.AdminCreateUser))
	r.mux.HandleFunc("PUT /api/admin/users/{id}", r.requireAdmin(r.auth.AdminUpdateUser))
	r.mux.HandleFunc("DELETE /api/admin/users/{id}", r.requireAdmin(r.auth.AdminDeleteUser))
	r.mux.HandleFunc("POST /api/admin/users/{id}/avatar", r.requireAdmin(r.account.AdminUploadUserAvatar))

	// Admin panel updates & GitHub repository management
	r.mux.HandleFunc("GET /api/admin/updates", r.requireAdmin(HandleGetUpdateStatus(db)))
	r.mux.HandleFunc("POST /api/admin/updates/config", r.requireAdmin(HandleSaveUpdateConfig(db)))
	r.mux.HandleFunc("POST /api/admin/updates/apply", r.requireAdmin(HandleApplyUpdate(db)))
	r.mux.HandleFunc("GET /api/admin/updates/logs", r.requireAdmin(HandleGetUpdateLogs))

	// Admin SMTP email configuration & test
	r.mux.HandleFunc("GET /api/admin/email", r.requireAdmin(r.email.HandleGetSettings))
	r.mux.HandleFunc("POST /api/admin/email", r.requireAdmin(r.email.HandleSaveSettings))
	r.mux.HandleFunc("POST /api/admin/email/test", r.requireAdmin(r.email.HandleTestEmail))

	// Multi-node cluster & proxy routes
	r.mux.HandleFunc("GET /api/cluster/servers", r.requireAuth(r.listClusterServers))
	r.mux.HandleFunc("GET /api/nodes", r.requireAuth(r.nodes.ListNodes))
	r.mux.HandleFunc("POST /api/nodes", r.requireAdmin(r.nodes.CreateNode))
	r.mux.HandleFunc("POST /api/nodes/test", r.requireAdmin(r.nodes.TestNodePayload))
	r.mux.HandleFunc("POST /api/nodes/{id}/test", r.requireAdmin(r.nodes.TestNodeByID))
	r.mux.HandleFunc("PUT /api/nodes/{id}", r.requireAdmin(r.nodes.UpdateNode))
	r.mux.HandleFunc("DELETE /api/nodes/{id}", r.requireAdmin(r.nodes.DeleteNode))
	for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"} {
		r.mux.HandleFunc(method+" /api/nodes/{id}/proxy/{rest...}", r.requireAuth(r.proxyHTTP))
	}
	r.mux.HandleFunc("GET /api/nodes/{id}/ws/{serverId}", r.requireAuth(r.proxyWebSocket))

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
			cleanWithoutStatic := strings.TrimPrefix(clean, "static/")
			if info, err := fs.Stat(r.embeddedStatic, clean); err == nil && !info.IsDir() {
				fileServer.ServeHTTP(w, req)
				return
			}
			if info, err := fs.Stat(r.embeddedStatic, cleanWithoutStatic); err == nil && !info.IsDir() {
				req.URL.Path = "/" + cleanWithoutStatic
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
		claims, err = forwardedClaims(r.auth.db, req, claims)
		if err != nil {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		w.Header().Set("X-NeoPanel-User-ID", strconv.Itoa(claims.UserID))
		w.Header().Set("X-NeoPanel-Permission-Policy", "2")
		if strings.HasPrefix(req.URL.Path, "/api/servers/") {
			serverID := getURLParam(req, "id")
			suffix := strings.TrimPrefix(strings.TrimPrefix(req.URL.Path, "/api/servers/"+serverID), "/")
			permission := serverActionPermission(req.Method, suffix)
			if status := authorizeServer(r.auth.db, claims, serverID, permission); status != 0 {
				writeError(w, status, "Access denied: server membership and "+permission+" permission required")
				return
			}
		}

		ctx := context.WithValue(req.Context(), contextUserKey, claims)
		next(w, req.WithContext(ctx))
	}
}

func (r *Router) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return r.requireAuth(func(w http.ResponseWriter, req *http.Request) {
		claims, _ := req.Context().Value(contextUserKey).(*JWTClaims)
		if claims == nil || claims.Role != "admin" {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "Forbidden: Administrator privileges required"})
			return
		}
		next(w, req)
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]interface{}{
		"status": "error",
		"error":  message,
	})
}

func getClientIP(r *http.Request) string {
	return extractClientIP(r)
}

func extractIDFromPath(path, prefix, suffix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	trimmed := strings.TrimPrefix(path, prefix)
	if suffix != "" {
		idx := strings.Index(trimmed, suffix)
		if idx != -1 {
			return trimmed[:idx]
		}
	}
	parts := strings.Split(trimmed, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func getUserFromContext(r *http.Request) *database.User {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil {
		return nil
	}
	return &database.User{
		ID:       claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
	}
}

func getURLParam(r *http.Request, key string) string {
	return r.PathValue(key)
}

// serverActionPermission is shared by all operational routes. Unknown operations
// fail closed to owners/admins rather than silently granting any authenticated user.
func serverActionPermission(method, suffix string) string {
	read := method == http.MethodGet || method == http.MethodHead
	parts := strings.Split(suffix, "/")
	switch parts[0] {
	case "":
		if read {
			return ""
		}
		if method == http.MethodPut {
			return "settings.edit"
		}
		return "owner"
	case "permissions", "icon":
		if read {
			return ""
		}
		return "settings.edit"
	case "ws":
		return "console.view"
	case "command":
		return "console.command"
	case "power":
		return "server.power"
	case "files":
		if method == http.MethodPost && suffix == "files/delete-bulk" {
			return "files.delete"
		}
		if read {
			return "files.view"
		}
		if method == http.MethodDelete {
			return "files.delete"
		}
		if len(parts) > 1 {
			switch parts[1] {
			case "rename", "move", "unzip", "extract":
				return "files.delete"
			}
		}
		return "files.edit"
	case "plugins":
		if read {
			return "files.view"
		}
		if method == http.MethodDelete {
			return "files.delete"
		}
		return "files.edit"
	case "backups":
		return "backup.manage"
	case "databases":
		return "database.manage"
	case "subusers":
		return "subusers.manage"
	case "schedules":
		return "schedules.manage"
	case "players":
		if read {
			return "console.view"
		}
		return "console.command"
	case "sftp":
		return "sftp.access"
	case "configs", "properties", "ports", "subdomain", "software", "reinstall":
		if read {
			return "settings.view"
		}
		return "settings.edit"
	default:
		return "owner"
	}
}

func authorizeServer(db *database.DB, claims *JWTClaims, serverID, permission string) int {
	if claims == nil {
		return http.StatusUnauthorized
	}
	srv, err := db.GetServer(serverID)
	if err != nil || srv == nil {
		return http.StatusNotFound
	}
	if claims.Role == "admin" || srv.OwnerID == claims.UserID {
		return 0
	}
	perms, err := db.GetUserServerPermissions(serverID, claims.UserID)
	if err != nil {
		return http.StatusForbidden
	}
	if permission == "" {
		return 0
	}
	for _, p := range perms {
		if p == permission && permission != "owner" {
			return 0
		}
	}
	return http.StatusForbidden
}

// Nodes already trust their configured node/admin API token and synchronize user
// IDs from the master. Carry the original caller over that authenticated hop, so
// the node evaluates its own server ownership/subusers instead of acting as admin.
// Ordinary tokens can never assert a forwarded identity.
func forwardedClaims(db *database.DB, req *http.Request, claims *JWTClaims) (*JWTClaims, error) {
	id := req.Header.Get("X-NeoPanel-User-ID")
	if id == "" {
		return claims, nil
	}
	if claims == nil || claims.Role != "admin" {
		return nil, fmt.Errorf("untrusted forwarded identity")
	}
	userID, err := strconv.Atoi(id)
	if err != nil || userID <= 0 {
		return nil, fmt.Errorf("invalid forwarded user")
	}
	user, err := db.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("forwarded user is not synchronized to this node")
	}
	role := req.Header.Get("X-NeoPanel-User-Role")
	if role != "admin" && role != "user" {
		return nil, fmt.Errorf("invalid forwarded role")
	}
	allowCreate := req.Method == http.MethodPost && req.URL.Path == "/api/servers" && req.Header.Get("X-NeoPanel-Can-Create-Server") == "true"
	return &JWTClaims{UserID: user.ID, Username: user.Username, Role: role, AuthorizedServerCreate: allowCreate}, nil
}

func setProxyIdentity(header http.Header, claims *JWTClaims, nodeToken string) {
	header.Del("X-NeoPanel-User-ID")
	header.Del("X-NeoPanel-User-Role")
	header.Del("X-NeoPanel-Can-Create-Server")
	header.Set("Authorization", "Bearer "+nodeToken)
	// Native node tokens have user ID zero and are already privileged machine identities.
	if claims.UserID > 0 && claims.Role != "admin" {
		header.Set("X-NeoPanel-User-ID", strconv.Itoa(claims.UserID))
		header.Set("X-NeoPanel-User-Role", claims.Role)
	}
}

func (r *Router) proxyHTTP(w http.ResponseWriter, req *http.Request) {
	node, err := r.auth.db.GetNode(getURLParam(req, "id"))
	if err != nil {
		writeError(w, 404, "Node not found")
		return
	}
	rest := strings.TrimPrefix(path.Clean("/"+getURLParam(req, "rest")), "/")
	if !strings.HasPrefix(rest, "api/") || rest == "api/nodes" || strings.HasPrefix(rest, "api/nodes/") || strings.HasPrefix(rest, "api/cluster/") {
		writeError(w, 403, "Invalid proxy path")
		return
	}
	claims, _ := req.Context().Value(contextUserKey).(*JWTClaims)
	// Only server operations and public catalogs are available through a user proxy.
	if claims.Role != "admin" && !strings.HasPrefix(rest, "api/servers/") && rest != "api/servers" && rest != "api/plugins/search" && rest != "api/versions" && rest != "api/steel/versions" && rest != "api/system/java" {
		writeError(w, 403, "Proxy operation requires administrator privileges")
		return
	}
	createAuthorized := false
	if req.Method == http.MethodPost && rest == "api/servers" {
		allowed, err := r.auth.CanCreateServer(req)
		if err != nil {
			writeError(w, 503, "Could not verify server creation permission")
			return
		}
		if !allowed {
			writeError(w, 403, "Your panel role does not allow server creation")
			return
		}
		createAuthorized = true
	}
	if err := r.checkProxyIdentity(req, node, claims); err != nil {
		writeError(w, 502, err.Error())
		return
	}
	q := req.URL.Query()
	q.Del("token")
	remote, err := http.NewRequestWithContext(req.Context(), req.Method, strings.TrimRight(node.BaseURL, "/")+"/"+rest+"?"+q.Encode(), req.Body)
	if err != nil {
		writeError(w, 400, "Invalid proxy request")
		return
	}
	remote.Header.Set("Content-Type", req.Header.Get("Content-Type"))
	remote.Header.Set("Accept", req.Header.Get("Accept"))
	for _, key := range []string{"Range", "If-Range", "If-Modified-Since", "If-Unmodified-Since"} {
		if value := req.Header.Get(key); value != "" {
			remote.Header.Set(key, value)
		}
	}
	remote.ContentLength = req.ContentLength
	setProxyIdentity(remote.Header, claims, node.APIToken)
	if createAuthorized && claims.UserID > 0 && claims.Role != "admin" {
		remote.Header.Set("X-NeoPanel-Can-Create-Server", "true")
	}
	resp, err := fileMediaProxyClient(r.nodes.httpClient, req.Method, rest).Do(remote)
	if err != nil {
		writeError(w, 502, "Failed to reach node: "+err.Error())
		return
	}
	defer resp.Body.Close()
	for _, key := range []string{"Content-Type", "Content-Disposition", "Content-Length", "Cache-Control", "Content-Range", "Accept-Ranges", "Last-Modified", "X-Content-Type-Options", "Content-Security-Policy"} {
		if v := resp.Header.Get(key); v != "" {
			w.Header().Set(key, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (r *Router) proxyWebSocket(w http.ResponseWriter, req *http.Request) {
	node, err := r.auth.db.GetNode(getURLParam(req, "id"))
	if err != nil {
		writeError(w, 404, "Node not found")
		return
	}
	u, err := url.Parse(node.BaseURL)
	if err != nil {
		writeError(w, 502, "Invalid node URL")
		return
	}
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/api/servers/" + getURLParam(req, "serverId") + "/ws"
	u.RawQuery = ""
	header := http.Header{}
	claims, _ := req.Context().Value(contextUserKey).(*JWTClaims)
	// Capture the same original credential selected by requireAuth. The node
	// connection uses a machine credential and cannot revoke this client session.
	token := ""
	if value := req.Header.Get("Authorization"); strings.HasPrefix(value, "Bearer ") {
		token = strings.TrimPrefix(value, "Bearer ")
	}
	if token == "" {
		if cookie, err := req.Cookie("neopanel_token"); err == nil {
			token = cookie.Value
		}
	}
	if token == "" {
		token = strings.TrimSpace(req.URL.Query().Get("token"))
	}
	revalidate := func() bool {
		_, err := r.auth.revalidateConsoleClaims(req, token, claims)
		return err == nil
	}
	if !revalidate() {
		writeError(w, http.StatusUnauthorized, "Console session expired; reconnect required")
		return
	}
	if err := r.checkProxyIdentity(req, node, claims); err != nil {
		writeError(w, 502, err.Error())
		return
	}
	setProxyIdentity(header, claims, node.APIToken)
	dialer := *websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second
	remote, resp, err := dialer.DialContext(req.Context(), u.String(), header)
	if err != nil {
		status := http.StatusBadGateway
		if resp != nil {
			if resp.Body != nil {
				resp.Body.Close()
			}
			if resp.StatusCode == 403 || resp.StatusCode == 404 {
				status = resp.StatusCode
			}
		}
		writeError(w, status, "Remote console access failed")
		return
	}
	defer remote.Close()
	client, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer client.Close()
	client.SetReadLimit(65536)
	stopAuthorization := watchConsoleAuthorization(revalidate, func() {
		_ = client.Close()
		_ = remote.Close()
	})
	defer stopAuthorization()
	done := make(chan struct{}, 2)
	relay := func(dst, src *websocket.Conn, fromClient bool) {
		defer func() { done <- struct{}{} }()
		for {
			kind, payload, err := src.ReadMessage()
			if err != nil {
				return
			}
			// Check before forwarding commands (and other client messages), not
			// for every high-volume log frame flowing in the opposite direction.
			if fromClient && !revalidate() {
				return
			}
			if err := dst.WriteMessage(kind, payload); err != nil {
				return
			}
		}
	}
	go relay(remote, client, true)
	go relay(client, remote, false)
	<-done
}

// Refuse user proxying to older nodes that would ignore the identity headers and
// execute the request with the privileged node token. No mutation is sent first.
func (r *Router) checkProxyIdentity(req *http.Request, node *database.ClusterNode, claims *JWTClaims) error {
	if claims.Role == "admin" {
		return nil
	}
	probe, err := http.NewRequestWithContext(req.Context(), http.MethodGet, strings.TrimRight(node.BaseURL, "/")+"/api/auth/me", nil)
	if err != nil {
		return err
	}
	setProxyIdentity(probe.Header, claims, node.APIToken)
	resp, err := r.nodes.httpClient.Do(probe)
	if err != nil {
		return fmt.Errorf("node identity check failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK || resp.Header.Get("X-NeoPanel-User-ID") != strconv.Itoa(claims.UserID) || resp.Header.Get("X-NeoPanel-Permission-Policy") != "2" {
		return fmt.Errorf("node must be updated to the current permission policy and user identities synchronized before user proxy access")
	}
	return nil
}

func (r *Router) listClusterServers(w http.ResponseWriter, req *http.Request) {
	claims, _ := req.Context().Value(contextUserKey).(*JWTClaims)
	if claims.Role == "admin" {
		r.nodes.ListClusterServers(w, req)
		return
	}
	local, err := r.auth.db.ListServersForUser(claims.UserID, claims.Role)
	if err != nil {
		writeError(w, 500, "Failed to list servers")
		return
	}
	items := []interface{}{}
	for _, srv := range local {
		items = append(items, struct {
			*database.Server
			Stats    database.ServerStats `json:"stats"`
			MOTD     string               `json:"motd"`
			NodeID   string               `json:"node_id"`
			NodeName string               `json:"node_name"`
			NodeFQDN string               `json:"node_fqdn"`
		}{srv, r.servers.manager.GetStats(srv), readServerMOTD(r.servers.manager.GetWorkingDir(srv.ID)), "local", "Local Master Node", r.nodes.resolveLocalFQDN(req.Host)})
	}
	nodes, err := r.auth.db.ListNodes()
	if err != nil {
		writeError(w, 500, "Failed to list nodes")
		return
	}
	nodeErrors := map[string]string{}
	for _, node := range nodes {
		list, err := r.listRemoteServers(req, node, claims)
		if err != nil {
			nodeErrors[node.ID] = err.Error()
			continue
		}
		for _, item := range list {
			item["node_id"], item["node_name"], item["node_fqdn"] = node.ID, node.Name, node.FQDN
			items = append(items, item)
		}
	}
	writeJSON(w, 200, map[string]interface{}{"servers": items, "node_errors": nodeErrors})
}

func (r *Router) listRemoteServers(req *http.Request, node *database.ClusterNode, claims *JWTClaims) ([]map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	req = req.WithContext(ctx)
	if err := r.checkProxyIdentity(req, node, claims); err != nil {
		return nil, err
	}
	remote, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(node.BaseURL, "/")+"/api/servers", nil)
	if err != nil {
		return nil, err
	}
	setProxyIdentity(remote.Header, claims, node.APIToken)
	resp, err := r.nodes.httpClient.Do(remote)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("node returned HTTP %d", resp.StatusCode)
	}
	var items []map[string]interface{}
	err = json.NewDecoder(io.LimitReader(resp.Body, 16<<20)).Decode(&items)
	return items, err
}
