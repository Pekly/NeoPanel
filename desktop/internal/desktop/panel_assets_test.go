package desktop

import (
	"github.com/Pekly/NeoPanel/web"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

func TestDesktopAssetsInjectReceiverBeforeApp(t *testing.T) {
	original := web.GetFS()
	assets := desktopPanelAssets{original}
	data, err := fs.ReadFile(assets, "index.html")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	receiver, app := strings.Index(text, `src="/js/desktop-session.js"`), strings.Index(text, `src="/js/app.js"`)
	if receiver < 0 || receiver >= app {
		t.Fatal("receiver must load before application")
	}
	info, err := fs.Stat(assets, "index.html")
	if err != nil || info.Size() != int64(len(data)) {
		t.Fatal("incorrect wrapped asset size", err)
	}
	if _, err := fs.ReadFile(assets, "js/desktop-session.js"); err != nil {
		t.Fatal(err)
	}
	normal, _ := fs.ReadFile(original, "index.html")
	if strings.Contains(string(normal), "desktop-session.js") {
		t.Fatal("ordinary web assets changed")
	}
	if _, err := fs.ReadFile(original, "js/desktop-session.js"); err == nil {
		t.Fatal("CLI serves desktop receiver")
	}
}
func TestDesktopAssetsRejectUnexpectedStartupDocument(t *testing.T) {
	assets := desktopPanelAssets{fstest.MapFS{"index.html": {Data: []byte("unexpected document")}}}
	if _, err := assets.Open("index.html"); err == nil {
		t.Fatal("missing startup anchor accepted")
	}
}
