package runner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestAutomaticRetrySurvivesOperationContention(t *testing.T) {
	for _, mode := range []string{"resume", "stop", "restore", "shutdown"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			db, err := database.InitDB(filepath.Join(root, "panel.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			srv := &database.Server{ID: "alpha", Name: "alpha", Type: "paper", Version: "1.21", JarFile: "missing.jar", JavaPath: filepath.Join(root, "missing-java"), AutoRestart: true}
			if err := db.CreateServer(srv); err != nil {
				t.Fatal(err)
			}
			m := NewServerManager(filepath.Join(root, "servers"), db)
			defer m.ShutdownAll()
			os.MkdirAll(m.GetWorkingDir(srv.ID), 0755)
			// A pending crash restart whose three-second deadline expires during backup.
			cancel := make(chan struct{})
			m.retryCancels[srv.ID] = cancel
			release, err := m.beginOperation(srv.ID)
			if err != nil {
				t.Fatal(err)
			}
			done := make(chan struct{})
			go func() { m.waitAndRetryStart(srv.ID, 1, cancel); close(done) }()
			time.Sleep(3200 * time.Millisecond)
			select {
			case <-done:
				t.Fatal("contention discarded pending restart")
			default:
			}
			switch mode {
			case "stop":
				if err := m.Stop(srv.ID, false); !errors.Is(err, ErrServerOperation) {
					t.Fatal(err)
				}
			case "restore":
				// Restore cancels pending retries under this same gate before extracting.
				m.mu.Lock()
				m.cancelRetryLocked(srv.ID)
				m.restorePaused.Store(srv.ID, true)
				m.mu.Unlock()
			case "shutdown":
				m.ShutdownAll()
			}
			release()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("retry worker did not complete")
			}
			m.mu.RLock()
			attempts := m.startupFailures[srv.ID]
			m.mu.RUnlock()
			if mode == "resume" && attempts != 1 {
				t.Fatal("automatic launch not attempted after release", attempts)
			}
			if mode != "resume" && attempts != 0 {
				t.Fatal("cancelled restart resurrected", attempts)
			}
		})
	}
}

func TestDeleteGateSpansCleanupAndReleasesAfterFailure(t *testing.T) {
	m, srv, _ := restoreFixture(t)
	entered, proceed := make(chan struct{}), make(chan struct{})
	done := make(chan error, 1)
	failure := errors.New("cleanup failed")
	go func() {
		done <- m.DeleteServer(context.Background(), srv.ID, func() error { close(entered); <-proceed; return failure })
	}()
	<-entered
	if _, err := m.RestoreBackup(context.Background(), srv, "test.tar.gz"); !errors.Is(err, ErrServerOperation) {
		t.Fatal("restore bypassed deletion", err)
	}
	if err := m.RemoveInstance(srv.ID); !errors.Is(err, ErrServerOperation) {
		t.Fatal("runtime removal bypassed gate", err)
	}
	close(proceed)
	if err := <-done; !errors.Is(err, failure) {
		t.Fatal(err)
	}
	release, err := m.BeginServerMaintenance(context.Background(), srv.ID)
	if err != nil {
		t.Fatal("cleanup failure leaked gate", err)
	}
	release()
	got, err := os.ReadFile(filepath.Join(m.GetWorkingDir(srv.ID), "original"))
	if err != nil || string(got) != "keep me" {
		t.Fatal("conflict changed live data", err)
	}
}
