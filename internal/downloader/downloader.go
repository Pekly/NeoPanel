package downloader

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var httpClient = &http.Client{
	Timeout: 60 * time.Second,
}

const userAgent = "NeoPanel/1.0 (+https://github.com/Pekly/NeoPanel)"

func httpGet(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	return httpClient.Do(req)
}

// DownloadServerJar downloads the specified server flavor and version into destDir as jarFileName.
// It also writes eula.txt (eula=true) automatically so the server is ready to run immediately.
func DownloadServerJar(flavor, version, destDir, jarFileName string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Try fast, reliable mcjarfiles.com API first for supported engines
	if err := DownloadWithMCJarFiles(flavor, version, destDir, jarFileName); err == nil {
		return nil
	}

	destPath := filepath.Join(destDir, jarFileName)

	var downloadURL string
	var err error

	switch flavor {
	case "paper":
		downloadURL, err = getPaperURL(version)
	case "purpur":
		downloadURL, err = getPurpurURL(version)
	case "vanilla":
		downloadURL, err = getVanillaURL(version)
	case "fabric":
		downloadURL, err = getFabricURL(version)
	case "velocity":
		downloadURL, err = getVelocityURL(version)
	case "waterfall":
		downloadURL, err = getWaterfallURL(version)
	case "bungeecord":
		downloadURL, err = getBungeeCordURL()
	default:
		return fmt.Errorf("unsupported server flavor for auto-download: %s", flavor)
	}

	if err != nil {
		return fmt.Errorf("failed to resolve download URL for %s %s: %w", flavor, version, err)
	}

	if err := downloadFile(destPath, downloadURL); err != nil {
		return fmt.Errorf("failed to download file from %s: %w", downloadURL, err)
	}

	// Automatically agree to Minecraft EULA
	eulaPath := filepath.Join(destDir, "eula.txt")
	_ = os.WriteFile(eulaPath, []byte("# EULA auto-accepted by NeoPanel\neula=true\n"), 0644)

	return nil
}

