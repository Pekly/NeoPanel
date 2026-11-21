package runner

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func restartFixture(t *testing.T) (*database.DB, string, *database.Server, *database.Server) {
	t.Helper()
	root := t.TempDir()
	db, err := database.InitDB(filepath.Join(root, "panel.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(root, "servers")
	online := &database.Server{ID: "online-before-update", Name: "Online", Type: "paper", JarFile: "server.jar", JavaPath: exe, AutoRestart: false}
	offline := &database.Server{ID: "offline-before-update", Name: "Offline", Type: "paper", JarFile: "server.jar", JavaPath: exe, AutoRestart: true}
	for _, srv := range []*database.Server{online, offline} {
		if err := db.CreateServer(srv); err != nil {
			t.Fatal(err)
		}
		dir := filepath.Join(base, srv.ID)
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "server.jar"), []byte("fixture"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return db, base, online, offline
}

func TestRestartRecoveryResumesOnlyOnlineServer(t *testing.T) {
	t.Setenv("NEOPANEL_LIFECYCLE_HELPER", "graceful")
	db, base, online, offline := restartFixture(t)
	old := NewServerManager(base, db)
	t.Cleanup(old.ShutdownAll)
	if err := old.RestoreRunningServers(); err != nil {
		t.Fatal(err)
	}
	old.getOrCreateInstance(online).SetIsolationConfig("process", 512)
	if err := old.Start(online); err != nil {
		t.Fatal(err)
	}
	inst, _ := old.GetInstance(online.ID)
	waitProcessStatus(t, inst, "online")
	var calls sync.WaitGroup
	for i := 0; i < 8; i++ {
		calls.Add(1)
		go func() { defer calls.Done(); old.ShutdownAll() }()
	}
	calls.Wait()
	old.ShutdownAll() // Must not overwrite the saved online ID with an empty list.
	if old.GetStatus(online.ID) != "offline" {
		t.Fatal("shutdown did not stop old process")
	}
	next := NewServerManager(base, db)
	t.Cleanup(next.ShutdownAll)
	next.getOrCreateInstance(online).SetIsolationConfig("process", 512)
	if err := next.RestoreRunningServers(); err != nil {
		t.Fatal(err)
	}
	resumed, _ := next.GetInstance(online.ID)
	waitProcessStatus(t, resumed, "online")
	if next.GetStatus(offline.ID) != "offline" {
		t.Fatal("offline server was started")
	}
	if err := next.RestoreRunningServers(); err != nil {
		t.Fatal(err)
	}
	if current, _ := next.GetInstance(online.ID); current != resumed {
		t.Fatal("duplicate restoration")
	}
	if ids, err := db.TakeServerResumeState(base); err != nil || len(ids) != 0 {
		t.Fatal("snapshot replayable", ids, err)
	}
	// A deliberate stop must remain stopped across the following clean restart.
	if err := next.Stop(online.ID, true); err != nil {
		t.Fatal(err)
	}
	next.ShutdownAll()
	final := NewServerManager(base, db)
	t.Cleanup(final.ShutdownAll)
	if err := final.RestoreRunningServers(); err != nil {
		t.Fatal(err)
	}
	if final.GetStatus(online.ID) != "offline" {
		t.Fatal("explicit stop was undone")
	}
	if _, exists := final.GetInstance(online.ID); exists {
		t.Fatal("offline server was launched")
	}
}

func TestRestartRecoverySnapshotSelection(t *testing.T) {
	db, base, _, _ := restartFixture(t)
	m := NewServerManager(base, db)
	t.Cleanup(func() { m.mu.Lock(); m.instances = map[string]*ProcessInstance{}; m.mu.Unlock(); m.ShutdownAll() })
	m.recoveryEnabled = true
	m.recoveryPending = map[string]bool{"not-yet-attempted": true, "paused-pending": true}
	m.recoveryStarting = map[string]bool{"resuming": true}
	for id, status := range map[string]string{"online": "online", "new-start": "starting", "resuming": "starting", "stopping": "stopping", "offline": "offline", "paused": "online", "halted": "online", "busy": "online", "blocked": "online"} {
		p := NewProcessInstance(&database.Server{ID: id}, filepath.Join(base, id), time.Second, nil, nil)
		p.status = status
		m.instances[id] = p
	}
	m.restorePaused.Store("paused", true)
	m.restorePaused.Store("paused-pending", true)
	m.startupHalted["halted"] = true
	m.restoreBlocked.Store("blocked", ErrRestoreRecovery)
	release, err := m.beginOperation("busy")
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if err := m.saveRestartState(); err != nil {
		t.Fatal(err)
	}
	ids, err := db.TakeServerResumeState(base)
	if err != nil || !reflect.DeepEqual(ids, []string{"not-yet-attempted", "online", "resuming"}) {
		t.Fatal(ids, err)
	}
}

func TestRestartRecoverySkipsDeletedAndFailedLaunchesOnce(t *testing.T) {
	db, base, broken, _ := restartFixture(t)
	broken.JavaPath = filepath.Join(base, "missing-java")
	if err := db.UpdateServer(broken); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveServerResumeState(base, []string{broken.ID, "deleted"}); err != nil {
		t.Fatal(err)
	}
	m := NewServerManager(base, db)
	defer m.ShutdownAll()
	m.getOrCreateInstance(broken).SetIsolationConfig("process", 512)
	if err := m.RestoreRunningServers(); err != nil {
		t.Fatal(err)
	}
	if m.GetStatus(broken.ID) != "offline" {
		t.Fatal("failed launch reported active")
	}
	logs, err := db.ListActivity(0, true, 100)
	if err != nil || len(logs) != 2 {
		t.Fatal(logs, err)
	}
	if err := m.RestoreRunningServers(); err != nil {
		t.Fatal(err)
	}
	after, _ := db.ListActivity(0, true, 100)
	if len(after) != len(logs) {
		t.Fatal("replayed failed recovery")
	}
	if ids, err := db.TakeServerResumeState(base); err != nil || len(ids) != 0 {
		t.Fatal(ids, err)
	}
}

func TestRestartRecoveryPreservesSnapshotWhenStartupNeverCompletes(t *testing.T) {
	db, base, srv, _ := restartFixture(t)
	if err := db.SaveServerResumeState(base, []string{srv.ID}); err != nil {
		t.Fatal(err)
	}
	// Equivalent to a bind/setup failure: construction alone never arms recovery.
	m := NewServerManager(base, db)
	m.ShutdownAll()
	if ids, err := db.TakeServerResumeState(base); err != nil || !reflect.DeepEqual(ids, []string{srv.ID}) {
		t.Fatal(ids, err)
	}
	if err := m.RestoreRunningServers(); !errors.Is(err, errDaemonShuttingDown) {
		t.Fatal(err)
	}
	if err := m.Start(srv); !errors.Is(err, errDaemonShuttingDown) {
		t.Fatal(err)
	}
}

func TestRestartRecoveryRespectsExplicitStopBeforeRestore(t *testing.T) {
	db, base, srv, _ := restartFixture(t)
	if err := db.SaveServerResumeState(base, []string{srv.ID}); err != nil {
		t.Fatal(err)
	}
	m := NewServerManager(base, db)
	defer m.ShutdownAll()
	_ = m.Stop(srv.ID, true)
	if err := m.RestoreRunningServers(); err != nil {
		t.Fatal(err)
	}
	if _, exists := m.GetInstance(srv.ID); exists {
		t.Fatal("stop was undone by restore")
	}
}

func TestRestartRecoveryShutdownRacingRestoreRetainsIntent(t *testing.T) {
	t.Setenv("NEOPANEL_LIFECYCLE_HELPER", "graceful")
	db, base, srv, _ := restartFixture(t)
	if err := db.SaveServerResumeState(base, []string{srv.ID}); err != nil {
		t.Fatal(err)
	}
	m := NewServerManager(base, db)
	defer m.ShutdownAll()
	m.getOrCreateInstance(srv).SetIsolationConfig("process", 512)
	m.launchMu.Lock()
	restored, stopped := make(chan error, 1), make(chan struct{})
	go func() { restored <- m.RestoreRunningServers() }()
	go func() { m.ShutdownAll(); close(stopped) }()
	m.launchMu.Unlock()
	select {
	case err := <-restored:
		if err != nil && !errors.Is(err, errDaemonShuttingDown) {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("restore did not finish")
	}
	select {
	case <-stopped:
	case <-time.After(10 * time.Second):
		t.Fatal("shutdown did not finish")
	}
	if m.GetStatus(srv.ID) != "offline" {
		t.Fatal("process survived shutdown")
	}
	ids, err := db.TakeServerResumeState(base)
	if err != nil || !reflect.DeepEqual(ids, []string{srv.ID}) {
		t.Fatal("pending recovery was lost", ids, err)
	}
}

func TestRestartRecoveryRepeatedFailureReturnsOriginalError(t *testing.T) {
	m := NewServerManager(t.TempDir(), nil)
	defer m.ShutdownAll()
	first := m.RestoreRunningServers()
	if first == nil {
		t.Fatal("missing database accepted")
	}
	if again := m.RestoreRunningServers(); again != first {
		t.Fatal(first, again)
	}
}
