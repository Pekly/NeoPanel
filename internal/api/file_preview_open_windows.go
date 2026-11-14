package api

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

func openPreviewPath(base, rel string, flags int) (*os.File, error) {
	// Pin every parent against rename/deletion, and inspect reparse points without
	// following them. Reject Windows device names and alternate data streams too.
	parts := strings.Split(rel, string(filepath.Separator))
	for _, part := range parts {
		if strings.ContainsAny(part, ":\x00") || strings.TrimRight(part, ". ") != part {
			return nil, fmt.Errorf("Unsafe Windows file name")
		}
		stem := strings.ToUpper(strings.SplitN(part, ".", 2)[0])
		if stem == "CON" || stem == "PRN" || stem == "AUX" || stem == "NUL" || strings.HasPrefix(stem, "COM") && len(stem) == 4 || strings.HasPrefix(stem, "LPT") && len(stem) == 4 {
			return nil, fmt.Errorf("Device names are not supported")
		}
	}
	var parents []windows.Handle
	defer func() {
		for _, h := range parents {
			windows.CloseHandle(h)
		}
	}()
	current := base
	for i := 0; i < len(parts); i++ {
		p, err := windows.UTF16PtrFromString(current)
		if err != nil {
			return nil, err
		}
		h, err := windows.CreateFile(p, windows.FILE_READ_ATTRIBUTES, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
		if err != nil {
			return nil, err
		}
		parents = append(parents, h)
		var info windows.ByHandleFileInformation
		if err := windows.GetFileInformationByHandle(h, &info); err != nil {
			return nil, err
		}
		if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 {
			return nil, fmt.Errorf("Symbolic links and non-directory parents are not supported")
		}
		current = filepath.Join(current, parts[i])
	}
	p, err := windows.UTF16PtrFromString(current)
	if err != nil {
		return nil, err
	}
	access := uint32(windows.GENERIC_READ)
	if flags&os.O_RDWR != 0 {
		access |= windows.GENERIC_WRITE
	}
	disposition := uint32(windows.OPEN_EXISTING)
	if flags&os.O_CREATE != 0 {
		disposition = windows.CREATE_NEW
	}
	h, err := windows.CreateFile(p, access, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, disposition, windows.FILE_FLAG_OPEN_REPARSE_POINT|windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return nil, err
	}
	var info windows.ByHandleFileInformation
	err = windows.GetFileInformationByHandle(h, &info)
	if err != nil || info.FileAttributes&(windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_DIRECTORY) != 0 {
		windows.CloseHandle(h)
		return nil, fmt.Errorf("Only regular files without symbolic links are supported")
	}
	return os.NewFile(uintptr(h), current), nil
}
