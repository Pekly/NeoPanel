package api

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/downloader"
	"github.com/Pekly/NeoPanel/internal/rooted"
	"github.com/Pekly/NeoPanel/internal/runner"
)

type ServerHandler struct {
	cfg        *config.Config
	db         *database.DB
	manager    *runner.ServerManager
	httpClient *http.Client
	allowPort  func(int, string)
	removePort func(int)
}

func NewServerHandler(cfg *config.Config, db *database.DB, manager *runner.ServerManager) *ServerHandler {
	return &ServerHandler{
		cfg:        cfg,
		db:         db,
		manager:    manager,
		httpClient: &http.Client{Timeout: 6 * time.Second},
		allowPort:  allowHostPort,
		removePort: removeHostPort,
	}
}

type ServerWithStats struct {
	*database.Server
	Stats database.ServerStats `json:"stats"`
	MOTD  string               `json:"motd,omitempty"`
}

func readServerMOTD(workingDir string) string {
	if strings.TrimSpace(workingDir) == "" {
		return ""
	}
	f, err := rooted.OpenFile(workingDir, "server.properties")
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == "motd" {
			return strings.TrimSpace(parts[1])
		}
	}
	return ""
}

func (h *ServerHandler) ListServers(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	var userID int
	role := "user"
	if claims != nil {
		userID = claims.UserID
		role = claims.Role
	}
	servers, err := h.db.ListServersForUser(userID, role)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	result := make([]ServerWithStats, len(servers))
	for i, s := range servers {
		workDir := h.manager.GetWorkingDir(s.ID)
		if s.IconURL == "" {
			if s.NodeID != "" && s.NodeID != "local" {
				s.IconURL = fmt.Sprintf("/api/servers/%s/icon", s.ID)
			} else if fi, err := os.Stat(filepath.Join(workDir, "server-icon.png")); err == nil && !fi.IsDir() {
				s.IconURL = fmt.Sprintf("/api/servers/%s/icon", s.ID)
			}
		}
		result[i] = ServerWithStats{
			Server: s,
			Stats:  h.manager.GetStats(s),
			MOTD:   readServerMOTD(workDir),
		}
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *ServerHandler) GetServer(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	server, err := h.db.GetServer(serverID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims != nil && claims.Role != "admin" {
		if server.OwnerID != claims.UserID {
			member, _ := h.db.IsServerSubuser(server.ID, claims.UserID)
			if !member {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "Access denied to this server"})
				return
			}
		}
	}

	workDir := h.manager.GetWorkingDir(server.ID)
	if server.IconURL == "" {
		if server.NodeID != "" && server.NodeID != "local" {
			server.IconURL = fmt.Sprintf("/api/servers/%s/icon", server.ID)
		} else if fi, err := os.Stat(filepath.Join(workDir, "server-icon.png")); err == nil && !fi.IsDir() {
			server.IconURL = fmt.Sprintf("/api/servers/%s/icon", server.ID)
		}
	}

	writeJSON(w, http.StatusOK, ServerWithStats{
		Server: server,
		Stats:  h.manager.GetStats(server),
		MOTD:   readServerMOTD(h.manager.GetWorkingDir(server.ID)),
	})
}

type CreateServerRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	Type            string `json:"type"`       // paper, purpur, vanilla, fabric, custom
	Version         string `json:"version"`    // e.g. 1.20.4, 1.21
	Port            int    `json:"port"`       // default 25565
	MemoryMinMB     int    `json:"memory_min"` // e.g. 1024
	MemoryMaxMB     int    `json:"memory_max"` // e.g. 2048
	JavaArgs        string `json:"java_args"`  // custom flags
	JavaPath        string `json:"java_path"`  // custom java path
	AutoRestart     bool   `json:"auto_restart"`
	AutoDownload    bool   `json:"auto_download"` // true to download jar immediately
	OwnerID         int    `json:"owner_id"`
	CPULimitPercent int    `json:"cpu_limit_percent"`
	DiskLimitMB     int64  `json:"disk_limit_mb"`
}

