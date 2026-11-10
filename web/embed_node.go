//go:build node

package web

import "io/fs"

// GetFS returns nil in headless node builds so no web panel assets are bundled or served.
func GetFS() fs.FS {
	return nil
}
