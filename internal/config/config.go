package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
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
	// Empty means socket peers only. For local Caddy explicitly configure
	// ["127.0.0.1/32", "::1/128"]. Only X-Forwarded-For is accepted.
	TrustedProxies []string `yaml:"trusted_proxies,omitempty"`
}

type JavaConfig struct {
	DefaultPath string `yaml:"default_path"`
}

func DefaultConfig() *Config {
	dataDir := "./data"
	if os.Getenv("NEOPANEL_DATA_DIR") != "" {
		dataDir = os.Getenv("NEOPANEL_DATA_DIR")
	}

	secret := generateRandomKey(32)

	return &Config{
		Server: ServerConfig{
			Host:              "0.0.0.0",
			Port:              8080,
			Mode:              "behind_proxy",
			Role:              "panel",
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
			Enabled: true,
			Port:    2022,
			KeyPath: filepath.Join(dataDir, "sftp_host_key"),
		},
	}
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()
	cfg.FilePath = path

	// If file doesn't exist, try to save default config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Ensure directory exists
		dir := filepath.Dir(path)
		if dir != "." && dir != "" {
			_ = os.MkdirAll(dir, 0755)
		}
		if err := SaveConfig(path, cfg); err != nil {
			return nil, fmt.Errorf("failed to write initial config: %w", err)
		}
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Optional environment variable overrides for Role, NodeToken, and Turnstile
	if envRole := strings.ToLower(strings.TrimSpace(os.Getenv("NEOPANEL_ROLE"))); envRole == "node" || envRole == "panel" {
		cfg.Server.Role = envRole
	}
	if strings.TrimSpace(cfg.Server.Role) == "" {
		cfg.Server.Role = "panel"
	} else {
		cfg.Server.Role = strings.ToLower(strings.TrimSpace(cfg.Server.Role))
	}
	if envNodeToken := strings.TrimSpace(os.Getenv("NEOPANEL_NODE_TOKEN")); envNodeToken != "" {
		cfg.Security.NodeToken = envNodeToken
	}
	if cfg.IsNode() && strings.TrimSpace(cfg.Security.NodeToken) == "" {
		cfg.Security.NodeToken = GenerateNodeToken()
		if path != "" {
			_ = SaveConfig(path, cfg)
		}
	}

	// Optional environment variable overrides for Turnstile
	if envEnabled := os.Getenv("NEOPANEL_TURNSTILE_ENABLED"); envEnabled != "" {
		cfg.Security.Turnstile.Enabled = envEnabled == "true" || envEnabled == "1"
	}
	if envSiteKey := os.Getenv("NEOPANEL_TURNSTILE_SITE_KEY"); envSiteKey != "" {
		cfg.Security.Turnstile.SiteKey = envSiteKey
	}
	if envSecretKey := os.Getenv("NEOPANEL_TURNSTILE_SECRET_KEY"); envSecretKey != "" {
		cfg.Security.Turnstile.SecretKey = envSecretKey
	}

	// Ensure directories exist
	_ = os.MkdirAll(cfg.Paths.DataDir, 0755)
	_ = os.MkdirAll(cfg.Paths.ServersDir, 0755)

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
