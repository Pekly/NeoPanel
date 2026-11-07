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
	mu         sync.RWMutex
	baseDir    string
	db         *database.DB
	instances  map[string]*ProcessInstance
	shutdownCh chan struct{}
}

func NewServerManager(baseDir string, db *database.DB) *ServerManager {
	_ = os.MkdirAll(baseDir, 0755)
	return &ServerManager{
		baseDir:    baseDir,
		db:         db,
		instances:  make(map[string]*ProcessInstance),
		shutdownCh: make(chan struct{}),
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

func (m *ServerManager) getOrCreateInstance(srv *database.Server) *ProcessInstance {
	m.mu.Lock()
	defer m.mu.Unlock()

	if inst, ok := m.instances[srv.ID]; ok {
		// Update server config reference in case it changed
		inst.server = srv
		return inst
	}

	workingDir := filepath.Join(m.baseDir, srv.ID)
	_ = os.MkdirAll(workingDir, 0755)

	inst := NewProcessInstance(srv, workingDir, func(srvID string, unexpected bool) {
		if unexpected && srv.AutoRestart {
			go func() {
				time.Sleep(3 * time.Second)
				latestSrv, err := m.db.GetServer(srvID)
				if err == nil {
					_ = m.Start(latestSrv)
				}
			}()
		}
	})

	m.instances[srv.ID] = inst
	return inst
}

func (m *ServerManager) Start(srv *database.Server) error {
	inst := m.getOrCreateInstance(srv)
	return inst.Start()
}

func (m *ServerManager) Stop(serverID string, graceful bool) error {
	m.mu.RLock()
	inst, ok := m.instances[serverID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("server instance not found or already offline")
	}

	return inst.Stop(graceful)
}

func (m *ServerManager) RemoveInstance(serverID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if inst, ok := m.instances[serverID]; ok {
		_ = inst.Stop(false)
		delete(m.instances, serverID)
	}
}

func (m *ServerManager) Restart(srv *database.Server) error {
	_ = m.Stop(srv.ID, true)
	time.Sleep(1 * time.Second)
	return m.Start(srv)
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
	m.mu.RUnlock()

	if !ok {
		workingDir := filepath.Join(m.baseDir, srv.ID)
		return database.ServerStats{
			ServerID:  srv.ID,
			Status:    "offline",
			CPUUsage:  0,
			MemoryMB:  0,
			DiskMB:    calculateDirSizeMB(workingDir),
			UptimeSec: 0,
		}
	}

	return inst.GetStats()
}

func (m *ServerManager) ShutdownAll() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var wg sync.WaitGroup
	for _, inst := range m.instances {
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

