package database

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"net"
	"strings"
	"unicode"
	"unicode/utf8"
)

// NodeWakeConfig is desired configuration on the panel and applied configuration
// under "local" on the host. It is independent of ClusterNode policy defaults.
type NodeWakeConfig struct {
	Enabled            bool   `json:"enabled"`
	BindAddress        string `json:"bind_address"`
	OfflineMOTD        string `json:"offline_motd"`
	VersionLabel       string `json:"version_label"`
	MaxPlayers         int    `json:"max_players"`
	IconDataURL        string `json:"icon_data_url"`
	UseServerIcon      bool   `json:"use_server_icon"`
	StartingMessage    string `json:"starting_message"`
	UnavailableMessage string `json:"unavailable_message"`
	CooldownSeconds    int    `json:"cooldown_seconds"`
}

func DefaultNodeWakeConfig() NodeWakeConfig {
	return NodeWakeConfig{
		OfflineMOTD:  "&9&lNeoPanel &8» &f{server}\n&a&lJOIN TO START &8• &7Your adventure is one click away",
		VersionLabel: "&cSleeping &8• &aJoin to start", MaxPlayers: 20, UseServerIcon: true,
		StartingMessage:    "&9&lNeoPanel &8» &f{server}\n\n&a&lWAKE REQUEST ACCEPTED\n&7Your server is starting up.\n&ePlease reconnect in a moment!",
		UnavailableMessage: "&9&lNeoPanel &8» &f{server}\n\n&6&lNOT READY YET\n&7The server cannot accept a wake request right now.\n&ePlease wait a moment and try again.",
		CooldownSeconds:    30,
	}
}

func ValidateNodeWakeConfig(c NodeWakeConfig) error {
	if c.BindAddress != "" && net.ParseIP(c.BindAddress) == nil {
		return fmt.Errorf("bind_address must be empty or a literal IP address")
	}
	for _, field := range []struct {
		name, value string
		limit       int
	}{
		{"offline_motd", c.OfflineMOTD, 2048}, {"version_label", c.VersionLabel, 128},
		{"starting_message", c.StartingMessage, 4096}, {"unavailable_message", c.UnavailableMessage, 4096},
	} {
		if !utf8.ValidString(field.value) || len(field.value) > field.limit {
			return fmt.Errorf("%s must be valid UTF-8 of at most %d bytes", field.name, field.limit)
		}
		for _, r := range field.value {
			if unicode.IsControl(r) && r != '\n' {
				return fmt.Errorf("%s must not contain control characters other than newline", field.name)
			}
		}
	}
	if c.MaxPlayers < 0 || c.MaxPlayers > 100000 {
		return fmt.Errorf("max_players must be between 0 and 100000")
	}
	if c.CooldownSeconds < 10 || c.CooldownSeconds > 600 {
		return fmt.Errorf("cooldown_seconds must be between 10 and 600")
	}
	return ValidateWakeIcon(c.IconDataURL)
}

// ValidateWakeIcon never fetches URLs. Check dimensions before allocating a
// decoded image, then fully decode to reject truncated/corrupt PNG pixel data.
func ValidateWakeIcon(dataURL string) error {
	if dataURL == "" {
		return nil
	}
	const prefix = "data:image/png;base64,"
	if len(dataURL) > 64*1024 || !strings.HasPrefix(dataURL, prefix) {
		return fmt.Errorf("icon_data_url must be a PNG data URL of at most 65536 bytes")
	}
	payload := strings.TrimPrefix(dataURL, prefix)
	if strings.ContainsAny(payload, "\r\n") {
		return fmt.Errorf("invalid PNG base64")
	}
	data, err := base64.StdEncoding.Strict().DecodeString(payload)
	if err != nil {
		return fmt.Errorf("invalid PNG base64")
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil || cfg.Width != 64 || cfg.Height != 64 {
		return fmt.Errorf("icon must be a 64x64 PNG")
	}
	if _, err := png.Decode(bytes.NewReader(data)); err != nil {
		return fmt.Errorf("invalid PNG image: %w", err)
	}
	return nil
}

func nodeWakeSettingKey(nodeID string) (string, error) {
	if nodeID == "" || len(nodeID) > 256 || !utf8.ValidString(nodeID) {
		return "", fmt.Errorf("invalid node ID")
	}
	return "node_wake_config:" + nodeID, nil
}

func (db *DB) GetNodeWakeConfig(nodeID string) (NodeWakeConfig, error) {
	c := DefaultNodeWakeConfig()
	key, err := nodeWakeSettingKey(nodeID)
	if err != nil {
		return c, err
	}
	raw, err := db.GetPanelSetting(key)
	if errors.Is(err, sql.ErrNoRows) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		return DefaultNodeWakeConfig(), err
	}
	if err := ValidateNodeWakeConfig(c); err != nil {
		return DefaultNodeWakeConfig(), err
	}
	return c, nil
}

func (db *DB) SetNodeWakeConfig(nodeID string, c NodeWakeConfig) error {
	key, err := nodeWakeSettingKey(nodeID)
	if err != nil {
		return err
	}
	if err := ValidateNodeWakeConfig(c); err != nil {
		return err
	}
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return db.SetPanelSetting(key, string(data))
}
