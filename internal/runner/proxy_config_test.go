package runner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/proxyconfig"
)

func TestProxyConfigOperationNeverStopsRunningServer(t *testing.T) {
	for _, status := range []string{"online", "starting", "stopping"} {
		t.Run(status, func(t *testing.T) {
			m, srv, _ := restoreFixture(t)
			inst := NewProcessInstance(srv, m.GetWorkingDir(srv.ID), DefaultStartupTimeout, nil, nil)
			m.mu.Lock()
			m.instances[srv.ID] = inst
			m.mu.Unlock()
			inst.mu.Lock()
			inst.status = status
			inst.mu.Unlock()
			defer func() { inst.mu.Lock(); inst.status = "offline"; inst.mu.Unlock() }()
			if release, err := m.BeginProxyConfigOperation(context.Background(), srv.ID); !errors.Is(err, ErrProxyConfigRunning) || release != nil {
				if release != nil {
					release()
				}
				t.Fatalf("release returned: %t, error: %v", release != nil, err)
			}
			if inst.GetStatus() != status {
				t.Fatal("configuration action stopped the server")
			}
		})
	}
}

func TestProxyConfigOperationHoldsPowerAndRestoreGate(t *testing.T) {
	m, srv, _ := restoreFixture(t)
	release, err := m.BeginProxyConfigOperation(context.Background(), srv.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Start(srv); !errors.Is(err, ErrServerOperation) {
		release()
		t.Fatal("start bypassed gate", err)
	}
	if _, err := m.RestoreBackup(context.Background(), srv, "valid.tar.gz"); !errors.Is(err, ErrServerOperation) {
		release()
		t.Fatal("restore bypassed gate", err)
	}
	release()
	release()
	again, err := m.BeginProxyConfigOperation(context.Background(), srv.ID)
	if err != nil {
		t.Fatal("gate not released", err)
	}
	again()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := m.BeginProxyConfigOperation(ctx, srv.ID); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestProxyStartupCreatesMissingConfigAtAssignedPort(t *testing.T) {
	for _, kind := range []string{"velocity", "bungeecord", "waterfall"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			if err := syncServerPort(dir, kind, 25588); err != nil {
				t.Fatal(err)
			}
			cfg, err := proxyconfig.Read(dir, kind, 25588)
			if err != nil || !cfg.Exists {
				t.Fatal(cfg, err)
			}
			data, err := os.ReadFile(filepath.Join(dir, cfg.File))
			if err != nil || !strings.Contains(string(data), "25588") {
				t.Fatal(string(data), err)
			}
			if _, err := os.Stat(filepath.Join(dir, "server.properties")); !os.IsNotExist(err) {
				t.Fatal("proxy generated world settings", err)
			}
			if kind == "velocity" {
				if secret, err := proxyconfig.ReadSecret(dir, kind); err != nil || len(secret) < 32 {
					t.Fatal("missing forwarding secret", err)
				}
			}
		})
	}
}
