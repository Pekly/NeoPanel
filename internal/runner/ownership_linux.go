//go:build linux

package runner

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// The file is never unlinked: removing a locked inode allows a second owner.
func acquireRunLock(dir string) (*os.File, error) {
	path, err := serverRunLockPath(dir)
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, fmt.Errorf("another panel process owns this server directory; stop its existing instance before starting: %w", err)
	}
	return f, nil
}

func configureProcessOwnership(cmd *exec.Cmd, lock *os.File) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGKILL}
	// Keep ownership if a native child outlives an abruptly terminated panel.
	cmd.ExtraFiles = append(cmd.ExtraFiles, lock)
}

func killProcessTree(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}

// Minecraft uses a POSIX file lock, not mere file existence. Probe without
// deleting/truncating the lock or killing a process we do not own.
func checkWorldLock(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("cannot inspect world lock %s: %w", path, err)
	}
	defer f.Close()
	lock := syscall.Flock_t{Type: syscall.F_WRLCK, Whence: 0, Start: 0, Len: 0}
	if err := syscall.FcntlFlock(f.Fd(), syscall.F_SETLK, &lock); err != nil {
		return fmt.Errorf("world is already in use (%s); stop the existing Minecraft instance before starting; do not delete session.lock: %w", path, err)
	}
	return nil
}
