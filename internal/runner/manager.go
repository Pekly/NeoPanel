package runner

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

type ServerManager struct {
	mu              sync.RWMutex
	baseDir         string
	backupDir       string
	db              *database.DB
	instances       map[string]*ProcessInstance
	startupFailures map[string]int
	startupHalted   map[string]bool
	retryCancels    map[string]chan struct{}
	startupTimeout  time.Duration
	shutdownCh      chan struct{}
	shutdownOnce    sync.Once
	operations      sync.Map // server ID -> *sync.Mutex; shared by power, backups and restore
	restoreBlocked  sync.Map // unrecoverable publication failures block subsequent power operations
	restorePaused   sync.Map // suppress late crash callbacks until an explicit start
}

func NewServerManager(baseDir string, db *database.DB) *ServerManager {
	_ = os.MkdirAll(baseDir, 0755)
	mgr := &ServerManager{
		baseDir:         baseDir,
		backupDir:       filepath.Join(filepath.Dir(baseDir), "backups"),
		db:              db,
		instances:       make(map[string]*ProcessInstance),
		startupFailures: make(map[string]int),
		startupHalted:   make(map[string]bool),
		retryCancels:    make(map[string]chan struct{}),
		startupTimeout:  DefaultStartupTimeout,
		shutdownCh:      make(chan struct{}),
	}
	mgr.startIdleWatchdog()
	mgr.startScheduleRunner()
	return mgr
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

	if m.db != nil {
		if srv.NodeID != "" {
			if n, err := m.db.GetNode(srv.NodeID); err == nil && n != nil {
				inst.SetIsolationConfig(n.IsolationEngine, n.TasksMax)
			}
		} else {
			if nodes, err := m.db.ListNodes(); err == nil && len(nodes) > 0 {
				inst.SetIsolationConfig(nodes[0].IsolationEngine, nodes[0].TasksMax)
			}
		}
	}

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
	if _, paused := m.restorePaused.Load(srvID); paused {
		m.mu.Unlock()
		return
	}
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

	var unlock func()
	for {
		var err error
		unlock, err = m.beginOperation(srvID)
		if err == nil {
			break
		}
		if !errors.Is(err, ErrServerOperation) {
			return
		}
		// A live backup is temporary contention, not a cancelled restart.
		timer.Reset(100 * time.Millisecond)
		select {
		case <-cancelCh:
			return
		case <-m.shutdownCh:
			return
		case <-timer.C:
		}
	}
	defer unlock()
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
	unlock, err := m.beginOperation(srv.ID)
	if err != nil {
		return err
	}
	defer unlock()
	return m.startLocked(srv)
}

func (m *ServerManager) startLocked(srv *database.Server) error {
	if m.isShuttingDown() {
		return fmt.Errorf("daemon shutting down")
	}
	m.mu.Lock()
	m.cancelRetryLocked(srv.ID)
	m.restorePaused.Delete(srv.ID)
	m.startupFailures[srv.ID] = 0
	m.startupHalted[srv.ID] = false
	m.mu.Unlock()

	inst := m.getOrCreateInstance(srv)
	// Settings changes take effect on the next launch, including reused instances.
	if m.db != nil {
		nodeID := srv.NodeID
		if nodeID == "" {
			nodeID = "local"
		}
		if node, err := m.db.GetNode(nodeID); err == nil && node != nil {
			inst.SetIsolationConfig(node.IsolationEngine, node.TasksMax)
		}
	}
	return inst.Start()
}

func (m *ServerManager) Stop(serverID string, graceful bool) error {
	// Explicit stop cancels pending automatic retries even if another operation
	// currently owns the filesystem gate. No process/files are changed on conflict.
	m.mu.Lock()
	m.cancelRetryLocked(serverID)
	m.restorePaused.Store(serverID, true)
	m.mu.Unlock()
	unlock, err := m.beginOperation(serverID)
	if err != nil {
		return err
	}
	defer unlock()
	m.mu.Lock()
	m.cancelRetryLocked(serverID)
	inst, ok := m.instances[serverID]
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("server instance not found or already offline")
	}

	return inst.Stop(graceful)
}

func (m *ServerManager) RemoveInstance(serverID string) error {
	return m.DeleteServer(context.Background(), serverID, func() error { return nil })
}

