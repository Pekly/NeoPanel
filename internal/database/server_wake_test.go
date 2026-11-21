package database

import (
	"path/filepath"
	"testing"
)

func TestServerWakePersistenceDeletionAndIsolation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "panel.db")
	db, err := InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{ID: "sleeping", Name: "Sleeping", Type: "paper", NodeID: "local"}
	if err := db.CreateServer(s); err != nil {
		t.Fatal(err)
	}
	check := func(pref, idle bool) {
		t.Helper()
		got, err := db.GetServerWakeState(s.ID)
		if err != nil || got.JoinToStartEnabled != pref || got.StoppedByIdlePolicy != idle {
			t.Fatal(got, err)
		}
	}
	check(true, false)
	if err := db.SetServerIdleStopped(s.ID, true); err != nil {
		t.Fatal(err)
	}
	if err := db.SetServerWakePreference(s.ID, false); err != nil {
		t.Fatal(err)
	}
	check(false, true)
	db.Close()
	db, err = InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	check(false, true)
	if err := db.SetServerWakePreference(s.ID, true); err != nil {
		t.Fatal(err)
	}
	check(true, true)
	if err := db.DeleteServer(s.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetServerWakeState(s.ID); err == nil {
		t.Fatal("deleted state readable")
	}
	if err := db.CreateServer(s); err != nil {
		t.Fatal(err)
	}
	check(true, false)
	if err := db.SetServerIdleStopped(s.ID, true); err != nil {
		t.Fatal(err)
	}
	s.NodeID = "worker"
	if err := db.UpdateServer(s); err != nil {
		t.Fatal(err)
	}
	if err := db.SetServerIdleStopped(s.ID, true); err == nil {
		t.Fatal("foreign metadata became local authority")
	}
	s.NodeID = "local"
	if err := db.UpdateServer(s); err != nil {
		t.Fatal(err)
	}
	check(true, false)
	if err := db.SetServerWakePreference("missing", false); err == nil {
		t.Fatal("orphan preference")
	}
}

func TestServerWakeInvalidStateFailsClosed(t *testing.T) {
	db, err := InitDB(filepath.Join(t.TempDir(), "panel.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.CreateServer(&Server{ID: "test", Name: "Test"}); err != nil {
		t.Fatal(err)
	}
	if err := db.SetServerIdleStopped("test", true); err != nil {
		t.Fatal(err)
	}
	if _, err := db.conn.Exec(`UPDATE server_wake_state SET stopped_by_idle_policy=2`); err == nil {
		t.Fatal("constraint allowed invalid marker")
	}
	if _, err := db.conn.Exec(`PRAGMA ignore_check_constraints=ON; UPDATE server_wake_state SET stopped_by_idle_policy=2`); err != nil {
		t.Fatal(err)
	}
	if state, err := db.GetServerWakeState("test"); err == nil || state.StoppedByIdlePolicy {
		t.Fatal("invalid state accepted", state, err)
	}
}
