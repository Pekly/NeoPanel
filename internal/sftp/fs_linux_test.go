//go:build linux

package sftp

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func linuxHandler(t *testing.T, root string) *chrootHandler {
	t.Helper()
	h := &chrootHandler{rootDir: root}
	pin, err := h.fsRoot()
	must(t, err)
	t.Cleanup(func() { pin.Close() })
	h.rootInfo, err = pin.Stat()
	must(t, err)
	return h
}

// Deterministically replace the parent in the exact former check/use window.
// Both source and destination descriptors must remain bound to the old directory.
func TestLinuxParentDescriptorSurvivesSymlinkSwap(t *testing.T) {
	base := t.TempDir()
	root, outside := filepath.Join(base, "jail"), filepath.Join(base, "outside")
	must(t, os.MkdirAll(filepath.Join(root, "parent"), 0755))
	must(t, os.Mkdir(outside, 0755))
	must(t, os.WriteFile(filepath.Join(root, "parent", "file"), []byte("inside"), 0644))
	must(t, os.WriteFile(filepath.Join(outside, "file"), []byte("outside"), 0644))
	h := linuxHandler(t, root)
	parent, name, err := h.fsParent("/parent/file")
	must(t, err)
	defer parent.Close()
	must(t, os.Rename(filepath.Join(root, "parent"), filepath.Join(root, "saved")))
	must(t, os.Symlink(outside, filepath.Join(root, "parent")))
	file, err := openRegularAt(parent, name, os.O_RDWR)
	must(t, err)
	data, err := io.ReadAll(file)
	must(t, err)
	if string(data) != "inside" {
		t.Fatalf("escaped through replacement: %q", data)
	}
	must(t, file.Truncate(2))
	must(t, file.Chmod(0600))
	must(t, unix.Futimes(int(file.Fd()), []unix.Timeval{{Sec: 100}, {Sec: 100}}))
	file.Close()
	must(t, unix.Mkdirat(int(parent.Fd()), "dir", 0755))
	must(t, unix.Renameat(int(parent.Fd()), name, int(parent.Fd()), "renamed"))
	must(t, unix.Unlinkat(int(parent.Fd()), "renamed", 0))
	must(t, unix.Unlinkat(int(parent.Fd()), "dir", unix.AT_REMOVEDIR))
	data, err = os.ReadFile(filepath.Join(outside, "file"))
	must(t, err)
	if string(data) != "outside" {
		t.Fatalf("outside file changed: %q", data)
	}
	_, err = h.fsStat("/parent/file")
	denied(t, err)
	_, err = h.fsList("/parent")
	denied(t, err)
	file, err = h.fsOpen("/parent/file", os.O_WRONLY|os.O_TRUNC)
	if file != nil {
		file.Close()
	}
	denied(t, err)
}

func TestLinuxSymlinkAncestorsAndLeavesDenied(t *testing.T) {
	f := newFixture(t)
	_, c := f.connect(t, "owner", "password")
	root := filepath.Join(f.root, "srv-alpha")
	outside := filepath.Join(f.root, "srv-beta")
	must(t, os.Symlink(outside, filepath.Join(root, "parent")))
	must(t, os.Symlink(filepath.Join(outside, "file.txt"), filepath.Join(root, "leaf")))
	for _, p := range []string{"/parent/file.txt", "/leaf"} {
		file, err := c.Open(p)
		if file != nil {
			file.Close()
		}
		denied(t, err)
		file, err = c.Create(p)
		if file != nil {
			file.Close()
		}
		denied(t, err)
		_, err = c.Stat(p)
		denied(t, err)
		_, err = c.Lstat(p)
		denied(t, err)
		denied(t, c.Chmod(p, 0600))
		denied(t, c.Chtimes(p, time.Unix(100, 0), time.Unix(100, 0)))
		denied(t, c.Truncate(p, 0))
		denied(t, c.Remove(p))
		denied(t, c.Rename(p, "/renamed"))
		denied(t, c.Rename("/file.txt", p))
	}
	denied(t, c.Mkdir("/parent/new"))
	denied(t, c.RemoveDirectory("/parent/world"))
	_, err := c.ReadDir("/parent")
	denied(t, err)
	// Jail creation must not traverse a symlink in the configured ancestors either.
	denied(t, prepareJail(filepath.Join(root, "parent", "new-jail")))
	if _, err := os.Stat(filepath.Join(outside, "new-jail")); !os.IsNotExist(err) {
		t.Fatalf("created jail outside: %v", err)
	}
	// Positive descriptor-based metadata operations.
	must(t, c.Chmod("/file.txt", 0600))
	must(t, c.Chtimes("/file.txt", time.Unix(100, 0), time.Unix(100, 0)))
	must(t, c.Truncate("/file.txt", 3))
	info, err := os.Stat(filepath.Join(root, "file.txt"))
	must(t, err)
	if info.Size() != 3 || info.Mode().Perm() != 0600 {
		t.Fatalf("metadata not applied: %v", info)
	}
	must(t, c.Chtimes("/file.txt", time.Unix(100, 0), time.Unix(100, 0)))
	info, err = os.Stat(filepath.Join(root, "file.txt"))
	must(t, err)
	if info.ModTime().Unix() != 100 {
		t.Fatalf("mtime not applied: %v", info.ModTime())
	}
}

