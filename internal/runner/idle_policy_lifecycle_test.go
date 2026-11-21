package runner

import (
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func idleManagerFixture(t *testing.T) (*ServerManager, *ProcessInstance, *database.Server) {
	t.Helper()
	t.Setenv("NEOPANEL_LIFECYCLE_HELPER", "graceful")
	db, base, s, _ := restartFixture(t)
	s.AutoRestart = true
	if err := db.UpdateServer(s); err != nil {
		t.Fatal(err)
	}
	m := NewServerManager(base, db)
	t.Cleanup(m.ShutdownAll)
	p := m.getOrCreateInstance(s)
	p.SetIsolationConfig("process", 512)
	if err := m.Start(s); err != nil {
		t.Fatal(err)
	}
	waitProcessStatus(t, p, "online")
	if err := db.SetNodeIdlePolicy("local", database.NodeIdlePolicy{Enabled: true, TimeoutMinutes: 1}); err != nil {
		t.Fatal(err)
	}
	p.mu.Lock()
	p.zeroPlayersSince = time.Now().Add(-2 * time.Minute)
	p.mu.Unlock()
	return m, p, s
}

func TestIdleWatchdogDisabledOccupiedAndGracefulStop(t *testing.T) {
	m, p, s := idleManagerFixture(t)
	policy := database.NodeIdlePolicy{Enabled: false, TimeoutMinutes: 1}
	if err := m.db.SetNodeIdlePolicy("local", policy); err != nil {
		t.Fatal(err)
	}
	m.checkIdleTimeouts()
	if p.GetStatus() != "online" {
		t.Fatal("off toggle with positive timeout stopped server")
	}
	policy.Enabled = true
	if err := m.db.SetNodeIdlePolicy("local", policy); err != nil {
		t.Fatal(err)
	}
	p.mu.RLock()
	run := p.runID
	p.mu.RUnlock()
	p.trackPlayerLine("Player joined the game", run)
	m.checkIdleTimeouts()
	if p.GetStatus() != "online" {
		t.Fatal("occupied server stopped")
	}
	p.trackPlayerLine("Player left the game", run)
	m.checkIdleTimeouts()
	if p.GetStatus() != "online" {
		t.Fatal("stopped before full timeout")
	}
	p.mu.Lock()
	p.zeroPlayersSince = time.Now().Add(-2 * time.Minute)
	p.mu.Unlock()
	// Repeated player-list refreshes must not keep an empty game running forever.
	for i := 0; i < 5; i++ {
		p.trackPlayerLine("There are 0 of a max of 20 players online:", run)
	}
	m.checkIdleTimeouts()
	if p.GetStatus() != "offline" {
		t.Fatal("idle policy did not stop server")
	}
	logs := strings.Join(p.GetLogs(), "\n")
	if !strings.Contains(logs, "Sending 'stop' command") || !strings.Contains(logs, "FINAL-OUTPUT-PRESERVED") || strings.Contains(logs, "Terminating the owned") {
		t.Fatal("idle shutdown was not graceful", logs)
	}
	m.mu.RLock()
	retry := m.retryCancels[s.ID]
	m.mu.RUnlock()
	if retry != nil {
		t.Fatal("idle stop scheduled auto-restart")
	}
	if _, paused := m.restorePaused.Load(s.ID); !paused {
		t.Fatal("idle stop did not suppress late retries")
	}
}

func TestIdleWatchdogSkipsBusyBlockedAndForeignInventory(t *testing.T) {
	m, p, s := idleManagerFixture(t)
	release, err := m.beginOperation(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	m.checkIdleTimeouts()
	release()
	if p.GetStatus() != "online" {
		t.Fatal("busy operation bypassed")
	}
	m.restoreBlocked.Store(s.ID, ErrRestoreRecovery)
	m.checkIdleTimeouts()
	m.restoreBlocked.Delete(s.ID)
	if p.GetStatus() != "online" {
		t.Fatal("blocked server stopped")
	}
	s.NodeID = "remote"
	if err := m.db.UpdateServer(s); err != nil {
		t.Fatal(err)
	}
	m.UpdateInstanceServer(s)
	m.checkIdleTimeouts()
	if p.GetStatus() != "online" {
		t.Fatal("foreign server affected by local policy")
	}
	s.NodeID = "local"
	if err := m.db.UpdateServer(s); err != nil {
		t.Fatal(err)
	}
	m.UpdateInstanceServer(s)
	m.checkIdleTimeouts()
	if p.GetStatus() != "offline" {
		t.Fatal("eligible idle server not stopped")
	}
}

func TestIdleFinalStopRechecksPlayerAndRun(t *testing.T) {
	m, p, _ := idleManagerFixture(t)
	p.mu.RLock()
	run := p.runID
	p.mu.RUnlock()
	if stopped, err := p.stopIfIdle(run+1, time.Minute, m.isShuttingDown); stopped || err != nil {
		t.Fatal("stale run stop", stopped, err)
	}
	p.lifecycleMu.Lock()
	done := make(chan bool, 1)
	go func() {
		stopped, err := p.stopIfIdle(run, time.Minute, m.isShuttingDown)
		done <- stopped || err != nil
	}()
	p.trackPlayerLine("JustJoined joined the game", run)
	p.lifecycleMu.Unlock()
	if <-done {
		t.Fatal("player joined before final stop guard but server stopped")
	}
	if p.GetStatus() != "online" {
		t.Fatal("status changed on rejected idle stop")
	}
	if stopped, err := p.stopIfIdle(run, time.Minute, func() bool { return true }); stopped || err != nil {
		t.Fatal("stop accepted during daemon shutdown")
	}
}
