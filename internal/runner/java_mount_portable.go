//go:build !linux

package runner

// Bubblewrap is Linux-only. Never infer host mounts on other platforms.
func hostManagedJavaPath(string) bool { return false }
