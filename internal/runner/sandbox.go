package runner

import "path/filepath"

// Mount the resolver file, never its parent (/etc on many hosts).
func sandboxResolverArgs(path string) []string {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil
	}
	return []string{"--ro-bind-try", resolved, "/etc/resolv.conf"}
}
