package runner

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"github.com/Pekly/NeoPanel/internal/rooted"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CreateBackup publishes a complete archive atomically, without replacing an existing
// backup. Temporary files live on the destination filesystem and are always removed.
// This is a live filesystem backup, not a transactional snapshot of a running JVM.
func CreateBackup(sourceDir, backupDir, name string) (int64, error) {
	unreserve, err := reserveBackupTarget(backupDir, name)
	if err != nil {
		return 0, err
	}
	defer unreserve()
	ctx, cancel := context.WithTimeout(context.Background(), ArchiveTimeout)
	defer cancel()
	release, err := archiveSlot(ctx)
	if err != nil {
		return 0, err
	}
	defer release()
	mutation, err := rooted.TryExclusiveMutation(sourceDir)
	if err != nil {
		return 0, err
	}
	defer mutation()
	return createBackupContext(ctx, sourceDir, backupDir, name, ArchiveMaxBytes)
}

func createBackupContext(ctx context.Context, sourceDir, backupDir, name string, limit int64) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if name == "" || filepath.Base(name) != name || strings.ContainsAny(name, "/\\") || !strings.HasSuffix(name, ".tar.gz") {
		return 0, fmt.Errorf("invalid backup filename")
	}
	source, err := filepath.Abs(sourceDir)
	if err != nil {
		return 0, err
	}
	source, err = filepath.EvalSymlinks(source)
	if err != nil {
		return 0, fmt.Errorf("resolve backup source: %w", err)
	}
	dest, err := filepath.Abs(backupDir)
	if err != nil {
		return 0, err
	}
	if rel, e := filepath.Rel(source, dest); e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return 0, fmt.Errorf("backup directory must be outside the server directory")
	}
	st, err := os.Stat(source)
	if err != nil {
		return 0, fmt.Errorf("read backup source: %w", err)
	}
	if !st.IsDir() {
		return 0, fmt.Errorf("backup source is not a directory")
	}
	if err := os.MkdirAll(dest, 0755); err != nil {
		return 0, fmt.Errorf("prepare backup directory: %w", err)
	}
	dest, err = filepath.EvalSymlinks(dest)
	if err != nil {
		return 0, fmt.Errorf("resolve backup destination: %w", err)
	}
	if rel, e := filepath.Rel(source, dest); e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return 0, fmt.Errorf("backup directory must be outside the server directory")
	}
	out, err := os.CreateTemp(dest, ".backup-*.tmp")
	if err != nil {
		return 0, fmt.Errorf("create temporary backup: %w", err)
	}
	defer os.Remove(out.Name())
	archiveErr := writeBackupArchiveContext(ctx, source, out, limit)
	syncErr := out.Sync()
	closeErr := out.Close()
	if err := errors.Join(archiveErr, syncErr, closeErr); err != nil {
		return 0, fmt.Errorf("backup archive failed: %w", err)
	}
	info, err := os.Stat(out.Name())
	if err != nil {
		return 0, fmt.Errorf("stat completed backup: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	// Link is an atomic no-replace publication (Rename would overwrite on Unix).
	if err := os.Link(out.Name(), filepath.Join(dest, name)); err != nil {
		return 0, fmt.Errorf("publish backup: %w", err)
	}
	return info.Size(), nil
}

func writeBackupArchive(source string, out io.Writer) error {
	ctx, cancel := context.WithTimeout(context.Background(), ArchiveTimeout)
	defer cancel()
	return writeBackupArchiveContext(ctx, source, out, ArchiveMaxBytes)
}

func writeBackupArchiveContext(ctx context.Context, source string, out io.Writer, limit int64) error {
	root, err := rooted.Open(source)
	if err != nil {
		return err
	}
	defer root.Close()
	type entry struct {
		path string
		info os.FileInfo
		link string
	}
	var entries []entry
	err = NewArchiveBudget(limit).Walk(ctx, root, ".", func(p string, info os.FileInfo) error {
		if p == "." {
			return nil
		}
		link := ""
		if info.Mode()&os.ModeSymlink != 0 {
			var e error
			link, e = root.Readlink(p)
			if e != nil {
				return e
			}
		} else if !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported file type: %s", p)
		}
		entries = append(entries, entry{p, info, link})
		return nil
	})
	if err != nil {
		return err
	} // No file content or compressed output before preflight succeeds.
	gw := gzip.NewWriter(out)
	tw := tar.NewWriter(gw)
	writeErr := func() error {
		for _, item := range entries {
			if err := ctx.Err(); err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(item.info, item.link)
			if err != nil {
				return err
			}
			header.Name = filepath.ToSlash(item.path)
			if err = tw.WriteHeader(header); err != nil {
				return err
			}
			if item.info.Mode().IsRegular() {
				f, err := root.Open(item.path)
				if err != nil {
					return err
				}
				copyErr := CopyArchiveFile(ctx, tw, f, item.info.Size())
				if err = errors.Join(copyErr, f.Close()); err != nil {
					return fmt.Errorf("archive %s: %w", item.path, err)
				}
			}
		}
		return ctx.Err()
	}()
	return errors.Join(writeErr, tw.Close(), gw.Close(), ctx.Err())
}
