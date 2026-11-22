//go:build linux

package proxyconfig

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

func TestFIFORejectedWithoutBlocking(t *testing.T) {
	dir := t.TempDir()
	for _, kind := range []string{"velocity", "bungeecord"} {
		name := filepath.Join(dir, FileName(kind))
		if err := unix.Mkfifo(name, 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := Read(dir, kind, 25565); err == nil {
			t.Fatal("FIFO read accepted")
		}
		if _, err := Ensure(dir, kind, 25565); err == nil {
			t.Fatal("FIFO Ensure accepted")
		}
		if err := SyncPort(dir, kind, 25565); err == nil {
			t.Fatal("FIFO SyncPort accepted")
		}
		if err := os.Remove(name); err != nil {
			t.Fatal(err)
		}
	}
	writeDoc(t, dir, "velocity", starter("velocity", 25565))
	if err := unix.Mkfifo(filepath.Join(dir, "forwarding.secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadSecret(dir, "velocity"); err == nil {
		t.Fatal("FIFO secret accepted")
	}
	if _, err := Ensure(dir, "velocity", 25565); err == nil {
		t.Fatal("FIFO secret Ensure accepted")
	}
}