func (m *ServerManager) removeInstanceLocked(serverID string) {
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
	unlock, err := m.beginOperation(srv.ID)
	if err != nil {
		return err
	}
	defer unlock()
	m.mu.Lock()
	m.cancelRetryLocked(srv.ID)
	inst, ok := m.instances[srv.ID]
	m.mu.Unlock()

	if ok && inst != nil && inst.GetStatus() != "offline" {
		if err := inst.Stop(true); err != nil {
			return fmt.Errorf("restart refused because shutdown did not complete: %w", err)
		}
	}
	return m.startLocked(srv)
}

func (m *ServerManager) StopServer(serverID string, graceful bool) error {
	return m.Stop(serverID, graceful)
}

func (m *ServerManager) StartServer(serverID string) error {
	srv, err := m.db.GetServer(serverID)
	if err != nil {
		return err
	}
	return m.Start(srv)
}

func (m *ServerManager) RestartServer(serverID string) error {
	srv, err := m.db.GetServer(serverID)
	if err != nil {
		return err
	}
	return m.Restart(srv)
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
		unlock, err := m.beginOperation(serverID)
		if err != nil {
			return nil, nil, err
		}
		defer unlock()
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

func (m *ServerManager) startIdleWatchdog() {
	ticker := time.NewTicker(20 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-m.shutdownCh:
				return
			case <-ticker.C:
				m.checkIdleTimeouts()
			}
		}
	}()
}

func (m *ServerManager) checkIdleTimeouts() {
	if m.db == nil {
		return
	}
	m.mu.RLock()
	instances := make([]*ProcessInstance, 0, len(m.instances))
	for _, inst := range m.instances {
		instances = append(instances, inst)
	}
	m.mu.RUnlock()

	for _, inst := range instances {
		if inst.GetStatus() != "online" {
			continue
		}
		inst.mu.RLock()
		srv := inst.server
		inst.mu.RUnlock()
		if srv == nil {
			continue
		}

		nodeID := srv.NodeID
		if nodeID == "" {
			nodeID = "local"
		}
		node, err := m.db.GetNode(nodeID)
		if err != nil || node == nil {
			continue
		}

		if !node.IsFree && node.IdleTimeoutMin <= 0 {
			continue
		}

		timeoutMin := node.IdleTimeoutMin
		if timeoutMin <= 0 && node.IsFree {
			timeoutMin = 15 // Default 15 minutes for free node
		}

		idleDur, isIdle := inst.GetIdleDuration()
		if !isIdle || idleDur < time.Duration(timeoutMin)*time.Minute {
			continue
		}

		// Timeout reached!
		inst.broadcast(fmt.Sprintf("\x1b[33;1m[NeoPanel] Server idle timeout reached (0 players connected for %d minutes on free node).\x1b[0m", int(idleDur.Minutes())))
		inst.broadcast("\x1b[33m[NeoPanel] Gracefully stopping server to conserve host resources...\x1b[0m")
		_ = m.StopServer(srv.ID, false)
		_ = m.db.CreateAuditLog("system", "server.idle_timeout", fmt.Sprintf("Server %s auto-stopped after %d min idle on free node %s", srv.Name, int(idleDur.Minutes()), node.Name), "")
	}
}

func (m *ServerManager) startScheduleRunner() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-m.shutdownCh:
				return
			case now := <-ticker.C:
				m.runScheduledTasks(now.UTC())
			}
		}
	}()
}

func (m *ServerManager) runScheduledTasks(now time.Time) {
	if m.db == nil {
		return
	}
	schedules, err := m.db.ListAllActiveSchedules()
	if err != nil || len(schedules) == 0 {
		return
	}

	for _, sched := range schedules {
		if MatchesCron(sched.CronExpr, now) {
			if err := m.startSchedule(sched, now, true); err != nil {
				log.Printf("schedule %s: %v", sched.ID, err)
			}
		}
	}
}

func (m *ServerManager) SetBackupDirectory(dir string) {
	m.mu.Lock()
	m.backupDir = dir
	m.mu.Unlock()
}

// StartSchedule records acceptance before returning; completion is recorded by the worker.
func (m *ServerManager) StartSchedule(sched *database.ServerSchedule) error {
	return m.startSchedule(sched, time.Now().UTC(), false)
}

