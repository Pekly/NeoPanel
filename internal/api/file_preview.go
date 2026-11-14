package api

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	fileEditorLimit = 5 << 20
	fileSniffLimit  = 64 << 10
)

type filePreview struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Kind     string `json:"kind"`
	MIME     string `json:"mime"`
	Editable bool   `json:"editable"`
	Reason   string `json:"reason,omitempty"`
}

// Extensions only restrict editing; they never authorize an inline MIME type.
func binaryFileExtension(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".zip", ".jar", ".gz", ".tgz", ".bz2", ".xz", ".7z", ".rar", ".tar", ".zst",
		".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".ico", ".avif", ".heic", ".tif", ".tiff",
		".mp4", ".webm", ".ogv", ".ogg", ".mov", ".avi", ".mkv", ".mp3", ".wav", ".flac",
		".exe", ".dll", ".so", ".dylib", ".class", ".wasm", ".bin", ".db", ".sqlite", ".sqlite3",
		".dat", ".mca", ".mcr", ".nbt", ".ldb", ".sst", ".pdf", ".woff", ".woff2", ".ttf":
		return true
	}
	return false
}

// Permit whitespace and well-formed ANSI CSI escapes in logs, not arbitrary
// control bytes (including NUL, DEL and standalone ESC).
func validFileText(data []byte) bool {
	if !utf8.Valid(data) {
		return false
	}
	for _, r := range string(data) {
		if r >= 0x80 && r <= 0x9f {
			return false
		}
	}
	for i := 0; i < len(data); i++ {
		c := data[i]
		if c == 0x1b {
			if i+1 >= len(data) || data[i+1] != '[' {
				return false
			}
			i += 2
			for i < len(data) && data[i] >= 0x30 && data[i] <= 0x3f {
				i++
			}
			for i < len(data) && data[i] >= 0x20 && data[i] <= 0x2f {
				i++
			}
			if i >= len(data) || data[i] < 0x40 || data[i] > 0x7e {
				return false
			}
		} else if c < 0x20 && c != '\t' && c != '\n' && c != '\r' || c == 0x7f {
			return false
		}
	}
	return true
}

func sniffFileKind(data []byte) (string, string) {
	if bytes.HasPrefix(data, []byte("PK\x03\x04")) || bytes.HasPrefix(data, []byte("PK\x05\x06")) || bytes.HasPrefix(data, []byte("PK\x07\x08")) {
		return "zip", "application/zip"
	}
	m := http.DetectContentType(data)
	switch m {
	case "image/png", "image/jpeg", "image/gif", "image/webp", "image/bmp", "image/x-icon":
		return "image", m
	case "video/mp4":
		return "video", m
	}
	// MP4 major/compatible brands; do not label generic ISO BMFF (e.g. AVIF) video.
	if len(data) >= 16 && string(data[4:8]) == "ftyp" {
		n := int(binary.BigEndian.Uint32(data[:4]))
		if n >= 16 && n <= len(data) && n%4 == 0 {
			for i := 8; i < n; i += 4 {
				if i == 12 {
					continue
				}
				switch string(data[i : i+4]) {
				case "isom", "iso2", "mp41", "mp42", "avc1", "M4V ":
					return "video", "video/mp4"
				}
			}
		}
	}
	// EBML DocType must explicitly identify WebM, not arbitrary Matroska.
	if webMDocType(data) {
		return "video", "video/webm"
	}
	// Ogg alone may be audio. Require the first packet's Theora identifier.
	if len(data) >= 28 && string(data[:4]) == "OggS" && data[4] == 0 {
		start := 27 + int(data[26])
		if start+7 <= len(data) && bytes.Equal(data[start:start+7], []byte("\x80theora")) {
			return "video", "video/ogg"
		}
	}
	return "binary", "application/octet-stream"
}

// EBML values are length-prefixed. Only accept a DocType element within the
// bounded EBML header, not a coincidental "webm" marker elsewhere in the file.
func webMDocType(data []byte) bool {
	if !bytes.HasPrefix(data, []byte{0x1a, 0x45, 0xdf, 0xa3}) {
		return false
	}
	headerSize, n := previewEBMLSize(data[4:])
	if n == 0 || headerSize > uint64(len(data)-4-n) {
		return false
	}
	header := data[4+n : 4+n+int(headerSize)]
	for len(header) > 0 {
		_, idLen := previewEBMLSize(header)
		if idLen == 0 || idLen > 4 {
			return false
		}
		valueSize, sizeLen := previewEBMLSize(header[idLen:])
		if sizeLen == 0 || valueSize > uint64(len(header)-idLen-sizeLen) {
			return false
		}
		end := idLen + sizeLen + int(valueSize)
		if idLen == 2 && header[0] == 0x42 && header[1] == 0x82 {
			return string(header[idLen+sizeLen:end]) == "webm"
		}
		header = header[end:]
	}
	return false
}

