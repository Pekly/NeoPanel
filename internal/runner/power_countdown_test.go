package runner

import (
	"errors"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestCountdownPowerPreservesOperationGateAndProcessLifetime(t *testing.T) {
	m := NewServerManager(t.TempDir(), nil)
	t.Cleanup(m.ShutdownAll)
	srv := &database.Server{ID: "countdown", Name: "countdown"}
	inst := NewProcessInstance(srv, t.TempDir(), 0, nil, nil)
	inst.status, inst.runID = "online", 1
	m.mu.Lock()
	m.instances[srv.ID] = inst
	m.mu.Unlock()
	// The instance is synthetic: no OS process is launched by these tests.
	t.Cleanup(func() { inst.mu.Lock(); inst.status = "offline"; inst.mu.Unlock() })
	run, online := m.GetPowerRun(srv.ID)
	if !online {
		t.Fatal("online instance not recognized")
	}
	unlock, err := m.beginOperation(srv.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{"stop", "restart"} {
		if err := m.CountdownPower(srv, action, run); err == nil {
			t.Fatal("bypassed operation gate")
		}
	}
	unlock()
	for _, status := range []string{"offline", "starting", "stopping"} {
		inst.mu.Lock()
		inst.status = status
		inst.mu.Unlock()
		if err := m.CountdownPower(srv, "restart", run); err == nil || !strings.Contains(err.Error(), "expired") {
			t.Fatal(status, err)
		}
	}
	inst.mu.Lock()
	inst.status, inst.runID = "online", 2
	inst.mu.Unlock()
	if err := m.CountdownPower(srv, "restart", run); err == nil {
		t.Fatal("affected a new process generation")
	}
	current, _ := m.GetPowerRun(srv.ID)
	if current == run {
		t.Fatal("generation identity did not change")
	}
	// A reused ID with a new ProcessInstance must not compare equal, either.
	replacement := NewProcessInstance(srv, t.TempDir(), 0, nil, nil)
	replacement.status, replacement.runID = "online", 1
	m.mu.Lock()
	m.instances[srv.ID] = replacement
	m.mu.Unlock()
	if err := m.CountdownPower(srv, "restart", run); err == nil {
		t.Fatal("affected replacement instance")
	}
	m.mu.Lock()
	delete(m.instances, srv.ID)
	m.mu.Unlock()
	if err := m.CountdownPower(srv, "restart", run); err == nil {
		t.Fatal("started deleted instance")
	}
	if err := m.CountdownPower(srv, "kill", run); err == nil {
		t.Fatal("accepted unsupported action")
	}
}

func TestCountdownPowerDoesNotRestartAfterFailedShutdown(t *testing.T) {
	m := NewServerManager(t.TempDir(), nil)
	t.Cleanup(m.ShutdownAll)
	srv := &database.Server{ID: "failed-stop"}
	inst := NewProcessInstance(srv, t.TempDir(), 0, nil, nil)
	inst.status, inst.runID, inst.cleanupErr = "online", 1, errors.New("retained process ownership")
	m.mu.Lock()
	m.instances[srv.ID] = inst
	m.mu.Unlock()
	t.Cleanup(func() { m.mu.Lock(); delete(m.instances, srv.ID); m.mu.Unlock() })
	run, _ := m.GetPowerRun(srv.ID)
	if err := m.CountdownPower(srv, "restart", run); err == nil {
		t.Fatal("restart ignored failed shutdown")
	}
}
