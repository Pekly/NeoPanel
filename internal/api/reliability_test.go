package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/runner"
)

type reliabilityFixture struct {
	router  *Router
	db      *database.DB
	manager *runner.ServerManager
	cfg     *config.Config
	tokens  map[string]string
	users   map[string]*database.User
}

func newReliabilityFixture(t *testing.T) *reliabilityFixture {
	t.Helper()
	dir := t.TempDir()
	db, err := database.InitDB(filepath.Join(dir, "panel.db"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{}
	cfg.Paths.DataDir = dir
	cfg.Paths.ServersDir = filepath.Join(dir, "servers")
	cfg.Security.JWTSecret = "regression-test-secret-" + dir
	cfg.Security.TokenExpiration = 1
	cfg.Security.NodeToken = "neo_node_regression"
	manager := runner.NewServerManager(cfg.Paths.ServersDir, db)
	router := NewRouter(cfg, db, manager, nil)
	router.globalLimiter = nil
	f := &reliabilityFixture{router, db, manager, cfg, map[string]string{}, map[string]*database.User{}}
	for _, name := range []string{"admin", "owner", "stranger", "subuser"} {
		role := "user"
		if name == "admin" {
			role = "admin"
		}
		user, err := db.CreateUser(name, "unused", role)
		if err != nil {
			t.Fatal(err)
		}
		f.users[name] = user
		f.tokens[name], err = router.auth.GenerateToken(user)
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"alpha", "beta"} {
		srv := &database.Server{ID: id, Name: id, OwnerID: f.users["owner"].ID, NodeID: "local", Type: "paper", Version: "1.21", JarFile: "server.jar", Description: "keep", JavaArgs: "-Dkeep=true", JavaPath: "java-custom", AutoRestart: true, CPULimitPercent: 120, DiskLimitMB: 4096}
		if err := db.CreateServer(srv); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(manager.GetWorkingDir(id), 0755); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.AddServerSubuser("alpha", f.users["subuser"].ID, []string{"files.view", "console.view"}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { manager.ShutdownAll(); db.Close() })
	return f
}

func (f *reliabilityFixture) request(method, path, body, user string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if user != "" {
		req.Header.Set("Authorization", "Bearer "+f.tokens[user])
	}
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	return w
}

func TestOperationalRoutesRequireMembershipAndAction(t *testing.T) {
	f := newReliabilityFixture(t)
	routes := []string{
		"GET ", "PUT ", "DELETE ", "POST power", "POST command", "GET icon", "POST icon", "DELETE icon",
		"GET configs", "POST configs", "GET subusers", "POST subusers", "PUT subusers/1", "DELETE subusers/1", "GET permissions",
		"GET schedules", "POST schedules", "PUT schedules/missing", "DELETE schedules/missing", "POST schedules/missing/toggle", "POST schedules/missing/run",
		"GET ports", "POST ports", "DELETE ports/1", "POST ports/1/primary", "GET ws", "GET files", "GET files/content", "POST files/content", "POST files/create", "DELETE files", "POST files/upload", "GET files/download", "POST files/rename", "POST files/move", "POST files/compress", "POST files/archive", "POST files/unzip", "POST files/extract",
		"GET plugins", "POST plugins/install", "DELETE plugins", "POST plugins/toggle", "GET plugins/versions", "GET backups", "POST backups", "DELETE backups", "GET backups/download", "GET players", "POST players/action", "POST software", "POST reinstall", "GET properties", "POST properties", "GET databases", "POST databases", "DELETE databases/missing", "POST databases/missing/rotate", "GET databases/missing/tables", "POST databases/missing/query", "GET databases/missing/export", "GET subdomain", "PUT subdomain", "POST subdomain/verify", "GET sftp",
	}
	for _, route := range routes {
		parts := strings.SplitN(route, " ", 2)
		path := "/api/servers/alpha"
		if parts[1] != "" {
			path += "/" + parts[1]
		}
		for _, user := range []string{"", "stranger"} {
			want := 403
			if user == "" {
				want = 401
			}
			if w := f.request(parts[0], path, "{}", user); w.Code != want {
				t.Errorf("%s %s: got %d, want %d: %s", user, route, w.Code, want, w.Body.String())
			}
		}
	}
	if w := f.request("GET", "/api/servers/alpha/files", "", "subuser"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, route := range []string{"POST files/content", "POST command", "GET databases", "POST plugins/toggle"} {
		p := strings.SplitN(route, " ", 2)
		if w := f.request(p[0], "/api/servers/alpha/"+p[1], "{}", "subuser"); w.Code != 403 {
			t.Errorf("subuser %s: %d", route, w.Code)
		}
	}
	for _, user := range []string{"owner", "admin"} {
		if w := f.request("GET", "/api/servers/alpha/databases", "", user); w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
}

func TestDatabaseDuplicatesPreserveData(t *testing.T) {
	f := newReliabilityFixture(t)
	create := func(name string) *httptest.ResponseRecorder {
		return f.request("POST", "/api/servers/alpha/databases", `{"name":"`+name+`"}`, "owner")
	}
	if w := create("data"); w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	dbPath := filepath.Join(f.cfg.Paths.DataDir, "databases", "alpha", "data.db")
	conn, err := f.router.databases.openServerSQLite(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(`CREATE TABLE valuable (value TEXT); INSERT INTO valuable VALUES ('keep me')`); err != nil {
		t.Fatal(err)
	}
	conn.Close()
	var wg sync.WaitGroup
	codes := make(chan int, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); codes <- create("data").Code }()
	}
	wg.Wait()
	close(codes)
	for code := range codes {
		if code != 409 {
			t.Errorf("duplicate returned %d", code)
		}
	}
	conn, err = f.router.databases.openServerSQLite(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	var value string
	if err := conn.QueryRow(`SELECT value FROM valuable`).Scan(&value); err != nil || value != "keep me" {
		t.Fatalf("existing data lost: %q %v", value, err)
	}
	orphan := filepath.Join(filepath.Dir(dbPath), "orphan.db")
	if err := os.WriteFile(orphan, []byte("unregistered data"), 0600); err != nil {
		t.Fatal(err)
	}
	if w := create("orphan"); w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	b, _ := os.ReadFile(orphan)
	if string(b) != "unregistered data" {
		t.Fatal("existing orphan overwritten")
	}
	codes = make(chan int, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); codes <- create("fresh").Code }()
	}
	wg.Wait()
	close(codes)
	created := 0
	for code := range codes {
		if code == 201 {
			created++
		} else if code != 409 {
			t.Errorf("concurrent create: %d", code)
		}
	}
	if created != 1 {
		t.Fatalf("created %d databases", created)
	}
}

func TestDestructiveRootAliases(t *testing.T) {
	f := newReliabilityFixture(t)
	root := f.manager.GetWorkingDir("alpha")
	os.MkdirAll(filepath.Join(root, "folder"), 0755)
	os.WriteFile(filepath.Join(root, "keep.txt"), []byte("keep"), 0600)
	for _, alias := range []string{"", "/", ".", "./", "folder/..", "folder/../.", "folder\\.."} {
		encoded, _ := json.Marshal(alias)
		cases := []struct{ method, path, body string }{
			{"DELETE", "/api/servers/alpha/files?path=" + url.QueryEscape(alias), ""},
			{"POST", "/api/servers/alpha/files/rename", `{"old_path":` + string(encoded) + `,"new_path":"renamed"}`},
			{"POST", "/api/servers/alpha/files/move", `{"paths":[` + string(encoded) + `],"destination":"moved"}`},
		}
		for _, c := range cases {
			if w := f.request(c.method, c.path, c.body, "owner"); w.Code != 400 {
				t.Errorf("%s %q: %d %s", c.path, alias, w.Code, w.Body.String())
			}
		}
	}
	if _, err := os.Stat(filepath.Join(root, "keep.txt")); err != nil {
		t.Fatal("server root destroyed", err)
	}
}

func TestScopedResourcesAndSchedulePartialUpdates(t *testing.T) {
	f := newReliabilityFixture(t)
	s := &database.ServerSchedule{ID: "schedule", ServerID: "beta", Name: "keep", CronExpr: "*/5 * * * *", Action: "command", Payload: "say keep", IsActive: true}
	if err := f.db.CreateServerSchedule(s); err != nil {
		t.Fatal(err)
	}
	for _, route := range []string{"PUT schedule", "DELETE schedule", "POST schedule/toggle", "POST schedule/run"} {
		p := strings.SplitN(route, " ", 2)
		if w := f.request(p[0], "/api/servers/alpha/schedules/"+p[1], `{"is_active":false}`, "owner"); w.Code != 404 {
			t.Errorf("%s: %d %s", route, w.Code, w.Body.String())
		}
	}
	su, err := f.db.AddServerSubuser("beta", f.users["stranger"].ID, []string{"files.view"})
	if err != nil {
		t.Fatal(err)
	}
	for _, method := range []string{"PUT", "DELETE"} {
		w := f.request(method, fmt.Sprintf("/api/servers/alpha/subusers/%d", su.ID), `{"permissions":["files.delete"]}`, "owner")
		if w.Code != 404 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if w := f.request("PUT", "/api/servers/beta/schedules/schedule", `{"name":"renamed"}`, "owner"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	got, _ := f.db.GetServerSchedule("beta", "schedule")
	if got.Payload != "say keep" || !got.IsActive || got.CronExpr != s.CronExpr || got.Action != s.Action {
		t.Fatalf("partial update cleared fields: %+v", got)
	}
	for _, body := range []string{`{"cron_expr":"60 * * * *"}`, `{"cron_expr":"*/0 * * * *"}`, `{"cron_expr":""}`} {
		if w := f.request("PUT", "/api/servers/beta/schedules/schedule", body, "owner"); w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if w := f.request("POST", "/api/servers/alpha/schedules", `{"name":"bad","cron_expr":"nonsense","action":"backup"}`, "owner"); w.Code != 400 {
		t.Fatal(w.Code, w.Body.String())
	}
	f.db.SetScheduleRunStatus("schedule", "failed", "test failure")
	w := f.request("GET", "/api/servers/beta/schedules", "", "owner")
	var result struct {
		Schedules []map[string]interface{} `json:"schedules"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	item := result.Schedules[0]
	if item["timezone"] != "UTC" || item["last_status"] != "failed" || item["last_error"] != "test failure" {
		t.Fatal(item)
	}
	if _, err := time.Parse(time.RFC3339, item["next_run_at"].(string)); err != nil {
		t.Fatal(err)
	}
	if _, ok := item["last_run"]; !ok {
		t.Fatal("last_run alias missing")
	}
}

func TestServerOptionalUpdatePreservesFields(t *testing.T) {
	f := newReliabilityFixture(t)
	w := f.request("PUT", "/api/servers/alpha", `{"name":"renamed"}`, "admin")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	s, _ := f.db.GetServer("alpha")
	if s.Description != "keep" || s.JavaArgs != "-Dkeep=true" || s.JavaPath != "java-custom" || !s.AutoRestart || s.CPULimitPercent != 120 || s.DiskLimitMB != 4096 {
		t.Fatalf("fields cleared: %+v", s)
	}
	w = f.request("PUT", "/api/servers/alpha", `{"description":"","java_args":"","java_path":"","auto_restart":false,"cpu_limit_percent":0,"disk_limit_mb":0}`, "admin")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	s, _ = f.db.GetServer("alpha")
	if s.Description != "" || s.JavaArgs != "" || s.JavaPath != "" || s.AutoRestart || s.CPULimitPercent != 0 || s.DiskLimitMB != 0 {
		t.Fatalf("explicit zeros ignored: %+v", s)
	}
}

func TestManualBackupCollisionAndSize(t *testing.T) {
	f := newReliabilityFixture(t)
	os.WriteFile(filepath.Join(f.manager.GetWorkingDir("alpha"), "world"), []byte(strings.Repeat("world", 500)), 0600)
	w := f.request("POST", "/api/servers/alpha/backups", `{"name":"same"}`, "owner")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var result struct {
		Size int64  `json:"size"`
		Name string `json:"name"`
	}
	json.Unmarshal(w.Body.Bytes(), &result)
	p := filepath.Join(f.cfg.Paths.DataDir, "backups", "alpha", result.Name)
	st, err := os.Stat(p)
	if err != nil || st.Size() != result.Size {
		t.Fatal(result, err)
	}
	before, _ := os.ReadFile(p)
	if w := f.request("POST", "/api/servers/alpha/backups", `{"name":"same"}`, "owner"); w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	after, _ := os.ReadFile(p)
	if string(after) != string(before) {
		t.Fatal("backup overwritten")
	}
	os.RemoveAll(f.manager.GetWorkingDir("beta"))
	if w := f.request("POST", "/api/servers/beta/backups", `{}`, "owner"); w.Code != 500 {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestProxyPreservesCallerAndRejectsOldNodes(t *testing.T) {
	master := newReliabilityFixture(t)
	node := newReliabilityFixture(t)
	nodeHTTP := httptest.NewServer(node.router)
	defer nodeHTTP.Close()
	if err := master.db.CreateNode(&database.ClusterNode{ID: "remote", Name: "remote", BaseURL: nodeHTTP.URL, APIToken: node.cfg.Security.NodeToken}); err != nil {
		t.Fatal(err)
	}
	for _, user := range []string{"owner", "subuser", "stranger"} {
		want := 403
		if user != "stranger" {
			want = 200
		}
		w := master.request("GET", "/api/nodes/remote/proxy/api/servers/alpha/files", "", user)
		if w.Code != want {
			t.Errorf("proxy %s: %d %s", user, w.Code, w.Body.String())
		}
	}
	for _, user := range []string{"stranger", "subuser"} {
		w := master.request("GET", "/api/cluster/servers", "", user)
		var result struct {
			Servers []database.Server `json:"servers"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		want := 0
		if user == "subuser" {
			want = 2
		} // alpha on each node, never beta
		if w.Code != 200 || len(result.Servers) != want {
			t.Fatalf("cluster leak for %s: %s", user, w.Body.String())
		}
		for _, srv := range result.Servers {
			if srv.ID != "alpha" {
				t.Fatal("unrelated server exposed", srv.ID)
			}
		}
	}
	if w := master.request("POST", "/api/nodes/remote/proxy/api/servers/alpha/command", `{"command":"stop"}`, "subuser"); w.Code != 403 {
		t.Fatal(w.Code, w.Body.String())
	}
	forged := httptest.NewRequest("GET", "/api/servers/alpha/files", nil)
	forged.Header.Set("Authorization", "Bearer "+node.tokens["stranger"])
	forged.Header.Set("X-NeoPanel-User-ID", fmt.Sprint(node.users["owner"].ID))
	forged.Header.Set("X-NeoPanel-User-Role", "admin")
	w := httptest.NewRecorder()
	node.router.ServeHTTP(w, forged)
	if w.Code != 403 {
		t.Fatal("forged forwarded identity accepted", w.Code)
	}
	old := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/auth/me" {
			t.Error("operation sent before identity check")
		}
		w.Write([]byte(`{}`))
	}))
	defer old.Close()
	master.db.CreateNode(&database.ClusterNode{ID: "old", Name: "old", BaseURL: old.URL, APIToken: "old-admin-token"})
	if w := master.request("DELETE", "/api/nodes/old/proxy/api/servers/alpha/files?path=world", "", "owner"); w.Code != 502 {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestEncodedServerIDCannotBypassAuthorization(t *testing.T) {
	f := newReliabilityFixture(t)
	for _, method := range []string{"GET", "DELETE"} {
		w := f.request(method, "/api/servers/alpha%2F..%2Fbeta/files?path=world", "", "subuser")
		if w.Code != 404 {
			t.Fatalf("encoded server ID: %d %s", w.Code, w.Body.String())
		}
	}
}

func TestManualScheduleAcceptanceAndOverlapResponse(t *testing.T) {
	f := newReliabilityFixture(t)
	s := &database.ServerSchedule{ID: "run", ServerID: "alpha", Name: "run", CronExpr: "* * * * *", Action: "backup", IsActive: false}
	if err := f.db.CreateServerSchedule(s); err != nil {
		t.Fatal(err)
	}
	if ok, err := f.db.ClaimScheduleRun(s.ID, time.Now().UTC(), false); err != nil || !ok {
		t.Fatal(ok, err)
	}
	if w := f.request("POST", "/api/servers/alpha/schedules/run/run", `{}`, "owner"); w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	if err := f.db.SetScheduleRunStatus(s.ID, "failed", "previous run"); err != nil {
		t.Fatal(err)
	}
	w := f.request("POST", "/api/servers/alpha/schedules/run/run", `{}`, "owner")
	if w.Code != 202 || !strings.Contains(w.Body.String(), `"status":"started"`) {
		t.Fatal(w.Code, w.Body.String())
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		got, err := f.db.GetServerSchedule("alpha", s.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.LastStatus == "succeeded" {
			if viewSchedule(got).NextRunAt != nil {
				t.Fatal("disabled schedule has next run")
			}
			return
		}
		if got.LastStatus == "failed" {
			t.Fatal(got.LastError)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("manual backup did not finish")
}
