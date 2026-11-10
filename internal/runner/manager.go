package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

type ServerManager struct {
	mu              sync.RWMutex
	baseDir         string
	db              *database.DB
	instances       map[string]*ProcessInstance
	startupFailures map[string]int
	startupHalted   map[string]bool
	retryCancels    map[string]chan struct{}
	startupTimeout  time.Duration
	shutdownCh      chan struct{}
	shutdownOnce    sync.Once
}

func NewServerManager(baseDir string, db *database.DB) *ServerManager {
	_ = os.MkdirAll(baseDir, 0755)
	return &ServerManager{
		baseDir:         baseDir,
		db:              db,
		instances:       make(map[string]*ProcessInstance),
		startupFailures: make(map[string]int),
		startupHalted:   make(map[string]bool),
		retryCancels:    make(map[string]chan struct{}),
		startupTimeout:  DefaultStartupTimeout,
		shutdownCh:      make(chan struct{}),
	}
}

func (m *ServerManager) SetStartupTimeout(d time.Duration) {
	if d <= 0 {
		return
	}
	m.mu.Lock()
	m.startupTimeout = d
	for _, inst := range m.instances {
		inst.SetStartupTimeout(d)
	}
	m.mu.Unlock()
}

func (m *ServerManager) isShuttingDown() bool {
	select {
	case <-m.shutdownCh:
		return true
	default:
		return false
	}
}

func (m *ServerManager) cancelRetryLocked(srvID string) {
	if ch, ok := m.retryCancels[srvID]; ok {
		close(ch)
		delete(m.retryCancels, srvID)
	}
}

func (m *ServerManager) GetWorkingDir(serverID string) string {
	return filepath.Join(m.baseDir, serverID)
}

func (m *ServerManager) GetInstance(serverID string) (*ProcessInstance, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	inst, ok := m.instances[serverID]
	return inst, ok
}

func (m *ServerManager) UpdateInstanceServer(srv *database.Server) {
	m.mu.RLock()
	inst, ok := m.instances[srv.ID]
	m.mu.RUnlock()
	if ok {
		inst.mu.Lock()
		inst.server = srv
		inst.mu.Unlock()
	}
}

func (m *ServerManager) getOrCreateInstance(srv *database.Server) *ProcessInstance {
	m.mu.Lock()
	defer m.mu.Unlock()

	if inst, ok := m.instances[srv.ID]; ok {
		// Update server config reference in case it changed
		inst.mu.Lock()
		inst.server = srv
		inst.mu.Unlock()
		return inst
	}

	workingDir := filepath.Join(m.baseDir, srv.ID)
	_ = os.MkdirAll(workingDir, 0755)

	timeout := m.startupTimeout
	if timeout <= 0 {
		timeout = DefaultStartupTimeout
	}

	inst := NewProcessInstance(
		srv,
		workingDir,
		timeout,
		func(srvID string) {
			m.mu.Lock()
			m.startupHalted[srvID] = false
			m.mu.Unlock()
		},
		func(srvID string, unexpected bool, reachedOnline bool, timedOut bool, uptime time.Duration) {
			m.handleProcessExit(srvID, unexpected, reachedOnline, timedOut, uptime)
		},
	)

	m.instances[srv.ID] = inst
	return inst
}

