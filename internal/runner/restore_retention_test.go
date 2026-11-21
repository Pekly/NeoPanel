package runner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestRestoreRetentionFourthDeniedWithoutDeletingData(t *testing.T) {
	m, srv, backups := restoreFixture(t)
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "restored"), []byte("backup data"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateBackup(source, backups, "small.tar.gz"); err != nil {
		t.Fatal(err)
	}
	// Another server's recovery directories must not consume alpha's allowance.
	for i := 0; i < restoreMaxRetained; i++ {
		if err := os.Mkdir(filepath.Join(m.baseDir, ".restore-beta-"+strconv.Itoa(i)), 0700); err != nil {
			t.Fatal(err)
		}
	}
	live := m.GetWorkingDir(srv.ID)
	copies := make(map[string]string)
	for i := 0; i < restoreMaxRetained; i++ {
		marker := "original generation " + strconv.Itoa(i)
		if err := os.WriteFile(filepath.Join(live, "marker"), []byte(marker), 0600); err != nil {
			t.Fatal(err)
		}
		result, err := m.RestoreBackup(context.Background(), srv, "small.tar.gz")
		if err != nil {
			t.Fatalf("restore %d: %v", i+1, err)
		}
		copies[filepath.Join(m.baseDir, result.SafetyCopy, "marker")] = marker
	}
	if err := os.WriteFile(filepath.Join(live, "marker"), []byte("live must survive"), 0600); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadDir(m.baseDir)
	if err != nil {
		t.Fatal(err)
	}
	result, err := m.RestoreBackup(context.Background(), srv, "small.tar.gz")
	if !errors.Is(err, ErrRestoreRetention) || !strings.Contains(err.Error(), "administrator cleanup required") || result.SafetyCopy != "" {
		t.Fatalf("fourth restore should require cleanup: %+v, %v", result, err)
	}
	after, err := os.ReadDir(m.baseDir)
	if err != nil || len(before) != len(after) {
		t.Fatalf("denied restore changed recovery parent: %v", err)
	}
	for i := range before {
		if before[i].Name() != after[i].Name() {
			t.Fatal("denied restore changed recovery directories")
		}
	}
	copies[filepath.Join(live, "marker")] = "live must survive"
	copies[filepath.Join(live, "restored")] = "backup data"
	for path, want := range copies {
		if got, err := os.ReadFile(path); err != nil || string(got) != want {
			t.Fatalf("data not preserved at %s: %q, %v", path, got, err)
		}
	}
}

func TestRestoreRetentionUnsafeEntriesFailClosed(t *testing.T) {
	for _, kind := range []string{"file", "symlink", "dangling-symlink"} {
		t.Run(kind, func(t *testing.T) {
			m, srv, _ := restoreFixture(t)
			entry := filepath.Join(m.baseDir, ".restore-"+srv.ID+"-ambiguous")
			outside := t.TempDir()
			sentinel := filepath.Join(outside, "sentinel")
			if err := os.WriteFile(sentinel, []byte("untouched"), 0600); err != nil {
				t.Fatal(err)
			}
			if kind == "file" {
				if err := os.WriteFile(entry, []byte("retained"), 0600); err != nil {
					t.Fatal(err)
				}
			} else {
				target := outside
				if kind == "dangling-symlink" {
					target = filepath.Join(outside, "missing")
				}
				if err := os.Symlink(target, entry); err != nil {
					t.Skipf("symlink unavailable: %v", err)
				}
			}
			// Retention is checked before even opening the requested archive.
			if _, err := m.RestoreBackup(context.Background(), srv, "missing.tar.gz"); !errors.Is(err, ErrRestoreRetention) {
				t.Fatalf("unsafe recovery entry accepted: %v", err)
			}
			for path, want := range map[string]string{
				sentinel: "untouched",
				filepath.Join(m.GetWorkingDir(srv.ID), "original"): "keep me",
			} {
				if got, err := os.ReadFile(path); err != nil || string(got) != want {
					t.Fatalf("data changed at %s: %q, %v", path, got, err)
				}
			}
			if _, err := os.Lstat(entry); err != nil {
				t.Fatalf("ambiguous recovery entry deleted: %v", err)
			}
		})
	}
}
