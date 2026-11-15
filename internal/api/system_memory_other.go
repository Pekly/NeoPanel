//go:build !windows

package api

// Linux uses /proc/meminfo. Unsupported hosts report unavailable rather than fake RAM.
func readPlatformMemInfo(info *SystemInfo) {}
