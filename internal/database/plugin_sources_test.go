package database

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestPluginSourceDurableHashAndIdentityBinding(t *testing.T) {
	path := filepath.Join(t.TempDir(), "panel.db")
	db, err := InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	hash := strings.Repeat("a", 128)
	if err := db.ConfirmPluginSource("server", "renamed.jar.disabled", hash, "P1"); err != nil {
		t.Fatal(err)
	}
	db.Close()
	db, err = InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, tc := range []struct{ server, file, hash, want string }{
		{"server", "renamed.jar.disabled", hash, "P1"},
		{"other", "renamed.jar.disabled", hash, ""},
		{"server", "renamed.jar", hash, ""},
		{"server", "renamed.jar.disabled", strings.Repeat("b", 128), ""},
	} {
		got, err := db.ConfirmedPluginSource(tc.server, tc.file, tc.hash)
		if err != nil || got != tc.want {
			t.Fatal(tc, got, err)
		}
	}
	if pluginSourceKey("ab", "c") == pluginSourceKey("a", "bc") {
		t.Fatal("ambiguous identity key")
	}
	if err := db.ConfirmPluginSource("server", "renamed.jar.disabled", hash, "P2"); err != nil {
		t.Fatal(err)
	}
	if got, err := db.ConfirmedPluginSource("server", "renamed.jar.disabled", hash); err != nil || got != "P2" {
		t.Fatal(got, err)
	}
}
