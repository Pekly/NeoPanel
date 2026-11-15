//go:build !node

package web

import (
	"embed"
	"io/fs"
)

const NodeOnlyBuild = false

//go:embed static
var staticFS embed.FS

// GetFS returns the filesystem rooted at static/
func GetFS() fs.FS {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return sub
}
