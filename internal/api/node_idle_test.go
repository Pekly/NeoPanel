package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func idleJSON(p database.NodeIdlePolicy) string { b, _ := json.Marshal(p); return string(b) }
func idleView(t *testing.T, w *httptest.ResponseRecorder, status int) NodeIdlePolicyView {
	t.Helper()
	var v NodeIdlePolicyView
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil || w.Code != status {
		t.Fatal(w.Code, w.Body.String(), err)
	}
	return v
}
func idleNodeView(t *testing.T, w *httptest.ResponseRecorder, status int, applied bool) NodeStatusView {
	t.Helper()
	var v NodeStatusView
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil || w.Code != status {
		t.Fatal(w.Code, w.Body.String(), err)
	}
	if v.IdlePolicyApplied == nil || *v.IdlePolicyApplied != applied || (v.IdlePolicyError != "") == applied {
		t.Fatal("incorrect saved-node delivery status", v)
	}
	return v
}

func TestNodeIdleLocalAuthorizationAndStrictValidation(t *testing.T) {
	f := newReliabilityFixture(t)
	path := "/api/admin/nodes/local/idle-policy"
	p := database.DefaultNodeIdlePolicy()
	if v := idleView(t, f.request("GET", path, "", "admin"), 200); v.Config != p || !v.Applied {
		t.Fatal(v)
	}
	for _, endpoint := range []string{path, "/api/cluster/idle-policy"} {
		for _, user := range []string{"", "owner", "stranger", "subuser"} {
			want := 403
			if user == "" {
				want = 401
			}
			if w := f.request("PUT", endpoint, idleJSON(p), user); w.Code != want {
				t.Fatal(endpoint, user, w.Code)
			}
			if endpoint == path {
				if w := f.request("GET", endpoint, "", user); w.Code != want {
					t.Fatal(user, w.Code)
				}
			}
		}
		invalid := []string{"", "null", "[]", "{}", `{"enabled":true}`, `{"enabled":true,"timeout_minutes":0}`, `{"enabled":false,"timeout_minutes":1441}`, `{"enabled":true,"timeout_minutes":-1}`, `{"enabled":true,"timeout_minutes":1.5}`, `{"enabled":null,"timeout_minutes":15}`, `{"Enabled":true,"timeout_minutes":15}`, `{"enabled":true,"enabled":false,"timeout_minutes":15}`, `{"enabled":true,"timeout_minutes":15,"extra":1}`, idleJSON(p) + " {}", idleJSON(p) + strings.Repeat(" ", nodeIdleBodyLimit)}
		for _, body := range invalid {
			if w := f.request("PUT", endpoint, body, "admin"); w.Code != 400 {
				t.Fatal(body, w.Code, w.Body.String())
			}
			if got, err := f.db.GetNodeIdlePolicy("local"); err != nil || got != p {
				t.Fatal("invalid mutation", got, err)
			}
		}
	}
	for _, handler := range []http.HandlerFunc{f.router.nodes.GetNodeIdlePolicy, f.router.nodes.PutNodeIdlePolicy, f.router.nodes.ApplyClusterIdlePolicy} {
		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest("PUT", "/", strings.NewReader(idleJSON(p))))
		if w.Code != 403 {
			t.Fatal("direct handler auth", w.Code)
		}
	}
	p.Enabled = true
	if v := idleView(t, f.request("PUT", path, idleJSON(p), "admin"), 200); !v.Applied || v.Config != p {
		t.Fatal(v)
	}
	if nodes, err := f.db.ListNodes(); err != nil || len(nodes) != 0 {
		t.Fatal("synthetic local", nodes, err)
	}
}

