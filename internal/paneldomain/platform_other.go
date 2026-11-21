//go:build !linux

package paneldomain

import (
	"errors"
	"os"
	"os/exec"
)

func supportedPlatform() error {
	return errors.New("automatic Caddy provisioning is available only on Linux")
}
func safePath(string, pathKind) error          { return supportedPlatform() }
func caddyIdentity(*exec.Cmd) error            { return supportedPlatform() }
func ownership(os.FileInfo) (int, int, uint64) { return -1, -1, 1 }

// These helpers also allow unprivileged transaction tests on non-Linux hosts.
// Production always fails supportedPlatform before reaching filesystem operations.
func openNoFollow(path string) (*os.File, error) { return os.Open(path) }
func preserveMetadata(f *os.File, original fileRecord) error {
	return f.Chmod(original.info.Mode().Perm())
}
func syncDirectory(string) error { return supportedPlatform() }
