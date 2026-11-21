package api

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
)

func wakeStatusView(t *testing.T, w *httptest.ResponseRecorder) runner.NodeWakeStatus {
	t.Helper()
	var v runner.NodeWakeStatus
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil || w.Code != 200 {
		t.Fatal(w.Code, w.Body.String(), err)
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("cacheable live status")
	}
	return v
}

func TestNodeWakeStatusAuthorizationAndReadOnly(t *testing.T) {
	f := newReliabilityFixture(t)
	for _, path := range []string{"/api/admin/nodes/local/wake/status", "/api/cluster/wake/status"} {
		for _, user := range []string{"", "owner", "stranger", "subuser"} {
			want := 403
			if user == "" {
				want = 401
			}
			if w := f.request("GET", path, "", user); w.Code != want {
				t.Fatal(path, user, w.Code)
			}
		}
		v := wakeStatusView(t, f.request("GET", path, "", "admin"))
		if v.CoordinatorRunning || v.ConfigMatches || !v.CheckedAt.IsZero() || v.Listeners != 0 {
			t.Fatal(v)
		}
	}
	for _, handler := range []http.HandlerFunc{f.router.nodes.GetNodeWakeStatus, f.router.nodes.GetClusterWakeStatus} {
		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest("GET", "/", nil))
		if w.Code != 403 {
			t.Fatal(w.Code)
		}
	}
	if w := f.request("GET", "/api/admin/nodes/missing/wake/status", "", "admin"); w.Code != 404 {
		t.Fatal(w.Code)
	}
}