// DownloadWithMCJarFiles downloads a Minecraft server JAR from mcjarfiles.com REST API.
// Automatically follows 302 redirects to the official upstream builds.
func DownloadWithMCJarFiles(flavor, version, destDir, jarFileName string) error {
	fl := strings.ToLower(strings.TrimSpace(flavor))
	ver := strings.TrimSpace(version)

	var endpoint string
	switch fl {
	case "paper":
		if ver == "" || strings.EqualFold(ver, "latest") {
			endpoint = "https://mcjarfiles.com/api/get-latest-jar/servers/paper"
		} else {
			endpoint = fmt.Sprintf("https://mcjarfiles.com/api/get-jar/servers/paper/%s", ver)
		}
	case "purpur":
		if ver == "" || strings.EqualFold(ver, "latest") {
			endpoint = "https://mcjarfiles.com/api/get-latest-jar/servers/purpur"
		} else {
			endpoint = fmt.Sprintf("https://mcjarfiles.com/api/get-jar/servers/purpur/%s", ver)
		}
	case "leaf", "leafmc":
		if ver == "" || strings.EqualFold(ver, "latest") {
			endpoint = "https://mcjarfiles.com/api/get-latest-jar/servers/leafmc"
		} else {
			endpoint = fmt.Sprintf("https://mcjarfiles.com/api/get-jar/servers/leafmc/%s", ver)
		}
	case "folia":
		if ver == "" || strings.EqualFold(ver, "latest") {
			endpoint = "https://mcjarfiles.com/api/get-latest-jar/servers/folia"
		} else {
			endpoint = fmt.Sprintf("https://mcjarfiles.com/api/get-jar/servers/folia/%s", ver)
		}
	case "fabric":
		if ver == "" || strings.EqualFold(ver, "latest") {
			endpoint = "https://mcjarfiles.com/api/get-latest-jar/modded/fabric"
		} else {
			endpoint = fmt.Sprintf("https://mcjarfiles.com/api/get-jar/modded/fabric/%s", ver)
		}
	case "forge":
		if ver == "" || strings.EqualFold(ver, "latest") {
			endpoint = "https://mcjarfiles.com/api/get-latest-jar/modded/forge"
		} else {
			endpoint = fmt.Sprintf("https://mcjarfiles.com/api/get-jar/modded/forge/%s", ver)
		}
	case "neoforge":
		if ver == "" || strings.EqualFold(ver, "latest") {
			endpoint = "https://mcjarfiles.com/api/get-latest-jar/modded/neoforge"
		} else {
			endpoint = fmt.Sprintf("https://mcjarfiles.com/api/get-jar/modded/neoforge/%s", ver)
		}
	case "vanilla":
		if ver == "" || strings.EqualFold(ver, "latest") {
			endpoint = "https://mcjarfiles.com/api/get-latest-jar/vanilla/release"
		} else {
			endpoint = fmt.Sprintf("https://mcjarfiles.com/api/get-jar/vanilla/release/%s", ver)
		}
	case "velocity":
		if ver == "" || strings.EqualFold(ver, "latest") || strings.HasPrefix(ver, "1.") {
			endpoint = "https://mcjarfiles.com/api/get-latest-jar/proxies/velocity"
		} else {
			endpoint = fmt.Sprintf("https://mcjarfiles.com/api/get-jar/proxies/velocity/%s", ver)
		}
	default:
		return fmt.Errorf("flavor %s is not supported on mcjarfiles", fl)
	}

	_ = os.MkdirAll(destDir, 0755)
	destPath := filepath.Join(destDir, jarFileName)

	resp, err := httpGet(endpoint)
	if err != nil {
		return fmt.Errorf("mcjarfiles download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mcjarfiles returned non-200 status: %d %s", resp.StatusCode, resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write server jar: %w", err)
	}

	// Auto-accept EULA
	eulaPath := filepath.Join(destDir, "eula.txt")
	_ = os.WriteFile(eulaPath, []byte("# EULA auto-accepted by NeoPanel\neula=true\n"), 0644)

	return nil
}

func downloadFile(filepath string, url string) error {
	resp, err := httpGet(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func getPaperURL(version string) (string, error) {
	// PaperMC v3 Fill API
	apiURL := fmt.Sprintf("https://fill.papermc.io/v3/projects/paper/versions/%s/builds", version)
	resp, err := httpGet(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("PaperMC v3 API returned status %d for version %s", resp.StatusCode, version)
	}

	var builds []struct {
		ID        int    `json:"id"`
		Channel   string `json:"channel"`
		Downloads map[string]struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"downloads"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&builds); err != nil {
		return "", err
	}

	if len(builds) == 0 {
		return "", fmt.Errorf("no builds found for Paper version %s", version)
	}

	// First item is the newest build in v3 Fill API
	latest := builds[0]
	if dl, ok := latest.Downloads["server:default"]; ok && dl.URL != "" {
		return dl.URL, nil
	}

	// Fallback to any download URL in the map
	for _, dl := range latest.Downloads {
		if dl.URL != "" {
			return dl.URL, nil
		}
	}

	return "", fmt.Errorf("no server download URL found in latest build %d for Paper %s", latest.ID, version)
}

func getPurpurURL(version string) (string, error) {
	// Purpur direct latest build endpoint
	return fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s/latest/download", version), nil
}

func getVanillaURL(version string) (string, error) {
	manifestURL := "https://launchermeta.mojang.com/mc/game/version_manifest_v2.json"
	resp, err := httpGet(manifestURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var manifest struct {
		Versions []struct {
			ID  string `json:"id"`
			URL string `json:"url"`
		} `json:"versions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return "", err
	}

	var versionDetailURL string
	for _, v := range manifest.Versions {
		if v.ID == version {
			versionDetailURL = v.URL
			break
		}
	}

	if versionDetailURL == "" {
		return "", fmt.Errorf("version %s not found in Mojang manifest", version)
	}

	vResp, err := httpGet(versionDetailURL)
	if err != nil {
		return "", err
	}
	defer vResp.Body.Close()

	var versionData struct {
		Downloads struct {
			Server struct {
				URL string `json:"url"`
			} `json:"server"`
		} `json:"downloads"`
	}

	if err := json.NewDecoder(vResp.Body).Decode(&versionData); err != nil {
		return "", err
	}

	if versionData.Downloads.Server.URL == "" {
		return "", fmt.Errorf("no server download URL found for vanilla %s", version)
	}

	return versionData.Downloads.Server.URL, nil
}

func getFabricURL(version string) (string, error) {
	loaderURL := "https://meta.fabricmc.net/v2/versions/loader"
	resp, err := httpGet(loaderURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var loaders []struct {
		Version string `json:"version"`
		Stable  bool   `json:"stable"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loaders); err != nil {
		return "", err
	}

	loaderVersion := "0.16.9" // modern fallback
	for _, l := range loaders {
		if l.Stable {
			loaderVersion = l.Version
			break
		}
	}

	installerVersion := "1.0.1"
	return fmt.Sprintf("https://meta.fabricmc.net/v2/versions/loader/%s/%s/%s/server/jar",
		version, loaderVersion, installerVersion), nil
}

func getVelocityURL(version string) (string, error) {
	v := strings.TrimSpace(version)
	// If version is empty, "latest", or looks like Minecraft version ("1.21.4"), default to "3.4.0-SNAPSHOT"
	if v == "" || strings.EqualFold(v, "latest") || strings.HasPrefix(v, "1.") {
		v = "3.4.0-SNAPSHOT"
	}

	apiURL := fmt.Sprintf("https://fill.papermc.io/v3/projects/velocity/versions/%s/builds", v)
	resp, err := httpGet(apiURL)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var builds []struct {
				Downloads map[string]struct {
					URL string `json:"url"`
				} `json:"downloads"`
			}
			if json.NewDecoder(resp.Body).Decode(&builds) == nil && len(builds) > 0 {
				if dl, ok := builds[0].Downloads["server:default"]; ok && dl.URL != "" {
					return dl.URL, nil
				}
				for _, dl := range builds[0].Downloads {
					if dl.URL != "" {
						return dl.URL, nil
					}
				}
			}
		}
	}
	// Fallback to verified direct PaperMC fill-data object
	return "https://fill-data.papermc.io/v1/objects/fe53021f3168322cb6cb68f78699866fd098df3c306e4359847a10b0d02689ef/velocity-3.4.0-SNAPSHOT-563.jar", nil
}

func getWaterfallURL(version string) (string, error) {
	v := strings.TrimSpace(version)
	if v == "" || strings.EqualFold(v, "latest") || (!strings.HasPrefix(v, "1.20") && !strings.HasPrefix(v, "1.21") && !strings.HasPrefix(v, "1.19")) {
		v = "1.20"
	}

	apiURL := fmt.Sprintf("https://fill.papermc.io/v3/projects/waterfall/versions/%s/builds", v)
	resp, err := httpGet(apiURL)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var builds []struct {
				Downloads map[string]struct {
					URL string `json:"url"`
				} `json:"downloads"`
			}
			if json.NewDecoder(resp.Body).Decode(&builds) == nil && len(builds) > 0 {
				if dl, ok := builds[0].Downloads["server:default"]; ok && dl.URL != "" {
					return dl.URL, nil
				}
				for _, dl := range builds[0].Downloads {
					if dl.URL != "" {
						return dl.URL, nil
					}
				}
			}
		}
	}
	// Fallback to verified direct PaperMC fill-data object
	return "https://fill-data.papermc.io/v1/objects/a6786337d4d41e19b5d0f771027b01956e6fb11a62244e762f5dae2e76c34644/waterfall-1.20-578.jar", nil
}

func getBungeeCordURL() (string, error) {
	return "https://ci.md-5.net/job/BungeeCord/lastSuccessfulBuild/artifact/bootstrap/target/BungeeCord.jar", nil
}
