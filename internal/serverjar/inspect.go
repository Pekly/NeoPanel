package serverjar

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

const fileLimit int64 = 1 << 30

var (
	errArchive  = errors.New("invalid or unsupported server JAR archive")
	errLimit    = errors.New("server JAR inspection resource limit exceeded")
	errMetadata = errors.New("invalid or ambiguous embedded server metadata")
	errChanged  = errors.New("server JAR changed during inspection")
)

// Inspect reads only the exact configured basename, using a single confined file
// descriptor for both its digest and ZIP evidence. It never runs or extracts code.
// Unknown valid archives still have a digest and file facts. Partial means some
// recognized evidence exists but metadata was missing, unsupported, or invalid.
// RequiredJava is explicit runtime metadata; MinimumJava is only the entry point's
// bytecode floor, not a supported-runtime recommendation. Callers set deadlines.
func Inspect(ctx context.Context, dir, filename string) (Info, error) {
	info := Info{Status: "error", Source: "embedded"}
	fail := func(err error) (Info, error) {
		info.Message = err.Error()
		return info, err
	}
	if !rooted.Basename(filename) || len(filename) > 255 || strings.TrimSpace(filename) != filename || strings.HasSuffix(filename, ".") || strings.IndexFunc(filename, unicode.IsControl) >= 0 {
		return fail(errors.New("invalid server JAR filename"))
	}
	info.Filename = filename
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	root, err := rooted.Open(dir)
	if err != nil {
		return fail(errors.New("cannot open server JAR directory"))
	}
	defer root.Close()
	f, err := root.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			info.Status = "missing"
			return fail(errors.New("server JAR is missing"))
		}
		return fail(errors.New("cannot open regular server JAR"))
	}
	defer f.Close()
	before, err := f.Stat()
	if err != nil {
		return fail(errors.New("cannot inspect server JAR"))
	}
	info.Size, info.Modified = before.Size(), before.ModTime()
	if before.Size() < 0 || before.Size() > fileLimit {
		return fail(errLimit)
	}
	h := sha256.New()
	n, err := io.Copy(h, io.NewSectionReader(contextAt{ctx, f}, 0, before.Size()))
	if err != nil {
		if ctx.Err() != nil {
			return fail(ctx.Err())
		}
		return fail(errors.New("cannot read server JAR"))
	}
	if n != before.Size() {
		return fail(errChanged)
	}
	info.SHA256 = hex.EncodeToString(h.Sum(nil))
	budget := &inspectionBudget{}
	evidence, inspectErr := inspectArchive(ctx, contextAt{ctx, f}, before.Size(), budget, 0)
	// Check both the still-open object and the current directory entry. Lstat
	// rejects a replacement symlink, even one pointing back at the open object.
	if err := stableFile(root, f, filename, before); err != nil {
		info.SHA256 = ""
		return fail(err)
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if inspectErr != nil {
		return fail(inspectErr)
	}
	evidence.Filename, evidence.Size, evidence.Modified = info.Filename, info.Size, info.Modified
	evidence.SHA256, evidence.Source = info.SHA256, "embedded"
	return evidence, nil
}

func stableFile(root *rooted.Root, f *os.File, filename string, before os.FileInfo) error {
	after, err := f.Stat()
	if err != nil || !sameFileState(before, after) {
		return errChanged
	}
	current, err := root.Lstat(filename)
	if err != nil || !current.Mode().IsRegular() || !sameFileState(before, current) {
		return errChanged
	}
	return nil
}

func sameFileState(a, b os.FileInfo) bool {
	return b != nil && os.SameFile(a, b) && a.Size() == b.Size() && a.ModTime().Equal(b.ModTime()) && a.Mode() == b.Mode()
}

type contextAt struct {
	ctx context.Context
	io.ReaderAt
}

func (r contextAt) ReadAt(p []byte, off int64) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.ReaderAt.ReadAt(p, off)
}

type contextReader struct {
	ctx context.Context
	io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.Reader.Read(p)
}
