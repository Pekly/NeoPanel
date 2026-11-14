package api

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	zipPreviewMaxSize      = 256 << 20
	zipPreviewMaxDirectory = 8 << 20
	zipPreviewMaxEntries   = 10000
	zipPreviewEntryLimit   = 1 << 20
)

type zipPreviewEntry struct {
	Index          int    `json:"index"`
	Name           string `json:"name"`
	IsDir          bool   `json:"is_dir"`
	Size           uint64 `json:"size"`
	CompressedSize uint64 `json:"compressed_size"`
}

// Pin the validated central directory in memory. A concurrently modified archive
// cannot replace it with an unbounded directory between preflight and zip.NewReader.
type previewZIPReader struct {
	f         *os.File
	directory []byte
	offset    int64
}

func (r *previewZIPReader) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 {
		return 0, fmt.Errorf("Invalid ZIP offset")
	}
	n := 0
	if off < r.offset {
		count := int64(len(p))
		if count > r.offset-off {
			count = r.offset - off
		}
		read, err := r.f.ReadAt(p[:int(count)], off)
		n += read
		off += int64(read)
		p = p[read:]
		if err != nil {
			return n, err
		}
	}
	if len(p) > 0 {
		pos := off - r.offset
		if pos >= int64(len(r.directory)) {
			return n, io.EOF
		}
		read := copy(p, r.directory[pos:])
		n += read
		if read < len(p) {
			return n, io.EOF
		}
	}
	return n, nil
}

func safeZIPName(name string) bool {
	if name == "" || len(name) > 4096 || !utf8.ValidString(name) || strings.ContainsAny(name, "\\:\x00") || strings.HasPrefix(name, "/") {
		return false
	}
	for _, c := range name {
		if c < 32 || c == 127 {
			return false
		}
	}
	for _, part := range strings.Split(name, "/") {
		if part == ".." {
			return false
		}
	}
	return path.Clean(name) != "."
}

func boundedPreviewZIP(f *os.File, size int64) (*zip.Reader, error) {
	if size > zipPreviewMaxSize {
		return nil, fmt.Errorf("ZIP exceeds 256 MiB archive limit")
	}
	if size < 22 {
		return nil, fmt.Errorf("Invalid ZIP archive")
	}
	tailSize := int64(22 + 65535)
	if size < tailSize {
		tailSize = size
	}
	tail := make([]byte, int(tailSize))
	if _, err := f.ReadAt(tail, size-tailSize); err != nil {
		return nil, err
	}
	end := -1
	for i := len(tail) - 22; i >= 0; i-- {
		if string(tail[i:i+4]) == "PK\x05\x06" && i+22+int(binary.LittleEndian.Uint16(tail[i+20:i+22])) == len(tail) {
			end = i
			break
		}
	}
	if end < 0 {
		return nil, fmt.Errorf("Invalid ZIP end of directory (trailing data is not supported)")
	}
	e := tail[end:]
	u16 := func(i int) uint16 { return binary.LittleEndian.Uint16(e[i : i+2]) }
	u32 := func(i int) uint32 { return binary.LittleEndian.Uint32(e[i : i+4]) }
	count, cdSize, cdOffset := int(u16(10)), int64(u32(12)), int64(u32(16))
	if count == 65535 || cdSize == 0xffffffff || cdOffset == 0xffffffff || end >= 20 && string(tail[end-20:end-16]) == "PK\x06\x07" {
		return nil, fmt.Errorf("ZIP64 archives are not supported for preview")
	}
	if u16(4) != 0 || u16(6) != 0 || int(u16(8)) != count {
		return nil, fmt.Errorf("Multi-disk ZIP archives are not supported")
	}
	if count > zipPreviewMaxEntries {
		return nil, fmt.Errorf("ZIP exceeds 10000 entry limit")
	}
	if cdSize > zipPreviewMaxDirectory {
		return nil, fmt.Errorf("ZIP central directory exceeds 8 MiB limit")
	}
	endOffset := size - tailSize + int64(end)
	if cdOffset+cdSize != endOffset {
		return nil, fmt.Errorf("Invalid ZIP directory offsets (ZIP64 and prefixed archives are not supported)")
	}
	directory := make([]byte, int(size-cdOffset)) // <= 8 MiB + 65557 bytes
	if _, err := f.ReadAt(directory, cdOffset); err != nil {
		return nil, err
	}
	// The EOCD must match the preflight too.
	if !bytes.Equal(directory[cdSize:], e) {
		return nil, fmt.Errorf("ZIP changed while reading")
	}
	central := directory[:cdSize]
	actual := 0
	for len(central) > 0 {
		actual++
		if actual > zipPreviewMaxEntries {
			return nil, fmt.Errorf("ZIP exceeds 10000 entry limit")
		}
		if len(central) < 46 || string(central[:4]) != "PK\x01\x02" {
			return nil, fmt.Errorf("Corrupt ZIP central directory")
		}
		u16c := func(i int) uint16 { return binary.LittleEndian.Uint16(central[i : i+2]) }
		u32c := func(i int) uint32 { return binary.LittleEndian.Uint32(central[i : i+4]) }
		if u16c(8)&(1|64|8192) != 0 {
			return nil, fmt.Errorf("Encrypted ZIP archives are not supported")
		}
		if method := u16c(10); method != zip.Store && method != zip.Deflate {
			return nil, fmt.Errorf("Unsupported ZIP compression method")
		}
		if u16c(34) != 0 || u32c(20) == 0xffffffff || u32c(24) == 0xffffffff || u32c(42) == 0xffffffff {
			return nil, fmt.Errorf("ZIP64 and multi-disk entries are not supported")
		}
		nameLen, extraLen, commentLen := int(u16c(28)), int(u16c(30)), int(u16c(32))
		length := 46 + nameLen + extraLen + commentLen
		if length > len(central) {
			return nil, fmt.Errorf("Truncated ZIP directory entry")
		}
		if !safeZIPName(string(central[46 : 46+nameLen])) {
			return nil, fmt.Errorf("Unsafe ZIP entry path")
		}
		extra := central[46+nameLen : 46+nameLen+extraLen]
		for len(extra) > 0 {
			if len(extra) < 4 {
				return nil, fmt.Errorf("Corrupt ZIP extra data")
			}
			id, n := binary.LittleEndian.Uint16(extra), int(binary.LittleEndian.Uint16(extra[2:]))
			if id == 1 {
				return nil, fmt.Errorf("ZIP64 entries are not supported")
			}
			if n > len(extra)-4 {
				return nil, fmt.Errorf("Corrupt ZIP extra data")
			}
			extra = extra[4+n:]
		}
		central = central[length:]
	}
	if actual != count {
		return nil, fmt.Errorf("ZIP entry count mismatch")
	}
	reader, err := zip.NewReader(&previewZIPReader{f, directory, cdOffset}, size)
	if err != nil {
		return nil, fmt.Errorf("Invalid ZIP: %w", err)
	}
	if len(reader.File) != count {
		return nil, fmt.Errorf("ZIP entry count mismatch")
	}
	for _, entry := range reader.File {
		if !safeZIPName(entry.Name) || entry.Mode()&os.ModeSymlink != 0 || !entry.Mode().IsRegular() && !entry.Mode().IsDir() {
			return nil, fmt.Errorf("Unsafe ZIP entry (links and special files are not supported)")
		}
		offset, err := entry.DataOffset()
		if err != nil || offset < 0 || offset > cdOffset || entry.CompressedSize64 > uint64(cdOffset-offset) {
			return nil, fmt.Errorf("Corrupt ZIP entry offsets")
		}
	}
	return reader, nil
}

