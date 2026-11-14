//go:build !windows

package api

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

func openPreviewPath(base, rel string, flags int) (*os.File, error) {
	dir, err := unix.Open(base, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	defer func() { unix.Close(dir) }()
	parts := strings.Split(rel, string(filepath.Separator))
	for _, part := range parts[:len(parts)-1] {
		next, err := unix.Openat(dir, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		if err != nil {
			return nil, err
		}
		unix.Close(dir)
		dir = next
	}
	name := parts[len(parts)-1]
	var st unix.Stat_t
	err = unix.Fstatat(dir, name, &st, unix.AT_SYMLINK_NOFOLLOW)
	if err == nil && st.Mode&unix.S_IFMT != unix.S_IFREG {
		return nil, fmt.Errorf("Only regular files are supported")
	}
	if err != nil && !(err == unix.ENOENT && flags&os.O_CREATE != 0) {
		return nil, err
	}
	fd, err := unix.Openat(dir, name, flags|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0644)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), filepath.Join(base, rel)), nil
}
