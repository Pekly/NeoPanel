package sftp

import (
	"net"
	"sync"
	"time"
)

// Limits apply to the whole built-in listener, not to individual accounts.
// The total includes authenticated clients, so authentication cannot bypass it.
const (
	defaultPendingConnections = 64
	defaultPendingPerSource   = 8
	defaultTotalConnections   = 256
	maxSessionChannels        = 8
	defaultHandshakeTimeout   = 30 * time.Second
)

type connectionAdmission struct {
	mu          sync.Mutex
	stopped     bool
	connections map[net.Conn]bool // true until the SSH handshake completes
	sources     map[string]int
	pending     int
	// Zero uses production defaults; tests may inject smaller limits before dialing.
	maxPending, maxPerSource, maxTotal int
	handshakeTimeout                   time.Duration
}

func sourceKey(c net.Conn) string {
	host, _, err := net.SplitHostPort(c.RemoteAddr().String())
	if err != nil {
		return c.RemoteAddr().String()
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.String()
	}
	return host
}

func positiveOr(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

// Called by acceptLoop BEFORE starting a goroutine. All tracking is bounded by
// maxTotal; rejected sockets do not create workers or retain source entries.
func (a *connectionAdmission) admit(c net.Conn) (time.Duration, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	key := sourceKey(c)
	if a.stopped || len(a.connections) >= positiveOr(a.maxTotal, defaultTotalConnections) ||
		a.pending >= positiveOr(a.maxPending, defaultPendingConnections) ||
		a.sources[key] >= positiveOr(a.maxPerSource, defaultPendingPerSource) {
		return 0, false
	}
	if a.connections == nil {
		a.connections = make(map[net.Conn]bool)
		a.sources = make(map[string]int)
	}
	a.connections[c] = true
	a.pending++
	a.sources[key]++
	timeout := a.handshakeTimeout
	if timeout <= 0 {
		timeout = defaultHandshakeTimeout
	}
	return timeout, true
}

func (a *connectionAdmission) finishPending(c net.Conn) {
	if !a.connections[c] {
		return
	}
	a.connections[c] = false
	a.pending--
	key := sourceKey(c)
	a.sources[key]--
	if a.sources[key] == 0 {
		delete(a.sources, key)
	}
}

func (a *connectionAdmission) authenticated(c net.Conn) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.finishPending(c)
}

func (a *connectionAdmission) release(c net.Conn) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.finishPending(c)
	delete(a.connections, c)
}

func (a *connectionAdmission) stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.stopped = true
	for c := range a.connections {
		_ = c.Close()
	}
}
