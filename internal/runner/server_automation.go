package runner

import (
	"fmt"
	"sync"

	"github.com/Pekly/NeoPanel/internal/database"
)

// ServerAutomation is effective policy on the host, not master-side node metadata.
type ServerAutomation struct {
	JoinToStartEnabled     bool   `json:"join_to_start_enabled"`
	NodeJoinToStartEnabled bool   `json:"node_join_to_start_enabled"`
	IdleShutdownEnabled    bool   `json:"idle_shutdown_enabled"`
	IdleTimeoutMinutes     int    `json:"idle_timeout_minutes"`
	IdleShutdownLocked     bool   `json:"idle_shutdown_locked"`
	StoppedByIdlePolicy    bool   `json:"stopped_by_idle_policy"`
	WakeEligible           bool   `json:"wake_eligible"`
	CanEditJoinToStart     bool   `json:"can_edit_join_to_start"`
	Reason                 string `json:"reason"`
}

// Caller holds the operation gate. Suppression survives a failed durable clear
// for this daemon; operations must not proceed until the durable clear succeeds.
func (m *ServerManager) invalidateIdleWake(id string) error {
	m.wakeSuppressed.Store(id, true)
	m.releaseWakeListener(id)
	if m.db == nil {
		return nil
	}
	if err := m.db.SetServerIdleStopped(id, false); err != nil {
		return fmt.Errorf("clear idle wake authority: %w", err)
	}
	m.wakeSuppressed.Delete(id)
	return nil
}

func (m *ServerManager) wakePolicyView(s *database.Server) (ServerAutomation, error) {
	v := ServerAutomation{IdleShutdownLocked: true}
	if m.db == nil {
		v.Reason = "blocked"
		return v, fmt.Errorf("database unavailable")
	}
	cfg, err := m.db.GetNodeWakeConfig("local")
	if err != nil {
		v.Reason = "blocked"
		return v, err
	}
	idle, err := m.db.GetNodeIdlePolicy("local")
	if err != nil {
		v.Reason = "blocked"
		return v, err
	}
	state, err := m.db.GetServerWakeState(s.ID)
	if err != nil {
		v.Reason = "blocked"
		return v, err
	}
	v.JoinToStartEnabled = state.JoinToStartEnabled
	v.NodeJoinToStartEnabled = cfg.Enabled
	v.IdleShutdownEnabled = idle.Enabled
	v.IdleTimeoutMinutes = idle.TimeoutMinutes
	v.StoppedByIdlePolicy = state.StoppedByIdlePolicy
	if _, suppressed := m.wakeSuppressed.Load(s.ID); suppressed {
		v.StoppedByIdlePolicy = false
	}
	switch {
	case !localWakeServer(s):
		v.Reason = "ineligible"
	case !cfg.Enabled:
		v.Reason = "disabled"
	case !idle.Enabled:
		v.Reason = "node_idle_disabled"
	case !state.JoinToStartEnabled:
		v.Reason = "user_disabled"
	case !v.StoppedByIdlePolicy:
		v.Reason = "not_idle_stopped"
	default:
		v.Reason = "eligible"
	}
	return v, nil
}

// GetServerAutomation uses the same eligibility guards as queued automatic starts.
func (m *ServerManager) GetServerAutomation(id string) (ServerAutomation, error) {
	if m.db == nil {
		return ServerAutomation{}, fmt.Errorf("database unavailable")
	}
	m.launchMu.Lock()
	defer m.launchMu.Unlock()
	s, err := m.db.GetServer(id)
	if err != nil {
		return ServerAutomation{}, err
	}
	v, err := m.wakePolicyView(s)
	if err != nil || v.Reason != "eligible" {
		return v, err
	}
	if !m.wakeEligible(s) {
		v.Reason = "blocked"
		return v, nil
	}
	gate, _ := m.operations.LoadOrStore(id, &sync.Mutex{})
	lock := gate.(*sync.Mutex)
	if !lock.TryLock() {
		v.Reason = "busy"
		return v, nil
	}
	lock.Unlock()
	servers, err := m.db.ListServers()
	if err != nil {
		return v, err
	}
	for _, other := range servers {
		if other.ID != id && (other.NodeID == "" || other.NodeID == "local") && other.Port == s.Port {
			v.Reason = "conflict"
			return v, nil
		}
	}
	v.WakeEligible = true
	return v, nil
}

// SetServerWakePreference changes only the preference, never the idle marker or
// administrator policy. Closing the old generation also cancels accepted joins.
func (m *ServerManager) SetServerWakePreference(id string, enabled bool) error {
	release, err := m.beginOperation(id)
	if err != nil {
		return err
	}
	defer release()
	m.launchMu.Lock()
	defer m.launchMu.Unlock()
	m.releaseWakeListener(id)
	if m.db == nil {
		return fmt.Errorf("database unavailable")
	}
	if err := m.db.SetServerWakePreference(id, enabled); err != nil {
		m.wakeSuppressed.Store(id, true)
		return err
	}
	return nil
}
