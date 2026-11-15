//go:build windows

package api

import "testing"

func TestWindowsSystemMemoryUsesActualHost(t *testing.T) {
	info := collectSystemInfo()
	if info.OS != "windows" || info.TotalRAMMB <= 0 || info.FreeRAMMB < 0 || info.FreeRAMMB > info.TotalRAMMB {
		t.Fatalf("invalid Windows memory stats: %+v", info)
	}
	var expected SystemInfo
	readPlatformMemInfo(&expected)
	if info.TotalRAMMB != expected.TotalRAMMB {
		t.Fatal("host capacity mismatch", info, expected)
	}
}
