package runner

import (
	"fmt"
	"strings"
)

func normalizeIsolationEngine(engine string) string {
	engine = strings.ToLower(strings.TrimSpace(engine))
	switch engine {
	case "":
		return "auto"
	case "bubblewrap", "bubblewrap + systemd slices":
		return "bwrap"
	}
	return engine
}

func requireLinuxSandboxTools(goos, engine string, allowInsecure, bwrap, systemd bool) error {
	if goos != "linux" || allowInsecure || (engine != "auto" && engine != "bwrap") {
		return nil
	}
	if !bwrap || !systemd {
		return fmt.Errorf("Linux %s isolation requires both bubblewrap and systemd-run; install/enable them before starting untrusted workloads. No automatic downgrade is allowed (security.allow_insecure_game_fallback is for explicitly trusted hosts only)", engine)
	}
	return nil
}

// SetAllowInsecureGameFallback is host configuration, not tenant preference.
// Existing processes retain their current sandbox; this controls future starts.
func (m *ServerManager) SetAllowInsecureGameFallback(allow bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.allowInsecureGameFallback = allow
	for _, inst := range m.instances {
		inst.mu.Lock()
		inst.allowInsecureGameFallback = allow
		inst.mu.Unlock()
	}
}
