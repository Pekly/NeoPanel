package rooted

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type ArchiveLimits struct {
	Entries                        int
	Compressed, Expanded, Metadata int64
}

var DefaultArchiveLimits = ArchiveLimits{10000, 256 << 20, 1 << 30, 8 << 20}

func archiveName(s string) (string, error) {
	if s == "" || len(s) > 4096 || !utf8.ValidString(s) || strings.ContainsAny(s, "\\:\x00") || strings.HasPrefix(s, "/") || strings.Count(s, "/") > 64 {
		return "", fmt.Errorf("unsafe archive name")
	}
	for _, c := range s {
		if c < 32 || c == 127 {
			return "", fmt.Errorf("unsafe archive name")
		}
	}
	for _, c := range strings.Split(s, "/") {
		if c == ".." {
			return "", fmt.Errorf("archive traversal is not supported")
		}
	}
	s = path.Clean(s)
	if s == "." {
		return "", fmt.Errorf("empty archive member")
	}
	return filepath.FromSlash(s), nil
}

// Extract validates and stages the complete archive before touching live files.
// Publication is confined and per-file atomic, not a transaction over the tree.
func Extract(ctx context.Context, root *Root, f *os.File, filename, dest string, l ArchiveLimits) (int, error) {
	if l.Entries < 1 || l.Compressed < 1 || l.Expanded < 1 || l.Metadata < 1 {
		return 0, fmt.Errorf("invalid archive limits")
	}
	if e := ctx.Err(); e != nil {
		return 0, e
	}
	st, e := f.Stat()
	if e != nil {
		return 0, e
	}
	if !st.Mode().IsRegular() || st.Size() > l.Compressed {
		return 0, fmt.Errorf("archive exceeds compressed size limit")
	}
	stageName, stage, e := root.TempDir()
	if e != nil {
		return 0, e
	}
	defer root.RemoveAll(stageName)
	defer stage.Close()
	type member struct {
		name string
		dir  bool
		mode os.FileMode
	}
	var members []member
	seen := map[string]bool{}
	var total int64
	count := 0
	add := func(name string, mode os.FileMode, size int64, in io.Reader) error {
		if e := ctx.Err(); e != nil {
			return e
		}
		if len(members) >= l.Entries {
			return fmt.Errorf("archive exceeds entry limit")
		}
		n, e := archiveName(name)
		if e != nil {
			return e
		}
		key := strings.ToLower(n)
		if seen[key] {
			return fmt.Errorf("duplicate archive member")
		}
		seen[key] = true
		if !mode.IsDir() && !mode.IsRegular() {
			return fmt.Errorf("unsupported archive entry: links and special files are forbidden")
		}
		if e := root.Check(filepath.Join(dest, n)); e != nil {
			return e
		}
		members = append(members, member{n, mode.IsDir(), mode.Perm()})
		if mode.IsDir() {
			return stage.MkdirAll(n, 0755)
		}
		if size < 0 || size > l.Expanded-total {
			return fmt.Errorf("archive exceeds expanded byte limit")
		}
		if e := stage.MkdirAll(filepath.Dir(n), 0755); e != nil {
			return e
		}
		got, e := stage.Copy(ctx, n, in, l.Expanded-total, 0644)
		total += got
		if e != nil {
			return e
		}
		if got != size {
			return fmt.Errorf("archive member size mismatch")
		}
		count++
		return nil
	}
	lower := strings.ToLower(filename)
	if strings.HasSuffix(lower, ".tar") || strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") {
		if _, e = f.Seek(0, io.SeekStart); e != nil {
			return 0, e
		}
		var input io.Reader = &archiveBudget{r: f, left: l.Compressed}
		if !strings.HasSuffix(lower, ".tar") {
			gz, e := gzip.NewReader(input)
			if e != nil {
				return 0, e
			}
			defer gz.Close()
			input = gz
		}
		// Includes headers, padding and hidden PAX/GNU metadata consumed by Next.
		budget := &archiveBudget{r: contextReader{ctx, input}, left: l.Expanded + l.Metadata}
		tr := tar.NewReader(budget)
		for {
			h, e := tr.Next()
			if e == io.EOF {
				break
			}
			if e != nil {
				return 0, e
			}
			if h.Typeflag != tar.TypeDir && h.Typeflag != tar.TypeReg && h.Typeflag != tar.TypeRegA {
				return 0, fmt.Errorf("unsupported TAR entry type")
			}
			if e = add(h.Name, h.FileInfo().Mode(), h.Size, tr); e != nil {
				return 0, e
			}
		}
	} else {
		zr, e := boundedZIP(f, st.Size(), l)
		if e != nil {
			return 0, e
		}
		for _, z := range zr.File {
			rc, e := z.Open()
			if e != nil {
				return 0, e
			}
			e = add(z.Name, z.Mode(), int64(z.UncompressedSize64), rc)
			ce := rc.Close()
			if e != nil {
				return 0, e
			}
			if ce != nil {
				return 0, ce
			}
		}
	}
	// Preflight existing destinations before publication; Root handles concurrent swaps.
	for _, m := range members {
		if e := root.Check(filepath.Join(dest, m.name)); e != nil {
			return 0, e
		}
	}
	remaining := l.Expanded
	for _, m := range members {
		if e := ctx.Err(); e != nil {
			return 0, e
		}
		target := filepath.Join(dest, m.name)
		if m.dir {
			if e := root.MkdirAll(target, 0755); e != nil {
				return 0, e
			}
			continue
		}
		if e := root.MkdirAll(filepath.Dir(target), 0755); e != nil {
			return 0, e
		}
		src, e := stage.Open(m.name)
		if e != nil {
			return 0, e
		}
		n, copyErr := root.Copy(ctx, target, src, remaining, m.mode|0600)
		remaining -= n
		e = copyErr
		src.Close()
		if e != nil {
			return 0, e
		}
	}
	return count, nil
}

