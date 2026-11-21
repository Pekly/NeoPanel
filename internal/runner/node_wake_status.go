package runner

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

const WakeStatusMaxAge = 10 * time.Second
const WakeStatusMaxServers = maxWakeListeners

// NodeWakeStatus describes a reconciliation, not a configuration delivery ack.
// A zero checked_at and empty config_hash mean no reconciliation has run yet.
// Listeners counts bound sockets, including cooldown/busy sockets, not readiness.
type NodeWakeStatus struct {
	CheckedAt          time.Time          `json:"checked_at"`
	Enabled            bool               `json:"enabled"`
	CoordinatorRunning bool               `json:"coordinator_running"`
	ConfigMatches      bool               `json:"config_matches"`
	ConfigHash         string             `json:"config_hash"`
	Listeners          int                `json:"listeners"`
	Servers            []WakeServerStatus `json:"servers"`
	Error              string             `json:"error"`
}

type WakeServerStatus struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Port   int    `json:"port"`
	State  string `json:"state"`
	Detail string `json:"detail"`
}

func NodeWakeConfigHash(cfg database.NodeWakeConfig) string {
	data, _ := json.Marshal(cfg)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func wakeStatusText(s string, limit int) string {
	r := []rune(strings.ToValidUTF8(s, ""))
	if len(r) > limit {
		r = r[:limit]
	}
	return string(r)
}

func wakeStatusRows(servers []*database.Server) ([]WakeServerStatus, int) {
	rows := make([]WakeServerStatus, 0)
	omitted := 0
	for _, s := range servers {
		if s.NodeID != "" && s.NodeID != "local" {
			continue
		}
		if len(rows) == WakeStatusMaxServers {
			omitted++
			continue
		}
		rows = append(rows, WakeServerStatus{ID: s.ID, Name: wakeStatusText(s.Name, 128), Port: s.Port, State: "pending", Detail: "Waiting for listener reconciliation"})
	}
	return rows, omitted
}

func wakeStatusError(report *NodeWakeStatus, message string) {
	if report.Error != "" {
		report.Error += "; "
	}
	report.Error += message
}

// Called only at the end of reconciliation with launchMu and wake.mu held.
func (m *ServerManager) captureWakeStatusLocked(cfg database.NodeWakeConfig, servers []*database.Server, failed bool) {
	w := &m.wake
	rows, omitted := wakeStatusRows(servers)
	report := NodeWakeStatus{CheckedAt: time.Now().UTC(), Enabled: cfg.Enabled, CoordinatorRunning: true, ConfigHash: NodeWakeConfigHash(cfg), ConfigMatches: !failed, Listeners: len(w.listeners), Servers: rows}
	if failed {
		report.ConfigHash = ""
		report.Error = "Unable to read local wake configuration or server inventory"
	}
	if omitted > 0 {
		wakeStatusError(&report, fmt.Sprintf("Server details truncated: %d additional local servers omitted", omitted))
	}
	counts := make(map[int]int)
	inventory := make(map[string]*database.Server)
	for _, s := range servers {
		if s.NodeID == "" || s.NodeID == "local" {
			counts[s.Port]++
			inventory[s.ID] = s
		}
	}
	for i := range report.Servers {
		row := &report.Servers[i]
		s := inventory[row.ID]
		set := func(state, detail string) { row.State, row.Detail = state, detail }
		if failed {
			continue
		}
		if !cfg.Enabled {
			set("disabled", "Join-to-start is disabled")
			continue
		}
		if !localWakeServer(s) {
			set("ineligible", "Unsupported server type, server ID or game port (must be 1024–65535)")
			continue
		}
		switch m.GetStatus(s.ID) {
		case "online", "running":
			set("online", "Server is already running")
			continue
		case "starting":
			set("starting", "Server startup is in progress")
			continue
		case "offline":
		default:
			set("busy", "Server is transitioning or stopping")
			continue
		}
		_, blocked := m.restoreBlocked.Load(s.ID)
		m.mu.RLock()
		halted, retrying := m.startupHalted[s.ID], m.retryCancels[s.ID] != nil
		m.mu.RUnlock()
		if blocked || halted {
			set("blocked", "Automatic startup is blocked; administrator action required")
			continue
		}
		if retrying {
			set("starting", "A startup retry is scheduled")
			continue
		}
		if counts[s.Port] != 1 {
			set("conflict", "Multiple local servers use this game port")
			continue
		}
		e := w.listeners[s.ID]
		if e != nil && e.queued {
			set("starting", "A player join has queued startup")
			continue
		}
		if gate, ok := m.operations.Load(s.ID); ok {
			lock := gate.(*sync.Mutex)
			if !lock.TryLock() {
				set("busy", "Another server operation is in progress")
				continue
			}
			lock.Unlock()
		}
		if e != nil {
			if time.Now().Before(w.cooldown[s.ID]) {
				set("cooldown", "Listener bound; startup cooldown is active")
			} else {
				set("listening", "Listener bound and waiting for a player join")
			}
		} else if w.errors[s.ID] != "" {
			set("bind_failed", "Port unavailable or bind address not available; check node logs")
		} else if len(w.listeners) >= maxWakeListeners {
			set("pending", "Node listener limit reached")
		}
	}
	w.status = report
}

// GetNodeWakeStatus never reconciles, probes sockets, binds ports or launches.
// It copies a cached bounded reconciliation and checks live socket ownership and
// the latest desired configuration without waiting for the launch lock.
func (m *ServerManager) GetNodeWakeStatus() NodeWakeStatus {
	var desired database.NodeWakeConfig
	var err error
	if m.db != nil {
		desired, err = m.db.GetNodeWakeConfig("local")
	} else {
		err = fmt.Errorf("database unavailable")
	}
	w := &m.wake
	w.mu.Lock()
	report := w.status
	report.Servers = append([]WakeServerStatus{}, report.Servers...)
	report.CoordinatorRunning = w.started && !m.isShuttingDown()
	report.Listeners = len(w.listeners)
	report.ConfigMatches = err == nil && report.ConfigHash == NodeWakeConfigHash(desired)
	if report.CheckedAt.IsZero() {
		report.Enabled = desired.Enabled
	}
	if !report.CoordinatorRunning {
		wakeStatusError(&report, "Wake coordinator is not running")
	}
	if report.CheckedAt.IsZero() || time.Since(report.CheckedAt) > WakeStatusMaxAge {
		report.CoordinatorRunning = false
		wakeStatusError(&report, "No recent listener reconciliation")
	}
	if err != nil {
		wakeStatusError(&report, "Unable to read desired wake configuration")
	} else if !report.ConfigMatches {
		wakeStatusError(&report, "Saved configuration is pending reconciliation")
	}
	for i := range report.Servers {
		row := &report.Servers[i]
		e := w.listeners[row.ID]
		if !report.CoordinatorRunning || !report.ConfigMatches {
			row.State, row.Detail = "pending", "Waiting for a current reconciliation of the saved configuration"
		} else if row.State == "listening" || row.State == "cooldown" {
			if e == nil {
				row.State, row.Detail = "pending", "Listener released; waiting for reconciliation"
			} else if e.queued {
				row.State, row.Detail = "starting", "A player join has queued startup"
			} else if time.Now().Before(w.cooldown[row.ID]) {
				row.State, row.Detail = "cooldown", "Listener bound; startup cooldown is active"
			}
		}
	}
	w.mu.Unlock()
	if report.CheckedAt.IsZero() && m.db != nil {
		servers, listErr := m.db.ListServers()
		if listErr == nil {
			var omitted int
			report.Servers, omitted = wakeStatusRows(servers)
			if omitted > 0 {
				wakeStatusError(&report, fmt.Sprintf("Server details truncated: %d additional local servers omitted", omitted))
			}
			if !desired.Enabled {
				for i := range report.Servers {
					report.Servers[i].State, report.Servers[i].Detail = "disabled", "Join-to-start is disabled"
				}
			}
		} else {
			wakeStatusError(&report, "Unable to read local server inventory")
		}
	}
	return report
}