func (h *ServerHandler) CreateServer(w http.ResponseWriter, r *http.Request) {
	allowed, err := canCreateServer(h.db, r)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unable to resolve server creation permission"})
		return
	}
	if !allowed {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Server creation requires the servers.create permission"})
		return
	}
	var req CreateServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Server name is required"})
		return
	}

	if req.Port <= 0 || req.Port > 65535 {
		req.Port = 25565
	}
	if req.MemoryMinMB <= 0 {
		req.MemoryMinMB = 1024
	}
	if req.MemoryMaxMB <= 0 {
		req.MemoryMaxMB = 2048
	}
	req.Type = strings.ToLower(strings.TrimSpace(req.Type))
	if req.Type == "" {
		req.Type = "paper"
	}
	if strings.TrimSpace(req.Version) == "" {
		if downloader.IsNativeSoftware(req.Type) {
			req.Version = "latest"
		} else {
			req.Version = "1.20.4"
		}
	}
	if downloader.IsNativeSoftware(req.Type) {
		req.MemoryMinMB, req.JavaArgs, req.JavaPath = 0, "", ""
		if req.AutoDownload {
			if err := downloader.ValidateSteelPlatform(); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
		}
	}

	// Generate clean unique server ID (e.g. srv-3f8a1e)
	idBytes := make([]byte, 3)
	_, _ = rand.Read(idBytes)
	srvID := fmt.Sprintf("srv-%s", hex.EncodeToString(idBytes))

	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	ownerID := 1
	if claims != nil && claims.UserID > 0 {
		ownerID = claims.UserID
	}
	if req.OwnerID > 0 && claims != nil && claims.Role == "admin" {
		ownerID = req.OwnerID
	}

	server := &database.Server{
		ID:              srvID,
		Name:            req.Name,
		Description:     req.Description,
		Type:            req.Type,
		Version:         req.Version,
		JarFile:         downloader.ServerFileName(req.Type),
		Port:            req.Port,
		MemoryMinMB:     req.MemoryMinMB,
		MemoryMaxMB:     req.MemoryMaxMB,
		JavaArgs:        req.JavaArgs,
		JavaPath:        req.JavaPath,
		AutoRestart:     req.AutoRestart,
		OwnerID:         ownerID,
		CPULimitPercent: req.CPULimitPercent,
		DiskLimitMB:     req.DiskLimitMB,
	}

	workingDir := h.manager.GetWorkingDir(srvID)
	if err := os.MkdirAll(workingDir, 0755); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create server directory"})
		return
	}

	if downloader.IsNativeSoftware(req.Type) {
		// Native installs are synchronous so failures are visible and start cannot race a download.
		if req.AutoDownload {
			if err := downloader.DownloadServerJar(req.Type, req.Version, workingDir, server.JarFile); err != nil {
				_ = os.RemoveAll(workingDir)
				writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Steel installation failed: " + err.Error()})
				return
			}
			if err := runner.SyncSteelPort(workingDir, req.Port); err != nil {
				writeServerOperationError(w, err)
				return
			}
		}
	} else {
		// Create initial server.properties with the chosen port
		root, err := rooted.Open(workingDir)
		if err != nil {
			writeServerOperationError(w, err)
			return
		}
		defer root.Close()
		if _, err := root.Lstat("server.properties"); os.IsNotExist(err) {
			propsContent := fmt.Sprintf("# Minecraft server properties generated by NeoPanel\nserver-port=%d\nquery.port=%d\nmotd=A NeoPanel Minecraft Server\nenable-rcon=false\n", req.Port, req.Port)
			if err := root.WriteFile("server.properties", []byte(propsContent), 0644); err != nil {
				writeServerOperationError(w, err)
				return
			}
		}

		// Agree to EULA
		if err := root.WriteFile("eula.txt", []byte("eula=true\n"), 0644); err != nil {
			writeServerOperationError(w, err)
			return
		}

		// Auto-download jar if requested and flavor is not custom
		if req.AutoDownload && req.Type != "custom" {
			go func() {
				_ = downloader.DownloadServerJar(req.Type, req.Version, workingDir, server.JarFile)
			}()
		}
	}

	if err := h.db.CreateServer(server); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save server: " + err.Error()})
		return
	}

	// Record default primary port allocation in database
	_, _ = h.db.AddServerPort(server.ID, server.Port, "tcp/udp", "Primary Game Port", true)

	logAudit(h.db, r, "server_create", fmt.Sprintf("Created server '%s' (%s %s, port %d, %d MiB RAM)", server.Name, server.Type, server.Version, server.Port, server.MemoryMaxMB))

	writeJSON(w, http.StatusCreated, ServerWithStats{
		Server: server,
		Stats:  h.manager.GetStats(server),
		MOTD:   readServerMOTD(workingDir),
	})
}

