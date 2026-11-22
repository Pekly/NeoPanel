package proxyconfig

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// Stock schema 2.8, comments elided. Source: PaperMC/Velocity
// proxy/src/main/resources/default-velocity.toml (dev/3.0.0 branch).
// ForcedHosts in proxy/.../config/VelocityConfiguration.java also accepts strings.
const stockVelocity = `config-version = "2.8"
bind = "0.0.0.0:25565"
motd = "<#09add3>A Velocity Server"
show-max-players = 500
online-mode = true
force-key-authentication = true
prevent-client-proxy-connections = false
player-info-forwarding-mode = "NONE"
forwarding-secret-file = "forwarding.secret"
announce-forge = false
kick-existing-players = false
ping-passthrough = "DISABLED"
sample-players-in-ping = false
enable-player-address-logging = true
[packet-limiter]
interval = 7
packets-per-second = -1
bytes-per-second = -1
decompressed-bytes-per-second = 5242880
[servers]
lobby = "127.0.0.1:30066"
factions = "127.0.0.1:30067"
minigames = "127.0.0.1:30068"
try = ["lobby"]
[forced-hosts]
"lobby.example.com" = ["lobby"]
"factions.example.com" = ["factions"]
"minigames.example.com" = ["minigames"]
[advanced]
compression-threshold = 256
compression-level = -1
login-ratelimit = 3000
connection-timeout = 5000
read-timeout = 30000
haproxy-protocol = false
tcp-fast-open = false
bungee-plugin-message-channel = true
show-ping-requests = false
failover-on-unexpected-server-disconnect = true
announce-proxy-commands = true
log-command-executions = false
log-player-connections = true
accepts-transfers = false
enable-reuse-port = false
command-rate-limit = 50
forward-commands-if-rate-limited = true
kick-after-rate-limited-commands = 0
tab-complete-rate-limit = 10
kick-after-rate-limited-tab-completes = 0
[query]
enabled = false
port = 25565
map = "Velocity"
show-plugins = false
`

func TestVelocityStockCompatibility(t *testing.T) {
	for name, text := range map[string]string{
		"stock-2.8":                  stockVelocity,
		"single-string-forced-host":  strings.Replace(stockVelocity, `"lobby.example.com" = ["lobby"]`, `"lobby.example.com" = "lobby"`, 1),
		"lowercase-modes":            strings.ReplaceAll(strings.ReplaceAll(stockVelocity, `"NONE"`, `"none"`), `"DISABLED"`, `"disabled"`),
		"rewritten-root-and-metrics": "root-plugin-setting = 'keep'\n" + stockVelocity + "\n[metrics]\nenabled = true\nid = 'test-id'\n",
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			put(t, dir, "velocity.toml", []byte(text))
			put(t, dir, "forwarding.secret", []byte("keep-existing-secret"))
			before := doc(t, dir, "velocity")
			s := read(t, dir, "velocity")
			if len(s.Config.Backends) != 3 || s.Config.ForwardingMode != "NONE" {
				t.Fatal(s)
			}
			if !bytes.Equal(get(t, dir, "velocity.toml"), []byte(text)) {
				t.Fatal("Read changed file")
			}
			s.Config.MOTD = "Changed deliberately"
			if _, err := Save(dir, "velocity", 25590, s.Revision, s.Config); err != nil {
				t.Fatal(err)
			}
			after := doc(t, dir, "velocity")
			for _, key := range []string{"packet-limiter", "advanced", "sample-players-in-ping", "config-version", "root-plugin-setting", "metrics"} {
				if !reflect.DeepEqual(before[key], after[key]) {
					t.Fatal("lost setting", key)
				}
			}
			if string(get(t, dir, "forwarding.secret")) != "keep-existing-secret" {
				t.Fatal("secret rotated")
			}
		})
	}
}

func TestVelocityRecoverableValues(t *testing.T) {
	for _, change := range []struct{ old, value string }{
		{"show-max-players = 500", "show-max-players = -1"},
		{`"NONE"`, `"typo"`},
		{`"DISABLED"`, `"typo"`},
		{`"127.0.0.1:30066"`, `"bad address"`},
		{`try = ["lobby"]`, `try = ["missing"]`},
	} {
		t.Run(change.value, func(t *testing.T) {
			dir := t.TempDir()
			raw := []byte(strings.Replace(stockVelocity, change.old, change.value, 1))
			put(t, dir, "velocity.toml", raw)
			s := read(t, dir, "velocity")
			if Validate("velocity", s.Config) == nil || s.ValidationError == "" {
				t.Fatal("invalid value lost")
			}
			if _, err := Save(dir, "velocity", 25590, s.Revision, s.Config); err == nil {
				t.Fatal("invalid save accepted")
			}
			if !bytes.Equal(raw, get(t, dir, "velocity.toml")) {
				t.Fatal("rejected save changed file")
			}
			if _, err := Save(dir, "velocity", 25590, s.Revision, usable("velocity")); err != nil {
				t.Fatal("cannot repair editable value", err)
			}
		})
	}
}

func TestVelocityEmptySectionsAndPlayerBounds(t *testing.T) {
	for _, text := range []string{"[servers]\n[forced-hosts]\n", "show-max-players = 0\n", "show-max-players = 2147483647\n"} {
		dir := t.TempDir()
		put(t, dir, "velocity.toml", []byte(text))
		s := read(t, dir, "velocity")
		data, _ := json.Marshal(s)
		if bytes.Contains(data, []byte(":null")) {
			t.Fatal("null collection", string(data))
		}
	}
	for _, text := range []string{"show-max-players = 2147483648\n", "show-max-players = 1.2\n", "online-mode = 'false'\n", "[forced-hosts]\n'example.com' = 12\n", "[servers]\ntry = 'lobby'\n"} {
		dir := t.TempDir()
		put(t, dir, "velocity.toml", []byte(text))
		if _, err := Read(dir, "velocity", 25565); err == nil {
			t.Fatal("invalid structure accepted", text)
		}
	}
}

func TestVelocityReadCollectionLimits(t *testing.T) {
	for _, field := range []string{"backends", "try", "forced-hosts", "route-targets"} {
		t.Run(field, func(t *testing.T) {
			dir := t.TempDir()
			m, err := parse("velocity", []byte(stockVelocity))
			if err != nil {
				t.Fatal(err)
			}
			refs := make([]string, 129)
			for i := range refs {
				refs[i] = "lobby"
			}
			switch field {
			case "backends":
				for i := 0; i < 129; i++ {
					m["servers"].(object)[strings.Repeat("x", i+1)] = "localhost:25565"
				}
			case "try":
				m["servers"].(object)["try"] = refs
			case "forced-hosts":
				for i := 0; i < 129; i++ {
					m["forced-hosts"].(object)[strings.Repeat("x", i+1)] = []string{"lobby"}
				}
			case "route-targets":
				m["forced-hosts"].(object)["lobby.example.com"] = refs
			}
			writeDoc(t, dir, "velocity", m)
			if _, err := Read(dir, "velocity", 25565); err == nil {
				t.Fatal("collection limit bypassed")
			}
		})
	}
}
