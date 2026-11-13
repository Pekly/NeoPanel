//go:build !linux

package sftp

import (
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
)

// Windows directory handles prevent renaming; retain identity but do not pin.
const pinRootHandle = false

// NON-LINUX LIMITATION (including Windows): these are checked pathname-based
// operations, not atomic descriptor-relative confinement. An external process
// racing a symlink/junction/ancestor replacement can defeat the checks. Protect
// the servers tree from untrusted local writers; use Linux for race-safe walks.
func rejectSymlinks(full string) error {
	parent := filepath.Dir(full)
	if parent != full {
		if err := rejectSymlinks(parent); err != nil {
			return err
		}
	}
	info, err := os.Lstat(full)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&(os.ModeSymlink|os.ModeIrregular) != 0 {
		return sftp.ErrSSHFxPermissionDenied
	}
	return nil
}

func prepareJail(p string) error {
	if err := rejectSymlinks(p); err != nil {
		return err
	}
	return os.MkdirAll(p, 0755)
}

func (h *chrootHandler) fsRoot() (*os.File, error) {
	if err := rejectSymlinks(h.rootDir); err != nil {
		return nil, err
	}
	dir, err := os.Open(h.rootDir)
	if err != nil {
		return nil, err
	}
	info, err := dir.Stat()
	if err != nil || !info.IsDir() || (h.rootInfo != nil && !os.SameFile(h.rootInfo, info)) {
		dir.Close()
		return nil, sftp.ErrSSHFxPermissionDenied
	}
	return dir, nil
}

func (h *chrootHandler) portablePath(p string) (string, error) {
	full, err := h.safePath(p)
	if err != nil {
		return "", err
	}
	root, err := h.fsRoot()
	if err != nil {
		return "", err
	}
	root.Close()
	if err := rejectSymlinks(full); err != nil {
		return "", err
	}
	return full, nil
}

func (h *chrootHandler) fsStat(p string) (os.FileInfo, error) {
	full, err := h.portablePath(p)
	if err != nil {
		return nil, err
	}
	return os.Lstat(full)
}

func (h *chrootHandler) fsOpen(p string, flags int) (*os.File, error) {
	full, err := h.portablePath(p)
	if err != nil {
		return nil, err
	}
	if info, err := os.Stat(full); err == nil && !info.Mode().IsRegular() {
		return nil, sftp.ErrSSHFxPermissionDenied
	}
	return os.OpenFile(full, flags, 0644)
}

func (h *chrootHandler) fsList(p string) ([]os.FileInfo, error) {
	full, err := h.portablePath(p)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(full)
	if err != nil {
		return nil, err
	}
	var infos []os.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		if info.Mode()&(os.ModeSymlink|os.ModeIrregular) != 0 {
			return nil, sftp.ErrSSHFxPermissionDenied
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func (h *chrootHandler) fsCommand(r *sftp.Request) error {
	full, err := h.portablePath(r.Filepath)
	if err != nil {
		return err
	}
	if full == h.rootDir {
		return sftp.ErrSSHFxPermissionDenied
	}
	switch r.Method {
	case "Rename", "PosixRename":
		target, err := h.portablePath(r.Target)
		if err != nil {
			return err
		}
		if target == h.rootDir {
			return sftp.ErrSSHFxPermissionDenied
		}
		return os.Rename(full, target)
	case "Remove", "Rmdir":
		info, err := os.Lstat(full)
		if err != nil {
			return err
		}
		if info.IsDir() != (r.Method == "Rmdir") {
			return sftp.ErrSSHFxPermissionDenied
		}
		return os.Remove(full)
	case "Mkdir":
		return os.Mkdir(full, 0755)
	case "Symlink", "Link":
		return sftp.ErrSSHFxPermissionDenied
	case "Setstat":
		flags, attrs := r.AttrFlags(), r.Attributes()
		if attrs == nil {
			return nil
		}
		if flags.UidGid {
			return sftp.ErrSSHFxPermissionDenied
		}
		if flags.Size {
			if err := os.Truncate(full, int64(attrs.Size)); err != nil {
				return err
			}
		}
		if flags.Permissions {
			if err := os.Chmod(full, attrs.FileMode().Perm()); err != nil {
				return err
			}
		}
		if flags.Acmodtime {
			return os.Chtimes(full, time.Unix(int64(attrs.Atime), 0), time.Unix(int64(attrs.Mtime), 0))
		}
		return nil
	default:
		return sftp.ErrSSHFxOpUnsupported
	}
}
