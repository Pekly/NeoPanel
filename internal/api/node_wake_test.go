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
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func wakeJSON(c database.NodeWakeConfig) string { b, _ := json.Marshal(c); return string(b) }
func wakeView(t *testing.T, w *httptest.ResponseRecorder, status int) NodeWakeView {
	t.Helper()
	var v NodeWakeView
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil || w.Code != status {
		t.Fatal(w.Code, w.Body.String(), err)
	}
	return v
}

func TestNodeWakeLocalAndAuthorization(t *testing.T) {
	f := newReliabilityFixture(t)
	path := "/api/admin/nodes/local/wake"
	c := database.DefaultNodeWakeConfig()
	if v := wakeView(t, f.request("GET", path, "", "admin"), 200); v.Config != c || !v.Applied || v.Error != "" {
		t.Fatal(v)
	}
	for _, endpoint := range []string{path, "/api/cluster/wake"} {
		for _, user := range []string{"", "owner", "stranger", "subuser"} {
			want := 403
			if user == "" {
				want = 401
			}
			if w := f.request("PUT", endpoint, wakeJSON(c), user); w.Code != want {
				t.Fatal(endpoint, user, w.Code)
			}
			if endpoint == path {
				if w := f.request("GET", endpoint, "", user); w.Code != want {
					t.Fatal(user, w.Code)
				}
			}
		}
	}
	for _, handler := range []http.HandlerFunc{f.router.nodes.GetNodeWake, f.router.nodes.PutNodeWake, f.router.nodes.ApplyClusterWake} {
		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest("PUT", "/", strings.NewReader(wakeJSON(c))))
		if w.Code != 403 {
			t.Fatal("missing direct handler guard", w.Code)
		}
	}
	c.Enabled = true
	c.BindAddress = "127.0.0.1"
	c.OfflineMOTD = "§a{server}\nWake up!"
	if v := wakeView(t, f.request("PUT", path, wakeJSON(c), "admin"), 200); v.Config != c || !v.Applied || v.Error != "" {
		t.Fatal(v)
	}
	if got, err := f.db.GetNodeWakeConfig("local"); err != nil || got != c {
		t.Fatal(got, err)
	}
	if nodes, err := f.db.ListNodes(); err != nil || len(nodes) != 0 {
		t.Fatal(nodes, err)
	}
	c.Enabled = false
	if v := wakeView(t, f.request("PUT", path, wakeJSON(c), "admin"), 200); v.Config != c || !v.Applied {
		t.Fatal(v)
	}
	if w := f.request("PUT", "/api/admin/nodes/missing/wake", wakeJSON(c), "admin"); w.Code != 404 {
		t.Fatal(w.Code)
	}
}

func TestNodeWakeStrictRequests(t *testing.T) {
	f := newReliabilityFixture(t)
	c := database.DefaultNodeWakeConfig()
	wireConfig := c
	wireConfig.VersionLabel = "Join to start" // Stable input for mutation cases, independent of themed defaults.
	good := wakeJSON(wireConfig)
	invalid := []string{"", "null", "[]", "{}", good + " {}", good + " trailing", good[:len(good)-1],
		strings.Replace(good, `"enabled":false`, `"enabled":false,"enabled":true`, 1),
		strings.Replace(good, `"enabled":false`, `"Enabled":false`, 1),
		strings.Replace(good, `"enabled":false`, `"enabled":null`, 1),
		strings.Replace(good, `"enabled":false`, `"enabled":1`, 1),
		strings.Replace(good, `"enabled":false`, `"unknown":false`, 1),
		strings.Replace(good, `"cooldown_seconds":30`, `"cooldown_seconds":0`, 1),
		strings.Replace(good, `"max_players":20`, `"max_players":1.2`, 1),
		strings.Replace(good, `"bind_address":""`, `"bind_address":"localhost"`, 1),
		strings.Replace(good, `"icon_data_url":""`, `"icon_data_url":"https://example.com/icon.png"`, 1),
		strings.Replace(good, `"version_label":"Join to start"`, `"version_label":"x\u0000"`, 1),
		strings.Replace(good, "Join to start", "\xff", 1), good + strings.Repeat(" ", nodeWakeBodyLimit),
	}
	for _, path := range []string{"/api/admin/nodes/local/wake", "/api/cluster/wake"} {
		for i, body := range invalid {
			if w := f.request("PUT", path, body, "admin"); w.Code != 400 {
				t.Fatalf("case %d: %d %s", i, w.Code, w.Body.String())
			}
			if got, err := f.db.GetNodeWakeConfig("local"); err != nil || got != c {
				t.Fatal("invalid request mutated config", got, err)
			}
		}
	}
}

