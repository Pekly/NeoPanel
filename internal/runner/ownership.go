package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Serialize late watchdog/stop requests with final cleanup for this exact run.
type ownedCommand struct {
	mu     sync.Mutex
	cmd    *exec.Cmd
	reaped bool
	err    error
}

func (c *ownedCommand) stop(final bool) error {
	if c == nil {
		return fmt.Errorf("process ownership is unavailable")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.reaped {
		return c.err
	}
	err := terminateOwnedCommand(c.cmd, final)
	if final {
		c.reaped, c.err = true, err
	}
	return err
}

// Each launch has a unique scope name, so cleanup can only address this run.
func stopOwnedScope(cmd *exec.Cmd) error {
	if cmd == nil || filepath.Base(cmd.Path) != "systemd-run" {
		return nil
	}
	for _, arg := range cmd.Args[1:] {
		if strings.HasPrefix(arg, "--unit=neopanel-") && strings.HasSuffix(arg, ".scope") {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			stop := exec.CommandContext(ctx, "systemctl", "stop", strings.TrimPrefix(arg, "--unit="))
			stop.Env = append(os.Environ(), "LC_ALL=C")
			out, err := stop.CombinedOutput()
			// A scope that has already exited is collected automatically.
			if err != nil && !strings.Contains(string(out), "not loaded") && !strings.Contains(string(out), "not found") {
				return fmt.Errorf("stop owned systemd scope: %w: %s", err, strings.TrimSpace(string(out)))
			}
			return nil
		}
	}
	return nil
}

func terminateOwnedCommand(cmd *exec.Cmd, final bool) error {
	// Kill the process group first, including the scope launcher; systemctl also
	// catches descendants that deliberately create another process group.
	var treeErr error
	if !final || runtime.GOOS == "linux" {
		treeErr = killProcessTree(cmd)
	}
	scopeErr := stopOwnedScope(cmd)
	if scopeErr != nil {
		return scopeErr
	}
	if treeErr != nil && !os.IsNotExist(treeErr) && treeErr != os.ErrProcessDone {
		return treeErr
	}
	return nil
}

func checkServerWorldLocks(dir, serverType string) error {
	switch strings.ToLower(serverType) {
	case "velocity", "bungeecord", "waterfall":
		return nil
	}
	name := "world"
	if raw, err := os.ReadFile(filepath.Join(dir, "server.properties")); err == nil {
		for _, line := range strings.Split(string(raw), "\n") {
			if key, value, ok := strings.Cut(strings.TrimSpace(line), "="); ok && strings.TrimSpace(key) == "level-name" && strings.TrimSpace(value) != "" {
				name = strings.TrimSpace(value)
			}
		}
	}
	for _, world := range []string{name, name + "_nether", name + "_the_end"} {
		if err := checkWorldLock(filepath.Join(dir, world, "session.lock")); err != nil {
			return err
		}
	}
	return nil
}
