//go:build windows

package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/Pekly/NeoPanel/internal/database"
	"golang.org/x/sys/windows"
)

func jobHelperCommand(mode, dir string) *exec.Cmd {
	exe, _ := os.Executable()
	cmd := exec.Command(exe)
	cmd.Env = append(os.Environ(), "NEOPANEL_JOB_HELPER="+mode, "NEOPANEL_JOB_DIR="+dir)
	return cmd
}

func init() {
	mode, dir := os.Getenv("NEOPANEL_JOB_HELPER"), os.Getenv("NEOPANEL_JOB_DIR")
	if mode == "" {
		return
	}
	write := func(name, value string) {
		if os.WriteFile(filepath.Join(dir, name), []byte(value), 0600) != nil {
			os.Exit(90)
		}
	}
	wait := func(name string) {
		for i := 0; i < 3000; i++ {
			if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
		os.Exit(91)
	}
	write(mode+".pid", strconv.Itoa(os.Getpid()))
	switch mode {
	case "daemon", "prelaunch":
		lock, err := acquireRunLock(dir)
		if err != nil {
			os.Exit(97)
		}
		defer lock.Close()
		if mode == "prelaunch" {
			j, err := newWindowsJob(256)
			if err != nil {
				os.Exit(92)
			}
			pid, err := windows.GetProcessId(j.parent)
			if err != nil {
				os.Exit(93)
			}
			write("surrogate.pid", strconv.Itoa(int(pid)))
		} else {
			_, err := startOwnedCommand(jobHelperCommand("root", dir), nil, 256)
			if err != nil {
				os.Exit(94)
			}
			wait("leaf.pid")
		}
		write("ready", "yes")
		time.Sleep(time.Minute)
	case "root", "child":
		next := "child"
		if mode == "child" {
			next = "leaf"
		}
		cmd := jobHelperCommand(next, dir)
		if err := cmd.Start(); err != nil {
			os.Exit(95)
		}
		_ = cmd.Process.Release()
		wait("exit-" + mode)
	case "leaf":
		time.Sleep(time.Minute)
	case "allocate":
		// The 256 MiB job ceiling must reject this 512 MiB committed allocation.
		address, err := windows.VirtualAlloc(0, 512*1024*1024, windows.MEM_COMMIT|windows.MEM_RESERVE, windows.PAGE_READWRITE)
		if err == nil {
			windows.VirtualFree(address, 0, windows.MEM_RELEASE)
			os.Exit(96)
		}
		write("allocation-denied", "yes")
	case "busy":
		memory := make([]byte, 32*1024*1024)
		for i := range memory {
			memory[i] = byte(i)
		}
		write("ready", "yes")
		end := time.Now().Add(time.Minute)
		for time.Now().Before(end) {
			for i := range memory {
				memory[i]++
			}
		}
	}
	os.Exit(0)
}

func awaitJobFile(t *testing.T, dir, name string) string {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if raw, err := os.ReadFile(filepath.Join(dir, name)); err == nil && len(raw) > 0 {
			return string(raw)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("helper did not publish", name)
	return ""
}

func pinHelper(t *testing.T, dir, name string) windows.Handle {
	t.Helper()
	pid, err := strconv.Atoi(awaitJobFile(t, dir, name+".pid"))
	if err != nil {
		t.Fatal(err)
	}
	h, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { windows.CloseHandle(h) })
	return h
}
func assertHelperExited(t *testing.T, h windows.Handle) {
	t.Helper()
	state, err := windows.WaitForSingleObject(h, 5000)
	if err != nil || state != windows.WAIT_OBJECT_0 {
		t.Fatalf("descendant survived: wait=%d err=%v", state, err)
	}
}

func TestWindowsJobTreeLifecycle(t *testing.T) {
	for _, mode := range []string{"natural-root", "natural-child", "force"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			c, err := startOwnedCommand(jobHelperCommand("root", dir), nil, 256)
			if err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { _, err := c.wait(); done <- err }()
			t.Cleanup(func() { _ = c.stop(false) })
			j := c.owner.(*windowsJob)
			child := pinHelper(t, dir, "child")
			leaf := pinHelper(t, dir, "leaf")
			if mode == "natural-child" {
				os.WriteFile(filepath.Join(dir, "exit-child"), []byte("yes"), 0600)
				assertHelperExited(t, child)
				if state, _ := windows.WaitForSingleObject(leaf, 0); state != uint32(windows.WAIT_TIMEOUT) {
					t.Fatal("leaf unexpectedly exited with its immediate parent")
				}
			}
			if mode == "force" {
				var wg sync.WaitGroup
				for i := 0; i < 8; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						if err := c.stop(false); err != nil {
							t.Error(err)
						}
					}()
				}
				wg.Wait()
			} else {
				os.WriteFile(filepath.Join(dir, "exit-root"), []byte("yes"), 0600)
			}
			select {
			case err := <-done:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(10 * time.Second):
				t.Fatal("wait blocked")
			}
			assertHelperExited(t, child)
			assertHelperExited(t, leaf)
			if err := c.stop(false); err != nil {
				t.Fatal("late stop", err)
			}
			if j.job != 0 || j.parent != 0 || j.root != 0 {
				t.Fatal("owner handles leaked")
			}
		})
	}
}

