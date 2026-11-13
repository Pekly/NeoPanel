package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type invitationSMTP struct {
	settings database.SMTPSettings
	messages chan string
	fail     atomic.Bool
}

func newInvitationSMTP(t *testing.T) *invitationSMTP {
	t.Helper()
	listener, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { listener.Close() })
	host, port, _ := net.SplitHostPort(listener.Addr().String())
	p, _ := strconv.Atoi(port)
	s := &invitationSMTP{settings: database.SMTPSettings{Enabled: true, Host: host, Port: p, FromEmail: "panel@example.test", FromName: "NeoPanel", Encryption: "none"}, messages: make(chan string, 30)}
	go func() {
		for {
			c, e := listener.Accept()
			if e != nil {
				return
			}
			go func() {
				defer c.Close()
				c.SetDeadline(time.Now().Add(10 * time.Second))
				reader := bufio.NewReader(c)
				fmt.Fprint(c, "220 localhost SMTP test\r\n")
				for {
					line, e := reader.ReadString('\n')
					if e != nil {
						return
					}
					switch {
					case strings.HasPrefix(line, "EHLO"), strings.HasPrefix(line, "HELO"):
						fmt.Fprint(c, "250 localhost\r\n")
					case strings.HasPrefix(line, "MAIL"), strings.HasPrefix(line, "RCPT"):
						if s.fail.Load() {
							fmt.Fprint(c, "550 delivery rejected\r\n")
						} else {
							fmt.Fprint(c, "250 OK\r\n")
						}
					case strings.HasPrefix(line, "DATA"):
						fmt.Fprint(c, "354 Send message\r\n")
						var message strings.Builder
						for {
							line, e = reader.ReadString('\n')
							if e != nil {
								return
							}
							if line == ".\r\n" {
								break
							}
							message.WriteString(line)
						}
						s.messages <- message.String()
						fmt.Fprint(c, "250 queued\r\n")
					case strings.HasPrefix(line, "QUIT"):
						fmt.Fprint(c, "221 bye\r\n")
						return
					default:
						fmt.Fprint(c, "250 OK\r\n")
					}
				}
			}()
		}
	}()
	return s
}
func (s *invitationSMTP) token(t *testing.T) string {
	t.Helper()
	select {
	case message := <-s.messages:
		match := regexp.MustCompile(`https?://[^\s]+/#invite-token=([a-f0-9]{64})`).FindStringSubmatch(message)
		if len(match) != 2 {
			t.Fatal("email missing fragment invitation link")
		}
		if strings.Contains(message, "evil.example") || strings.Contains(message, "worker.example") {
			t.Fatal("untrusted origin in invitation")
		}
		return match[1]
	case <-time.After(3 * time.Second):
		t.Fatal("no SMTP message received")
	}
	return ""
}

type invitationHarness struct {
	router  *Router
	handler *InvitationHandler
	db      *database.DB
	owner   *database.User
	session string
	mux     *http.ServeMux
}

