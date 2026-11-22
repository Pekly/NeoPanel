package proxyconfig

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRejectedSaveHasNoSideEffects(t *testing.T) {
	dir := t.TempDir()
	writeDoc(t, dir, "velocity", starter("velocity", 25565))
	raw := get(t, dir, "velocity.toml")
	if _, err := Save(dir, "velocity", 25565, "stale", usable("velocity")); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	s := usable("velocity")
	s.Try = []string{"missing"}
	if _, err := Save(dir, "velocity", 25565, revision(raw), s); err == nil {
		t.Fatal("invalid settings accepted")
	}
	if !bytes.Equal(raw, get(t, dir, "velocity.toml")) {
		t.Fatal("rejected save changed config")
	}
	if _, err := os.Stat(filepath.Join(dir, "forwarding.secret")); !os.IsNotExist(err) {
		t.Fatal("rejected save created secret")
	}
}
func TestSecretCreationOnlyWhenRequired(t *testing.T) {
	for _, mode := range []string{"NONE", "LEGACY", "MODERN", "BUNGEEGUARD"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			m := starter("velocity", 25565)
			m["forwarding-secret-file"] = "custom.secret"
			writeDoc(t, dir, "velocity", m)
			if _, err := Ensure(dir, "velocity", 25565); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(dir, "custom.secret")); !os.IsNotExist(err) {
				t.Fatal("Ensure created secret for existing config")
			}
			s := usable("velocity")
			s.ForwardingMode = mode
			if _, err := Save(dir, "velocity", 25565, read(t, dir, "velocity").Revision, s); err != nil {
				t.Fatal(err)
			}
			_, err := os.Stat(filepath.Join(dir, "custom.secret"))
			required := mode == "MODERN" || mode == "BUNGEEGUARD"
			if (err == nil) != required {
				t.Fatal("unexpected secret creation", err)
			}
			if _, err := os.Stat(filepath.Join(dir, "forwarding.secret")); !os.IsNotExist(err) {
				t.Fatal("default secret created")
			}
		})
	}
}
func TestSecretLinks(t *testing.T) {
	for _, hard := range []bool{false, true} {
		t.Run(map[bool]string{false: "symlink", true: "hardlink"}[hard], func(t *testing.T) {
			dir := t.TempDir()
			writeDoc(t, dir, "velocity", starter("velocity", 25565))
			outside := filepath.Join(t.TempDir(), "secret")
			raw := []byte("do-not-rotate")
			if err := os.WriteFile(outside, raw, 0600); err != nil {
				t.Fatal(err)
			}
			link := os.Symlink
			if hard {
				link = os.Link
			}
			if err := link(outside, filepath.Join(dir, "forwarding.secret")); err != nil {
				t.Skip("link unavailable:", err)
			}
			_, err := Ensure(dir, "velocity", 25565)
			if hard && err != nil {
				t.Fatal(err)
			}
			if !hard && err == nil {
				t.Fatal("symlink accepted")
			}
			_, err = ReadSecret(dir, "velocity")
			if !hard && err == nil {
				t.Fatal("symlink secret read")
			}
			_, err = Save(dir, "velocity", 25565, read(t, dir, "velocity").Revision, usable("velocity"))
			if hard && err != nil {
				t.Fatal(err)
			}
			if !hard && err == nil {
				t.Fatal("symlink secret Save")
			}
			got, err := os.ReadFile(outside)
			if err != nil || !bytes.Equal(got, raw) {
				t.Fatal("outside secret changed", err)
			}
		})
	}
}
func TestLegacyBungeePriorities(t *testing.T) {
	dir := t.TempDir()
	m := starter("bungeecord", 25565)
	apply("bungeecord", m, usable("bungeecord"))
	l, _ := listener(m, true)
	l["priorities"] = []string{}
	l["default_server"] = "lobby"
	l["fallback_server"] = "lobby"
	writeDoc(t, dir, "bungeecord", m)
	raw := get(t, dir, "config.yml")
	s := read(t, dir, "bungeecord")
	if !reflect.DeepEqual(s.Config.Try, []string{"lobby", "lobby"}) {
		t.Fatal(s.Config.Try)
	}
	if !bytes.Equal(raw, get(t, dir, "config.yml")) {
		t.Fatal("read migrated file")
	}
	if _, err := Save(dir, "bungeecord", 25565, s.Revision, s.Config); err != nil {
		t.Fatal(err)
	}
	l, _ = listener(doc(t, dir, "bungeecord"), true)
	if _, ok := l["default_server"]; ok {
		t.Fatal("legacy priority retained")
	}
	if _, ok := l["fallback_server"]; ok {
		t.Fatal("legacy fallback retained")
	}
	if !reflect.DeepEqual(read(t, dir, "bungeecord").Config.Try, s.Config.Try) {
		t.Fatal("priority migration lost values")
	}
}
func TestSyncInvalidSyntaxUntouched(t *testing.T) {
	for _, kind := range []string{"velocity", "bungeecord"} {
		dir := t.TempDir()
		raw := []byte("[not valid\n")
		put(t, dir, FileName(kind), raw)
		if err := SyncPort(dir, kind, 25565); err == nil {
			t.Fatal("invalid sync accepted")
		}
		if !bytes.Equal(raw, get(t, dir, FileName(kind))) {
			t.Fatal("invalid config rewritten")
		}
	}
}

func TestUnchangedPortKeepsOriginalFormatting(t *testing.T) {
	for _, tc := range []struct{ kind, text string }{
		{"velocity", "# keep this comment\nbind = \"127.0.0.1:25588\" # listener\n[query]\nenabled = false\nport = 25588\n"},
		{"bungeecord", "# keep this comment\nlisteners:\n- host: 127.0.0.1:25588 # listener\n  query_enabled: false\n  query_port: 25588\n"},
		{"waterfall", "# keep this comment\nlisteners:\n- host: 127.0.0.1:25588 # listener\n  query_enabled: false\n  query_port: 25588\n"},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			dir := t.TempDir()
			file := filepath.Join(dir, FileName(tc.kind))
			if err := os.WriteFile(file, []byte(tc.text), 0600); err != nil {
				t.Fatal(err)
			}
			if err := SyncPort(dir, tc.kind, 25588); err != nil {
				t.Fatal(err)
			}
			if got := get(t, dir, FileName(tc.kind)); string(got) != tc.text {
				t.Fatal("unchanged port rewrote config", string(got))
			}
		})
	}
}
