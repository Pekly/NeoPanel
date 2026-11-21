package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStartupPropertiesBoundAndWorldPaths(t *testing.T) {
	for _, tc := range []struct {
		name, content, want string
	}{
		{"oversized", strings.Repeat("#", startupPropertiesMaxBytes+1), "maximum 1048576 bytes"},
		{"at-limit", strings.Repeat("#", startupPropertiesMaxBytes), ""},
		{"nested", "level-name=worlds/main\n", ""},
		{"absolute", "level-name=/outside\n", "invalid level-name"},
		{"traversal", "level-name=../outside\n", "invalid level-name"},
		{"cleaned-traversal", "level-name=worlds/../outside\n", "invalid level-name"},
		{"windows-absolute", "level-name=C:/outside\n", "invalid level-name"},
		{"windows-traversal", "level-name=..\\outside\n", "invalid level-name"},
		{"root", "level-name=.\n", "invalid level-name"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "server.properties"), []byte(tc.content), 0600); err != nil {
				t.Fatal(err)
			}
			err := checkServerWorldLocks(dir, "paper")
			if tc.want == "" && err != nil || tc.want != "" && (err == nil || !strings.Contains(err.Error(), tc.want)) {
				t.Fatalf("got %v, want %q", err, tc.want)
			}
		})
	}
	dir := t.TempDir()
	if err := checkServerWorldLocks(dir, "paper"); err != nil {
		t.Fatalf("missing properties/locks should be accepted: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "server.properties"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := checkServerWorldLocks(dir, "paper"); err == nil {
		t.Fatal("directory configuration accepted")
	}
}

func TestStartupRejectsOutsideLinksWithoutChangingSentinel(t *testing.T) {
	for _, target := range []string{"server.properties", "world/session.lock", "world"} {
		t.Run(target, func(t *testing.T) {
			dir, outside := t.TempDir(), t.TempDir()
			sentinel := filepath.Join(outside, "sentinel")
			data := "level-name=outside-sentinel\n"
			if err := os.WriteFile(sentinel, []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			link := filepath.Join(dir, filepath.FromSlash(target))
			if err := os.MkdirAll(filepath.Dir(link), 0700); err != nil {
				t.Fatal(err)
			}
			dest := sentinel
			if target == "world" {
				dest = outside
			}
			if err := os.Symlink(dest, link); err != nil {
				t.Skipf("symlink unavailable: %v", err)
			}
			if err := checkServerWorldLocks(dir, "paper"); err == nil {
				t.Fatal("outside symlink accepted")
			}
			if got, err := os.ReadFile(sentinel); err != nil || string(got) != data {
				t.Fatalf("outside sentinel changed: %q, %v", got, err)
			}
		})
	}
}

func TestStartupNestedWorldLockIsInspected(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "server.properties"), []byte("level-name=worlds/main\n"), 0600); err != nil {
		t.Fatal(err)
	}
	// A directory is never a valid session.lock, including for nested worlds.
	if err := os.MkdirAll(filepath.Join(dir, "worlds", "main", "session.lock"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := checkServerWorldLocks(dir, "paper"); err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("nested world lock not checked: %v", err)
	}
}