func newInvitationHarness(t *testing.T, smtp *invitationSMTP) *invitationHarness {
	t.Helper()
	db, e := database.InitDB(filepath.Join(t.TempDir(), "panel.db"))
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	if e = db.MigrateInvitations(); e != nil {
		t.Fatal(e)
	}
	cfg := config.DefaultConfig()
	cfg.Paths.DataDir = t.TempDir()
	cfg.Paths.DBPath = filepath.Join(cfg.Paths.DataDir, "panel.db")
	cfg.FilePath = filepath.Join(cfg.Paths.DataDir, "config.yml")
	cfg.Security.NodeToken = "neo_node_test"
	cfg.Server.Role = "panel"
	auth := NewAuthHandler(db, cfg)
	router := &Router{cfg: cfg, auth: auth}
	router.email = NewEmailHandler(db, cfg, auth)
	router.nodes = NewNodeHandler(cfg, db, auth, nil)
	h := NewInvitationHandler(router)
	owner, e := db.CreateUserWithDetails("owner", "owner@example.test", "hash", "user", "")
	if e != nil {
		t.Fatal(e)
	}
	if e = db.CreateServer(&database.Server{ID: "server", Name: "SMTP Test Server", OwnerID: owner.ID}); e != nil {
		t.Fatal(e)
	}
	session, e := auth.GenerateToken(owner)
	if e != nil {
		t.Fatal(e)
	}
	if smtp != nil {
		if e = router.email.SaveSMTPSettings(smtp.settings); e != nil {
			t.Fatal(e)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/team/invitations", router.requireAuth(h.Create))
	mux.HandleFunc("GET /api/team/invitations", router.requireAuth(h.List))
	mux.HandleFunc("POST /api/team/invitations/{inviteId}/resend", router.requireAuth(h.Resend))
	mux.HandleFunc("DELETE /api/team/invitations/{inviteId}", router.requireAuth(h.Revoke))
	mux.HandleFunc("POST /api/auth/invitations/inspect", h.Inspect)
	mux.HandleFunc("POST /api/auth/invitations/accept", h.Accept)
	return &invitationHarness{router: router, handler: h, db: db, owner: owner, session: session, mux: mux}
}
func (h *invitationHarness) request(method, path string, body any, session string) *httptest.ResponseRecorder {
	payload, _ := json.Marshal(body)
	r := httptest.NewRequest(method, "http://panel.example"+path, bytes.NewReader(payload))
	if session != "" {
		r.Header.Set("Authorization", "Bearer "+session)
	}
	r.Header.Set("X-Forwarded-Host", "evil.example")
	r.Header.Set("X-Forwarded-Proto", "https")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	return w
}
func invitationWant(t *testing.T, w *httptest.ResponseRecorder, status int) {
	t.Helper()
	if w.Code != status {
		t.Fatalf("HTTP %d want %d: %s", w.Code, status, w.Body.String())
	}
}
func (h *invitationHarness) create(t *testing.T, s *invitationSMTP, name, node string) (*database.Invitation, string) {
	t.Helper()
	w := h.request("POST", "/api/team/invitations", map[string]any{"server_id": "server", "node_id": node, "username": name, "email": strings.ToUpper(name) + "@EXAMPLE.TEST", "permissions": []string{"console.view", "files.view"}}, h.session)
	invitationWant(t, w, 201)
	i := &database.Invitation{}
	if e := json.Unmarshal(w.Body.Bytes(), i); e != nil {
		t.Fatal(e)
	}
	token := s.token(t)
	if strings.Contains(w.Body.String(), token) || strings.Contains(w.Body.String(), "token_hash") {
		t.Fatal("secret in create response")
	}
	return i, token
}
func TestInvitationSMTPRegistrationAndSingleUse(t *testing.T) {
	s := newInvitationSMTP(t)
	h := newInvitationHarness(t, s)
	i, token := h.create(t, s, "newuser", "local")
	w := h.request("POST", "/api/auth/invitations/inspect", map[string]string{"token": token}, "")
	invitationWant(t, w, 200)
	if !strings.Contains(w.Body.String(), `"account_exists":false`) {
		t.Fatal(w.Body.String())
	}
	w = h.request("GET", "/api/team/invitations?server_id=server", nil, h.session)
	invitationWant(t, w, 200)
	if strings.Contains(w.Body.String(), token) || strings.Contains(w.Body.String(), "token_hash") {
		t.Fatal("list leaked token")
	}
	accept := map[string]string{"token": token, "username": "newuser", "password": "a-strong-password"}
	w = h.request("POST", "/api/auth/invitations/accept", accept, "")
	invitationWant(t, w, 200)
	if !strings.Contains(w.Body.String(), `"login_required":true`) {
		t.Fatal(w.Body.String())
	}
	u, e := h.db.GetUserByUsername(i.Username)
	if e != nil {
		t.Fatal(e)
	}
	if u.Role != "user" || u.Email != "newuser@example.test" || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(accept["password"])) != nil {
		t.Fatalf("bad registration: %+v", u)
	}
	ps, e := h.db.GetUserServerPermissions("server", u.ID)
	if e != nil || len(ps) != 2 {
		t.Fatalf("permissions %v %v", ps, e)
	}
	invitationWant(t, h.request("POST", "/api/auth/invitations/accept", accept, ""), 410)
	invitationWant(t, h.request("POST", "/api/auth/invitations/inspect", map[string]string{"token": token}, ""), 410)
}
func TestInvitationSMTPConcurrentAcceptance(t *testing.T) {
	s := newInvitationSMTP(t)
	h := newInvitationHarness(t, s)
	_, token := h.create(t, s, "concurrent", "local")
	var success atomic.Int32
	var wg sync.WaitGroup
	for n := 0; n < 5; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := h.request("POST", "/api/auth/invitations/accept", map[string]string{"token": token, "password": "a-strong-password"}, "")
			if w.Code == 200 {
				success.Add(1)
			} else if w.Code != 410 {
				t.Errorf("unexpected %d %s", w.Code, w.Body.String())
			}
		}()
	}
	wg.Wait()
	if success.Load() != 1 {
		t.Fatal(success.Load())
	}
}
func TestInvitationSMTPResendRevokeFailureValidation(t *testing.T) {
	s := newInvitationSMTP(t)
	h := newInvitationHarness(t, s)
	i, token := h.create(t, s, "resenduser", "local")
	invitationWant(t, h.request("POST", "/api/team/invitations/"+i.ID+"/resend", nil, h.session), 429)
	// Re-seed a past send timestamp through the public database API, without sleeping.
	raw, hash, e := invitationToken()
	if e != nil {
		t.Fatal(e)
	}
	past := time.Now().UTC().Add(-2 * time.Minute)
	seed := &database.Invitation{ID: "resend-test", ServerID: "server", NodeID: "local", Username: "rotateduser", Email: "rotateduser@example.test", Permissions: []string{}, InviterID: h.owner.ID, TokenHash: hash, Status: "pending", DeliveryStatus: "sent", CreatedAt: past, LastSentAt: past, ExpiresAt: past.Add(72 * time.Hour)}
	if e = h.db.CreateInvitation(seed); e != nil {
		t.Fatal(e)
	}
	invitationWant(t, h.request("POST", "/api/team/invitations/"+seed.ID+"/resend", nil, h.session), 200)
	rotated := s.token(t)
	if raw == rotated {
		t.Fatal("token not rotated")
	}
	invitationWant(t, h.request("POST", "/api/auth/invitations/inspect", map[string]string{"token": raw}, ""), 400)
	invitationWant(t, h.request("DELETE", "/api/team/invitations/"+i.ID, nil, h.session), 200)
	invitationWant(t, h.request("POST", "/api/auth/invitations/accept", map[string]string{"token": token, "password": "strong-password"}, ""), 410)
	s.fail.Store(true)
	w := h.request("POST", "/api/team/invitations", map[string]any{"server_id": "server", "username": "faileduser", "email": "failed@example.test", "permissions": []string{}}, h.session)
	invitationWant(t, w, 502)
	list, e := h.db.ListInvitations("server", "local")
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for _, i := range list {
		if i.Username == "faileduser" {
			found = i.DeliveryStatus == "failed" && i.Status == "pending"
		}
	}
	if !found {
		t.Fatal("delivery failure lost invite")
	}
	for _, body := range []map[string]any{{"server_id": "server", "username": "bad user", "email": "ok@example.test"}, {"server_id": "server", "username": "validuser", "email": "Name <mail@example.test>"}, {"server_id": "server", "username": "validuser", "email": "mail@example.test", "permissions": []string{"admin"}}} {
		invitationWant(t, h.request("POST", "/api/team/invitations", body, h.session), 400)
	}
	s.settings.Enabled = false
	h.router.email.SaveSMTPSettings(s.settings)
	invitationWant(t, h.request("POST", "/api/team/invitations", map[string]any{"server_id": "server", "username": "disabled", "email": "disabled@example.test"}, h.session), 503)
}
func TestInvitationExistingSessionAndStaleInviter(t *testing.T) {
	s := newInvitationSMTP(t)
	h := newInvitationHarness(t, s)
	existing, e := h.db.CreateUserWithDetails("existing", "existing@example.test", "unchanged", "user", "")
	if e != nil {
		t.Fatal(e)
	}
	_, token := h.create(t, s, "existing", "local")
	payload := map[string]string{"token": token, "password": "must-not-reset"}
	invitationWant(t, h.request("POST", "/api/auth/invitations/accept", payload, ""), 403)
	invitationWant(t, h.request("POST", "/api/auth/invitations/accept", payload, h.session), 403)
	pending, e := h.router.auth.Generate2FATempToken(existing)
	if e != nil {
		t.Fatal(e)
	}
	invitationWant(t, h.request("POST", "/api/auth/invitations/accept", payload, pending), 401)
	signed, e := h.router.auth.GenerateToken(existing)
	if e != nil {
		t.Fatal(e)
	}
	invitationWant(t, h.request("POST", "/api/auth/invitations/accept", payload, signed), 200)
	u, _ := h.db.GetUserByID(existing.ID)
	if u.PasswordHash != "unchanged" {
		t.Fatal("invitation reset password")
	}
	_, token = h.create(t, s, "staleuser", "local")
	h.db.UpdateServerOwner("server", 999)
	invitationWant(t, h.request("POST", "/api/auth/invitations/accept", map[string]string{"token": token, "password": "strong-password"}, ""), 403)
}

