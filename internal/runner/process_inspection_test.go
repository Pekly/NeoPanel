package runner

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestProcessIdentityParsing(t *testing.T) {
	got, err := parseProcIdentity("4242 (server ) name) S 1 4240 0 0")
	if err != nil || got.PID != 4242 || got.Group != 4240 || got.State != "S" {
		t.Fatal(got, err)
	}
	for _, invalid := range []string{"", "42 nope", "42 (java) S", "x (java) S 1 42"} {
		if _, err := parseProcIdentity(invalid); err == nil {
			t.Fatal("accepted invalid stat", invalid)
		}
	}
}

func TestProcessInspectionDetectsUntrackedAndDeletedDirectories(t *testing.T) {
	root, server := t.TempDir(), filepath.Join(t.TempDir(), "server")
	pid := 424242
	if pid == os.Getpid() {
		pid++
	}
	proc := filepath.Join(root, fmt.Sprint(pid))
	if err := os.MkdirAll(proc, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proc, "stat"), []byte(fmt.Sprintf("%d (steel) S 1 %d 0", pid, pid)), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		link, target string
		blocked      bool
	}{
		{"cwd", server, true}, {"cwd", server + " (deleted)", true}, {"cwd", filepath.Join(server, "world"), true},
		{"exe", filepath.Join(server, "steel") + " (deleted)", true}, {"cwd", server + "-other", false}, {"cwd", filepath.Dir(server), false},
	} {
		readlink := func(path string) (string, error) {
			if filepath.Base(path) == tc.link {
				return tc.target, nil
			}
			return "", os.ErrNotExist
		}
		err := inspectServerProcesses(root, server, readlink)
		if errors.Is(err, ErrServerProcessAlive) != tc.blocked {
			t.Errorf("%+v: %v", tc, err)
		}
		if tc.blocked && !strings.Contains(err.Error(), fmt.Sprint(pid)) {
			t.Fatal("missing surviving PID", err)
		}
	}
	if err := inspectServerProcesses(root, server, func(string) (string, error) { return "", os.ErrPermission }); err == nil {
		t.Fatal("inspection failure treated as offline")
	}
	if err := os.WriteFile(filepath.Join(proc, "stat"), []byte(fmt.Sprintf("%d (steel) Z 1 %d 0", pid, pid)), 0600); err != nil {
		t.Fatal(err)
	}
	if err := inspectServerProcesses(root, server, func(string) (string, error) { t.Fatal("inspected zombie"); return "", nil }); err != nil {
		t.Fatal(err)
	}
}

func TestScopeShutdownRequiresExplicitInactiveState(t *testing.T) {
	for _, tc := range []struct {
		output  string
		stopped bool
	}{
		{"LoadState=loaded\nActiveState=inactive\n", true}, {"LoadState=not-found\nActiveState=inactive\n", true},
		{"LoadState=loaded\nActiveState=active\n", false}, {"LoadState=loaded\nActiveState=deactivating\n", false},
		{"Failed to connect to bus: file not found", false}, {"Unit not loaded", false}, {"", false},
	} {
		if scopeStopped(tc.output) != tc.stopped {
			t.Errorf("incorrect scope confirmation: %q", tc.output)
		}
	}
}

func TestProcessExitRacesAndUIDChecks(t *testing.T) {
	if !processDisappeared(&os.PathError{Op: "read", Path: "/proc/1/stat", Err: syscall.ESRCH}) || !processDisappeared(os.ErrNotExist) || processDisappeared(os.ErrPermission) {
		t.Fatal("incorrect proc exit-race handling")
	}
	for _, tc := range []struct {
		status  string
		ours    bool
		invalid bool
	}{
		{"Name: java\nUid:\t1000\t1000\t1000\t1000\n", true, false},
		{"Uid:\t0\t1000\t0\t0\n", true, false},
		{"Uid:\t0\t0\t0\t0\n", false, false},
		{"Uid: 1000\n", false, true}, {"", false, true},
	} {
		ours, err := processHasUID(tc.status, 1000)
		if ours != tc.ours || (err != nil) != tc.invalid {
			t.Fatalf("UID check %q: %v %v", tc.status, ours, err)
		}
	}
}