func (m *ServerManager) handleProcessExit(srvID string, unexpected bool, reachedOnline bool, timedOut bool, uptime time.Duration) {
	if !unexpected || m.isShuttingDown() {
		return
	}

	var latestSrv *database.Server
	if m.db != nil {
		s, err := m.db.GetServer(srvID)
		if err != nil {
			return
		}
		latestSrv = s
	}
	if latestSrv == nil || !latestSrv.AutoRestart {
		return
	}

	inst, ok := m.GetInstance(srvID)
	if !ok {
		return
	}

	stableRun := reachedOnline && !timedOut && uptime >= minStableUptime

	m.mu.Lock()
	if stableRun {
		m.startupFailures[srvID] = 0
		m.startupHalted[srvID] = false
	}

	isStartupFailure := !reachedOnline || timedOut || uptime < minStableUptime
	if isStartupFailure {
		m.startupFailures[srvID]++
		fails := m.startupFailures[srvID]
		if fails >= MaxStartupAttempts {
			m.startupHalted[srvID] = true
			m.cancelRetryLocked(srvID)
			m.mu.Unlock()

			inst.broadcast("\x1b[31;1m========================================================================\x1b[0m")
			inst.broadcast(fmt.Sprintf("\x1b[31;1m[NeoPanel ERROR] STARTUP TIMED OUT — %d/%d ATTEMPTS FAILED!\x1b[0m", fails, MaxStartupAttempts))
			inst.broadcast("\x1b[31mThe server failed to finish starting across 2 consecutive attempts.\x1b[0m")
			inst.broadcast("\x1b[31mAuto-restart has been halted to prevent an infinite crash loop.\x1b[0m")
			inst.broadcast("\x1b[33;1m👉 Fix the error/Java settings above, then click Start manually to try again.\x1b[0m")
			inst.broadcast("\x1b[31;1m========================================================================\x1b[0m")

			if m.db != nil {
				_ = m.db.LogActivity(0, "startup_failed", fmt.Sprintf("Server '%s' (%s) startup timed out after %d failed attempts", latestSrv.Name, srvID, fails), "system")
			}
			return
		}

		m.cancelRetryLocked(srvID)
		cancelCh := make(chan struct{})
		m.retryCancels[srvID] = cancelCh
		attemptNum := fails + 1
		m.mu.Unlock()

		inst.broadcast(fmt.Sprintf("\x1b[33;1m[NeoPanel WARN] Startup attempt %d/%d failed before coming online. Retrying in 3s (attempt %d/%d)...\x1b[0m", fails, MaxStartupAttempts, attemptNum, MaxStartupAttempts))
		go m.waitAndRetryStart(srvID, attemptNum, cancelCh)
		return
	}

	m.cancelRetryLocked(srvID)
	cancelCh := make(chan struct{})
	m.retryCancels[srvID] = cancelCh
	m.mu.Unlock()

	inst.broadcast("\x1b[33;1m[NeoPanel WARN] Server exited unexpectedly. Auto-restarting in 3s...\x1b[0m")
	go m.waitAndRetryStart(srvID, 1, cancelCh)
}

func (m *ServerManager) waitAndRetryStart(srvID string, attemptNum int, cancelCh <-chan struct{}) {
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()

	select {
	case <-cancelCh:
		return
	case <-m.shutdownCh:
		return
	case <-timer.C:
	}

	m.mu.Lock()
	curCh, exists := m.retryCancels[srvID]
	if !exists || curCh != cancelCh {
		m.mu.Unlock()
		return
	}
	delete(m.retryCancels, srvID)
	m.mu.Unlock()

	if m.isShuttingDown() || m.db == nil {
		return
	}

	latestSrv, err := m.db.GetServer(srvID)
	if err != nil || !latestSrv.AutoRestart {
		return
	}

	inst := m.getOrCreateInstance(latestSrv)
	inst.broadcast(fmt.Sprintf("\x1b[36m[NeoPanel] Executing auto-restart startup attempt %d/%d...\x1b[0m", attemptNum, MaxStartupAttempts))

	if err := inst.Start(); err != nil {
		m.mu.Lock()
		m.startupFailures[srvID]++
		fails := m.startupFailures[srvID]
		if fails >= MaxStartupAttempts {
			m.startupHalted[srvID] = true
		}
		m.mu.Unlock()

		if fails >= MaxStartupAttempts {
			inst.broadcast("\x1b[31;1m========================================================================\x1b[0m")
			inst.broadcast(fmt.Sprintf("\x1b[31;1m[NeoPanel ERROR] STARTUP TIMED OUT — %d/%d ATTEMPTS FAILED!\x1b[0m", fails, MaxStartupAttempts))
			inst.broadcast(fmt.Sprintf("\x1b[31mRetry attempt failed to launch (%v). Auto-restart is halted.\x1b[0m", err))
			inst.broadcast("\x1b[33;1m👉 Fix the error/Java settings above, then click Start manually.\x1b[0m")
			inst.broadcast("\x1b[31;1m========================================================================\x1b[0m")
			if m.db != nil {
				_ = m.db.LogActivity(0, "startup_failed", fmt.Sprintf("Server '%s' (%s) startup timed out after %d failed attempts (%v)", latestSrv.Name, srvID, fails, err), "system")
			}
		}
	}
}

func (m *ServerManager) Start(srv *database.Server) error {
	m.mu.Lock()
	m.cancelRetryLocked(srv.ID)
	m.startupFailures[srv.ID] = 0
	m.startupHalted[srv.ID] = false
	m.mu.Unlock()

	inst := m.getOrCreateInstance(srv)
	return inst.Start()
}

func (m *ServerManager) Stop(serverID string, graceful bool) error {
	m.mu.Lock()
	m.cancelRetryLocked(serverID)
	inst, ok := m.instances[serverID]
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("server instance not found or already offline")
	}

	return inst.Stop(graceful)
}

