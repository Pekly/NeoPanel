package downloader

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Pekly/NeoPanel/internal/rooted"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const steelReleasesURL = "https://api.github.com/repos/Steel-Foundation/SteelMC/releases"

// IsNativeSoftware distinguishes host executables from JVM server distributions.
func IsNativeSoftware(software string) bool {
	return strings.EqualFold(strings.TrimSpace(software), "steel")
}

// ServerFileName retains the legacy jar_file API field for native executables.
func ServerFileName(software string) string {
	if IsNativeSoftware(software) {
		if runtime.GOOS == "windows" {
			return "steel.exe"
		}
		return "steel"
	}
	return "server.jar"
}

func steelAssetName(goos, arch string) (string, error) {
	switch {
	case goos == "linux" && arch == "amd64":
		return "steel-linux-musl", nil
	case goos == "windows" && arch == "amd64":
		return "steel-windows.exe", nil
	case goos == "darwin" && arch == "arm64":
		return "steel-macos", nil
	default:
		return "", fmt.Errorf("Steel has no official native binary for %s/%s; build Steel from source on that host and upload it with automatic download disabled", goos, arch)
	}
}

func ValidateSteelPlatform() error {
	_, err := steelAssetName(runtime.GOOS, runtime.GOARCH)
	return err
}

type steelAsset struct {
	Name   string `json:"name"`
	URL    string `json:"browser_download_url"`
	Size   int64  `json:"size"`
	Digest string `json:"digest"`
}
type steelRelease struct {
	Tag        string       `json:"tag_name"`
	Draft      bool         `json:"draft"`
	Prerelease bool         `json:"prerelease"`
	Assets     []steelAsset `json:"assets"`
}

func steelJSON(endpoint string, dest interface{}) error {
	resp, err := httpGet(endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Steel release API returned HTTP %d", resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(dest)
}

// GetSteelVersions returns stable release tags, not Minecraft versions.
func GetSteelVersions() ([]string, error) {
	var releases []steelRelease
	if err := steelJSON(steelReleasesURL+"?per_page=100", &releases); err != nil {
		return nil, err
	}
	versions := make([]string, 0, len(releases))
	for _, release := range releases {
		if !release.Draft && !release.Prerelease && release.Tag != "" && len(release.Assets) > 0 {
			versions = append(versions, release.Tag)
		}
	}
	return versions, nil
}

func downloadSteel(version, destDir, filename string) error {
	assetName, err := steelAssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}
	if filename == "" || filepath.Base(filename) != filename || strings.ContainsAny(filename, `/\\:`) || filename == "." || filename == ".." {
		return fmt.Errorf("invalid Steel executable filename")
	}
	if err := ValidateSoftware("steel", version); err != nil {
		return err
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	root, err := rooted.Open(destDir)
	if err != nil {
		return err
	}
	defer root.Close()
	endpoint := steelReleasesURL + "/latest"
	if version = strings.TrimSpace(version); version != "" && !strings.EqualFold(version, "latest") {
		endpoint = steelReleasesURL + "/tags/" + url.PathEscape(version)
	}
	var release steelRelease
	if err := steelJSON(endpoint, &release); err != nil {
		return err
	}
	if release.Tag == "" || release.Draft {
		return fmt.Errorf("Steel release is unavailable")
	}
	var asset *steelAsset
	for i := range release.Assets {
		if release.Assets[i].Name == assetName {
			asset = &release.Assets[i]
			break
		}
	}
	// Older Linux releases did not publish a static musl build.
	if asset == nil && assetName == "steel-linux-musl" {
		for i := range release.Assets {
			if release.Assets[i].Name == "steel-linux" {
				asset = &release.Assets[i]
				break
			}
		}
	}
	if asset == nil {
		return fmt.Errorf("Steel %s does not publish %s", release.Tag, assetName)
	}
	u, err := url.Parse(asset.URL)
	if err != nil || u.Scheme != "https" || u.Host != "github.com" || !strings.HasPrefix(u.Path, "/Steel-Foundation/SteelMC/releases/download/") {
		return fmt.Errorf("invalid official Steel asset URL")
	}
	// The destination is pinned before release metadata or asset requests.
	dir, stage, err := root.TempDir()
	if err != nil {
		return err
	}
	defer root.RemoveAll(dir)
	defer stage.Close()
	temp, err := stage.OpenFile("binary", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer temp.Close()
	// Native release binaries can exceed the shared metadata timeout on slow links.
	client := *httpClient
	client.Timeout = 5 * time.Minute
	req, err := http.NewRequest(http.MethodGet, asset.URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Steel binary download returned HTTP %d", resp.StatusCode)
	}
	hash := sha256.New()
	size, err := rooted.Copy(context.Background(), io.MultiWriter(temp, hash), resp.Body, 1<<30)
	if err != nil {
		return err
	}
	if size == 0 || (asset.Size > 0 && size != asset.Size) {
		return fmt.Errorf("incomplete Steel binary download")
	}
	if asset.Digest != "" && asset.Digest != "sha256:"+hex.EncodeToString(hash.Sum(nil)) {
		return fmt.Errorf("Steel binary checksum mismatch")
	}
	if err := temp.Chmod(0755); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	configPath := filepath.Join("config", "config.toml")
	if err := root.Check(configPath); err != nil {
		return err
	}
	if _, err := root.Lstat(configPath); os.IsNotExist(err) {
		configURL := "https://raw.githubusercontent.com/Steel-Foundation/SteelMC/" + url.PathEscape(release.Tag) + "/package-content/config.toml"
		configResp, err := httpGet(configURL)
		if err != nil {
			return err
		}
		defer configResp.Body.Close()
		if configResp.StatusCode != http.StatusOK {
			return fmt.Errorf("Steel default configuration download returned HTTP %d", configResp.StatusCode)
		}
		config, err := rooted.ReadBounded(configResp.Body, 2<<20)
		if err != nil {
			return err
		}
		if !strings.Contains(string(config), "[server]") || !strings.Contains(string(config), "server_port") {
			return fmt.Errorf("invalid Steel default configuration")
		}
		if err := root.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return err
		}
		if err := root.WriteFile(configPath, config, 0644); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if err := root.Check(filename); err != nil {
		return err
	}
	return root.Rename(filepath.Join(dir, "binary"), filename)
}
