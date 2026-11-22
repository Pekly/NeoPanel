package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

// Archive creation limits count logical INPUT, not compressed output or allocated
// blocks. These are workload bounds, not a filesystem quota or CPU/RAM cgroup.
const ArchiveMaxBytes int64 = 100 << 30
const ArchiveMaxEntries = 100000
const ArchiveMaxDepth = 64
const ArchiveTimeout = 30 * time.Minute
const ArchiveConcurrency = 2

var ErrArchiveLimit = errors.New("archive creation input limit exceeded")
var ErrArchiveBusy = errors.New("archive creation capacity is busy; retry later")
var archiveSlots = make(chan struct{}, ArchiveConcurrency)

// Reserve standalone backup destinations so concurrent no-replace callers retain
// os.ErrExist semantics even while the first archive is still being written.
var backupTargets sync.Map

func reserveBackupTarget(dir, name string) (func(), error) {
	target, err := filepath.Abs(filepath.Join(dir, name))
	if err != nil {
		return nil, err
	}
	if _, loaded := backupTargets.LoadOrStore(target, struct{}{}); loaded {
		return nil, fmt.Errorf("backup destination already reserved: %w", os.ErrExist)
	}
	return func() { backupTargets.Delete(target) }, nil
}

func archiveSlot(ctx context.Context) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case archiveSlots <- struct{}{}:
		return func() { <-archiveSlots }, nil
	default:
		return nil, ErrArchiveBusy
	}
}

func ArchiveByteLimit(diskMB int64) int64 {
	if diskMB > 0 && diskMB < ArchiveMaxBytes/(1<<20) {
		return diskMB * (1 << 20)
	}
	return ArchiveMaxBytes
}

// BeginArchiveCreation excludes other managed writes and server operations,
// shares the daemon-wide backup/compression slots, and joins shutdown. It never
// stops a server. Call release exactly once, after publication/cleanup.
func (m *ServerManager) BeginArchiveCreation(parent context.Context, id string) (context.Context, int64, func(), error) {
	unlock, err := m.beginOperation(id)
	if err != nil {
		return nil, 0, nil, err
	}
	releaseMutation, err := rooted.TryExclusiveMutation(m.GetWorkingDir(id))
	if err != nil {
		unlock()
		return nil, 0, nil, err
	}
	ctx, cancel := context.WithTimeout(parent, ArchiveTimeout)
	slot, err := archiveSlot(ctx)
	if err != nil {
		cancel()
		releaseMutation()
		unlock()
		return nil, 0, nil, err
	}
	cleanup := func() { cancel(); slot(); releaseMutation(); unlock() }
	m.launchMu.Lock()
	if m.isShuttingDown() {
		m.launchMu.Unlock()
		cleanup()
		return nil, 0, nil, errDaemonShuttingDown
	}
	m.importWG.Add(1) // Existing shutdown auxiliary-work fence.
	m.launchMu.Unlock()
	done := make(chan struct{})
	go func() {
		defer close(done)
		select {
		case <-m.shutdownCh:
			cancel()
		case <-ctx.Done():
		}
	}()
	release := func() { cleanup(); <-done; m.importWG.Done() }
	limit := ArchiveMaxBytes
	if m.db != nil {
		srv, e := m.db.GetServer(id)
		if e != nil {
			release()
			return nil, 0, nil, e
		}
		limit = ArchiveByteLimit(srv.DiskLimitMB)
	}
	return ctx, limit, release, nil
}

// CreateServerBackupContext is the common bounded path for scheduled and manual
// backups. CreateServerBackup should delegate here with context.Background().
func (m *ServerManager) CreateServerBackupContext(parent context.Context, id, name string) (int64, error) {
	ctx, limit, release, err := m.BeginArchiveCreation(parent, id)
	if err != nil {
		return 0, err
	}
	defer release()
	m.mu.RLock()
	dir := m.backupDir
	m.mu.RUnlock()
	return createBackupContext(ctx, m.GetWorkingDir(id), filepath.Join(dir, id), name, limit)
}

// ArchiveBudget is per preflight; all selected roots share one instance. Walk
// uses bounded directory batches (Root.Walk otherwise materializes whole dirs).
type ArchiveBudget struct {
	Bytes   int64
	Entries int
	limit   int64
}

func NewArchiveBudget(limit int64) *ArchiveBudget {
	if limit <= 0 || limit > ArchiveMaxBytes {
		limit = ArchiveMaxBytes
	}
	return &ArchiveBudget{limit: limit}
}
func (b *ArchiveBudget) Walk(ctx context.Context, root *rooted.Root, name string, fn func(string, os.FileInfo) error) error {
	var walk func(string, int) error
	walk = func(p string, depth int) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if depth > ArchiveMaxDepth || b.Entries >= ArchiveMaxEntries {
			return ErrArchiveLimit
		}
		info, err := root.Lstat(p)
		if err != nil {
			return err
		}
		b.Entries++
		if info.Mode().IsRegular() {
			if info.Size() < 0 || info.Size() > b.limit-b.Bytes {
				return fmt.Errorf("%w: logical bytes (maximum %d)", ErrArchiveLimit, b.limit)
			}
			b.Bytes += info.Size()
		}
		if err = fn(p, info); err != nil {
			if err == filepath.SkipDir && info.IsDir() {
				return nil
			}
			return err
		}
		if !info.IsDir() {
			return nil
		}
		if err = root.Check(p); err != nil {
			return err
		}
		d, err := root.Root.OpenRoot(p)
		if err != nil {
			return err
		}
		defer d.Close()
		f, err := d.Open(".")
		if err != nil {
			return err
		}
		defer f.Close()
		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			names, e := f.Readdirnames(128)
			for _, n := range names {
				if err := walk(filepath.Join(p, n), depth+1); err != nil {
					return err
				}
			}
			if e == io.EOF {
				return nil
			}
			if e != nil {
				return e
			}
		}
	}
	depth := len(strings.Split(filepath.ToSlash(filepath.Clean(name)), "/"))
	if name == "." {
		depth = 0
	}
	return walk(name, depth)
}

type archiveContextReader struct {
	ctx context.Context
	in  io.Reader
}

func (r archiveContextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	if len(p) > 32<<10 {
		p = p[:32<<10]
	}
	return r.in.Read(p)
}

// CopyArchiveFile checks the opened handle against preflight BEFORE reading and
// copies exactly the approved size. A one-byte probe detects growth without
// compressing it. Shrink/growth/cancellation fail the entire unpublished archive.
func CopyArchiveFile(ctx context.Context, out io.Writer, f *os.File, size int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || size < 0 || info.Size() != size {
		return fmt.Errorf("%w: source changed size", ErrArchiveLimit)
	}
	in := archiveContextReader{ctx, f}
	if _, err = io.CopyN(out, in, size); err != nil {
		return err
	}
	var extra [1]byte
	n, err := in.Read(extra[:])
	if n != 0 {
		return fmt.Errorf("%w: source grew", ErrArchiveLimit)
	}
	if err != io.EOF {
		if err != nil {
			return err
		}
		return io.ErrNoProgress
	}
	return ctx.Err()
}
