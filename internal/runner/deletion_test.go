package runner

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestDeleteStopsLiveProcessDespiteStaleOfflineStatus(t *testing.T) {
	p := lifecycleInstance(t, "listener")
	id := filepath.Base(p.workingDir)
	p.server.ID = id
	m := NewServerManager(filepath.Dir(p.workingDir), nil)
	t.Cleanup(m.ShutdownAll)
	m.instances[id] = p
	if err := p.Start(); err != nil {
		t.Fatal(err)
	}
	waitProcessStatus(t, p, "online")
	address, err := os.ReadFile(filepath.Join(p.workingDir, "listener.addr"))
	if err != nil {
		t.Fatal(err)
	}
	if listener, err := net.Listen("tcp4", string(address)); err == nil {
		listener.Close()
		t.Fatal("test server never owned port")
	}
	p.mu.Lock()
	p.status = "offline" // reproduce stale status without losing the live process handle
	p.mu.Unlock()
	called := false
	err = m.DeleteServer(context.Background(), id, func() error {
		called = true
		listener, err := net.Listen("tcp4", string(address))
		if err != nil {
			return errors.New("deletion reached file cleanup while server still owned port")
		}
		listener.Close()
		return os.RemoveAll(p.workingDir)
	})
	if err != nil || !called {
		t.Fatal("deletion failed", err)
	}
	if _, ok := m.GetInstance(id); ok {
		t.Fatal("deleted runtime still registered")
	}
	listener, err := net.Listen("tcp4", string(address))
	if err != nil {
		t.Fatal("deleted server still owns port", err)
	}
	listener.Close()
}

func TestDeleteRetainsRecordAndInstanceWhenStopCannotBeConfirmed(t *testing.T) {
	dir := t.TempDir()
	m := NewServerManager(dir, nil)
	failure := errors.New("process cleanup not confirmed")
	p := NewProcessInstance(&database.Server{ID: "server"}, filepath.Join(dir, "server"), time.Second, nil, nil)
	p.cleanupErr = failure
	m.instances["server"] = p
	called := false
	if err := m.DeleteServer(context.Background(), "server", func() error { called = true; return nil }); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	if called {
		t.Fatal("metadata removed after cleanup failure")
	}
	if got, ok := m.GetInstance("server"); !ok || got != p {
		t.Fatal("lost process ownership on failed delete")
	}
	p.cleanupErr = nil
	m.ShutdownAll()
}

func TestDeleteKeepsStoppedInstanceWhenCleanupFails(t *testing.T) {
	dir := t.TempDir()
	m := NewServerManager(dir, nil)
	defer m.ShutdownAll()
	p := NewProcessInstance(&database.Server{ID: "server"}, filepath.Join(dir, "server"), time.Second, nil, nil)
	m.instances["server"] = p
	failure := errors.New("database cleanup failed")
	if err := m.DeleteServer(context.Background(), "server", func() error { return failure }); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	if got, ok := m.GetInstance("server"); !ok || got != p {
		t.Fatal("failed cleanup discarded runtime record")
	}
}

func TestStopCanReconcilePreviousCleanupFailure(t *testing.T) {
	p := lifecycleInstance(t, "graceful")
	if err := p.Start(); err != nil {
		t.Fatal(err)
	}
	waitProcessStatus(t, p, "online")
	p.mu.RLock()
	cmd := p.cmd
	p.mu.RUnlock()
	if err := p.Stop(true); err != nil {
		t.Fatal(err)
	}
	lock, err := acquireRunLock(p.workingDir)
	if err != nil {
		t.Fatal(err)
	}
	p.mu.Lock()
	p.status, p.cmd, p.runLock, p.cleanupErr = "stopping", cmd, lock, errors.New("temporary verification failure")
	p.mu.Unlock()
	if err := p.Stop(false); err != nil {
		t.Fatal("failed to recover after process was already cleaned up", err)
	}
	if p.GetStatus() != "offline" {
		t.Fatal("cleanup retry did not unblock server")
	}
	lock, err = acquireRunLock(p.workingDir)
	if err != nil {
		t.Fatal("cleanup retry retained lock", err)
	}
	lock.Close()
}
