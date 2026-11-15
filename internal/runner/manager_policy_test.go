package runner

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestLocalInstanceDoesNotInheritRemoteNodePolicy(t *testing.T) {
	dir := t.TempDir()
	db, err := database.InitDB(filepath.Join(dir, "panel.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.CreateNode(&database.ClusterNode{ID: "remote", Name: "remote", IsolationEngine: "bwrap", TasksMax: 42}); err != nil {
		t.Fatal(err)
	}
	manager := NewServerManager(filepath.Join(dir, "servers"), db)
	defer manager.ShutdownAll()
	inst := manager.getOrCreateInstance(&database.Server{ID: "local-default"})
	if inst.isolationEngine != "" || inst.tasksMax != 0 {
		t.Fatalf("inherited remote policy: %q, %d", inst.isolationEngine, inst.tasksMax)
	}
	if err := db.CreateNode(&database.ClusterNode{ID: "local", Name: "local", IsolationEngine: "process", TasksMax: 123}); err != nil {
		t.Fatal(err)
	}
	inst = manager.getOrCreateInstance(&database.Server{ID: "local-configured"})
	if inst.isolationEngine != "process" || inst.tasksMax != 123 {
		t.Fatalf("local policy not applied: %q, %d", inst.isolationEngine, inst.tasksMax)
	}
}

func TestOfflineStatsDeclareNetworkCapability(t *testing.T) {
	manager := NewServerManager(t.TempDir(), nil)
	defer manager.ShutdownAll()
	srv := &database.Server{ID: "offline", MemoryMaxMB: 1024}
	check := func(stats database.ServerStats) {
		t.Helper()
		if stats.NetworkSupported == nil || *stats.NetworkSupported != (runtime.GOOS == "linux") {
			t.Fatalf("incorrect network capability: %+v", stats)
		}
	}
	check(manager.GetStats(srv))
	manager.getOrCreateInstance(srv)
	check(manager.GetStats(srv))
}