func TestNodeWakeRemoteActualWorkerAndDisable(t *testing.T) {
	master, worker := newReliabilityFixture(t), newReliabilityFixture(t)
	worker.cfg.Server.Role = "node"
	server := httptest.NewServer(worker.router)
	defer server.Close()
	node := nodeNameRemote(t, master, server.URL)
	worker.cfg.Security.NodeToken = node.APIToken
	path := "/api/admin/nodes/remote/wake"
	if v := wakeView(t, master.request("GET", path, "", "admin"), 200); v.Applied || v.Error == "" {
		t.Fatal("unconfirmed default is not applied", v)
	}
	c := database.DefaultNodeWakeConfig()
	c.Enabled = true
	c.VersionLabel = "Wake worker"
	for _, enabled := range []bool{true, false} {
		c.Enabled = enabled
		v := wakeView(t, master.request("PUT", path, wakeJSON(c), "admin"), 200)
		if v.Config != c || !v.Applied || v.Error != "" {
			t.Fatal(v)
		}
		if got, err := worker.db.GetNodeWakeConfig("local"); err != nil || got != c {
			t.Fatal("worker did not apply", got, err)
		}
		if got, err := master.db.GetNodeWakeConfig("remote"); err != nil || got != c {
			t.Fatal(got, err)
		}
		if got, _ := master.db.GetNodeWakeConfig("local"); got.Enabled {
			t.Fatal("remote update enabled master")
		}
		if v := wakeView(t, master.request("GET", path, "", "admin"), 200); !v.Applied || v.Config != c {
			t.Fatal(v)
		}
	}
	if got, _ := master.db.GetNode("remote"); !reflect.DeepEqual(got, node) {
		t.Fatal("wake changed generic node policy", got)
	}
	if nodes, err := worker.db.ListNodes(); err != nil || len(nodes) != 0 {
		t.Fatal(nodes, err)
	}
	// A freshly constructed handler and reopened DB retain acknowledgement.
	db, err := database.InitDB(filepath.Join(master.cfg.Paths.DataDir, "panel.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := NewNodeHandler(master.cfg, db, master.router.auth, master.manager)
	req := httptest.NewRequest("GET", path, nil)
	req.SetPathValue("id", "remote")
	req = req.WithContext(context.WithValue(req.Context(), contextUserKey, &JWTClaims{Role: "admin"}))
	w := httptest.NewRecorder()
	h.GetNodeWake(w, req)
	if v := wakeView(t, w, 200); !v.Applied || v.Config != c {
		t.Fatal(v)
	}
	worker.cfg.Security.NodeToken = "different-token"
	c.Enabled = true
	if v := wakeView(t, master.request("PUT", path, wakeJSON(c), "admin"), 502); v.Applied || v.Config != c || !strings.Contains(v.Error, "401") {
		t.Fatal(v)
	}
}

func TestNodeWakeRemoteFailuresRemainPending(t *testing.T) {
	f := newReliabilityFixture(t)
	var mode atomic.Int32
	var redirected atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { redirected.Add(1) }))
	defer target.Close()
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" || r.URL.Path != "/api/cluster/wake" || r.Header.Get("Authorization") != "Bearer secret-token-preserved" {
			t.Error("incorrect delivery route/auth")
		}
		var c database.NodeWakeConfig
		_ = json.NewDecoder(r.Body).Decode(&c)
		switch mode.Load() {
		case 0:
			writeJSON(w, 200, NodeWakeView{Config: c, Applied: true})
		case 1:
			w.WriteHeader(404)
		case 2:
			w.Write([]byte(`<html>old worker</html>`))
		case 3:
			c.VersionLabel = "wrong"
			writeJSON(w, 200, NodeWakeView{Config: c, Applied: true})
		case 4:
			writeJSON(w, 200, NodeWakeView{Config: c, Applied: false, Error: "refused"})
		case 5:
			w.WriteHeader(204)
		case 6:
			http.Redirect(w, r, target.URL, 307)
		case 7:
			w.Write([]byte(strings.Repeat("x", nodeWakeBodyLimit+1)))
		case 8:
			writeJSON(w, 200, map[string]any{"applied": true, "config": c})
		case 9:
			writeJSON(w, 200, NodeWakeView{Config: c, Applied: true, Error: "not really"})
		case 10:
			w.WriteHeader(405)
		case 11:
			w.WriteHeader(403)
		}
	}))
	defer remote.Close()
	nodeNameRemote(t, f, remote.URL)
	path := "/api/admin/nodes/remote/wake"
	c := database.DefaultNodeWakeConfig()
	c.Enabled = true
	if v := wakeView(t, f.request("PUT", path, wakeJSON(c), "admin"), 200); !v.Applied {
		t.Fatal(v)
	}
	c.Enabled = false
	for i := int32(1); i <= 11; i++ {
		mode.Store(i)
		v := wakeView(t, f.request("PUT", path, wakeJSON(c), "admin"), 502)
		if v.Applied || v.Error == "" || v.Config != c {
			t.Fatal(i, v)
		}
		if (i == 1 || i == 10) && (!strings.Contains(v.Error, "PUT /api/cluster/wake") || !strings.Contains(v.Error, "Update and restart the worker") || !strings.Contains(v.Error, "reverse proxy")) {
			t.Fatal("unsupported endpoint lacks actionable guidance", v.Error)
		}
		if i == 11 && !strings.Contains(v.Error, "Node Token or administrator API key") {
			t.Fatal("missing authentication guidance", v.Error)
		}
		if got := wakeView(t, f.request("GET", path, "", "admin"), 200); got != v {
			t.Fatal("GET lost pending state", got, v)
		}
		if got, err := f.db.GetNodeWakeConfig("remote"); err != nil || got != c {
			t.Fatal(got, err)
		}
	}
	if redirected.Load() != 0 {
		t.Fatal("followed redirect, leaking node token")
	}
	mode.Store(0)
	if v := wakeView(t, f.request("PUT", path, wakeJSON(c), "admin"), 200); !v.Applied || v.Config.Enabled {
		t.Fatal("disable retry failed", v)
	}
	remote.Close()
	if v := wakeView(t, f.request("PUT", path, wakeJSON(c), "admin"), 502); v.Applied || v.Config != c || v.Error == "" {
		t.Fatal(v)
	}
}

