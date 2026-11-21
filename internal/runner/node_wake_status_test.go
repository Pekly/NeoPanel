package runner

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func statusCoordinator(m *ServerManager) {
	m.wake.mu.Lock()
	defer m.wake.mu.Unlock()
	m.wake.started = true
	m.wake.listeners = make(map[string]*wakeListener)
	m.wake.cooldown = make(map[string]time.Time)
	m.wake.ips = make(map[string]wakeIPBudget)
	m.wake.errors = make(map[string]string)
	m.wake.slots = make(chan struct{}, 64)
}

func statusRow(t *testing.T, v NodeWakeStatus, id, state string) {
	t.Helper()
	for _, row := range v.Servers {
		if row.ID == id {
			if row.State != state {
				t.Fatalf("want %s: %+v", state, v)
			}
			return
		}
	}
	t.Fatalf("missing server %s: %+v", id, v)
}

func TestWakeStatusBoundPendingDisabledAndShutdown(t *testing.T) {
	m, s, cfg := wakeFixture(t)
	v := m.GetNodeWakeStatus()
	if !v.CheckedAt.IsZero() || v.CoordinatorRunning || v.ConfigMatches || v.Listeners != 0 {
		t.Fatal(v)
	}
	statusRow(t, v, s.ID, "pending")
	// Merely reading status did not bind a socket.
	probe, err := net.Listen("tcp", wakeAddress(s))
	if err != nil {
		t.Fatal(err)
	}
	probe.Close()
	statusCoordinator(m)
	m.reconcileWakeListeners()
	v = m.GetNodeWakeStatus()
	if !v.CoordinatorRunning || !v.ConfigMatches || v.Listeners != 1 || v.ConfigHash != NodeWakeConfigHash(cfg) || v.CheckedAt.IsZero() {
		t.Fatal(v)
	}
	statusRow(t, v, s.ID, "listening")
	if probe, err = net.Listen("tcp", wakeAddress(s)); err == nil {
		probe.Close()
		t.Fatal("reported listener was not bound")
	}
	for i := 0; i < 3; i++ {
		if got := m.GetNodeWakeStatus(); got.CheckedAt != v.CheckedAt {
			t.Fatal("GET fabricated a reconciliation", got)
		}
	}
	if e := wakeEntry(m, s.ID); e == nil || e.queued {
		t.Fatal("status queued a player wake")
	}
	m.wake.mu.Lock()
	if len(m.wake.ips) != 0 {
		t.Error("status probed a socket")
	}
	m.wake.cooldown[s.ID] = time.Now().Add(time.Minute)
	m.wake.mu.Unlock()
	statusRow(t, m.GetNodeWakeStatus(), s.ID, "cooldown")
	cfg.Enabled = false
	if err := m.db.SetNodeWakeConfig("local", cfg); err != nil {
		t.Fatal(err)
	}
	v = m.GetNodeWakeStatus()
	if v.ConfigMatches || v.CheckedAt.IsZero() {
		t.Fatal(v)
	}
	statusRow(t, v, s.ID, "pending")
	m.reconcileWakeListeners()
	v = m.GetNodeWakeStatus()
	if v.Enabled || !v.ConfigMatches || v.Listeners != 0 {
		t.Fatal(v)
	}
	statusRow(t, v, s.ID, "disabled")
	stamp := v.CheckedAt
	m.ShutdownAll()
	v = m.GetNodeWakeStatus()
	if v.CoordinatorRunning || v.Listeners != 0 || v.CheckedAt != stamp {
		t.Fatal(v)
	}
}

