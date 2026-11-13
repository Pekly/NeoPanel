package downloader

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompatiblePluginVersions(t *testing.T) {
	var versions []ModrinthVersion
	if err := json.Unmarshal([]byte(`[
 {"id":"new","game_versions":["1.21.4"],"loaders":["paper"],"files":[{"filename":"new.jar"}]},
 {"id":"right","game_versions":["1.20.1"],"loaders":["paper"],"files":[{"filename":"right.jar"}]},
 {"id":"fabric","game_versions":["1.20.1"],"loaders":["fabric"],"files":[{"filename":"mod.jar"}]},
 {"id":"folia","game_versions":["1.20.1"],"loaders":["folia"],"files":[{"filename":"folia.jar"}]},
 {"id":"unsafe","game_versions":["1.20.1"],"loaders":["paper"],"files":[{"filename":"../escape.jar"}]}
 ]`), &versions); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ loader, version, want string }{
		{"paper", "1.20.1", "right"}, {"purpur", "1.20.1", "right"}, {"leafmc", "1.20.1", "right"},
		{"fabric", "1.20.1", "fabric"}, {"folia", "1.20.1", "folia"}, {"forge", "1.20.1", ""}, {"paper", "1.19.4", ""},
	} {
		t.Run(tc.loader+tc.version, func(t *testing.T) {
			got := filterCompatibleVersions(versions, tc.loader, tc.version)
			if tc.want == "" {
				if len(got) != 0 {
					t.Fatalf("unexpected fallback: %+v", got)
				}
				return
			}
			if len(got) != 1 || got[0].ID != tc.want {
				t.Fatalf("got %+v, want %s", got, tc.want)
			}
		})
	}
}

type modrinthTransport func(*http.Request) (*http.Response, error)

func (f modrinthTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestPluginDownloadPublication(t *testing.T) {
	old := httpClient
	t.Cleanup(func() { httpClient = old })
	body := "jar-data"
	httpClient = &http.Client{Transport: modrinthTransport(func(r *http.Request) (*http.Response, error) {
		data := body
		if strings.Contains(r.URL.Path, "/version") {
			data = `[{"id":"release","game_versions":["1.20.1"],"loaders":["paper"],"files":[{"filename":"plugin.jar","url":"https://cdn.modrinth.com/plugin.jar","size":8,"primary":true}]}]`
		}
		return &http.Response{StatusCode: 200, Status: "200 OK", Body: io.NopCloser(strings.NewReader(data)), Header: make(http.Header)}, nil
	})}
	dir := t.TempDir()
	if _, err := DownloadModrinthPlugin("project", dir, "paper", "1.20.1", "wrong-id"); err == nil {
		t.Fatal("accepted incompatible release ID")
	}
	if _, err := DownloadModrinthPlugin("project", dir, "forge", "1.20.1", ""); err == nil {
		t.Fatal("accepted wrong loader")
	}
	name, err := DownloadModrinthPlugin("project", dir, "paper", "1.20.1", "release")
	if err != nil || name != "plugin.jar" {
		t.Fatalf("download: %q %v", name, err)
	}
	if _, err := DownloadModrinthPlugin("project", dir, "paper", "1.20.1", "release"); err == nil {
		t.Fatal("overwrote existing JAR")
	}
	got, _ := os.ReadFile(filepath.Join(dir, name))
	if string(got) != body {
		t.Fatal("existing content changed")
	}
	if err := os.Rename(filepath.Join(dir, name), filepath.Join(dir, name+".disabled")); err != nil {
		t.Fatal(err)
	}
	if _, err := DownloadModrinthPlugin("project", dir, "paper", "1.20.1", "release"); err == nil {
		t.Fatal("installed over disabled plugin")
	}
	body = "short"
	empty := t.TempDir()
	if _, err := DownloadModrinthPlugin("project", empty, "paper", "1.20.1", "release"); err == nil {
		t.Fatal("published incomplete download")
	}
	entries, _ := os.ReadDir(empty)
	if len(entries) != 0 {
		t.Fatalf("incomplete artifacts remain: %v", entries)
	}
}
