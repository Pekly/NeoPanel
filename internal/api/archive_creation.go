package api

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Pekly/NeoPanel/internal/rooted"
	"github.com/Pekly/NeoPanel/internal/runner"
)

func writeArchiveCreationError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, runner.ErrArchiveLimit):
		status = http.StatusRequestEntityTooLarge
	case errors.Is(err, runner.ErrArchiveBusy):
		status = http.StatusTooManyRequests
		w.Header().Set("Retry-After", "5")
	case errors.Is(err, runner.ErrServerOperation), errors.Is(err, rooted.ErrMutationBusy):
		status = http.StatusConflict
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		status = http.StatusRequestTimeout
	}
	writeError(w, status, "Failed to create archive: "+err.Error())
}

func buildZipArchive(root *rooted.Root, out io.Writer, sources []archiveSource, skip map[string]bool) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), runner.ArchiveTimeout)
	defer cancel()
	return buildArchiveContext(ctx, root, out, sources, skip, "zip", runner.ArchiveMaxBytes)
}
func buildTarGzArchive(root *rooted.Root, out io.Writer, sources []archiveSource, skip map[string]bool) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), runner.ArchiveTimeout)
	defer cancel()
	return buildArchiveContext(ctx, root, out, sources, skip, "tar.gz", runner.ArchiveMaxBytes)
}

func buildArchiveContext(ctx context.Context, root *rooted.Root, out io.Writer, sources []archiveSource, skip map[string]bool, format string, limit int64) (int, error) {
	type entry struct {
		path, name string
		info       os.FileInfo
	}
	var entries []entry
	budget := runner.NewArchiveBudget(limit)
	for _, src := range sources {
		err := budget.Walk(ctx, root, src.full, func(p string, info os.FileInfo) error {
			if skip[p] || strings.HasPrefix(filepath.Base(p), ".neopanel-") {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if info.Mode()&os.ModeSymlink != 0 {
				return nil
			}
			if !info.IsDir() && !info.Mode().IsRegular() {
				return fmt.Errorf("unsupported file type: %s", p)
			}
			rel, err := filepath.Rel(src.full, p)
			if err != nil {
				return err
			}
			name := filepath.ToSlash(src.entryRoot)
			if rel != "." {
				name = filepath.ToSlash(filepath.Join(src.entryRoot, rel))
			}
			if len(strings.Split(name, "/")) > runner.ArchiveMaxDepth {
				return runner.ErrArchiveLimit
			}
			entries = append(entries, entry{p, name, info})
			return nil
		})
		if err != nil {
			return 0, err
		}
	}
	// Nothing is opened for content or sent to a compressor until ALL selections
	// pass logical input preflight. The immutable plan also bounds added files.
	var header func(entry) (io.Writer, error)
	var finish func() error
	if format == "tar.gz" {
		gw := gzip.NewWriter(out)
		tw := tar.NewWriter(gw)
		finish = func() error { return errors.Join(tw.Close(), gw.Close()) }
		header = func(e entry) (io.Writer, error) {
			h, err := tar.FileInfoHeader(e.info, "")
			if err != nil {
				return nil, err
			}
			h.Name = e.name
			if e.info.IsDir() {
				h.Name = strings.TrimSuffix(h.Name, "/") + "/"
			}
			return tw, tw.WriteHeader(h)
		}
	} else {
		zw := zip.NewWriter(out)
		finish = zw.Close
		header = func(e entry) (io.Writer, error) {
			h, err := zip.FileInfoHeader(e.info)
			if err != nil {
				return nil, err
			}
			h.Name = e.name
			if e.info.IsDir() {
				h.Name = strings.TrimSuffix(h.Name, "/") + "/"
			} else {
				h.Method = zip.Deflate
			}
			return zw.CreateHeader(h)
		}
	}
	count := 0
	writeErr := func() error {
		for _, e := range entries {
			if err := ctx.Err(); err != nil {
				return err
			}
			w, err := header(e)
			if err != nil {
				return err
			}
			if e.info.IsDir() {
				continue
			}
			f, err := root.Open(e.path)
			if err != nil {
				return err
			}
			copyErr := runner.CopyArchiveFile(ctx, w, f, e.info.Size())
			if err = errors.Join(copyErr, f.Close()); err != nil {
				return err
			}
			count++
		}
		return nil
	}()
	return count, errors.Join(writeErr, finish(), ctx.Err())
}