func invitationWorker(t *testing.T, worker *invitationHarness, failGrant *atomic.Bool, loseReply *atomic.Bool) *httptest.Server {
	t.Helper()
	r := worker.router
	r.cfg.Server.Role = "node"
	sub := NewSubuserHandler(worker.db)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/auth/me", r.requireAuth(func(w http.ResponseWriter, req *http.Request) { writeJSON(w, 200, getUserFromContext(req)) }))
	mux.HandleFunc("GET /api/servers/{id}/permissions", r.requireAuth(sub.GetMyPermissions))
	mux.HandleFunc("GET /api/servers/{id}", r.requireAuth(func(w http.ResponseWriter, req *http.Request) {
		srv, e := worker.db.GetServer(req.PathValue("id"))
		if e != nil {
			writeError(w, 404, "missing")
			return
		}
		writeJSON(w, 200, srv)
	}))
	mux.HandleFunc("POST /api/cluster/sync-auth", r.requireAdmin(r.nodes.SyncAuthHandler))
	mux.HandleFunc("POST /api/servers/{id}/subusers", r.requireAuth(func(w http.ResponseWriter, req *http.Request) {
		if failGrant.Load() {
			writeError(w, 503, "offline")
			return
		}
		if loseReply.Swap(false) {
			rr := httptest.NewRecorder()
			sub.AddSubuser(rr, req)
			writeError(w, 502, "reply lost after commit")
			return
		}
		sub.AddSubuser(w, req)
	}))
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}
func TestInvitationRemoteSMTPProvisionGrantRetryAndRevoke(t *testing.T) {
	s := newInvitationSMTP(t)
	master := newInvitationHarness(t, s)
	worker := newInvitationHarness(t, nil)
	var failGrant, loseReply atomic.Bool
	remote := invitationWorker(t, worker, &failGrant, &loseReply)
	if e := master.db.CreateNode(&database.ClusterNode{ID: "remote", Name: "Worker", BaseURL: remote.URL, APIToken: "neo_node_test", FQDN: "worker.example"}); e != nil {
		t.Fatal(e)
	}
	i, token := master.create(t, s, "remoteuser", "remote")
	failGrant.Store(true)
	payload := map[string]string{"token": token, "password": "first-strong-password"}
	invitationWant(t, master.request("POST", "/api/auth/invitations/accept", payload, ""), 502)
	saved, e := master.db.GetInvitation(i.ID)
	if e != nil || saved.Status != "granting" || saved.AccountID == 0 {
		t.Fatalf("not recoverable: %+v %v", saved, e)
	}
	inspection := master.request("POST", "/api/auth/invitations/inspect", map[string]string{"token": token}, "")
	invitationWant(t, inspection, 200)
	if !strings.Contains(inspection.Body.String(), `"account_reserved":true`) {
		t.Fatal("UI cannot distinguish retry from new registration: " + inspection.Body.String())
	}
	u, e := master.db.GetUserByUsername("remoteuser")
	if e != nil {
		t.Fatal(e)
	}
	originalHash := u.PasswordHash
	failGrant.Store(false)
	loseReply.Store(true)
	payload["password"] = "different-password"
	invitationWant(t, master.request("POST", "/api/auth/invitations/accept", payload, ""), 502)
	invitationWant(t, master.request("POST", "/api/auth/invitations/accept", payload, ""), 200)
	u, _ = master.db.GetUserByUsername("remoteuser")
	if u.PasswordHash != originalHash {
		t.Fatal("remote retry reset master password")
	}
	wu, e := worker.db.GetUserByID(u.ID)
	if e != nil || wu.Username != u.Username || wu.PasswordHash != originalHash {
		t.Fatalf("bad worker account %+v %v", wu, e)
	}
	ps, e := worker.db.GetUserServerPermissions("server", u.ID)
	if e != nil || len(ps) != 2 {
		t.Fatalf("worker grant %v %v", ps, e)
	}
	if ok, _ := master.db.IsServerSubuser("server", u.ID); ok {
		t.Fatal("remote grant incorrectly added local membership")
	}
	rows, e := worker.db.ListInvitations("server", "local")
	if e != nil || len(rows) != 0 {
		t.Fatal("invitation record leaked to worker")
	}
	i, token = master.create(t, s, "remoterevoke", "remote")
	loseReply.Store(true)
	invitationWant(t, master.request("POST", "/api/auth/invitations/accept", map[string]string{"token": token, "password": "strong-password"}, ""), 502)
	invitationWant(t, master.request("DELETE", "/api/team/invitations/"+i.ID, nil, master.session), 200)
	invitationWant(t, master.request("POST", "/api/auth/invitations/accept", map[string]string{"token": token, "password": "strong-password"}, ""), 410)
	revoked, _ := master.db.GetUserByUsername("remoterevoke")
	if ok, _ := worker.db.IsServerSubuser("server", revoked.ID); ok {
		t.Fatal("unresolved remote grant survived revocation")
	}
}
func TestInvitationRemoteAuthorizationAndOffline(t *testing.T) {
	s := newInvitationSMTP(t)
	master := newInvitationHarness(t, s)
	worker := newInvitationHarness(t, nil)
	var fail, lose atomic.Bool
	remote := invitationWorker(t, worker, &fail, &lose)
	master.db.CreateNode(&database.ClusterNode{ID: "remote", Name: "Worker", BaseURL: remote.URL, APIToken: "neo_node_test"})
	worker.db.UpdateServerOwner("server", 999)
	payload := map[string]any{"server_id": "server", "node_id": "remote", "username": "denieduser", "email": "denied@example.test"}
	invitationWant(t, master.request("POST", "/api/team/invitations", payload, master.session), 403)
	worker.db.UpdateServerOwner("server", worker.owner.ID)
	_, token := master.create(t, s, "offlineuser", "remote")
	remote.Close()
	invitationWant(t, master.request("POST", "/api/auth/invitations/accept", map[string]string{"token": token, "password": "strong-password"}, ""), 502)
	if u, _ := master.db.GetUserByUsername("offlineuser"); u != nil {
		t.Fatal("offline preflight created account")
	}
}
func TestInvitationSMTPHeadersAndOrigin(t *testing.T) {
	s := newInvitationSMTP(t)
	h := newInvitationHarness(t, s)
	if e := h.router.email.SendEmail(&s.settings, "to@example.test", "name\r\nBcc: evil@example.test", "subject", "", "body"); e == nil {
		t.Fatal("header injection accepted")
	}
	h.db.SetPanelSetting(panelSettingDaemonKey, `{"fqdn":"https://configured.example:8443"}`)
	r := httptest.NewRequest("POST", "http://panel.example/", nil)
	r.Header.Set("X-Forwarded-Host", "evil.example")
	origin, e := h.handler.origin(r)
	if e != nil || origin != "https://configured.example:8443" {
		t.Fatalf("origin %q %v", origin, e)
	}
	h.db.SetPanelSetting(panelSettingDaemonKey, `{"fqdn":"https://evil@example.test"}`)
	if _, e = h.handler.origin(r); e == nil {
		t.Fatal("userinfo origin accepted")
	}
	// Test receiver accepts real SMTP DATA; no external SMTP server is required.
	if e = h.router.email.SendEmail(&s.settings, "to@example.test", "Quoted, Name", "subject", "", "body"); e != nil {
		t.Fatal(e)
	}
	select {
	case message := <-s.messages:
		if _, e = io.ReadAll(strings.NewReader(message)); e != nil {
			t.Fatal(e)
		}
	case <-time.After(time.Second):
		t.Fatal("missing SMTP DATA")
	}
}
