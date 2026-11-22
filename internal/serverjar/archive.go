package serverjar

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/binary"
	"io"
)

const (
	entryLimit         = 50000
	centralLimit       = 8 << 20
	metadataLimit      = 128 << 10
	metadataTotalLimit = 512 << 10
	nestedLimit        = 128 << 20
	expandedLimit      = 256 << 20
)

type inspectionBudget struct{ metadata, expanded uint64 }

// Preflight the real central directory before archive/zip allocates its index.
// Deliberately reject ZIP64, split archives, prefixes, trailers and ambiguous
// duplicate names. Counts in EOCD are not trusted to bound the parser's work.
func boundedZip(ctx context.Context, r io.ReaderAt, size int64) (*zip.Reader, error) {
	if size < 22 {
		return nil, errArchive
	}
	n := min(size, int64(65557))
	tail := make([]byte, n)
	if _, err := r.ReadAt(tail, size-n); err != nil {
		return nil, errArchive
	}
	for i := len(tail) - 22; i >= 0; i-- {
		if binary.LittleEndian.Uint32(tail[i:]) != 0x06054b50 || i+22+int(binary.LittleEndian.Uint16(tail[i+20:])) > len(tail) {
			continue
		}
		// archive/zip accepts a short comment with trailing bytes. Do not
		// preflight a different EOCD from the one its backward scan will use.
		if i+22+int(binary.LittleEndian.Uint16(tail[i+20:])) != len(tail) {
			return nil, errArchive
		}
		h := tail[i:]
		count := int(binary.LittleEndian.Uint16(h[10:]))
		length := int64(binary.LittleEndian.Uint32(h[12:]))
		start := int64(binary.LittleEndian.Uint32(h[16:]))
		if count > entryLimit || length > centralLimit {
			return nil, errLimit
		}
		if binary.LittleEndian.Uint16(h[4:]) != 0 || binary.LittleEndian.Uint16(h[6:]) != 0 || int(binary.LittleEndian.Uint16(h[8:])) != count || start == 0xffffffff || start+length != size-n+int64(i) {
			return nil, errArchive
		}
		central := make([]byte, length)
		if _, err := r.ReadAt(central, start); err != nil {
			return nil, errArchive
		}
		pos, first := 0, int64(-1)
		for j := 0; j < count; j++ {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			if pos+46 > len(central) {
				return nil, errArchive
			}
			c := central[pos:]
			if binary.LittleEndian.Uint32(c) != 0x02014b50 {
				return nil, errArchive
			}
			nl, el, cl := int(binary.LittleEndian.Uint16(c[28:])), int(binary.LittleEndian.Uint16(c[30:])), int(binary.LittleEndian.Uint16(c[32:]))
			end := pos + 46 + nl + el + cl
			off := int64(binary.LittleEndian.Uint32(c[42:]))
			compressed := binary.LittleEndian.Uint32(c[20:])
			if end > len(central) || binary.LittleEndian.Uint16(c[34:]) != 0 || compressed == 0xffffffff || binary.LittleEndian.Uint32(c[24:]) == 0xffffffff || off == 0xffffffff || off+30 > start || binary.LittleEndian.Uint16(c[8:])&1 != 0 {
				return nil, errArchive
			}
			for extra := c[46+nl : 46+nl+el]; len(extra) > 0; {
				if len(extra) < 4 {
					return nil, errArchive
				}
				l := int(binary.LittleEndian.Uint16(extra[2:]))
				if binary.LittleEndian.Uint16(extra) == 1 || l+4 > len(extra) {
					return nil, errArchive
				}
				extra = extra[l+4:]
			}
			var local [30]byte
			if _, err := r.ReadAt(local[:], off); err != nil || binary.LittleEndian.Uint32(local[:]) != 0x04034b50 {
				return nil, errArchive
			}
			ln, le := int(binary.LittleEndian.Uint16(local[26:])), int(binary.LittleEndian.Uint16(local[28:]))
			if ln != nl || off+30+int64(ln+le)+int64(compressed) > start || binary.LittleEndian.Uint16(local[6:]) != binary.LittleEndian.Uint16(c[8:]) || binary.LittleEndian.Uint16(local[8:]) != binary.LittleEndian.Uint16(c[10:]) {
				return nil, errArchive
			}
			name := make([]byte, ln)
			if _, err := r.ReadAt(name, off+30); err != nil || !bytes.Equal(name, c[46:46+nl]) {
				return nil, errArchive
			}
			if first < 0 || off < first {
				first = off
			}
			pos = end
		}
		if pos != len(central) || (count > 0 && first != 0) || (count == 0 && start != 0) {
			return nil, errArchive
		}
		z, err := zip.NewReader(r, size)
		if err != nil || len(z.File) != count {
			return nil, errArchive
		}
		names := make(map[string]bool, count)
		for _, f := range z.File {
			if names[f.Name] {
				return nil, errMetadata
			}
			names[f.Name] = true
		}
		return z, nil
	}
	return nil, errArchive
}

func (b *inspectionBudget) read(ctx context.Context, f *zip.File, nested bool) ([]byte, error) {
	limit := uint64(metadataLimit)
	if nested {
		limit = nestedLimit
	}
	n := f.UncompressedSize64
	if n > limit || n > expandedLimit-b.expanded {
		return nil, errLimit
	}
	if !nested {
		if n > metadataTotalLimit-b.metadata {
			return nil, errLimit
		}
		b.metadata += n
	}
	b.expanded += n
	in, err := f.Open()
	if err != nil {
		return nil, errMetadata
	}
	defer in.Close()
	data := make([]byte, int(n))
	r := contextReader{ctx, in}
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, errMetadata
	}
	var extra [1]byte
	if n, err := r.Read(extra[:]); n != 0 || err != io.EOF {
		return nil, errMetadata
	}
	return data, nil
}
