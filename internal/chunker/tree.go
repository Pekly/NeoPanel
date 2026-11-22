package chunker

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

const MaxEntries = 50000
const MaxBytes int64 = 8 << 30

// Tree walks bounded directory batches through pinned directory handles. dst is
// a fresh private directory, or nil for a fingerprint only. No links are made.
// Settings are rejected rather than silently letting input control the CLI.
func Tree(ctx context.Context, src *rooted.Root, dst *rooted.Root, limit int64) (string, int64, error) {
	h := sha256.New()
	count := 0
	var total int64
	var walk func(*rooted.Root, *rooted.Root, string, int) error
	walk = func(r, d *rooted.Root, prefix string, depth int) error {
		if depth > 64 {
			return errors.New("world nesting limit exceeded")
		}
		dir, err := r.Root.Open(".")
		if err != nil {
			return err
		}
		defer dir.Close()
		var names []string
		for {
			batch, e := dir.Readdirnames(128)
			count += len(batch)
			if count > MaxEntries {
				return errors.New("world file count limit exceeded")
			}
			names = append(names, batch...)
			if e == io.EOF {
				break
			}
			if e != nil {
				return e
			}
			if err := ctx.Err(); err != nil {
				return err
			}
		}
		sort.Strings(names)
		for _, name := range names {
			if err := ctx.Err(); err != nil {
				return err
			}
			if !rooted.Basename(name) || strings.HasSuffix(strings.ToLower(name), ".chunker.json") {
				return errors.New("unsafe world name or uploaded .chunker.json settings; remove settings before conversion")
			}
			info, err := r.Lstat(name)
			if err != nil {
				return err
			}
			if info.Mode()&os.ModeSymlink != 0 || (!info.IsDir() && !info.Mode().IsRegular()) {
				return errors.New("world must contain only real directories and regular files")
			}
			path := prefix + name
			entrySize := info.Size()
			if info.IsDir() {
				entrySize = 0
			}
			fmt.Fprintf(h, "%d:%s:%t:%d\n", len(path), path, info.IsDir(), entrySize)
			if info.IsDir() {
				child, e := r.Sub(name)
				if e != nil {
					return e
				}
				var out *rooted.Root
				if d != nil {
					if e = d.Mkdir(name, 0700); e == nil {
						out, e = d.Sub(name)
					}
					if e != nil {
						child.Close()
						return e
					}
				}
				e = walk(child, out, path+"/", depth+1)
				child.Close()
				if out != nil {
					out.Close()
				}
				if e != nil {
					return e
				}
				continue
			}
			if info.Size() < 0 || info.Size() > limit-total {
				return errors.New("world byte limit exceeded")
			}
			f, e := r.Open(name)
			if e != nil {
				return e
			}
			actual, e := f.Stat()
			if e != nil || !os.SameFile(info, actual) || info.Size() != actual.Size() {
				f.Close()
				return errors.New("world changed during read")
			}
			var output *os.File
			var w io.Writer = h
			if d != nil {
				output, e = d.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
				if e != nil {
					f.Close()
					return e
				}
				w = io.MultiWriter(h, output)
			}
			n, e := rooted.Copy(ctx, w, f, info.Size())
			after, se := f.Stat()
			f.Close()
			if output != nil {
				e = errors.Join(e, output.Close())
			}
			if e != nil {
				return e
			}
			if se != nil || n != info.Size() || after.Size() != info.Size() || !after.ModTime().Equal(info.ModTime()) {
				return errors.New("world changed during copy")
			}
			total += n
		}
		return nil
	}
	err := walk(src, dst, "", 0)
	return hex.EncodeToString(h.Sum(nil)), total, err
}

func OpenWorld(serverDir, name string) (*rooted.Root, error) {
	// Deliberately only one direct child: excludes root, aliases and overlap.
	if !rooted.Basename(name) || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "converted-") {
		return nil, errors.New("select a direct world directory, not server root or converted output")
	}
	st, err := os.Lstat(serverDir)
	if err != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("unsafe server directory")
	}
	r, err := rooted.Open(serverDir)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return r.Sub(name)
}
func Scan(ctx context.Context, dir string, limit int64) (string, int64, error) {
	r, err := rooted.Open(filepath.Clean(dir))
	if err != nil {
		return "", 0, err
	}
	defer r.Close()
	return Tree(ctx, r, nil, limit)
}
