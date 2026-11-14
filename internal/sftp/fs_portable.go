//go:build !linux

package sftp

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pkg/sftp"
)

// Keep connection identity, but pin roots only for individual operations.
// Windows root handles can prevent renaming for the lifetime of the handle.
const pinRootHandle = false

// As on Linux, hardlinks, mounts, or moving an opened directory out of the jail
// require OS ownership/mount isolation. No operation falls back to raw paths.
func portableAllowed(info os.FileInfo) bool {
	return info.Mode().IsRegular() || info.IsDir() && info.Mode()&(os.ModeSymlink|os.ModeIrregular) == 0
}

// Pin one directory at a time. The lstat policy is checked against the opened
// object's identity; confinement itself is provided by Root.OpenRoot.
func portableSub(root *os.Root, name string) (*os.Root, error) {
	info, err := root.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || !portableAllowed(info) {
		return nil, sftp.ErrSSHFxPermissionDenied
	}
	next, err := root.OpenRoot(name)
	if err != nil {
		return nil, err
	}
	actual, err := next.Stat(".")
	if err != nil || !os.SameFile(info, actual) {
		next.Close()
		return nil, sftp.ErrSSHFxPermissionDenied
	}
	return next, nil
}

// Even configured ancestors are traversed through pinned roots, never by a
// check-then-open absolute pathname. Volume roots are trusted configuration.
func portableAbsoluteRoot(p string, create bool) (*os.Root, error) {
	if !filepath.IsAbs(p) {
		return nil, os.ErrPermission
	}
	p = filepath.Clean(p)
	volume := filepath.VolumeName(p)
	root, err := os.OpenRoot(volume + string(filepath.Separator))
	if err != nil {
		return nil, err
	}
	for _, part := range strings.Split(strings.TrimPrefix(p, volume), string(filepath.Separator)) {
		if part == "" {
			continue
		}
		if create {
			if err := root.Mkdir(part, 0755); err != nil && !os.IsExist(err) {
				root.Close()
				return nil, err
			}
		}
		next, err := portableSub(root, part)
		root.Close()
		if err != nil {
			return nil, err
		}
		root = next
	}
	return root, nil
}
func prepareJail(p string) error {
	root, err := portableAbsoluteRoot(p, true)
	if err != nil {
		return err
	}
	return root.Close()
}

func (h *chrootHandler) portableRoot() (*os.Root, error) {
	root, err := portableAbsoluteRoot(h.rootDir, false)
	if err != nil {
		return nil, err
	}
	info, err := root.Stat(".")
	if err != nil || !info.IsDir() || h.rootInfo != nil && !os.SameFile(h.rootInfo, info) {
		root.Close()
		return nil, sftp.ErrSSHFxPermissionDenied
	}
	return root, nil
}

// Used only to establish the connection's immutable root identity.
func (h *chrootHandler) fsRoot() (*os.File, error) {
	root, err := h.portableRoot()
	if err != nil {
		return nil, err
	}
	defer root.Close()
	return root.Open(".")
}

func (h *chrootHandler) portableRelative(p string) (string, error) {
	full, err := h.safePath(p)
	if err != nil {
		return "", err
	}
	return filepath.Rel(h.rootDir, full)
}
func portableParent(root *os.Root, rel string) (*os.Root, string, error) {
	parent, err := root.OpenRoot(".")
	if err != nil {
		return nil, "", err
	}
	parts := strings.Split(rel, string(filepath.Separator))
	for _, part := range parts[:len(parts)-1] {
		next, err := portableSub(parent, part)
		parent.Close()
		if err != nil {
			return nil, "", err
		}
		parent = next
	}
	return parent, parts[len(parts)-1], nil
}
func portableStat(root *os.Root, name string) (os.FileInfo, error) {
	info, err := root.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !portableAllowed(info) {
		return nil, sftp.ErrSSHFxPermissionDenied
	}
	return info, nil
}
func (h *chrootHandler) fsStat(p string) (os.FileInfo, error) {
	rel, err := h.portableRelative(p)
	if err != nil {
		return nil, err
	}
	root, err := h.portableRoot()
	if err != nil {
		return nil, err
	}
	defer root.Close()
	parent, name, err := portableParent(root, rel)
	if err != nil {
		return nil, err
	}
	defer parent.Close()
	return portableStat(parent, name)
}

