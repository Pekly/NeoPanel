package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestNodeNameLocalPersistsAndRelists(t *testing.T) {
	f := newReliabilityFixture(t)
	if got := f.router.nodes.resolveLocalNodeName(); got != "Local Master Node" {
		t.Fatal(got)
	}
	cfgBefore := *f.cfg
	w := f.request("PATCH", "/api/nodes/local/name", `{"name":"  Praha 🐧  "}`, "admin")
	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil || w.Code != 200 || !reflect.DeepEqual(response, map[string]string{"id": "local", "name": "Praha 🐧"}) {
		t.Fatal(w.Code, w.Body.String(), err)
	}
	if !reflect.DeepEqual(cfgBefore, *f.cfg) {
		t.Fatal("rename mutated startup configuration")
	}
	if nodes, err := f.db.ListNodes(); err != nil || len(nodes) != 0 {
		t.Fatal("local rename inserted a remote node", nodes, err)
	}
	reopened, err := database.InitDB(filepath.Join(f.cfg.Paths.DataDir, "panel.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	h := NewNodeHandler(f.cfg, reopened, f.router.auth, f.manager)
	if got, err := reopened.GetPanelSetting(localNodeNameSettingKey); err != nil || got != "Praha 🐧" {
		t.Fatal(got, err)
	}
	if got := h.buildLocalNodeView("panel.example"); got.Name != "Praha 🐧" || got.ID != "local" || !got.IsLocal {
		t.Fatal(got)
	}
	for _, user := range []string{"admin", "owner"} {
		w = f.request("GET", "/api/nodes", "", user)
		var nodes []NodeStatusView
		if err := json.Unmarshal(w.Body.Bytes(), &nodes); err != nil || w.Code != 200 || len(nodes) != 1 || nodes[0].Name != "Praha 🐧" {
			t.Fatal(w.Code, w.Body.String(), err)
		}
	}
	w = f.request("GET", "/api/cluster/servers", "", "admin")
	var cluster struct {
		Servers []struct {
			NodeID   string `json:"node_id"`
			NodeName string `json:"node_name"`
		} `json:"servers"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &cluster); err != nil || w.Code != 200 || len(cluster.Servers) != 2 {
		t.Fatal(w.Code, w.Body.String(), err)
	}
	for _, server := range cluster.Servers {
		if server.NodeID != "local" || server.NodeName != "Praha 🐧" {
			t.Fatal(server)
		}
	}
	logs, err := f.db.ListActivity(f.users["admin"].ID, true, 100)
	if err != nil || len(logs) != 1 || logs[0].Action != "node_rename" || !strings.Contains(logs[0].Details, "Praha 🐧") {
		t.Fatal(logs, err)
	}
}

func nodeNameRemote(t *testing.T, f *reliabilityFixture, baseURL string) *database.ClusterNode {
	t.Helper()
	node := &database.ClusterNode{ID: "remote", Name: "Original", BaseURL: baseURL, APIToken: "secret-token-preserved", FQDN: "game.example", Location: "Prague", IsFree: true, IdleTimeoutMin: 17, AllowUserRAM: true, AllowUserPorts: true, AllowUserJava: false, IsolationEngine: "docker", CPULimitPercent: 150, TasksMax: 321}
	if err := f.db.CreateNode(node); err != nil {
		t.Fatal(err)
	}
	stored, err := f.db.GetNode(node.ID)
	if err != nil {
		t.Fatal(err)
	}
	return stored
}

func TestNodeNameRemotePreservesPropertiesWithoutNetwork(t *testing.T) {
	f := newReliabilityFixture(t)
	var calls atomic.Int32
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); w.WriteHeader(500) }))
	defer remote.Close()
	before := nodeNameRemote(t, f, remote.URL)
	// Exactly 100 non-ASCII code points must be allowed, not counted as bytes.
	name := strings.Repeat("🐧", 100)
	body, _ := json.Marshal(map[string]string{"name": name})
	w := f.request("PATCH", "/api/nodes/remote/name", string(body), "admin")
	if w.Code != 200 || strings.Contains(w.Body.String(), before.APIToken) {
		t.Fatal(w.Code, w.Body.String())
	}
	after, err := f.db.GetNode("remote")
	if err != nil || after.Name != name || after.UpdatedAt.Before(before.UpdatedAt) {
		t.Fatal(after, err)
	}
	after.Name, after.UpdatedAt = before.Name, before.UpdatedAt
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("rename changed unrelated properties: before=%+v after=%+v", before, after)
	}
	if calls.Load() != 0 {
		t.Fatal("rename contacted the remote node")
	}
}

func TestNodeNameRejectedRequestsDoNotMutate(t *testing.T) {
	f := newReliabilityFixture(t)
	before := nodeNameRemote(t, f, "http://127.0.0.1:1")
	if err := f.db.SetPanelSetting(localNodeNameSettingKey, "Local original"); err != nil {
		t.Fatal(err)
	}
	assertUnchanged := func() {
		t.Helper()
		after, err := f.db.GetNode("remote")
		if err != nil || !reflect.DeepEqual(before, after) {
			t.Fatal("rejected request mutated remote node", after, err)
		}
		if name, err := f.db.GetPanelSetting(localNodeNameSettingKey); err != nil || name != "Local original" {
			t.Fatal("rejected request mutated local name", name, err)
		}
		if logs, err := f.db.ListActivity(0, true, 100); err != nil || len(logs) != 0 {
			t.Fatal("rejected request logged a successful rename", logs, err)
		}
	}
	invalid := []string{"", `{}`, `null`, `[]`, `{"name":null}`, `{"name":1}`, `{"name":""}`, `{"name":"   "}`, `{"name":"x","unknown":true}`, `{"name":"x","name":"y"}`, `{"Name":"x"}`, `{"name":"x"} {}`, `{"name":"x"} trailing`, `{"name":"x"`, `{"name":"x\n"}`, `{"name":"\tx"}`, `{"name":"x\u0000y"}`, `{"name":"x\u007fy"}`, `{"name":"x\u0085y"}`, `{"name":"` + strings.Repeat("🐧", 101) + `"}`, `{"name":"` + strings.Repeat("x", 5000) + `"}`, `{"name":"x"}` + strings.Repeat(" ", 4096), "{\"name\":\"\xff\"}"}
	for _, id := range []string{"local", "remote"} {
		for _, body := range invalid {
			w := f.request("PATCH", "/api/nodes/"+id+"/name", body, "admin")
			if w.Code != 400 {
				t.Fatalf("accepted invalid body %q: %d %s", body, w.Code, w.Body.String())
			}
			assertUnchanged()
		}
		for _, user := range []string{"", "owner", "subuser", "stranger"} {
			want := 403
			if user == "" {
				want = 401
			}
			w := f.request("PATCH", "/api/nodes/"+id+"/name", `{"name":"Forbidden"}`, user)
			if w.Code != want {
				t.Fatal(user, w.Code, w.Body.String())
			}
			assertUnchanged()
		}
		for _, claims := range []*JWTClaims{nil, {Role: "user"}} {
			req := httptest.NewRequest("PATCH", "/", strings.NewReader(`{"name":"Forbidden"}`))
			req.SetPathValue("id", id)
			if claims != nil {
				req = req.WithContext(context.WithValue(req.Context(), contextUserKey, claims))
			}
			w := httptest.NewRecorder()
			f.router.nodes.RenameNode(w, req)
			if w.Code != 403 {
				t.Fatal("handler lacks admin guard", w.Code)
			}
			assertUnchanged()
		}
	}
	if w := f.request("PATCH", "/api/nodes/missing/name", `{"name":"Unknown"}`, "admin"); w.Code != 404 {
		t.Fatal(w.Code, w.Body.String())
	}
	if nodes, err := f.db.ListNodes(); err != nil || len(nodes) != 1 {
		t.Fatal(nodes, err)
	}
	assertUnchanged()
}
