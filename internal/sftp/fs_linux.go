//go:build linux

package sftp

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/sys/unix"
)

// Linux never falls back to pathname-based filesystem operations. Every parent
// is opened with O_NOFOLLOW; the root identity is pinned for the SSH connection.
// Directory descriptors survive pathname/symlink replacement without following
// the replacement. This is not isolation from a local process with filesystem
// authority to move open directories outside the jail, install mounts, or create
// hard links to outside files; those require OS ownership/mount isolation.
const directoryFlags = unix.O_RDONLY | unix.O_DIRECTORY | unix.O_NOFOLLOW | unix.O_CLOEXEC
const pinRootHandle = true

func openDirectoryAt(parent int, name string) (*os.File, error) {
	fd, err := unix.Openat(parent, name, directoryFlags, 0)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), name), nil
}

// Start at / and walk even the configured jail ancestors without following
// links. Creation uses mkdirat on the verified parent, then reopens NOFOLLOW.
func walkAbsoluteDirectory(p string, create bool) (*os.File, error) {
	if !filepath.IsAbs(p) {
		return nil, os.ErrPermission
	}
	dir, err := openDirectoryAt(unix.AT_FDCWD, "/")
	if err != nil {
		return nil, err
	}
	for _, part := range strings.Split(filepath.Clean(p), "/") {
		if part == "" {
			continue
		}
		next, err := openDirectoryAt(int(dir.Fd()), part)
		if err == unix.ENOENT && create {
			err = unix.Mkdirat(int(dir.Fd()), part, 0755)
			if err == nil || err == unix.EEXIST {
				next, err = openDirectoryAt(int(dir.Fd()), part)
			}
		}
		dir.Close()
		if err != nil {
			return nil, err
		}
		dir = next
	}
	return dir, nil
}

func prepareJail(p string) error {
	dir, err := walkAbsoluteDirectory(p, true)
	if err != nil {
		return err
	}
	return dir.Close()
}

func (h *chrootHandler) fsRoot() (*os.File, error) {
	dir, err := walkAbsoluteDirectory(h.rootDir, false)
	if err != nil {
		return nil, err
	}
	if h.rootInfo != nil {
		info, err := dir.Stat()
		if err != nil || !os.SameFile(h.rootInfo, info) {
			dir.Close()
			return nil, sftp.ErrSSHFxPermissionDenied
		}
	}
	return dir, nil
}

// fsParent returns an owned descriptor and a single basename (never a path).
// Callers must retain the descriptor until their final *at syscall completes.
func (h *chrootHandler) fsParent(p string) (*os.File, string, error) {
	full, err := h.safePath(p) // lexical validation only
	if err != nil {
		return nil, "", err
	}
	rel, err := filepath.Rel(h.rootDir, full)
	if err != nil {
		return nil, "", err
	}
	dir, err := h.fsRoot()
	if err != nil {
		return nil, "", err
	}
	parts := strings.Split(rel, "/")
	for _, part := range parts[:len(parts)-1] {
		next, err := openDirectoryAt(int(dir.Fd()), part)
		dir.Close()
		if err != nil {
			return nil, "", err
		}
		dir = next
	}
	return dir, parts[len(parts)-1], nil
}

// O_PATH avoids opening devices/FIFOs for metadata; O_NOFOLLOW opens a link
// itself rather than its target, then fstat explicitly rejects that link.
func statAt(dir *os.File, name string) (os.FileInfo, error) {
	fd, err := unix.Openat(int(dir.Fd()), name, unix.O_PATH|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), name)
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, sftp.ErrSSHFxPermissionDenied
	}
	return info, nil
}

func (h *chrootHandler) fsStat(p string) (os.FileInfo, error) {
	dir, name, err := h.fsParent(p)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	return statAt(dir, name)
}

func openRegularAt(dir *os.File, name string, flags int) (*os.File, error) {
	// Delay truncation until the opened object is verified to be a regular file.
	fd, err := unix.Openat(int(dir.Fd()), name, (flags&^os.O_TRUNC)|unix.O_NOFOLLOW|unix.O_CLOEXEC|unix.O_NONBLOCK, 0644)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), name)
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		file.Close()
		return nil, sftp.ErrSSHFxPermissionDenied
	}
	if flags&os.O_TRUNC != 0 {
		if err := file.Truncate(0); err != nil {
			file.Close()
			return nil, err
		}
	}
	return file, nil
}

