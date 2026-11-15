// Package desktop contains the Windows desktop shell's configuration and lifecycle helpers.
package desktop

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Pekly/NeoPanel/internal/config"
)

const Title = "NeoPanel Desktop"
const Version = "1.0.0-windows-beta.1"

func DataDirectory(override string) (string, error) {
	if strings.TrimSpace(override) == "" {
		base, err := os.UserCacheDir()
		if err != nil {
			return "", err
		}
		override = filepath.Join(base, "NeoPanel-Desktop")
	}
	return filepath.Abs(override)
}

func InstanceID(dataDir string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(filepath.Clean(dataDir))))
	return "NeoPanel.Desktop." + hex.EncodeToString(sum[:12])
}

func LoadConfig(dataDir string) (*config.Config, error) {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, err
	}
	path := filepath.Join(dataDir, "config.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg := config.DefaultConfigForRole("panel")
		cfg.Paths.DataDir = dataDir
		cfg.Paths.ServersDir = filepath.Join(dataDir, "servers")
		cfg.Paths.DBPath = filepath.Join(dataDir, "neopanel.db")
		cfg.SFTP.KeyPath = filepath.Join(dataDir, "sftp_host_key")
		cfg.Server.Host, cfg.Server.Port, cfg.Server.Mode = "127.0.0.1", 0, "standalone"
		cfg.SFTP.Enabled = false
		if err := config.SaveConfig(path, cfg); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	cfg, err := config.LoadConfigForRole(path, "panel")
	if err != nil {
		return nil, err
	}
	// The desktop never silently publishes control APIs or starts an SFTP listener.
	cfg.Server.Host, cfg.Server.Mode, cfg.Server.Role = "127.0.0.1", "standalone", "panel"
	cfg.SFTP.Enabled = false
	return cfg, nil
}

type WindowState struct {
	Width     int  `json:"width"`
	Height    int  `json:"height"`
	Maximized bool `json:"maximized"`
}

func (s WindowState) normalized() WindowState {
	if s.Width < 900 || s.Width > 3840 {
		s.Width = 1280
	}
	if s.Height < 620 || s.Height > 2160 {
		s.Height = 820
	}
	return s
}

func LoadWindowState(dir string) WindowState {
	var state WindowState
	if data, err := os.ReadFile(filepath.Join(dir, "window.json")); err == nil {
		_ = json.Unmarshal(data, &state)
	}
	return state.normalized()
}

func SaveWindowState(dir string, state WindowState) error {
	data, err := json.MarshalIndent(state.normalized(), "", "  ")
	if err != nil {
		return err
	}
	temp := filepath.Join(dir, "window.json.tmp")
	if err := os.WriteFile(temp, data, 0600); err != nil {
		return err
	}
	if err := os.Rename(temp, filepath.Join(dir, "window.json")); err != nil {
		_ = os.Remove(temp)
		return fmt.Errorf("save window state: %w", err)
	}
	return nil
}
