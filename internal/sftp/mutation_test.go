package sftp

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

func TestSFTPMaintenanceExcludesOpenHandlesAndMutators(t *testing.T) {
	f := newFixture(t)
	_, client := f.connect(t, "editor", "password")
	opened, err := client.OpenFile("file.txt", os.O_RDWR)
	must(t, err)
	defer opened.Close()
	dir := filepath.Join(f.root, "srv-alpha")
	release, err := rooted.TryExclusiveMutation(dir)
	must(t, err)
	defer release()

	// The network server runs these competing writes on separate goroutines.
	if n, err := opened.WriteAt([]byte("changed"), 0); err == nil || n != 0 {
		t.Fatalf("old handle wrote during maintenance: %d %v", n, err)
	}
	for _, name := range []string{"file.txt", "new.txt"} {
		file, err := client.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
		if file != nil {
			file.Close()
		}
		denied(t, err)
	}
	denied(t, client.Truncate("file.txt", 0))
	denied(t, client.Rename("file.txt", "moved.txt"))
	denied(t, client.PosixRename("file.txt", "moved.txt"))
	denied(t, client.Remove("file.txt"))
	denied(t, client.Mkdir("newdir"))
	denied(t, client.Chmod("file.txt", 0600))
	buf := make([]byte, len("srv-alpha contents"))
	_, err = opened.ReadAt(buf, 0)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	if string(buf) != "srv-alpha contents" {
		t.Fatal("blocked mutation changed content", string(buf))
	}
	for _, name := range []string{"new.txt", "moved.txt", "newdir"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatalf("blocked operation created %s: %v", name, err)
		}
	}
	release()
	_, err = opened.WriteAt([]byte("changed"), 0)
	must(t, err)
	must(t, opened.Close())
	must(t, client.Rename("file.txt", "moved.txt"))
	got, err := os.ReadFile(filepath.Join(dir, "moved.txt"))
	must(t, err)
	if string(got[:7]) != "changed" {
		t.Fatal("write did not resume")
	}
}
