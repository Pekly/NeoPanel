package api

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type previewZIPFixtureEntry struct {
	name, content string
	mode          os.FileMode
}

func previewZIPBytes(t *testing.T, entries []previewZIPFixtureEntry) []byte {
	t.Helper()
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.name, Method: zip.Deflate}
		if entry.mode != 0 {
			header.SetMode(entry.mode)
		}
		dest, err := w.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := dest.Write([]byte(entry.content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestFilePreviewZIPListingAndEntries(t *testing.T) {
	f := newReliabilityFixture(t)
	data := previewZIPBytes(t, []previewZIPFixtureEntry{
		{"config/", "", 0}, {"config/server.properties", "motd=世界\n", 0},
		{"duplicate.txt", "first", 0}, {"duplicate.txt", "second", 0},
		{"binary.txt", strings.Repeat("a", 1024) + "\x00", 0},
		{"large.txt", strings.Repeat("a", zipPreviewEntryLimit+1), 0},
		{"exact.txt", strings.Repeat("a", zipPreviewEntryLimit), 0},
	})
	previewPut(t, f, "plugin.jar", data)
	w := f.request("GET", previewURL+"zip?path=plugin.jar", "", "subuser")
	var result struct {
		Path    string            `json:"path"`
		Entries []zipPreviewEntry `json:"entries"`
		Count   int               `json:"entry_count"`
	}
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &result) != nil {
		t.Fatal(w.Code, w.Body.String())
	}
	if result.Path != "plugin.jar" || result.Count != 7 || len(result.Entries) != 7 || !result.Entries[0].IsDir || result.Entries[1].Index != 1 || result.Entries[1].Size != uint64(len("motd=世界\n")) || result.Entries[1].CompressedSize == 0 {
		t.Fatalf("bad listing: %+v", result)
	}
	for _, tc := range []struct {
		index, content string
		status         int
	}{
		{"0", "", 415}, {"1", "motd=世界\n", 200}, {"2", "first", 200}, {"3", "second", 200},
		{"4", "", 415}, {"5", "", 413}, {"6", strings.Repeat("a", zipPreviewEntryLimit), 200},
		{"-1", "", 400}, {"no", "", 400}, {"7", "", 400}, {"", "", 400},
	} {
		w := f.request("GET", previewURL+"zip-entry?path=plugin.jar&index="+tc.index, "", "subuser")
		if w.Code != tc.status {
			t.Fatalf("index=%s: %d %s", tc.index, w.Code, w.Body.String())
		}
		if tc.status == 200 {
			var entry map[string]string
			if json.Unmarshal(w.Body.Bytes(), &entry) != nil || entry["content"] != tc.content || entry["name"] == "" {
				t.Fatal("wrong entry content")
			}
		}
	}
	contents, err := os.ReadDir(f.manager.GetWorkingDir("alpha"))
	if err != nil || len(contents) != 1 || contents[0].Name() != "plugin.jar" {
		t.Fatal("ZIP preview extracted files", contents, err)
	}
	previewPut(t, f, "empty.zip", previewZIPBytes(t, nil))
	w = f.request("GET", previewURL+"zip?path=empty.zip", "", "owner")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"entries":[]`) || !strings.Contains(w.Body.String(), `"entry_count":0`) {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestFilePreviewZIPRejectsUnsafeArchives(t *testing.T) {
	f := newReliabilityFixture(t)
	for _, entry := range []previewZIPFixtureEntry{
		{"../escape.txt", "bad", 0}, {"folder/../../escape", "bad", 0}, {"/absolute", "bad", 0},
		{"C:/escape.txt", "bad", 0}, {"folder\\escape.txt", "bad", 0}, {"nul\x00.txt", "bad", 0},
		{"symlink", "../../outside", os.ModeSymlink | 0777}, {"device", "", os.ModeDevice | 0600},
	} {
		previewPut(t, f, "unsafe.zip", previewZIPBytes(t, []previewZIPFixtureEntry{entry}))
		for _, route := range []string{"zip", "zip-entry"} {
			if w := f.request("GET", previewURL+route+"?path=unsafe.zip&index=0", "", "owner"); w.Code != 400 {
				t.Fatalf("%q %s: %d %s", entry.name, route, w.Code, w.Body.String())
			}
		}
	}
	base := previewZIPBytes(t, []previewZIPFixtureEntry{{"file.txt", "text data", 0}})
	central := bytes.Index(base, []byte("PK\x01\x02"))
	end := len(base) - 22
	for _, tc := range []struct {
		name, reason string
		mutate       func([]byte)
	}{
		{"encrypted", "Encrypted", func(b []byte) { b[central+8] |= 1 }},
		{"unsupported", "Unsupported", func(b []byte) { binary.LittleEndian.PutUint16(b[central+10:], 99) }},
		{"zip64", "ZIP64", func(b []byte) { binary.LittleEndian.PutUint32(b[central+24:], 0xffffffff) }},
		{"count", "10000", func(b []byte) {
			binary.LittleEndian.PutUint16(b[end+8:], 10001)
			binary.LittleEndian.PutUint16(b[end+10:], 10001)
		}},
		{"directory-size", "8 MiB", func(b []byte) { binary.LittleEndian.PutUint32(b[end+12:], zipPreviewMaxDirectory+1) }},
		{"bad-count", "count mismatch", func(b []byte) {
			binary.LittleEndian.PutUint16(b[end+8:], 0)
			binary.LittleEndian.PutUint16(b[end+10:], 0)
		}},
		{"bad-offset", "offset", func(b []byte) { binary.LittleEndian.PutUint32(b[central+42:], 0xfffffffe) }},
		{"multi-disk", "Multi-disk", func(b []byte) { binary.LittleEndian.PutUint16(b[end+4:], 1) }},
	} {
		b := append([]byte(nil), base...)
		tc.mutate(b)
		previewPut(t, f, "bad.zip", b)
		w := f.request("GET", previewURL+"zip?path=bad.zip", "", "owner")
		if w.Code != 400 || !strings.Contains(w.Body.String(), tc.reason) {
			t.Fatal(tc.name, w.Code, w.Body.String())
		}
	}
	// CRC corruption is discovered only on the selected bounded entry read, not by decompressing during listing.
	corrupt := append([]byte(nil), base...)
	corrupt[central+16] ^= 0xff
	previewPut(t, f, "crc.zip", corrupt)
	if w := f.request("GET", previewURL+"zip?path=crc.zip", "", "owner"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := f.request("GET", previewURL+"zip-entry?path=crc.zip&index=0", "", "owner"); w.Code != 400 {
		t.Fatal("CRC not verified", w.Code, w.Body.String())
	}
	for _, b := range [][]byte{[]byte("not a ZIP"), base[:len(base)-10], append(append([]byte(nil), base...), []byte("trailing")...)} {
		previewPut(t, f, "broken.zip", b)
		if w := f.request("GET", previewURL+"zip?path=broken.zip", "", "owner"); w.Code != 400 {
			t.Fatal("corrupt archive accepted", w.Code)
		}
	}
}

func TestFilePreviewZIPResourceBounds(t *testing.T) {
	if _, err := boundedPreviewZIP(nil, zipPreviewMaxSize+1); err == nil || !strings.Contains(err.Error(), "256 MiB") {
		t.Fatal("size limit must precede file reads", err)
	}
	f := newReliabilityFixture(t)
	entries := make([]previewZIPFixtureEntry, zipPreviewMaxEntries+1)
	for i := range entries {
		entries[i].name = fmt.Sprintf("file-%d.txt", i)
	}
	data := previewZIPBytes(t, entries)
	// Lie in EOCD: the actual central-directory record count must also be capped.
	end := len(data) - 22
	binary.LittleEndian.PutUint16(data[end+8:], 1)
	binary.LittleEndian.PutUint16(data[end+10:], 1)
	previewPut(t, f, "many.zip", data)
	w := f.request("GET", previewURL+"zip?path=many.zip", "", "owner")
	if w.Code != 400 || !strings.Contains(w.Body.String(), "10000") {
		t.Fatal(w.Code, w.Body.String())
	}
	// A size lie must never allow a highly expanded entry past the bounded reader.
	data = previewZIPBytes(t, []previewZIPFixtureEntry{{"bomb.txt", strings.Repeat("a", zipPreviewEntryLimit+1), 0}})
	central := bytes.Index(data, []byte("PK\x01\x02"))
	binary.LittleEndian.PutUint32(data[central+24:], 1)
	previewPut(t, f, "bomb.zip", data)
	if w := f.request("GET", previewURL+"zip-entry?path=bomb.zip&index=0", "", "owner"); w.Code != 400 {
		t.Fatal(w.Code, w.Body.String())
	}
	// No entry or traversal artifact was created on disk.
	if _, err := os.Stat(filepath.Join(f.manager.GetWorkingDir("alpha"), "bomb.txt")); !os.IsNotExist(err) {
		t.Fatal("entry extracted", err)
	}
}
