package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestOwnershipLockSurvivesServerDirectoryReplacement(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "server")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	lock, err := acquireRunLock(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	path, err := serverRunLockPath(dir)
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(path, dir+string(os.PathSeparator)) {
		t.Fatal("lock is inside writable server tree")
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	second, err := acquireRunLock(dir)
	if err == nil {
		second.Close()
		t.Fatal("replacing server directory bypassed ownership")
	}
}

func TestOwnershipBlocksOtherManagerMaintenanceAndRestore(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "server")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	m := NewServerManager(root, nil)
	defer m.ShutdownAll()
	lock, err := acquireRunLock(dir)
	if err != nil {
		t.Fatal(err)
	}
	called := false
	if err := m.DeleteServer(context.Background(), "server", func() error { called = true; return nil }); err == nil || called {
		t.Fatal("maintenance bypassed another process ownership", err)
	}
	if _, err := m.RestoreBackup(context.Background(), &database.Server{ID: "server"}, "test.tar.gz"); err == nil || !strings.Contains(err.Error(), "restore refused") {
		t.Fatal("restore bypassed ownership", err)
	}
	lock.Close()
	release, err := m.BeginServerMaintenance(context.Background(), "server")
	if err != nil {
		t.Fatal(err)
	}
	if lock, err := acquireRunLock(dir); err == nil {
		lock.Close()
		t.Fatal("maintenance did not retain ownership")
	}
	release()
	release()
	lock, err = acquireRunLock(dir)
	if err != nil {
		t.Fatal("maintenance lock leaked", err)
	}
	lock.Close()
}

func TestResolverBindNeverExposesParentDirectory(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "resolv.conf")
	if err := os.WriteFile(file, []byte("nameserver 127.0.0.1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	args := sandboxResolverArgs(file)
	if len(args) != 3 || args[0] != "--ro-bind-try" || args[1] != file || args[2] != "/etc/resolv.conf" {
		t.Fatalf("unexpected resolver mounts: %v", args)
	}
	if args := sandboxResolverArgs(filepath.Join(dir, "missing")); len(args) != 0 {
		t.Fatal(args)
	}
}
