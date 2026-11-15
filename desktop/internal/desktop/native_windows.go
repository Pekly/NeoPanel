//go:build windows && !node

package desktop

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var user32 = windows.NewLazySystemDLL("user32.dll")

// AcquireInstance holds a named kernel object before any database or WebView is opened.
// A second launch focuses the existing window instead of starting a second backend.
func AcquireInstance(id string) (release func(), acquired bool, err error) {
	name, err := windows.UTF16PtrFromString(`Local\` + id)
	if err != nil {
		return nil, false, err
	}
	handle, err := windows.CreateMutex(nil, false, name)
	if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		if handle != 0 {
			_ = windows.CloseHandle(handle)
		}
		focusWindow(id)
		return func() {}, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("single-instance lock: %w", err)
	}
	return func() { _ = windows.CloseHandle(handle) }, true, nil
}

func focusWindow(id string) {
	class, _ := windows.UTF16PtrFromString(id)
	hwnd, _, _ := user32.NewProc("FindWindowW").Call(uintptr(unsafe.Pointer(class)), 0)
	if hwnd != 0 {
		minimized, _, _ := user32.NewProc("IsIconic").Call(hwnd)
		if minimized != 0 {
			user32.NewProc("ShowWindow").Call(hwnd, 9)
		}
		user32.NewProc("SetForegroundWindow").Call(hwnd)
	}
}

func ShowError(message string) {
	text, _ := windows.UTF16PtrFromString(message)
	title, _ := windows.UTF16PtrFromString(Title)
	_, _ = windows.MessageBox(0, text, title, windows.MB_OK|windows.MB_ICONERROR)
}

func openLocalPath(path string) error {
	value, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	return windows.ShellExecute(0, nil, value, nil, nil, windows.SW_SHOWNORMAL)
}