func (h *ServerHandler) UpdateServer(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	server, err := h.db.GetServer(serverID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	var req struct {
		database.Server
		Description     *string `json:"description"`
		JavaArgs        *string `json:"java_args"`
		JavaPath        *string `json:"java_path"`
		AutoRestart     *bool   `json:"auto_restart"`
		CPULimitPercent *int    `json:"cpu_limit_percent"`
		DiskLimitMB     *int64  `json:"disk_limit_mb"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	user := getUserFromContext(r)
	isAdmin := user != nil && user.Role == "admin"
	if req.NodeID != "" {
		currentNode := server.NodeID
		if currentNode == "" {
			currentNode = "local"
		}
		if req.NodeID != currentNode {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "Changing node_id does not migrate a running server and is not supported. Stop and transfer the server to the destination node instead."})
			return
		}
	}

	// Enforce node-level restrictions for normal / subusers
	if !isAdmin {
		node := h.serverUserPolicy(r, server)
		if node != nil {
			if !node.AllowUserRAM && ((req.MemoryMaxMB > 0 && req.MemoryMaxMB != server.MemoryMaxMB) || (req.MemoryMinMB > 0 && req.MemoryMinMB != server.MemoryMinMB)) {
				writeError(w, http.StatusForbidden, "Node policy prohibits non-admin users from altering server memory")
				return
			}
			// JVM options can load other options indirectly; do not attempt a
			// heap-flag denylist at this policy boundary.
			if !node.AllowUserRAM && req.JavaArgs != nil && *req.JavaArgs != server.JavaArgs {
				writeError(w, http.StatusForbidden, "Node policy prohibits non-admin users from altering JVM arguments while memory is locked")
				return
			}
			if !node.AllowUserPorts && (req.Port > 0 && req.Port != server.Port) {
				writeError(w, http.StatusForbidden, "Node policy prohibits non-admin users from altering server network ports")
				return
			}
			if !node.AllowUserJava && req.JavaPath != nil && *req.JavaPath != server.JavaPath {
				writeError(w, http.StatusForbidden, "Node policy prohibits non-admin users from altering Java runtime binary")
				return
			}
		}
	}

	if strings.TrimSpace(req.Name) != "" {
		server.Name = strings.TrimSpace(req.Name)
	}
	if req.Description != nil {
		server.Description = *req.Description
	}
	if req.Port > 0 {
		server.Port = req.Port
	}
	if req.Type != "" {
		server.Type = req.Type
	}
	if req.Version != "" {
		server.Version = req.Version
	}
	if req.MemoryMinMB > 0 {
		server.MemoryMinMB = req.MemoryMinMB
	}
	if req.MemoryMaxMB > 0 {
		server.MemoryMaxMB = req.MemoryMaxMB
	}
	if req.JavaArgs != nil {
		server.JavaArgs = *req.JavaArgs
	}
	if req.JavaPath != nil {
		server.JavaPath = *req.JavaPath
	}
	if req.AutoRestart != nil {
		server.AutoRestart = *req.AutoRestart
	}
	if req.JarFile != "" {
		server.JarFile = req.JarFile
	}
	if isAdmin {
		if req.OwnerID > 0 {
			server.OwnerID = req.OwnerID
		}
		if req.CPULimitPercent != nil && *req.CPULimitPercent >= 0 {
			server.CPULimitPercent = *req.CPULimitPercent
		}
		if req.DiskLimitMB != nil && *req.DiskLimitMB >= 0 {
			server.DiskLimitMB = *req.DiskLimitMB
		}
	}

	if err := h.db.UpdateServer(server); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update server: " + err.Error()})
		return
	}

	h.manager.UpdateInstanceServer(server)

	logAudit(h.db, r, "server_update", fmt.Sprintf("Updated settings for server '%s' (%s)", server.Name, server.ID))

	writeJSON(w, http.StatusOK, server)
}

func (h *ServerHandler) DeleteServer(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	server, err := h.db.GetServer(serverID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	user := getUserFromContext(r)
	if user != nil && user.Role != "admin" && server.OwnerID != user.ID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Only administrators or the server owner can delete this server instance"})
		return
	}
	// Master-side remote metadata is not proof of shutdown on its worker.
	// Never silently erase it using only this host's process manager.
	if server.NodeID != "" && server.NodeID != "local" && (h.cfg == nil || h.cfg.Server.Role != "node") {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "Deletion refused: this record is assigned to another node. Delete the server on its hosting worker and reconcile this record only after confirming shutdown."})
		return
	}

	err = h.manager.DeleteServer(r.Context(), serverID, func() error {
		// The manager retains the restore/power gate through all disk and DB work.
		if err := os.RemoveAll(h.manager.GetWorkingDir(serverID)); err != nil {
			return fmt.Errorf("remove server files: %w", err)
		}
		if h.cfg != nil && h.cfg.Paths.DataDir != "" {
			if err := os.RemoveAll(filepath.Join(h.cfg.Paths.DataDir, "databases", serverID)); err != nil {
				return fmt.Errorf("remove server databases: %w", err)
			}
		}
		if err := h.db.DeleteAllServerDatabases(serverID); err != nil {
			return err
		}
		if err := h.db.DeleteAllServerPorts(serverID); err != nil {
			return err
		}
		return h.db.DeleteServer(serverID)
	})
	if err != nil {
		writeServerOperationError(w, err)
		return
	}

	logAudit(h.db, r, "server_delete", fmt.Sprintf("Deleted server '%s' (%s)", server.Name, server.ID))

	writeJSON(w, http.StatusOK, map[string]interface{}{"message": "Server deleted successfully", "deleted": true, "server_id": serverID})
}

func (h *ServerHandler) PowerAction(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	server, err := h.db.GetServer(serverID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Server not found"})
		return
	}

	var req struct {
		Action string `json:"action"` // start, stop, restart, kill
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid action request"})
		return
	}

	var actionErr error
	switch req.Action {
	case "start":
		actionErr = h.manager.Start(server)
	case "stop":
		actionErr = h.manager.Stop(server.ID, true)
	case "restart":
		actionErr = h.manager.Restart(server)
	case "kill":
		actionErr = h.manager.Stop(server.ID, false)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Unknown action: " + req.Action})
		return
	}

	if actionErr != nil {
		writeServerOperationError(w, actionErr)
		return
	}

	logAudit(h.db, r, "power_"+req.Action, fmt.Sprintf("Triggered %s on server '%s' (%s)", req.Action, server.Name, server.ID))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"action":  req.Action,
		"current": h.manager.GetStats(server),
	})
}

func (h *ServerHandler) SendCommand(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	var req struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Command == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid command"})
		return
	}

	if err := h.manager.SendCommand(serverID, req.Command); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	logAudit(h.db, r, "console_command", fmt.Sprintf("Executed command on %s: %s", auditServerLabel(h.db, serverID), truncateAuditText(req.Command, 120)))

	writeJSON(w, http.StatusOK, map[string]string{"status": "command sent"})
}

// --- Port Allocation Endpoints ---

func (h *ServerHandler) ListServerPorts(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	ports, err := h.db.ListServerPorts(serverID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, ports)
}

type AddPortRequest struct {
	Port      int    `json:"port"`
	Protocol  string `json:"protocol"`
	Notes     string `json:"notes"`
	IsPrimary bool   `json:"is_primary"`
}

// Prefer the policy carried by trusted forwarding middleware: a worker's
// local/empty node ID need not identify the master's node policy row.
func (h *ServerHandler) serverUserPolicy(r *http.Request, server *database.Server) *nodeUserPolicy {
	if claims, _ := r.Context().Value(contextUserKey).(*JWTClaims); claims != nil && claims.ForwardedNodePolicy != nil {
		return claims.ForwardedNodePolicy
	}
	nodeID := server.NodeID
	if nodeID == "" {
		nodeID = "local"
	}
	node, _ := h.db.GetNode(nodeID)
	return userPolicyForNode(node)
}

// Check policy before decoding or performing any database, file or firewall mutation.
func (h *ServerHandler) requirePortPolicy(w http.ResponseWriter, r *http.Request, serverID string) bool {
	server, err := h.db.GetServer(serverID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Server not found")
		return false
	}
	if user := getUserFromContext(r); user != nil && user.Role == "admin" {
		return true
	}
	// Match UpdateServer, including trusted master policy on remote workers.
	if node := h.serverUserPolicy(r, server); node != nil && !node.AllowUserPorts {
		writeError(w, http.StatusForbidden, "Node policy prohibits non-admin users from altering server network ports")
		return false
	}
	return true
}

func (h *ServerHandler) AddServerPort(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	if !h.requirePortPolicy(w, r, serverID) {
		return
	}
	var req AddPortRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Port < 1024 || req.Port > 65535 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Port number must be between 1024 and 65535"})
		return
	}

	port, err := h.db.AddServerPort(serverID, req.Port, req.Protocol, req.Notes, req.IsPrimary)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// If set as primary, also update server.properties and in-memory instance
	if req.IsPrimary {
		if err := h.updateServerPropertiesPort(serverID, req.Port); err != nil {
			writeServerOperationError(w, err)
			return
		}
		if srv, err := h.db.GetServer(serverID); err == nil {
			h.manager.UpdateInstanceServer(srv)
		}
	}

	// Open port in host Linux firewall if UFW is active
	h.allowPort(req.Port, req.Protocol)

	logAudit(h.db, r, "port_add", fmt.Sprintf("Assigned port %d (%s) to %s", req.Port, port.Protocol, auditServerLabel(h.db, serverID)))

	writeJSON(w, http.StatusCreated, port)
}

func (h *ServerHandler) DeleteServerPort(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	if !h.requirePortPolicy(w, r, serverID) {
		return
	}
	portIDStr := getURLParam(r, "portId")
	portID, err := strconv.Atoi(portIDStr)
	if err != nil || portID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid port ID"})
		return
	}

	// Retrieve port number for firewall rule cleanup
	ports, _ := h.db.ListServerPorts(serverID)
	var portNum int
	for _, p := range ports {
		if p.ID == portID {
			portNum = p.Port
			break
		}
	}

	if err := h.db.DeleteServerPort(serverID, portID); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if portNum > 0 {
		h.removePort(portNum)
	}

	logAudit(h.db, r, "port_delete", fmt.Sprintf("Removed port %d allocation (ID %d) from %s", portNum, portID, auditServerLabel(h.db, serverID)))

	writeJSON(w, http.StatusOK, map[string]string{"message": "Port allocation deleted successfully"})
}

func (h *ServerHandler) SetPrimaryPort(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	if !h.requirePortPolicy(w, r, serverID) {
		return
	}
	portIDStr := getURLParam(r, "portId")
	portID, err := strconv.Atoi(portIDStr)
	if err != nil || portID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid port ID"})
		return
	}

	newPort, err := h.db.SetPrimaryServerPort(serverID, portID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Update server.properties with the new primary port
	if err := h.updateServerPropertiesPort(serverID, newPort); err != nil {
		writeServerOperationError(w, err)
		return
	}

	// Live sync in-memory server instance
	if srv, err := h.db.GetServer(serverID); err == nil {
		h.manager.UpdateInstanceServer(srv)
	}

	logAudit(h.db, r, "port_primary", fmt.Sprintf("Set port %d as primary game port on %s", newPort, auditServerLabel(h.db, serverID)))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Primary port updated",
		"port":    newPort,
	})
}

func allowHostPort(port int, protocol string) {
	if runtime.GOOS != "linux" {
		return
	}
	if _, err := exec.LookPath("ufw"); err == nil {
		proto := strings.ToLower(protocol)
		if strings.Contains(proto, "udp") && !strings.Contains(proto, "tcp") {
			_ = exec.Command("ufw", "allow", fmt.Sprintf("%d/udp", port)).Run()
		} else if strings.Contains(proto, "tcp") && !strings.Contains(proto, "udp") {
			_ = exec.Command("ufw", "allow", fmt.Sprintf("%d/tcp", port)).Run()
		} else {
			_ = exec.Command("ufw", "allow", fmt.Sprintf("%d", port)).Run()
		}
	}
}

func removeHostPort(port int) {
	if runtime.GOOS != "linux" {
		return
	}
	if _, err := exec.LookPath("ufw"); err == nil {
		_ = exec.Command("ufw", "delete", "allow", fmt.Sprintf("%d", port)).Run()
		_ = exec.Command("ufw", "delete", "allow", fmt.Sprintf("%d/tcp", port)).Run()
		_ = exec.Command("ufw", "delete", "allow", fmt.Sprintf("%d/udp", port)).Run()
	}
}

func (h *ServerHandler) updateServerPropertiesPort(serverID string, port int) error {
	root, err := rooted.Open(h.manager.GetWorkingDir(serverID))
	if err != nil {
		return err
	}
	defer root.Close()
	content, err := root.ReadFile("server.properties")
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	lines := strings.Split(string(content), "\n")
	hasPort := false
	hasQuery := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "server-port=") {
			lines[i] = fmt.Sprintf("server-port=%d", port)
			hasPort = true
		} else if strings.HasPrefix(trimmed, "query.port=") {
			lines[i] = fmt.Sprintf("query.port=%d", port)
			hasQuery = true
		}
	}
	if !hasPort {
		lines = append(lines, fmt.Sprintf("server-port=%d", port))
	}
	if !hasQuery {
		lines = append(lines, fmt.Sprintf("query.port=%d", port))
	}
	return root.WriteFile("server.properties", []byte(strings.Join(lines, "\n")), 0644)
}

// UploadServerIcon saves a custom server-icon.png for the server
func (h *ServerHandler) UploadServerIcon(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	server, err := h.db.GetServer(serverID)
	if err != nil || server == nil {
		writeError(w, http.StatusNotFound, "Server not found")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)
	defer func() {
		if r.MultipartForm != nil {
			r.MultipartForm.RemoveAll()
		}
	}()
	if err := r.ParseMultipartForm(10 * 1024 * 1024); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid image upload (max 10MB)")
		return
	}

	file, header, err := r.FormFile("icon")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Missing 'icon' form field")
		return
	}
	defer file.Close()

	workDir := h.manager.GetWorkingDir(serverID)
	_ = os.MkdirAll(workDir, 0755)
	destPath := filepath.Join(workDir, "server-icon.png")

	if err := processAndSaveServerIcon(file, destPath); err != nil {
		writeError(w, http.StatusBadRequest, "Failed to process icon (must be a valid PNG, JPG, or GIF): "+err.Error())
		return
	}

	iconURL := fmt.Sprintf("/api/servers/%s/icon?t=%d", serverID, time.Now().Unix())
	_ = h.db.UpdateServerIcon(serverID, iconURL)

	if server.NodeID != "" && server.NodeID != "local" {
		if node, _ := h.db.GetNode(server.NodeID); node != nil && node.BaseURL != "" {
			go func() {
				if f, err := rooted.OpenFile(workDir, "server-icon.png"); err == nil {
					defer f.Close()
					var b bytes.Buffer
					mw := multipart.NewWriter(&b)
					if part, err := mw.CreateFormFile("icon", "server-icon.png"); err == nil {
						_, _ = io.Copy(part, f)
						_ = mw.Close()
						targetURL := strings.TrimRight(node.BaseURL, "/") + "/api/servers/" + url.PathEscape(serverID) + "/icon"
						req, err := http.NewRequest(http.MethodPost, targetURL, &b)
						if err == nil {
							req.Header.Set("Authorization", "Bearer "+node.APIToken)
							req.Header.Set("Content-Type", mw.FormDataContentType())
							client := &http.Client{Timeout: 10 * time.Second}
							resp, err := client.Do(req)
							if err == nil {
								_ = resp.Body.Close()
							}
						}
					}
				}
			}()
		}
	}

	logAudit(h.db, r, "server_icon_upload", fmt.Sprintf("Uploaded custom 64x64 server-icon.png for %s (%s)", server.Name, header.Filename))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "success",
		"icon_url": iconURL,
	})
}

// processAndSaveServerIcon decodes an uploaded image, center-crops to square,
// resizes to exactly 64x64 pixels (Minecraft specification), and encodes as PNG.
func processAndSaveServerIcon(r io.Reader, destPath string) error {
	img, _, err := image.Decode(r)
	if err != nil {
		return fmt.Errorf("unable to decode image: %w", err)
	}

	resized := cropAndResizeTo64x64(img)

	root, err := rooted.Open(filepath.Dir(destPath))
	if err != nil {
		return err
	}
	defer root.Close()
	return root.Publish(filepath.Base(destPath), 0644, func(f *os.File) error { return png.Encode(f, resized) })
}

func cropAndResizeTo64x64(src image.Image) *image.RGBA {
	const targetW = 64
	const targetH = 64
	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))

	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	if srcW == targetW && srcH == targetH {
		draw.Draw(dst, dst.Bounds(), src, bounds.Min, draw.Src)
		return dst
	}

	minDim := srcW
	if srcH < minDim {
		minDim = srcH
	}

	cropX := bounds.Min.X + (srcW-minDim)/2
	cropY := bounds.Min.Y + (srcH-minDim)/2

	for y := 0; y < targetH; y++ {
		srcY := float64(cropY) + (float64(y)*float64(minDim))/float64(targetH)
		y0 := int(math.Floor(srcY))
		y1 := y0 + 1
		if y1 >= bounds.Max.Y {
			y1 = bounds.Max.Y - 1
		}
		fy := srcY - float64(y0)

		for x := 0; x < targetW; x++ {
			srcX := float64(cropX) + (float64(x)*float64(minDim))/float64(targetW)
			x0 := int(math.Floor(srcX))
			x1 := x0 + 1
			if x1 >= bounds.Max.X {
				x1 = bounds.Max.X - 1
			}
			fx := srcX - float64(x0)

			c00 := colorToRGBA(src.At(x0, y0))
			c10 := colorToRGBA(src.At(x1, y0))
			c01 := colorToRGBA(src.At(x0, y1))
			c11 := colorToRGBA(src.At(x1, y1))

			// Bilinear interpolation
			r0 := float64(c00.R)*(1-fx) + float64(c10.R)*fx
			g0 := float64(c00.G)*(1-fx) + float64(c10.G)*fx
			b0 := float64(c00.B)*(1-fx) + float64(c10.B)*fx
			a0 := float64(c00.A)*(1-fx) + float64(c10.A)*fx

			r1 := float64(c01.R)*(1-fx) + float64(c11.R)*fx
			g1 := float64(c01.G)*(1-fx) + float64(c11.G)*fx
			b1 := float64(c01.B)*(1-fx) + float64(c11.B)*fx
			a1 := float64(c01.A)*(1-fx) + float64(c11.A)*fx

			r := uint8(math.Round(r0*(1-fy) + r1*fy))
			g := uint8(math.Round(g0*(1-fy) + g1*fy))
			b := uint8(math.Round(b0*(1-fy) + b1*fy))
			a := uint8(math.Round(a0*(1-fy) + a1*fy))

			dst.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}
	return dst
}

func colorToRGBA(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	if a == 0 {
		return color.RGBA{0, 0, 0, 0}
	}
	return color.RGBA{
		R: uint8((r * 255) / a),
		G: uint8((g * 255) / a),
		B: uint8((b * 255) / a),
		A: uint8(a >> 8),
	}
}

// GetServerIcon returns the server's server-icon.png
func (h *ServerHandler) GetServerIcon(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	server, _ := h.db.GetServer(serverID)
	if server != nil && server.NodeID != "" && server.NodeID != "local" {
		node, _ := h.db.GetNode(server.NodeID)
		if node != nil && node.BaseURL != "" {
			targetURL := strings.TrimRight(node.BaseURL, "/") + "/api/servers/" + url.PathEscape(serverID) + "/icon"
			if r.URL.RawQuery != "" {
				targetURL += "?" + r.URL.RawQuery
			}
			req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, targetURL, nil)
			if err == nil {
				if node.APIToken != "" {
					req.Header.Set("Authorization", "Bearer "+node.APIToken)
				}
				resp, err := h.httpClient.Do(req)
				if err == nil {
					defer resp.Body.Close()
					if resp.StatusCode == http.StatusOK {
						if ct := resp.Header.Get("Content-Type"); ct != "" {
							w.Header().Set("Content-Type", ct)
						}
						w.Header().Set("Cache-Control", "no-cache, must-revalidate")
						w.WriteHeader(http.StatusOK)
						_, _ = io.Copy(w, resp.Body)
						return
					}
				}
			}
		}
	}

	workDir := h.manager.GetWorkingDir(serverID)

	if f, err := rooted.OpenFile(workDir, "server-icon.png"); err == nil {
		defer f.Close()
		fi, err := f.Stat()
		if err != nil {
			writeError(w, 400, err.Error())
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
		http.ServeContent(w, r, fi.Name(), fi.ModTime(), f)
		return
	}

	if r.URL.Query().Get("default") == "none" {
		http.NotFound(w, r)
		return
	}

	defaultPath := filepath.Join("web", "static", "img", "default-server-icon.png")
	if fi, err := os.Stat(defaultPath); err == nil && !fi.IsDir() {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		http.ServeFile(w, r, defaultPath)
		return
	}

	http.NotFound(w, r)
}

// DeleteServerIcon removes the custom icon
func (h *ServerHandler) DeleteServerIcon(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	server, _ := h.db.GetServer(serverID)
	if server != nil && server.NodeID != "" && server.NodeID != "local" {
		if node, _ := h.db.GetNode(server.NodeID); node != nil && node.BaseURL != "" {
			targetURL := strings.TrimRight(node.BaseURL, "/") + "/api/servers/" + url.PathEscape(serverID) + "/icon"
			req, err := http.NewRequestWithContext(r.Context(), http.MethodDelete, targetURL, nil)
			if err == nil {
				req.Header.Set("Authorization", "Bearer "+node.APIToken)
				_, _ = h.httpClient.Do(req)
			}
		}
	}

	workDir := h.manager.GetWorkingDir(serverID)
	root, err := rooted.Open(workDir)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	defer root.Close()
	if err := root.Remove("server-icon.png"); err != nil && !os.IsNotExist(err) {
		writeError(w, 400, err.Error())
		return
	}
	_ = h.db.UpdateServerIcon(serverID, "")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
	})
}

// GetQuickConfigs reads server.properties and eula.txt for easy web editing
func (h *ServerHandler) GetQuickConfigs(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	if rejectSteelJavaConfig(h.db, serverID, w) {
		return
	}
	workDir := h.manager.GetWorkingDir(serverID)
	root, err := rooted.Open(workDir)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	defer root.Close()

	configs := map[string]interface{}{
		"view_distance":       10,
		"simulation_distance": 8,
		"spawn_protection":    16,
		"allow_flight":        false,
		"pvp":                 true,
		"difficulty":          "easy",
		"gamemode":            "survival",
		"whitelist":           false,
		"command_blocks":      false,
		"online_mode":         true,
		"motd":                "A Minecraft Server",
		"eula_accepted":       false,
	}

	eulaData, err := root.ReadFile("eula.txt")
	if err != nil && !os.IsNotExist(err) {
		writeError(w, 400, err.Error())
		return
	}
	if err == nil {
		if strings.Contains(strings.ToLower(string(eulaData)), "eula=true") {
			configs["eula_accepted"] = true
		}
	}

	data, err := root.ReadFile("server.properties")
	if err != nil && !os.IsNotExist(err) {
		writeError(w, 400, err.Error())
		return
	}
	if err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])

			switch k {
			case "motd":
				configs["motd"] = v
			case "view-distance":
				if iv, err := strconv.Atoi(v); err == nil {
					configs["view_distance"] = iv
				}
			case "simulation-distance":
				if iv, err := strconv.Atoi(v); err == nil {
					configs["simulation_distance"] = iv
				}
			case "spawn-protection":
				if iv, err := strconv.Atoi(v); err == nil {
					configs["spawn_protection"] = iv
				}
			case "allow-flight":
				configs["allow_flight"] = v == "true"
			case "pvp":
				configs["pvp"] = v == "true"
			case "difficulty":
				configs["difficulty"] = v
			case "gamemode":
				configs["gamemode"] = v
			case "white-list":
				configs["whitelist"] = v == "true"
			case "enable-command-block":
				configs["command_blocks"] = v == "true"
			case "online-mode":
				configs["online_mode"] = v == "true"
			}
		}
	}

	writeJSON(w, http.StatusOK, configs)
}

// UpdateQuickConfigs updates key server properties and applies optimization presets
func (h *ServerHandler) UpdateQuickConfigs(w http.ResponseWriter, r *http.Request) {
	serverID := getURLParam(r, "id")
	if rejectSteelJavaConfig(h.db, serverID, w) {
		return
	}
	workDir := h.manager.GetWorkingDir(serverID)
	root, err := rooted.Open(workDir)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	defer root.Close()

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	for k, v := range req {
		if !rooted.PropertyText(k) || !rooted.PropertyText(fmt.Sprint(v)) {
			writeError(w, 400, "Configuration must not contain line controls")
			return
		}
	}

	// 1. EULA toggle
	if val, ok := req["eula_accepted"]; ok {
		if bVal, ok := val.(bool); ok && bVal {
			if err := root.WriteFile("eula.txt", []byte("eula=true\n"), 0644); err != nil {
				writeError(w, 400, err.Error())
				return
			}
		} else {
			if err := root.WriteFile("eula.txt", []byte("eula=false\n"), 0644); err != nil {
				writeError(w, 400, err.Error())
				return
			}
		}
	}

	// 2. Read existing server.properties lines
	existing := make(map[string]string)
	orderedKeys := []string{}
	data, err := root.ReadFile("server.properties")
	if err != nil && !os.IsNotExist(err) {
		writeError(w, 400, err.Error())
		return
	}
	if err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			if _, exists := existing[k]; !exists {
				orderedKeys = append(orderedKeys, k)
			}
			existing[k] = v
		}
	}

	// Apply presets if specified
	if preset, ok := req["preset"].(string); ok {
		switch preset {
		case "extreme-performance":
			existing["view-distance"] = "4"
			existing["simulation-distance"] = "3"
			existing["spawn-protection"] = "0"
			existing["network-compression-threshold"] = "256"
		case "balanced":
			existing["view-distance"] = "8"
			existing["simulation-distance"] = "6"
			existing["spawn-protection"] = "16"
		case "ultra":
			existing["view-distance"] = "16"
			existing["simulation-distance"] = "12"
			existing["spawn-protection"] = "16"
		}
	}

	// Map incoming keys to Minecraft server.properties keys
	keyMap := map[string]string{
		"view_distance":       "view-distance",
		"simulation_distance": "simulation-distance",
		"spawn_protection":    "spawn-protection",
		"allow_flight":        "allow-flight",
		"pvp":                 "pvp",
		"difficulty":          "difficulty",
		"gamemode":            "gamemode",
		"whitelist":           "white-list",
		"command_blocks":      "enable-command-block",
		"online_mode":         "online-mode",
		"motd":                "motd",
	}

	for jsonKey, propKey := range keyMap {
		if val, ok := req[jsonKey]; ok {
			strVal := fmt.Sprintf("%v", val)
			if _, exists := existing[propKey]; !exists {
				orderedKeys = append(orderedKeys, propKey)
			}
			existing[propKey] = strVal
		}
	}

	var sb strings.Builder
	sb.WriteString("# Minecraft server properties\n# Modified by NeoPanel Quick Config Editor\n")
	for _, k := range orderedKeys {
		if !rooted.PropertyText(k) || !rooted.PropertyText(existing[k]) {
			writeError(w, 400, "Configuration must not contain line controls")
			return
		}
		sb.WriteString(fmt.Sprintf("%s=%s\n", k, existing[k]))
	}

	if err := root.WriteFile("server.properties", []byte(sb.String()), 0644); err != nil {
		writeError(w, 400, err.Error())
		return
	}

	logAudit(h.db, r, "quick_config_update", fmt.Sprintf("Updated game configurations on server %s", serverID))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Game configuration updated successfully",
	})
}