var ErrScheduleRunning = fmt.Errorf("schedule is already running or was already triggered this minute")

func (m *ServerManager) startSchedule(sched *database.ServerSchedule, now time.Time, scheduled bool) error {
	if _, err := database.ParseCron(sched.CronExpr); err != nil {
		return err
	}
	snapshot, tasks, claimed, err := m.db.ClaimScheduleWithTasks(sched.ID, now, scheduled)
	if err != nil {
		return err
	}
	if !claimed {
		return ErrScheduleRunning
	}
	go m.executeClaimedSchedule(snapshot, tasks)
	return nil
}

// ExecuteSchedule is retained for internal callers and also enforces non-overlap.
func (m *ServerManager) ExecuteSchedule(sched *database.ServerSchedule) {
	if err := m.StartSchedule(sched); err != nil {
		log.Printf("schedule %s: %v", sched.ID, err)
	}
}

func (m *ServerManager) executeClaimedSchedule(sched *database.ServerSchedule, tasks []database.ScheduleTask) {
	var runErr error
	defer func() {
		if recovered := recover(); recovered != nil {
			runErr = fmt.Errorf("schedule panic: %v", recovered)
		}
		status, message := "succeeded", ""
		if runErr != nil {
			status, message = "failed", runErr.Error()
		}
		if err := m.db.SetScheduleRunStatus(sched.ID, status, message); err != nil {
			log.Printf("record schedule %s completion: %v", sched.ID, err)
		}
		_ = m.db.CreateAuditLog("system", "schedule."+status, fmt.Sprintf("Schedule %s (%s): %s", sched.Name, sched.ID, message), "")
	}()
	if runErr = m.db.SetScheduleRunStatus(sched.ID, "running", ""); runErr != nil {
		return
	}
	if _, runErr = m.db.GetServer(sched.ServerID); runErr != nil {
		return
	}
	runErr = executeScheduleTasks(tasks, m.waitScheduleDelay, func(task database.ScheduleTask) error {
		step := *sched
		step.Action, step.Payload = task.Action, task.Payload
		return m.executeScheduleAction(&step)
	})
}

func (m *ServerManager) executeScheduleAction(sched *database.ServerSchedule) (runErr error) {
	switch sched.Action {
	case "command":
		if strings.TrimSpace(sched.Payload) == "" {
			runErr = fmt.Errorf("scheduled command is empty")
			return
		}
		runErr = m.SendCommand(sched.ServerID, sched.Payload)
	case "backup":
		name := fmt.Sprintf("backup-scheduled-%s-%d.tar.gz", sched.ID, time.Now().UTC().UnixNano())
		_, runErr = m.CreateServerBackup(sched.ServerID, name)
	case "restart", "start", "stop":
		target := "online"
		if sched.Action == "restart" {
			runErr = m.RestartServer(sched.ServerID)
		}
		if sched.Action == "start" {
			runErr = m.StartServer(sched.ServerID)
		}
		if sched.Action == "stop" {
			target = "offline"
			runErr = m.StopServer(sched.ServerID, true)
		}
		if runErr != nil {
			return
		}
		m.mu.RLock()
		timeout := m.startupTimeout
		m.mu.RUnlock()
		if timeout <= 0 {
			timeout = DefaultStartupTimeout
		}
		deadline := time.NewTimer(timeout + 60*time.Second)
		defer deadline.Stop()
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for m.GetStatus(sched.ServerID) != target {
			if target == "online" && m.GetStatus(sched.ServerID) == "offline" {
				runErr = fmt.Errorf("server stopped before becoming online")
				return
			}
			select {
			case <-m.shutdownCh:
				runErr = fmt.Errorf("daemon shutting down")
				return
			case <-deadline.C:
				runErr = fmt.Errorf("timed out waiting for server to become %s", target)
				return
			case <-ticker.C:
			}
		}
	default:
		runErr = fmt.Errorf("unsupported scheduled action %q", sched.Action)
	}
	return runErr
}

func MatchesCron(expr string, t time.Time) bool {
	c, err := database.ParseCron(expr)
	return err == nil && c.Matches(t)
}
