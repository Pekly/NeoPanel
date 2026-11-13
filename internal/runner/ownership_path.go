package runner

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Keep the lock inode outside the server's writable sandbox and outside the
// directory swapped by restore/reinstall. All panel processes derive the same
// key, including symlink aliases of an existing server directory.
func serverRunLockPath(dir string) (string, error) {
	canonical, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(canonical); err == nil {
		canonical = resolved
	} else if !os.IsNotExist(err) {
		return "", err
	} else {
		parent, err := filepath.EvalSymlinks(filepath.Dir(canonical))
		if err != nil {
			return "", err
		}
		canonical = filepath.Join(parent, filepath.Base(canonical))
	}
	lockDir := filepath.Join(filepath.Dir(canonical), ".neopanel-locks")
	if err := os.MkdirAll(lockDir, 0700); err != nil {
		return "", err
	}
	key := canonical
	if runtime.GOOS == "windows" {
		key = strings.ToLower(key)
	}
	return filepath.Join(lockDir, fmt.Sprintf("%x.lock", sha256.Sum256([]byte(key)))), nil
}
