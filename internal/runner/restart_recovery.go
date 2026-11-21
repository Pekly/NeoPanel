package runner

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/Pekly/NeoPanel/internal/database"
)

var errDaemonShuttingDown = errors.New("daemon shutting down")

// RestoreRunningServers is called once after the daemon has bound its listener,
// before accepting control requests. Constructing a manager or failing to bind
// must not consume recovery state. This also arms the next clean shutdown save.
func (m *ServerManager) RestoreRunningServers() (result error) {
	m.recoveryOnce.Do(func() {
		defer func() { m.recoveryErr = result }()
		m.recoveryMu.Lock()
		defer m.recoveryMu.Unlock()
		if m.isShuttingDown() {
			result = errDaemonShuttingDown
			return
		}
		if m.db == nil {
			result = errors.New("restart recovery requires a database")
			return
		}
		ids, err := m.db.TakeServerResumeState(m.baseDir)
		if err != nil {
			result = fmt.Errorf("read restart recovery state: %w", err)
			return
		}
		m.recoveryEnabled = true
		m.recoveryPending = make(map[string]bool, len(ids))
		m.recoveryStarting = make(map[string]bool, len(ids))
		for _, id := range ids {
			m.recoveryPending[id] = true
		}
		for _, id := range ids {
			if m.isShuttingDown() {
				break
			}
			srv, err := m.db.GetServer(id)
			if err == nil {
				err = m.restoreRunningServer(srv)
			}
			if errors.Is(err, errDaemonShuttingDown) {
				break
			}
			delete(m.recoveryPending, id)
			if err != nil {
				log.Printf("[WARN] Could not restore previously running server %s after panel restart: %v; start it manually after resolving the error", id, err)
				_ = m.db.LogActivity(0, "server_resume_failed", fmt.Sprintf("Could not resume server %s after panel restart: %v", id, err), "system")
				continue
			}
			m.recoveryStarting[id] = true
			log.Printf("[INFO] Restored previously running server %s after panel restart", id)
			_ = m.db.LogActivity(0, "server_resumed", fmt.Sprintf("Resumed server %s after panel restart", id), "system")
		}
	})
	return m.recoveryErr
}

func (m *ServerManager) restoreRunningServer(srv *database.Server) error {
	unlock, err := m.beginOperation(srv.ID)
	if err != nil {
		return err
	}
	defer unlock()
	m.launchMu.Lock()
	defer m.launchMu.Unlock()
	if m.isShuttingDown() {
		return errDaemonShuttingDown
	}
	if _, paused := m.restorePaused.Load(srv.ID); paused {
		return errors.New("server was explicitly stopped or paused")
	}
	m.mu.RLock()
	halted := m.startupHalted[srv.ID]
	m.mu.RUnlock()
	if halted {
		return errors.New("server startup is halted")
	}
	if m.GetStatus(srv.ID) != "offline" {
		return errors.New("server is already active; duplicate start skipped")
	}
	return m.startWithLaunchLock(srv)
}

// saveRestartState is called with recoveryMu held after new launches have been
// blocked, but before Stop changes statuses. Operation contention is skipped
// conservatively: never resurrect a server being restored/deleted/reinstalled.
func (m *ServerManager) saveRestartState() error {
	if !m.recoveryEnabled || m.db == nil {
		return nil
	}
	selected := make(map[string]bool)
	m.mu.RLock()
	instances := make(map[string]*ProcessInstance, len(m.instances))
	halted := make(map[string]bool, len(m.startupHalted))
	for id, inst := range m.instances {
		instances[id] = inst
	}
	for id, value := range m.startupHalted {
		halted[id] = value
	}
	m.mu.RUnlock()
	for id := range m.recoveryPending {
		_, paused := m.restorePaused.Load(id)
		_, blocked := m.restoreBlocked.Load(id)
		if !paused && !blocked && !halted[id] {
			selected[id] = true
		}
	}
	for id, inst := range instances {
		if inst == nil || halted[id] {
			continue
		}
		if _, paused := m.restorePaused.Load(id); paused {
			continue
		}
		if _, blocked := m.restoreBlocked.Load(id); blocked {
			continue
		}
		gate, _ := m.operations.LoadOrStore(id, &sync.Mutex{})
		lock := gate.(*sync.Mutex)
		if !lock.TryLock() {
			log.Printf("[WARN] Skipping restart recovery for busy server %s; check its operation and start manually if needed", id)
			continue
		}
		status := inst.GetStatus()
		if status == "online" || (status == "starting" && m.recoveryStarting[id]) {
			selected[id] = true
		}
		lock.Unlock()
	}
	ids := make([]string, 0, len(selected))
	for id := range selected {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	if err := m.db.SaveServerResumeState(m.baseDir, ids); err != nil {
		return err
	}
	log.Printf("[INFO] Saved %d running server(s) for recovery after panel restart", len(ids))
	return nil
}

func (m *ServerManager) shutdownAll() {
	// Wait for any launch already in progress, then forbid all subsequent starts.
	m.launchMu.Lock()
	close(m.shutdownCh)
	m.launchMu.Unlock()
	m.stopWakeListeners()
	if m.trash != nil {
		m.trash.Close()
	}
	m.recoveryMu.Lock()
	m.launchMu.Lock()
	if err := m.saveRestartState(); err != nil {
		log.Printf("[ERROR] Could not save running-server state before panel shutdown: %v; servers may require manual startup", err)
	}
	m.launchMu.Unlock()
	m.recoveryMu.Unlock()
	m.mu.Lock()
	for id := range m.retryCancels {
		m.cancelRetryLocked(id)
	}
	instances := make([]*ProcessInstance, 0, len(m.instances))
	for _, inst := range m.instances {
		instances = append(instances, inst)
	}
	m.mu.Unlock()
	var wg sync.WaitGroup
	for _, inst := range instances {
		wg.Add(1)
		go func(i *ProcessInstance) {
			defer wg.Done()
			if err := i.Stop(true); err != nil {
				log.Printf("Server process cleanup failed during daemon shutdown: %v", err)
			}
		}(inst)
	}
	wg.Wait()
}
