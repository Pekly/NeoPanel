package desktop

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
)

//go:embed panel-session.js
var sessionAssets embed.FS

type desktopPanelAssets struct{ fs.FS }

// Only the desktop serves the session receiver. Static assets never contain a
// credential: the native shell supplies it to its initial frame in memory.
func (a desktopPanelAssets) Open(name string) (fs.File, error) {
	if name == "js/desktop-session.js" {
		return sessionAssets.Open("panel-session.js")
	}
	if name != "index.html" {
		return a.FS.Open(name)
	}
	data, err := fs.ReadFile(a.FS, name)
	if err != nil {
		return nil, err
	}
	anchor := []byte(`<script src="/js/app.js"></script>`)
	if bytes.Count(data, anchor) != 1 {
		return nil, fmt.Errorf("desktop panel startup anchor missing")
	}
	data = bytes.Replace(data, anchor, []byte(`<script src="/js/desktop-session.js"></script>`+"\n  "+string(anchor)), 1)
	info, err := fs.Stat(a.FS, name)
	if err != nil {
		return nil, err
	}
	return &panelAssetFile{Reader: bytes.NewReader(data), info: panelAssetInfo{FileInfo: info, size: int64(len(data))}}, nil
}

type panelAssetInfo struct {
	fs.FileInfo
	size int64
}

func (i panelAssetInfo) Size() int64 { return i.size }

type panelAssetFile struct {
	*bytes.Reader
	info fs.FileInfo
}

func (f *panelAssetFile) Stat() (fs.FileInfo, error) { return f.info, nil }
func (f *panelAssetFile) Close() error               { return nil }
