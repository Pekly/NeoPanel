package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func panelRoleRequest(t *testing.T, f *reliabilityFixture, handler http.HandlerFunc, user, body string, params map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	for k, v := range params {
		req.SetPathValue(k, v)
	}
	if user != "" {
		claims, err := f.router.auth.ValidateAuthToken(f.tokens[user])
		if err != nil {
			t.Fatal(err)
		}
		req = req.WithContext(context.WithValue(req.Context(), contextUserKey, claims))
	}
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func TestPanelRoleAdminAPI(t *testing.T) {
	f := newReliabilityFixture(t)
	h := f.router.auth
	call := func(fn http.HandlerFunc, user, body string, params map[string]string, status int) *httptest.ResponseRecorder {
		t.Helper()
		w := panelRoleRequest(t, f, fn, user, body, params)
		if w.Code != status {
			t.Fatalf("status %d want %d: %s", w.Code, status, w.Body)
		}
		return w
	}
	payload := `{"name":"Creators","description":"Provision servers","permissions":["servers.create"]}`
	call(h.CreatePanelRole, "", payload, nil, 401)
	for _, name := range []string{"owner", "subuser", "stranger"} {
		call(h.CreatePanelRole, name, payload, nil, 403)
		call(h.ListPanelRoles, name, "", nil, 403)
		call(h.AssignPanelRole, name, `{"role_id":"admin"}`, map[string]string{"id": strconv.Itoa(f.users[name].ID)}, 403)
	}
	for _, body := range []string{`null`, `{"name":"admin"}`, `{"name":"<b>"}`, `{"name":"Test","permissions":["*"]}`, `{"name":"Test","permissions":["files.view"]}`, `{"name":"Test","builtin":true}`, `{"name":"Test"} {}`} {
		call(h.CreatePanelRole, "admin", body, nil, 400)
	}
	w := call(h.CreatePanelRole, "admin", payload, nil, 201)
	var role database.PanelRole
	if err := json.Unmarshal(w.Body.Bytes(), &role); err != nil {
		t.Fatal(err)
	}
	call(h.CreatePanelRole, "admin", payload, nil, 409)
	w = call(h.ListPanelRoles, "admin", "", nil, 200)
	var catalog struct {
		Roles       []database.PanelRole `json:"roles"`
		Permissions []map[string]string  `json:"permissions"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &catalog); err != nil || len(catalog.Roles) != 3 || len(catalog.Permissions) != 1 || catalog.Permissions[0]["id"] != "servers.create" {
		t.Fatalf("catalog: %s %v", w.Body, err)
	}
	params := map[string]string{"id": strconv.Itoa(f.users["stranger"].ID)}
	for _, body := range []string{`{}`, `null`, `{"role_id":null}`, `{"role_id":12}`, `{"role_id":"","role":"admin"}`} {
		call(h.AssignPanelRole, "admin", body, params, 400)
	}
	call(h.AssignPanelRole, "admin", `{"role_id":"admin"}`, params, 404)
	call(h.AssignPanelRole, "admin", fmt.Sprintf(`{"role_id":%q}`, role.ID), params, 200)
	call(h.DeletePanelRole, "admin", "", map[string]string{"roleId": role.ID}, 409)
	call(h.UpdatePanelRole, "admin", `{"name":"Creators renamed","permissions":[]}`, map[string]string{"roleId": role.ID}, 200)
	w = call(h.ListUsers, "admin", "", nil, 200)
	var users []database.User
	if err := json.Unmarshal(w.Body.Bytes(), &users); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, user := range users {
		if user.ID == f.users["stranger"].ID {
			found = user.PanelRoleID == role.ID && user.PanelRoleName == "Creators renamed" && user.PanelPermissions != nil
		}
	}
	if !found {
		t.Fatalf("missing user role: %s", w.Body)
	}
	call(h.AssignPanelRole, "admin", `{"role_id":""}`, params, 200)
	call(h.DeletePanelRole, "admin", "", map[string]string{"roleId": role.ID}, 204)
}

func TestPanelRoleCreateJWTAndAPIKeyRevocation(t *testing.T) {
	for _, kind := range []string{"jwt", "api-key"} {
		t.Run(kind, func(t *testing.T) {
			f := newReliabilityFixture(t)
			if kind == "api-key" {
				token := "neo_panel_role_test"
				sum := sha256.Sum256([]byte(token))
				if _, err := f.db.AddAPIKey(f.users["stranger"].ID, "role-test", hex.EncodeToString(sum[:]), "neo_panel"); err != nil {
					t.Fatal(err)
				}
				f.tokens["stranger"] = token
			}
			create := func(user string, status int) {
				t.Helper()
				w := f.request("POST", "/api/servers", `{"name":"Test","type":"custom"}`, user)
				if w.Code != status {
					t.Fatalf("create: %d want %d %s", w.Code, status, w.Body)
				}
			}
			create("admin", 201)
			create("owner", 403)
			create("subuser", 403)
			create("stranger", 403)
			role := &database.PanelRole{Name: "Creators", Permissions: []string{"servers.create"}}
			if err := f.db.CreatePanelRole(role); err != nil {
				t.Fatal(err)
			}
			if err := f.db.AssignPanelRole(f.users["stranger"].ID, role.ID); err != nil {
				t.Fatal(err)
			}
			create("stranger", 201)
			if w := f.request("GET", "/api/servers/alpha", "", "stranger"); w.Code != 403 {
				t.Fatalf("custom role gained server rights: %d", w.Code)
			}
			if w := f.request("GET", "/api/admin/users", "", "stranger"); w.Code != 403 {
				t.Fatalf("custom role gained admin: %d", w.Code)
			}
			me := f.request("GET", "/api/auth/me", "", "stranger")
			if me.Code != 200 || !strings.Contains(me.Body.String(), `"panel_role_id":"`+role.ID+`"`) || !strings.Contains(me.Body.String(), `"panel_permissions":["servers.create"]`) {
				t.Fatalf("me: %d %s", me.Code, me.Body)
			}
			role.Permissions = []string{}
			if err := f.db.UpdatePanelRole(role); err != nil {
				t.Fatal(err)
			}
			create("stranger", 403)
			role.Permissions = []string{"servers.create"}
			if err := f.db.UpdatePanelRole(role); err != nil {
				t.Fatal(err)
			}
			create("stranger", 201)
			if err := f.db.AssignPanelRole(f.users["stranger"].ID, ""); err != nil {
				t.Fatal(err)
			}
			create("stranger", 403)
			if err := f.db.DeleteUser(f.users["stranger"].ID); err != nil {
				t.Fatal(err)
			}
			if _, err := f.router.auth.ValidateAuthToken(f.tokens["stranger"]); err == nil {
				t.Fatal("deleted account token accepted")
			}
		})
	}
}

func TestPanelRoleAdminDemotionRefreshesExistingTokens(t *testing.T) {
	f := newReliabilityFixture(t)
	token := "neo_admin_role_test"
	sum := sha256.Sum256([]byte(token))
	if _, err := f.db.AddAPIKey(f.users["admin"].ID, "admin-test", hex.EncodeToString(sum[:]), "neo_admin"); err != nil {
		t.Fatal(err)
	}
	if err := f.db.UpdateUserRole(f.users["admin"].ID, "user"); err != nil {
		t.Fatal(err)
	}
	if err := f.db.UpdateUserProfile(f.users["admin"].ID, "renamed", "", "", ""); err != nil {
		t.Fatal(err)
	}
	for _, tok := range []string{token, f.tokens["admin"]} {
		claims, err := f.router.auth.ValidateAuthToken(tok)
		if err != nil || claims.Role != "user" || claims.Username != "renamed" {
			t.Fatalf("stale identity: %+v %v", claims, err)
		}
		req := httptest.NewRequest("POST", "/api/servers", nil)
		req = req.WithContext(context.WithValue(req.Context(), contextUserKey, claims))
		if ok, err := f.router.auth.CanCreateServer(req); ok || err != nil {
			t.Fatalf("demoted admin can create: %v %v", ok, err)
		}
	}
	if w := f.request("GET", "/api/admin/users", "", "admin"); w.Code != 403 {
		t.Fatalf("demoted JWT admin: %d", w.Code)
	}
}

func TestPanelRoleLoginFieldsAndForwarding(t *testing.T) {
	f := newReliabilityFixture(t)
	role := &database.PanelRole{Name: "Creators", Permissions: []string{"servers.create"}}
	if err := f.db.CreatePanelRole(role); err != nil {
		t.Fatal(err)
	}
	if err := f.db.AssignPanelRole(f.users["stranger"].ID, role.ID); err != nil {
		t.Fatal(err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.db.UpdateUserPassword(f.users["stranger"].ID, string(hash)); err != nil {
		t.Fatal(err)
	}
	w := panelRoleRequest(t, f, f.router.auth.Login, "", `{"username":"stranger","password":"password"}`, nil)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"panel_permissions":["servers.create"]`) || !strings.Contains(w.Body.String(), `"panel_role_name":"Creators"`) {
		t.Fatalf("login: %d %s", w.Code, w.Body)
	}
	req := httptest.NewRequest("POST", "/api/servers", nil)
	req.Header.Set("X-NeoPanel-User-ID", strconv.Itoa(f.users["owner"].ID))
	req.Header.Set("X-NeoPanel-User-Role", "user")
	req.Header.Set("X-NeoPanel-Can-Create-Server", "true")
	nodeClaims, err := f.router.auth.ValidateAuthToken(f.cfg.Security.NodeToken)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := forwardedClaims(f.db, req, nodeClaims)
	if err != nil {
		t.Fatal(err)
	}
	// Clear any router-integrated authorization: this test verifies the helper's
	// contract separately from the parent's trusted forwarding-header integration.
	claims.AuthorizedServerCreate = false
	req = req.WithContext(context.WithValue(req.Context(), contextUserKey, claims))
	if ok, err := f.router.auth.CanCreateServer(req); ok || err != nil {
		t.Fatalf("forwarded member bypass: %v %v", ok, err)
	}
	claims.AuthorizedServerCreate = true
	if ok, err := f.router.auth.CanCreateServer(req); !ok || err != nil {
		t.Fatalf("trusted create claim rejected: %v %v", ok, err)
	}
	var parsed JWTClaims
	if err := json.Unmarshal([]byte(`{"user_id":2,"role":"user","AuthorizedServerCreate":true,"authorized_server_create":true}`), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.AuthorizedServerCreate {
		t.Fatal("trusted create claim parsed from JSON")
	}
}
