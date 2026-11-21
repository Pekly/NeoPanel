package runner

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

func assertIdleWake(t *testing.T, m *ServerManager, id string, want bool) {
	t.Helper()
	state, err := m.db.GetServerWakeState(id)
	if err != nil || state.StoppedByIdlePolicy != want {
		t.Fatal(state, err)
	}
}

func TestWakeOnlyCompletedIdleStopAndPreference(t *testing.T) {
	m, p, s := idleManagerFixture(t)
	s.Port = wakeFreePort(t)
	if err := m.db.UpdateServer(s); err != nil {
		t.Fatal(err)
	}
	cfg := database.DefaultNodeWakeConfig()
	cfg.Enabled = true
	if err := m.db.SetNodeWakeConfig("local", cfg); err != nil {
		t.Fatal(err)
	}
	// A disabled preference must not disable the node's enforced idle shutdown.
	if err := m.SetServerWakePreference(s.ID, false); err != nil {
		t.Fatal(err)
	}
	m.checkIdleTimeouts()
	if p.GetStatus() != "offline" {
		t.Fatal("preference disabled idle enforcement")
	}
	assertIdleWake(t, m, s.ID, true)
	if m.wakeEligible(s) {
		t.Fatal("disabled preference woke")
	}
	if err := m.SetServerWakePreference(s.ID, true); err != nil {
		t.Fatal(err)
	}
	if !m.wakeEligible(s) {
		t.Fatal("completed idle stop not eligible")
	}
	if err := m.db.SetNodeIdlePolicy("local", database.NodeIdlePolicy{TimeoutMinutes: 1}); err != nil {
		t.Fatal(err)
	}
	if m.wakeEligible(s) {
		t.Fatal("node idle off still eligible")
	}
	if err := m.db.SetNodeIdlePolicy("local", database.NodeIdlePolicy{Enabled: true, TimeoutMinutes: 1}); err != nil {
		t.Fatal(err)
	}
	if err := m.Stop(s.ID, true); err != nil {
		t.Fatal(err)
	}
	assertIdleWake(t, m, s.ID, false)
	if m.wakeEligible(s) {
		t.Fatal("manual stop remained eligible")
	}
	// User preference does not restrict ordinary manual power.
	if err := m.SetServerWakePreference(s.ID, false); err != nil {
		t.Fatal(err)
	}
	if err := m.Start(s); err != nil {
		t.Fatal(err)
	}
	waitProcessStatus(t, p, "online")
}

func TestWakeFreshManualCrashAndShutdownNeverMark(t *testing.T) {
	for _, kind := range []string{"fresh", "manual", "crash", "shutdown"} {
		t.Run(kind, func(t *testing.T) {
			m, s, _ := wakeFixture(t)
			if err := m.db.SetServerIdleStopped(s.ID, false); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "manual":
				if err := m.Stop(s.ID, true); err != nil {
					t.Fatal(err)
				}
			case "crash":
				s.AutoRestart = false
				if err := m.db.UpdateServer(s); err != nil {
					t.Fatal(err)
				}
				t.Setenv("NEOPANEL_LIFECYCLE_HELPER", "tail")
				if err := m.Start(s); err != nil {
					t.Fatal(err)
				}
				p, _ := m.GetInstance(s.ID)
				waitProcessStatus(t, p, "offline")
			case "shutdown":
				m.ShutdownAll()
			}
			assertIdleWake(t, m, s.ID, false)
			if m.wakeEligible(s) {
				t.Fatal("non-idle offline server eligible")
			}
		})
	}
}

func TestWakeMarkerSurvivesCleanRestartButNotFailedStart(t *testing.T) {
	for _, automatic := range []bool{false, true} {
		t.Run(fmt.Sprint(automatic), func(t *testing.T) {
			m, s, _ := wakeFixture(t)
			m.ShutdownAll()
			next := NewServerManager(m.baseDir, m.db)
			t.Cleanup(next.ShutdownAll)
			if !next.wakeEligible(s) {
				t.Fatal("trusted idle stop lost at restart")
			}
			s.JavaPath = filepath.Join(t.TempDir(), "missing-java")
			if err := m.db.UpdateServer(s); err != nil {
				t.Fatal(err)
			}
			if automatic {
				next.StartWakeListeners()
				wakeWait(t, func() bool { return wakeEntry(next, s.ID) != nil })
				e := wakeEntry(next, s.ID)
				if !next.reserveWake(e) {
					t.Fatal("not reserved")
				}
				next.launchWake(e)
			} else if err := next.Start(s); err == nil {
				t.Fatal("missing java started")
			}
			assertIdleWake(t, next, s.ID, false)
			next.ShutdownAll()
			last := NewServerManager(m.baseDir, m.db)
			t.Cleanup(last.ShutdownAll)
			if last.wakeEligible(s) {
				t.Fatal("failed launch replayed after restart")
			}
		})
	}
}

