package runner

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

// Execute this test binary as a Java stand-in, through the real launch path.
func init() {
	mode := os.Getenv("NEOPANEL_LIFECYCLE_HELPER")
	if mode == "" {
		return
	}
	if mode == "timeout" {
		time.Sleep(time.Minute)
		os.Exit(0)
	}
	if mode == "listener" {
		listener, err := net.Listen("tcp4", "127.0.0.1:0")
		if err != nil {
			os.Exit(4)
		}
		defer listener.Close()
		if err := os.WriteFile("listener.addr", []byte(listener.Addr().String()), 0600); err != nil {
			os.Exit(5)
		}
	}
	if mode == "steel" {
		if len(os.Args) != 1 {
			fmt.Println("Unexpected Java arguments:", os.Args[1:])
			os.Exit(2)
		}
		fmt.Println("Started Steel Server")
	} else {
		fmt.Println(`Done (0.01s)! For help, type "help"`)
	}
	if mode == "tail" {
		fmt.Println("FINAL-OUTPUT-PRESERVED")
		os.Exit(0)
	}
	if mode == "ignore" {
		time.Sleep(time.Minute)
		os.Exit(0)
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "stop" {
			time.Sleep(150 * time.Millisecond)
			fmt.Println("FINAL-OUTPUT-PRESERVED")
			os.Exit(0)
		}
	}
	os.Exit(0)
}

func lifecycleInstance(t *testing.T, mode string) *ProcessInstance {
	t.Helper()
	t.Setenv("NEOPANEL_LIFECYCLE_HELPER", mode)
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "server.jar"), []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}
	p := NewProcessInstance(&database.Server{ID: "lifecycle", Type: "paper", JarFile: "server.jar", JavaPath: exe}, dir, time.Second, nil, nil)
	p.SetIsolationConfig("process", 512)
	t.Cleanup(func() {
		if err := p.Stop(false); err != nil {
			t.Error(err)
		}
	})
	return p
}
func waitProcessStatus(t *testing.T, p *ProcessInstance, want string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if p.GetStatus() == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("status=%s want=%s logs=%v", p.GetStatus(), want, p.logBuffer)
}

func TestLifecycleStopWaitsAndRestartDoesNotOverlap(t *testing.T) {
	p := lifecycleInstance(t, "graceful")
	for i := 0; i < 3; i++ {
		if err := p.Start(); err != nil {
			t.Fatal(err)
		}
		waitProcessStatus(t, p, "online")
		if err := p.Start(); err == nil {
			t.Fatal("duplicate start accepted")
		}
		started := time.Now()
		if err := p.Stop(true); err != nil {
			t.Fatal(err)
		}
		if time.Since(started) < 100*time.Millisecond || p.GetStatus() != "offline" {
			t.Fatal("stop returned before process exited")
		}
		p.mu.RLock()
		logs := strings.Join(p.logBuffer, "\n")
		p.mu.RUnlock()
		if !strings.Contains(logs, "FINAL-OUTPUT-PRESERVED") {
			t.Fatal("lost final output")
		}
	}
}
func TestLifecycleForceKillAndTimeoutReleaseOwnership(t *testing.T) {
	for _, mode := range []string{"ignore", "timeout", "tail"} {
		t.Run(mode, func(t *testing.T) {
			p := lifecycleInstance(t, mode)
			p.SetStartupTimeout(200 * time.Millisecond)
			if err := p.Start(); err != nil {
				t.Fatal(err)
			}
			if mode == "ignore" {
				waitProcessStatus(t, p, "online")
				if err := p.Stop(false); err != nil {
					t.Fatal(err)
				}
			}
			waitProcessStatus(t, p, "offline")
			lock, err := acquireRunLock(p.workingDir)
			if err != nil {
				t.Fatal("ownership leaked", err)
			}
			lock.Close()
		})
	}
}
func TestLifecycleRejectsStoppingAndDuplicateDirectory(t *testing.T) {
	p := lifecycleInstance(t, "graceful")
	p.status = "stopping"
	if err := p.Start(); err == nil {
		t.Fatal("started during stopping")
	}
	p.status = "offline"
	lock, err := acquireRunLock(p.workingDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Start(); err == nil {
		t.Fatal("duplicate directory ownership allowed")
	}
	lock.Close()
	if p.GetStatus() != "offline" {
		t.Fatal(p.GetStatus())
	}
	if err := p.Start(); err != nil {
		t.Fatal(err)
	}
	if err := p.Stop(false); err != nil {
		t.Fatal(err)
	}
}
func TestLifecycleFailedStartReleasesLock(t *testing.T) {
	p := lifecycleInstance(t, "graceful")
	p.server.JavaPath = filepath.Join(p.workingDir, "missing-java")
	if err := p.Start(); err == nil {
		t.Fatal("missing executable accepted")
	}
	lock, err := acquireRunLock(p.workingDir)
	if err != nil {
		t.Fatal(err)
	}
	lock.Close()
	if p.GetStatus() != "offline" {
		t.Fatal(p.GetStatus())
	}
}
func TestExplicitIsolationNeverSilentlyDowngrades(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	p := NewProcessInstance(&database.Server{ID: "test"}, t.TempDir(), time.Second, nil, nil)
	for _, engine := range []string{"bwrap", "systemd", "typo"} {
		p.SetIsolationConfig(engine, 512)
		if cmd, _, err := p.buildCommand("java", nil); err == nil || cmd != nil {
			t.Fatalf("%s silently downgraded", engine)
		}
	}
	p.SetIsolationConfig("auto", 512)
	_, label, err := p.buildCommand("java", nil)
	if err != nil || !strings.Contains(label, "NO filesystem isolation") {
		t.Fatal(label, err)
	}
}
func TestSandboxAndScopeCommandConstruction(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	for _, name := range []string{"bwrap", "systemd-run"} {
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte("stub"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	p := NewProcessInstance(&database.Server{ID: "test", CPULimitPercent: 150}, t.TempDir(), time.Second, nil, nil)
	p.SetIsolationConfig("bwrap", 128)
	cmd, label, err := p.buildCommand("/usr/bin/java", []string{"-jar", "server.jar"})
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(cmd.Args, " ")
	for _, want := range []string{"--clearenv", "--cap-drop ALL", "--unit=neopanel-", "--property=TasksMax=128", "--property=CPUQuota=150%"} {
		if !strings.Contains(args, want) {
			t.Fatalf("missing %s: %s", want, args)
		}
	}
	if strings.Contains(args, "--ro-bind-try /etc /etc") || strings.Contains(args, "--ro-bind-try /opt /opt") {
		t.Fatal("host config exposed")
	}
	if !strings.Contains(label, "Bubblewrap") {
		t.Fatal(label)
	}
	// A finalized run ignores late watchdogs instead of signalling a reused PID.
	c := &ownedCommand{cmd: exec.Command("unused"), reaped: true}
	if err := c.stop(false); err != nil {
		t.Fatal(err)
	}
}
