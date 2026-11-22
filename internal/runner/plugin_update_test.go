package runner

import (
	"context"
	"testing"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

func TestPluginUpdateLifecycleGates(t *testing.T) {
	m, srv, _ := restoreFixture(t)
	ctx, release, err := m.BeginPluginUpdateOperation(context.Background(), srv.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.beginOperation(srv.ID); err == nil {
		t.Fatal("operation gate bypassed")
	}
	if unlock, err := rooted.TryMutation(m.GetWorkingDir(srv.ID)); err == nil {
		unlock()
		t.Fatal("file gate bypassed")
	}
	// No power calls/process launches: exercise exactly the gate power uses.
	if unlock, err := rooted.TryExclusiveMutation(m.GetWorkingDir(srv.ID)); err == nil {
		unlock()
		t.Fatal("exclusive gate bypassed")
	}
	release()
	release()
	if ctx.Err() == nil {
		t.Fatal("operation context not canceled")
	}
	if unlock, err := rooted.TryMutation(m.GetWorkingDir(srv.ID)); err != nil {
		t.Fatal(err)
	} else {
		unlock()
	}
	m.BlockPluginUpdateRecovery(srv.ID)
	if unlock, err := m.beginOperation(srv.ID); err == nil {
		unlock()
		t.Fatal("recovery did not gate startup")
	}
}

func TestPluginUpdateRefusesRunningAndRunLock(t *testing.T) {
	for _, status := range []string{"online", "starting", "stopping", "offline"} {
		t.Run(status, func(t *testing.T) {
			m, srv, _ := restoreFixture(t)
			inst := NewProcessInstance(srv, m.GetWorkingDir(srv.ID), DefaultStartupTimeout, nil, nil)
			m.instances[srv.ID] = inst
			inst.status = status
			defer func() { inst.status = "offline" }()
			if status == "offline" {
				lock, err := acquireRunLock(m.GetWorkingDir(srv.ID))
				if err != nil {
					t.Fatal(err)
				}
				defer lock.Close()
			}
			_, release, err := m.BeginPluginUpdateOperation(context.Background(), srv.ID)
			if err == nil {
				release()
				t.Fatal("accepted running server or held run lock")
			}
			if inst.status != status {
				t.Fatal("operation changed process state")
			}
		})
	}
}

func TestPluginUpdateShutdownContextIsJoined(t *testing.T) {
	m, srv, _ := restoreFixture(t)
	ctx, release, err := m.BeginPluginUpdateOperation(context.Background(), srv.ID)
	if err != nil {
		t.Fatal(err)
	}
	// Use the manager's shutdown signal without launching or stopping a process.
	m.shutdownOnce.Do(func() { close(m.shutdownCh) })
	<-ctx.Done()
	release()
	m.importWG.Wait()
	// No live instances were created; cleanup's shutdownOnce is already complete.
}
