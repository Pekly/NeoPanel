package runner

import (
	"github.com/Pekly/NeoPanel/internal/database"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/downloader"
)

func TestSteelNativeLifecycle(t *testing.T) {
	p := lifecycleInstance(t, "steel")
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	binary, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	p.server.Type = "steel"
	p.server.JarFile = downloader.ServerFileName("steel")
	p.server.JavaPath = "nonexistent-java"
	p.server.JavaArgs = "-XX:InvalidJavaFlag"
	p.server.MemoryMinMB, p.server.MemoryMaxMB = 4096, 1024 // JVM heap constraints do not apply.
	p.server.Port = 25579
	if err := os.WriteFile(filepath.Join(p.workingDir, p.server.JarFile), binary, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(p.workingDir, "config"), 0755); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(p.workingDir, "config", "config.toml")
	if err := os.WriteFile(config, []byte("[server]\nserver_port = 25565\n"), 0644); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := p.Start(); err != nil {
			t.Fatal(err)
		}
		waitProcessStatus(t, p, "online")
		if err := p.Stop(true); err != nil {
			t.Fatal(err)
		}
		waitProcessStatus(t, p, "offline")
	}
	got, _ := os.ReadFile(config)
	if !strings.Contains(string(got), "server_port = 25579") {
		t.Fatal(string(got))
	}
	if _, err := os.Stat(filepath.Join(p.workingDir, "server.properties")); !os.IsNotExist(err) {
		t.Fatal("created Java properties")
	}
}

func TestSyncSteelPort(t *testing.T) {
	dir := t.TempDir()
	if err := SyncSteelPort(dir, 25565); err == nil {
		t.Fatal("missing config accepted")
	}
	if err := os.MkdirAll(filepath.Join(dir, "config"), 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config", "config.toml")
	original := "# preserved\n[server] # server settings\nserver_port=25565 # managed\nmotd = \"hello\"\n[other]\nserver_port = 1234\n"
	if err := os.WriteFile(path, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}
	if err := SyncSteelPort(dir, 25580); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	want := strings.Replace(original, "server_port=25565 # managed", "server_port = 25580", 1)
	if string(got) != want {
		t.Fatalf("got %s", got)
	}
	if err := SyncSteelPort(dir, 70000); err == nil {
		t.Fatal("invalid port accepted")
	}
	if err := os.WriteFile(path, []byte("[other]\nserver_port = 10\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := SyncSteelPort(dir, 25565); err == nil {
		t.Fatal("wrong TOML table accepted")
	}
}

func TestSteelExecutableAndReadiness(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"../escape", "..", "/bin/sh", `C:\\windows\\cmd.exe`, "missing"} {
		if _, err := steelExecutable(dir, name); err == nil {
			t.Fatalf("accepted %q", name)
		}
	}
	name := downloader.ServerFileName("steel")
	if err := os.WriteFile(filepath.Join(dir, name), []byte("binary"), 0600); err != nil {
		t.Fatal(err)
	}
	path, err := steelExecutable(dir, name)
	if err != nil || !filepath.IsAbs(path) {
		t.Fatal(path, err)
	}
	info, _ := os.Stat(path)
	if runtime.GOOS != "windows" && info.Mode().Perm()&0100 == 0 {
		t.Fatal("upload not executable")
	}
	if !serverReadyLine("steel", "[INFO] Started Steel Server") || serverReadyLine("steel", "Starting Steel Server") || serverReadyLine("steel", "Done (1s)") || !serverReadyLine("paper", "Done (1s)") {
		t.Fatal("incorrect readiness detection")
	}
}

func TestSteelPlayerLogs(t *testing.T) {
	for _, prefix := range []string{"INFO ", "12:45:01:123 INFO ", "1.23s INFO steel_core::server::broadcasting ", "\x1b[32mINFO\x1b[0m ", "\r\x1b[J> \r\x1b[J 0.94s \x1b[0;1;34m[Info]\x1b[0m "} {
		p := NewProcessInstance(&database.Server{ID: "steel", Type: "steel"}, t.TempDir(), time.Second, nil, nil)
		p.status, p.runID, p.zeroPlayersSince = "online", 1, time.Now().Add(-time.Hour)
		p.trackPlayerLine(prefix+"steelman joined the game", 1)
		if p.GetPlayerCount() != 1 {
			t.Fatalf("missed join: %q", prefix)
		}
		if _, idle := p.GetIdleDuration(); idle {
			t.Fatal("occupied Steel server considered idle")
		}
		p.trackPlayerLine(prefix+"steelman left the game", 0)
		if p.GetPlayerCount() != 1 {
			t.Fatal("stale run changed players")
		}
		p.trackPlayerLine(prefix+"steelman left the game", 1)
		if p.GetPlayerCount() != 0 {
			t.Fatalf("missed leave: %q", prefix)
		}
		if _, idle := p.GetIdleDuration(); !idle {
			t.Fatal("empty Steel server not considered idle")
		}
	}
}
