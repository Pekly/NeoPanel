package trash

import (
	"github.com/Pekly/NeoPanel/internal/rooted"
	"golang.org/x/sys/unix"
	"os"
)

func moveNoReplace(from *rooted.Root, old string, to *rooted.Root, name string) error {
	if !rooted.Basename(old) || !rooted.Basename(name) {
		return os.ErrInvalid
	}
	a, err := from.Root.Open(".")
	if err != nil {
		return err
	}
	defer a.Close()
	b, err := to.Root.Open(".")
	if err != nil {
		return err
	}
	defer b.Close()
	// Both parent directories are pinned; leaf symlinks are never followed.
	err = unix.Renameat2(int(a.Fd()), old, int(b.Fd()), name, unix.RENAME_NOREPLACE)
	if err != nil {
		return &os.LinkError{Op: "rename", Old: old, New: name, Err: err}
	}
	return nil
}
