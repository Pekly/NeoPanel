package runner

import (
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/downloader"
)

func TestSteelSandboxCommand(t *testing.T) {
	binDir := t.TempDir()
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	for _, name := range []string{"bwrap", "systemd-run"} {
		if err := os.WriteFile(filepath.Join(binDir, name+suffix), []byte("stub"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", binDir)
	dir := t.TempDir()
	binary := filepath.Join(dir, downloader.ServerFileName("steel"))
	if err := os.WriteFile(binary, []byte("native"), 0755); err != nil {
		t.Fatal(err)
	}
	p := NewProcessInstance(&database.Server{ID: "steel", Type: "steel", MemoryMaxMB: 2048}, dir, time.Second, nil, nil)
	p.SetIsolationConfig("bwrap", 512)
	cmd, label, err := p.buildCommand(binary, nil)
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(cmd.Args, " ")
	if !strings.Contains(args, "--property=MemoryMax=2048M") {
		t.Fatal("missing native memory limit", args)
	}
	if !strings.Contains(label, "MemoryMax=2048 MiB enforced by systemd") {
		t.Fatal("missing active limit in isolation message", label)
	}
	for i, arg := range cmd.Args {
		if arg == "--ro-bind-try" && i+1 < len(cmd.Args) && (cmd.Args[i+1] == dir || cmd.Args[i+1] == filepath.Dir(dir)) {
			t.Fatal("native launch exposes Java-style runtime directory", args)
		}
	}
	if cmd.Args[len(cmd.Args)-1] != binary {
		t.Fatal("unexpected native arguments", args)
	}
	p.SetIsolationConfig("systemd", 512)
	systemdCmd, systemdLabel, err := p.buildCommand(binary, nil)
	if err != nil || !strings.Contains(strings.Join(systemdCmd.Args, " "), "--property=MemoryMax=2048M") || !strings.Contains(systemdLabel, "MemoryMax=2048 MiB enforced by systemd") {
		t.Fatal("systemd-only limit or description missing", systemdLabel, err)
	}
	p.SetIsolationConfig("process", 512)
	_, processLabel, err := p.buildCommand(binary, nil)
	if err != nil || strings.Contains(processLabel, "MemoryMax") {
		t.Fatal("unrestricted mode must not claim a memory cap", processLabel, err)
	}
}

// Opt in explicitly: downloads and executes an official Steel release on a spare port.
func TestSteelLiveRelease(t *testing.T) {
	if os.Getenv("NEOPANEL_STEEL_LIVE_TEST") != "1" {
		t.Skip("set NEOPANEL_STEEL_LIVE_TEST=1 for the official-release smoke test")
	}
	dir := t.TempDir()
	name := downloader.ServerFileName("steel")
	if err := downloader.DownloadServerJar("steel", "latest", dir, name); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	p := NewProcessInstance(&database.Server{ID: "steel-live", Type: "steel", JarFile: name, JavaPath: "missing-java", Port: port}, dir, 60*time.Second, nil, nil)
	p.SetIsolationConfig("process", 512)
	t.Cleanup(func() {
		if err := p.Stop(false); err != nil {
			t.Error(err)
		}
	})
	if err := p.Start(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(60 * time.Second)
	for p.GetStatus() == "starting" && time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
	}
	p.mu.RLock()
	logs := strings.Join(p.logBuffer, "\n")
	p.mu.RUnlock()
	t.Log(logs)
	for _, code := range terminalControlRe.FindAllString(logs, -1) {
		if !terminalSGRRe.MatchString(code) {
			t.Fatalf("raw terminal control leaked to logs: %q", code)
		}
	}
	if !strings.Contains(logs, "\x1b[") {
		t.Fatal("Steel colors were discarded")
	}
	for _, line := range strings.Split(logs, "\n") {
		visible := strings.TrimSpace(ansiEscapeRe.ReplaceAllString(line, ""))
		if visible == "" || visible == ">" {
			t.Fatalf("empty prompt leaked to logs: %q", line)
		}
	}
	if p.GetStatus() != "online" {
		t.Fatalf("Steel did not become online: %s", p.GetStatus())
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)), time.Second)
	if err != nil {
		t.Fatal("assigned port is not listening", err)
	}
	conn.Close()
	if err := p.Stop(true); err != nil {
		t.Fatal(err)
	}
	if p.GetStatus() != "offline" {
		t.Fatal("Steel did not stop")
	}
}