type archiveBudget struct {
	r    io.Reader
	left int64
}

func (b *archiveBudget) Read(p []byte) (int, error) {
	if b.left <= 0 {
		return 0, fmt.Errorf("archive stream exceeds byte budget")
	}
	if int64(len(p)) > b.left {
		p = p[:b.left]
	}
	n, e := b.r.Read(p)
	b.left -= int64(n)
	return n, e
}

// Snapshot central metadata so concurrent archive writes cannot bypass preflight.
type zipSnapshot struct {
	f      *os.File
	tail   []byte
	offset int64
}

func (s *zipSnapshot) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 {
		return 0, fmt.Errorf("negative ZIP offset")
	}
	n := 0
	if off < s.offset {
		take := int64(len(p))
		if take > s.offset-off {
			take = s.offset - off
		}
		k, e := s.f.ReadAt(p[:take], off)
		n += k
		off += int64(k)
		p = p[k:]
		if e != nil {
			return n, e
		}
	}
	if len(p) > 0 {
		pos := off - s.offset
		if pos >= int64(len(s.tail)) {
			return n, io.EOF
		}
		k := copy(p, s.tail[pos:])
		n += k
		if k < len(p) {
			return n, io.EOF
		}
	}
	return n, nil
}
func boundedZIP(f *os.File, size int64, l ArchiveLimits) (*zip.Reader, error) {
	bad := func() (*zip.Reader, error) {
		return nil, fmt.Errorf("invalid or unsupported ZIP metadata (ZIP64/multidisk/encryption unsupported)")
	}
	if size < 22 {
		return bad()
	}
	n := int64(65557)
	if n > size {
		n = size
	}
	tail := make([]byte, n)
	if _, e := f.ReadAt(tail, size-n); e != nil {
		return nil, e
	}
	end := -1
	for i := len(tail) - 22; i >= 0; i-- {
		if string(tail[i:i+4]) == "PK\x05\x06" && i+22+int(binary.LittleEndian.Uint16(tail[i+20:])) == len(tail) {
			end = i
			break
		}
	}
	if end < 0 {
		return bad()
	}
	eocd := tail[end:]
	u16 := func(i int) int { return int(binary.LittleEndian.Uint16(eocd[i:])) }
	u32 := func(i int) int64 { return int64(binary.LittleEndian.Uint32(eocd[i:])) }
	count, cdSize, off := u16(10), u32(12), u32(16)
	if count == 65535 || count > l.Entries || cdSize > l.Metadata || u16(4) != 0 || u16(6) != 0 || u16(8) != count || off+cdSize != size-n+int64(end) {
		return bad()
	}
	directory := make([]byte, size-off)
	if _, e := f.ReadAt(directory, off); e != nil {
		return nil, e
	}
	if !bytes.Equal(directory[cdSize:], eocd) {
		return bad()
	}
	central := directory[:cdSize]
	actual := 0
	for len(central) > 0 {
		actual++
		if actual > l.Entries || len(central) < 46 || string(central[:4]) != "PK\x01\x02" {
			return bad()
		}
		c16 := func(i int) int { return int(binary.LittleEndian.Uint16(central[i:])) }
		c32 := func(i int) uint32 { return binary.LittleEndian.Uint32(central[i:]) }
		if c16(8)&(1|64|8192) != 0 || c16(34) != 0 || c32(20) == 0xffffffff || c32(24) == 0xffffffff || c32(42) == 0xffffffff {
			return bad()
		}
		length := 46 + c16(28) + c16(30) + c16(32)
		if length > len(central) {
			return bad()
		}
		extra := central[46+c16(28) : 46+c16(28)+c16(30)]
		for len(extra) > 0 {
			if len(extra) < 4 {
				return bad()
			}
			id := binary.LittleEndian.Uint16(extra)
			n := int(binary.LittleEndian.Uint16(extra[2:]))
			if id == 1 || n > len(extra)-4 {
				return bad()
			}
			extra = extra[4+n:]
		}
		central = central[length:]
	}
	if actual != count {
		return bad()
	}
	zr, e := zip.NewReader(&zipSnapshot{f, directory, off}, size)
	if e != nil {
		return nil, e
	}
	if len(zr.File) != count {
		return bad()
	}
	for _, z := range zr.File {
		data, e := z.DataOffset()
		if e != nil || data < 0 || data > off || z.CompressedSize64 > uint64(off-data) {
			return bad()
		}
	}
	return zr, nil
}
