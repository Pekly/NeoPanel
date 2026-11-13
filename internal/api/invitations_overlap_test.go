package api

import (
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func invitationPair(t *testing.T, remote bool) (*invitationHarness, *invitationHarness, *invitationSMTP, string) {
	t.Helper()
	s := newInvitationSMTP(t)
	master := newInvitationHarness(t, s)
	if !remote {
		return master, master, s, "local"
	}
	worker := newInvitationHarness(t, nil)
	var fail, lose atomic.Bool
	server := invitationWorker(t, worker, &fail, &lose)
	if err := master.db.CreateNode(&database.ClusterNode{ID: "remote", Name: "Worker", BaseURL: server.URL, APIToken: "neo_node_test"}); err != nil {
		t.Fatal(err)
	}
	return master, worker, s, "remote"
}

func invitationNoMail(t *testing.T, s *invitationSMTP) {
	t.Helper()
	select {
	case <-s.messages:
		t.Fatal("conflict sent an extra email")
	default:
	}
}

func invitationExisting(t *testing.T, master, worker *invitationHarness, name, role string) *database.User {
	t.Helper()
	u, err := master.db.CreateUserWithDetails(name, name+"@example.test", "original-password", role, "")
	if err != nil {
		t.Fatal(err)
	}
	if master != worker {
		if err = worker.db.ProvisionInvitationUser(u); err != nil {
			t.Fatal(err)
		}
	}
	return u
}

func TestInvitationOverlapCreate(t *testing.T) {
	for _, remote := range []bool{false, true} {
		for _, kind := range []string{"member", "owner", "admin", "username-email", "email-username", "case"} {
			t.Run(fmtInvitationCase(remote, kind), func(t *testing.T) {
				master, worker, smtp, node := invitationPair(t, remote)
				role := "user"
				if kind == "admin" {
					role = "admin"
				}
				u := invitationExisting(t, master, worker, "existing", role)
				name, email := u.Username, u.Email
				switch kind {
				case "member":
					if _, err := worker.db.AddServerSubuser("server", u.ID, []string{"files.edit"}); err != nil {
						t.Fatal(err)
					}
				case "owner":
					name, email = worker.owner.Username, worker.owner.Email
				case "username-email":
					email = "different@example.test"
				case "email-username":
					name = "different"
				case "case":
					name = "Existing"
				}
				w := master.request("POST", "/api/team/invitations", map[string]any{"server_id": "server", "node_id": node, "username": name, "email": email}, master.session)
				invitationWant(t, w, 409)
				if !strings.Contains(w.Body.String(), "manage") && !strings.Contains(w.Body.String(), "exact") {
					t.Fatal("conflict lacks actionable guidance")
				}
				invitationNoMail(t, smtp)
				list, err := master.db.ListInvitations("server", node)
				if err != nil || len(list) != 0 {
					t.Fatalf("conflict persisted invitation: %v %v", list, err)
				}
				if kind == "member" {
					ps, err := worker.db.GetUserServerPermissions("server", u.ID)
					if err != nil || strings.Join(ps, ",") != "files.edit" {
						t.Fatalf("permissions overwritten: %v %v", ps, err)
					}
				}
			})
		}
	}
}

func fmtInvitationCase(remote bool, name string) string {
	if remote {
		return "remote/" + name
	}
	return "local/" + name
}

func TestInvitationRemoteOnlyAccountCollision(t *testing.T) {
	for _, kind := range []string{"username", "email", "numeric-id", "member", "admin"} {
		t.Run(kind, func(t *testing.T) {
			master, worker, smtp, node := invitationPair(t, true)
			name, email, role := "remoteonly", "different@example.test", "user"
			if kind == "email" {
				name, email = "different", "remoteonly@example.test"
			}
			if kind == "member" || kind == "admin" {
				email = "remoteonly@example.test"
			}
			if kind == "admin" {
				role = "admin"
			}
			if kind == "numeric-id" {
				if _, err := master.db.CreateUserWithDetails("remoteonly", "remoteonly@example.test", "hash", "user", ""); err != nil {
					t.Fatal(err)
				}
				name = "different"
			}
			u, err := worker.db.CreateUserWithDetails(name, email, "hash", role, "")
			if err != nil {
				t.Fatal(err)
			}
			if kind == "member" {
				if _, err = worker.db.AddServerSubuser("server", u.ID, []string{"files.edit"}); err != nil {
					t.Fatal(err)
				}
			}
			invitationWant(t, master.request("POST", "/api/team/invitations", map[string]any{"server_id": "server", "node_id": node, "username": "remoteonly", "email": "remoteonly@example.test"}, master.session), 409)
			invitationNoMail(t, smtp)
		})
	}
}

func TestInvitationExistingAccountOtherServerAndDuplicate(t *testing.T) {
	for _, remote := range []bool{false, true} {
		t.Run(fmtInvitationCase(remote, "other-server"), func(t *testing.T) {
			master, worker, smtp, node := invitationPair(t, remote)
			u := invitationExisting(t, master, worker, "existing", "user")
			if err := worker.db.CreateServer(&database.Server{ID: "other", Name: "Other", OwnerID: worker.owner.ID}); err != nil {
				t.Fatal(err)
			}
			if _, err := worker.db.AddServerSubuser("other", u.ID, []string{"files.edit"}); err != nil {
				t.Fatal(err)
			}
			i, token := master.create(t, smtp, "existing", node)
			invitationWant(t, master.request("POST", "/api/team/invitations", map[string]any{"server_id": "server", "node_id": node, "username": u.Username, "email": u.Email}, master.session), 409)
			invitationWant(t, master.request("POST", "/api/team/invitations/"+i.ID+"/resend", nil, master.session), 429)
			invitationNoMail(t, smtp)
			session, err := master.router.auth.GenerateToken(u)
			if err != nil {
				t.Fatal(err)
			}
			payload := map[string]string{"token": token, "password": "not-a-password-reset"}
			invitationWant(t, master.request("POST", "/api/auth/invitations/accept", payload, ""), 403)
			var success atomic.Int32
			var wg sync.WaitGroup
			for n := 0; n < 4; n++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					w := master.request("POST", "/api/auth/invitations/accept", payload, session)
					if w.Code == 200 {
						success.Add(1)
					} else if w.Code != 410 {
						t.Errorf("concurrent accept: %d %s", w.Code, w.Body.String())
					}
				}()
			}
			wg.Wait()
			if success.Load() != 1 {
				t.Fatalf("accepted %d times", success.Load())
			}
			ps, err := worker.db.GetUserServerPermissions("other", u.ID)
			if err != nil || strings.Join(ps, ",") != "files.edit" {
				t.Fatalf("other server changed: %v %v", ps, err)
			}
			subs, err := worker.db.ListServerSubusers("server")
			if err != nil || len(subs) != 1 {
				t.Fatalf("memberships %v %v", subs, err)
			}
			if err = worker.db.UpdateServerSubuser("server", subs[0].ID, []string{"files.delete"}); err != nil {
				t.Fatal(err)
			}
			invitationWant(t, master.request("POST", "/api/auth/invitations/accept", payload, session), 410)
			invitationWant(t, master.request("POST", "/api/team/invitations/"+i.ID+"/resend", nil, master.session), 410)
			ps, err = worker.db.GetUserServerPermissions("server", u.ID)
			if err != nil || strings.Join(ps, ",") != "files.delete" {
				t.Fatalf("retry reset permissions: %v %v", ps, err)
			}
			stored, _ := master.db.GetUserByID(u.ID)
			if stored.PasswordHash != "original-password" {
				t.Fatal("password reset")
			}
			invitationNoMail(t, smtp)
		})
	}
}

