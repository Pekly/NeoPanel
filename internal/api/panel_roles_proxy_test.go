package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestPanelRoleRetirementProtectsLocalAndRemoteCreation(t *testing.T) {
	for _, kind := range []string{"jwt", "api-key"} {
		t.Run(kind, func(t *testing.T) {
			master := newReliabilityFixture(t)
			worker := newReliabilityFixture(t)
			if kind == "api-key" {
				for _, name := range []string{"admin", "owner", "stranger", "subuser"} {
					master.tokens[name] = retirementAPIKey(t, master, name)
				}
			}
			retireLegacyFixture(t, master)
			retireLegacyFixture(t, worker)
			node := httptest.NewServer(worker.router)
			defer node.Close()
			if err := master.db.CreateNode(&database.ClusterNode{ID: "remote", Name: "Worker", BaseURL: node.URL, APIToken: worker.cfg.Security.NodeToken}); err != nil {
				t.Fatal(err)
			}
			paths := []string{"/api/servers", "/api/nodes/remote/proxy/api/servers"}
			for _, path := range paths {
				for _, user := range []string{"owner", "stranger", "subuser"} {
					if w := master.request("POST", path, `{"name":"Denied","type":"custom"}`, user); w.Code != 403 {
						t.Fatalf("former custom user %s created via %s: %d %s", user, path, w.Code, w.Body)
					}
				}
				if w := master.request("POST", path, `{"name":"Admin server","type":"custom"}`, "admin"); w.Code != 201 {
					t.Fatalf("admin creation %s: %d %s", path, w.Code, w.Body)
				}
			}
			// Retirement must not remove ownership or granular team access, nor
			// turn former creators into administrators on either installation.
			for _, prefix := range []string{"", "/api/nodes/remote/proxy"} {
				for _, check := range []struct {
					method, suffix, user string
					status               int
				}{
					{"GET", "/api/servers/alpha", "owner", 200},
					{"GET", "/api/servers/alpha/files", "subuser", 200},
					{"POST", "/api/servers/alpha/command", "subuser", 403},
					{"GET", "/api/servers/alpha", "stranger", 403},
				} {
					if w := master.request(check.method, prefix+check.suffix, "{}", check.user); w.Code != check.status {
						t.Fatalf("team rights %s %+v: %d %s", prefix, check, w.Code, w.Body)
					}
				}
			}
			if w := master.request("GET", "/api/admin/users", "", "stranger"); w.Code != 403 {
				t.Fatalf("former creator gained admin: %d", w.Code)
			}
			if err := master.db.UpdateUserRole(master.users["admin"].ID, "user"); err != nil {
				t.Fatal(err)
			}
			for _, path := range paths {
				if w := master.request("POST", path, `{"name":"Denied","type":"custom"}`, "admin"); w.Code != 403 {
					t.Fatalf("demoted admin created via %s: %d %s", path, w.Code, w.Body)
				}
			}
			if err := master.db.DeleteUser(master.users["stranger"].ID); err != nil {
				t.Fatal(err)
			}
			if _, err := master.router.auth.ValidateAuthToken(master.tokens["stranger"]); err == nil {
				t.Fatal("deleted account token accepted")
			}
		})
	}
}

func TestServerCreationIgnoresRetiredForwardedGrant(t *testing.T) {
	f := newReliabilityFixture(t)
	for _, token := range []string{f.tokens["owner"], f.cfg.Security.NodeToken} {
		for _, role := range []string{"user", "admin"} {
			req := httptest.NewRequest("POST", "/api/servers", strings.NewReader(`{"name":"Denied","type":"custom"}`))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("X-NeoPanel-User-ID", fmt.Sprint(f.users["owner"].ID))
			req.Header.Set("X-NeoPanel-User-Role", role)
			req.Header.Set("X-NeoPanel-Can-Create-Server", "true")
			w := httptest.NewRecorder()
			f.router.ServeHTTP(w, req)
			if w.Code != 403 {
				t.Fatalf("stale/forged forwarded create accepted (%s): %d %s", role, w.Code, w.Body)
			}
		}
	}
	// Header alone cannot promote an ordinary authenticated user.
	req := httptest.NewRequest("POST", "/api/servers", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer "+f.tokens["owner"])
	req.Header.Set("X-NeoPanel-Can-Create-Server", "true")
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	if w.Code != 403 {
		t.Fatalf("untrusted create header accepted: %d", w.Code)
	}
	// Native configured machine tokens still administer nodes without a user.
	f.tokens["machine"] = f.cfg.Security.NodeToken
	if w := f.request("POST", "/api/servers", `{"name":"Machine server","type":"custom"}`, "machine"); w.Code != 201 {
		t.Fatalf("machine creation: %d %s", w.Code, w.Body)
	}
	headers := http.Header{"X-Neopanel-User-Id": []string{"999"}, "X-Neopanel-User-Role": []string{"admin"}, "X-Neopanel-Can-Create-Server": []string{"true"}}
	setProxyIdentity(headers, &JWTClaims{UserID: 0, Role: "admin"}, "node-token")
	if headers.Get("X-NeoPanel-User-ID") != "" || headers.Get("X-NeoPanel-Can-Create-Server") != "" {
		t.Fatal("client identity headers survived", headers)
	}
}

func TestUserProxyRefusesOldPermissionPolicy(t *testing.T) {
	f := newReliabilityFixture(t)
	mutated := false
	node := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/me" {
			w.Header().Set("X-NeoPanel-User-ID", fmt.Sprint(f.users["owner"].ID))
			writeJSON(w, 200, map[string]string{"role": "user"})
			return
		}
		mutated = true
		writeJSON(w, 200, map[string]string{"unsafe": "old worker"})
	}))
	defer node.Close()
	if err := f.db.CreateNode(&database.ClusterNode{ID: "old", Name: "Old", BaseURL: node.URL, APIToken: "neo_node_old"}); err != nil {
		t.Fatal(err)
	}
	if w := f.request("GET", "/api/nodes/old/proxy/api/servers/alpha/sftp", "", "owner"); w.Code != 502 || mutated {
		t.Fatal(w.Code, w.Body.String(), mutated)
	}
}

func TestPanelRoleRetirementRoutesUnavailable(t *testing.T) {
	f := newReliabilityFixture(t)
	before, err := f.db.ListUsers()
	if err != nil {
		t.Fatal(err)
	}
	for _, route := range []struct{ method, path, body string }{
		{"GET", "/api/admin/roles", ""},
		{"POST", "/api/admin/roles", `{"name":"Creator","permissions":["servers.create"]}`},
		{"PUT", "/api/admin/roles/missing", `{"name":"Creator","permissions":[]}`},
		{"DELETE", "/api/admin/roles/missing", ""},
		{"PUT", fmt.Sprintf("/api/admin/users/%d/panel-role", f.users["owner"].ID), `{"role_id":"admin"}`},
	} {
		for _, user := range []string{"", "owner", "admin"} {
			if w := f.request(route.method, route.path, route.body, user); w.Code != 404 && w.Code != 410 {
				t.Fatalf("legacy route available (%s) %+v: %d %s", user, route, w.Code, w.Body)
			}
		}
	}
	after, err := f.db.ListUsers()
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("legacy routes mutated users: %v", err)
	}
}