func TestWakeQueuedJoinManualStopAndBusyGate(t *testing.T) {
	m, s, _ := wakeFixture(t)
	m.StartWakeListeners()
	wakeWait(t, func() bool { return wakeEntry(m, s.ID) != nil })
	e := wakeEntry(m, s.ID)
	if !m.reserveWake(e) {
		t.Fatal("not reserved")
	}
	release, err := m.beginOperation(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Stop(s.ID, true); !errors.Is(err, ErrServerOperation) {
		t.Fatal("busy stop reported success", err)
	}
	release()
	if err := m.Stop(s.ID, true); err != nil {
		t.Fatal(err)
	}
	assertIdleWake(t, m, s.ID, false)
	m.launchWake(e)
	if m.GetStatus(s.ID) != "offline" || wakeEntry(m, s.ID) != nil {
		t.Fatal("stale queued join survived manual stop")
	}
}

func TestWakeRestoreAndMaintenanceInvalidate(t *testing.T) {
	for _, kind := range []string{"restore", "maintenance"} {
		t.Run(kind, func(t *testing.T) {
			m, s, _ := wakeFixture(t)
			if kind == "restore" {
				_, _ = m.RestoreBackup(context.Background(), s, "missing.tar.gz")
			} else {
				release, err := m.BeginServerMaintenance(context.Background(), s.ID)
				if err != nil {
					t.Fatal(err)
				}
				release()
			}
			assertIdleWake(t, m, s.ID, false)
			if m.wakeEligible(s) {
				t.Fatal("destructive operation retained wake authority")
			}
		})
	}
}

func TestWakePolicyReadAndClearErrorsFailClosed(t *testing.T) {
	m, s, _ := wakeFixture(t)
	if err := m.db.SetPanelSetting("node_idle_policy:local", `{"enabled":true}`); err != nil {
		t.Fatal(err)
	}
	if m.wakeEligible(s) {
		t.Fatal("invalid node policy enabled wake")
	}
	if err := m.db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := m.Start(s); err == nil {
		t.Fatal("start proceeded without durable consume")
	}
	if _, suppressed := m.wakeSuppressed.Load(s.ID); !suppressed {
		t.Fatal("failed clear not suppressed")
	}
	if m.wakeEligible(s) {
		t.Fatal("read error enabled wake")
	}
}

func TestWakePreferenceClosesListenerAndCancelsQueuedJoin(t *testing.T) {
	m, s, _ := wakeFixture(t)
	m.StartWakeListeners()
	wakeWait(t, func() bool { return wakeEntry(m, s.ID) != nil })
	e := wakeEntry(m, s.ID)
	if !m.reserveWake(e) {
		t.Fatal("not reserved")
	}
	if err := m.SetServerWakePreference(s.ID, false); err != nil {
		t.Fatal(err)
	}
	if wakeEntry(m, s.ID) != nil {
		t.Fatal("preference disable did not close immediately")
	}
	assertIdleWake(t, m, s.ID, true)
	m.launchWake(e)
	if m.GetStatus(s.ID) != "offline" {
		t.Fatal("queued join bypassed disabled preference")
	}
	m.reconcileWakeListeners()
	if wakeEntry(m, s.ID) != nil {
		t.Fatal("disabled preference reopened")
	}
	if err := m.SetServerWakePreference(s.ID, true); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(1500 * time.Millisecond)
	for wakeEntry(m, s.ID) == nil && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if wakeEntry(m, s.ID) == nil {
		t.Fatal("preference enable not reconciled within one tick")
	}
	if m.reserveWake(e) {
		t.Fatal("old generation resurrected")
	}
}