func (m *ServerManager) RemoveInstance(serverID string) {
	m.mu.Lock()
	m.cancelRetryLocked(serverID)
	delete(m.startupFailures, serverID)
	delete(m.startupHalted, serverID)
	inst, ok := m.instances[serverID]
	delete(m.instances, serverID)
	m.mu.Unlock()

	if ok && inst != nil {
		_ = inst.Stop(false)
	}
}

func (m *ServerManager) Restart(srv *database.Server) error {
	m.mu.Lock()
	m.cancelRetryLocked(srv.ID)
	inst, ok := m.instances[srv.ID]
	m.mu.Unlock()

	if ok && inst != nil && inst.GetStatus() != "offline" {
		_ = inst.Stop(true)
		deadline := time.Now().Add(2 * time.Second)
		for inst.GetStatus() != "offline" && time.Now().Before(deadline) {
			time.Sleep(50 * time.Millisecond)
		}
	}
	return m.Start(srv)
}

func (m *ServerManager) GetStatus(serverID string) string {
	m.mu.RLock()
	inst, ok := m.instances[serverID]
	m.mu.RUnlock()
	if !ok || inst == nil {
		return "offline"
	}
	return inst.GetStatus()
}

func (m *ServerManager) GetOnlinePlayers(serverID string) []OnlinePlayer {
	m.mu.RLock()
	inst, ok := m.instances[serverID]
	m.mu.RUnlock()
	if !ok || inst == nil {
		return []OnlinePlayer{}
	}
	return inst.GetOnlinePlayers()
}

func (m *ServerManager) SendCommand(serverID, cmd string) error {
	m.mu.RLock()
	inst, ok := m.instances[serverID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("server is offline")
	}

	return inst.WriteCommand(cmd)
}

func (m *ServerManager) GetLogs(serverID string) []string {
	m.mu.RLock()
	inst, ok := m.instances[serverID]
	m.mu.RUnlock()

	if !ok {
		return []string{}
	}

	return inst.GetLogs()
}

func (m *ServerManager) Subscribe(serverID string) (chan string, func(), error) {
	m.mu.RLock()
	inst, ok := m.instances[serverID]
	m.mu.RUnlock()

	if !ok {
		// Server might not be running yet, but we still allow listening
		srv, err := m.db.GetServer(serverID)
		if err != nil {
			return nil, nil, fmt.Errorf("server not found: %w", err)
		}
		inst = m.getOrCreateInstance(srv)
	}

	ch, unsub := inst.Subscribe()
	return ch, unsub, nil
}

func (m *ServerManager) GetStats(srv *database.Server) database.ServerStats {
	m.mu.RLock()
	inst, ok := m.instances[srv.ID]
	fails := m.startupFailures[srv.ID]
	halted := m.startupHalted[srv.ID]
	m.mu.RUnlock()

	if !ok {
		workingDir := filepath.Join(m.baseDir, srv.ID)
		return database.ServerStats{
			ServerID:        srv.ID,
			Status:          "offline",
			CPUUsage:        0,
			MemoryMB:        0,
			MemoryMaxMB:     srv.MemoryMaxMB,
			DiskMB:          calculateDirSizeMB(workingDir),
			UptimeSec:       0,
			StartupFailures: fails,
			StartupHalted:   halted,
		}
	}

	stats := inst.GetStats()
	if stats.Status == "online" && stats.UptimeSec >= int64(minStableUptime.Seconds()) {
		fails = 0
		halted = false
	}
	stats.StartupFailures = fails
	stats.StartupHalted = halted
	return stats
}

func (m *ServerManager) GetStatsByID(serverID string) (database.ServerStats, bool) {
	m.mu.RLock()
	inst, ok := m.instances[serverID]
	fails := m.startupFailures[serverID]
	halted := m.startupHalted[serverID]
	m.mu.RUnlock()

	if ok {
		stats := inst.GetStats()
		if stats.Status == "online" && stats.UptimeSec >= int64(minStableUptime.Seconds()) {
			fails = 0
			halted = false
		}
		stats.StartupFailures = fails
		stats.StartupHalted = halted
		return stats, true
	}

	srv, err := m.db.GetServer(serverID)
	if err != nil {
		return database.ServerStats{}, false
	}
	return m.GetStats(srv), true
}

func (m *ServerManager) ShutdownAll() {
	m.shutdownOnce.Do(func() {
		close(m.shutdownCh)
	})

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
		if inst.GetStatus() != "offline" {
			wg.Add(1)
			go func(i *ProcessInstance) {
				defer wg.Done()
				_ = i.Stop(true)
			}(inst)
		}
	}
	wg.Wait()
}