func TestNodeIdleActualWorkerAndPersistentAck(t *testing.T) {
	master, worker := newReliabilityFixture(t), newReliabilityFixture(t)
	worker.cfg.Server.Role = "node"
	worker.cfg.Security.NodeToken = "secret-token-preserved"
	server := httptest.NewServer(worker.router)
	defer server.Close()
	node := nodeNameRemote(t, master, server.URL)
	path := "/api/admin/nodes/remote/idle-policy"
	if v := idleView(t, master.request("GET", path, "", "admin"), 200); v.Applied || v.Error == "" {
		t.Fatal(v)
	}
	p := database.NodeIdlePolicy{Enabled: true, TimeoutMinutes: 27}
	for _, enabled := range []bool{true, false} {
		p.Enabled = enabled
		if v := idleView(t, master.request("PUT", path, idleJSON(p), "admin"), 200); !v.Applied || v.Config != p || v.Error != "" {
			t.Fatal(v)
		}
		if got, err := worker.db.GetNodeIdlePolicy("local"); err != nil || got != p {
			t.Fatal("worker", got, err)
		}
		if got, err := master.db.GetNodeIdlePolicy("remote"); err != nil || got != p {
			t.Fatal("master", got, err)
		}
	}
	if got, _ := master.db.GetNodeIdlePolicy("local"); got.Enabled {
		t.Fatal("enabled master")
	}
	if nodes, _ := worker.db.ListNodes(); len(nodes) != 0 {
		t.Fatal("synthetic worker node")
	}
	db, err := database.InitDB(filepath.Join(master.cfg.Paths.DataDir, "panel.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := NewNodeHandler(master.cfg, db, master.router.auth, master.manager)
	r := httptest.NewRequest("GET", path, nil)
	r.SetPathValue("id", "remote")
	r = r.WithContext(context.WithValue(r.Context(), contextUserKey, &JWTClaims{Role: "admin"}))
	w := httptest.NewRecorder()
	h.GetNodeIdlePolicy(w, r)
	if v := idleView(t, w, 200); !v.Applied || v.Config != p {
		t.Fatal(v)
	}
	// GET is last confirmed delivery, not a live-health probe.
	server.Close()
	if v := idleView(t, master.request("GET", path, "", "admin"), 200); !v.Applied {
		t.Fatal(v)
	}
	node, err = master.db.GetNode(node.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, changed := range []string{"token", "url", "policy"} {
		modified := *node
		switch changed {
		case "token":
			modified.APIToken = "new-secret"
		case "url":
			modified.BaseURL += "/new-destination"
		case "policy":
			modified.IdleTimeoutMin++
		}
		if err := master.db.UpdateNode(&modified); err != nil {
			t.Fatal(err)
		}
		if v := idleView(t, master.request("GET", path, "", "admin"), 200); v.Applied {
			t.Fatal("stale ack after "+changed, v)
		}
		if err := master.db.UpdateNode(node); err != nil {
			t.Fatal(err)
		}
	}
	if v := idleView(t, master.request("PUT", path, idleJSON(p), "admin"), 502); v.Applied || v.Error == "" {
		t.Fatal(v)
	}
}

func TestNodeIdleGenericCreateUpdateAndRepair(t *testing.T) {
	master, worker := newReliabilityFixture(t), newReliabilityFixture(t)
	worker.cfg.Server.Role = "node"
	worker.cfg.Security.NodeToken = "machine-secret"
	server := httptest.NewServer(worker.router)
	defer server.Close()
	body := fmt.Sprintf(`{"name":"worker","base_url":%q,"api_token":"machine-secret","is_free":true,"idle_timeout_min":0}`, server.URL)
	node := idleNodeView(t, master.request("POST", "/api/nodes", body, "admin"), 201, true)
	if p, err := worker.db.GetNodeIdlePolicy("local"); err != nil || !p.Enabled || p.TimeoutMinutes != 15 {
		t.Fatal(p, err)
	}
	route := "/api/nodes/" + node.ID
	node = idleNodeView(t, master.request("PUT", route, `{"is_free":false,"idle_timeout_min":42}`, "admin"), 200, true)
	if p, err := worker.db.GetNodeIdlePolicy("local"); err != nil || p.Enabled || p.TimeoutMinutes != 42 {
		t.Fatal(p, err)
	}
	// An absent policy in upsert/re-pair preserves disabled+positive timeout.
	body = fmt.Sprintf(`{"name":"worker","base_url":%q,"api_token":"machine-secret","upsert":true}`, server.URL)
	repaired := idleNodeView(t, master.request("POST", "/api/nodes", body, "admin"), 200, true)
	if repaired.ID != node.ID || repaired.IsFree || repaired.IdleTimeoutMin != 42 {
		t.Fatal(repaired)
	}
	before, _ := master.db.GetNode(node.ID)
	for _, minutes := range []int{-1, 1441} {
		invalid := fmt.Sprintf(`{"name":"changed","is_free":true,"idle_timeout_min":%d}`, minutes)
		if w := master.request("PUT", route, invalid, "admin"); w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
		invalid = fmt.Sprintf(`{"name":"bad","base_url":%q,"api_token":"machine-secret","idle_timeout_min":%d}`, server.URL, minutes)
		if w := master.request("POST", "/api/nodes", invalid, "admin"); w.Code != 400 {
			t.Fatal(w.Code)
		}
	}
	after, _ := master.db.GetNode(node.ID)
	if *before != *after {
		t.Fatal("invalid save mutated node")
	}
	old := httptest.NewServer(http.NotFoundHandler())
	defer old.Close()
	body = fmt.Sprintf(`{"base_url":%q,"is_free":true,"idle_timeout_min":8}`, old.URL)
	idleNodeView(t, master.request("PUT", route, body, "admin"), 200, false)
	if p, _ := master.db.GetNodeIdlePolicy(node.ID); !p.Enabled || p.TimeoutMinutes != 8 {
		t.Fatal("desired not saved", p)
	}
	body = fmt.Sprintf(`{"name":"old","base_url":%q,"api_token":"machine-secret"}`, old.URL)
	idleNodeView(t, master.request("POST", "/api/nodes", body, "admin"), 201, false)
	// Re-pair to a new destination must deliver the preserved policy there.
	worker2 := newReliabilityFixture(t)
	worker2.cfg.Server.Role = "node"
	worker2.cfg.Security.NodeToken = "new-secret"
	server2 := httptest.NewServer(worker2.router)
	defer server2.Close()
	body = fmt.Sprintf(`{"name":"worker","base_url":%q,"api_token":"new-secret","upsert":true}`, server2.URL)
	idleNodeView(t, master.request("POST", "/api/nodes", body, "admin"), 200, true)
	if p, _ := worker2.db.GetNodeIdlePolicy("local"); !p.Enabled || p.TimeoutMinutes != 8 {
		t.Fatal("repair lost desired", p)
	}
}

func TestNodeIdleFailuresRedirectsAndBadAck(t *testing.T) {
	f := newReliabilityFixture(t)
	var mode, leaked atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { leaked.Add(1) }))
	defer target.Close()
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" || r.URL.Path != "/api/cluster/idle-policy" || r.Header.Get("Authorization") != "Bearer secret-token-preserved" {
			t.Error("wrong delivery request")
		}
		switch mode.Load() {
		case 0:
			w.WriteHeader(404)
		case 1:
			w.WriteHeader(405)
		case 2:
			http.Redirect(w, r, target.URL, 307)
		case 3:
			fmt.Fprint(w, `{"config":{"enabled":true,"timeout_minutes":15},"applied":true,"error":""}`)
		case 4:
			fmt.Fprint(w, `{"applied":true}`)
		case 5:
			fmt.Fprint(w, `{"config":{"enabled":false,"timeout_minutes":15},"applied":false,"error":""}`)
		case 6:
			fmt.Fprint(w, `{"config":{"enabled":false,"timeout_minutes":15},"applied":true,"error":""} {}`)
		}
	}))
	defer remote.Close()
	nodeNameRemote(t, f, remote.URL)
	path := "/api/admin/nodes/remote/idle-policy"
	p := database.DefaultNodeIdlePolicy()
	for i := 0; i < 7; i++ {
		mode.Store(int32(i))
		if v := idleView(t, f.request("PUT", path, idleJSON(p), "admin"), 502); v.Applied || v.Error == "" || v.Config != p {
			t.Fatal(i, v)
		}
		if v := idleView(t, f.request("GET", path, "", "admin"), 200); v.Applied || v.Error == "" {
			t.Fatal(i, v)
		}
	}
	if leaked.Load() != 0 {
		t.Fatal("redirect leaked credentials")
	}
}

