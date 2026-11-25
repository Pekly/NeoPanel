package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	FilePath string         `yaml:"-"`
	Server   ServerConfig   `yaml:"server"`
	Paths    PathsConfig    `yaml:"paths"`
	Security SecurityConfig `yaml:"security"`
	Java     JavaConfig     `yaml:"java"`
	SFTP     SFTPConfig     `yaml:"sftp"`
	FTP      FTPConfig      `yaml:"ftp"`
}

type ServerConfig struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	Mode              string `yaml:"mode"` // "standalone" or "behind_proxy"
	Role              string `yaml:"role"` // "panel" (Master Panel + Daemon) or "node" (Cluster Node Daemon)
	StartupTimeoutSec int    `yaml:"startup_timeout_seconds"`
	PublicURL         string `yaml:"public_url"` // Canonical panel origin for passkeys; never a game/node FQDN.
}

type SFTPConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	KeyPath string `yaml:"key_path"`
}

type FTPConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

type PathsConfig struct {
	DataDir    string `yaml:"data_dir"`
	ServersDir string `yaml:"servers_dir"`
	DBPath     string `yaml:"db_path"`
}

type TurnstileConfig struct {
	Enabled   bool   `yaml:"enabled"`
	SiteKey   string `yaml:"site_key"`
	SecretKey string `yaml:"secret_key"`
}

type SecurityConfig struct {
	JWTSecret        string          `yaml:"jwt_secret"`
	TokenExpiration  int             `yaml:"token_expiration_hours"` // in hours (default 72)
	NodeToken        string          `yaml:"node_token"`             // cluster node authentication token (neo_node_...)
	Turnstile        TurnstileConfig `yaml:"turnstile"`
	PasskeysDisabled bool            `yaml:"passkeys_disabled"`
	// Explicit trusted-host opt-out; never a per-tenant/free/premium setting.
	AllowInsecureGameFallback bool `yaml:"allow_insecure_game_fallback"`
	// Empty means socket peers only. For local Caddy explicitly configure
	// ["127.0.0.1/32", "::1/128"]. Only X-Forwarded-For is accepted.
	TrustedProxies []string `yaml:"trusted_proxies,omitempty"`
	// Exact host:port exceptions for imports from private networks; no tenant override.
	SFTPImportAllowedHosts []string `yaml:"sftp_import_allowed_hosts,omitempty"`
}

type JavaConfig struct {
	DefaultPath string `yaml:"default_path"`
}

func DefaultConfig() *Config { return DefaultConfigForRole("panel") }

// Windows defaults are per-user and local-only. A headless node gets a separate
// data root so starting both binaries on one PC never shares their database.
func DefaultConfigForRole(role string) *Config {
	if role != "node" {
		role = "panel"
	}
	dataDir, host, mode, sftpEnabled := "./data", "0.0.0.0", "behind_proxy", true
	if runtime.GOOS == "windows" {
		base, err := os.UserCacheDir()
		if err != nil || base == "" {
			base, _ = os.UserHomeDir()
		}
		name := "NeoPanel"
		if role == "node" {
			name = "NeoPanel-node"
		}
		dataDir = filepath.Join(base, name)
		host, mode, sftpEnabled = "127.0.0.1", "standalone", false
	}
	if value := strings.TrimSpace(os.Getenv("NEOPANEL_DATA_DIR")); value != "" {
		dataDir = value
	}
	secret := generateRandomKey(32)

	return &Config{
		Server: ServerConfig{
			Host:              host,
			Port:              8080,
			Mode:              mode,
			Role:              role,
			StartupTimeoutSec: 180,
		},
		Paths: PathsConfig{
			DataDir:    dataDir,
			ServersDir: filepath.Join(dataDir, "servers"),
			DBPath:     filepath.Join(dataDir, "neopanel.db"),
		},
		Security: SecurityConfig{
			JWTSecret:       secret,
			TokenExpiration: 72,
		},
		Java: JavaConfig{
			DefaultPath: "java",
		},
		SFTP: SFTPConfig{
			Enabled: sftpEnabled,
			Port:    2022,
			KeyPath: filepath.Join(dataDir, "sftp_host_key"),
		},
		FTP: FTPConfig{
			Enabled: sftpEnabled,
			Port:    2121,
		},
	}
}

func LoadConfig(path string) (*Config, error) { return LoadConfigForRole(path, "") }

