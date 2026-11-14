package runner

import (
	"os"
	"path/filepath"
	"syscall"
)

func hostManagedJavaPath(name string) bool {
	for {
		st, e := os.Lstat(name)
		if e != nil || st.Mode()&os.ModeSymlink != 0 || st.Mode().Perm()&0022 != 0 {
			return false
		}
		sys, ok := st.Sys().(*syscall.Stat_t)
		if !ok || sys.Uid != 0 {
			return false
		}
		parent := filepath.Dir(name)
		if parent == name {
			return true
		}
		name = parent
	}
}
