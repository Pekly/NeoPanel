//go:build windows

package runner

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"
	"unsafe"

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

// Job Objects manage lifetime/resources, NOT filesystem or security isolation.
// Only the daemon owns the non-inheritable, unnamed job handle. No breakaway
// flags are enabled. Closing that handle on daemon death kills every member.
// Access to the daemon itself by same-user/admin code is not a security boundary.
type windowsJob struct {
	job    windows.Handle
	parent windows.Handle
	root   windows.Handle
}

// Go's exec.Cmd does not expose PROC_THREAD_ATTRIBUTE_JOB_LIST. Instead create
// a never-resumed surrogate atomically IN the job, then use it as the workload's
// parent. Windows inherits job membership during CreateProcess, before user code
// can run. This avoids even the daemon-death gap of Start-suspended/Assign/Resume.
// The surrogate is this executable but NEVER executes its entry point. This
// requires Windows 10+ job-list attributes; unsupported/restricted hosts fail closed.
func startOwnedCommand(cmd *exec.Cmd, lock *os.File, memoryMB int64) (owned *ownedCommand, err error) {
	if cmd.SysProcAttr != nil {
		return nil, fmt.Errorf("custom Windows process attributes are not supported for owned launches")
	}
	j, err := newWindowsJob(memoryMB)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			cleanupErr := j.terminate()
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
				_ = cmd.Wait()
			}
			if cleanupErr == nil {
				cleanupErr = j.close()
			}
			err = errors.Join(err, cleanupErr)
			if cleanupErr != nil {
				// Preserve the exact job and run lock for explicit reconciliation.
				owned = &ownedCommand{cmd: cmd, owner: j, reaped: true, err: cleanupErr}
			}
		}
	}()
	cmd.SysProcAttr = &syscall.SysProcAttr{ParentProcess: syscall.Handle(j.parent), CreationFlags: windows.CREATE_NO_WINDOW}
	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("start job-owned workload: %w", err)
	}
	// exec.Cmd still pins the process object even if the workload already exited.
	j.root, err = windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(cmd.Process.Pid))
	if err != nil {
		return nil, fmt.Errorf("retain workload completion handle: %w", err)
	}
	return &ownedCommand{cmd: cmd, owner: j}, nil
}

func newWindowsJob(memoryMB int64) (_ *windowsJob, err error) {
	if memoryMB < 0 || uint64(memoryMB) > uint64(^uintptr(0))/(1024*1024) {
		return nil, fmt.Errorf("Windows job memory ceiling is out of range: %d MiB", memoryMB)
	}
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, fmt.Errorf("create Windows job: %w", err)
	}
	j := &windowsJob{job: job}
	defer func() {
		if err != nil {
			_ = j.close()
		}
	}()
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if memoryMB > 0 {
		limits.BasicLimitInformation.LimitFlags |= windows.JOB_OBJECT_LIMIT_JOB_MEMORY
		limits.JobMemoryLimit = uintptr(memoryMB) * 1024 * 1024
	}
	if _, err = windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits))); err != nil {
		return nil, fmt.Errorf("set Windows job limits: %w", err)
	}
	attrs, err := windows.NewProcThreadAttributeList(1)
	if err != nil {
		return nil, err
	}
	defer attrs.Delete()
	const procThreadAttributeJobList = 0x0002000d
	if err = attrs.Update(procThreadAttributeJobList, unsafe.Pointer(&job), unsafe.Sizeof(job)); err != nil {
		return nil, fmt.Errorf("atomic Windows job launch is unavailable: %w", err)
	}
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	app, err := windows.UTF16PtrFromString(exe)
	if err != nil {
		return nil, err
	}
	command, err := windows.UTF16PtrFromString(syscall.EscapeArg(exe))
	if err != nil {
		return nil, err
	}
	si := windows.StartupInfoEx{StartupInfo: windows.StartupInfo{Cb: uint32(unsafe.Sizeof(windows.StartupInfoEx{}))}, ProcThreadAttributeList: attrs.List()}
	var pi windows.ProcessInformation
	err = windows.CreateProcess(app, command, nil, nil, false, windows.CREATE_SUSPENDED|windows.CREATE_NO_WINDOW|windows.EXTENDED_STARTUPINFO_PRESENT, nil, nil, &si.StartupInfo, &pi)
	runtime.KeepAlive(job)
	if err != nil {
		return nil, fmt.Errorf("create atomic job parent: %w", err)
	}
	j.parent = pi.Process
	windows.CloseHandle(pi.Thread)
	return j, nil
}

func (j *windowsJob) waitExit() error {
	status, err := windows.WaitForSingleObject(j.root, windows.INFINITE)
	if err != nil {
		return err
	}
	if status != windows.WAIT_OBJECT_0 {
		return fmt.Errorf("unexpected process wait status %d", status)
	}
	return nil
}

type jobAccounting struct {
	UserTime, KernelTime, PeriodUserTime, PeriodKernelTime           int64
	PageFaults, TotalProcesses, ActiveProcesses, TerminatedProcesses uint32
}

func (j *windowsJob) terminate() error {
	if j.job == 0 {
		return nil
	}
	if err := windows.TerminateJobObject(j.job, 1); err != nil {
		return fmt.Errorf("terminate Windows job: %w", err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		var accounting jobAccounting
		if err := windows.QueryInformationJobObject(j.job, windows.JobObjectBasicAccountingInformation, uintptr(unsafe.Pointer(&accounting)), uint32(unsafe.Sizeof(accounting)), nil); err != nil {
			return fmt.Errorf("verify Windows job completion: %w", err)
		}
		if accounting.ActiveProcesses == 0 {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("Windows job still has %d active processes; restart blocked", accounting.ActiveProcesses)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// Called under ownedCommand.mu after waitExit, or before publishing a launch.
func (j *windowsJob) close() error {
	var errs []error
	for _, h := range []*windows.Handle{&j.job, &j.root, &j.parent} {
		if *h != 0 {
			if err := windows.CloseHandle(*h); err != nil {
				errs = append(errs, err)
			} else {
				*h = 0
			}
		}
	}
	return errors.Join(errs...)
}

// os.Root rejects escaping links; the common precheck and inode check reject links.
const worldLockOpenFlags = 0

func probeWorldLock(f *os.File, path string) error {
	// Java FileChannel.tryLock covers this byte. Never truncate/delete the file.
	if err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &windows.Overlapped{}); err != nil {
		return fmt.Errorf("world is in use or its lock cannot be verified (%s); do not delete session.lock: %w", path, err)
	}
	return nil
}

// Windows cannot reliably identify arbitrary pre-existing processes by working
// directory. This only probes conventional world locks; it does NOT certify that
// manually launched/legacy untracked servers are absent after a panel restart.
// Tracked runs are certified separately by their exact Job Object accounting.
func ensureServerProcessesStopped(dir string) error { return checkServerWorldLocks(dir, "") }

// A Windows command without its job owner is never safe to tree-kill or certify.
func killProcessTree(cmd *exec.Cmd) error {
	return fmt.Errorf("Windows process tree has no Job Object owner")
}
func waitForOwnedExit(cmd *exec.Cmd) error {
	return fmt.Errorf("Windows process has no Job Object completion handle")
}
func verifyProcessTreeStopped(cmd *exec.Cmd) error {
	return fmt.Errorf("Windows process tree has no Job Object owner")
}
