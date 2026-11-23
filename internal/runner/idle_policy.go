package runner

import (
	"fmt"
	"log"
	"time"
)

func (m *ServerManager) checkIdleTimeouts() {
	if m.db == nil || m.isShuttingDown() {
		return
	}
	policy, err := m.db.GetNodeIdlePolicy("local")
	if err != nil || !policy.Enabled {
		return
	}
	timeout := time.Duration(policy.TimeoutMinutes) * time.Minute
	m.mu.RLock()
	instances := make(map[string]*ProcessInstance, len(m.instances))
	for id, inst := range m.instances {
		instances[id] = inst
	}
	m.mu.RUnlock()
	for id, inst := range instances {
		if m.isShuttingDown() {
			return
		}
		inst.mu.RLock()
		runID := inst.runID
		local := inst.server != nil && (inst.server.NodeID == "" || inst.server.NodeID == "local")
		inst.mu.RUnlock()
		if !local {
			continue
		}
		duration, idle := inst.GetIdleDuration()
		if !idle || duration < timeout {
			continue
		}
		m.stopIdleServer(id, inst, runID)
	}
}

// Preserve the ordinary operation gate and re-read policy/inventory after
// contention. The process performs the final run/player/time check atomically
// with its transition to stopping, so an old scan cannot stop a replacement run.
func (m *ServerManager) stopIdleServer(id string, expected *ProcessInstance, runID uint64) {
	release, err := m.beginOperation(id)
	if err != nil {
		return
	}
	defer release()
	if m.isShuttingDown() {
		return
	}
	inst, ok := m.GetInstance(id)
	if !ok || inst != expected {
		return
	}
	srv, err := m.db.GetServer(id)
	if err != nil || (srv.NodeID != "" && srv.NodeID != "local") {
		return
	}
	policy, err := m.db.GetNodeIdlePolicy("local")
	if err != nil || !policy.Enabled {
		return
	}
	timeout := time.Duration(policy.TimeoutMinutes) * time.Minute
	stopped, err := inst.stopIfIdle(runID, timeout, m.isShuttingDown)
	if !stopped {
		return
	}
	m.mu.Lock()
	m.cancelRetryLocked(id)
	m.restorePaused.Store(id, true)
	m.mu.Unlock()
	if err != nil {
		log.Printf("[WARN] Idle auto-shutdown failed for %s: %v", id, err)
		inst.broadcast("[NeoPanel] Idle auto-shutdown failed; check process cleanup before retrying.")
		_ = m.db.CreateAuditLog("system", "server.idle_timeout_failed", fmt.Sprintf("Idle stop failed for server %s: %v", id, err), "")
		return
	}
	// stopIfIdle waits for run completion including owned-process cleanup.
	// Serialize publication with launches; the operation gate excludes manual stop.
	m.launchMu.Lock()
	if !m.isShuttingDown() && inst.GetStatus() == "offline" {
		if err := m.db.SetServerIdleStopped(id, true); err != nil {
			m.wakeSuppressed.Store(id, true)
			log.Printf("[WARN] Cannot persist idle wake authority for %s: %v", id, err)
		} else {
			m.wakeSuppressed.Delete(id)
		}
	}
	m.launchMu.Unlock()
	inst.broadcast(fmt.Sprintf("[NeoPanel] Idle auto-shutdown completed after at least %d minutes with no tracked players.", policy.TimeoutMinutes))
	_ = m.db.CreateAuditLog("system", "server.idle_timeout", fmt.Sprintf("Server %s auto-stopped after at least %d minutes idle", srv.Name, policy.TimeoutMinutes), "")
}

func (p *ProcessInstance) stopIfIdle(runID uint64, timeout time.Duration, shuttingDown func() bool) (bool, error) {
	stopped := false
	err := p.stopConditional(true, func() bool {
		if shuttingDown != nil && shuttingDown() {
			return false
		}
		if p.runID != runID || p.cmd == nil || p.status != "online" || len(p.onlinePlayers) > 0 || p.unlistedPlayers > 0 || p.zeroPlayersSince.IsZero() || timeout <= 0 || time.Since(p.zeroPlayersSince) < timeout {
			return false
		}
		stopped = true
		return true
	})
	return stopped, err
}

func (m *ServerManager) checkDiskLimits() {
	if m.isShuttingDown() {
		return
	}
	m.mu.RLock()
	instances := make(map[string]*ProcessInstance, len(m.instances))
	for id, inst := range m.instances {
		instances[id] = inst
	}
	m.mu.RUnlock()

	for id, inst := range instances {
		if m.isShuttingDown() {
			return
		}
		inst.mu.RLock()
		srv := inst.server
		cachedDiskMB := inst.cachedDiskMB
		status := inst.status
		inst.mu.RUnlock()

		if srv == nil || srv.DiskLimitMB <= 0 || (status != "online" && status != "starting") {
			continue
		}

		if cachedDiskMB > srv.DiskLimitMB {
			inst.broadcast(fmt.Sprintf("[NeoPanel] WARNING: Disk limit exceeded (%d MB used / %d MB limit). Auto-saving world.", cachedDiskMB, srv.DiskLimitMB))
			_ = inst.WriteCommand("save-all")

			if cachedDiskMB > (srv.DiskLimitMB * 115 / 100) {
				inst.broadcast(fmt.Sprintf("[NeoPanel] CRITICAL: Disk limit severely exceeded (%d MB used / %d MB limit). Halting server to protect host disk.", cachedDiskMB, srv.DiskLimitMB))
				log.Printf("server %s exceeded disk quota (%d MB > %d MB limit); halting instance", id, cachedDiskMB, srv.DiskLimitMB)
				go func(target *ProcessInstance) {
					_ = target.Stop(true)
				}(inst)
			}
		}
	}
}
