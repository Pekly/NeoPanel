package runner

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
)

var (
	ErrServerOperation  = errors.New("another server power, backup, restore, reinstall or deletion operation is in progress; retry when it finishes")
	ErrServerNotStopped = errors.New("server must be stopped before restoring a backup")
	ErrInvalidArchive   = errors.New("invalid backup archive")
	ErrRestoreRecovery  = errors.New("restore publication and rollback failed; manual recovery required")
)

const restoreMaxBytes int64 = 100 << 30 // hard ceiling, further bounded by the server disk limit
const restoreMaxEntries = 100000

func (m *ServerManager) beginOperation(id string) (func(), error) {
	value, _ := m.operations.LoadOrStore(id, &sync.Mutex{})
	lock := value.(*sync.Mutex)
	if !lock.TryLock() {
		return nil, ErrServerOperation
	}
	if reason, blocked := m.restoreBlocked.Load(id); blocked {
		lock.Unlock()
		return nil, reason.(error)
	}
	return lock.Unlock, nil
}

func (m *ServerManager) CreateServerBackup(serverID, name string) (int64, error) {
	unlock, err := m.beginOperation(serverID)
	if err != nil {
		return 0, err
	}
	defer unlock()
	m.mu.RLock()
	dir := m.backupDir
	m.mu.RUnlock()
	return CreateBackup(m.GetWorkingDir(serverID), filepath.Join(dir, serverID), name)
}

type RestoreResult struct {
	Message    string `json:"message"`
	Name       string `json:"name"`
	SafetyCopy string `json:"safety_copy"`
}

