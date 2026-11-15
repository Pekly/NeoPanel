package daemon

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
)

func testConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.Server.Host, cfg.Server.Port = "127.0.0.1", 0
	cfg.SFTP.Enabled = false
	cfg.Paths.DataDir = t.TempDir()
	cfg.Paths.DBPath = filepath.Join(cfg.Paths.DataDir, "neopanel.db")
	cfg.Paths.ServersDir = filepath.Join(cfg.Paths.DataDir, "servers")
	cfg.FilePath = filepath.Join(cfg.Paths.DataDir, "config.yaml")
	return cfg
}

func startTestService(t *testing.T, cfg *config.Config) *Service {
	t.Helper()
	s, err := Start(cfg, fstest.MapFS{"index.html": {Data: []byte("<html>daemon-test-panel</html>")}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = s.Shutdown(ctx)
	})
	return s
}

func get(t *testing.T, target string) (int, string) {
	t.Helper()
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(target)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, string(body)
}

func TestStartEphemeralPortAndRoutes(t *testing.T) {
	for _, role := range []string{"panel", "node"} {
		t.Run(role, func(t *testing.T) {
			cfg := testConfig(t)
			cfg.Server.Role = role
			s := startTestService(t, cfg)
			u, err := url.Parse(s.URL())
			if err != nil || u.Hostname() != "127.0.0.1" || u.Port() == "0" || u.Port() == "" {
				t.Fatalf("not an actual loopback URL: %q (%v)", s.URL(), err)
			}
			if s.http.ReadTimeout != 30*time.Second || s.http.WriteTimeout != 60*time.Second || s.http.IdleTimeout != 120*time.Second {
				t.Fatal("HTTP timeouts changed")
			}
			if status, _ := get(t, s.URL()+"/api/auth/status"); status != http.StatusOK {
				t.Fatalf("auth status API: %d", status)
			}
			if status, _ := get(t, s.URL()+"/api/servers"); status != http.StatusUnauthorized {
				t.Fatalf("server API must still require authentication: %d", status)
			}
			_, body := get(t, s.URL()+"/")
			if strings.Contains(body, "daemon-test-panel") != (role == "panel") {
				t.Fatalf("wrong UI for %s: %s", role, body)
			}
			if role == "node" && !strings.Contains(body, `"role":"node"`) {
				t.Fatalf("node metadata absent: %s", body)
			}
			resp, err := http.Post(s.URL()+"/api/shutdown", "application/json", nil)
			if err != nil {
				t.Fatal(err)
			}
			resp.Body.Close()
			if resp.StatusCode < 400 {
				t.Fatalf("unexpected shutdown endpoint: %d", resp.StatusCode)
			}
			select {
			case <-s.Done():
				t.Fatal("running service reported done")
			default:
			}
			if s.ActiveServerCount() != 0 {
				t.Fatal("empty service has active servers")
			}
			if err := s.Shutdown(context.Background()); err != nil {
				t.Fatal(err)
			}
			select {
			case <-s.Done():
			default:
				t.Fatal("shutdown returned before serving stopped")
			}
			if s.Err() != nil || s.ActiveServerCount() != 0 {
				t.Fatalf("normal shutdown: %v", s.Err())
			}
			if _, err := s.db.ListServers(); err == nil {
				t.Fatal("database still open after Shutdown")
			}
			conn, err := net.DialTimeout("tcp", u.Host, time.Second)
			if err == nil {
				conn.Close()
				t.Fatal("listener still open after Shutdown")
			}
		})
	}
}

func TestConcurrentAndRepeatedShutdown(t *testing.T) {
	s := startTestService(t, testConfig(t))
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			if err := s.Shutdown(context.Background()); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if err := s.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestBindFailure(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	cfg := testConfig(t)
	cfg.Server.Port = listener.Addr().(*net.TCPAddr).Port
	s, err := Start(cfg, nil)
	if err == nil || s != nil {
		t.Fatalf("bind failure: service=%v error=%v", s, err)
	}
	if _, err := os.Stat(cfg.Paths.DBPath); !os.IsNotExist(err) {
		t.Fatalf("bind failure should not initialize database: %v", err)
	}
}

func TestPartialStartFailureReleasesListener(t *testing.T) {
	cfg := testConfig(t)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	cfg.Server.Port = listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	// A regular file cannot contain a servers directory. This fails after DB init.
	blocker := filepath.Join(cfg.Paths.DataDir, "blocker")
	if err := os.WriteFile(blocker, nil, 0600); err != nil {
		t.Fatal(err)
	}
	cfg.Paths.ServersDir = filepath.Join(blocker, "servers")
	if s, err := Start(cfg, nil); err == nil || s != nil {
		t.Fatalf("expected partial start failure: %v, %v", s, err)
	}
	rebound, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("listener leaked: %v", err)
	}
	rebound.Close()
	if err := os.Remove(cfg.Paths.DBPath); err != nil {
		t.Fatalf("database handle leaked: %v", err)
	}
}

