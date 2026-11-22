package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Pekly/NeoPanel/internal/database"
)

var ErrImportTargetNotEmpty = errors.New("import requires a strictly empty server directory; clear all files first")

// ImportServerFiles imports into a stopped, local, strictly empty server. transfer
// receives a fresh private staging directory, never the live directory. It must
// honor ctx and finish all file writes before returning; shutdown cancels ctx and
// waits for it to return. Transport validation and transfer limits belong to the
// caller. Success leaves the server stopped and retains even the empty original
// directory in SafetyCopy (pre-existing file handles may still write into it).
func (m *ServerManager) ImportServerFiles(ctx context.Context, srv *database.Server, transfer func(context.Context, string) error) (RestoreResult, error) {
	var result RestoreResult
	if srv == nil || !safeBackupComponent(srv.ID) || strings.HasPrefix(srv.ID, ".") || (srv.NodeID != "" && srv.NodeID != "local") {
		return result, errors.New("import requires a safe local server ID")
	}
	if transfer == nil {
		return result, errors.New("import transfer callback is required")
	}
	unlock, err := m.beginOperation(srv.ID)
	if err != nil {
		return result, err
	}
	defer unlock()

	// Register inside the shutdown fence: no Add can race a shutdown Wait.
	// Never hold launchMu during the potentially slow network transfer.
	m.launchMu.Lock()
	if m.isShuttingDown() {
		m.launchMu.Unlock()
		return result, errDaemonShuttingDown
	}
	m.importWG.Add(1)
	defer m.importWG.Done()
	ctx, cancel := context.WithCancel(ctx)
	watchDone := make(chan struct{})
	go func() {
		defer close(watchDone)
		select {
		case <-m.shutdownCh:
			cancel()
		case <-ctx.Done():
		}
	}()
	defer func() { cancel(); <-watchDone }()
	if err = ctx.Err(); err == nil {
		err = m.invalidateIdleWake(srv.ID)
	}
	if err == nil && m.GetStatus(srv.ID) != "offline" {
		err = ErrServerNotStopped
	}
	if err == nil {
		m.mu.Lock()
		m.cancelRetryLocked(srv.ID)
		m.restorePaused.Store(srv.ID, true)
		m.mu.Unlock()
	}
	m.launchMu.Unlock()
	if err != nil {
		return result, err
	}

	base, err := filepath.EvalSymlinks(m.baseDir)
	if err != nil {
		return result, fmt.Errorf("resolve server root: %w", err)
	}
	base, err = filepath.Abs(base)
	if err != nil {
		return result, err
	}
	parent, err := os.OpenRoot(base)
	if err != nil {
		return result, err
	}
	defer parent.Close()
	info, err := emptyImportTarget(parent, srv.ID, nil)
	if err != nil {
		return result, err
	}
	live := filepath.Join(base, srv.ID)
	lock, err := acquireRunLock(live)
	if err != nil {
		return result, fmt.Errorf("import refused: %w", err)
	}
	defer lock.Close()
	if err := ensureServerProcessesStopped(live); err != nil {
		return result, err
	}
	if err := checkRestoreRetention(parent, srv.ID); err != nil {
		return result, err
	}
	if _, err := emptyImportTarget(parent, srv.ID, info); err != nil {
		return result, err
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	stage, err := os.MkdirTemp(base, ".restore-"+srv.ID+"-")
	if err != nil {
		return result, fmt.Errorf("create import staging directory: %w", err)
	}
	keep := false
	defer func() {
		if !keep {
			_ = os.RemoveAll(stage)
		}
	}()
	fresh := filepath.Join(stage, "restored")
	if err := os.Mkdir(fresh, 0700); err != nil {
		return result, err
	}
	if err := transfer(ctx, fresh); err != nil {
		return result, fmt.Errorf("import transfer failed (live data unchanged): %w", err)
	}

	// Publication and shutdown have a single ordering point. A shutdown that
	// wins this fence cannot publish, even if the transfer ignored cancellation.
	m.launchMu.Lock()
	defer m.launchMu.Unlock()
	if m.isShuttingDown() {
		return result, errDaemonShuttingDown
	}
	if _, err := emptyImportTarget(parent, srv.ID, info); err != nil {
		return result, err
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	stageName := filepath.Base(stage)
	old := filepath.Join(stageName, "original")
	if err := publishRestore(srv.ID, filepath.Join(stageName, "restored"), old, parent.Rename); err != nil {
		if errors.Is(err, ErrRestoreRecovery) {
			keep = true
			m.restoreBlocked.Store(srv.ID, err)
		}
		return result, err
	}
	keep = true
	result.Message = "Server files imported successfully; server remains stopped. Original directory retained in safety_copy; remove it manually after verification."
	result.SafetyCopy = old
	return result, nil
}

// Check both path and opened-directory identities; never accept a symlink or a
// replaced (even empty) directory. Read one name, including hidden/generated
// files: there are deliberately no exceptions to the empty-target requirement.
func emptyImportTarget(parent *os.Root, id string, expected os.FileInfo) (os.FileInfo, error) {
	info, err := parent.Lstat(id)
	if err != nil {
		return nil, fmt.Errorf("read import target: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("import target must be a real directory, not a symlink")
	}
	if expected != nil && !os.SameFile(expected, info) {
		return nil, errors.New("import target directory changed during transfer")
	}
	f, err := parent.Open(id)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !os.SameFile(info, opened) {
		return nil, errors.New("import target directory changed while opening")
	}
	names, err := f.Readdirnames(1)
	if len(names) != 0 {
		return nil, ErrImportTargetNotEmpty
	}
	if err != io.EOF {
		if err == nil {
			err = errors.New("could not confirm empty import target")
		}
		return nil, err
	}
	current, err := parent.Lstat(id)
	if err != nil || !os.SameFile(info, current) {
		return nil, errors.New("import target directory changed while checking emptiness")
	}
	return info, nil
}