func TestWakeStatusBindFailureBusyBlockedAndFreshness(t *testing.T) {
	m, s, _ := wakeFixture(t)
	occupied, err := net.Listen("tcp", wakeAddress(s))
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	statusCoordinator(m)
	m.reconcileWakeListeners()
	v := m.GetNodeWakeStatus()
	statusRow(t, v, s.ID, "bind_failed")
	if v.Listeners != 0 {
		t.Fatal(v)
	}
	for _, row := range v.Servers {
		if row.ID == s.ID && row.Detail != "Port unavailable or bind address not available; check node logs" {
			t.Fatal(row)
		}
	}
	occupied.Close()
	m.reconcileWakeListeners()
	e := wakeEntry(m, s.ID)
	gate, _ := m.operations.Load(s.ID)
	gate.(*sync.Mutex).Lock()
	m.reconcileWakeListeners()
	statusRow(t, m.GetNodeWakeStatus(), s.ID, "busy")
	if wakeEntry(m, s.ID) != e {
		t.Fatal("operation contention discarded accepted generation")
	}
	gate.(*sync.Mutex).Unlock()
	m.restoreBlocked.Store(s.ID, true)
	m.reconcileWakeListeners()
	statusRow(t, m.GetNodeWakeStatus(), s.ID, "blocked")
	m.restoreBlocked.Delete(s.ID)
	m.reconcileWakeListeners()
	m.wake.mu.Lock()
	m.wake.status.CheckedAt = time.Now().Add(-WakeStatusMaxAge - time.Second)
	stamp := m.wake.status.CheckedAt
	m.wake.mu.Unlock()
	v = m.GetNodeWakeStatus()
	if v.CoordinatorRunning || v.CheckedAt != stamp || !strings.Contains(v.Error, "recent") {
		t.Fatal(v)
	}
	statusRow(t, v, s.ID, "pending")
}

func TestWakeStatusInventoryStatesAndBound(t *testing.T) {
	m, s, cfg := wakeFixture(t)
	statusCoordinator(m)
	remote := *s
	remote.ID = "remote-only"
	remote.NodeID = "remote"
	conflict := *s
	conflict.ID = "duplicate"
	ineligible := *s
	ineligible.ID = "unsupported"
	ineligible.Type = "unknown"
	ineligible.Port++
	m.launchMu.Lock()
	m.wake.mu.Lock()
	m.captureWakeStatusLocked(cfg, []*database.Server{s, &remote, &conflict, &ineligible}, false)
	m.wake.mu.Unlock()
	m.launchMu.Unlock()
	v := m.GetNodeWakeStatus()
	if len(v.Servers) != 3 {
		t.Fatal(v)
	}
	statusRow(t, v, s.ID, "conflict")
	statusRow(t, v, ineligible.ID, "ineligible")
	servers := make([]*database.Server, WakeStatusMaxServers+3)
	for i := range servers {
		servers[i] = &database.Server{ID: fmt.Sprint(i), Name: strings.Repeat("X", 1000)}
	}
	m.launchMu.Lock()
	m.wake.mu.Lock()
	m.captureWakeStatusLocked(cfg, servers, false)
	m.wake.mu.Unlock()
	m.launchMu.Unlock()
	v = m.GetNodeWakeStatus()
	if len(v.Servers) != WakeStatusMaxServers || !strings.Contains(v.Error, "3 additional local servers omitted") || len(v.Servers[0].Name) > 128 {
		t.Fatal("unbounded report", len(v.Servers), v.Error)
	}
}

func TestWakeStatusProcessStatesAndReleasedSocket(t *testing.T) {
	m, s, _ := wakeFixture(t)
	statusCoordinator(m)
	inst := m.getOrCreateInstance(s)
	for _, state := range []string{"online", "starting", "stopping"} {
		inst.mu.Lock()
		inst.status = state
		inst.mu.Unlock()
		m.reconcileWakeListeners()
		want := state
		if state == "stopping" {
			want = "busy"
		}
		statusRow(t, m.GetNodeWakeStatus(), s.ID, want)
	}
	inst.mu.Lock()
	inst.status = "offline"
	inst.mu.Unlock()
	m.reconcileWakeListeners()
	statusRow(t, m.GetNodeWakeStatus(), s.ID, "listening")
	stamp := m.GetNodeWakeStatus().CheckedAt
	m.releaseWakeListener(s.ID)
	v := m.GetNodeWakeStatus()
	if v.Listeners != 0 || v.CheckedAt != stamp {
		t.Fatal(v)
	}
	statusRow(t, v, s.ID, "pending")
}