func TestNodeWakeStatusRemoteRealWorkerAndMismatch(t *testing.T) {
	master, worker := newReliabilityFixture(t), newReliabilityFixture(t)
	worker.cfg.Server.Role = "node"
	host := httptest.NewServer(worker.router)
	defer host.Close()
	node := nodeNameRemote(t, master, host.URL)
	worker.cfg.Security.NodeToken = node.APIToken
	path := "/api/admin/nodes/remote/wake/status"
	v := wakeStatusView(t, master.request("GET", path, "", "admin"))
	if v.CoordinatorRunning || v.ConfigMatches || !v.CheckedAt.IsZero() {
		t.Fatal(v)
	}
	cfg := database.DefaultNodeWakeConfig()
	cfg.Enabled = true
	cfg.BindAddress = "127.0.0.1"
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	s, _ := worker.db.GetServer("alpha")
	s.Port = port
	if err := worker.db.UpdateServer(s); err != nil {
		t.Fatal(err)
	}
	if err := worker.db.SetNodeWakeConfig("local", cfg); err != nil {
		t.Fatal(err)
	}
	if err := master.db.SetNodeWakeConfig("remote", cfg); err != nil {
		t.Fatal(err)
	}
	if err := worker.db.SetNodeIdlePolicy("local", database.NodeIdlePolicy{Enabled: true, TimeoutMinutes: 15}); err != nil {
		t.Fatal(err)
	}
	if err := worker.db.SetServerIdleStopped(s.ID, true); err != nil {
		t.Fatal(err)
	}
	worker.manager.StartWakeListeners()
	deadline := time.Now().Add(5 * time.Second)
	for {
		v = wakeStatusView(t, master.request("GET", path, "", "admin"))
		if v.Listeners == 1 && v.ConfigMatches {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal(v)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !v.CoordinatorRunning || v.ConfigHash != runner.NodeWakeConfigHash(cfg) {
		t.Fatal(v)
	}
	found := false
	for _, row := range v.Servers {
		if row.ID == "alpha" && row.State == "listening" {
			found = true
		}
	}
	if !found {
		t.Fatal(v)
	}
	cfg.VersionLabel = "New desired configuration"
	if err := master.db.SetNodeWakeConfig("remote", cfg); err != nil {
		t.Fatal(err)
	}
	v = wakeStatusView(t, master.request("GET", path, "", "admin"))
	if v.ConfigMatches || v.Error == "" {
		t.Fatal("remote falsely matches master desired", v)
	}
	for _, row := range v.Servers {
		if row.State == "listening" {
			t.Fatal(v)
		}
	}
	worker.cfg.Security.NodeToken = "wrong-token"
	if w := master.request("GET", path, "", "admin"); w.Code != 502 {
		t.Fatal(w.Code, w.Body.String())
	}
}

func goodWakeStatus() runner.NodeWakeStatus {
	cfg := database.DefaultNodeWakeConfig()
	cfg.Enabled = true
	return runner.NodeWakeStatus{CheckedAt: time.Now().UTC(), Enabled: true, CoordinatorRunning: true, ConfigMatches: true, ConfigHash: runner.NodeWakeConfigHash(cfg), Listeners: 1, Servers: []runner.WakeServerStatus{{ID: "alpha", Name: "Alpha", Port: 25565, State: "listening", Detail: "Listener bound"}}}
}

func TestNodeWakeStatusRejectsInvalidReports(t *testing.T) {
	good, _ := json.Marshal(goodWakeStatus())
	if _, err := decodeNodeWakeStatus(good, time.Now()); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*runner.NodeWakeStatus){
		func(v *runner.NodeWakeStatus) { v.CheckedAt = time.Now().Add(-11 * time.Second) },
		func(v *runner.NodeWakeStatus) { v.CheckedAt = time.Now().Add(time.Minute) },
		func(v *runner.NodeWakeStatus) { v.ConfigHash = "not-a-hash" },
		func(v *runner.NodeWakeStatus) { v.Listeners = -1 },
		func(v *runner.NodeWakeStatus) { v.Listeners = 0 },
		func(v *runner.NodeWakeStatus) { v.CoordinatorRunning = false },
		func(v *runner.NodeWakeStatus) { v.Servers[0].State = "ready" },
		func(v *runner.NodeWakeStatus) { v.Servers = append(v.Servers, v.Servers[0]) },
	} {
		v := goodWakeStatus()
		mutate(&v)
		body, _ := json.Marshal(v)
		if _, err := decodeNodeWakeStatus(body, time.Now()); err == nil {
			t.Fatal("accepted", string(body))
		}
	}
	for _, body := range []string{"{}", "null", string(good) + " {}", strings.Replace(string(good), `"enabled":true`, `"enabled":null`, 1), strings.Replace(string(good), `"enabled":true`, `"enabled":true,"enabled":true`, 1), strings.Repeat(" ", nodeWakeStatusLimit+1)} {
		if _, err := decodeNodeWakeStatus([]byte(body), time.Now()); err == nil {
			t.Fatal("accepted malformed report")
		}
	}
}

func TestNodeWakeStatusRemoteFailuresBoundedAndSafe(t *testing.T) {
	f := newReliabilityFixture(t)
	var mode atomic.Int32
	var redirects atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { redirects.Add(1) }))
	defer target.Close()
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/cluster/wake/status" || r.Header.Get("Authorization") != "Bearer secret-token-preserved" {
			t.Error("wrong route or machine auth")
		}
		switch mode.Load() {
		case 0:
			w.WriteHeader(404)
		case 1:
			w.WriteHeader(405)
		case 2:
			http.Redirect(w, r, target.URL, 302)
		case 3:
			_, _ = w.Write([]byte(strings.Repeat("x", nodeWakeStatusLimit+1)))
		case 4:
			v := goodWakeStatus()
			v.CheckedAt = time.Now().Add(-time.Minute)
			_ = json.NewEncoder(w).Encode(v)
		case 5:
			v := goodWakeStatus()
			v.Error = `secret-token-preserved C:\private\file`
			v.Servers[0].Detail = v.Error
			_ = json.NewEncoder(w).Encode(v)
		case 6:
			<-r.Context().Done()
		}
	}))
	defer remote.Close()
	node := nodeNameRemote(t, f, remote.URL)
	path := "/api/admin/nodes/remote/wake/status"
	for i := int32(0); i < 5; i++ {
		mode.Store(i)
		if w := f.request("GET", path, "", "admin"); w.Code != 502 {
			t.Fatal(i, w.Code, w.Body.String())
		}
	}
	if redirects.Load() != 0 {
		t.Fatal("redirect leaked credentials")
	}
	mode.Store(5)
	v, err := fetchNodeWakeStatus(context.Background(), node)
	if err != nil || strings.Contains(v.Error, "secret-token") || strings.Contains(v.Servers[0].Detail, "private") {
		t.Fatal(v, err)
	}
	mode.Store(6)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err = fetchNodeWakeStatus(ctx, node)
	if err == nil || time.Since(start) > time.Second {
		t.Fatal("context bound ignored", err)
	}
	remote.Close()
	if w := f.request("GET", path, "", "admin"); w.Code != 502 || strings.Contains(w.Body.String(), node.APIToken) {
		t.Fatal(w.Code, w.Body.String())
	}
}
