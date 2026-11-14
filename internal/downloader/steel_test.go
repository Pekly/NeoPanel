package downloader

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSteelPlatforms(t *testing.T) {
	for _, tc := range []struct{ os, arch, want string }{
		{"linux", "amd64", "steel-linux-musl"}, {"windows", "amd64", "steel-windows.exe"}, {"darwin", "arm64", "steel-macos"},
		{"linux", "arm64", ""}, {"windows", "arm64", ""}, {"darwin", "amd64", ""},
	} {
		got, err := steelAssetName(tc.os, tc.arch)
		if got != tc.want || (err != nil) != (tc.want == "") {
			t.Errorf("%s/%s = %q, %v", tc.os, tc.arch, got, err)
		}
	}
	if !IsNativeSoftware(" Steel ") || IsNativeSoftware("paper") || ServerFileName("paper") != "server.jar" {
		t.Fatal("bad runtime classification")
	}
}

func TestSteelDownload(t *testing.T) {
	asset, err := steelAssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Skip(err)
	}
	old := httpClient
	t.Cleanup(func() { httpClient = old })
	binary := "native-binary"
	sum := sha256.Sum256([]byte(binary))
	release := steelRelease{Tag: "v0.17.0+mc26.3", Assets: []steelAsset{{Name: asset, URL: "https://github.com/Steel-Foundation/SteelMC/releases/download/v0.17.0+mc26.3/" + asset, Size: int64(len(binary)), Digest: "sha256:" + hex.EncodeToString(sum[:])}}}
	status := 200
	var seen []string
	httpClient = &http.Client{Transport: modrinthTransport(func(r *http.Request) (*http.Response, error) {
		seen = append(seen, r.URL.String())
		data := binary
		code := 200
		switch r.URL.Host {
		case "api.github.com":
			raw, _ := json.Marshal(release)
			data = string(raw)
		case "raw.githubusercontent.com":
			data = "[server]\nserver_port = 25565\nmax_players = 20\n"
		case "github.com":
			code = status
		default:
			t.Fatalf("unexpected host: %s", r.URL)
		}
		return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader(data)), Header: make(http.Header)}, nil
	})}
	dir := t.TempDir()
	name := ServerFileName("steel")
	if err := DownloadServerJar("steel", release.Tag, dir, name); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(seen[0], "/tags/v0.17.0+mc26.3") {
		t.Fatal(seen)
	}
	if _, err := os.Stat(filepath.Join(dir, "eula.txt")); !os.IsNotExist(err) {
		t.Fatal("wrote Java EULA for native server")
	}
	configPath := filepath.Join(dir, "config", "config.toml")
	if err := os.WriteFile(configPath, []byte("custom configuration"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := DownloadServerJar("steel", "latest", dir, name); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(configPath)
	if string(got) != "custom configuration" {
		t.Fatal("overwrote config")
	}
	for _, failure := range []string{"status", "size", "digest"} {
		status = 200
		binary = "native-binary"
		if failure == "status" {
			status = 503
		}
		if failure == "size" {
			binary = "short"
		}
		if failure == "digest" {
			binary = "wrong--binary"
		}
		if err := DownloadServerJar("steel", "latest", dir, name); err == nil {
			t.Fatalf("accepted %s failure", failure)
		}
		got, _ := os.ReadFile(filepath.Join(dir, name))
		if string(got) != "native-binary" {
			t.Fatalf("%s overwrote binary", failure)
		}
	}
	if err := DownloadServerJar("steel", "latest", dir, "../escape"); err == nil {
		t.Fatal("accepted unsafe name")
	}
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".steel-download-") {
			t.Fatal("leaked partial download")
		}
	}
}

func TestSteelVersions(t *testing.T) {
	old := httpClient
	t.Cleanup(func() { httpClient = old })
	httpClient = &http.Client{Transport: modrinthTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`[{"tag_name":"v1+mc26.3","assets":[{"name":"steel-linux"}]},{"tag_name":"nightly","prerelease":true,"assets":[{}]},{"tag_name":"draft","draft":true,"assets":[{}]}]`))}, nil
	})}
	versions, err := GetSteelVersions()
	if err != nil || len(versions) != 1 || versions[0] != "v1+mc26.3" {
		t.Fatal(versions, err)
	}
}
