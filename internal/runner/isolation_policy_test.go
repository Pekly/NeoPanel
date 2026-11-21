package runner

import (
	"testing"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
)

func TestLinuxAutomaticIsolationNeverSilentlyDowngrades(t *testing.T) {
	for _, engine := range []string{"auto", "bwrap"} {
		for _, bwrap := range []bool{false, true} {
			for _, systemd := range []bool{false, true} {
				err := requireLinuxSandboxTools("linux", engine, false, bwrap, systemd)
				if (err == nil) != (bwrap && systemd) {
					t.Fatal(engine, bwrap, systemd, err)
				}
				if err := requireLinuxSandboxTools("linux", engine, true, bwrap, systemd); err != nil {
					t.Fatal("explicit trusted-host fallback unavailable", err)
				}
			}
		}
	}
	for _, engine := range []string{"process", "none", "systemd"} {
		if err := requireLinuxSandboxTools("linux", engine, false, false, false); err != nil {
			t.Fatal("explicit mode semantics changed", engine, err)
		}
	}
	if err := requireLinuxSandboxTools("windows", "auto", false, false, false); err != nil {
		t.Fatal("Windows trusted hosting broken", err)
	}
}
func TestIsolationLegacyNamesAndHostOverride(t *testing.T) {
	for input, want := range map[string]string{"": "auto", " Bubblewrap ": "bwrap", "Bubblewrap + Systemd Slices": "bwrap", "SYSTEMD": "systemd", "process": "process", "bogus": "bogus"} {
		if got := normalizeIsolationEngine(input); got != want {
			t.Fatal(input, got, want)
		}
	}
	if config.DefaultConfig().Security.AllowInsecureGameFallback {
		t.Fatal("unsafe fallback enabled by default")
	}
	m := NewServerManager(t.TempDir(), nil)
	defer m.ShutdownAll()
	first := m.getOrCreateInstance(&database.Server{ID: "one"})
	if first.allowInsecureGameFallback {
		t.Fatal("default instance weakens isolation")
	}
	m.SetAllowInsecureGameFallback(true)
	second := m.getOrCreateInstance(&database.Server{ID: "two"})
	if !first.allowInsecureGameFallback || !second.allowInsecureGameFallback {
		t.Fatal("host override not propagated")
	}
	m.SetAllowInsecureGameFallback(false)
	if first.allowInsecureGameFallback || second.allowInsecureGameFallback {
		t.Fatal("host override not cleared")
	}
}
