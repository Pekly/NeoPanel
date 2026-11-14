//go:build linux

package runner

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
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
	if err != nil && !errors.Is(err, syscall.ESRCH) {
		return err
	}
	// SIGKILL delivery is asynchronous. Do not publish "offline" while a
	// surviving group member can still own sockets or write server files.
	deadline := time.Now().Add(5 * time.Second)
	for {
		processes, err := liveProcIdentities("/proc")
		if err != nil {
			return err
		}
		alive := false
		for _, process := range processes {
			if process.Group == cmd.Process.Pid {
				alive = true
				break
			}
		}
		if !alive {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("process group %d did not exit after termination", cmd.Process.Pid)
		}
		time.Sleep(25 * time.Millisecond)
	}
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

func ensureServerProcessesStopped(dir string) error {
	return inspectServerProcesses("/proc", dir, func(path string) (string, error) {
		target, err := os.Readlink(path)
		if errors.Is(err, syscall.ESRCH) {
			return "", os.ErrNotExist
		}
		if os.IsPermission(err) {
			// Proc-directory ownership changes for non-dumpable processes.
			// Read the kernel's UID fields instead; an unreadable/unknown UID
			// cannot be treated as proof that this is an unrelated process.
			if raw, statusErr := os.ReadFile(filepath.Join(filepath.Dir(path), "status")); statusErr == nil {
				if ours, parseErr := processHasUID(string(raw), os.Geteuid()); parseErr == nil && !ours {
					return "", os.ErrNotExist
				}
			}
		}
		return target, err
	})
}

func waitForOwnedExit(cmd *exec.Cmd) error {
	var info unix.Siginfo
	for {
		err := unix.Waitid(unix.P_PID, cmd.Process.Pid, &info, unix.WEXITED|unix.WNOWAIT, nil)
		if err == unix.EINTR {
			continue
		}
		return err
	}
}

func verifyProcessTreeStopped(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	processes, err := liveProcIdentities("/proc")
	if err != nil {
		return err
	}
	for _, process := range processes {
		if process.Group == cmd.Process.Pid {
			return fmt.Errorf("%w: process group %d still contains PID %d", ErrServerProcessAlive, cmd.Process.Pid, process.PID)
		}
	}
	return nil
}
