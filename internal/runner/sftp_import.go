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

const (
	ImportConflictEmpty    = "empty"
	ImportConflictPreserve = "preserve"
	ImportConflictReplace  = "replace"
)

var ErrImportTargetNotEmpty = errors.New("import requires a strictly empty server directory; clear all files first")

// ImportServerFiles imports into a stopped, local, strictly empty server.
func (m *ServerManager) ImportServerFiles(ctx context.Context, srv *database.Server, transfer func(context.Context, string) error) (RestoreResult, error) {
	return m.ImportServerFilesWithMode(ctx, srv, ImportConflictEmpty, transfer)
}

// ImportServerFilesWithMode imports into a stopped, local server using the specified
// conflict mode: "empty" (default, requires empty target), "preserve" (keeps existing
// files and skips conflicts), or "replace" (replaces existing files while retaining a
// safety copy). transfer receives a fresh private staging directory, never the live
// directory. It must honor ctx and finish all file writes before returning; shutdown
// cancels ctx and waits for it to return.
func (m *ServerManager) ImportServerFilesWithMode(ctx context.Context, srv *database.Server, mode string, transfer func(context.Context, string) error) (RestoreResult, error) {
	var result RestoreResult
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		mode = ImportConflictEmpty
	}
	switch mode {
	case ImportConflictEmpty, ImportConflictPreserve, ImportConflictReplace:
	default:
		return result, fmt.Errorf("invalid import conflict mode: %q", mode)
	}
	requireEmpty := mode == ImportConflictEmpty
	if srv == nil || !safeBackupComponent(srv.ID) || strings.HasPrefix(srv.ID, ".") || (srv.NodeID != "" && srv.NodeID != "local") {
		return result, errors.New("import requires a safe local server ID")
	}
	if transfer == nil {
		return result, errors.New("import transfer callback is required")
	}
	unlock, err := m.beginFileOperation(srv.ID)
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
	info, err := importTargetDir(parent, srv.ID, nil, requireEmpty)
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
	if mode != ImportConflictPreserve {
		if err := checkRestoreRetention(parent, srv.ID); err != nil {
			return result, err
		}
	}
	if _, err := importTargetDir(parent, srv.ID, info, requireEmpty); err != nil {
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
	if _, err := importTargetDir(parent, srv.ID, info, requireEmpty); err != nil {
		return result, err
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if mode == ImportConflictPreserve {
		if err := mergePreserveImport(fresh, live); err != nil {
			return result, fmt.Errorf("merge imported files: %w", err)
		}
		result.Message = "Server files imported; existing files were preserved and server remains stopped."
		return result, nil
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
	if mode == ImportConflictReplace {
		result.Message = "Server files imported and existing files replaced; server remains stopped. Previous files retained in safety_copy; remove it manually after verification."
	} else {
		result.Message = "Server files imported successfully; server remains stopped. Original directory retained in safety_copy; remove it manually after verification."
	}
	result.SafetyCopy = old
	return result, nil
}

// mergePreserveImport moves files and directories from src into dst without
// overwriting any existing file, directory, or symlink in dst.
func mergePreserveImport(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		srcPath := filepath.Join(src, name)
		dstPath := filepath.Join(dst, name)
		dstInfo, err := os.Lstat(dstPath)
		if err == nil {
			// Destination exists: if both are real directories, recurse to merge children;
			// otherwise preserve existing destination entry and skip.
			if entry.IsDir() && dstInfo.IsDir() && dstInfo.Mode()&os.ModeSymlink == 0 {
				if err := mergePreserveImport(srcPath, dstPath); err != nil {
					return err
				}
			}
			continue
		}
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err := os.Rename(srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}

// Check both path and opened-directory identities; never accept a symlink or a
// replaced directory. When requireEmpty is true, reject any non-empty target.
func importTargetDir(parent *os.Root, id string, expected os.FileInfo, requireEmpty bool) (os.FileInfo, error) {
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
	if requireEmpty {
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
	}
	current, err := parent.Lstat(id)
	if err != nil || !os.SameFile(info, current) {
		return nil, errors.New("import target directory changed while checking target")
	}
	return info, nil
}
