package downloader

import (
	"github.com/Pekly/NeoPanel/internal/rooted"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJavaDownloadTraversalRejectedBeforeHTTP(t *testing.T) {
	old := httpClient
	t.Cleanup(func() { httpClient = old })
	requests := 0
	httpClient = &http.Client{Transport: modrinthTransport(func(r *http.Request) (*http.Response, error) {
		requests++
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("jar"))}, nil
	})}
	parent := t.TempDir()
	a := filepath.Join(parent, "A")
	b := filepath.Join(parent, "B")
	os.Mkdir(a, 0755)
	os.Mkdir(b, 0755)
	victim := filepath.Join(b, "server.jar")
	os.WriteFile(victim, []byte("secret"), 0600)
	for _, download := range []func(string, string, string, string) error{DownloadServerJar, DownloadWithMCJarFiles} {
		for _, v := range []string{"../../../B/server", `..\..\..\B\server`} {
			if e := download("bungeecord", v, a, "bungeecord-"+v+".jar"); e == nil {
				t.Fatal("accepted traversal", v)
			}
		}
		for _, name := range []string{"../B/server.jar", `..\B\server.jar`} {
			if e := download("paper", "latest", a, name); e == nil {
				t.Fatal("accepted filename", name)
			}
		}
	}
	if requests != 0 {
		t.Fatal("invalid input performed network requests", requests)
	}
	got, _ := os.ReadFile(victim)
	if string(got) != "secret" {
		t.Fatal("sibling overwritten")
	}
}
func TestJavaDownloadHardlinkAndEULASymlink(t *testing.T) {
	old := httpClient
	t.Cleanup(func() { httpClient = old })
	httpClient = &http.Client{Transport: modrinthTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("jar"))}, nil
	})}
	for _, download := range []func(string, string, string, string) error{DownloadServerJar, DownloadWithMCJarFiles} {
		dir := t.TempDir()
		outside := filepath.Join(t.TempDir(), "victim")
		os.WriteFile(outside, []byte("secret"), 0600)
		if e := os.Link(outside, filepath.Join(dir, "server.jar")); e != nil {
			t.Fatal(e)
		}
		if e := download("paper", "latest", dir, "server.jar"); e != nil {
			t.Fatal(e)
		}
		got, _ := os.ReadFile(outside)
		if string(got) != "secret" {
			t.Fatal("hardlink truncated")
		}
		os.Remove(filepath.Join(dir, "eula.txt"))
		if e := os.Symlink(outside, filepath.Join(dir, "eula.txt")); e != nil {
			t.Skip(e)
		}
		if e := download("paper", "latest", dir, "server.jar"); e == nil {
			t.Fatal("accepted eula symlink")
		}
		got, _ = os.ReadFile(outside)
		if string(got) != "secret" {
			t.Fatal("eula escaped")
		}
	}
}
func TestModrinthRootPinnedAcrossRequest(t *testing.T) {
	old := httpClient
	t.Cleanup(func() { httpClient = old })
	base := t.TempDir()
	dir := filepath.Join(base, "plugins")
	os.Mkdir(dir, 0755)
	outside := t.TempDir()
	victim := filepath.Join(outside, "plugin.jar")
	os.WriteFile(victim, []byte("secret"), 0600)
	// Probe symlink permission before performing a swap while a root is open.
	probe := filepath.Join(base, "probe")
	if e := os.Symlink(outside, probe); e != nil {
		t.Skip(e)
	}
	os.Remove(probe)
	serverRoot, e := rooted.Open(base)
	if e != nil {
		t.Fatal(e)
	}
	defer serverRoot.Close()
	pluginRoot, e := serverRoot.Sub("plugins")
	if e != nil {
		t.Fatal(e)
	}
	defer pluginRoot.Close()
	swapped := false
	httpClient = &http.Client{Transport: modrinthTransport(func(r *http.Request) (*http.Response, error) {
		data := "jar"
		if strings.Contains(r.URL.Path, "/version") {
			if !swapped {
				if e := os.Rename(dir, dir+"-old"); e != nil {
					t.Fatal(e)
				}
				if e := os.Symlink(outside, dir); e != nil {
					t.Fatal(e)
				}
				swapped = true
			}
			data = `[{"id":"release","game_versions":["1.20.1"],"loaders":["paper"],"files":[{"filename":"plugin.jar","url":"https://mock.invalid/plugin.jar","size":3}]}]`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(data))}, nil
	})}
	if _, e := DownloadModrinthPluginRoot("project", pluginRoot, "paper", "1.20.1", "release"); e != nil {
		t.Fatal(e)
	}
	got, _ := os.ReadFile(victim)
	if string(got) != "secret" {
		t.Fatal("plugin escaped pinned root")
	}
	got, e = os.ReadFile(filepath.Join(dir+"-old", "plugin.jar"))
	if e != nil || string(got) != "jar" {
		t.Fatal("not published in pinned root", e)
	}
}