// RestoreBackup never extracts into the live tree. The old directory remains in
// a private sibling recovery directory after success (operator cleanup is explicit).
// The runtime uses the daemon's identity, so archive owners/setuid bits are not restored.
func (m *ServerManager) RestoreBackup(ctx context.Context, srv *database.Server, name string) (RestoreResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()
	result := RestoreResult{Name: name}
	if !safeBackupComponent(srv.ID) || !safeBackupComponent(name) || !strings.HasSuffix(name, ".tar.gz") {
		return result, fmt.Errorf("%w: use an existing local .tar.gz backup basename", ErrInvalidArchive)
	}
	unlock, err := m.beginOperation(srv.ID)
	if err != nil {
		return result, err
	}
	defer unlock()
	if m.GetStatus(srv.ID) != "offline" {
		return result, ErrServerNotStopped
	}
	m.mu.Lock()
	m.cancelRetryLocked(srv.ID)
	m.restorePaused.Store(srv.ID, true)
	dir := m.backupDir
	m.mu.Unlock()
	base, err := filepath.EvalSymlinks(m.baseDir)
	if err != nil {
		return result, fmt.Errorf("resolve server root: %w", err)
	}
	base, err = filepath.Abs(base)
	if err != nil {
		return result, err
	}
	live := filepath.Join(base, srv.ID)
	info, err := os.Lstat(live)
	if err != nil {
		return result, fmt.Errorf("read stopped server directory: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return result, fmt.Errorf("%w: server directory must not be a symlink", ErrInvalidArchive)
	}
	root, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return result, fmt.Errorf("resolve backup root: %w", err)
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return result, err
	}
	if rel, err := filepath.Rel(live, root); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return result, fmt.Errorf("%w: backup root must be outside the live server directory", ErrInvalidArchive)
	}
	backupDir := filepath.Join(root, srv.ID)
	st, err := os.Lstat(backupDir)
	if err != nil {
		return result, err
	}
	if !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return result, fmt.Errorf("%w: backup directory must not be a symlink", ErrInvalidArchive)
	}
	archive := filepath.Join(backupDir, name)
	st, err = os.Lstat(archive)
	if err != nil {
		return result, err
	}
	if !st.Mode().IsRegular() {
		return result, fmt.Errorf("%w: backup must be a regular file, not a link", ErrInvalidArchive)
	}
	f, err := os.Open(archive)
	if err != nil {
		return result, err
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil {
		return result, err
	}
	if !os.SameFile(st, opened) {
		return result, fmt.Errorf("%w: backup changed while opening; retry", ErrInvalidArchive)
	}
	stage, err := os.MkdirTemp(base, ".restore-"+srv.ID+"-")
	if err != nil {
		return result, fmt.Errorf("create restore staging directory: %w", err)
	}
	keep := false
	defer func() {
		if !keep {
			_ = os.RemoveAll(stage)
		}
	}()
	fresh := filepath.Join(stage, "restored")
	if err := os.Mkdir(fresh, 0755); err != nil {
		return result, err
	}
	budget := restoreMaxBytes
	if srv.DiskLimitMB > 0 && int64(srv.DiskLimitMB) < budget/(1<<20) {
		budget = int64(srv.DiskLimitMB) * (1 << 20)
	}
	if err := extractRestoreArchive(ctx, f, fresh, budget, restoreMaxEntries); err != nil {
		return result, fmt.Errorf("restore extraction failed (live data unchanged): %w", err)
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if m.isShuttingDown() {
		return result, fmt.Errorf("daemon shutting down; live data unchanged")
	}
	// Recheck the root identity before publication, rather than following a replaced link.
	current, err := os.Lstat(live)
	if err != nil || !os.SameFile(info, current) {
		return result, fmt.Errorf("%w: server directory changed during extraction; retry", ErrInvalidArchive)
	}
	old := filepath.Join(stage, "original")
	if err := publishRestore(live, fresh, old, os.Rename); err != nil {
		if errors.Is(err, ErrRestoreRecovery) {
			keep = true
			m.restoreBlocked.Store(srv.ID, err)
		}
		return result, err
	}
	keep = true
	result.Message = "Backup restored successfully; server remains stopped. Original data retained in safety_copy; remove it manually after verification."
	result.SafetyCopy = filepath.Join(filepath.Base(stage), "original")
	return result, nil
}

func safeBackupComponent(name string) bool {
	return name != "" && name != "." && name != ".." && filepath.IsLocal(name) && !strings.ContainsAny(name, "/\\:\x00") && filepath.Base(name) == name && !strings.HasSuffix(name, ".") && !strings.HasSuffix(name, " ")
}

// Rename is injected for deterministic publication/rollback failure tests.
func publishRestore(live, fresh, old string, rename func(string, string) error) error {
	if err := rename(live, old); err != nil {
		return fmt.Errorf("retain original directory (live data unchanged): %w", err)
	}
	if err := rename(fresh, live); err != nil {
		if rollback := rename(old, live); rollback != nil {
			return fmt.Errorf("%w: publish: %v; rollback: %v; original at %s, extracted data at %s; recover original directory before starting", ErrRestoreRecovery, err, rollback, old, fresh)
		}
		return fmt.Errorf("publish restored directory failed; original data rolled back: %w", err)
	}
	return nil
}

type restoreReader struct {
	ctx       context.Context
	r         io.Reader
	remaining int64
	read      int64
}

func (r *restoreReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	if r.remaining <= 0 {
		return 0, fmt.Errorf("%w: decompressed archive budget exceeded", ErrInvalidArchive)
	}
	if int64(len(p)) > r.remaining {
		p = p[:r.remaining]
	}
	n, err := r.r.Read(p)
	r.read += int64(n)
	r.remaining -= int64(n)
	return n, err
}

func extractRestoreArchive(ctx context.Context, in io.Reader, dest string, maxBytes int64, maxEntries int) error {
	gz, err := gzip.NewReader(in)
	if err != nil {
		return fmt.Errorf("%w: gzip header: %v", ErrInvalidArchive, err)
	}
	defer gz.Close()
	reader := &restoreReader{ctx: ctx, r: gz, remaining: maxBytes + int64(maxEntries)*4096 + (1 << 20)}
	tr := tar.NewReader(reader)
	names := map[string]bool{} // true = directory; includes implicit parents
	explicit := map[string]bool{}
	var total int64
	for count := 0; ; count++ {
		before := reader.read
		h, err := tr.Next()
		if err == io.EOF {
			// archive/tar accepts EOF without the required two end blocks. We don't.
			if reader.read-before < 1024 {
				return fmt.Errorf("%w: missing tar end-of-archive blocks", ErrInvalidArchive)
			}
			break
		}
		if err != nil {
			return fmt.Errorf("%w: tar header: %v", ErrInvalidArchive, err)
		}
		if count >= maxEntries {
			return fmt.Errorf("%w: archive exceeds %d entries", ErrInvalidArchive, maxEntries)
		}
		name := strings.TrimSuffix(h.Name, "/")
		if name == "" || len(name) > 4096 || strings.ContainsAny(name, "\\:\x00") || strings.HasPrefix(name, "/") {
			return fmt.Errorf("%w: unsafe entry %q", ErrInvalidArchive, h.Name)
		}
		parts := strings.Split(name, "/")
		for _, part := range parts {
			if !safeBackupComponent(part) {
				return fmt.Errorf("%w: unsafe entry %q", ErrInvalidArchive, h.Name)
			}
		}
		if h.Typeflag != tar.TypeReg && h.Typeflag != tar.TypeDir {
			return fmt.Errorf("%w: links/devices/unsupported entry %q", ErrInvalidArchive, h.Name)
		}
		isDir := h.Typeflag == tar.TypeDir
		// Case-folding also prevents aliases/conflicts on Windows filesystems.
		key := strings.ToLower(name)
		if explicit[key] {
			return fmt.Errorf("%w: duplicate entry %q", ErrInvalidArchive, h.Name)
		}
		if dir, exists := names[key]; exists && (!dir || !isDir) {
			return fmt.Errorf("%w: conflicting entry %q", ErrInvalidArchive, h.Name)
		}
		for i := 1; i < len(parts); i++ {
			parent := strings.ToLower(strings.Join(parts[:i], "/"))
			if dir, exists := names[parent]; exists && !dir {
				return fmt.Errorf("%w: file used as directory %q", ErrInvalidArchive, h.Name)
			}
			names[parent] = true
		}
		explicit[key], names[key] = true, isDir
		if len(names) > maxEntries {
			return fmt.Errorf("%w: archive exceeds %d paths including parent directories", ErrInvalidArchive, maxEntries)
		}
		target := filepath.Join(dest, filepath.FromSlash(name))
		if isDir {
			if h.Size != 0 {
				return fmt.Errorf("%w: directory with content", ErrInvalidArchive)
			}
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}
		if h.Size < 0 || h.Size > maxBytes-total {
			return fmt.Errorf("%w: extracted data exceeds %d bytes", ErrInvalidArchive, maxBytes)
		}
		total += h.Size
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		_, copyErr := io.CopyN(out, tr, h.Size)
		if copyErr != nil {
			out.Close()
			return fmt.Errorf("%w: incomplete entry %q: %v", ErrInvalidArchive, name, copyErr)
		}
		// Only ordinary owner access and executable bits; never restore privileged modes/owners.
		modeErr := out.Chmod(0600 | os.FileMode(h.Mode)&0111)
		if err := errors.Join(modeErr, out.Sync(), out.Close()); err != nil {
			return err
		}
	}
	// Consume all gzip data to verify CRC/footer, rejecting nonzero trailing data
	// (including concatenated tar archives) and unbounded padding.
	var padding int64
	buf := make([]byte, 32768)
	for {
		n, err := reader.Read(buf)
		padding += int64(n)
		if padding > 1<<20 {
			return fmt.Errorf("%w: excessive tar padding", ErrInvalidArchive)
		}
		for _, b := range buf[:n] {
			if b != 0 {
				return fmt.Errorf("%w: unexpected data after tar terminator", ErrInvalidArchive)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("%w: gzip footer/data: %v", ErrInvalidArchive, err)
		}
	}
	return nil
}
