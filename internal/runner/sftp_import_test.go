package runner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func importFixture(t *testing.T) (*ServerManager, *database.Server, string) {
	t.Helper()
	m, srv, _ := restoreFixture(t)
	live := m.GetWorkingDir(srv.ID)
	if err := os.Remove(filepath.Join(live, "original")); err != nil {
		t.Fatal(err)
	}
	return m, srv, live
}

func assertNoImportStage(t *testing.T, m *ServerManager) {
	t.Helper()
	stages, err := filepath.Glob(filepath.Join(m.baseDir, ".restore-*"))
	if err != nil || len(stages) != 0 {
		t.Fatalf("staging leaked: %v, %v", stages, err)
	}
}

func TestImportServerFilesSuccessAndOperationGate(t *testing.T) {
	m, srv, live := importFixture(t)
	before, err := os.Stat(live)
	if err != nil {
		t.Fatal(err)
	}
	// Windows resolves Stat's file identity lazily; pin it before the rename.
	if !os.SameFile(before, before) {
		t.Fatal("could not read original directory identity")
	}
	retry := make(chan struct{})
	m.retryCancels[srv.ID] = retry
	result, err := m.ImportServerFiles(context.Background(), srv, func(ctx context.Context, fresh string) error {
		if filepath.Dir(filepath.Dir(fresh)) != m.baseDir || !strings.HasPrefix(filepath.Base(filepath.Dir(fresh)), ".restore-"+srv.ID+"-") {
			t.Fatalf("not a private sibling stage: %s", fresh)
		}
		stageInfo, err := os.Stat(filepath.Dir(fresh))
		if err != nil {
			return err
		}
		if runtime.GOOS != "windows" && stageInfo.Mode().Perm() != 0700 {
			t.Fatalf("stage permissions: %v", stageInfo.Mode())
		}
		freshInfo, err := os.Stat(fresh)
		if err != nil {
			return err
		}
		if runtime.GOOS != "windows" && freshInfo.Mode().Perm()&0077 != 0 {
			t.Fatalf("transfer directory must satisfy the transport privacy check: %v", freshInfo.Mode())
		}
		entries, err := os.ReadDir(fresh)
		if err != nil || len(entries) != 0 {
			t.Fatalf("stage is not fresh: %v, %v", entries, err)
		}
		if !m.launchMu.TryLock() {
			t.Fatal("global launch lock held during transfer")
		}
		m.launchMu.Unlock()
		for name, operation := range map[string]func() error{
			"start": func() error { return m.Start(srv) },
			"delete": func() error {
				return m.DeleteServer(ctx, srv.ID, func() error { t.Error("delete callback ran"); return nil })
			},
			"restore": func() error { _, err := m.RestoreBackup(ctx, srv, "valid.tar.gz"); return err },
			"import": func() error {
				_, err := m.ImportServerFiles(ctx, srv, func(context.Context, string) error { t.Error("second import ran"); return nil })
				return err
			},
		} {
			if err := operation(); !errors.Is(err, ErrServerOperation) {
				t.Errorf("%s bypassed gate: %v", name, err)
			}
		}
		return os.WriteFile(filepath.Join(fresh, "world.dat"), []byte("imported"), 0600)
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.SafetyCopy == "" || m.GetStatus(srv.ID) != "offline" {
		t.Fatalf("unexpected result/status: %+v, %s", result, m.GetStatus(srv.ID))
	}
	got, err := os.ReadFile(filepath.Join(live, "world.dat"))
	if err != nil || string(got) != "imported" {
		t.Fatalf("missing published data: %q, %v", got, err)
	}
	original := filepath.Join(m.baseDir, result.SafetyCopy)
	after, err := os.Stat(original)
	if err != nil || !os.SameFile(before, after) {
		t.Fatalf("original directory was not retained: %v", err)
	}
	entries, err := os.ReadDir(original)
	if err != nil || len(entries) != 0 {
		t.Fatalf("recovery original not empty: %v, %v", entries, err)
	}
	select {
	case <-retry:
	default:
		t.Fatal("retry not canceled")
	}
	if _, paused := m.restorePaused.Load(srv.ID); !paused {
		t.Fatal("automatic restart not paused")
	}
}

func TestImportServerFilesRejectsNonemptyAndActive(t *testing.T) {
	for _, name := range []string{"server.properties", "eula.txt", ".hidden", "empty-subdir"} {
		t.Run(name, func(t *testing.T) {
			m, srv, live := importFixture(t)
			var err error
			if name == "empty-subdir" {
				err = os.Mkdir(filepath.Join(live, name), 0700)
			} else {
				err = os.WriteFile(filepath.Join(live, name), nil, 0600)
			}
			if err != nil {
				t.Fatal(err)
			}
			_, err = m.ImportServerFiles(context.Background(), srv, func(context.Context, string) error { t.Error("transfer ran"); return nil })
			if !errors.Is(err, ErrImportTargetNotEmpty) {
				t.Fatal(err)
			}
			assertNoImportStage(t, m)
		})
	}
	for _, status := range []string{"online", "starting", "stopping"} {
		t.Run(status, func(t *testing.T) {
			m, srv, live := importFixture(t)
			inst := NewProcessInstance(srv, live, DefaultStartupTimeout, nil, nil)
			inst.status = status
			m.instances[srv.ID] = inst
			defer func() { inst.mu.Lock(); inst.status = "offline"; inst.mu.Unlock() }()
			_, err := m.ImportServerFiles(context.Background(), srv, func(context.Context, string) error { t.Error("transfer ran"); return nil })
			if !errors.Is(err, ErrServerNotStopped) || inst.GetStatus() != status {
				t.Fatalf("running server changed/accepted: %s, %v", inst.GetStatus(), err)
			}
			assertNoImportStage(t, m)
		})
	}
}

func TestImportServerFilesFailureAndCancellation(t *testing.T) {
	for _, mode := range []string{"error", "cancel-before", "cancel-after"} {
		t.Run(mode, func(t *testing.T) {
			m, srv, live := importFixture(t)
			before, _ := os.Stat(live)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			failure := errors.New("transfer failed")
			want := failure
			if mode != "error" {
				want = context.Canceled
			}
			if mode == "cancel-before" {
				cancel()
			}
			_, err := m.ImportServerFiles(ctx, srv, func(ctx context.Context, fresh string) error {
				if mode == "cancel-before" {
					t.Error("canceled callback ran")
				}
				if err := os.WriteFile(filepath.Join(fresh, "partial"), []byte("partial"), 0600); err != nil {
					return err
				}
				if mode == "cancel-after" {
					cancel()
					return nil // A successful transfer cannot override cancellation.
				}
				return failure
			})
			if !errors.Is(err, want) {
				t.Fatalf("wanted %v, got %v", want, err)
			}
			after, _ := os.Stat(live)
			entries, readErr := os.ReadDir(live)
			if !os.SameFile(before, after) || readErr != nil || len(entries) != 0 {
				t.Fatal("live directory changed")
			}
			assertNoImportStage(t, m)
		})
	}
}

func TestImportServerFilesTargetChangedDuringTransfer(t *testing.T) {
	for _, mode := range []string{"late-write", "replacement"} {
		t.Run(mode, func(t *testing.T) {
			m, srv, live := importFixture(t)
			_, err := m.ImportServerFiles(context.Background(), srv, func(ctx context.Context, fresh string) error {
				if mode == "late-write" {
					return os.WriteFile(filepath.Join(live, "late"), []byte("preserve"), 0600)
				}
				if err := os.Rename(live, live+"-moved"); err != nil {
					return err
				}
				return os.Mkdir(live, 0755)
			})
			if err == nil {
				t.Fatal("changed target accepted")
			}
			if mode == "late-write" {
				got, readErr := os.ReadFile(filepath.Join(live, "late"))
				if !errors.Is(err, ErrImportTargetNotEmpty) || readErr != nil || string(got) != "preserve" {
					t.Fatalf("late write lost: %q, %v, %v", got, readErr, err)
				}
			} else if !strings.Contains(err.Error(), "changed during transfer") {
				t.Fatalf("wrong replacement error: %v", err)
			}
			assertNoImportStage(t, m)
		})
	}
}

func TestImportServerFilesRetentionSharedWithRestore(t *testing.T) {
	m, srv, _ := importFixture(t)
	noop := func(context.Context, string) error { return nil }
	for i := 0; i < restoreMaxRetained; i++ {
		if _, err := m.ImportServerFiles(context.Background(), srv, noop); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := m.ImportServerFiles(context.Background(), srv, func(context.Context, string) error { t.Error("transfer ran over retention limit"); return nil }); !errors.Is(err, ErrRestoreRetention) {
		t.Fatal(err)
	}
	if _, err := m.RestoreBackup(context.Background(), srv, "valid.tar.gz"); !errors.Is(err, ErrRestoreRetention) {
		t.Fatalf("restore did not share import retention: %v", err)
	}
	stages, _ := filepath.Glob(filepath.Join(m.baseDir, ".restore-*"))
	if len(stages) != restoreMaxRetained {
		t.Fatalf("retained %d directories", len(stages))
	}
}

func TestImportServerFilesUnsafeTargetAndRunLock(t *testing.T) {
	for _, mode := range []string{"nil-server", "unsafe-id", "private-id", "remote", "file", "missing", "symlink", "locked"} {
		t.Run(mode, func(t *testing.T) {
			m, srv, live := importFixture(t)
			switch mode {
			case "nil-server":
				srv = nil
			case "unsafe-id":
				srv.ID = "../escape"
			case "private-id":
				srv.ID = ".restore-private"
			case "remote":
				srv.NodeID = "remote"
			case "file", "missing", "symlink":
				if err := os.Remove(live); err != nil {
					t.Fatal(err)
				}
				if mode == "file" {
					if err := os.WriteFile(live, nil, 0600); err != nil {
						t.Fatal(err)
					}
				}
				if mode == "symlink" {
					if err := os.Symlink(t.TempDir(), live); err != nil {
						t.Skip("symlinks unavailable: ", err)
					}
				}
			case "locked":
				lock, err := acquireRunLock(live)
				if err != nil {
					t.Fatal(err)
				}
				defer lock.Close()
			}
			if _, err := m.ImportServerFiles(context.Background(), srv, func(context.Context, string) error { t.Error("transfer ran"); return nil }); err == nil {
				t.Fatal("unsafe target accepted")
			}
			assertNoImportStage(t, m)
		})
	}
}

func TestImportServerFilesShutdownCancelsAndJoins(t *testing.T) {
	db, base, srv, _ := restartFixture(t)
	if err := os.Remove(filepath.Join(base, srv.ID, "server.jar")); err != nil {
		t.Fatal(err)
	}
	m := NewServerManager(base, db)
	t.Cleanup(m.ShutdownAll)
	if err := db.SetServerIdleStopped(srv.ID, true); err != nil {
		t.Fatal(err)
	}
	entered, canceled, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	defer close(release)
	importDone := make(chan error, 1)
	go func() {
		_, err := m.ImportServerFiles(context.Background(), srv, func(ctx context.Context, fresh string) error {
			close(entered)
			<-ctx.Done()
			close(canceled)
			<-release
			// Database access remains valid until the callback has finished.
			if _, err := db.GetServer(srv.ID); err != nil {
				return err
			}
			return nil // Deliberately ignore cancellation; publication must refuse.
		})
		importDone <- err
	}()
	select {
	case <-entered:
	case err := <-importDone:
		t.Fatalf("callback never entered: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("callback did not start")
	}
	state, err := db.GetServerWakeState(srv.ID)
	if err != nil || state.StoppedByIdlePolicy {
		t.Fatalf("idle wake not invalidated: %+v, %v", state, err)
	}
	shutdownDone := make(chan struct{})
	go func() { m.ShutdownAll(); close(shutdownDone) }()
	select {
	case <-canceled:
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown did not cancel transfer")
	}
	select {
	case <-shutdownDone:
		t.Fatal("shutdown returned before callback exited")
	case <-time.After(50 * time.Millisecond):
	}
	if _, err := m.ImportServerFiles(context.Background(), &database.Server{ID: "other"}, func(context.Context, string) error { t.Error("post-shutdown callback ran"); return nil }); !errors.Is(err, errDaemonShuttingDown) {
		t.Fatalf("import crossed shutdown fence: %v", err)
	}
	release <- struct{}{}
	select {
	case err := <-importDone:
		if !errors.Is(err, errDaemonShuttingDown) {
			t.Fatalf("shutdown import: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("import failed to exit")
	}
	select {
	case <-shutdownDone:
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown failed to join")
	}
	assertNoImportStage(t, m)
	entries, err := os.ReadDir(m.GetWorkingDir(srv.ID))
	if err != nil || len(entries) != 0 {
		t.Fatalf("shutdown published data: %v, %v", entries, err)
	}
}
