package runner

import (
	"archive/tar"
	"compress/gzip"
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
	archiveErr := writeBackupArchive(source, out)
	syncErr := out.Sync()
	closeErr := out.Close()
	if err := errors.Join(archiveErr, syncErr, closeErr); err != nil {
		return 0, fmt.Errorf("backup archive failed: %w", err)
	}
	info, err := os.Stat(out.Name())
	if err != nil {
		return 0, fmt.Errorf("stat completed backup: %w", err)
	}
	// Link is an atomic no-replace publication (Rename would overwrite on Unix).
	if err := os.Link(out.Name(), filepath.Join(dest, name)); err != nil {
		return 0, fmt.Errorf("publish backup: %w", err)
	}
	return info.Size(), nil
}

func writeBackupArchive(source string, out io.Writer) error {
	root, err := rooted.Open(source)
	if err != nil {
		return err
	}
	defer root.Close()
	gw := gzip.NewWriter(out)
	tw := tar.NewWriter(gw)
	walkErr := root.Walk(".", func(p string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if p == "." {
			return nil
		}
		rel := p
		var err error
		link := ""
		if info.Mode()&os.ModeSymlink != 0 {
			link, err = root.Readlink(p)
			if err != nil {
				return err
			}
		} else if !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported file type: %s", rel)
		}
		var f *os.File
		if info.Mode().IsRegular() {
			f, err = root.Open(p)
			if err != nil {
				return err
			}
			defer f.Close()
			info, err = f.Stat()
			if err != nil {
				return err
			}
		}
		header, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(rel)
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("archive header %s: %w", rel, err)
		}
		if info.Mode().IsRegular() {
			// Read from the same regular-file handle used for the header.
			_, copyErr := io.Copy(tw, f)
			if err := errors.Join(copyErr, f.Close()); err != nil {
				return fmt.Errorf("archive %s: %w", rel, err)
			}
		}
		return nil
	})
	// Close in order, before measuring or publishing; gzip writes its footer on Close.
	tarErr := tw.Close()
	gzipErr := gw.Close()
	return errors.Join(walkErr, tarErr, gzipErr)
}
