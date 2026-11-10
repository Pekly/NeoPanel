package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

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
	StartupTimeoutSec int    `yaml:"startup_timeout_seconds"`
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
	JWTSecret       string          `yaml:"jwt_secret"`
	TokenExpiration int             `yaml:"token_expiration_hours"` // in hours (default 72)
	Turnstile       TurnstileConfig `yaml:"turnstile"`
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
