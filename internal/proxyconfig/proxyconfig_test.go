package proxyconfig

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

func put(t *testing.T, dir, name string, b []byte) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), b, 0600); err != nil {
		t.Fatal(err)
	}
}
func get(t *testing.T, dir, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func read(t *testing.T, dir, kind string) Snapshot {
	t.Helper()
	s, err := Read(dir, kind, 25590)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func usable(kind string) Settings {
	s := defaults(kind)
	s.Backends = []Backend{{Name: "lobby", Address: "[::1]:25566", MOTD: "Hello\nworld"}}
	s.Try = []string{"lobby"}
	s.ForcedHosts = []ForcedHost{{Host: "play.example.com", Servers: []string{"lobby"}}}
	return s
}
func doc(t *testing.T, dir, kind string) object {
	t.Helper()
	m, err := parse(kind, get(t, dir, FileName(kind)))
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func writeDoc(t *testing.T, dir, kind string, m object) {
	t.Helper()
	b, err := encode(kind, m)
	if err != nil {
		t.Fatal(err)
	}
	put(t, dir, FileName(kind), b)
}
func TestStarter(t *testing.T) {
	for _, kind := range []string{"velocity", "bungeecord", "waterfall"} {
		t.Run(kind, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "new", "server")
			s, err := Read(dir, strings.ToUpper(kind), 25590)
			if err != nil || s.Exists || s.Revision != "missing" {
				t.Fatalf("missing: %+v %v", s, err)
			}
			if _, err = os.Stat(dir); !os.IsNotExist(err) {
				t.Fatal("Read created directory")
			}
			if err = SyncPort(dir, kind, 25590); !errors.Is(err, ErrMissing) {
				t.Fatal(err)
			}
			if _, err = Save(dir, kind, 25590, "missing", usable(kind)); !errors.Is(err, ErrMissing) {
				t.Fatal(err)
			}
			created, err := Ensure(dir, strings.ToUpper(kind), 25590)
			if err != nil || !created {
				t.Fatalf("Ensure: %v %v", created, err)
			}
			s = read(t, dir, kind)
			if !s.Exists || s.Type != kind || s.File != FileName(kind) || len(s.Revision) != 64 || !reflect.DeepEqual(s.Config, defaults(kind)) {
				t.Fatalf("starter: %+v", s)
			}
			if len(s.Config.Backends) != 0 || len(s.Config.Try) != 0 || len(s.Config.ForcedHosts) != 0 {
				t.Fatal("invented routes")
			}
			if Validate(kind, s.Config) == nil {
				t.Fatal("empty starter accepted for Save")
			}
			m := doc(t, dir, kind)
			if kind == "velocity" {
				if m["bind"] != "0.0.0.0:25590" || m["config-version"] != "2.7" || m["forwarding-secret-file"] != "forwarding.secret" {
					t.Fatal(m)
				}
				q := m["query"].(object)
				if q["enabled"] != false || q["port"] != int64(25590) {
					t.Fatal(q)
				}
				secret, err := ReadSecret(dir, kind)
				if err != nil || len(secret) != 64 {
					t.Fatalf("secret %v", err)
				}
				j, _ := json.Marshal(s)
				if bytes.Contains(j, []byte(secret)) || bytes.Contains(j, []byte("forwarding-secret")) {
					t.Fatal("secret exposed")
				}
				if runtime.GOOS != "windows" {
					st, _ := os.Stat(filepath.Join(dir, "forwarding.secret"))
					if st.Mode().Perm() != 0600 {
						t.Fatal(st.Mode())
					}
				}
			} else {
				l, _ := listener(m, true)
				if l["host"] != "0.0.0.0:25590" || l["query_port"] != 25590 || l["query_enabled"] != false || m["player_limit"] != -1 {
					t.Fatal(m)
				}
				if _, err = ReadSecret(dir, kind); err == nil {
					t.Fatal("Bungee secret reveal accepted")
				}
			}
			before := get(t, dir, FileName(kind))
			created, err = Ensure(dir, kind, 25600)
			if err != nil || created || !bytes.Equal(before, get(t, dir, FileName(kind))) {
				t.Fatal("existing config rewritten", err)
			}
			for _, n := range []string{"server.properties", "eula.txt"} {
				if _, err = os.Stat(filepath.Join(dir, n)); !os.IsNotExist(err) {
					t.Fatal("created", n)
				}
			}
		})
	}
	if Supported("paper") || FileName("paper") != "" || !Supported("VeLoCiTy") {
		t.Fatal("supported")
	}
}
func TestSavePreservesUnknown(t *testing.T) {
	for _, kind := range []string{"velocity", "bungeecord", "waterfall"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			m := starter(kind, 25590)
			apply(kind, m, usable(kind))
			m["plugin"] = object{"nested": object{"token": "do-not-expose", "enabled": true}, "list": []any{"a", 17}}
			if kind == "velocity" {
				m["bind"] = "[::]:25590"
				m["forwarding-secret-file"] = "keys/custom.secret"
				m["advanced"].(object)["custom"] = 99
			} else {
				l, _ := listener(m, true)
				l["host"] = "127.0.0.1:25590"
				l["custom"] = object{"keep": true}
				m["servers"].(object)["lobby"].(object)["plugin"] = object{"keep": "yes"}
			}
			writeDoc(t, dir, kind, m)
			before := doc(t, dir, kind)
			s := read(t, dir, kind)
			encoded, _ := json.Marshal(s)
			if bytes.Contains(encoded, []byte("do-not-expose")) {
				t.Fatal("unknown secret exposed")
			}
			s.Config.MOTD = "Quotes \" and newline\n[a]\nUnicode ☃"
			s.Config.MaxPlayers = 123
			s.Config.PingPassthrough = "ALL"
			saved, err := Save(dir, kind, 25601, s.Revision, s.Config)
			if err != nil {
				t.Fatal(err)
			}
			after := doc(t, dir, kind)
			if !reflect.DeepEqual(before["plugin"], after["plugin"]) {
				t.Fatal("plugin lost")
			}
			if saved.Config.MOTD != s.Config.MOTD || saved.Revision == s.Revision || !reflect.DeepEqual(saved, read(t, dir, kind)) {
				t.Fatal("snapshot mismatch")
			}
			if kind == "velocity" {
				if after["bind"] != "[::]:25601" || after["forwarding-secret-file"] != "keys/custom.secret" || after["advanced"].(object)["custom"] != int64(99) {
					t.Fatal(after)
				}
				if secret, err := ReadSecret(dir, kind); err != nil || len(secret) != 64 {
					t.Fatal(err)
				}
				if _, err = os.Stat(filepath.Join(dir, "forwarding.secret")); !os.IsNotExist(err) {
					t.Fatal("default secret invented")
				}
			} else {
				l, _ := listener(after, true)
				if l["host"] != "127.0.0.1:25601" || l["query_port"] != 25601 || !reflect.DeepEqual(l["custom"], object{"keep": true}) {
					t.Fatal(l)
				}
				if !reflect.DeepEqual(after["servers"].(object)["lobby"].(object)["plugin"], object{"keep": "yes"}) {
					t.Fatal("backend extra lost")
				}
			}
			b := get(t, dir, FileName(kind))
			if _, err = Save(dir, kind, 25601, s.Revision, s.Config); !errors.Is(err, ErrConflict) {
				t.Fatal("stale", err)
			}
			if !bytes.Equal(b, get(t, dir, FileName(kind))) {
				t.Fatal("stale save wrote")
			}
			if _, err = Save(dir, kind, 25601, "", s.Config); err == nil {
				t.Fatal("empty revision")
			}
			// Removing a backend is explicit and leaves retained backend extras intact.
			s = saved
			s.Config.Backends = append(s.Config.Backends, Backend{Name: "second", Address: "backend.example:25567"})
			s, err = Save(dir, kind, 25601, s.Revision, s.Config)
			if err != nil {
				t.Fatal(err)
			}
			s.Config.Backends = s.Config.Backends[:1]
			if _, err = Save(dir, kind, 25601, s.Revision, s.Config); err != nil {
				t.Fatal(err)
			}
			if _, ok := doc(t, dir, kind)["servers"].(object)["second"]; ok {
				t.Fatal("backend not removed")
			}
		})
	}
}
func TestValidation(t *testing.T) {
	cases := map[string]func(*Settings){
		"collision":        func(s *Settings) { s.Backends = append(s.Backends, Backend{Name: "LOBBY", Address: "host:25565"}) },
		"reserved":         func(s *Settings) { s.Backends[0].Name = "TrY" },
		"unsafe name":      func(s *Settings) { s.Backends[0].Name = "bad.name" },
		"long name":        func(s *Settings) { s.Backends[0].Name = strings.Repeat("a", 65) },
		"unknown fallback": func(s *Settings) { s.Try = []string{"other"} },
		"unknown route":    func(s *Settings) { s.ForcedHosts[0].Servers = []string{"other"} },
		"empty route":      func(s *Settings) { s.ForcedHosts[0].Servers = nil },
		"host injection":   func(s *Settings) { s.ForcedHosts[0].Host = "host\nkey = value" },
		"route collision": func(s *Settings) {
			s.ForcedHosts = append(s.ForcedHosts, ForcedHost{Host: "PLAY.EXAMPLE.COM", Servers: []string{"lobby"}})
		},
		"motd limit":     func(s *Settings) { s.MOTD = strings.Repeat("a", 4097) },
		"motd utf8":      func(s *Settings) { s.MOTD = string([]byte{255}) },
		"negative slots": func(s *Settings) { s.MaxPlayers = -1 },
		"empty backends": func(s *Settings) { s.Backends = nil },
		"empty fallback": func(s *Settings) { s.Try = nil },
		"refs limit":     func(s *Settings) { s.Try = make([]string, 129) },
		"routes limit":   func(s *Settings) { s.ForcedHosts = make([]ForcedHost, 129) },
		"backend limit":  func(s *Settings) { s.Backends = make([]Backend, 129) },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			s := usable("velocity")
			mutate(&s)
			if err := Validate("velocity", s); err == nil {
				t.Fatal("accepted")
			}
		})
	}
	for _, a := range []string{"http://host:25565", "host:0", "host:65536", "host:25565/path", "host\n:25565", "host:+25565", "host:", "::1:25565", "bad host:25565", "[fe80::1%eth0]:25565"} {
		s := usable("velocity")
		s.Backends[0].Address = a
		if Validate("velocity", s) == nil {
			t.Fatal("bad address", a)
		}
	}
	for _, a := range []string{"localhost:1", "127.0.0.1:65535", "[2001:db8::1]:25565", "mc-1.example.com:25565"} {
		s := usable("velocity")
		s.Backends[0].Address = a
		if err := Validate("velocity", s); err != nil {
			t.Fatal(a, err)
		}
	}
	for _, mode := range []string{"NONE", "LEGACY", "MODERN", "BUNGEEGUARD"} {
		s := usable("velocity")
		s.ForwardingMode = mode
		if err := Validate("VeLoCiTy", s); err != nil {
			t.Fatal(err)
		}
	}
	s := usable("bungeecord")
	s.ForcedHosts[0].Servers = []string{"lobby", "lobby"}
	if Validate("bungeecord", s) == nil {
		t.Fatal("multi-backend Bungee route accepted")
	}
	s = usable("velocity")
	s.ForwardingMode = "modern"
	if Validate("velocity", s) == nil {
		t.Fatal("lowercase DTO mode accepted")
	}
}
func TestInvalidFilesUntouchedAndRedacted(t *testing.T) {
	fixtures := []struct{ kind, text string }{
		{"velocity", "secret = \"super-secret\n"},
		{"velocity", "online-mode = 'super-secret'\n"},
		{"velocity", "show-max-players = 1.5\n"},
		{"velocity", "[servers]\ntry = [123]\n"},
		{"bungeecord", "listeners: [super-secret"},
		{"bungeecord", "listeners: []\n"},
		{"bungeecord", "host: 0.0.0.0:25565\n"},
		{"bungeecord", "listeners: [{host: '0.0.0.0:25565', max_players: 'super-secret'}]\n"},
		{"bungeecord", "listeners: [{host: '0.0.0.0:25565'}]\nx: &value {secret: super-secret}\ny: *value\n"},
		{"bungeecord", "listeners: [{host: '0.0.0.0:25565'}]\nx: {<<: {secret: super-secret}}\n"},
		{"bungeecord", "listeners: [{host: '0.0.0.0:25565'}]\nx: 1\nx: 2\n"},
		{"bungeecord", "listeners: [{host: '0.0.0.0:25565'}]\n---\nsecret: super-secret\n"},
		{"bungeecord", "listeners: [{host: '0.0.0.0:25565'}]\nx: !custom super-secret\n"},
	}
	for _, f := range fixtures {
		t.Run(f.kind+f.text[:8], func(t *testing.T) {
			dir := t.TempDir()
			raw := []byte(f.text)
			put(t, dir, FileName(f.kind), raw)
			if _, err := Read(dir, f.kind, 25565); err == nil || strings.Contains(err.Error(), "super-secret") {
				t.Fatal("unsafe error", err)
			}
			if _, err := Save(dir, f.kind, 25565, revision(raw), usable(f.kind)); err == nil || strings.Contains(err.Error(), "super-secret") {
				t.Fatal("unsafe save", err)
			}
			if created, err := Ensure(dir, f.kind, 25565); err != nil || created {
				t.Fatal("ensure repaired", err)
			}
			if !bytes.Equal(raw, get(t, dir, FileName(f.kind))) {
				t.Fatal("invalid file rewritten")
			}
		})
	}
}
func TestSyncPort(t *testing.T) {
	dir := t.TempDir()
	put(t, dir, "velocity.toml", []byte("bind = '[::]:25565'\n[plugin]\nbind = '127.0.0.1:1000'\n[query]\nenabled = false\nport = 1234\n"))
	if err := SyncPort(dir, "velocity", 25600); err != nil {
		t.Fatal(err)
	}
	m := doc(t, dir, "velocity")
	if m["bind"] != "[::]:25600" || m["plugin"].(object)["bind"] != "127.0.0.1:1000" || m["query"].(object)["port"] != int64(25600) || m["query"].(object)["enabled"] != false {
		t.Fatal(m)
	}
	put(t, dir, "velocity.toml", []byte("bind = '0.0.0.0:25565'\n"))
	if err := SyncPort(dir, "velocity", 25601); err != nil {
		t.Fatal(err)
	}
	put(t, dir, "config.yml", []byte("host: unrelated\nlisteners:\n  - host: '[::]:25565'\n    query_enabled: true\n    query_port: 25565\n    plugin: keep\n  - host: '127.0.0.1:25570'\n    query_port: 25571\n    motd: secondary\nservers: {}\n"))
	before := doc(t, dir, "bungeecord")["listeners"].([]any)[1]
	if _, err := Read(dir, "bungeecord", 25600); err == nil || err.Error() != "Multiple listeners: edit config.yml in Files" {
		t.Fatal(err)
	}
	b := get(t, dir, "config.yml")
	if _, err := Save(dir, "bungeecord", 25600, revision(b), usable("bungeecord")); err == nil {
		t.Fatal("multi-listener Save")
	}
	if !bytes.Equal(b, get(t, dir, "config.yml")) {
		t.Fatal("multi-listener Save wrote")
	}
	if err := SyncPort(dir, "bungeecord", 25600); err != nil {
		t.Fatal(err)
	}
	m = doc(t, dir, "bungeecord")
	a := m["listeners"].([]any)
	l := a[0].(object)
	if m["host"] != "unrelated" || !reflect.DeepEqual(before, a[1]) || l["host"] != "[::]:25600" || l["query_port"] != 25600 || l["query_enabled"] != true || l["plugin"] != "keep" {
		t.Fatal(m)
	}
	for _, kind := range []string{"velocity", "bungeecord"} {
		b := get(t, dir, FileName(kind))
		if SyncPort(dir, kind, 0) == nil {
			t.Fatal("invalid port")
		}
		if !bytes.Equal(b, get(t, dir, FileName(kind))) {
			t.Fatal("invalid port wrote")
		}
	}
}
func TestSecrets(t *testing.T) {
	first := ""
	for i := 0; i < 2; i++ {
		dir := t.TempDir()
		if _, err := Ensure(dir, "velocity", 25565); err != nil {
			t.Fatal(err)
		}
		secret, err := ReadSecret(dir, "velocity")
		if err != nil {
			t.Fatal(err)
		}
		if i == 1 && first == secret {
			t.Fatal("reused secret")
		}
		first = secret
	}
	for _, value := range []string{"", "\r\n", strings.Repeat("x", 4097), "hello\x00world", "hello\nworld", string([]byte{255})} {
		t.Run("invalid secret", func(t *testing.T) {
			dir := t.TempDir()
			writeDoc(t, dir, "velocity", starter("velocity", 25565))
			put(t, dir, "forwarding.secret", []byte(value))
			before := get(t, dir, "velocity.toml")
			if _, err := ReadSecret(dir, "velocity"); err == nil {
				t.Fatal("invalid secret read")
			}
			if _, err := Ensure(dir, "velocity", 25565); err == nil {
				t.Fatal("invalid secret accepted")
			}
			if _, err := Save(dir, "velocity", 25565, revision(before), usable("velocity")); err == nil {
				t.Fatal("invalid secret saved")
			}
			if !bytes.Equal(before, get(t, dir, "velocity.toml")) || string(get(t, dir, "forwarding.secret")) != value {
				t.Fatal("invalid secret/config overwritten")
			}
		})
	}
	for _, name := range []string{"../outside", "/absolute", "https://example/secret", "C:/secret", "keys/../secret", "keys\\secret", "keys//secret", "keys/./secret", "keys/secret.", strings.Repeat("a", 1025)} {
		t.Run("path", func(t *testing.T) {
			dir := t.TempDir()
			m := starter("velocity", 25565)
			m["forwarding-secret-file"] = name
			writeDoc(t, dir, "velocity", m)
			if _, err := ReadSecret(dir, "velocity"); err == nil {
				t.Fatal("unsafe path read", name)
			}
			if _, err := Read(dir, "velocity", 25565); err == nil {
				t.Fatal("unsafe path builder", name)
			}
		})
	}
	dir := t.TempDir()
	m := starter("velocity", 25565)
	m["forwarding-secret-file"] = "custom.secret"
	writeDoc(t, dir, "velocity", m)
	put(t, dir, "custom.secret", []byte("existing-secret\r\n"))
	before := get(t, dir, "velocity.toml")
	if _, err := Ensure(dir, "velocity", 25565); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, get(t, dir, "velocity.toml")) {
		t.Fatal("ensure changed custom config")
	}
	if _, err := Save(dir, "velocity", 25565, revision(before), usable("velocity")); err != nil {
		t.Fatal(err)
	}
	if string(get(t, dir, "custom.secret")) != "existing-secret\r\n" {
		t.Fatal("secret rotated")
	}
	if s, err := ReadSecret(dir, "velocity"); err != nil || s != "existing-secret" {
		t.Fatal("line ending", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "forwarding.secret")); !os.IsNotExist(err) {
		t.Fatal("default secret created")
	}
}
func TestFileGuards(t *testing.T) {
	for _, kind := range []string{"velocity", "bungeecord"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			name := FileName(kind)
			put(t, dir, name, bytes.Repeat([]byte("x"), maxConfig+1))
			if _, err := Read(dir, kind, 25565); err == nil {
				t.Fatal("oversize read")
			}
			if _, err := Ensure(dir, kind, 25565); err == nil {
				t.Fatal("oversize ensure")
			}
			if SyncPort(dir, kind, 25565) == nil {
				t.Fatal("oversize sync")
			}
			if err := os.Remove(filepath.Join(dir, name)); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(filepath.Join(dir, name), 0700); err != nil {
				t.Fatal(err)
			}
			if _, err := Read(dir, kind, 25565); err == nil {
				t.Fatal("directory read")
			}
			if _, err := Ensure(dir, kind, 25565); err == nil {
				t.Fatal("directory ensure")
			}
		})
	}
	t.Run("symlink", func(t *testing.T) {
		dir := t.TempDir()
		outside := filepath.Join(t.TempDir(), "outside")
		if err := os.WriteFile(outside, []byte("bind = '0.0.0.0:25565'\n"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(dir, "velocity.toml")); err != nil {
			t.Skip("symlink unavailable:", err)
		}
		if _, err := Read(dir, "velocity", 25565); err == nil {
			t.Fatal("symlink read")
		}
		if _, err := Ensure(dir, "velocity", 25565); err == nil {
			t.Fatal("symlink ensure")
		}
		if SyncPort(dir, "velocity", 25600) == nil {
			t.Fatal("symlink sync")
		}
	})
	t.Run("directory symlink", func(t *testing.T) {
		dir := t.TempDir()
		outside := t.TempDir()
		link := filepath.Join(dir, "linked")
		if err := os.Symlink(outside, link); err != nil {
			t.Skip("symlink unavailable:", err)
		}
		if _, err := Ensure(filepath.Join(link, "new"), "velocity", 25565); err == nil {
			t.Fatal("ancestor symlink accepted")
		}
	})
	t.Run("hardlink", func(t *testing.T) {
		dir := t.TempDir()
		outside := filepath.Join(t.TempDir(), "outside")
		raw := []byte("bind = '0.0.0.0:25565'\n")
		if err := os.WriteFile(outside, raw, 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Link(outside, filepath.Join(dir, "velocity.toml")); err != nil {
			t.Skip("hardlink unavailable:", err)
		}
		if err := SyncPort(dir, "velocity", 25600); err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(outside)
		if err != nil || !bytes.Equal(raw, b) {
			t.Fatal("hardlink target changed", err)
		}
	})
}
func TestPublishConflictIdentityAndRevision(t *testing.T) {
	dir := t.TempDir()
	put(t, dir, "velocity.toml", []byte("bind = '0.0.0.0:25565'\n"))
	root, err := rooted.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	old, st, err := readFile(root, "velocity.toml", maxConfig)
	if err != nil {
		t.Fatal(err)
	}
	put(t, dir, "velocity.toml", []byte("bind = '0.0.0.0:25566'\n"))
	if err = publish(root, "velocity.toml", old, st, old); !errors.Is(err, ErrConflict) {
		t.Fatal("revision", err)
	}
	if err = os.Rename(filepath.Join(dir, "velocity.toml"), filepath.Join(dir, "retained-inode")); err != nil {
		t.Fatal(err)
	}
	put(t, dir, "velocity.toml", old)
	// Different inode with identical bytes must also conflict (not merely hash).
	if err = publish(root, "velocity.toml", old, st, old); !errors.Is(err, ErrConflict) {
		t.Fatal("identity", err)
	}
}
