package api

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

type InvitationHandler struct {
	router *Router
	db     *database.DB
	auth   *AuthHandler
	email  *EmailHandler
	nodes  *NodeHandler
}

func NewInvitationHandler(r *Router) *InvitationHandler {
	return &InvitationHandler{router: r, db: r.auth.db, auth: r.auth, email: r.email, nodes: r.nodes}
}

var invitationUsername = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{2,31}$`)
var invitationID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

func validateInvitationPermissions(ps []string) error {
	allowed := map[string]bool{"console.view": true, "console.command": true, "server.power": true, "files.view": true, "files.edit": true, "files.delete": true, "settings.view": true, "settings.edit": true, "settings.ram": true, "database.manage": true, "backup.manage": true, "subusers.manage": true, "schedules.manage": true}
	allowed["sftp.access"] = true
	seen := map[string]bool{}
	if len(ps) > len(allowed) {
		return fmt.Errorf("invalid permissions")
	}
	for _, p := range ps {
		if !allowed[p] || seen[p] {
			return fmt.Errorf("invalid or duplicate permission: %s", p)
		}
		seen[p] = true
	}
	return nil
}
func invitationToken() (string, string, error) {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		return "", "", e
	}
	raw := hex.EncodeToString(b)
	return raw, invitationHash(raw), nil
}
func invitationHash(raw string) string {
	s := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(s[:])
}
func invitationDecode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 16384)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if d.Decode(v) != nil {
		writeError(w, 400, "Invalid invitation request")
		return false
	}
	if d.Decode(&struct{}{}) != io.EOF {
		writeError(w, 400, "Invalid invitation request")
		return false
	}
	return true
}
func invitationError(w http.ResponseWriter, err error) {
	code := 500
	message := "Invitation operation failed"
	switch {
	case errors.Is(err, database.ErrInvitationGone):
		code = 410
		message = err.Error()
	case errors.Is(err, sql.ErrNoRows):
		code = 400
		message = "Invalid invitation"
	case errors.Is(err, database.ErrInvitationIdentity):
		code = 403
		message = err.Error()
	case errors.Is(err, database.ErrInvitationForbidden):
		code = 403
		message = err.Error()
	case errors.Is(err, database.ErrInvitationConflict), errors.Is(err, database.ErrInvitationBusy):
		code = 409
		message = err.Error()
	case errors.Is(err, database.ErrInvitationCooldown):
		code = 429
		message = err.Error()
	default:
		var remote *invitationRemoteError
		if errors.As(err, &remote) {
			code = remote.code
			message = remote.message
		}
	}
	writeError(w, code, message)
}

type invitationRemoteError struct {
	code    int
	message string
}

func (e *invitationRemoteError) Error() string { return e.message }
func (h *InvitationHandler) master(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Cache-Control", "no-store")
	if h.router.cfg.Server.Role == "node" || r.Header.Get("X-NeoPanel-User-ID") != "" {
		writeError(w, 403, "Invitations must be managed on the master panel")
		return false
	}
	return true
}
func (h *InvitationHandler) caller(r *http.Request) (*JWTClaims, error) {
	c, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if c == nil || c.UserID <= 0 {
		return nil, database.ErrInvitationForbidden
	}
	u, e := h.db.GetUserByID(c.UserID)
	if e != nil {
		return nil, database.ErrInvitationForbidden
	}
	return &JWTClaims{UserID: u.ID, Username: u.Username, Role: u.Role}, nil
}
func (h *InvitationHandler) node(id string) (*database.ClusterNode, error) {
	if id == "local" {
		return nil, nil
	}
	return h.db.GetNode(id)
}
func (h *InvitationHandler) remote(r *http.Request, node *database.ClusterNode, c *JWTClaims, method, path string, body any, provision bool) (*http.Response, error) {
	var data []byte
	var err error
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequestWithContext(r.Context(), method, strings.TrimRight(node.BaseURL, "/")+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	setProxyIdentity(req.Header, c, node.APIToken)
	// Preserve even administrator identity rather than relying on the node token role.
	req.Header.Set("X-NeoPanel-User-ID", strconv.Itoa(c.UserID))
	req.Header.Set("X-NeoPanel-User-Role", c.Role)
	if provision {
		req.Header.Del("X-NeoPanel-User-ID")
		req.Header.Del("X-NeoPanel-User-Role")
		req.Header.Set("X-NeoPanel-Invitation-Provision", "1")
	}
	req.Header.Set("Content-Type", "application/json")
	client := *h.nodes.httpClient
	client.Timeout = 15 * time.Second
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if err != nil {
		return nil, &invitationRemoteError{502, "Remote node unavailable; invitation retained, retry safely"}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		code := 502
		if resp.StatusCode == 409 {
			return nil, &invitationRemoteError{409, "Remote invitation conflicts with an account or existing server access; check the exact username/email and account synchronization, or manage existing permissions"}
		}
		if resp.StatusCode == 403 || resp.StatusCode == 404 || resp.StatusCode == 409 || resp.StatusCode == 410 {
			code = resp.StatusCode
		}
		return nil, &invitationRemoteError{code, "Remote node rejected invitation operation; check server, permissions and account synchronization"}
	}
	return resp, nil
}
func (h *InvitationHandler) authorize(r *http.Request, server string, node *database.ClusterNode, c *JWTClaims, grants ...[]string) (string, error) {
	if !invitationID.MatchString(server) {
		return "", &invitationRemoteError{400, "Invalid server ID"}
	}
	if node == nil {
		if code := authorizeServer(h.db, c, server, "subusers.manage"); code != 0 {
			return "", &invitationRemoteError{code, "Server missing or subusers.manage permission required"}
		}
		for _, requested := range grants {
			if !canGrantServerPermissions(h.db, c, server, requested) {
				return "", database.ErrInvitationForbidden
			}
		}
		s, e := h.db.GetServer(server)
		if e != nil {
			return "", e
		}
		return s.Name, nil
	}
	if err := h.router.checkProxyIdentity(r, node, c); err != nil {
		return "", &invitationRemoteError{502, "Node identity verification failed; update and synchronize the node"}
	}
	resp, err := h.remote(r, node, c, "GET", "/api/servers/"+server+"/permissions", nil, false)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.Header.Get("X-NeoPanel-Invitations") != "1" || resp.Header.Get("X-NeoPanel-Invitation-Check") != "1" {
		return "", &invitationRemoteError{502, "Update the remote node before using invitations"}
	}
	var p struct {
		Owner       bool     `json:"is_owner"`
		Admin       bool     `json:"is_admin"`
		Permissions []string `json:"permissions"`
	}
	if json.NewDecoder(io.LimitReader(resp.Body, 65536)).Decode(&p) != nil {
		return "", &invitationRemoteError{502, "Invalid node permission response"}
	}
	allowed := p.Owner || p.Admin
	for _, permission := range p.Permissions {
		allowed = allowed || permission == "subusers.manage"
	}
	if !allowed {
		return "", database.ErrInvitationForbidden
	}
	if !p.Owner && !p.Admin {
		for _, requested := range grants {
			if !permissionsContainAll(p.Permissions, requested) {
				return "", database.ErrInvitationForbidden
			}
		}
	}
	// Fetch the name only after authorization; never use the node hostname for email links.
	sr, err := h.remote(r, node, c, "GET", "/api/servers/"+server, nil, false)
	if err != nil {
		return "", err
	}
	defer sr.Body.Close()
	var result struct {
		Name   string `json:"name"`
		Server struct {
			Name string `json:"name"`
		} `json:"server"`
	}
	if json.NewDecoder(io.LimitReader(sr.Body, 65536)).Decode(&result) != nil {
		return "", &invitationRemoteError{502, "Invalid node server response"}
	}
	if result.Name == "" {
		result.Name = result.Server.Name
	}
	if result.Name == "" {
		result.Name = server
	}
	return result.Name, nil
}
func (h *InvitationHandler) origin(r *http.Request) (string, error) {
	return h.auth.panelLinkOrigin(r)
}
func (h *InvitationHandler) smtp() (*database.SMTPSettings, error) {
	s, e := h.email.GetSMTPSettings()
	if e != nil || s == nil || !s.Enabled || s.Host == "" {
		return nil, &invitationRemoteError{503, "SMTP is not configured or disabled"}
	}
	return s, nil
}
func (h *InvitationHandler) deliver(w http.ResponseWriter, i *database.Invitation, raw, origin, serverName string, s *database.SMTPSettings, code int) {
	bodyHTML, bodyPlain, err := invitationEmail(LoadBrandingConfig(h.db), i, serverName, origin+"/#invite-token="+raw)
	if err == nil {
		err = h.email.SendEmail(s, i.Email, i.Username, "Team invitation", bodyHTML, bodyPlain)
	}
	status := "sent"
	if err != nil {
		status = "failed"
	}
	i.DeliveryStatus = status
	if e := h.db.InvitationDelivery(i.ID, i.TokenHash, status); e != nil {
		writeError(w, 500, "Could not persist delivery result; invitation retained")
		return
	}
	if err != nil {
		writeJSON(w, 502, map[string]any{"error": "SMTP delivery failed; invitation retained. Correct SMTP settings and resend.", "invitation": i})
		return
	}
	writeJSON(w, code, i)
}
func (h *InvitationHandler) List(w http.ResponseWriter, r *http.Request) {
	if !h.master(w, r) {
		return
	}
	server, nodeID := r.URL.Query().Get("server_id"), r.URL.Query().Get("node_id")
	if nodeID == "" {
		nodeID = "local"
	}
	c, e := h.caller(r)
	if e != nil {
		invitationError(w, e)
		return
	}
	n, e := h.node(nodeID)
	if e != nil {
		invitationError(w, e)
		return
	}
	if _, e = h.authorize(r, server, n, c); e != nil {
		invitationError(w, e)
		return
	}
	list, e := h.db.ListInvitations(server, nodeID)
	if e != nil {
		invitationError(w, e)
		return
	}
	_, e = h.smtp()
	writeJSON(w, 200, map[string]any{"invitations": list, "smtp_configured": e == nil})
}
func (h *InvitationHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !h.master(w, r) {
		return
	}
	var req struct {
		ServerID    string   `json:"server_id"`
		NodeID      string   `json:"node_id"`
		Username    string   `json:"username"`
		Email       string   `json:"email"`
		Permissions []string `json:"permissions"`
	}
	if !invitationDecode(w, r, &req) {
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.NodeID == "" {
		req.NodeID = "local"
	}
	address, e := mail.ParseAddress(req.Email)
	if !invitationUsername.MatchString(req.Username) || e != nil || address.Address != req.Email || len(req.Email) > 254 || strings.ContainsAny(req.Email, "\r\n") {
		writeError(w, 400, "Valid username (3–32 letters, numbers, _, . or -) and email required")
		return
	}
	if e = validateInvitationPermissions(req.Permissions); e != nil {
		writeError(w, 400, e.Error())
		return
	}
	if req.Permissions == nil {
		req.Permissions = []string{}
	}
	c, e := h.caller(r)
	if e != nil {
		invitationError(w, e)
		return
	}
	n, e := h.node(req.NodeID)
	if e != nil {
		invitationError(w, e)
		return
	}
	serverName, e := h.authorize(r, req.ServerID, n, c, req.Permissions)
	if e != nil {
		invitationError(w, e)
		return
	}
	s, e := h.smtp()
	if e != nil {
		invitationError(w, e)
		return
	}
	origin, e := h.origin(r)
	if e != nil {
		writeError(w, 400, "Configure a valid master panel FQDN")
		return
	}
	raw, hash, e := invitationToken()
	if e != nil {
		invitationError(w, e)
		return
	}
	id, _, e := invitationToken()
	if e != nil {
		invitationError(w, e)
		return
	}
	now := time.Now().UTC()
	i := &database.Invitation{ID: id, ServerID: req.ServerID, NodeID: req.NodeID, Username: req.Username, Email: req.Email, Permissions: req.Permissions, InviterID: c.UserID, TokenHash: hash, Status: "pending", DeliveryStatus: "sending", CreatedAt: now, LastSentAt: now, ExpiresAt: now.Add(72 * time.Hour)}
	if e = h.checkTarget(r, n, c, i); e != nil {
		invitationError(w, e)
		return
	}
	if e = h.db.CreateInvitation(i); e != nil {
		invitationError(w, e)
		return
	}
	h.deliver(w, i, raw, origin, serverName, s, 201)
}
func (h *InvitationHandler) managed(r *http.Request) (*database.Invitation, *database.ClusterNode, *JWTClaims, error) {
	i, e := h.db.GetInvitation(r.PathValue("inviteId"))
	if e != nil {
		return nil, nil, nil, e
	}
	c, e := h.caller(r)
	if e != nil {
		return nil, nil, nil, e
	}
	n, e := h.node(i.NodeID)
	if e != nil {
		return nil, nil, nil, e
	}
	_, e = h.authorize(r, i.ServerID, n, c)
	return i, n, c, e
}
func (h *InvitationHandler) Resend(w http.ResponseWriter, r *http.Request) {
	if !h.master(w, r) {
		return
	}
	i, n, c, e := h.managed(r)
	if e != nil {
		invitationError(w, e)
		return
	}
	// Do not rotate tokens or send mail for an account already granted access.
	if i.Status == "granting" {
		invitationError(w, database.ErrInvitationBusy)
		return
	}
	if i.Status != "pending" && i.Status != "expired" {
		invitationError(w, database.ErrInvitationGone)
		return
	}
	if e = h.checkTarget(r, n, c, i); e != nil {
		invitationError(w, e)
		return
	}
	serverName, e := h.authorize(r, i.ServerID, n, c, i.Permissions)
	if e != nil {
		invitationError(w, e)
		return
	}
	s, e := h.smtp()
	if e != nil {
		invitationError(w, e)
		return
	}
	origin, e := h.origin(r)
	if e != nil {
		writeError(w, 400, "Configure a valid master panel FQDN")
		return
	}
	raw, hash, e := invitationToken()
	if e != nil {
		invitationError(w, e)
		return
	}
	i, e = h.db.RotateInvitation(i.ID, hash, time.Now().UTC())
	if e != nil {
		invitationError(w, e)
		return
	}
	h.deliver(w, i, raw, origin, serverName, s, 200)
}
func (h *InvitationHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	if !h.master(w, r) {
		return
	}
	i, n, c, e := h.managed(r)
	if e != nil {
		invitationError(w, e)
		return
	}
	var remote func(*database.Invitation) error
	if n != nil {
		remote = func(current *database.Invitation) error { return h.remoteGrant(r, n, c, current, nil, true) }
	}
	if e = h.db.RevokeInvitation(i.ID, remote); e != nil {
		invitationError(w, e)
		return
	}
	writeJSON(w, 200, map[string]string{"message": "Invitation revoked"})
}
func (h *InvitationHandler) token(raw string) (*database.Invitation, error) {
	b, e := hex.DecodeString(raw)
	if e != nil || len(b) != 32 {
		return nil, sql.ErrNoRows
	}
	i, e := h.db.InvitationByToken(invitationHash(raw))
	if e != nil {
		return nil, e
	}
	if (i.Status != "pending" && i.Status != "granting") || !i.ExpiresAt.After(time.Now()) {
		return nil, database.ErrInvitationGone
	}
	return i, nil
}
func (h *InvitationHandler) inviter(r *http.Request, i *database.Invitation) (*database.ClusterNode, string, error) {
	u, e := h.db.GetUserByID(i.InviterID)
	if e != nil {
		return nil, "", database.ErrInvitationForbidden
	}
	n, e := h.node(i.NodeID)
	if e != nil {
		return nil, "", e
	}
	name, e := h.authorize(r, i.ServerID, n, &JWTClaims{UserID: u.ID, Username: u.Username, Role: u.Role}, i.Permissions)
	return n, name, e
}
func (h *InvitationHandler) Inspect(w http.ResponseWriter, r *http.Request) {
	if !h.master(w, r) {
		return
	}
	var req struct {
		Token string `json:"token"`
	}
	if !invitationDecode(w, r, &req) {
		return
	}
	i, e := h.token(req.Token)
	if e != nil {
		invitationError(w, e)
		return
	}
	_, name, e := h.inviter(r, i)
	if e != nil {
		invitationError(w, e)
		return
	}
	u, e := h.db.GetUserByEmailOrUsername(i.Email)
	if e != nil && !errors.Is(e, sql.ErrNoRows) {
		invitationError(w, e)
		return
	}
	writeJSON(w, 200, map[string]any{"username": i.Username, "email": i.Email, "server_name": name, "server_id": i.ServerID, "node_id": i.NodeID, "permissions": i.Permissions, "expires_at": i.ExpiresAt, "account_exists": u != nil && u.ID != i.AccountID, "account_reserved": u != nil && i.AccountID > 0 && u.ID == i.AccountID})
}
func (h *InvitationHandler) remoteGrant(r *http.Request, n *database.ClusterNode, c *JWTClaims, i *database.Invitation, u *database.User, revoke bool) error {
	action := "grant"
	username := ""
	if revoke {
		action = "revoke"
	} else {
		username = u.Username
	}
	body := map[string]any{"username": username, "permissions": i.Permissions, "invitation_id": i.ID, "invitation_action": action, "invitation_expires_at": i.ExpiresAt}
	resp, e := h.remote(r, n, c, "POST", "/api/servers/"+i.ServerID+"/subusers", body, false)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	if resp.Header.Get("X-NeoPanel-Invitations") != "1" {
		return &invitationRemoteError{502, "Node did not acknowledge durable invitation grant"}
	}
	return nil
}
func (h *InvitationHandler) Accept(w http.ResponseWriter, r *http.Request) {
	if !h.master(w, r) {
		return
	}
	var req struct {
		Token    string `json:"token"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if !invitationDecode(w, r, &req) {
		return
	}
	i, e := h.token(req.Token)
	if e != nil {
		invitationError(w, e)
		return
	}
	if req.Username != "" && req.Username != i.Username {
		writeError(w, 400, "Use the username requested in the invitation")
		return
	}
	n, _, e := h.inviter(r, i)
	if e != nil {
		invitationError(w, e)
		return
	}
	sessionID := 0
	if header := r.Header.Get("Authorization"); strings.HasPrefix(header, "Bearer ") {
		c, err := h.auth.ValidateToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			writeError(w, 401, "Valid signed-in session required")
			return
		}
		sessionID = c.UserID
	}
	existing, e := h.db.GetUserByEmailOrUsername(i.Email)
	if e != nil && !errors.Is(e, sql.ErrNoRows) {
		invitationError(w, e)
		return
	}
	passwordHash := ""
	// Identity checks belong in the transaction, after rechecking single use.
	if existing == nil {
		if len(req.Password) < 10 || len(req.Password) > 72 {
			writeError(w, 400, "Password must contain 10–72 bytes")
			return
		}
		passwordHash, e = h.auth.HashPassword(req.Password)
		if e != nil {
			invitationError(w, e)
			return
		}
	}
	var remote func(*database.Invitation, *database.User, *database.User) error
	if n != nil {
		if e = h.db.ReserveInvitationAccount(i.ID, i.TokenHash, sessionID, passwordHash); e != nil {
			invitationError(w, e)
			return
		}
		remote = func(current *database.Invitation, u, inviter *database.User) error {
			c := &JWTClaims{UserID: inviter.ID, Username: inviter.Username, Role: inviter.Role}
			// Single-user, non-overwriting mode of the existing cluster auth sync interface.
			payload := SyncAuthRequest{Users: []SyncedUser{{ID: u.ID, Username: u.Username, Email: u.Email, PasswordHash: u.PasswordHash, Role: u.Role}}}
			resp, err := h.remote(r, n, c, "POST", "/api/cluster/sync-auth", payload, true)
			if err != nil {
				return err
			}
			resp.Body.Close()
			if resp.Header.Get("X-NeoPanel-Invitation-Provision") != "1" {
				return &invitationRemoteError{502, "Node lacks safe invitation provisioning"}
			}
			return h.remoteGrant(r, n, c, current, u, false)
		}
	}
	u, e := h.db.AcceptInvitation(i.ID, i.TokenHash, sessionID, passwordHash, remote)
	if e != nil {
		invitationError(w, e)
		return
	}
	writeJSON(w, 200, map[string]any{"message": "Invitation accepted. Sign in normally to continue.", "server_id": i.ServerID, "node_id": i.NodeID, "username": u.Username, "email": u.Email, "login_required": true})
}

// Preflight is intentionally read-only; acceptance checks again transactionally.
// Never run it ahead of remote receipt recovery during acceptance.
func (h *InvitationHandler) checkTarget(r *http.Request, n *database.ClusterNode, c *JWTClaims, i *database.Invitation) error {
	if err := h.db.CheckInvitationTarget(i); err != nil {
		return err
	}
	if n == nil {
		return nil
	}
	uid := 0
	u, err := h.db.GetUserByUsername(i.Username)
	if err == nil {
		uid = u.ID
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	body := map[string]any{"invitation_id": i.ID, "invitation_action": "check", "username": i.Username, "user_email": i.Email, "invitation_user_id": uid}
	resp, err := h.remote(r, n, c, "POST", "/api/servers/"+i.ServerID+"/subusers", body, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.Header.Get("X-NeoPanel-Invitation-Check") != "1" {
		return &invitationRemoteError{502, "Update the remote node before using invitations: safe target checks required"}
	}
	return nil
}
