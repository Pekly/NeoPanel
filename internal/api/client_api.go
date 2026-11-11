package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
	"golang.org/x/crypto/bcrypt"
)

type ClientAPIHandler struct {
	cfg     *config.Config
	db      *database.DB
	manager *runner.ServerManager
}

func NewClientAPIHandler(cfg *config.Config, db *database.DB, manager *runner.ServerManager) *ClientAPIHandler {
	return &ClientAPIHandler{cfg: cfg, db: db, manager: manager}
}

// ProvisionRequest defines the incoming payload from a billing / client area portal
type ProvisionRequest struct {
	ClientEmail      string `json:"client_email"`
	ClientUsername   string `json:"client_username"`
	ServerName       string `json:"server_name"`
	NodeID           string `json:"node_id"`
	ServerType       string `json:"server_type"`
	MinecraftVersion string `json:"minecraft_version"`
	MemoryMB         int    `json:"memory_mb"`
	DiskMB           int64  `json:"disk_mb"`
	CPULimitPercent  int    `json:"cpu_limit_percent"`
	AutoStart        bool   `json:"auto_start"`
}

func (h *ClientAPIHandler) authenticateClientAPI(r *http.Request) bool {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		authHdr := r.Header.Get("Authorization")
		if strings.HasPrefix(authHdr, "Bearer ") {
			apiKey = strings.TrimPrefix(authHdr, "Bearer ")
		}
	}
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return false
	}

	// Accept node token or cluster secret
	if h.cfg.Security.NodeToken != "" && apiKey == h.cfg.Security.NodeToken {
		return true
	}
	if h.cfg.Security.JWTSecret != "" && apiKey == h.cfg.Security.JWTSecret {
		return true
	}
	// Accept active admin user session/token or API key hash
	keyHash := fmt.Sprintf("%x", sha256.Sum256([]byte(apiKey)))
	if user, err := h.db.GetUserByAPIKeyHash(keyHash); err == nil && user != nil && user.Role == "admin" {
		return true
	}
	if claims, err := NewAuthHandler(h.db, h.cfg).ValidateAuthToken(apiKey); err == nil && claims != nil && claims.Role == "admin" {
		return true
	}
	return false
}