func TestNodeWakeConcurrentRemoteUpdatesSerialized(t *testing.T) {
	f := newReliabilityFixture(t)
	entered, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var c database.NodeWakeConfig
		_ = json.NewDecoder(r.Body).Decode(&c)
		if calls.Add(1) == 1 {
			close(entered)
			<-release
			writeJSON(w, 200, NodeWakeView{Config: c, Applied: true})
			return
		}
		w.WriteHeader(404) // newer disable must remain pending, not inherit success
	}))
	defer remote.Close()
	nodeNameRemote(t, f, remote.URL)
	path := "/api/admin/nodes/remote/wake"
	c := database.DefaultNodeWakeConfig()
	c.Enabled = true
	first, second := make(chan *httptest.ResponseRecorder, 1), make(chan *httptest.ResponseRecorder, 1)
	go func() { first <- f.request("PUT", path, wakeJSON(c), "admin") }()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("delivery never entered")
	}
	disabled := c
	disabled.Enabled = false
	go func() { second <- f.request("PUT", path, wakeJSON(disabled), "admin") }()
	time.Sleep(75 * time.Millisecond)
	count := calls.Load()
	close(release)
	if count != 1 {
		t.Fatal("concurrent delivery escaped serialization", count)
	}
	if v := wakeView(t, <-first, 200); !v.Applied {
		t.Fatal(v)
	}
	if v := wakeView(t, <-second, 502); v.Applied || v.Config != disabled {
		t.Fatal(v)
	}
	if v := wakeView(t, f.request("GET", path, "", "admin"), 200); v.Applied || v.Config != disabled || v.Error == "" {
		t.Fatal("stale success marked disable applied", v)
	}
}

