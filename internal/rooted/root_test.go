package rooted

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testRoot(t *testing.T) (string, *Root) {
	t.Helper()
	dir := t.TempDir()
	r, e := Open(dir)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { r.Close() })
	return dir, r
}
func put(t *testing.T, name, data string) {
	t.Helper()
	if e := os.WriteFile(name, []byte(data), 0600); e != nil {
		t.Fatal(e)
	}
}
func symlink(t *testing.T, target, name string) {
	t.Helper()
	if e := os.Symlink(target, name); e != nil {
		t.Skipf("symlinks unavailable: %v", e)
	}
}
func sentinel(t *testing.T, name string) {
	t.Helper()
	b, e := os.ReadFile(name)
	if e != nil || string(b) != "secret" {
		t.Fatalf("outside sentinel changed: %q %v", b, e)
	}
}
func TestPublicationRejectsLinksAndDoesNotTruncateHardlinks(t *testing.T) {
	dir, r := testRoot(t)
	outside := filepath.Join(t.TempDir(), "victim")
	put(t, outside, "secret")
	if e := os.Link(outside, filepath.Join(dir, "hard")); e != nil {
		t.Fatal(e)
	}
	if e := r.WriteFile("hard", []byte("new"), 0644); e != nil {
		t.Fatal(e)
	}
	sentinel(t, outside)
	symlink(t, outside, filepath.Join(dir, "link"))
	if e := r.WriteFile("link", []byte("bad"), 0644); e == nil {
		t.Fatal("accepted symlink")
	}
	if f, e := r.Open("link"); e == nil {
		f.Close()
		t.Fatal("read symlink")
	}
	sentinel(t, outside)
}
func TestPublicationParentSwap(t *testing.T) {
	dir, r := testRoot(t)
	outside := t.TempDir()
	put(t, filepath.Join(outside, "victim"), "secret")
	if e := r.Mkdir("d", 0755); e != nil {
		t.Fatal(e)
	}
	// Synchronize the swap after validation and after opening the private output.
	e := r.Publish("d/victim", 0644, func(f *os.File) error {
		if e := os.Rename(filepath.Join(dir, "d"), filepath.Join(dir, "old")); e != nil {
			t.Fatal(e)
		}
		symlink(t, outside, filepath.Join(dir, "d"))
		_, e := io.WriteString(f, "bad")
		return e
	})
	if e == nil {
		t.Fatal("parent swap accepted")
	}
	sentinel(t, filepath.Join(outside, "victim"))
}
func TestPinnedSubRootAndWalkSwap(t *testing.T) {
	dir, r := testRoot(t)
	r.Mkdir("d", 0755)
	put(t, filepath.Join(dir, "d", "file"), "inside")
	outside := t.TempDir()
	put(t, filepath.Join(outside, "file"), "secret")
	sub, e := r.Sub("d")
	if e != nil {
		t.Fatal(e)
	}
	defer sub.Close()
	swapped := false
	e = r.Walk(".", func(p string, st os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if p == "d" && !swapped {
			swapped = true
			if e := os.Rename(filepath.Join(dir, "d"), filepath.Join(dir, "old")); e != nil {
				t.Fatal(e)
			}
			symlink(t, outside, filepath.Join(dir, "d"))
		}
		return nil
	})
	if e == nil {
		t.Fatal("walk accepted swapped symlink")
	}
	b, e := sub.ReadFile("file")
	if e != nil || string(b) != "inside" {
		t.Fatal(string(b), e)
	}
	if e := sub.WriteFile("file", []byte("safe"), 0600); e != nil {
		t.Fatal(e)
	}
	sentinel(t, filepath.Join(outside, "file"))
}
func TestCopyLimitsCancellationAndCleanup(t *testing.T) {
	_, r := testRoot(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := r.Copy(ctx, "cancel", strings.NewReader("abc"), 10, 0600); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if _, e := r.Copy(context.Background(), "bomb", strings.NewReader(strings.Repeat("x", 100)), 4, 0600); e == nil {
		t.Fatal("byte limit ignored")
	}
	entries, e := r.ReadDir(".")
	if e != nil || len(entries) != 0 {
		t.Fatal("partial files", entries, e)
	}
	for _, p := range []string{"../outside", "a/../../outside", "a\x00b", "C:/outside"} {
		if _, e := r.Path(p); e == nil {
			t.Fatal("accepted", p)
		}
	}
}