func TestInvitationMemberAddedWhilePending(t *testing.T) {
	for _, remote := range []bool{false, true} {
		t.Run(fmtInvitationCase(remote, "pending"), func(t *testing.T) {
			master, worker, smtp, node := invitationPair(t, remote)
			u := invitationExisting(t, master, worker, "pendinguser", "user")
			i, token := master.create(t, smtp, u.Username, node)
			if _, err := worker.db.AddServerSubuser("server", u.ID, []string{"files.edit"}); err != nil {
				t.Fatal(err)
			}
			invitationWant(t, master.request("POST", "/api/team/invitations/"+i.ID+"/resend", nil, master.session), 409)
			stored, err := master.db.GetInvitation(i.ID)
			if err != nil || stored.TokenHash != invitationHash(token) {
				t.Fatal("conflicting resend rotated token")
			}
			session, err := master.router.auth.GenerateToken(u)
			if err != nil {
				t.Fatal(err)
			}
			for n := 0; n < 2; n++ {
				invitationWant(t, master.request("POST", "/api/auth/invitations/accept", map[string]string{"token": token}, session), 409)
			}
			if err = master.db.RevokeInvitation(i.ID, func(current *database.Invitation) error {
				return worker.db.InvitationMembership(current.ID, "server", 0, worker.owner.ID, "user", current.Permissions, current.ExpiresAt, true)
			}); err != nil {
				t.Fatal(err)
			}
			ps, err := worker.db.GetUserServerPermissions("server", u.ID)
			if err != nil || strings.Join(ps, ",") != "files.edit" {
				t.Fatalf("manual membership overwritten or revoked: %v %v", ps, err)
			}
			invitationNoMail(t, smtp)
		})
	}
}

