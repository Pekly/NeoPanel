package runner

import (
	"context"
	"fmt"
	"time"
)

// BeginServerOperation excludes restore/power/backup for the entire caller's
// filesystem and metadata operation. The caller must defer the returned release.
func (m *ServerManager) BeginServerOperation(id string) (func(), error) {
	return m.beginOperation(id)
}

// BeginServerMaintenance stops the process before destructive work and retains
// the operation gate until the caller has finished ALL filesystem/DB changes.
// It deliberately uses the instance directly, not the guarded public Stop.
func (m *ServerManager) BeginServerMaintenance(ctx context.Context, id string) (func(), error) {
	release, err := m.beginOperation(id)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	m.cancelRetryLocked(id)
	m.restorePaused.Store(id, true)
	inst := m.instances[id]
	m.mu.Unlock()
	if inst != nil && inst.GetStatus() != "offline" {
		if err := inst.Stop(false); err != nil {
			release()
			return nil, err
		}
		timer := time.NewTimer(15 * time.Second)
		defer timer.Stop()
		tick := time.NewTicker(25 * time.Millisecond)
		defer tick.Stop()
		for inst.GetStatus() != "offline" {
			select {
			case <-ctx.Done():
				release()
				return nil, ctx.Err()
			case <-m.shutdownCh:
				release()
				return nil, fmt.Errorf("daemon shutting down")
			case <-timer.C:
				release()
				return nil, fmt.Errorf("server did not stop; no files removed")
			case <-tick.C:
			}
		}
	}
	if err := ctx.Err(); err != nil {
		release()
		return nil, err
	}
	return release, nil
}

// DeleteServer keeps the gate through runtime removal and the caller's complete
// disk/metadata cleanup. Cleanup must not call other guarded manager methods.
func (m *ServerManager) DeleteServer(ctx context.Context, id string, cleanup func() error) error {
	release, err := m.BeginServerMaintenance(ctx, id)
	if err != nil {
		return err
	}
	defer release()
	m.removeInstanceLocked(id)
	return cleanup()
}
