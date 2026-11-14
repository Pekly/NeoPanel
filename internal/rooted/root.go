// Package rooted provides descriptor-confined regular-file I/O for tenant trees.
package rooted

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Root struct{ *os.Root }

func Open(dir string) (*Root, error) {
	r, e := os.OpenRoot(dir)
	if e != nil {
		return nil, e
	}
	return &Root{r}, nil
}
func Wrap(r *os.Root) *Root { return &Root{r} }
func Basename(s string) bool {
	return s != "" && s != "." && s != ".." && !strings.ContainsAny(s, "/\\:\x00") && filepath.IsLocal(s)
}
func (r *Root) Path(s string) (string, error) {
	s = strings.TrimLeft(strings.ReplaceAll(s, "\\", "/"), "/")
	if s == "" {
		s = "."
	}
	if !filepath.IsLocal(s) || strings.ContainsAny(s, ":\x00") {
		return "", fmt.Errorf("invalid relative path")
	}
	s = filepath.Clean(s)
	if e := r.Check(s); e != nil {
		return "", e
	}
	return s, nil
}
func (r *Root) Destructive(s string) (string, error) {
	s, e := r.Path(s)
	if e == nil && s == "." {
		e = fmt.Errorf("cannot modify server root directory")
	}
	return s, e
}

// Check retains the API's explicit symlink policy. Security does not depend on
// this precheck: all subsequent lookups are performed by the pinned os.Root.
func (r *Root) Check(s string) error {
	if !filepath.IsLocal(s) || strings.ContainsAny(s, ":\x00") {
		return fmt.Errorf("invalid relative path")
	}
	for p := filepath.Clean(s); p != "."; p = filepath.Dir(p) {
		st, e := r.Lstat(p)
		if os.IsNotExist(e) {
			continue
		}
		if e != nil {
			return e
		}
		if st.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symbolic links are not supported")
		}
	}
	return nil
}
func (r *Root) Open(name string) (*os.File, error) {
	if e := r.Check(name); e != nil {
		return nil, e
	}
	f, e := r.OpenFile(name, os.O_RDONLY|nonblock, 0)
	if e != nil {
		return nil, e
	}
	st, e := f.Stat()
	if e != nil || !st.Mode().IsRegular() {
		f.Close()
		if e == nil {
			e = fmt.Errorf("not a regular file")
		}
		return nil, e
	}
	return f, nil
}
func (r *Root) ReadFile(name string) ([]byte, error) {
	f, e := r.Open(name)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	return ReadBounded(f, 16<<20)
}
func ReadBounded(in io.Reader, limit int64) ([]byte, error) {
	b, e := io.ReadAll(io.LimitReader(in, limit+1))
	if e == nil && int64(len(b)) > limit {
		e = fmt.Errorf("file exceeds byte limit")
	}
	return b, e
}
func (r *Root) ReadDir(name string) ([]os.DirEntry, error) {
	if e := r.Check(name); e != nil {
		return nil, e
	}
	d, e := r.Root.OpenRoot(name)
	if e != nil {
		return nil, e
	}
	defer d.Close()
	f, e := d.Open(".")
	if e != nil {
		return nil, e
	}
	defer f.Close()
	names, e := f.Readdirnames(-1)
	if e != nil {
		return nil, e
	}
	sort.Strings(names)
	entries := make([]os.DirEntry, 0, len(names))
	for _, n := range names {
		st, e := d.Lstat(n)
		if e != nil {
			return nil, e
		}
		entries = append(entries, fs.FileInfoToDirEntry(st))
	}
	return entries, nil
}
func (r *Root) Sub(name string) (*Root, error) {
	if e := r.Check(name); e != nil {
		return nil, e
	}
	d, e := r.Root.OpenRoot(name)
	if e != nil {
		return nil, e
	}
	return Wrap(d), nil
}
func (r *Root) TempDir() (string, *Root, error) {
	var b [16]byte
	if _, e := rand.Read(b[:]); e != nil {
		return "", nil, e
	}
	name := ".neopanel-" + hex.EncodeToString(b[:])
	if e := r.Mkdir(name, 0700); e != nil {
		return "", nil, e
	}
	d, e := r.Sub(name)
	if e != nil {
		r.Remove(name)
	}
	return name, d, e
}

