package sftp

import (
	"os"
	"time"

	"golang.org/x/sys/windows"
)

const portableOpenFlags = 0

var portableReOpenFile = windows.NewLazySystemDLL("kernel32.dll").NewProc("ReOpenFile")

func portableFileTimes(f *os.File, atime, mtime time.Time) error {
	// ReOpenFile addresses the existing object, not its mutable pathname. Request
	// only timestamp access rather than requiring file-data write permission.
	h, _, err := portableReOpenFile.Call(uintptr(f.Fd()), windows.FILE_WRITE_ATTRIBUTES, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, windows.FILE_FLAG_BACKUP_SEMANTICS)
	if windows.Handle(h) == windows.InvalidHandle {
		return err
	}
	defer windows.CloseHandle(windows.Handle(h))
	a, m := windows.NsecToFiletime(atime.UnixNano()), windows.NsecToFiletime(mtime.UnixNano())
	return windows.SetFileTime(windows.Handle(h), nil, &a, &m)
}
