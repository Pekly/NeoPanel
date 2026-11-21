package database

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"path/filepath"
	"strings"
	"testing"
)

func TestNodeWakePersistenceAndDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "panel.db")
	db, err := InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defaults := DefaultNodeWakeConfig()
	if defaults.Enabled || defaults.MaxPlayers != 20 || !defaults.UseServerIcon || defaults.CooldownSeconds != 30 || defaults.OfflineMOTD != "{server} is sleeping. Join to start!" || defaults.VersionLabel != "Join to start" {
		t.Fatal(defaults)
	}
	if got, err := db.GetNodeWakeConfig("local"); err != nil || got != defaults {
		t.Fatal(got, err)
	}
	c := defaults
	c.Enabled = true
	c.OfflineMOTD = "§a{server}\nSleeping 💤"
	c.BindAddress = "::1"
	if err := db.SetNodeWakeConfig("local", c); err != nil {
		t.Fatal(err)
	}
	remote := defaults
	remote.VersionLabel = "Remote"
	if err := db.SetNodeWakeConfig("remote", remote); err != nil {
		t.Fatal(err)
	}
	if nodes, err := db.ListNodes(); err != nil || len(nodes) != 0 {
		t.Fatal("wake inserted synthetic node", nodes, err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if got, err := db.GetNodeWakeConfig("local"); err != nil || got != c {
		t.Fatal(got, err)
	}
	if got, err := db.GetNodeWakeConfig("remote"); err != nil || got != remote {
		t.Fatal(got, err)
	}
	bad := c
	bad.CooldownSeconds = 0
	if err := db.SetNodeWakeConfig("local", bad); err == nil {
		t.Fatal("invalid config persisted")
	}
	if got, err := db.GetNodeWakeConfig("local"); err != nil || got != c {
		t.Fatal(got, err)
	}
	key, _ := nodeWakeSettingKey("local")
	if err := db.SetPanelSetting(key, `{"enabled":true,"cooldown_seconds":0}`); err != nil {
		t.Fatal(err)
	}
	if got, err := db.GetNodeWakeConfig("local"); err == nil || got.Enabled {
		t.Fatal("corrupt config did not fail closed", got, err)
	}
	if _, err := db.GetNodeWakeConfig(""); err == nil {
		t.Fatal("empty ID accepted")
	}
}

func TestNodeWakeValidation(t *testing.T) {
	for name, mutate := range map[string]func(*NodeWakeConfig){
		"hostname":         func(c *NodeWakeConfig) { c.BindAddress = "localhost" },
		"port":             func(c *NodeWakeConfig) { c.BindAddress = "127.0.0.1:25565" },
		"zone":             func(c *NodeWakeConfig) { c.BindAddress = "fe80::1%eth0" },
		"motd bytes":       func(c *NodeWakeConfig) { c.OfflineMOTD = strings.Repeat("é", 1025) },
		"label":            func(c *NodeWakeConfig) { c.VersionLabel = strings.Repeat("x", 129) },
		"starting":         func(c *NodeWakeConfig) { c.StartingMessage = strings.Repeat("x", 4097) },
		"unavailable":      func(c *NodeWakeConfig) { c.UnavailableMessage = strings.Repeat("x", 4097) },
		"UTF8":             func(c *NodeWakeConfig) { c.OfflineMOTD = "\xff" },
		"NUL":              func(c *NodeWakeConfig) { c.StartingMessage = "x\x00" },
		"C1":               func(c *NodeWakeConfig) { c.VersionLabel = "\u0085" },
		"tab":              func(c *NodeWakeConfig) { c.OfflineMOTD = "\t" },
		"CR":               func(c *NodeWakeConfig) { c.UnavailableMessage = "\r" },
		"negative players": func(c *NodeWakeConfig) { c.MaxPlayers = -1 },
		"players":          func(c *NodeWakeConfig) { c.MaxPlayers = 100001 },
		"cooldown low":     func(c *NodeWakeConfig) { c.CooldownSeconds = 9 },
		"cooldown high":    func(c *NodeWakeConfig) { c.CooldownSeconds = 601 },
	} {
		t.Run(name, func(t *testing.T) {
			c := DefaultNodeWakeConfig()
			mutate(&c)
			if ValidateNodeWakeConfig(c) == nil {
				t.Fatal("accepted invalid config")
			}
		})
	}
	c := DefaultNodeWakeConfig()
	c.OfflineMOTD = strings.Repeat("é", 1024)
	c.VersionLabel = strings.Repeat("x", 128)
	c.StartingMessage = strings.Repeat("x", 4096)
	c.UnavailableMessage = "§aHello\nWorld"
	c.MaxPlayers = 100000
	c.CooldownSeconds = 600
	if err := ValidateNodeWakeConfig(c); err != nil {
		t.Fatal(err)
	}
	c.MaxPlayers = 0
	c.CooldownSeconds = 10
	c.BindAddress = "0.0.0.0"
	if err := ValidateNodeWakeConfig(c); err != nil {
		t.Fatal(err)
	}
}

func TestNodeWakeIconFullDecode(t *testing.T) {
	encode := func(w, h int) []byte {
		var b bytes.Buffer
		if err := png.Encode(&b, image.NewRGBA(image.Rect(0, 0, w, h))); err != nil {
			t.Fatal(err)
		}
		return b.Bytes()
	}
	uri := func(b []byte) string { return "data:image/png;base64," + base64.StdEncoding.EncodeToString(b) }
	good := encode(64, 64)
	for _, data := range []string{"", uri(good)} {
		if err := ValidateWakeIcon(data); err != nil {
			t.Fatal(err)
		}
	}
	// The 33-byte header contains valid IHDR dimensions, but no pixel data.
	if _, err := png.DecodeConfig(bytes.NewReader(good[:33])); err != nil {
		t.Fatal(err)
	}
	for _, data := range []string{"https://example.com/icon.png", "data:image/jpeg;base64,AAAA", "data:image/png;base64,!!!!", uri(encode(65, 64)), uri(good[:33]), uri(good[:len(good)-5]), "data:image/png;base64," + strings.Repeat("A", 65536)} {
		if ValidateWakeIcon(data) == nil {
			t.Fatal("accepted invalid image")
		}
	}
}
