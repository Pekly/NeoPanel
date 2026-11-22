package runner

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/rooted"
)

type archiveWriteFunc func([]byte) (int, error)

func (f archiveWriteFunc) Write(p []byte) (int, error) { return f(p) }

func TestBackupLogicalPreflight(t *testing.T) {
	source, dest := t.TempDir(), t.TempDir()
	f, err := os.Create(filepath.Join(source, "sparse"))
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Truncate(16 << 20); err != nil {
		t.Fatal(err)
	}
	f.Close()
	wrote := false
	err = writeBackupArchiveContext(context.Background(), source, archiveWriteFunc(func(p []byte) (int, error) { wrote = true; return len(p), nil }), 1<<20)
	if !errors.Is(err, ErrArchiveLimit) || wrote {
		t.Fatalf("preflight: %v, wrote=%v", err, wrote)
	}
	if _, err = createBackupContext(context.Background(), source, dest, "too-large.tar.gz", 1<<20); !errors.Is(err, ErrArchiveLimit) {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dest)
	if err != nil || len(entries) != 0 {
		t.Fatal(entries, err)
	}
	st, err := os.Stat(filepath.Join(source, "sparse"))
	if err != nil || st.Size() != 16<<20 {
		t.Fatal(st, err)
	}
}

func TestBackupConfiguredDiskBudget(t *testing.T) {
	base := t.TempDir()
	db, err := database.InitDB(filepath.Join(base, "panel.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	srv := &database.Server{ID: "small", Name: "small", Type: "paper", DiskLimitMB: 1}
	if err = db.CreateServer(srv); err != nil {
		t.Fatal(err)
	}
	m := NewServerManager(filepath.Join(base, "servers"), db)
	defer m.ShutdownAll()
	m.SetBackupDirectory(filepath.Join(base, "backups"))
	if err = os.MkdirAll(m.GetWorkingDir(srv.ID), 0700); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(m.GetWorkingDir(srv.ID), "sparse"))
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Truncate(16 << 20); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if _, err = m.CreateServerBackupContext(context.Background(), srv.ID, "rejected.tar.gz"); !errors.Is(err, ErrArchiveLimit) {
		t.Fatal(err)
	}
	if err = os.Remove(filepath.Join(m.GetWorkingDir(srv.ID), "sparse")); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(m.GetWorkingDir(srv.ID), "small"), []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = m.CreateServerBackupContext(context.Background(), srv.ID, "accepted.tar.gz"); err != nil {
		t.Fatal(err)
	}
}

func TestArchiveStreamGrowthAndCancellation(t *testing.T) {
	for _, mode := range []string{"grow-before", "grow-during", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			f, err := os.CreateTemp(t.TempDir(), "source")
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			data := bytes.Repeat([]byte("x"), 64<<10)
			if _, err = f.Write(data); err != nil {
				t.Fatal(err)
			}
			f.Seek(0, 0)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if mode == "grow-before" {
				if err = f.Truncate(int64(len(data) + 1)); err != nil {
					t.Fatal(err)
				}
			}
			copied := 0
			writer := archiveWriteFunc(func(p []byte) (int, error) {
				copied += len(p)
				if mode == "cancel" {
					cancel()
				}
				if mode == "grow-during" {
					if err = f.Truncate(int64(len(data) + 1)); err != nil {
						t.Fatal(err)
					}
				}
				return len(p), nil
			})
			err = CopyArchiveFile(ctx, writer, f, int64(len(data)))
			expected := ErrArchiveLimit
			if mode == "cancel" {
				expected = context.Canceled
			}
			if !errors.Is(err, expected) {
				t.Fatal(err)
			}
			if copied > len(data) || mode == "grow-before" && copied != 0 || mode == "cancel" && copied > 32<<10 {
				t.Fatal("unbounded read", copied)
			}
		})
	}
}

func TestArchiveEntryAndDepthBudgets(t *testing.T) {
	dir := t.TempDir()
	root, err := rooted.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	b := NewArchiveBudget(10)
	b.Entries = ArchiveMaxEntries
	if err = b.Walk(context.Background(), root, ".", func(string, os.FileInfo) error { t.Fatal("entry callback after budget"); return nil }); !errors.Is(err, ErrArchiveLimit) {
		t.Fatal(err)
	}
	// Depth is checked before metadata I/O, even for a nonexistent deep path.
	deep := strings.TrimSuffix(strings.Repeat("a/", ArchiveMaxDepth+1), "/")
	if err = NewArchiveBudget(10).Walk(context.Background(), root, deep, func(string, os.FileInfo) error { return nil }); !errors.Is(err, ErrArchiveLimit) {
		t.Fatal(err)
	}
	if ArchiveByteLimit(1) != 1<<20 || ArchiveByteLimit(1<<62) != ArchiveMaxBytes || ArchiveByteLimit(0) != ArchiveMaxBytes {
		t.Fatal("limit overflow/default")
	}
}

func TestArchiveConcurrencyCancellationAndShutdown(t *testing.T) {
	m := NewServerManager(t.TempDir(), nil)
	defer m.ShutdownAll()
	ctx, _, release, err := m.BeginArchiveCreation(context.Background(), "one")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err = m.BeginArchiveCreation(context.Background(), "one"); !errors.Is(err, ErrServerOperation) {
		t.Fatal(err)
	}
	_, _, release2, err := m.BeginArchiveCreation(context.Background(), "two")
	if err != nil {
		release()
		t.Fatal(err)
	}
	if _, _, _, err = m.BeginArchiveCreation(context.Background(), "three"); !errors.Is(err, ErrArchiveBusy) {
		t.Fatal(err)
	}
	release2()
	// Saturation did not leak the failed server gate or mutation lease.
	_, _, release3, err := m.BeginArchiveCreation(context.Background(), "three")
	if err != nil {
		t.Fatal(err)
	}
	release3()
	stopped := make(chan struct{})
	go func() { m.ShutdownAll(); close(stopped) }()
	select {
	case <-ctx.Done():
	case <-time.After(3 * time.Second):
		t.Fatal("shutdown did not cancel archive")
	}
	select {
	case <-stopped:
		t.Fatal("shutdown failed to join archive")
	default:
	}
	release()
	select {
	case <-stopped:
	case <-time.After(3 * time.Second):
		t.Fatal("shutdown did not finish")
	}
}

func TestBackupCanceledNoPublicationOrGateLeak(t *testing.T) {
	m, srv, dir := restoreFixture(t)
	m.SetBackupDirectory(filepath.Dir(dir))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := m.CreateServerBackupContext(ctx, srv.ID, "cancel.tar.gz"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "cancel.tar.gz")); !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := m.CreateServerBackupContext(context.Background(), srv.ID, "ok.tar.gz"); err != nil {
		t.Fatal(err)
	}
	// Cancellation during compression is propagated (not just before the call).
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	err := writeBackupArchiveContext(ctx, m.GetWorkingDir(srv.ID), archiveWriteFunc(func(p []byte) (int, error) { cancel(); return len(p), nil }), ArchiveMaxBytes)
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
