//go:build !linux && !windows

package sftp

import (
	"golang.org/x/sys/unix"
	"os"
	"time"
)

const portableOpenFlags = unix.O_NONBLOCK | unix.O_NOFOLLOW

func portableFileTimes(f *os.File, atime, mtime time.Time) error {
	return unix.Futimes(int(f.Fd()), []unix.Timeval{unix.NsecToTimeval(atime.UnixNano()), unix.NsecToTimeval(mtime.UnixNano())})
}
