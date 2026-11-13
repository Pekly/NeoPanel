package api

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPluginTogglePersistsAndPreservesContents(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "example.jar"), []byte("plugin data"), 0644); err != nil {
		t.Fatal(err)
	}
	disabled, err := setPluginEnabled(dir, "example.jar", false)
	if err != nil || disabled != "example.jar.disabled" {
		t.Fatalf("disable: %s %v", disabled, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "example.jar")); !os.IsNotExist(err) {
		t.Fatal("enabled JAR still exists")
	}
	data, err := os.ReadFile(filepath.Join(dir, disabled))
	if err != nil || string(data) != "plugin data" {
		t.Fatal("contents lost")
	}
	if same, err := setPluginEnabled(dir, disabled, false); err != nil || same != disabled {
		t.Fatal("idempotent toggle failed")
	}
	enabled, err := setPluginEnabled(dir, disabled, true)
	if err != nil || enabled != "example.jar" {
		t.Fatalf("enable: %s %v", enabled, err)
	}
	if err := os.WriteFile(filepath.Join(dir, disabled), []byte("existing disabled"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := setPluginEnabled(dir, enabled, false); err == nil {
		t.Fatal("overwrote existing destination")
	}
	data, _ = os.ReadFile(filepath.Join(dir, disabled))
	if string(data) != "existing disabled" {
		t.Fatal("collision destroyed destination")
	}
	for _, name := range []string{"../example.jar", "config.yml", "a.jar/..", "C:escape.jar", ""} {
		if _, err := setPluginEnabled(dir, name, false); err == nil {
			t.Fatalf("accepted unsafe filename %q", name)
		}
	}
}
