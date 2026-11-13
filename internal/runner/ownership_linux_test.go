//go:build linux

package runner

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func init() {
	path := os.Getenv("NEOPANEL_WORLD_LOCK_HELPER")
	if path == "" {
		return
	}
	f, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		os.Exit(2)
	}
	lock := syscall.Flock_t{Type: syscall.F_WRLCK, Whence: 0, Len: 0}
	if syscall.FcntlFlock(f.Fd(), syscall.F_SETLK, &lock) != nil {
		os.Exit(3)
	}
	fmt.Println("locked")
	time.Sleep(time.Minute)
	f.Close()
	os.Exit(0)
}

func TestLinuxWorldLockDetectsLiveOwnerWithoutDeletingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.lock")
	if err := os.WriteFile(path, []byte("keep-lock-inode"), 0600); err != nil {
		t.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(exe, "-test.run=^$")
	cmd.Env = append(os.Environ(), "NEOPANEL_WORLD_LOCK_HELPER="+path)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { cmd.Process.Kill(); cmd.Wait() }()
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() || scanner.Text() != "locked" {
		t.Fatal("lock helper failed")
	}
	if err := checkWorldLock(path); err == nil || !strings.Contains(err.Error(), "already in use") {
		t.Fatal("live lock not detected", err)
	}
	cmd.Process.Kill()
	cmd.Wait()
	if err := checkWorldLock(path); err != nil {
		t.Fatal("stale file mistaken for active lock", err)
	}
	if raw, err := os.ReadFile(path); err != nil || string(raw) != "keep-lock-inode" {
		t.Fatal("lock file modified", err)
	}
}

func TestLinuxOwnedProcessGroupStopsChildren(t *testing.T) {
	dir := t.TempDir()
	lock, err := acquireRunLock(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	cmd := exec.Command("/bin/sh", "-c", "sleep 60 & echo $!; wait")
	configureProcessOwnership(cmd, lock)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { killProcessTree(cmd); cmd.Wait() }()
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() {
		t.Fatal("no child pid")
	}
	child, err := strconv.Atoi(scanner.Text())
	if err != nil {
		t.Fatal(err)
	}
	if err := killProcessTree(cmd); err != nil {
		t.Fatal(err)
	}
	cmd.Wait()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		raw, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", child))
		if os.IsNotExist(err) {
			return
		}
		if err == nil {
			fields := strings.Fields(string(raw)[strings.LastIndex(string(raw), ")")+1:])
			if len(fields) > 0 && fields[0] == "Z" {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("child survived process-group termination")
}
