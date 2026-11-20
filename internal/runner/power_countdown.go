package runner

import (
	"fmt"

	"github.com/Pekly/NeoPanel/internal/database"
)

// PowerRun identifies one online process lifetime. Keeping the instance pointer
// also distinguishes deletion/recreation, where a new instance's runID resets.
type PowerRun struct {
	instance   *ProcessInstance
	generation uint64
}

func (m *ServerManager) GetPowerRun(serverID string) (PowerRun, bool) {
	if m.isShuttingDown() {
		return PowerRun{}, false
	}
	m.mu.RLock()
	inst := m.instances[serverID]
	m.mu.RUnlock()
	if inst == nil {
		return PowerRun{}, false
	}
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return PowerRun{inst, inst.runID}, inst.status == "online"
}

// CountdownPower checks the original process lifetime under the same operation
// gate used by direct power, deletion and restore. In particular, a countdown
// must not turn a stopped server back on or affect a replacement process.
func (m *ServerManager) CountdownPower(srv *database.Server, action string, expected PowerRun) error {
	if action != "stop" && action != "restart" {
		return fmt.Errorf("invalid countdown action")
	}
	unlock, err := m.beginOperation(srv.ID)
	if err != nil {
		return err
	}
	defer unlock()
	current, online := m.GetPowerRun(srv.ID)
	if !online || current != expected {
		return fmt.Errorf("countdown expired: server process changed or stopped")
	}
	m.mu.Lock()
	m.cancelRetryLocked(srv.ID)
	if action == "stop" {
		m.restorePaused.Store(srv.ID, true)
	}
	m.mu.Unlock()
	if err := current.instance.Stop(true); err != nil {
		return err
	}
	if action == "restart" {
		return m.startLocked(srv)
	}
	return nil
}