func TestUnexpectedServeFailure(t *testing.T) {
	s := startTestService(t, testConfig(t))
	get(t, s.URL()+"/api/auth/status")
	if err := s.listener.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-s.Done():
	case <-time.After(3 * time.Second):
		t.Fatal("Serve failure did not close Done")
	}
	if !errors.Is(s.Err(), net.ErrClosed) {
		t.Fatalf("underlying Serve error not retained: %v", s.Err())
	}
	select {
	case <-s.cleaned:
	case <-time.After(3 * time.Second):
		t.Fatal("unexpected Serve failure did not clean up automatically")
	}
}

func TestShutdownJoinsHandlersAndRejectsNewRequests(t *testing.T) {
	s := startTestService(t, testConfig(t))
	entered, release, handled := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	releaseHandler := func() { releaseOnce.Do(func() { close(release) }) }
	defer releaseHandler()
	go func() {
		defer close(handled)
		s.guard(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			close(entered)
			<-release
		})).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	}()
	<-entered
	stopped := make(chan error, 1)
	go func() { stopped <- s.Shutdown(context.Background()) }()
	<-s.Done()
	w := httptest.NewRecorder()
	s.guard(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("request admitted during shutdown")
	})).ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("shutdown admission status=%d", w.Code)
	}
	select {
	case <-stopped:
		t.Fatal("shutdown failed to join request handler")
	default:
	}
	if _, err := s.db.ListServers(); err != nil {
		t.Fatalf("database closed while request is active: %v", err)
	}
	releaseHandler()
	<-handled
	if err := <-stopped; err != nil {
		t.Fatal(err)
	}
}

// The test binary acts as a Java stand-in through the real runner launch path.
func init() {
	if os.Getenv("NEOPANEL_DAEMON_TEST_PROCESS") != "1" {
		return
	}
	wait := func(name string) {
		for range 6000 {
			if _, err := os.Stat(name); err == nil {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
		os.Exit(2)
	}
	wait("ready")
	fmt.Println(`Done (0.01s)! For help, type "help"`)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "stop" {
			wait("exit")
			os.Exit(0)
		}
	}
	os.Exit(0)
}

func TestActiveCountAndShutdownWaitsForProcess(t *testing.T) {
	t.Setenv("NEOPANEL_DAEMON_TEST_PROCESS", "1")
	cfg := testConfig(t)
	s := startTestService(t, cfg)
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	server := &database.Server{ID: "daemon-test", Name: "test", Type: "paper", JarFile: "server.jar", JavaPath: exe, NodeID: "local"}
	if err := s.db.CreateServer(server); err != nil {
		t.Fatal(err)
	}
	node := &database.ClusterNode{ID: "local", Name: "Local", IsolationEngine: "process"}
	if err := s.db.CreateNode(node); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(cfg.Paths.ServersDir, server.ID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	touch := func(name string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	touch("server.jar")
	t.Cleanup(func() { touch("ready"); touch("exit") })
	if err := s.manager.Start(server); err != nil {
		t.Fatal(err)
	}
	waitStatus := func(want string) {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for s.manager.GetStatus(server.ID) != want && time.Now().Before(deadline) {
			time.Sleep(10 * time.Millisecond)
		}
		if got := s.manager.GetStatus(server.ID); got != want {
			t.Fatalf("status=%s want=%s logs=%v", got, want, s.manager.GetLogs(server.ID))
		}
		if s.ActiveServerCount() != 1 {
			t.Fatalf("%s server not counted as active", want)
		}
	}
	waitStatus("starting")
	touch("ready")
	waitStatus("online")
	stopped := make(chan error, 1)
	go func() { stopped <- s.Shutdown(context.Background()) }()
	waitStatus("stopping")
	select {
	case <-stopped:
		t.Fatal("Shutdown returned before process exit")
	default:
	}
	if _, err := s.db.ListServers(); err != nil {
		t.Fatalf("DB closed before process exit: %v", err)
	}
	touch("exit")
	select {
	case err := <-stopped:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Shutdown did not finish after process exit")
	}
	if s.ActiveServerCount() != 0 || s.manager.GetStatus(server.ID) != "offline" {
		t.Fatal("process still active after Shutdown")
	}
}