// Never pass O_TRUNC until the opened handle has been verified. Creating a
// missing file uses O_EXCL, so a concurrently planted final link cannot win.
func portableOpen(root *os.Root, name string, flags int, allowDir bool) (*os.File, error) {
	before, err := portableStat(root, name)
	missing := os.IsNotExist(err)
	if err != nil && !(missing && flags&os.O_CREATE != 0) {
		return nil, err
	}
	if !missing && !before.Mode().IsRegular() && !allowDir {
		return nil, sftp.ErrSSHFxPermissionDenied
	}
	openFlags := flags &^ os.O_TRUNC
	if missing {
		openFlags |= os.O_EXCL
	}
	f, err := root.OpenFile(name, openFlags|portableOpenFlags, 0644)
	if err != nil {
		return nil, err
	}
	actual, err := f.Stat()
	if err != nil || !portableAllowed(actual) || !allowDir && !actual.Mode().IsRegular() || !missing && !os.SameFile(before, actual) {
		f.Close()
		return nil, sftp.ErrSSHFxPermissionDenied
	}
	if flags&os.O_TRUNC != 0 {
		if err := f.Truncate(0); err != nil {
			f.Close()
			return nil, err
		}
	}
	return f, nil
}
func (h *chrootHandler) fsOpen(p string, flags int) (*os.File, error) {
	rel, err := h.portableRelative(p)
	if err != nil {
		return nil, err
	}
	root, err := h.portableRoot()
	if err != nil {
		return nil, err
	}
	defer root.Close()
	parent, name, err := portableParent(root, rel)
	if err != nil {
		return nil, err
	}
	defer parent.Close()
	return portableOpen(parent, name, flags, false)
}
func (h *chrootHandler) fsList(p string) ([]os.FileInfo, error) {
	rel, err := h.portableRelative(p)
	if err != nil {
		return nil, err
	}
	root, err := h.portableRoot()
	if err != nil {
		return nil, err
	}
	defer root.Close()
	parent, name, err := portableParent(root, rel)
	if err != nil {
		return nil, err
	}
	defer parent.Close()
	dir, err := portableSub(parent, name)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	f, err := dir.Open(".")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	infos := make([]os.FileInfo, 0, len(names))
	for _, name := range names {
		info, err := portableStat(dir, name)
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return infos, nil
}
func (h *chrootHandler) fsCommand(r *sftp.Request) error {
	rel, err := h.portableRelative(r.Filepath)
	if err != nil {
		return err
	}
	if rel == "." {
		return sftp.ErrSSHFxPermissionDenied
	}
	root, err := h.portableRoot()
	if err != nil {
		return err
	}
	defer root.Close()
	parent, name, err := portableParent(root, rel)
	if err != nil {
		return err
	}
	defer parent.Close()
	switch r.Method {
	case "Rename", "PosixRename":
		target, err := h.portableRelative(r.Target)
		if err != nil {
			return err
		}
		if target == "." {
			return sftp.ErrSSHFxPermissionDenied
		}
		dest, base, err := portableParent(root, target)
		if err != nil {
			return err
		}
		defer dest.Close()
		if _, err := portableStat(parent, name); err != nil {
			return err
		}
		if _, err := portableStat(dest, base); err != nil && !os.IsNotExist(err) {
			return err
		}
		// Root.Rename confines both paths to the same verified server root even if
		// parents change after the policy check. It never dereferences final links.
		return root.Rename(rel, target)
	case "Remove", "Rmdir":
		info, err := portableStat(parent, name)
		if err != nil {
			return err
		}
		if info.IsDir() != (r.Method == "Rmdir") {
			return sftp.ErrSSHFxPermissionDenied
		}
		return parent.Remove(name)
	case "Mkdir":
		return parent.Mkdir(name, 0755)
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
		info, err := portableStat(parent, name)
		if err != nil {
			return err
		}
		if info.IsDir() {
			if flags.Size {
				return sftp.ErrSSHFxPermissionDenied
			}
			dir, err := portableSub(parent, name)
			if err != nil {
				return err
			}
			defer dir.Close()
			// The name "." addresses the pinned directory itself, not a mutable entry.
			if flags.Permissions {
				if err := dir.Chmod(".", attrs.FileMode().Perm()); err != nil {
					return err
				}
			}
			if flags.Acmodtime {
				return dir.Chtimes(".", time.Unix(int64(attrs.Atime), 0), time.Unix(int64(attrs.Mtime), 0))
			}
			return nil
		}
		openFlags := os.O_RDONLY
		if flags.Size {
			openFlags = os.O_WRONLY
		}
		f, err := portableOpen(parent, name, openFlags, false)
		if err != nil {
			return err
		}
		defer f.Close()
		if flags.Size {
			if attrs.Size > 1<<63-1 {
				return sftp.ErrSSHFxPermissionDenied
			}
			if err := f.Truncate(int64(attrs.Size)); err != nil {
				return err
			}
		}
		if flags.Permissions {
			if err := f.Chmod(attrs.FileMode().Perm()); err != nil {
				return err
			}
		}
		if flags.Acmodtime {
			return portableFileTimes(f, time.Unix(int64(attrs.Atime), 0), time.Unix(int64(attrs.Mtime), 0))
		}
		return nil
	default:
		return sftp.ErrSSHFxOpUnsupported
	}
}
