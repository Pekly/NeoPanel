package runner

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/mcwake"
	"github.com/Pekly/NeoPanel/internal/rooted"
)

const maxWakeListeners = 1024

type wakeListener struct {
	listener    net.Listener
	server      database.Server
	config      database.NodeWakeConfig
	icon        string
	queued      bool
	connections map[net.Conn]bool
}
type wakeIPBudget struct {
	since time.Time
	count int
}
type wakeCoordinator struct {
	mu        sync.Mutex
	started   bool
	listeners map[string]*wakeListener
	cooldown  map[string]time.Time
	ips       map[string]wakeIPBudget
	errors    map[string]string
	slots     chan struct{}
	wg        sync.WaitGroup
}

// StartWakeListeners is called only after daemon binding and restart recovery.
// Construction alone never binds game ports. Configuration is local to this host.
func (m *ServerManager) StartWakeListeners() {
	m.launchMu.Lock()
	defer m.launchMu.Unlock()
	w := &m.wake
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.started || m.isShuttingDown() || m.db == nil {
		return
	}
	w.started = true
	w.listeners = make(map[string]*wakeListener)
	w.cooldown = make(map[string]time.Time)
	w.ips = make(map[string]wakeIPBudget)
	w.errors = make(map[string]string)
	w.slots = make(chan struct{}, 64)
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		tick := time.NewTicker(time.Second)
		defer tick.Stop()
		for {
			m.reconcileWakeListeners()
			select {
			case <-m.shutdownCh:
				return
			case <-tick.C:
			}
		}
	}()
}

func localWakeServer(s *database.Server) bool {
	if s == nil || (s.NodeID != "" && s.NodeID != "local") || !rooted.Basename(s.ID) || s.Port < 1024 || s.Port > 65535 {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(s.Type)) {
	case "paper", "purpur", "leaf", "leafmc", "folia", "fabric", "forge", "neoforge", "vanilla", "spigot", "custom", "velocity", "waterfall", "bungeecord", "steel":
		return true
	}
	return false
}

