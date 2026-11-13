//go:build windows

package runner

import (
	"fmt"
	"os"
	"os/exec"

	"golang.org/x/sys/windows"
)

func acquireRunLock(dir string) (*os.File, error) {
	path, err := serverRunLockPath(dir)
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &windows.Overlapped{}); err != nil {
		f.Close()
		return nil, fmt.Errorf("another panel process owns this server directory: %w", err)
	}
	return f, nil
}

func configureProcessOwnership(cmd *exec.Cmd, lock *os.File) {}
func killProcessTree(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	// Windows development fallback; Linux production uses process groups/scopes.
	return cmd.Process.Kill()
}
func checkWorldLock(path string) error { return nil }