func TestNodeWakeDeliveryHonorsContext(t *testing.T) {
	f := newReliabilityFixture(t)
	release := make(chan struct{})
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-release }))
	defer remote.Close()
	defer close(release)
	node := nodeNameRemote(t, f, remote.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	if err := f.router.nodes.deliverNodeWake(ctx, node, database.DefaultNodeWakeConfig()); err == nil {
		t.Fatal("cancelled delivery succeeded")
	}
	if time.Since(start) > time.Second {
		t.Fatal("ignored bounded request context")
	}
}

func TestNodeWakeUnicodeEscapes(t *testing.T) {
	c := database.DefaultNodeWakeConfig()
	c.VersionLabel = "Join to start"
	good := wakeJSON(c)
	for _, escaped := range []string{`\ud800`, `\udfff`, `\ud800x`, `\ud800\u0041`, `\ud800\ud800`} {
		body := strings.Replace(good, `"version_label":"Join to start"`, `"version_label":"`+escaped+`"`, 1)
		if _, err := decodeWakeConfig([]byte(body)); err == nil {
			t.Fatal("accepted invalid Unicode escape", escaped)
		}
	}
	for escaped, want := range map[string]string{`\ud83d\udca4`: "💤", `\\ud800`: `\ud800`, `\u00a7a\nSleep`: "§a\nSleep"} {
		body := strings.Replace(good, `"version_label":"Join to start"`, `"version_label":"`+escaped+`"`, 1)
		if c, err := decodeWakeConfig([]byte(body)); err != nil || c.VersionLabel != want {
			t.Fatal(escaped, c.VersionLabel, err)
		}
	}
}

func TestNodeWakeDeliveryFingerprintAndPendingSurviveRestart(t *testing.T) {
	f := newReliabilityFixture(t)
	node := nodeNameRemote(t, f, "http://127.0.0.1:1")
	c := database.DefaultNodeWakeConfig()
	if err := f.db.SetNodeWakeConfig(node.ID, c); err != nil {
		t.Fatal(err)
	}
	state := nodeWakeDelivery{Fingerprint: wakeFingerprint(c, node), Applied: true}
	if err := f.router.nodes.saveWakeDelivery(node.ID, state); err != nil {
		t.Fatal(err)
	}
	path := "/api/admin/nodes/remote/wake"
	if v := wakeView(t, f.request("GET", path, "", "admin"), 200); !v.Applied {
		t.Fatal(v)
	}
	// A changed destination/token is not covered by the old acknowledgement.
	node.APIToken = "changed-token"
	if err := f.db.UpdateNode(node); err != nil {
		t.Fatal(err)
	}
	if v := wakeView(t, f.request("GET", path, "", "admin"), 200); v.Applied || v.Error == "" {
		t.Fatal(v)
	}
	pending := wakeView(t, f.request("PUT", path, wakeJSON(c), "admin"), 502)
	db, err := database.InitDB(filepath.Join(f.cfg.Paths.DataDir, "panel.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := NewNodeHandler(f.cfg, db, f.router.auth, f.manager)
	req := httptest.NewRequest("GET", path, nil)
	req.SetPathValue("id", "remote")
	req = req.WithContext(context.WithValue(req.Context(), contextUserKey, &JWTClaims{Role: "admin"}))
	w := httptest.NewRecorder()
	h.GetNodeWake(w, req)
	if v := wakeView(t, w, 200); v != pending || v.Applied {
		t.Fatal(v, pending)
	}
}