// Caller owns launchMu. A wake is an external request, never a manual reset of
// the startup failure circuit breaker, and never a recovery snapshot entry.
func (m *ServerManager) wakeEligible(s *database.Server) bool {
	if !localWakeServer(s) || m.isShuttingDown() || m.GetStatus(s.ID) != "offline" {
		return false
	}
	if _, blocked := m.restoreBlocked.Load(s.ID); blocked {
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return !m.startupHalted[s.ID] && m.retryCancels[s.ID] == nil
}

// Caller owns wake.mu. Closing a listener synchronously relinquishes its port;
// connection goroutines are joined only at shutdown, never under launchMu.
func (m *ServerManager) closeWakeLocked(id string) {
	w := &m.wake
	e := w.listeners[id]
	if e == nil {
		return
	}
	delete(w.listeners, id)
	_ = e.listener.Close()
	for c := range e.connections {
		_ = c.Close()
	}
}
func (m *ServerManager) releaseWakeListener(id string) {
	m.wake.mu.Lock()
	defer m.wake.mu.Unlock()
	m.closeWakeLocked(id)
}
func (m *ServerManager) stopWakeListeners() {
	w := &m.wake
	w.mu.Lock()
	for id := range w.listeners {
		m.closeWakeLocked(id)
	}
	w.mu.Unlock()
	w.wg.Wait()
}

func (m *ServerManager) reconcileWakeListeners() {
	m.launchMu.Lock()
	defer m.launchMu.Unlock()
	if m.isShuttingDown() || m.db == nil {
		return
	}
	cfg, err := m.db.GetNodeWakeConfig("local")
	servers, listErr := m.db.ListServers()
	w := &m.wake
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.started {
		return
	}
	if err != nil || listErr != nil || !cfg.Enabled {
		for id := range w.listeners {
			m.closeWakeLocked(id)
		}
		return
	}
	counts := make(map[int]int)
	for _, s := range servers {
		if s.NodeID == "" || s.NodeID == "local" {
			counts[s.Port]++
		}
	}
	keep := make(map[string]bool)
	inventory := make(map[string]bool)
	for _, s := range servers {
		inventory[s.ID] = true
		if !m.wakeEligible(s) || counts[s.Port] != 1 {
			continue
		}
		old := w.listeners[s.ID]
		if old != nil && (old.config != cfg || old.server.Port != s.Port || old.server.Name != s.Name || old.server.Version != s.Version || old.server.Type != s.Type) {
			m.closeWakeLocked(s.ID)
			old = nil
		}
		gate, _ := m.operations.LoadOrStore(s.ID, &sync.Mutex{})
		lock := gate.(*sync.Mutex)
		if !lock.TryLock() {
			// The gate may belong to an accepted wake waiting for launchMu.
			// Keep its unchanged listener/generation so reconciliation cannot
			// cancel that request. Destructive operations explicitly close it.
			keep[s.ID] = old != nil
			continue
		}
		keep[s.ID] = true
		if old == nil && len(w.listeners) < maxWakeListeners {
			address := net.JoinHostPort(cfg.BindAddress, strconv.Itoa(s.Port))
			listener, bindErr := net.Listen("tcp", address)
			if bindErr != nil {
				message := bindErr.Error()
				if w.errors[s.ID] != message {
					log.Printf("[WARN] Join-to-start cannot listen for %s on %s: %v", s.ID, address, bindErr)
					w.errors[s.ID] = message
				}
			} else {
				delete(w.errors, s.ID)
				e := &wakeListener{listener: listener, server: *s, config: cfg, icon: m.wakeIcon(s.ID, cfg), connections: make(map[net.Conn]bool)}
				w.listeners[s.ID] = e
				w.wg.Add(1)
				go m.acceptWake(e)
			}
		}
		lock.Unlock()
	}
	for id := range w.listeners {
		if !keep[id] {
			m.closeWakeLocked(id)
		}
	}
	for id := range w.cooldown {
		if !inventory[id] {
			delete(w.cooldown, id)
		}
	}
	for id := range w.errors {
		if !inventory[id] {
			delete(w.errors, id)
		}
	}
	for ip, budget := range w.ips {
		if time.Since(budget.since) > time.Minute {
			delete(w.ips, ip)
		}
	}
}

func (m *ServerManager) wakeIcon(id string, cfg database.NodeWakeConfig) string {
	if !cfg.UseServerIcon {
		return cfg.IconDataURL
	}
	root, err := rooted.Open(m.baseDir)
	if err != nil {
		return cfg.IconDataURL
	}
	defer root.Close()
	dir, err := root.Sub(id)
	if err != nil {
		return cfg.IconDataURL
	}
	defer dir.Close()
	file, err := dir.Open("server-icon.png")
	if err != nil {
		return cfg.IconDataURL
	}
	defer file.Close()
	data, err := rooted.ReadBounded(file, 48*1024)
	if err != nil {
		return cfg.IconDataURL
	}
	icon := "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
	if database.ValidateWakeIcon(icon) != nil {
		return cfg.IconDataURL
	}
	return icon
}

func wakeText(text string, server database.Server) string {
	text = strings.NewReplacer("{server}", server.Name, "{version}", server.Version).Replace(text)
	var out strings.Builder
	for i := 0; i < len(text); i++ {
		if text[i] == '&' && i+1 < len(text) && strings.ContainsRune("0123456789abcdefklmnorABCDEFKLMNOR", rune(text[i+1])) {
			out.WriteString("§")
			continue
		}
		out.WriteByte(text[i])
	}
	return out.String()
}

func (m *ServerManager) acceptWake(e *wakeListener) {
	w := &m.wake
	defer w.wg.Done()
	for {
		conn, err := e.listener.Accept()
		if err != nil {
			return
		}
		ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
		w.mu.Lock()
		budget := w.ips[ip]
		if time.Since(budget.since) >= 10*time.Second {
			budget = wakeIPBudget{since: time.Now()}
		}
		allowed := w.listeners[e.server.ID] == e && !m.isShuttingDown() && budget.count < 12
		if _, known := w.ips[ip]; !known && len(w.ips) >= 4096 {
			allowed = false
		}
		if allowed {
			select {
			case w.slots <- struct{}{}:
			default:
				allowed = false
			}
		}
		if allowed {
			budget.count++
			w.ips[ip] = budget
			e.connections[conn] = true
			w.wg.Add(1)
		}
		w.mu.Unlock()
		if !allowed {
			_ = conn.Close()
			continue
		}
		go func() {
			defer w.wg.Done()
			defer func() { <-w.slots; w.mu.Lock(); delete(e.connections, conn); w.mu.Unlock() }()
			requested := false
			_ = mcwake.Handle(conn, func(_ int32) mcwake.Status {
				return mcwake.Status{MOTD: wakeText(e.config.OfflineMOTD, e.server), VersionName: wakeText(e.config.VersionLabel, e.server), MaxPlayers: e.config.MaxPlayers, IconDataURL: e.icon}
			}, func(_ mcwake.Join) string {
				requested = m.reserveWake(e)
				if requested {
					return wakeText(e.config.StartingMessage, e.server)
				}
				return wakeText(e.config.UnavailableMessage, e.server)
			})
			// Handle has written the disconnect and closed the connection before the
			// port is handed to the real game. Do not wait for player authentication.
			if requested {
				m.launchWake(e)
			}
		}()
	}
}

func (m *ServerManager) reserveWake(e *wakeListener) bool {
	unlock, err := m.beginOperation(e.server.ID)
	if err != nil {
		return false
	}
	defer unlock()
	m.launchMu.Lock()
	defer m.launchMu.Unlock()
	if !m.wakeEligible(&e.server) {
		return false
	}
	w := &m.wake
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.listeners[e.server.ID] != e || e.queued || time.Now().Before(w.cooldown[e.server.ID]) {
		return false
	}
	e.queued = true
	w.cooldown[e.server.ID] = time.Now().Add(time.Duration(e.config.CooldownSeconds) * time.Second)
	return true
}

func (m *ServerManager) launchWake(e *wakeListener) {
	defer func() { m.wake.mu.Lock(); e.queued = false; m.wake.mu.Unlock() }()
	unlock, err := m.beginOperation(e.server.ID)
	if err != nil {
		return
	}
	defer unlock()
	m.launchMu.Lock()
	defer m.launchMu.Unlock()
	cfg, err := m.db.GetNodeWakeConfig("local")
	if err != nil || !cfg.Enabled || cfg != e.config {
		return
	}
	srv, err := m.db.GetServer(e.server.ID)
	if err != nil || srv.Port != e.server.Port || !m.wakeEligible(srv) {
		return
	}
	m.wake.mu.Lock()
	current := m.wake.listeners[srv.ID] == e
	m.wake.mu.Unlock()
	if !current {
		return
	}
	if err := m.startWithLaunchPolicy(srv, false); err != nil {
		// A synchronous failed wake must not create a remotely-triggerable launch
		// loop. An administrator can explicitly start after fixing the error.
		m.mu.Lock()
		m.startupHalted[srv.ID] = true
		m.mu.Unlock()
		log.Printf("[WARN] Join-to-start failed for %s: %v; manual start required", srv.ID, err)
		_ = m.db.LogActivity(0, "join_start_failed", fmt.Sprintf("Join-to-start failed for %s: %v", srv.ID, err), "system")
		return
	}
	_ = m.db.LogActivity(0, "join_start", fmt.Sprintf("Minecraft join requested startup for %s", srv.ID), "system")
}
