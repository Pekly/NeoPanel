package trash

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

func TestMoveFailureRollsBackMetadata(t *testing.T) {
	s, r, dir := fixture(t)
	putFile(t, r, "locked")
	path, err := windows.UTF16PtrFromString(filepath.Join(dir, "locked"))
	if err != nil {
		t.Fatal(err)
	}
	// Deny DELETE sharing to force the move to fail after metadata is committed.
	h, err := windows.CreateFile(path, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(h)
	if err := s.Put("alpha", r, "locked"); err == nil {
		t.Fatal("move unexpectedly succeeded")
	}
	if data, err := r.ReadFile("locked"); err != nil || string(data) != "original" {
		t.Fatal("source changed", string(data), err)
	}
	entries, err := os.ReadDir(filepath.Join(s.dir, serverKey("alpha")))
	if err != nil || len(entries) != 0 {
		t.Fatal("orphan metadata after failed move", entries, err)
	}
}
