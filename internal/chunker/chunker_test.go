package chunker

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

type mockTransport func(*http.Request) (*http.Response, error)

func (f mockTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestChunkerDownloadPinAndBounds(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		ok         bool
	}{{"valid", "jar", true}, {"wrong digest", "bad", false}, {"oversized", "jar-more", false}, {"short", "ja", false}} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "tool.jar")
			h := sha256.Sum256([]byte("jar"))
			client := &http.Client{Transport: mockTransport(func(r *http.Request) (*http.Response, error) {
				if r.URL.Scheme != "https" {
					t.Fatal("insecure")
				}
				return &http.Response{StatusCode: 200, ContentLength: -1, Body: io.NopCloser(strings.NewReader(tc.body)), Header: make(http.Header)}, nil
			})}
			err := download(context.Background(), client, "https://github.com/test", path, 3, hex.EncodeToString(h[:]))
			if (err == nil) != tc.ok {
				t.Fatal(err)
			}
			if !tc.ok {
				if _, e := os.Stat(path); !os.IsNotExist(e) {
					t.Fatal("partial published")
				}
			}
		})
	}
}
func TestChunkerRealCloneAndSettings(t *testing.T) {
	srcDir, dstDir := t.TempDir(), t.TempDir()
	os.Mkdir(filepath.Join(srcDir, "region"), 0700)
	os.WriteFile(filepath.Join(srcDir, "region", "r.mca"), []byte("original"), 0600)
	src, _ := rooted.Open(srcDir)
	defer src.Close()
	dst, _ := rooted.Open(dstDir)
	defer dst.Close()
	hash, n, e := Tree(context.Background(), src, dst, 1024)
	if e != nil || n != 8 {
		t.Fatal(n, e)
	}
	other, _, e := Tree(context.Background(), dst, nil, 1024)
	if e != nil || other != hash {
		t.Fatal("clone fingerprint mismatch", e)
	}
	os.WriteFile(filepath.Join(dstDir, "region", "r.mca"), []byte("changed"), 0600)
	original, _ := os.ReadFile(filepath.Join(srcDir, "region", "r.mca"))
	if string(original) != "original" {
		t.Fatal("clone aliases original")
	}
	if _, _, e = Tree(context.Background(), src, nil, 1); e == nil {
		t.Fatal("byte bound bypass")
	}
	os.WriteFile(filepath.Join(srcDir, "converter.chunker.json"), []byte("{}"), 0600)
	if _, _, e = Tree(context.Background(), src, nil, 1024); e == nil {
		t.Fatal("settings allowed")
	}
}
func TestChunkerRejectLinks(t *testing.T) {
	srcDir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret")
	os.WriteFile(outside, []byte("secret"), 0600)
	if e := os.Symlink(outside, filepath.Join(srcDir, "link")); e != nil {
		t.Skip("symlinks unavailable", e)
	}
	src, _ := rooted.Open(srcDir)
	defer src.Close()
	if _, _, e := Tree(context.Background(), src, nil, 1024); e == nil {
		t.Fatal("symlink accepted")
	}
}
func testNBT(t *testing.T, edition string, version int32) []byte {
	t.Helper()
	var b bytes.Buffer
	var order binary.ByteOrder = binary.BigEndian
	if edition == "bedrock" {
		order = binary.LittleEndian
	}
	text := func(s string) { binary.Write(&b, order, uint16(len(s))); b.WriteString(s) }
	b.WriteByte(10)
	text("")
	if edition == "java" {
		b.WriteByte(10)
		text("Data")
		b.WriteByte(3)
		text("DataVersion")
		binary.Write(&b, order, version)
		b.WriteByte(0)
	} else {
		b.WriteByte(9)
		text("lastOpenedWithVersion")
		b.WriteByte(3)
		binary.Write(&b, order, int32(3))
		for _, v := range []int32{1, 20, version} {
			binary.Write(&b, order, v)
		}
	}
	b.WriteByte(0)
	if edition == "java" {
		var out bytes.Buffer
		gz := gzip.NewWriter(&out)
		gz.Write(b.Bytes())
		gz.Close()
		return out.Bytes()
	}
	var out bytes.Buffer
	binary.Write(&out, order, int32(8))
	binary.Write(&out, order, int32(b.Len()))
	out.Write(b.Bytes())
	return out.Bytes()
}
func TestChunkerOutputRequiresActualVersion(t *testing.T) {
	for _, id := range []string{"JAVA_1_20_5", "BEDROCK_R20_80"} {
		t.Run(id, func(t *testing.T) {
			f, _ := Lookup(id)
			v := f.DataVersion
			if f.Edition == "bedrock" {
				v = 80
			}
			dir := t.TempDir()
			root, _ := rooted.Open(dir)
			defer root.Close()
			for _, valid := range []bool{true, false} {
				version := v
				if !valid {
					version++
				}
				os.WriteFile(filepath.Join(dir, "level.dat"), testNBT(t, f.Edition, version), 0600)
				e := ValidateLevel(root, f)
				if (e == nil) != valid {
					t.Fatal(valid, e)
				}
			}
			os.WriteFile(filepath.Join(dir, "level.dat"), []byte("not NBT"), 0600)
			if ValidateLevel(root, f) == nil {
				t.Fatal("junk accepted")
			}
		})
	}
}
func TestChunkerBoundedLogsAndMarker(t *testing.T) {
	l := &Log{}
	l.Write([]byte(strings.Repeat("x", 100000)))
	l.Write([]byte("\n52.50%\n"))
	s, p, ok := l.Snapshot()
	if len(s) > 32768 || p != 52.5 || ok {
		t.Fatal(len(s), p, ok)
	}
	l.Write([]byte("Conversion complete! "))
	l.Write([]byte("Took 1 second\n"))
	_, _, ok = l.Snapshot()
	if !ok {
		t.Fatal("split success marker missed")
	}
}
func TestChunkerOutputArchiveBounds(t *testing.T) {
	for _, tc := range []struct {
		name string
		kind byte
		size int64
		ok   bool
	}{{"level.dat", tar.TypeReg, 4, true}, {"../escape", tar.TypeReg, 4, false}, {"link", tar.TypeSymlink, 0, false}, {"big", tar.TypeReg, 8, false}} {
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer
			tw := tar.NewWriter(&b)
			tw.WriteHeader(&tar.Header{Name: tc.name, Typeflag: tc.kind, Mode: 0600, Size: tc.size, Linkname: "outside"})
			if tc.size > 0 {
				tw.Write(bytes.Repeat([]byte("x"), int(tc.size)))
			}
			tw.Close()
			e := ExtractOutput(context.Background(), &b, t.TempDir(), 4)
			if (e == nil) != tc.ok {
				t.Fatal(e)
			}
		})
	}
}
func TestChunkerCatalogClosed(t *testing.T) {
	for _, id := range []string{"INPUT", "PREVIEW", "SETTINGS", "JAVA_999", "-jar"} {
		if _, ok := Lookup(id); ok {
			t.Fatal(id)
		}
	}
	formats := Formats()
	formats[0].ID = "changed"
	if Formats()[0].ID == "changed" {
		t.Fatal("catalog mutable")
	}
}
