//go:build node

package web

import "io/fs"

// GetFS returns nil in cluster node builds so resources stay focused on server workloads.
func GetFS() fs.FS {
	return nil
}
