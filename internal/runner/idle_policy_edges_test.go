package runner

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestIdleFailedStopNeverLogsSuccess(t *testing.T) {
	db, base, s, _ := restartFixture(t)
	m := NewServerManager(base, db)
	t.Cleanup(m.ShutdownAll)
	p := m.getOrCreateInstance(s)
	// No OS process: model a completion/cleanup failure at the stop boundary.
	p.mu.Lock()
	p.status = "online"
	p.runID = 1
	p.cmd = &exec.Cmd{}
	p.zeroPlayersSince = time.Now().Add(-2 * time.Minute)
	p.mu.Unlock()
	if err := db.SetNodeIdlePolicy("local", database.NodeIdlePolicy{Enabled: true, TimeoutMinutes: 1}); err != nil {
		t.Fatal(err)
	}
	m.checkIdleTimeouts()
	p.mu.Lock()
	p.cmd = nil
	p.status = "offline"
	p.mu.Unlock()
	rows, err := db.ListActivity(0, true, 50)
	if err != nil {
		t.Fatal(err)
	}
	failed := 0
	for _, row := range rows {
		if row.Action == "server.idle_timeout" {
			t.Fatal("failed stop audited as completed")
		}
		if row.Action == "server.idle_timeout_failed" {
			failed++
		}
	}
	if failed != 1 {
		t.Fatal("failure not audited", failed)
	}
}

func TestIdleStopAllowsEnabledJoinToStartToReturn(t *testing.T) {
	m, s, _ := wakeFixture(t)
	t.Setenv("NEOPANEL_LIFECYCLE_HELPER", "graceful")
	if err := m.Start(s); err != nil {
		t.Fatal(err)
	}
	p, _ := m.GetInstance(s.ID)
	waitProcessStatus(t, p, "online")
	if err := m.db.SetNodeIdlePolicy("local", database.NodeIdlePolicy{Enabled: true, TimeoutMinutes: 1}); err != nil {
		t.Fatal(err)
	}
	p.mu.Lock()
	p.zeroPlayersSince = time.Now().Add(-2 * time.Minute)
	p.mu.Unlock()
	m.checkIdleTimeouts()
	if p.GetStatus() != "offline" {
		t.Fatal("idle stop failed")
	}
	m.StartWakeListeners()
	wakeWait(t, func() bool { return wakeEntry(m, s.ID) != nil })
	if message := wakeRequest(t, s, true); !strings.Contains(message, "reconnect") {
		t.Fatal(message)
	}
	waitProcessStatus(t, p, "online")
	if elapsed, idle := p.GetIdleDuration(); !idle || elapsed > time.Second {
		t.Fatal("woken server inherited expired idle interval", elapsed, idle)
	}
	m.checkIdleTimeouts()
	if p.GetStatus() != "online" {
		t.Fatal("freshly woken server immediately stopped")
	}
}