func TestNodeIdleGenericAndDedicatedSerialized(t *testing.T) {
	for _, genericFirst := range []bool{true, false} {
		t.Run(fmt.Sprint(genericFirst), func(t *testing.T) {
			f := newReliabilityFixture(t)
			started, release := make(chan struct{}), make(chan struct{})
			var calls atomic.Int32
			remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/cluster/idle-policy" {
					w.WriteHeader(404)
					return
				}
				var p database.NodeIdlePolicy
				if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
					t.Error(err)
				}
				if calls.Add(1) == 1 {
					close(started)
					<-release
				}
				writeJSON(w, 200, NodeIdlePolicyView{Config: p, Applied: true})
			}))
			defer remote.Close()
			nodeNameRemote(t, f, remote.URL)
			first, second := make(chan *httptest.ResponseRecorder, 1), make(chan *httptest.ResponseRecorder, 1)
			genericPath, dedicatedPath := "/api/nodes/remote", "/api/admin/nodes/remote/idle-policy"
			go func() {
				if genericFirst {
					first <- f.request("PUT", genericPath, `{"is_free":true,"idle_timeout_min":5}`, "admin")
				} else {
					first <- f.request("PUT", dedicatedPath, `{"enabled":false,"timeout_minutes":9}`, "admin")
				}
			}()
			select {
			case <-started:
			case <-time.After(3 * time.Second):
				close(release)
				t.Fatal("delivery not started")
			}
			go func() {
				if genericFirst {
					second <- f.request("PUT", dedicatedPath, `{"enabled":false,"timeout_minutes":9}`, "admin")
				} else {
					second <- f.request("PUT", genericPath, `{"name":"preserve latest policy"}`, "admin")
				}
			}()
			select {
			case <-second:
				close(release)
				t.Fatal("save overtook in-flight delivery")
			case <-time.After(50 * time.Millisecond):
			}
			close(release)
			if genericFirst {
				idleNodeView(t, <-first, 200, true)
				idleView(t, <-second, 200)
			} else {
				idleView(t, <-first, 200)
				idleNodeView(t, <-second, 200, true)
			}
			if p, _ := f.db.GetNodeIdlePolicy("remote"); p.Enabled || p.TimeoutMinutes != 9 {
				t.Fatal(p)
			}
			if v := idleView(t, f.request("GET", dedicatedPath, "", "admin"), 200); !v.Applied || v.Config.TimeoutMinutes != 9 {
				t.Fatal(v)
			}
		})
	}
}

type idleDeadlineTransport struct{ t *testing.T }

func (transport idleDeadlineTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	deadline, ok := r.Context().Deadline()
	if !ok || time.Until(deadline) > 5*time.Second || time.Until(deadline) < 4*time.Second {
		transport.t.Error("missing five-second delivery deadline")
	}
	return nil, context.DeadlineExceeded
}

func TestNodeIdleDeliveryDeadline(t *testing.T) {
	f := newReliabilityFixture(t)
	nodeNameRemote(t, f, "https://worker.invalid")
	f.router.nodes.httpClient = &http.Client{Transport: idleDeadlineTransport{t}}
	if v := idleView(t, f.request("PUT", "/api/admin/nodes/remote/idle-policy", idleJSON(database.DefaultNodeIdlePolicy()), "admin"), 502); v.Applied || v.Error == "" {
		t.Fatal(v)
	}
}
