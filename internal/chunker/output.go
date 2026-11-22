package chunker

import (
	"archive/tar"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

// ExtractOutput receives only the sandbox's tar stream, never an uploaded archive.
// The producer has no mount of destination. Extraction is bounded and no-replace.
func ExtractOutput(ctx context.Context, in io.Reader, dest string, limit int64) error {
	root, err := rooted.Open(dest)
	if err != nil {
		return err
	}
	defer root.Close()
	tr := tar.NewReader(io.LimitReader(in, limit+64<<20))
	var total int64
	count := 0
	for {
		if err = ctx.Err(); err != nil {
			return err
		}
		h, e := tr.Next()
		if e == io.EOF {
			return nil
		}
		if e != nil {
			return e
		}
		count++
		if count > MaxEntries {
			return errors.New("output entry limit exceeded")
		}
		name := strings.TrimSuffix(strings.TrimPrefix(h.Name, "./"), "/")
		if name == "." || name == "" {
			if h.Typeflag == tar.TypeDir {
				continue
			}
			return errors.New("invalid output root")
		}
		if !filepath.IsLocal(name) || strings.ContainsAny(name, "\\:\x00") {
			return errors.New("unsafe converter output path")
		}
		for _, part := range strings.Split(name, "/") {
			if !rooted.Basename(part) {
				return errors.New("unsafe output component")
			}
		}
		if len(strings.Split(name, "/")) > 64 {
			return errors.New("output nesting limit exceeded")
		}
		switch h.Typeflag {
		case tar.TypeDir:
			if e = root.MkdirAll(name, 0700); e != nil {
				return e
			}
		case tar.TypeReg:
			if h.Size < 0 || h.Size > limit-total {
				return errors.New("output byte limit exceeded")
			}
			if e = root.MkdirAll(filepath.Dir(name), 0700); e != nil {
				return e
			}
			f, e := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
			if e != nil {
				return e
			}
			n, e := rooted.Copy(ctx, f, tr, h.Size)
			e = errors.Join(e, f.Close())
			if e != nil {
				return e
			}
			if n != h.Size {
				return io.ErrUnexpectedEOF
			}
			total += n
		default:
			return errors.New("converter output contains a link or special file")
		}
	}
}