func TestLinuxReplacedRealRootRevokesEstablishedHandles(t *testing.T) {
	f := newFixture(t)
	_, c := f.connect(t, "owner", "key")
	read, err := c.Open("/file.txt")
	must(t, err)
	defer read.Close()
	write, err := c.OpenFile("/file.txt", os.O_WRONLY)
	must(t, err)
	defer write.Close()
	root := filepath.Join(f.root, "srv-alpha")
	must(t, os.Rename(root, root+"-saved"))
	must(t, os.Mkdir(root, 0755))
	must(t, os.WriteFile(filepath.Join(root, "file.txt"), []byte("replacement"), 0644))
	_, err = read.ReadAt(make([]byte, 1), 0)
	denied(t, err)
	_, err = write.WriteAt([]byte("x"), 0)
	denied(t, err)
	_, err = c.ReadDir("/")
	denied(t, err)
	denied(t, c.Mkdir("/new"))
	file, err := c.Create("/file.txt")
	if file != nil {
		file.Close()
	}
	denied(t, err)
}

// Exercise actual SSH/SFTP operations while another process-like goroutine
// repeatedly exchanges a real in-jail parent with an outside-pointing symlink.
func TestLinuxConcurrentParentSwap(t *testing.T) {
	f := newFixture(t)
	_, c := f.connect(t, "owner", "password")
	root := filepath.Join(f.root, "srv-alpha")
	outside := filepath.Join(f.root, "srv-beta")
	parent, saved := filepath.Join(root, "race"), filepath.Join(root, "saved")
	must(t, os.Mkdir(parent, 0755))
	must(t, os.WriteFile(filepath.Join(parent, "file.txt"), []byte("inside"), 0644))
	must(t, os.Mkdir(filepath.Join(outside, "empty"), 0755))
	before, err := os.Stat(filepath.Join(outside, "file.txt"))
	must(t, err)
	stop := make(chan struct{})
	failures := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			if err := os.Rename(parent, saved); err != nil {
				failures <- err
				return
			}
			if err := os.Symlink(outside, parent); err != nil {
				failures <- err
				return
			}
			time.Sleep(time.Microsecond)
			if err := os.Remove(parent); err != nil {
				failures <- err
				return
			}
			if err := os.Rename(saved, parent); err != nil {
				failures <- err
				return
			}
		}
	}()
	defer func() { close(stop); wg.Wait() }()
	for i := 0; i < 150; i++ {
		file, err := c.Open("/race/file.txt")
		if err == nil {
			data, readErr := io.ReadAll(file)
			file.Close()
			if readErr == nil && string(data) == "srv-beta contents" {
				t.Fatal("outside data disclosed")
			}
		}
		file, err = c.OpenFile("/race/file.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
		if err == nil {
			_, _ = file.WriteAt([]byte("inside"), 0)
			file.Close()
		}
		_, _ = c.Stat("/race/file.txt")
		_, _ = c.ReadDir("/race")
		_ = c.Chmod("/race/file.txt", 0600)
		_ = c.Chtimes("/race/file.txt", time.Unix(100, 0), time.Unix(100, 0))
		_ = c.Truncate("/race/file.txt", 3)
		_ = c.Mkdir("/race/new")
		_ = c.RemoveDirectory("/race/new")
		_ = c.RemoveDirectory("/race/empty")
		_ = c.Rename("/race/file.txt", "/race/renamed")
		_ = c.Rename("/race/renamed", "/race/file.txt")
		_ = c.Remove("/race/file.txt")
	}
	select {
	case err := <-failures:
		t.Fatal(err)
	default:
	}
	data, err := os.ReadFile(filepath.Join(outside, "file.txt"))
	must(t, err)
	if string(data) != "srv-beta contents" {
		t.Fatalf("outside file modified: %q", data)
	}
	after, err := os.Stat(filepath.Join(outside, "file.txt"))
	must(t, err)
	if before.Mode() != after.Mode() || !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("outside metadata changed")
	}
	if _, err := os.Stat(filepath.Join(outside, "empty")); err != nil {
		t.Fatal("outside directory removed")
	}
	for _, name := range []string{"new", "renamed"} {
		if _, err := os.Stat(filepath.Join(outside, name)); !os.IsNotExist(err) {
			t.Fatalf("outside entry %s created: %v", name, err)
		}
	}
}

func TestLinuxLeafSwapBeforeOpenDoesNotTruncateTarget(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "jail")
	must(t, os.Mkdir(root, 0755))
	outside := filepath.Join(base, "secret")
	must(t, os.WriteFile(outside, []byte("untouched"), 0644))
	must(t, os.WriteFile(filepath.Join(root, "file"), []byte("inside"), 0644))
	h := linuxHandler(t, root)
	dir, name, err := h.fsParent("/file")
	must(t, err)
	defer dir.Close()
	_, err = statAt(dir, name)
	must(t, err)
	must(t, os.Remove(filepath.Join(root, "file")))
	must(t, os.Symlink(outside, filepath.Join(root, "file")))
	file, err := openRegularAt(dir, name, os.O_WRONLY|os.O_TRUNC)
	if file != nil {
		file.Close()
	}
	denied(t, err)
	_, err = statAt(dir, name)
	denied(t, err)
	data, err := os.ReadFile(outside)
	must(t, err)
	if string(data) != "untouched" {
		t.Fatalf("symlink target truncated: %q", data)
	}
}
