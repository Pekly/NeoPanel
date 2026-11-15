package desktop

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Pekly/NeoPanel/internal/config"
)

func TestDesktopConfigSeparateAndLoopbackOnly(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "desktop data")
	t.Setenv("NEOPANEL_DATA_DIR", filepath.Join(t.TempDir(), "console"))
	t.Setenv("NEOPANEL_ROLE", "node")
	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.IsNode() || cfg.Server.Host != "127.0.0.1" || cfg.Server.Port != 0 || cfg.Server.Mode != "standalone" || cfg.SFTP.Enabled {
		t.Fatalf("unsafe desktop defaults: %+v", cfg)
	}
	if cfg.Paths.DataDir != dir || cfg.Paths.DBPath != filepath.Join(dir, "neopanel.db") {
		t.Fatalf("desktop used console data: %+v", cfg.Paths)
	}
	secret := cfg.Security.JWTSecret
	cfg.Server.Host, cfg.Server.Port, cfg.Server.Role = "0.0.0.0", 12345, "node"
	cfg.SFTP.Enabled = true
	if err := config.SaveConfig(cfg.FilePath, cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Server.Host != "127.0.0.1" || loaded.IsNode() || loaded.SFTP.Enabled {
		t.Fatal("desktop did not override network exposure")
	}
	if loaded.Server.Port != 12345 || loaded.Security.JWTSecret != secret {
		t.Fatal("desktop lost stable port/credentials")
	}
}

func TestWindowStateRoundTripAndBounds(t *testing.T) {
	dir := t.TempDir()
	if got := LoadWindowState(dir); got.Width != 1280 || got.Height != 820 {
		t.Fatal(got)
	}
	for _, state := range []WindowState{{Width: 1000, Height: 700, Maximized: true}, {Width: 1440, Height: 900}, {Width: -1, Height: 99999}} {
		if err := SaveWindowState(dir, state); err != nil {
			t.Fatal(err)
		}
		if got := LoadWindowState(dir); got != state.normalized() {
			t.Fatal(got, state)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "window.json"), []byte("corrupt"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := LoadWindowState(dir); got.Width != 1280 {
		t.Fatal(got)
	}
}

func TestInstanceIdentity(t *testing.T) {
	dir, err := DataDirectory(t.TempDir())
	if err != nil || !filepath.IsAbs(dir) {
		t.Fatal(dir, err)
	}
	if InstanceID(dir) != InstanceID(filepath.Join(dir, ".")) {
		t.Fatal("noncanonical identity")
	}
	if InstanceID(dir) == InstanceID(dir+"-other") {
		t.Fatal("instances collide")
	}
}