func TestInvitationStyledEmailMIME(t *testing.T) {
	s := newInvitationSMTP(t)
	h := newInvitationHarness(t, s)
	brand := `Acme <b>Cloud</b> & Co`
	server := `Game <img src=x onerror=alert(1)> & friends`
	cfg, _ := json.Marshal(map[string]string{"brand_name": brand, "accent_color": "#123456"})
	if err := h.db.SetPanelSetting(panelSettingBrandingKey, string(cfg)); err != nil {
		t.Fatal(err)
	}
	if err := h.db.CreateServer(&database.Server{ID: "html", Name: server, OwnerID: h.owner.ID}); err != nil {
		t.Fatal(err)
	}
	if err := h.db.SetPanelSetting(panelSettingDaemonKey, `{"fqdn":"https://configured.example:8443"}`); err != nil {
		t.Fatal(err)
	}
	w := h.request("POST", "/api/team/invitations", map[string]any{"server_id": "html", "username": "htmluser", "email": "htmluser@example.test", "permissions": []string{"console.view", "files.view"}}, h.session)
	invitationWant(t, w, 201)
	var i database.Invitation
	if err := json.Unmarshal(w.Body.Bytes(), &i); err != nil {
		t.Fatal(err)
	}
	var raw string
	select {
	case raw = <-s.messages:
	case <-time.After(time.Second):
		t.Fatal("missing email")
	}
	message, err := mail.ReadMessage(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	media, params, err := mime.ParseMediaType(message.Header.Get("Content-Type"))
	if err != nil || media != "multipart/alternative" {
		t.Fatalf("not multipart alternative: %s %v", media, err)
	}
	reader := multipart.NewReader(message.Body, params["boundary"])
	parts := map[string]string{}
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		kind, _, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		parts[kind] = string(body)
	}
	plain, html := parts["text/plain"], parts["text/html"]
	match := regexp.MustCompile(`https://configured\.example:8443/#invite-token=([a-f0-9]{64})`).FindStringSubmatch(plain)
	if len(parts) != 2 || len(match) != 2 {
		t.Fatal("missing plain/html or configured-origin link")
	}
	link, token := match[0], match[1]
	if strings.Count(html, `href="`+link+`"`) != 2 || !strings.Contains(html, ">Accept invitation</a>") {
		t.Fatal("CTA/fallback does not match plain single-use link")
	}
	for _, text := range []string{"Acme &lt;b&gt;Cloud&lt;/b&gt; &amp; Co", "Game &lt;img src=x onerror=alert(1)&gt; &amp; friends", "#123456", "htmluser", "console.view, files.view", i.ExpiresAt.UTC().Format("02 Jan 2006 at 15:04 UTC"), "max-width:600px", "Sign in", "Register", "ignore this email"} {
		if !strings.Contains(html, text) {
			t.Errorf("HTML missing %q", text)
		}
	}
	if strings.Contains(html, "<img") || strings.Contains(html, "<b>Cloud") || strings.Contains(html, "ZgotmplZ") {
		t.Fatal("unescaped or unsafe HTML")
	}
	if !strings.Contains(plain, brand) || !strings.Contains(plain, server) || !strings.Contains(plain, "UTC") {
		t.Fatal("plain alternative incomplete")
	}
	if strings.Contains(w.Body.String(), token) {
		t.Fatal("API leaked email secret")
	}
	invitationWant(t, h.request("POST", "/api/auth/invitations/accept", map[string]string{"token": token, "password": "a-strong-password"}, ""), 200)
	invitationWant(t, h.request("POST", "/api/auth/invitations/accept", map[string]string{"token": token, "password": "a-strong-password"}, ""), 410)
	// Defense in depth: template rendering escapes even names not admitted by the API.
	i.Username = `<script>alert(1)</script>`
	html, _, err = invitationEmail(PanelBrandingConfig{BrandName: brand, AccentColor: `red; background:url(https://attacker.test)`}, &i, server, link)
	if err != nil || strings.Contains(html, "<script>") || strings.Contains(html, "attacker.test") || !strings.Contains(html, "#387aff") {
		t.Fatalf("unsafe template rendering: %v", err)
	}
}