func (h *FileHandler) ZipPreview(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Query().Get("path")
	f, info, err := h.openPreviewFile(getURLParam(r, "id"), p, os.O_RDONLY)
	if err != nil {
		writeError(w, 400, "Cannot open ZIP: "+err.Error())
		return
	}
	defer f.Close()
	z, err := boundedPreviewZIP(f, info.Size())
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	entries := make([]zipPreviewEntry, 0, len(z.File))
	for i, e := range z.File {
		entries = append(entries, zipPreviewEntry{i, e.Name, e.FileInfo().IsDir(), e.UncompressedSize64, e.CompressedSize64})
	}
	writeJSON(w, 200, struct {
		Path       string            `json:"path"`
		Entries    []zipPreviewEntry `json:"entries"`
		EntryCount int               `json:"entry_count"`
	}{p, entries, len(entries)})
}

func (h *FileHandler) ZipEntryPreview(w http.ResponseWriter, r *http.Request) {
	index, err := strconv.Atoi(r.URL.Query().Get("index"))
	if err != nil || index < 0 {
		writeError(w, 400, "Invalid ZIP entry index")
		return
	}
	f, info, err := h.openPreviewFile(getURLParam(r, "id"), r.URL.Query().Get("path"), os.O_RDONLY)
	if err != nil {
		writeError(w, 400, "Cannot open ZIP: "+err.Error())
		return
	}
	defer f.Close()
	z, err := boundedPreviewZIP(f, info.Size())
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	if index >= len(z.File) {
		writeError(w, 400, "Invalid ZIP entry index")
		return
	}
	entry := z.File[index]
	if entry.FileInfo().IsDir() {
		writeError(w, 415, "ZIP directories cannot be opened as text")
		return
	}
	if entry.UncompressedSize64 > zipPreviewEntryLimit {
		writeError(w, 413, "ZIP entry exceeds 1 MiB expanded text limit")
		return
	}
	source, err := entry.Open()
	if err != nil {
		writeError(w, 400, "Cannot read ZIP entry: "+err.Error())
		return
	}
	defer source.Close()
	data, err := readFileBounded(source, zipPreviewEntryLimit)
	// Reading through EOF validates the ZIP CRC, including exactly-limit entries.
	if err != nil || uint64(len(data)) != entry.UncompressedSize64 {
		writeError(w, 400, "Corrupt ZIP entry or expanded size limit exceeded")
		return
	}
	if !editableFileText(entry.Name, data) {
		writeError(w, 415, "ZIP entry is not UTF-8 text")
		return
	}
	writeJSON(w, 200, map[string]string{"name": entry.Name, "content": string(data)})
}
