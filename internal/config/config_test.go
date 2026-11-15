package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFirstRunAppliesEnvironmentCreatesDirectoriesAndPersistsSecrets(t *testing.T) {
	root := t.TempDir()
	t.Setenv("NEOPANEL_DATA_DIR", filepath.Join(root, "data with spaces"))
	t.Setenv("NEOPANEL_ROLE", "node")
	t.Setenv("NEOPANEL_NODE_TOKEN", "neo_node_config_test")
	t.Setenv("NEOPANEL_TURNSTILE_ENABLED", "true")
	t.Setenv("NEOPANEL_TURNSTILE_SITE_KEY", "site")
	file := filepath.Join(root, "config", "config.yaml")
	cfg, err := LoadConfig(file)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.IsNode() || cfg.Security.NodeToken != "neo_node_config_test" || !cfg.Security.Turnstile.Enabled || cfg.Security.Turnstile.SiteKey != "site" {
		t.Fatalf("first-run overrides ignored: %+v", cfg)
	}
	for _, dir := range []string{cfg.Paths.DataDir, cfg.Paths.ServersDir, filepath.Dir(cfg.Paths.DBPath)} {
		if st, err := os.Stat(dir); err != nil || !st.IsDir() {
			t.Fatal("directory not initialized", dir, err)
		}
	}
	again, err := LoadConfig(file)
	if err != nil || again.Security.JWTSecret != cfg.Security.JWTSecret || again.Security.NodeToken != cfg.Security.NodeToken {
		t.Fatal("restart changed secrets", err)
	}
}

func TestExplicitRoleWinsAndNodeGetsPersistentToken(t *testing.T) {
	t.Setenv("NEOPANEL_DATA_DIR", filepath.Join(t.TempDir(), "data"))
	t.Setenv("NEOPANEL_ROLE", "panel")
	t.Setenv("NEOPANEL_NODE_TOKEN", "")
	file := filepath.Join(t.TempDir(), "config.yaml")
	cfg, err := LoadConfigForRole(file, "node")
	if err != nil || !cfg.IsNode() || cfg.Security.NodeToken == "" {
		t.Fatal(cfg, err)
	}
	again, err := LoadConfigForRole(file, "node")
	if err != nil || cfg.Security.NodeToken != again.Security.NodeToken {
		t.Fatal("node credential was not persisted", err)
	}
}

func TestWindowsDefaultsAreLocalAndRoleSeparated(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows defaults")
	}
	t.Setenv("NEOPANEL_DATA_DIR", "")
	t.Setenv("LOCALAPPDATA", t.TempDir())
	panel, node := DefaultConfigForRole("panel"), DefaultConfigForRole("node")
	if panel.Server.Host != "127.0.0.1" || panel.Server.Mode != "standalone" || panel.SFTP.Enabled {
		t.Fatal("unsafe Windows defaults", panel)
	}
	if !filepath.IsAbs(panel.Paths.DataDir) || panel.Paths.DataDir == node.Paths.DataDir || panel.Paths.DBPath == node.Paths.DBPath {
		t.Fatal("panel/node storage overlaps", panel.Paths, node.Paths)
	}
}

func TestWindowsRelativePathsAnchorToConfig(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows paths")
	}
	t.Setenv("NEOPANEL_ROLE", "")
	dir := t.TempDir()
	file := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(file, []byte("paths:\n  data_dir: portable\nsecurity:\n  jwt_secret: fixed-test-secret\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(file)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Paths.DataDir != filepath.Join(dir, "portable") || cfg.Paths.ServersDir != filepath.Join(dir, "portable", "servers") || cfg.Paths.DBPath != filepath.Join(dir, "portable", "neopanel.db") || cfg.SFTP.KeyPath != filepath.Join(dir, "portable", "sftp_host_key") {
		t.Fatal("relative paths depend on process cwd", cfg.Paths, cfg.SFTP)
	}
}

func TestMissingJWTSecretIsRepairedOnceAndFilesystemErrorsSurface(t *testing.T) {
	t.Setenv("NEOPANEL_ROLE", "")
	dir := t.TempDir()
	t.Setenv("NEOPANEL_DATA_DIR", filepath.Join(dir, "data"))
	file := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(file, []byte("server:\n  role: panel\n"), 0600); err != nil {
		t.Fatal(err)
	}
	first, err := LoadConfig(file)
	if err != nil {
		t.Fatal(err)
	}
	second, err := LoadConfig(file)
	if err != nil || first.Security.JWTSecret != second.Security.JWTSecret {
		t.Fatal("missing secret changes every restart", err)
	}
	blocked := filepath.Join(dir, "not-a-directory")
	if err := os.WriteFile(blocked, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	first.Paths.DataDir = blocked
	if err := SaveConfig(file, first); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig(file); err == nil {
		t.Fatal("directory creation failure was hidden")
	}
}
