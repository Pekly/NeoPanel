package paneldomain

import (
	"errors"
	"os"
)

// Kept platform independent so permission/link policies can be tested on Windows.
func checkSafety(mode os.FileMode, uid int, links uint64, kind pathKind) error {
	if uid != 0 {
		return errors.New("must be owned by root")
	}
	if mode.Perm()&0022 != 0 {
		return errors.New("must not be group- or world-writable")
	}
	if mode&(os.ModeSymlink|os.ModeSetuid|os.ModeSetgid|os.ModeSticky) != 0 {
		return errors.New("symlinks and special mode bits are unsupported")
	}
	if kind == directoryPath {
		if !mode.IsDir() {
			return errors.New("expected directory")
		}
		return nil
	}
	if !mode.IsRegular() || links != 1 {
		return errors.New("expected regular single-link file")
	}
	if kind == executablePath && mode.Perm()&0111 == 0 {
		return errors.New("expected executable")
	}
	return nil
}