func previewEBMLSize(data []byte) (uint64, int) {
	if len(data) == 0 || data[0] == 0 {
		return 0, 0
	}
	mask, n := byte(0x80), 1
	for data[0]&mask == 0 {
		mask >>= 1
		n++
	}
	if n > len(data) {
		return 0, 0
	}
	value := uint64(data[0] & (mask - 1))
	for _, b := range data[1:n] {
		value = value<<8 | uint64(b)
	}
	return value, n
}

func binaryFileSignature(data []byte) bool {
	for _, signature := range []string{"MZ", "%PDF-", "%!PS-Adobe-", "SQLite format 3", "\x7fELF", "\xca\xfe\xba\xbe", "\x00asm", "\x1f\x8b", "Rar!", "7z\xbc\xaf\x27\x1c"} {
		if bytes.HasPrefix(data, []byte(signature)) {
			return true
		}
	}
	return false
}

func editableFileText(name string, data []byte) bool {
	kind, _ := sniffFileKind(data)
	return kind == "binary" && !binaryFileExtension(name) && !binaryFileSignature(data) && validFileText(data)
}

func readFileBounded(r io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("File exceeds %d byte limit", limit)
	}
	return data, nil
}

// The platform open pins parent directories and refuses symlinks/reparse points.
// No truncation occurs until the opened descriptor has been validated.
func (h *FileHandler) openPreviewFile(serverID, reqPath string, flags int) (*os.File, os.FileInfo, error) {
	for _, part := range strings.Split(strings.ReplaceAll(reqPath, "\\", "/"), "/") {
		if part == ".." {
			return nil, nil, fmt.Errorf("Access denied: path traversal is not supported")
		}
	}
	full, err := h.safePath(serverID, reqPath)
	if err != nil {
		return nil, nil, err
	}
	base, err := h.safePath(serverID, "")
	if err != nil {
		return nil, nil, err
	}
	rel, err := filepath.Rel(base, full)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, nil, fmt.Errorf("Only regular files within the server directory are supported")
	}
	f, err := openPreviewPath(base, rel, flags)
	if err != nil {
		return nil, nil, err
	}
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		f.Close()
		return nil, nil, fmt.Errorf("Only regular files are supported")
	}
	return f, info, nil
}

func (h *FileHandler) Preview(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Query().Get("path")
	f, info, err := h.openPreviewFile(getURLParam(r, "id"), p, os.O_RDONLY)
	if err != nil {
		writeError(w, 400, "Cannot preview file: "+err.Error())
		return
	}
	defer f.Close()
	limit := int64(fileEditorLimit)
	if info.Size() > fileEditorLimit {
		limit = fileSniffLimit
	}
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		writeError(w, 400, "Cannot read file")
		return
	}
	kind, m := sniffFileKind(data)
	result := filePreview{Path: p, Name: info.Name(), Size: info.Size(), Kind: kind, MIME: m}
	if kind == "binary" && !binaryFileExtension(info.Name()) && !binaryFileSignature(data) {
		text := data
		// A bounded prefix may end within a UTF-8 rune. It is never editor approval.
		if info.Size() > fileEditorLimit {
			for n := 0; n < 3 && len(text) > 0 && !utf8.Valid(text); n++ {
				text = text[:len(text)-1]
			}
		}
		if validFileText(text) {
			result.Kind = "text"
			result.MIME = "text/plain; charset=utf-8"
		}
	}
	result.Editable = result.Kind == "text" && info.Size() <= fileEditorLimit && len(data) <= fileEditorLimit
	if !result.Editable {
		result.Reason = "This file is not editable text"
		if result.Kind == "text" {
			result.Reason = "File exceeds 5 MiB text editor limit"
		}
	}
	writeJSON(w, 200, result)
}

// ServeContent may remove cache headers on a range error; previews remain
// private and non-cacheable on every response, including 416.
type mediaPreviewWriter struct{ http.ResponseWriter }

func (w mediaPreviewWriter) WriteHeader(status int) {
	w.Header().Set("Cache-Control", "private, no-store")
	w.ResponseWriter.WriteHeader(status)
}

func (h *FileHandler) Media(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	f, info, err := h.openPreviewFile(getURLParam(r, "id"), r.URL.Query().Get("path"), os.O_RDONLY)
	if err != nil {
		writeError(w, 400, "Cannot preview file: "+err.Error())
		return
	}
	defer f.Close()
	data := make([]byte, fileSniffLimit)
	n, err := f.ReadAt(data, 0)
	if err != nil && err != io.EOF {
		writeError(w, 400, "Cannot read file")
		return
	}
	kind, m := sniffFileKind(data[:n])
	if kind != "image" && kind != "video" {
		writeError(w, 415, "Inline preview supports raster images and videos only")
		return
	}
	w.Header().Set("Content-Type", m)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": info.Name()}))
	http.ServeContent(mediaPreviewWriter{w}, r, "", info.ModTime(), f)
}
