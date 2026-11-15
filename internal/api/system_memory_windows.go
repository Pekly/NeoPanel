//go:build windows

package api

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var globalMemoryStatusEx = windows.NewLazySystemDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx")

type windowsMemoryStatus struct {
	Length, Load                                             uint32
	TotalPhysical, AvailablePhysical                         uint64
	TotalPageFile, AvailablePageFile                         uint64
	TotalVirtual, AvailableVirtual, AvailableExtendedVirtual uint64
}

func readPlatformMemInfo(info *SystemInfo) {
	status := windowsMemoryStatus{}
	status.Length = uint32(unsafe.Sizeof(status))
	ok, _, _ := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&status)))
	if ok == 0 {
		return
	} // Zero means unavailable, never invented capacity.
	info.TotalRAMMB = int64(status.TotalPhysical / (1024 * 1024))
	info.FreeRAMMB = int64(status.AvailablePhysical / (1024 * 1024))
}
