package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestPanelRolesProtectLocalAndRemoteCreation(t *testing.T) {
	master := newReliabilityFixture(t)
	worker := newReliabilityFixture(t)
	node := httptest.NewServer(worker.router)
	defer node.Close()
	if err := master.db.CreateNode(&database.ClusterNode{ID: "remote", Name: "Worker", BaseURL: node.URL, APIToken: worker.cfg.Security.NodeToken}); err != nil {
		t.Fatal(err)
	}
	paths := []string{"/api/servers", "/api/nodes/remote/proxy/api/servers"}
	for _, path := range paths {
		if w := master.request("POST", path, "{}", "owner"); w.Code != 403 {
			t.Fatal("ordinary owner created server", path, w.Code, w.Body.String())
		}
	}
	role := &database.PanelRole{Name: "Server creators", Permissions: []string{database.PermissionServersCreate}}
	if err := master.db.CreatePanelRole(role); err != nil {
		t.Fatal(err)
	}
	if err := master.db.AssignPanelRole(master.users["owner"].ID, role.ID); err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		if w := master.request("POST", path, "{}", "owner"); w.Code != 400 {
			t.Fatal("authorized creator did not reach validation", path, w.Code, w.Body.String())
		}
	}
	if w := worker.request("POST", "/api/servers", "{}", "owner"); w.Code != 403 {
		t.Fatal("worker synchronization conferred global creation", w.Code, w.Body.String())
	}
	if err := master.db.AssignPanelRole(master.users["owner"].ID, ""); err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		if w := master.request("POST", path, "{}", "owner"); w.Code != 403 {
			t.Fatal("revoked role persisted in existing session", path, w.Code, w.Body.String())
		}
	}
	for _, withIdentity := range []bool{false, true} {
		req := httptest.NewRequest("POST", "/api/servers", strings.NewReader("{}"))
		req.Header.Set("Authorization", "Bearer "+worker.tokens["owner"])
		req.Header.Set("X-NeoPanel-Can-Create-Server", "true")
		if withIdentity {
			req.Header.Set("X-NeoPanel-User-ID", fmt.Sprint(worker.users["owner"].ID))
			req.Header.Set("X-NeoPanel-User-Role", "user")
		}
		w := httptest.NewRecorder()
		worker.router.ServeHTTP(w, req)
		if w.Code != 403 {
			t.Fatal("forged authorization header accepted", w.Code, w.Body.String())
		}
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
func TestPanelRoleRoutesAreAdminOnly(t *testing.T) {
	f := newReliabilityFixture(t)
	for _, route := range []struct{ method, path, body string }{{"GET", "/api/admin/roles", ""}, {"POST", "/api/admin/roles", `{"name":"Creator","permissions":["servers.create"]}`}, {"PUT", "/api/admin/roles/missing", `{"name":"Creator","permissions":[]}`}, {"DELETE", "/api/admin/roles/missing", ""}, {"PUT", fmt.Sprintf("/api/admin/users/%d/panel-role", f.users["owner"].ID), `{"role_id":""}`}} {
		if w := f.request(route.method, route.path, route.body, "owner"); w.Code != 403 {
			t.Fatal(route, w.Code, w.Body.String())
		}
	}
	if w := f.request("GET", "/api/admin/roles", "", "admin"); w.Code != 200 || !strings.Contains(w.Body.String(), "servers.create") {
		t.Fatal(w.Code, w.Body.String())
	}
}