// Provision handles automated server creation from billing platforms
func (h *ClientAPIHandler) Provision(w http.ResponseWriter, r *http.Request) {
	if !h.authenticateClientAPI(r) {
		writeError(w, http.StatusUnauthorized, "Invalid or missing API key in X-API-Key or Bearer token")
		return
	}

	var req ProvisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
		return
	}

	clientEmail := strings.TrimSpace(req.ClientEmail)
	if clientEmail == "" {
		writeError(w, http.StatusBadRequest, "client_email is required")
		return
	}

	username := strings.TrimSpace(req.ClientUsername)
	if username == "" {
		parts := strings.Split(clientEmail, "@")
		username = parts[0]
	}

	// 1. Locate or create client user
	user, err := h.db.GetUserByEmailOrUsername(clientEmail)
	if err != nil || user == nil {
		user, _ = h.db.GetUserByUsername(username)
	}
	generatedPassword := ""
	if user == nil {
		rawBytes := make([]byte, 8)
		_, _ = rand.Read(rawBytes)
		generatedPassword = hex.EncodeToString(rawBytes)

		hash, _ := bcrypt.GenerateFromPassword([]byte(generatedPassword), 12)
		newUser, createErr := h.db.CreateUser(username, string(hash), "user")
		if createErr != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create client user account: "+createErr.Error())
			return
		}
		user = newUser
	}

	// 2. Select target node
	nodeID := strings.TrimSpace(req.NodeID)
	var targetNode *database.ClusterNode
	if nodeID != "" {
		n, err := h.db.GetNode(nodeID)
		if err == nil && n != nil {
			targetNode = n
		}
	}
	if targetNode == nil {
		nodes, err := h.db.ListNodes()
		if err == nil && len(nodes) > 0 {
			targetNode = nodes[0]
		}
	}
	if targetNode == nil {
		targetNode = &database.ClusterNode{
			ID:   "local",
			Name: "Local Host",
			FQDN: "127.0.0.1",
		}
	}

	// 3. Find next available game port
	port := 25565
	existingServers, _ := h.db.ListServers()
	usedPorts := make(map[int]bool)
	for _, s := range existingServers {
		usedPorts[s.Port] = true
	}
	for usedPorts[port] {
		port++
	}

	// 4. Default memory & configuration
	memMB := req.MemoryMB
	if memMB <= 0 {
		memMB = 2048
	}
	srvType := strings.ToLower(strings.TrimSpace(req.ServerType))
	if srvType == "" {
		srvType = "paper"
	}
	srvName := strings.TrimSpace(req.ServerName)
	if srvName == "" {
		srvName = fmt.Sprintf("%s's Minecraft Server", user.Username)
	}

	srvID := fmt.Sprintf("srv-%x", time.Now().UnixNano()%0xFFFFFF)
	jarName := "server.jar"

	srv := database.Server{
		ID:              srvID,
		Name:            srvName,
		Description:     fmt.Sprintf("Provisioned via Client Area API for %s", clientEmail),
		Type:            srvType,
		Version:         req.MinecraftVersion,
		Port:            port,
		MemoryMinMB:     1024,
		MemoryMaxMB:     memMB,
		JavaPath:        "",
		JavaArgs:        "-XX:+UseG1GC -XX:+ParallelRefProcEnabled -XX:MaxGCPauseMillis=200",
		JarFile:         jarName,
		NodeID:          targetNode.ID,
		OwnerID:         user.ID,
		CPULimitPercent: req.CPULimitPercent,
		DiskLimitMB:     req.DiskMB,
	}

	err = h.db.CreateServer(&srv)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to provision server in database: "+err.Error())
		return
	}

	// Create working directory
	srvDir := h.manager.GetWorkingDir(srvID)
	_ = os.MkdirAll(srvDir, 0755)

	// Accept EULA
	_ = os.WriteFile(filepath.Join(srvDir, "eula.txt"), []byte("eula=true\n"), 0644)

	// Sync server.properties port
	props := fmt.Sprintf("server-port=%d\nquery.port=%d\nmotd=%s\n", port, port, srvName)
	_ = os.WriteFile(filepath.Join(srvDir, "server.properties"), []byte(props), 0644)

	_ = h.db.CreateAuditLog("client-api", "server.provision", fmt.Sprintf("Provisioned server %s (%s) for %s on node %s", srvName, srvID, clientEmail, targetNode.Name), getClientIP(r))

	publicIP := targetNode.FQDN
	if publicIP == "" {
		publicIP = "127.0.0.1"
	}

	resp := map[string]interface{}{
		"status":      "success",
		"server_id":   srv.ID,
		"server_name": srv.Name,
		"node_id":     targetNode.ID,
		"node_name":   targetNode.Name,
		"ip":          publicIP,
		"port":        srv.Port,
		"user_id":     user.ID,
		"user_email":  clientEmail,
		"username":    user.Username,
		"memory_mb":   srv.MemoryMaxMB,
		"login_url":   fmt.Sprintf("http://%s:%d/login", h.cfg.Server.Host, h.cfg.Server.Port),
		"console_url": fmt.Sprintf("http://%s:%d/server/%s/console", h.cfg.Server.Host, h.cfg.Server.Port, srv.ID),
	}
	if generatedPassword != "" {
		resp["temporary_password"] = generatedPassword
	}

	writeJSON(w, http.StatusCreated, resp)
}