// An explicit CLI/build role wins over YAML/environment. Paths remain operator
// configuration; relative Windows paths are anchored to the configuration file.
func LoadConfigForRole(path, role string) (*Config, error) {
	cfg := DefaultConfigForRole(role)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration path: %w", err)
	}
	cfg.FilePath = absPath
	data, err := os.ReadFile(absPath)
	created := os.IsNotExist(err)
	if err != nil && !created {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var present struct {
		Paths struct {
			DataDir    *string `yaml:"data_dir"`
			ServersDir *string `yaml:"servers_dir"`
			DBPath     *string `yaml:"db_path"`
		} `yaml:"paths"`
		Security struct {
			JWTSecret *string `yaml:"jwt_secret"`
		} `yaml:"security"`
		SFTP struct {
			KeyPath *string `yaml:"key_path"`
		} `yaml:"sftp"`
	}
	if !created {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
		if err := yaml.Unmarshal(data, &present); err != nil {
			return nil, fmt.Errorf("failed to parse config fields: %w", err)
		}
	}
	if envRole := strings.ToLower(strings.TrimSpace(os.Getenv("NEOPANEL_ROLE"))); envRole == "node" || envRole == "panel" {
		cfg.Server.Role = envRole
	}
	if role == "node" || role == "panel" {
		cfg.Server.Role = role
	}
	cfg.Server.Role = strings.ToLower(strings.TrimSpace(cfg.Server.Role))
	if cfg.Server.Role == "" {
		cfg.Server.Role = "panel"
	}
	if cfg.Server.Role != "panel" && cfg.Server.Role != "node" {
		return nil, fmt.Errorf("invalid server role %q", cfg.Server.Role)
	}
	defaults := DefaultConfigForRole(cfg.Server.Role)
	if present.Paths.DataDir == nil {
		cfg.Paths.DataDir = defaults.Paths.DataDir
	}
	if present.Paths.ServersDir == nil {
		cfg.Paths.ServersDir = filepath.Join(cfg.Paths.DataDir, "servers")
	}
	if present.Paths.DBPath == nil {
		cfg.Paths.DBPath = filepath.Join(cfg.Paths.DataDir, "neopanel.db")
	}
	if present.SFTP.KeyPath == nil {
		cfg.SFTP.KeyPath = filepath.Join(cfg.Paths.DataDir, "sftp_host_key")
	}
	if runtime.GOOS == "windows" {
		for _, value := range []*string{&cfg.Paths.DataDir, &cfg.Paths.ServersDir, &cfg.Paths.DBPath, &cfg.SFTP.KeyPath} {
			if *value != "" && !filepath.IsAbs(*value) {
				*value = filepath.Join(filepath.Dir(absPath), *value)
			}
		}
	}
	if cfg.Paths.DataDir == "" || cfg.Paths.ServersDir == "" || cfg.Paths.DBPath == "" {
		return nil, fmt.Errorf("data, servers, and database paths must not be empty")
	}
	if cfg.Server.Port < 0 || cfg.Server.Port > 65535 {
		return nil, fmt.Errorf("invalid HTTP port %d", cfg.Server.Port)
	}
	if envToken := strings.TrimSpace(os.Getenv("NEOPANEL_NODE_TOKEN")); envToken != "" {
		cfg.Security.NodeToken = envToken
	}
	persist := created
	if present.Security.JWTSecret == nil || strings.TrimSpace(cfg.Security.JWTSecret) == "" {
		cfg.Security.JWTSecret = generateRandomKey(32)
		persist = true
	}
	if cfg.IsNode() && strings.TrimSpace(cfg.Security.NodeToken) == "" {
		cfg.Security.NodeToken = GenerateNodeToken()
		persist = true
	}
	if value := os.Getenv("NEOPANEL_TURNSTILE_ENABLED"); value != "" {
		cfg.Security.Turnstile.Enabled = value == "true" || value == "1"
	}
	if value := os.Getenv("NEOPANEL_TURNSTILE_SITE_KEY"); value != "" {
		cfg.Security.Turnstile.SiteKey = value
	}
	if value := os.Getenv("NEOPANEL_TURNSTILE_SECRET_KEY"); value != "" {
		cfg.Security.Turnstile.SecretKey = value
	}
	for _, dir := range []string{cfg.Paths.DataDir, cfg.Paths.ServersDir, filepath.Dir(cfg.Paths.DBPath)} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("cannot create data directory %q: %w", dir, err)
		}
	}
	if persist {
		if err := os.MkdirAll(filepath.Dir(absPath), 0700); err != nil {
			return nil, err
		}
		if err := SaveConfig(absPath, cfg); err != nil {
			return nil, fmt.Errorf("failed to save initial configuration: %w", err)
		}
	}
	return cfg, nil
}

func (c *Config) IsNode() bool {
	return c != nil && strings.EqualFold(strings.TrimSpace(c.Server.Role), "node")
}

func GenerateNodeToken() string {
	return "neo_node_" + generateRandomKey(28)
}

func SaveConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0640)
}

func generateRandomKey(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "fallback-secret-key-neopanel-ultra-lightweight-2026"
	}
	return hex.EncodeToString(bytes)
}
