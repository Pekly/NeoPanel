package runner

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestDiskLimitPreflightRejection(t *testing.T) {
	tempDir := t.TempDir()
	largeData := make([]byte, 2*1024*1024) // 2 MB
	if err := os.WriteFile(filepath.Join(tempDir, "filler.bin"), largeData, 0644); err != nil {
		t.Fatal(err)
	}

	// Disk limit is 1 MB, directory has 2 MB -> should refuse start
	srv := &database.Server{
		ID:          "disk-test-1",
		Type:        "paper",
		DiskLimitMB: 1,
	}
	p := NewProcessInstance(srv, tempDir, time.Second, nil, nil)
	err := p.Start()
	if err == nil {
		t.Fatal("expected Start() to fail due to disk limit, but got nil")
	}
	if !strings.Contains(err.Error(), "disk limit exceeded") {
		t.Fatalf("expected 'disk limit exceeded' error, got: %v", err)
	}

	// Raising limit to 10 MB -> disk check should pass (Start will then fail later due to missing jar, not disk quota)
	srv.DiskLimitMB = 10
	err = p.Start()
	if err != nil && strings.Contains(err.Error(), "disk limit exceeded") {
		t.Fatalf("did not expect disk limit error with 10 MB limit, got: %v", err)
	}
}

func TestDiskLimitWatchdogCheck(t *testing.T) {
	m := &ServerManager{
		instances:  make(map[string]*ProcessInstance),
		shutdownCh: make(chan struct{}),
	}
	srv := &database.Server{
		ID:          "srv-watchdog",
		Type:        "paper",
		DiskLimitMB: 50,
	}
	p := NewProcessInstance(srv, t.TempDir(), time.Second, nil, nil)
	done := make(chan struct{})
	p.mu.Lock()
	p.status = "online"
	p.cachedDiskMB = 60 // 60 MB > 50 MB (exceeds 115%)
	p.cmd = &exec.Cmd{}
	p.runDone = done
	p.mu.Unlock()

	m.instances[srv.ID] = p

	// Run checkDiskLimits: should trigger save-all and halt
	m.checkDiskLimits()

	// Wait up to 500ms for the stop goroutine to transition p.status or set p.stopExpected
	var stopped bool
	for i := 0; i < 50; i++ {
		p.mu.RLock()
		stopped = p.stopExpected
		p.mu.RUnlock()
		if stopped {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !stopped {
		t.Fatal("expected stopExpected to be true after critical disk limit breach")
	}

	// Also verify that the warning and critical broadcasts were added to logBuffer
	p.mu.RLock()
	logs := strings.Join(p.logBuffer, "\n")
	p.mu.RUnlock()
	if !strings.Contains(logs, "Disk limit exceeded") || !strings.Contains(logs, "CRITICAL: Disk limit severely exceeded") {
		t.Fatalf("expected warning and critical logs in logBuffer, got: %v", logs)
	}
}