// Publish writes a fresh inode in a private same-root directory, then renames it.
// It never truncates an existing inode (which could be an external hardlink).
func (r *Root) Publish(name string, mode os.FileMode, write func(*os.File) error) error {
	if e := r.Check(name); e != nil {
		return e
	}
	if st, e := r.Lstat(name); e == nil && !st.Mode().IsRegular() {
		return fmt.Errorf("not a regular file")
	} else if e != nil && !os.IsNotExist(e) {
		return e
	}
	dir, d, e := r.TempDir()
	if e != nil {
		return e
	}
	defer r.RemoveAll(dir)
	defer d.Close()
	f, e := d.OpenFile("data", os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode.Perm())
	if e != nil {
		return e
	}
	defer f.Close()
	e = write(f)
	if ce := f.Close(); e == nil {
		e = ce
	}
	if e != nil {
		return e
	}
	if e = r.Check(name); e != nil {
		return e
	}
	return r.Rename(filepath.Join(dir, "data"), name)
}
func (r *Root) WriteFile(name string, data []byte, mode os.FileMode) error {
	return r.Publish(name, mode, func(f *os.File) error { _, e := f.Write(data); return e })
}
func (r *Root) Copy(ctx context.Context, name string, in io.Reader, limit int64, mode os.FileMode) (int64, error) {
	var n int64
	e := r.Publish(name, mode, func(f *os.File) error { var e error; n, e = Copy(ctx, f, in, limit); return e })
	return n, e
}

type contextReader struct {
	context.Context
	io.Reader
}

func (c contextReader) Read(p []byte) (int, error) {
	if e := c.Err(); e != nil {
		return 0, e
	}
	return c.Reader.Read(p)
}
func Copy(ctx context.Context, out io.Writer, in io.Reader, limit int64) (int64, error) {
	n, e := io.Copy(out, io.LimitReader(contextReader{ctx, in}, limit+1))
	if e == nil && n > limit {
		e = fmt.Errorf("expanded data exceeds byte limit")
	}
	return n, e
}
func (r *Root) Walk(name string, fn filepath.WalkFunc) error {
	var walk func(string) error
	walk = func(p string) error {
		st, e := r.Lstat(p)
		if e != nil {
			return fn(p, nil, e)
		}
		if e = fn(p, st, nil); e != nil {
			if e == filepath.SkipDir && st.IsDir() {
				return nil
			}
			return e
		}
		if !st.IsDir() {
			return nil
		}
		entries, e := r.ReadDir(p)
		if e != nil {
			return fn(p, st, e)
		}
		for _, d := range entries {
			if e := walk(filepath.Join(p, d.Name())); e != nil {
				return e
			}
		}
		return nil
	}
	return walk(name)
}
func ReadFile(dir, name string) ([]byte, error) {
	r, e := Open(dir)
	if e != nil {
		return nil, e
	}
	defer r.Close()
	return r.ReadFile(name)
}
func WriteFile(dir, name string, b []byte, m os.FileMode) error {
	r, e := Open(dir)
	if e != nil {
		return e
	}
	defer r.Close()
	return r.WriteFile(name, b, m)
}
func OpenFile(dir, name string) (*os.File, error) {
	r, e := Open(dir)
	if e != nil {
		return nil, e
	}
	defer r.Close()
	return r.Open(name)
}

// PropertyText rejects physical line injection, including Unicode line controls.
func PropertyText(s string) bool { return !strings.ContainsAny(s, "\r\n\v\f\x00\u0085\u2028\u2029") }
