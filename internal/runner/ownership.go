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
	return err
}

// Observe exit without reaping first. Keeping the child unreaped pins its PID
// and process-group identity until cleanup finishes under the same mutex used
// by Stop. Never send a group signal after cmd.Wait releases that identity.
func (c *ownedCommand) wait() (error, error) {
	observeErr := waitForOwnedExit(c.cmd)
	c.mu.Lock()
	defer c.mu.Unlock()
	var cleanupErr error
	if observeErr == nil {
		cleanupErr = terminateOwnedCommand(c.cmd, true)
	} else {
		cleanupErr = fmt.Errorf("cannot safely observe server exit: %w", observeErr)
		// Fail closed without leaving Stop blocked behind a running launcher.
		// This targets only the still-owned child, never an unverified group.
		_ = c.cmd.Process.Kill()
	}
	waitErr := c.cmd.Wait()
	c.reaped, c.err = true, cleanupErr
	return waitErr, cleanupErr
}

func (c *ownedCommand) reconcile(dir string) error {
	if c == nil {
		return fmt.Errorf("process ownership is unavailable")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.reaped {
		return fmt.Errorf("server process has not been reaped")
	}
	// Only the unique scope may be stopped after reap. Numeric process groups
	// are inspected, never signalled here, since their IDs can now be reused.
	if err := stopOwnedScope(c.cmd); err != nil {
		return err
	}
	if err := verifyProcessTreeStopped(c.cmd); err != nil {
		return err
	}
	if err := ensureServerProcessesStopped(dir); err != nil {
		return err
	}
	c.err = nil
	return nil
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
			// Do not interpret generic "not found" errors (for example a missing
			// system bus) as proof the scope and its children are gone.
			verifyCtx, verifyCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer verifyCancel()
			verify := exec.CommandContext(verifyCtx, "systemctl", "show", strings.TrimPrefix(arg, "--unit="), "--property=LoadState,ActiveState")
			verify.Env = append(os.Environ(), "LC_ALL=C")
			state, verifyErr := verify.CombinedOutput()
			if verifyErr == nil && scopeStopped(string(state)) {
				return nil
			}
			return fmt.Errorf("cannot confirm owned systemd scope stopped: stop=%v (%s), verify=%v (%s)", err, strings.TrimSpace(string(out)), verifyErr, strings.TrimSpace(string(state)))
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

func scopeStopped(output string) bool {
	properties := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		if key, value, ok := strings.Cut(strings.TrimSpace(line), "="); ok {
			properties[key] = value
		}
	}
	return (properties["LoadState"] == "not-found" || properties["LoadState"] == "loaded") && properties["ActiveState"] == "inactive"
}
