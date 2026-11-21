package daemon

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func waitRecoveredStatus(t *testing.T, s *Service, id, want string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if s.manager.GetStatus(id) == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("server %s status=%s want=%s logs=%v", id, s.manager.GetStatus(id), want, s.manager.GetLogs(id))
}

func TestServiceRestartRestoresOnlineServers(t *testing.T) {
	t.Setenv("NEOPANEL_DAEMON_TEST_PROCESS", "1")
	for _, role := range []string{"panel", "node"} {
		t.Run(role, func(t *testing.T) {
			cfg := testConfig(t)
			cfg.Server.Role = role
			first := startTestService(t, cfg)
			exe, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			online := &database.Server{ID: "resume-online", Name: "Online", Type: "paper", JarFile: "server.jar", JavaPath: exe, NodeID: "local", AutoRestart: false}
			offline := &database.Server{ID: "resume-offline", Name: "Offline", Type: "paper", JarFile: "server.jar", JavaPath: exe, NodeID: "local", AutoRestart: true}
			for _, srv := range []*database.Server{online, offline} {
				if err := first.db.CreateServer(srv); err != nil {
					t.Fatal(err)
				}
				dir := filepath.Join(cfg.Paths.ServersDir, srv.ID)
				if err := os.MkdirAll(dir, 0700); err != nil {
					t.Fatal(err)
				}
				for _, name := range []string{"server.jar", "ready", "exit"} {
					if err := os.WriteFile(filepath.Join(dir, name), nil, 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			if err := first.db.CreateNode(&database.ClusterNode{ID: "local", Name: "Local", IsolationEngine: "process"}); err != nil {
				t.Fatal(err)
			}
			if err := first.manager.Start(online); err != nil {
				t.Fatal(err)
			}
			waitRecoveredStatus(t, first, online.ID, "online")
			if err := first.Shutdown(context.Background()); err != nil {
				t.Fatal(err)
			}
			// Repeated service shutdown must not erase the recorded running state.
			if err := first.Shutdown(context.Background()); err != nil {
				t.Fatal(err)
			}
			next := startTestService(t, cfg)
			waitRecoveredStatus(t, next, online.ID, "online")
			if next.manager.GetStatus(offline.ID) != "offline" || next.ActiveServerCount() != 1 {
				t.Fatal("restored the wrong server set")
			}
			if err := next.manager.RestoreRunningServers(); err != nil {
				t.Fatal(err)
			}
			if next.ActiveServerCount() != 1 {
				t.Fatal("duplicate restore")
			}
			if err := next.manager.Stop(online.ID, true); err != nil {
				t.Fatal(err)
			}
			if err := next.Shutdown(context.Background()); err != nil {
				t.Fatal(err)
			}
			last := startTestService(t, cfg)
			if last.ActiveServerCount() != 0 {
				t.Fatal("manual stop was undone")
			}
		})
	}
}

func TestServiceBindFailurePreservesResumeSnapshot(t *testing.T) {
	cfg := testConfig(t)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	cfg.Server.Port = listener.Addr().(*net.TCPAddr).Port
	db, err := database.InitDB(cfg.Paths.DBPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.SaveServerResumeState(cfg.Paths.ServersDir, []string{"previously-online"}); err != nil {
		t.Fatal(err)
	}
	service, err := Start(cfg, nil)
	if err == nil || service != nil {
		if service != nil {
			service.Shutdown(context.Background())
		}
		t.Fatal("occupied listener accepted")
	}
	ids, err := db.TakeServerResumeState(cfg.Paths.ServersDir)
	if err != nil || !reflect.DeepEqual(ids, []string{"previously-online"}) {
		t.Fatal("bind failure consumed recovery", ids, err)
	}
}
