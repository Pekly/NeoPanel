package runner

import (
	"archive/zip"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// Validate Windows semantics even in non-Windows tests. In particular, validate
// the first segment BEFORE stripping the vendor's top-level runtime directory.
func windowsJavaZIPName(name string) (string, string, error) {
	bad := func() (string, string, error) { return "", "", fmt.Errorf("unsafe Windows Java ZIP path %q", name) }
	if len(name) == 0 || len(name) > 4096 || !utf8.ValidString(name) || strings.ContainsAny(name, "\\:*?\"<>|\x00") || strings.HasPrefix(name, "/") || strings.Count(name, "/") > 64 {
		return bad()
	}
	for _, r := range name {
		if r < 32 || r == 127 {
			return bad()
		}
	}
	parts := strings.Split(strings.TrimSuffix(name, "/"), "/")
	for _, p := range parts {
		if p == "" || p == "." || p == ".." || strings.HasSuffix(p, " ") || strings.HasSuffix(p, ".") {
			return bad()
		}
		base := strings.TrimRight(strings.ToUpper(strings.SplitN(p, ".", 2)[0]), " ")
		if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" || base == "CONIN$" || base == "CONOUT$" {
			return bad()
		}
		if strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT") {
			suffix := base[3:]
			if utf8.RuneCountInString(suffix) == 1 && strings.Contains("123456789¹²³", suffix) {
				return bad()
			}
		}
	}
	return parts[0], filepath.FromSlash(strings.Join(parts[1:], "/")), nil
}

// Preflight the bounded central directory before archive/zip allocates entries.
// ZIP64, multi-disk, encrypted entries and self-extracting ZIPs are unnecessary
// for these <512 MiB official runtime archives and are deliberately unsupported.
func boundedJavaZIP(f *os.File, size int64) (*zip.Reader, error) {
	bad := func() (*zip.Reader, error) { return nil, fmt.Errorf("invalid or excessive Java ZIP metadata") }
	if size < 22 || size > javaDownloadLimit {
		return bad()
	}
	n := min(size, int64(65557))
	tail := make([]byte, n)
	if _, err := f.ReadAt(tail, size-n); err != nil {
		return nil, err
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
	if count == 0 || count > javaEntryLimit || cdSize > javaZIPMetadataLimit || u16(4) != 0 || u16(6) != 0 || u16(8) != count || off+cdSize != size-n+int64(end) {
		return bad()
	}
	central := make([]byte, cdSize)
	if _, err := f.ReadAt(central, off); err != nil {
		return nil, err
	}
	actual := 0
	for len(central) > 0 {
		actual++
		if actual > javaEntryLimit || len(central) < 46 || string(central[:4]) != "PK\x01\x02" {
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
			id, n := binary.LittleEndian.Uint16(extra), int(binary.LittleEndian.Uint16(extra[2:]))
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
	zr, err := zip.NewReader(f, size)
	if err != nil {
		return nil, err
	}
	if len(zr.File) != count {
		return bad()
	}
	for _, z := range zr.File {
		data, err := z.DataOffset()
		if err != nil || data < 0 || data > off || z.CompressedSize64 > uint64(off-data) {
			return bad()
		}
	}
	return zr, nil
}

func extractWindowsJavaZIP(ctx context.Context, archive, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return err
	}
	zr, err := boundedJavaZIP(f, st.Size())
	if err != nil {
		return err
	}
	root, err := os.OpenRoot(dest)
	if err != nil {
		return err
	}
	defer root.Close()
	top := ""
	seen := map[string]bool{}
	var total uint64
	// Preflight all names, types and declared sizes before writing any file.
	for _, z := range zr.File {
		head, rel, err := windowsJavaZIPName(z.Name)
		if err != nil {
			return err
		}
		if top == "" {
			top = head
		}
		if head != top {
			return fmt.Errorf("Java ZIP has multiple top-level directories")
		}
		if kind := z.Mode() & os.ModeType; kind != 0 && kind != os.ModeDir {
			return fmt.Errorf("Java ZIP links and special files are forbidden")
		}
		if rel == "" && !z.Mode().IsDir() {
			return fmt.Errorf("Java ZIP file lacks runtime directory")
		}
		key := strings.ToLower(rel)
		if seen[key] {
			return fmt.Errorf("duplicate Java ZIP member %q", z.Name)
		}
		seen[key] = true
		if z.UncompressedSize64 > javaDownloadLimit || z.UncompressedSize64 > javaExpandedLimit-total {
			return fmt.Errorf("Java ZIP expanded size exceeds limit")
		}
		total += z.UncompressedSize64
	}
	for _, z := range zr.File {
		if err := ctx.Err(); err != nil {
			return err
		}
		_, rel, _ := windowsJavaZIPName(z.Name)
		if rel == "" {
			continue
		}
		if z.Mode().IsDir() {
			if err := root.MkdirAll(rel, 0700); err != nil {
				return err
			}
			continue
		}
		if err := root.MkdirAll(filepath.Dir(rel), 0700); err != nil {
			return err
		}
		in, err := z.Open()
		if err != nil {
			return err
		}
		out, err := root.OpenFile(rel, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			in.Close()
			return err
		}
		n, copyErr := io.Copy(out, io.LimitReader(javaContextReader{ctx, in}, int64(z.UncompressedSize64)+1))
		in.Close()
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		if uint64(n) != z.UncompressedSize64 {
			return fmt.Errorf("Java ZIP member size mismatch")
		}
	}
	return nil
}

type javaContextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r javaContextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}