func (h *chrootHandler) fsOpen(p string, flags int) (*os.File, error) {
	dir, name, err := h.fsParent(p)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	return openRegularAt(dir, name, flags)
}

func (h *chrootHandler) fsList(p string) ([]os.FileInfo, error) {
	parent, name, err := h.fsParent(p)
	if err != nil {
		return nil, err
	}
	defer parent.Close()
	dir, err := openDirectoryAt(int(parent.Fd()), name)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	// Do not use Readdir/DirEntry.Info: those may stat using File.Name(), not dirfd.
	names, err := dir.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	infos := make([]os.FileInfo, 0, len(names))
	for _, name := range names {
		info, err := statAt(dir, name)
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func (h *chrootHandler) fsCommand(r *sftp.Request) error {
	dir, name, err := h.fsParent(r.Filepath)
	if err != nil {
		return err
	}
	defer dir.Close()
	if name == "." {
		return sftp.ErrSSHFxPermissionDenied
	}
	switch r.Method {
	case "Mkdir":
		return unix.Mkdirat(int(dir.Fd()), name, 0755)
	case "Rename", "PosixRename":
		targetDir, targetName, err := h.fsParent(r.Target)
		if err != nil {
			return err
		}
		defer targetDir.Close()
		if targetName == "." {
			return sftp.ErrSSHFxPermissionDenied
		}
		if _, err := statAt(dir, name); err != nil {
			return err
		}
		if _, err := statAt(targetDir, targetName); err != nil && err != unix.ENOENT {
			return err
		}
		// Neither renameat nor unlinkat follows a final symlink, including a link
		// swapped in after statAt. Parent descriptors cannot be redirected by links.
		return unix.Renameat(int(dir.Fd()), name, int(targetDir.Fd()), targetName)
	case "Remove", "Rmdir":
		info, err := statAt(dir, name)
		if err != nil {
			return err
		}
		if info.IsDir() != (r.Method == "Rmdir") {
			return sftp.ErrSSHFxPermissionDenied
		}
		flags := 0
		if r.Method == "Rmdir" {
			flags = unix.AT_REMOVEDIR
		}
		return unix.Unlinkat(int(dir.Fd()), name, flags)
	case "Setstat":
		flags, attrs := r.AttrFlags(), r.Attributes()
		if attrs == nil {
			return nil
		}
		if flags.UidGid {
			return sftp.ErrSSHFxPermissionDenied
		}
		// Open once and change metadata through that descriptor, never by pathname.
		mode := unix.O_RDONLY
		if flags.Size {
			mode = unix.O_WRONLY
		}
		fd, err := unix.Openat(int(dir.Fd()), name, mode|unix.O_NOFOLLOW|unix.O_CLOEXEC|unix.O_NONBLOCK, 0)
		if err != nil {
			return err
		}
		file := os.NewFile(uintptr(fd), name)
		defer file.Close()
		info, err := file.Stat()
		if err != nil {
			return err
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return sftp.ErrSSHFxPermissionDenied
		}
		if flags.Size {
			if !info.Mode().IsRegular() {
				return sftp.ErrSSHFxPermissionDenied
			}
			if err := file.Truncate(int64(attrs.Size)); err != nil {
				return err
			}
		}
		if flags.Permissions {
			if err := file.Chmod(attrs.FileMode().Perm()); err != nil {
				return err
			}
		}
		if flags.Acmodtime {
			return unix.Futimes(fd, []unix.Timeval{unix.NsecToTimeval(int64(attrs.Atime) * 1e9), unix.NsecToTimeval(int64(attrs.Mtime) * 1e9)})
		}
		return nil
	case "Symlink", "Link":
		return sftp.ErrSSHFxPermissionDenied
	default:
		return sftp.ErrSSHFxOpUnsupported
	}
}
