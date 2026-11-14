//go:build linux

package runner

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func init() {
	if os.Getenv("NEOPANEL_UNTRACKED_LISTENER") != "1" {
		return
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		os.Exit(3)
	}
	fmt.Println(listener.Addr().String())
	time.Sleep(time.Minute)
	listener.Close()
	os.Exit(0)
}

func TestLinuxDeleteRefusesUntrackedListenerAfterManagerRestart(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "server")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(dir, "world-data")
	if err := os.WriteFile(marker, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(exe, "-test.run=^$")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "NEOPANEL_UNTRACKED_LISTENER=1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { cmd.Process.Kill(); cmd.Wait() }()
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() {
		t.Fatal("untracked listener did not start")
	}
	address := scanner.Text()
	// A fresh manager has no instance or inherited lock for the old process.
	m := NewServerManager(root, nil)
	defer m.ShutdownAll()
	deleted := false
	cleanup := func() error { deleted = true; return os.RemoveAll(dir) }
	if err := m.DeleteServer(context.Background(), "server", cleanup); !errors.Is(err, ErrServerProcessAlive) {
		t.Fatal("untracked process not detected", err)
	}
	if deleted {
		t.Fatal("deleted live server")
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("live files removed", err)
	}
	if listener, err := net.Listen("tcp4", address); err == nil {
		listener.Close()
		t.Fatal("test unexpectedly stopped unrelated/untracked process")
	}
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	cmd.Wait()
	if err := m.DeleteServer(context.Background(), "server", cleanup); err != nil {
		t.Fatal("delete refused after confirmed cleanup", err)
	}
	if !deleted {
		t.Fatal("cleanup not performed")
	}
	listener, err := net.Listen("tcp4", address)
	if err != nil {
		t.Fatal("old server retained port", err)
	}
	listener.Close()
}

func TestLinuxCleanupPinsLauncherUntilChildrenStop(t *testing.T) {
	lock, err := acquireRunLock(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	cmd := exec.Command("/bin/sh", "-c", "sleep 60 & echo $!")
	configureProcessOwnership(cmd, lock)
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if cmd.ProcessState == nil {
			killProcessTree(cmd)
			cmd.Wait()
		}
	}()
	scanner := bufio.NewScanner(out)
	if !scanner.Scan() {
		t.Fatal("missing child PID")
	}
	child, err := strconv.Atoi(scanner.Text())
	if err != nil {
		t.Fatal(err)
	}
	if err := waitForOwnedExit(cmd); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", cmd.Process.Pid))
	if err != nil {
		t.Fatal("launcher was reaped before cleanup", err)
	}
	identity, err := parseProcIdentity(string(raw))
	if err != nil || identity.State != "Z" {
		t.Fatal("launcher identity not pinned", identity, err)
	}
	owned := &ownedCommand{cmd: cmd}
	if waitErr, cleanupErr := owned.wait(); waitErr != nil || cleanupErr != nil {
		t.Fatal(waitErr, cleanupErr)
	}
	if raw, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", child)); err == nil {
		identity, err := parseProcIdentity(string(raw))
		if err != nil || (identity.State != "Z" && identity.State != "X") {
			t.Fatal("child survived launcher cleanup", identity, err)
		}
	} else if !processDisappeared(err) {
		t.Fatal(err)
	}
	if err := owned.stop(false); err != nil {
		t.Fatal("late stop should not signal reaped group", err)
	}
}
