package sftp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSFTPRenameRequiresEditAndDelete(t *testing.T) {
	f := newFixture(t)
	must(t, os.WriteFile(filepath.Join(f.root, "srv-alpha", "existing.txt"), []byte("preserve target"), 0600))
	must(t, f.db.UpdateServerSubuser("srv-alpha", f.members["editor"].ID, []string{"sftp.access", "files.view", "files.edit"}))
	_, client := f.connect(t, "editor", "password")
	denied(t, client.Rename("file.txt", "moved.txt"))
	denied(t, client.PosixRename("file.txt", "existing.txt"))
	source, err := os.ReadFile(filepath.Join(f.root, "srv-alpha", "file.txt"))
	must(t, err)
	target, err := os.ReadFile(filepath.Join(f.root, "srv-alpha", "existing.txt"))
	must(t, err)
	if string(source) != "srv-alpha contents" || string(target) != "preserve target" {
		t.Fatal("denied rename changed files")
	}
	must(t, f.db.UpdateServerSubuser("srv-alpha", f.members["editor"].ID, []string{"sftp.access", "files.view", "files.delete"}))
	denied(t, client.Rename("file.txt", "moved.txt"))
	denied(t, client.PosixRename("file.txt", "existing.txt"))
	must(t, f.db.UpdateServerSubuser("srv-alpha", f.members["editor"].ID, []string{"sftp.access", "files.view", "files.edit", "files.delete"}))
	must(t, client.PosixRename("file.txt", "existing.txt"))
	target, err = os.ReadFile(filepath.Join(f.root, "srv-alpha", "existing.txt"))
	must(t, err)
	if string(target) != "srv-alpha contents" {
		t.Fatal("authorized rename failed")
	}
}
