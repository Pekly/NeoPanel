//go:build windows

package runner

import (
	"os"
	"sync"
	"testing"
	"time"
)

func TestWindowsProcessMetrics(t *testing.T) {
	dir := t.TempDir()
	c, err := startOwnedCommand(jobHelperCommand("busy", dir), nil, 256)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = c.stop(false); _, _ = c.wait() }()
	awaitJobFile(t, dir, "ready")
	pm := NewProcessMetrics(c.cmd.Process.Pid)
	time.Sleep(400 * time.Millisecond)
	snap, err := pm.ReadStats()
	if err != nil {
		t.Fatal(err)
	}
	if snap.MemoryMB < 32 || snap.CPUPercent <= 0 || snap.CPUPercent > 100 {
		t.Fatalf("not real workload metrics: %+v", snap)
	}
	if snap.NetworkSupported || snap.NetInKiB != 0 || snap.NetOutKiB != 0 {
		t.Fatal("fabricated network", snap)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := pm.ReadStats(); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	pm.processCreated++
	if snap, err := pm.ReadStats(); err == nil || snap != (StatsSnapshot{}) {
		t.Fatal("PID reuse accepted", snap, err)
	}
}

func TestWindowsUnavailableMetrics(t *testing.T) {
	for _, pid := range []int{0, -1, 2147483647} {
		snap, err := NewProcessMetrics(pid).ReadStats()
		if snap != (StatsSnapshot{}) {
			t.Fatal("fabricated unavailable metrics", snap)
		}
		if pid > 0 && err == nil {
			t.Fatal("nonexistent PID accepted")
		}
	}
	pm := NewProcessMetrics(os.Getpid())
	if snap, err := pm.ReadStats(); err != nil || snap.MemoryMB <= 0 {
		t.Fatal(snap, err)
	}
}
