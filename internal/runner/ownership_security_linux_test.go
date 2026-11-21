//go:build linux

package runner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

// The dangerous operation runs only in a killable child, never the test runner.
func TestStartupFIFOHelper(t *testing.T) {
	base := os.Getenv("NEOPANEL_STARTUP_FIFO_ROOT")
	if base == "" {
		t.Skip("subprocess helper")
	}
	m := NewServerManager(base, nil)
	defer m.ShutdownAll()
	if err := m.Start(&database.Server{ID: "bad", Type: "paper"}); err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("FIFO start not rejected: %v", err)
	}
	// Both use the node-wide launch mutex; neither may remain blocked by bad.
	if err := m.Stop("other", false); err != nil {
		t.Fatal(err)
	}
	if err := m.Start(&database.Server{ID: "other", Type: "paper"}); err == nil || !strings.Contains(err.Error(), "invalid level-name") {
		t.Fatalf("second server did not reach its own startup check: %v", err)
	}
}

func TestStartupFIFORejectsWithoutHoldingNodeLaunchMutex(t *testing.T) {
	for _, target := range []string{"server.properties", "world/session.lock"} {
		t.Run(target, func(t *testing.T) {
			base := t.TempDir()
			fifo := filepath.Join(base, "bad", filepath.FromSlash(target))
			if err := os.MkdirAll(filepath.Dir(fifo), 0700); err != nil {
				t.Fatal(err)
			}
			if err := syscall.Mkfifo(fifo, 0600); err != nil {
				t.Fatal(err)
			}
			other := filepath.Join(base, "other")
			if err := os.Mkdir(other, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(other, "server.properties"), []byte("level-name=../outside\n"), 0600); err != nil {
				t.Fatal(err)
			}
			exe, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, exe, "-test.run=^TestStartupFIFOHelper$")
			cmd.Env = append(os.Environ(), "NEOPANEL_STARTUP_FIFO_ROOT="+base)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("startup FIFO child failed (timeout=%v): %v\n%s", ctx.Err(), err, out)
			}
		})
	}
}
