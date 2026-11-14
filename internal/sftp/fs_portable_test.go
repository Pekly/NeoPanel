//go:build !linux

package sftp

import (
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	sftpclient "github.com/pkg/sftp"
)

func portableTestHandler(t *testing.T) (*chrootHandler, string) {
	t.Helper()
	base, err := filepath.EvalSymlinks(t.TempDir())
	must(t, err)
	dir := filepath.Join(base, "servers", "alpha")
	must(t, prepareJail(dir))
	h := &chrootHandler{rootDir: dir}
	f, err := h.fsRoot()
	must(t, err)
	h.rootInfo, err = f.Stat()
	must(t, err)
	must(t, f.Close())
	return h, dir
}
func portableSizeRequest(p string, n uint64) *sftpclient.Request {
	attrs := make([]byte, 8)
	binary.BigEndian.PutUint64(attrs, n)
	return &sftpclient.Request{Method: "Setstat", Filepath: p, Flags: 1, Attrs: attrs}
}
func portableTimesRequest(p string) *sftpclient.Request {
	attrs := make([]byte, 8)
	binary.BigEndian.PutUint32(attrs, 100)
	binary.BigEndian.PutUint32(attrs[4:], 200)
	return &sftpclient.Request{Method: "Setstat", Filepath: p, Flags: 8, Attrs: attrs}
}
func portableSentinel(t *testing.T, p string) {
	t.Helper()
	b, err := os.ReadFile(p)
	must(t, err)
	if string(b) != "outside-secret" {
		t.Fatalf("outside changed: %q", b)
	}
}
func TestPortableRandomAccessAndCommands(t *testing.T) {
	h, dir := portableTestHandler(t)
	must(t, h.fsCommand(&sftpclient.Request{Method: "Mkdir", Filepath: "/d"}))
	f, err := h.fsOpen("/d/file", os.O_CREATE|os.O_RDWR|os.O_TRUNC)
	must(t, err)
	_, err = f.WriteAt([]byte("world"), 5)
	must(t, err)
	_, err = f.WriteAt([]byte("hello"), 0)
	must(t, err)
	b := make([]byte, 10)
	_, err = f.ReadAt(b, 0)
	must(t, err)
	if string(b) != "helloworld" {
		t.Fatal(string(b))
	}
	must(t, f.Close())
	must(t, h.fsCommand(&sftpclient.Request{Method: "Rename", Filepath: "/d/file", Target: "/renamed"}))
	entries, err := h.fsList("/")
	must(t, err)
	if len(entries) != 2 {
		t.Fatal(entries)
	}
	must(t, h.fsCommand(portableSizeRequest("/renamed", 5)))
	must(t, h.fsCommand(portableTimesRequest("/renamed")))
	must(t, h.fsCommand(portableTimesRequest("/d")))
	perms := make([]byte, 4)
	binary.BigEndian.PutUint32(perms, 0755)
	must(t, h.fsCommand(&sftpclient.Request{Method: "Setstat", Filepath: "/d", Flags: 4, Attrs: perms}))
	binary.BigEndian.PutUint32(perms, 0644)
	must(t, h.fsCommand(&sftpclient.Request{Method: "Setstat", Filepath: "/renamed", Flags: 4, Attrs: perms}))
	info, err := h.fsStat("/renamed")
	must(t, err)
	if info.Size() != 5 || info.ModTime().Unix() != 200 {
		t.Fatal(info.Size(), info.ModTime())
	}
	f, err = h.fsOpen("/renamed", os.O_WRONLY|os.O_TRUNC)
	must(t, err)
	must(t, f.Close())
	info, err = os.Stat(filepath.Join(dir, "renamed"))
	must(t, err)
	if info.Size() != 0 {
		t.Fatal("not truncated")
	}
	must(t, h.fsCommand(&sftpclient.Request{Method: "Remove", Filepath: "/renamed"}))
	must(t, h.fsCommand(&sftpclient.Request{Method: "Rmdir", Filepath: "/d"}))
}
func TestPortableRejectRootReplacement(t *testing.T) {
	h, dir := portableTestHandler(t)
	must(t, os.Rename(dir, dir+"-old"))
	must(t, os.Mkdir(dir, 0755))
	victim := filepath.Join(dir, "file")
	must(t, os.WriteFile(victim, []byte("outside-secret"), 0600))
	if f, err := h.fsOpen("/file", os.O_WRONLY|os.O_TRUNC); err == nil {
		f.Close()
		t.Fatal("replaced root accepted")
	}
	_, err := h.fsStat("/file")
	denied(t, err)
	_, err = h.fsList("/")
	denied(t, err)
	denied(t, h.fsCommand(portableSizeRequest("/file", 0)))
	denied(t, h.fsCommand(&sftpclient.Request{Method: "Remove", Filepath: "/file"}))
	portableSentinel(t, victim)
}
func TestPortableLinksRejectedBeforeTruncation(t *testing.T) {
	h, dir := portableTestHandler(t)
	outside := t.TempDir()
	victim := filepath.Join(outside, "file")
	must(t, os.WriteFile(victim, []byte("outside-secret"), 0600))
	if err := os.Symlink(victim, filepath.Join(dir, "link")); err != nil {
		t.Skip(err)
	}
	must(t, os.Symlink(outside, filepath.Join(dir, "parent")))
	for _, p := range []string{"/link", "/parent/file"} {
		if f, err := h.fsOpen(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC); err == nil {
			f.Close()
			t.Fatal("link accepted", p)
		}
		_, err := h.fsStat(p)
		denied(t, err)
		denied(t, h.fsCommand(portableSizeRequest(p, 0)))
		denied(t, h.fsCommand(portableTimesRequest(p)))
		denied(t, h.fsCommand(&sftpclient.Request{Method: "Rename", Filepath: p, Target: "/moved"}))
	}
	_, err := h.fsList("/")
	denied(t, err)
	portableSentinel(t, victim)
}
func TestPortablePinnedParentSurvivesReplacement(t *testing.T) {
	h, dir := portableTestHandler(t)
	must(t, os.Mkdir(filepath.Join(dir, "d"), 0755))
	must(t, os.WriteFile(filepath.Join(dir, "d", "file"), []byte("inside"), 0600))
	outside := t.TempDir()
	victim := filepath.Join(outside, "file")
	must(t, os.WriteFile(victim, []byte("outside-secret"), 0600))
	probe := filepath.Join(dir, "probe")
	if err := os.Symlink(outside, probe); err != nil {
		t.Skip(err)
	}
	must(t, os.Remove(probe))
	root, err := h.portableRoot()
	must(t, err)
	defer root.Close()
	parent, name, err := portableParent(root, "d"+string(filepath.Separator)+"file")
	must(t, err)
	defer parent.Close()
	// Synchronized after parent lookup. Some Windows filesystems refuse renaming
	// an open directory; that is also safe, but the normal NTFS path allows this.
	if err := os.Rename(filepath.Join(dir, "d"), filepath.Join(dir, "old")); err != nil {
		t.Skipf("OS prevents pinned-directory replacement: %v", err)
	}
	must(t, os.Symlink(outside, filepath.Join(dir, "d")))
	f, err := portableOpen(parent, name, os.O_RDWR|os.O_TRUNC, false)
	must(t, err)
	_, err = f.WriteAt([]byte("safe"), 0)
	must(t, err)
	must(t, portableFileTimes(f, time.Unix(100, 0), time.Unix(200, 0)))
	must(t, f.Close())
	portableSentinel(t, victim)
	b, err := os.ReadFile(filepath.Join(dir, "old", "file"))
	must(t, err)
	if string(b) != "safe" {
		t.Fatal(string(b))
	}
	if f, err := h.fsOpen("/d/file", os.O_RDONLY); err == nil {
		f.Close()
		t.Fatal("new operation followed replacement")
	}
}
func TestPortableRejectDirectoryAsUpload(t *testing.T) {
	h, _ := portableTestHandler(t)
	if f, err := h.fsOpen("/", os.O_WRONLY|os.O_TRUNC); err == nil {
		f.Close()
		t.Fatal("directory accepted")
	}
	f, err := h.fsOpen("/file", os.O_CREATE|os.O_RDWR)
	must(t, err)
	defer f.Close()
	_, err = io.WriteString(f, "data")
	must(t, err)
}
