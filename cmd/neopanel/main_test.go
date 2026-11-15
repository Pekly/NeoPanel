package main

import (
	"io/fs"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/web"
)

func TestConfigurationRolePaths(t *testing.T) {
	if got := resolveConfigPathForRole(" custom.yaml ", "panel"); got != "custom.yaml" {
		t.Fatal(got)
	}
	if runtime.GOOS != "windows" {
		return
	}
	t.Setenv("NEOPANEL_DATA_DIR", "")
	t.Setenv("LOCALAPPDATA", t.TempDir())
	panel, node := resolveConfigPathForRole("", "panel"), resolveConfigPathForRole("", "node")
	if panel == node || !filepath.IsAbs(panel) || !filepath.IsAbs(node) {
		t.Fatal(panel, node)
	}
	if panel != filepath.Join(config.DefaultConfigForRole("panel").Paths.DataDir, "config.yaml") {
		t.Fatal(panel)
	}
}

func TestWeblessBuildContainsNoPanelAssets(t *testing.T) {
	if web.NodeOnlyBuild {
		if web.GetFS() != nil {
			t.Fatal("node build contains web UI")
		}
	} else if _, err := fs.Stat(web.GetFS(), "index.html"); err != nil {
		t.Fatal("panel build missing browser UI", err)
	}
}
