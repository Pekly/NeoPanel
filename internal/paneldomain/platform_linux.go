//go:build linux

package paneldomain

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	"golang.org/x/sys/unix"
)

func supportedPlatform() error {
	if os.Geteuid() != 0 {
		return errors.New("automatic Caddy provisioning requires Linux root")
	}
	return nil
}

func ownership(info os.FileInfo) (int, int, uint64) {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return -1, -1, 0
	}
	return int(st.Uid), int(st.Gid), uint64(st.Nlink)
}

func checkInfo(info os.FileInfo, kind pathKind) error {
	uid, _, links := ownership(info)
	return checkSafety(info.Mode(), uid, links, kind)
}

func safePath(path string, kind pathKind) error {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return errors.New("unsafe noncanonical path")
	}
	for current, k := path, kind; ; current, k = filepath.Dir(current), directoryPath {
		info, err := os.Lstat(current)
		if err != nil {
			return fmt.Errorf("cannot inspect %s: %w", current, err)
		}
		if err = checkInfo(info, k); err != nil {
			return fmt.Errorf("unsafe %s: %w", current, err)
		}
		if current == "/" {
			break
		}
	}
	return nil
}

func openNoFollow(path string) (*os.File, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), path), nil
}

func preserveMetadata(f *os.File, original fileRecord) error {
	if err := f.Chown(original.uid, original.gid); err != nil {
		return err
	}
	return f.Chmod(original.info.Mode().Perm())
}

func syncDirectory(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}

func caddyIdentity(cmd *exec.Cmd) error {
	u, err := user.Lookup("caddy")
	if err != nil {
		return errors.New("cannot resolve packaged caddy service account")
	}
	uid, e1 := strconv.ParseUint(u.Uid, 10, 32)
	g, err := user.LookupGroup("caddy")
	if err != nil {
		return errors.New("cannot resolve packaged caddy service group")
	}
	gid, e2 := strconv.ParseUint(g.Gid, 10, 32)
	if e1 != nil || e2 != nil || uid == 0 || gid == 0 || u.HomeDir != "/var/lib/caddy" {
		return errors.New("unsupported caddy service identity/home")
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}}
	return nil
}
