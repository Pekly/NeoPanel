package rooted

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fixtureMember struct {
	name, data string
	mode       os.FileMode
}

func archiveFixture(t *testing.T, format string, ms []fixtureMember) *os.File {
	t.Helper()
	f, e := os.Create(filepath.Join(t.TempDir(), "archive."+format))
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { f.Close() })
	if format == "zip" {
		w := zip.NewWriter(f)
		for _, m := range ms {
			h := &zip.FileHeader{Name: m.name, Method: zip.Deflate}
			h.SetMode(m.mode)
			out, e := w.CreateHeader(h)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = io.WriteString(out, m.data); e != nil {
				t.Fatal(e)
			}
		}
		if e := w.Close(); e != nil {
			t.Fatal(e)
		}
	} else {
		var out io.Writer = f
		var gz *gzip.Writer
		if format == "tgz" {
			gz = gzip.NewWriter(f)
			out = gz
		}
		w := tar.NewWriter(out)
		for _, m := range ms {
			typ := byte(tar.TypeReg)
			size := int64(len(m.data))
			if m.mode.IsDir() {
				typ = tar.TypeDir
				size = 0
			}
			if m.mode&os.ModeSymlink != 0 {
				typ = tar.TypeSymlink
				size = 0
			}
			h := &tar.Header{Name: m.name, Mode: 0644, Size: size, Typeflag: typ, Linkname: "outside"}
			if e := w.WriteHeader(h); e != nil {
				t.Fatal(e)
			}
			if size > 0 {
				if _, e := io.WriteString(w, m.data); e != nil {
					t.Fatal(e)
				}
			}
		}
		if e := w.Close(); e != nil {
			t.Fatal(e)
		}
		if gz != nil {
			if e := gz.Close(); e != nil {
				t.Fatal(e)
			}
		}
	}
	return f
}
func TestArchiveLinksAndHardlinks(t *testing.T) {
	for _, format := range []string{"zip", "tar", "tgz"} {
		for _, kind := range []string{"hard", "final-link", "parent-link"} {
			t.Run(format+"/"+kind, func(t *testing.T) {
				dir, r := testRoot(t)
				outside := t.TempDir()
				victim := filepath.Join(outside, "victim")
				put(t, victim, "secret")
				name := "victim"
				switch kind {
				case "hard":
					if e := os.Link(victim, filepath.Join(dir, name)); e != nil {
						t.Fatal(e)
					}
				case "final-link":
					symlink(t, victim, filepath.Join(dir, name))
				case "parent-link":
					symlink(t, outside, filepath.Join(dir, "out"))
					name = "out/victim"
				}
				f := archiveFixture(t, format, []fixtureMember{{name, "changed", 0644}})
				_, e := Extract(context.Background(), r, f, "archive."+format, ".", DefaultArchiveLimits)
				if kind == "hard" && e != nil {
					t.Fatal(e)
				}
				if kind != "hard" && e == nil {
					t.Fatal("accepted planted symlink")
				}
				sentinel(t, victim)
			})
		}
	}
}
func TestArchiveBudgetsAndInvalidMembers(t *testing.T) {
	for _, format := range []string{"zip", "tar", "tgz"} {
		for _, kind := range []string{"expanded", "entries", "unsafe", "special", "cancel", "duplicate"} {
			t.Run(format+"/"+kind, func(t *testing.T) {
				_, r := testRoot(t)
				limits := ArchiveLimits{4, 1 << 20, 64, 8192}
				ms := []fixtureMember{{"ok", "ok", 0644}}
				ctx := context.Background()
				switch kind {
				case "expanded":
					ms = append(ms, fixtureMember{"bomb", strings.Repeat("x", 100), 0644})
				case "entries":
					limits.Entries = 1
					ms = []fixtureMember{{"a/", "", os.ModeDir | 0755}, {"b/", "", os.ModeDir | 0755}}
				case "unsafe":
					ms = append(ms, fixtureMember{"../outside", "bad", 0644})
				case "special":
					ms = append(ms, fixtureMember{"link", "", os.ModeSymlink | 0777})
				case "cancel":
					var cancel context.CancelFunc
					ctx, cancel = context.WithCancel(ctx)
					cancel()
				case "duplicate":
					ms = append(ms, ms[0])
				}
				f := archiveFixture(t, format, ms)
				if _, e := Extract(ctx, r, f, "archive."+format, ".", limits); e == nil {
					t.Fatal("archive accepted")
				}
				entries, e := r.ReadDir(".")
				if e != nil || len(entries) != 0 {
					t.Fatal("leaked partial extraction", entries, e)
				}
			})
		}
	}
}
func TestZIPCentralDirectoryPreflight(t *testing.T) {
	f := archiveFixture(t, "zip", []fixtureMember{{"ok", "data", 0644}})
	st, _ := f.Stat()
	l := DefaultArchiveLimits
	l.Metadata = 1
	if _, e := boundedZIP(f, st.Size(), l); e == nil {
		t.Fatal("metadata budget ignored")
	}
	b := make([]byte, st.Size())
	f.ReadAt(b, 0)
	end := bytes.LastIndex(b, []byte("PK\x05\x06"))
	binary.LittleEndian.PutUint16(b[end+8:], 0)
	binary.LittleEndian.PutUint16(b[end+10:], 0)
	if _, e := f.WriteAt(b, 0); e != nil {
		t.Fatal(e)
	}
	if _, e := boundedZIP(f, st.Size(), DefaultArchiveLimits); e == nil {
		t.Fatal("actual central entry count ignored")
	}
}
