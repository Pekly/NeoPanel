package trash

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/Pekly/NeoPanel/internal/rooted"
	"golang.org/x/sys/windows"
)

func moveNoReplace(from *rooted.Root, old string, to *rooted.Root, name string) error {
	if !rooted.Basename(old) || !rooted.Basename(name) {
		return os.ErrInvalid
	}
	st, err := from.Lstat(old)
	if err != nil {
		return err
	}
	if st.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("symbolic links are not supported")
	}
	parent, err := from.Root.Open(".")
	if err != nil {
		return err
	}
	defer parent.Close()
	oldNT, err := windows.NewNTUnicodeString(old)
	if err != nil {
		return err
	}
	oa := windows.OBJECT_ATTRIBUTES{RootDirectory: windows.Handle(parent.Fd()), ObjectName: oldNT, Attributes: windows.OBJ_CASE_INSENSITIVE}
	oa.Length = uint32(unsafe.Sizeof(oa))
	var h windows.Handle
	var opened windows.IO_STATUS_BLOCK
	err = windows.NtCreateFile(&h, windows.DELETE|windows.FILE_READ_ATTRIBUTES|windows.SYNCHRONIZE, &oa, &opened, nil, 0,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, windows.FILE_OPEN,
		windows.FILE_OPEN_REPARSE_POINT|windows.FILE_SYNCHRONOUS_IO_NONALERT, 0, 0)
	if err != nil {
		return err
	}
	f := os.NewFile(uintptr(h), old)
	defer f.Close()
	actual, err := f.Stat()
	if err != nil || actual.Mode()&os.ModeSymlink != 0 || !os.SameFile(st, actual) {
		return fmt.Errorf("source changed")
	}
	dest, err := to.Root.Open(".")
	if err != nil {
		return err
	}
	defer dest.Close()
	utf, err := windows.UTF16FromString(name)
	if err != nil {
		return err
	}
	utf = utf[:len(utf)-1]
	type renameInfo struct {
		Replace uint32
		Root    windows.Handle
		Length  uint32
		Name    [1]uint16
	}
	offset := unsafe.Offsetof(renameInfo{}.Name)
	buf := make([]byte, max(int(unsafe.Sizeof(renameInfo{})), int(offset)+2*len(utf)))
	info := (*renameInfo)(unsafe.Pointer(&buf[0]))
	info.Root = windows.Handle(dest.Fd())
	info.Length = uint32(2 * len(utf))
	copy(unsafe.Slice((*uint16)(unsafe.Pointer(&buf[offset])), len(utf)), utf)
	var iosb windows.IO_STATUS_BLOCK
	// FileRenameInformation, ReplaceIfExists=false: atomic no-overwrite for both
	// files and directories, relative to the pinned destination directory handle.
	err = windows.NtSetInformationFile(windows.Handle(h), &iosb, &buf[0], uint32(len(buf)), 10)
	if err != nil {
		if status, ok := err.(windows.NTStatus); ok {
			err = status.Errno()
		}
		return &os.LinkError{Op: "rename", Old: old, New: name, Err: err}
	}
	return nil
}