func TestWindowsJobDaemonDeath(t *testing.T) {
	for _, mode := range []string{"daemon", "prelaunch"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			cmd := jobHelperCommand(mode, dir)
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = cmd.Process.Kill() })
			awaitJobFile(t, dir, "ready")
			if lock, err := acquireRunLock(dir); err == nil {
				lock.Close()
				t.Fatal("daemon run lock was bypassed")
			}
			var handles []windows.Handle
			if mode == "prelaunch" {
				handles = append(handles, pinHelper(t, dir, "surrogate"))
			} else {
				for _, name := range []string{"root", "child", "leaf"} {
					handles = append(handles, pinHelper(t, dir, name))
				}
			}
			if err := cmd.Process.Kill(); err != nil {
				t.Fatal(err)
			}
			_ = cmd.Wait()
			for _, h := range handles {
				assertHelperExited(t, h)
			}
			lock, err := acquireRunLock(dir)
			if err != nil {
				t.Fatal("daemon-death run lock leaked", err)
			}
			lock.Close()
		})
	}
}

func TestWindowsJobLimitsAndFailedLaunch(t *testing.T) {
	j, err := newWindowsJob(256)
	if err != nil {
		t.Fatal(err)
	}
	var limits windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
	err = windows.QueryInformationJobObject(j.job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits)), nil)
	if err != nil {
		t.Fatal(err)
	}
	flags := limits.BasicLimitInformation.LimitFlags
	if limits.JobMemoryLimit != 256*1024*1024 || flags&windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE == 0 || flags&windows.JOB_OBJECT_LIMIT_JOB_MEMORY == 0 || flags&(windows.JOB_OBJECT_LIMIT_BREAKAWAY_OK|windows.JOB_OBJECT_LIMIT_SILENT_BREAKAWAY_OK) != 0 {
		t.Fatal("incorrect limits", limits)
	}
	if err := j.terminate(); err != nil {
		t.Fatal(err)
	}
	if err := j.close(); err != nil {
		t.Fatal(err)
	}
	for _, mb := range []int64{-1, 1 << 62} {
		if j, err := newWindowsJob(mb); err == nil {
			j.close()
			t.Fatal("invalid memory accepted", mb)
		}
	}
	if c, err := startOwnedCommand(exec.Command(filepath.Join(t.TempDir(), "missing.exe")), nil, 256); err == nil {
		c.stop(false)
		t.Fatal("invalid executable accepted")
	}
	dir := t.TempDir()
	c, err := startOwnedCommand(jobHelperCommand("allocate", dir), nil, 256)
	if err != nil {
		t.Fatal(err)
	}
	if waitErr, cleanupErr := c.wait(); waitErr != nil || cleanupErr != nil {
		t.Fatal(waitErr, cleanupErr)
	}
	awaitJobFile(t, dir, "allocation-denied")
}

func TestWindowsWorldSessionLock(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "custom"), 0700)
	os.WriteFile(filepath.Join(dir, "server.properties"), []byte("level-name=custom\n"), 0600)
	path := filepath.Join(dir, "custom", "session.lock")
	os.WriteFile(path, []byte("world-lock"), 0600)
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &windows.Overlapped{}); err != nil {
		t.Fatal(err)
	}
	if err := checkServerWorldLocks(dir, "paper"); err == nil {
		t.Fatal("held world lock accepted")
	}
	if err := ensureServerProcessesStopped(dir); err == nil {
		t.Fatal("maintenance ignored world lock")
	}
	f.Close()
	if err := checkServerWorldLocks(dir, "paper"); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != "world-lock" {
		t.Fatal("world lock modified")
	}
}

func TestWindowsCommandModes(t *testing.T) {
	bin := t.TempDir()
	for _, name := range []string{"bwrap.exe", "systemd-run.exe"} {
		os.WriteFile(filepath.Join(bin, name), []byte("stub"), 0700)
	}
	t.Setenv("PATH", bin)
	p := NewProcessInstance(&database.Server{Type: "paper", MemoryMaxMB: 2048, CPULimitPercent: 50}, t.TempDir(), time.Second, nil, nil)
	for _, mode := range []string{"auto", "process", "none", ""} {
		p.SetIsolationConfig(mode, 10)
		cmd, label, err := p.buildCommand("game.exe", []string{"a"})
		if err != nil || cmd.Path != "game.exe" || !strings.Contains(label, "NO filesystem isolation") || !strings.Contains(label, "CPU/task limits unsupported") || !strings.Contains(label, "3072 MiB") {
			t.Fatal(mode, cmd, label, err)
		}
	}
	for _, mode := range []string{"bwrap", "systemd", "typo"} {
		p.SetIsolationConfig(mode, 10)
		if cmd, _, err := p.buildCommand("game.exe", nil); err == nil || cmd != nil {
			t.Fatal(fmt.Sprint("Linux isolation accepted: ", mode))
		}
	}
}

func TestWindowsCleanupFailureBlocksRestart(t *testing.T) {
	p := lifecycleInstance(t, "ignore")
	if err := p.Start(); err != nil {
		t.Fatal(err)
	}
	waitProcessStatus(t, p, "online")
	c := p.owned
	j := c.owner.(*windowsJob)
	c.mu.Lock()
	valid := j.job
	j.job = windows.InvalidHandle // inject a real API failure, without losing ownership
	c.mu.Unlock()
	t.Cleanup(func() {
		c.mu.Lock()
		if j.job == windows.InvalidHandle {
			j.job = valid
		}
		c.mu.Unlock()
	})
	if err := p.Stop(false); err == nil {
		t.Fatal("failed job termination reported success")
	}
	if p.GetStatus() != "stopping" {
		t.Fatal("cleanup failure published offline")
	}
	if err := p.Start(); err == nil {
		t.Fatal("restart accepted after cleanup failure")
	}
	if lock, err := acquireRunLock(p.workingDir); err == nil {
		lock.Close()
		t.Fatal("cleanup failure released run lock")
	}
	c.mu.Lock()
	j.job = valid
	c.mu.Unlock()
	if err := p.Stop(false); err != nil {
		t.Fatal(err)
	}
	waitProcessStatus(t, p, "offline")
}
