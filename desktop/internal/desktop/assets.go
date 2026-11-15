package desktop

import (
	"embed"
	"io/fs"
)

//go:embed shell/*
var bundledShell embed.FS

func shellAssets() fs.FS {
	assets, err := fs.Sub(bundledShell, "shell")
	if err != nil {
		panic(err)
	}
	return assets
}
