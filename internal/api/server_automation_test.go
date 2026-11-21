package api

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
)

func automationView(t *testing.T, w *httptest.ResponseRecorder) runner.ServerAutomation {
	t.Helper()
	var view runner.ServerAutomation
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &view) != nil {
		t.Fatal(w.Code, w.Body.String())
	}
	var fields map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &fields); err != nil || len(fields) != 9 {
		t.Fatal("response contract", fields, err)
	}
	if !view.IdleShutdownLocked {
		t.Fatal("idle policy claims user override")
	}
	return view
}

func TestServerAutomationRouterAuthorizationAndStrictPatch(t *testing.T) {
	f := newReliabilityFixture(t)
	path := "/api/servers/alpha/automation"
	if err := f.db.SetNodeIdlePolicy("local", database.NodeIdlePolicy{Enabled: true, TimeoutMinutes: 27}); err != nil {
		t.Fatal(err)
	}
	for _, user := range []string{"owner", "admin", "subuser"} {
		v := automationView(t, f.request("GET", path, "", user))
		if !v.JoinToStartEnabled || v.StoppedByIdlePolicy || v.WakeEligible || !v.IdleShutdownEnabled || v.IdleTimeoutMinutes != 27 || v.CanEditJoinToStart != (user != "subuser") {
			t.Fatal(user, v)
		}
	}
	for _, method := range []string{"GET", "PATCH"} {
		if w := f.request(method, path, `{"join_to_start_enabled":false}`, "stranger"); w.Code != 403 {
			t.Fatal(method, w.Code, w.Body)
		}
	}
	if w := f.request("PATCH", path, `{"join_to_start_enabled":false}`, "subuser"); w.Code != 403 {
		t.Fatal(w.Code, w.Body)
	}
	members, _ := f.db.ListServerSubusers("alpha")
	for _, perms := range [][]string{{"settings.edit"}, {"server.power"}, {"settings.edit", "server.power"}} {
		if err := f.db.UpdateServerSubuser("alpha", members[0].ID, perms); err != nil {
			t.Fatal(err)
		}
		w := f.request("PATCH", path, `{"join_to_start_enabled":false}`, "subuser")
		if len(perms) == 1 {
			if w.Code != 403 {
				t.Fatal(perms, w.Code, w.Body)
			}
		} else {
			v := automationView(t, w)
			if v.JoinToStartEnabled || !v.CanEditJoinToStart || !v.IdleShutdownEnabled {
				t.Fatal(v)
			}
		}
	}
	for _, body := range []string{`{}`, `null`, `{"join_to_start_enabled":null}`, `{"join_to_start_enabled":true,"join_to_start_enabled":false}`, `{"join_to_start_enabled":true,"wake_eligible":true}`, `{"join_to_start_enabled":true,"stopped_by_idle_policy":true}`, `{"idle_shutdown_enabled":false}`, `{"join_to_start_enabled":"true"}`, `{"join_to_start_enabled":true} {}`, strings.Repeat(" ", 4097)} {
		if w := f.request("PATCH", path, body, "owner"); w.Code != 400 {
			t.Fatal(body, w.Code, w.Body)
		}
	}
	v := automationView(t, f.request("PATCH", path, `{"join_to_start_enabled":true}`, "owner"))
	if !v.JoinToStartEnabled {
		t.Fatal(v)
	}
	if w := f.request("PUT", "/api/admin/nodes/local/idle-policy", `{"enabled":false,"timeout_minutes":27}`, "owner"); w.Code != 403 {
		t.Fatal(w.Code, w.Body)
	}
	policy, err := f.db.GetNodeIdlePolicy("local")
	if err != nil || !policy.Enabled || policy.TimeoutMinutes != 27 {
		t.Fatal(policy, err)
	}
	// Preference-off never removes ordinary power permission.
	automationView(t, f.request("PATCH", path, `{"join_to_start_enabled":false}`, "owner"))
	if w := f.request("POST", "/api/servers/alpha/power", `{"action":"stop"}`, "owner"); w.Code != 200 {
		t.Fatal(w.Code, w.Body)
	}
	release, err := f.manager.BeginServerOperation("alpha")
	if err != nil {
		t.Fatal(err)
	}
	w := f.request("PATCH", path, `{"join_to_start_enabled":true}`, "owner")
	release()
	if w.Code != 409 {
		t.Fatal("operation gate", w.Code, w.Body)
	}
}

func TestServerAutomationRemoteHostPolicyAndForwardedPermissions(t *testing.T) {
	master, worker := newReliabilityFixture(t), newReliabilityFixture(t)
	host := httptest.NewServer(worker.router)
	defer host.Close()
	if err := master.db.CreateNode(&database.ClusterNode{ID: "remote", Name: "Worker", BaseURL: host.URL, APIToken: worker.cfg.Security.NodeToken}); err != nil {
		t.Fatal(err)
	}
	cfg := database.DefaultNodeWakeConfig()
	cfg.Enabled = true
	if err := worker.db.SetNodeWakeConfig("local", cfg); err != nil {
		t.Fatal(err)
	}
	if err := worker.db.SetNodeIdlePolicy("local", database.NodeIdlePolicy{Enabled: true, TimeoutMinutes: 19}); err != nil {
		t.Fatal(err)
	}
	path := "/api/nodes/remote/proxy/api/servers/alpha/automation"
	v := automationView(t, master.request("GET", path, "", "owner"))
	if !v.NodeJoinToStartEnabled || !v.IdleShutdownEnabled || v.IdleTimeoutMinutes != 19 {
		t.Fatal("not actual worker policy", v)
	}
	for _, user := range []string{"stranger", "subuser"} {
		if w := master.request("PATCH", path, `{"join_to_start_enabled":false}`, user); w.Code != 403 {
			t.Fatal(user, w.Code, w.Body)
		}
	}
	automationView(t, master.request("PATCH", path, `{"join_to_start_enabled":false}`, "owner"))
	local, _ := master.db.GetServerWakeState("alpha")
	remote, _ := worker.db.GetServerWakeState("alpha")
	if !local.JoinToStartEnabled || remote.JoinToStartEnabled {
		t.Fatal("preference stored on wrong host", local, remote)
	}
	members, _ := worker.db.ListServerSubusers("alpha")
	if err := worker.db.UpdateServerSubuser("alpha", members[0].ID, []string{"settings.edit"}); err != nil {
		t.Fatal(err)
	}
	if w := master.request("PATCH", path, `{"join_to_start_enabled":true}`, "subuser"); w.Code != 403 {
		t.Fatal(w.Code, w.Body)
	}
	if err := worker.db.UpdateServerSubuser("alpha", members[0].ID, []string{"settings.edit", "server.power"}); err != nil {
		t.Fatal(err)
	}
	automationView(t, master.request("PATCH", path, `{"join_to_start_enabled":true}`, "subuser"))
	// A machine credential carrying a real forwarded identity is not an admin override.
	req := httptest.NewRequest("PATCH", "/api/servers/alpha/automation", strings.NewReader(`{"join_to_start_enabled":false}`))
	req.Header.Set("Authorization", "Bearer "+worker.cfg.Security.NodeToken)
	req.Header.Set("X-NeoPanel-User-ID", fmt.Sprint(worker.users["stranger"].ID))
	req.Header.Set("X-NeoPanel-User-Role", "user")
	w := httptest.NewRecorder()
	worker.router.ServeHTTP(w, req)
	if w.Code != 403 {
		t.Fatal("forwarded identity escalated", w.Code, w.Body)
	}
}