// GetCodeSamples returns integration code snippets for external billing portals
func (h *ClientAPIHandler) GetCodeSamples(w http.ResponseWriter, r *http.Request) {
	panelURL := fmt.Sprintf("http://%s:%d", h.cfg.Server.Host, h.cfg.Server.Port)
	if h.cfg.Server.Host == "0.0.0.0" || h.cfg.Server.Host == "" {
		panelURL = "http://YOUR_PANEL_IP:" + fmt.Sprintf("%d", h.cfg.Server.Port)
	}

	apiKey := h.cfg.Security.NodeToken
	if apiKey == "" {
		apiKey = "YOUR_PANEL_API_KEY"
	}

	phpCode := fmt.Sprintf(`<?php
// NeoPanel Automated Server Provisioning (WHMCS / Blesta / Custom Hook)
$endpoint = '%s/api/client-api/v1/provision';
$apiKey   = '%s';

$payload = [
    'client_email'      => $params['clientsdetails']['email'] ?? 'client@example.com',
    'client_username'   => $params['clientsdetails']['username'] ?? 'clientuser',
    'server_name'       => $params['customfields']['ServerName'] ?? 'Minecraft Server',
    'node_id'           => 'local', // target cluster node
    'server_type'       => 'paper',
    'minecraft_version' => '1.21.4',
    'memory_mb'         => 4096,
    'disk_mb'           => 15360,
    'cpu_limit_percent' => 100,
    'auto_start'        => false
];

$ch = curl_init($endpoint);
curl_setopt_array($ch, [
    CURLOPT_POST           => true,
    CURLOPT_RETURNTRANSFER => true,
    CURLOPT_HTTPHEADER     => [
        'Content-Type: application/json',
        'X-API-Key: ' . $apiKey
    ],
    CURLOPT_POSTFIELDS     => json_encode($payload),
    CURLOPT_TIMEOUT        => 15
]);

$response = curl_exec($ch);
$httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
curl_close($ch);

$result = json_decode($response, true);
if ($httpCode === 201 && $result['status'] === 'success') {
    // Server provisioned successfully!
    $serverId = $result['server_id'];
    $serverPort = $result['port'];
    return 'success';
} else {
    throw new Exception('Provisioning failed: ' . ($result['error'] ?? $response));
}
`, panelURL, apiKey)

	nodeCode := fmt.Sprintf(`// NeoPanel Provisioning (Node.js / Express / Stripe Webhook)
const axios = require('axios');

async function provisionMinecraftServer(clientEmail, serverName, memoryMB = 4096) {
    const response = await axios.post('%s/api/client-api/v1/provision', {
        client_email: clientEmail,
        server_name: serverName,
        node_id: 'local',
        server_type: 'paper',
        minecraft_version: '1.21.4',
        memory_mb: memoryMB,
        disk_mb: 20480
    }, {
        headers: {
            'Content-Type': 'application/json',
            'X-API-Key': '%s'
        },
        timeout: 10000
    });

    console.log('Server created:', response.data);
    return response.data;
}
`, panelURL, apiKey)

	pythonCode := fmt.Sprintf(`# NeoPanel Provisioning (Python / FastAPI / Django)
import requests

def provision_server(client_email, server_name, memory_mb=4096):
    url = "%s/api/client-api/v1/provision"
    headers = {
        "Content-Type": "application/json",
        "X-API-Key": "%s"
    }
    payload = {
        "client_email": client_email,
        "server_name": server_name,
        "node_id": "local",
        "server_type": "paper",
        "minecraft_version": "1.21.4",
        "memory_mb": memory_mb,
        "disk_mb": 20480
    }
    res = requests.post(url, json=payload, headers=headers, timeout=15)
    res.raise_for_status()
    return res.json()
`, panelURL, apiKey)

	curlCode := fmt.Sprintf(`curl -X POST %s/api/client-api/v1/provision \
  -H "Content-Type: application/json" \
  -H "X-API-Key: %s" \
  -d '{
    "client_email": "gamer@example.com",
    "server_name": "Survival SMP",
    "node_id": "local",
    "server_type": "paper",
    "minecraft_version": "1.21.4",
    "memory_mb": 4096,
    "disk_mb": 20480
  }'`, panelURL, apiKey)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"php":    phpCode,
		"nodejs": nodeCode,
		"python": pythonCode,
		"curl":   curlCode,
	})
}

