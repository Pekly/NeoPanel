package downloader

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Pekly/NeoPanel/internal/proxyconfig"
	"github.com/Pekly/NeoPanel/internal/rooted"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
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

// DownloadServerJar installs the server distribution (a native binary for Steel).
// World-server distributions also get an automatically accepted eula.txt.
func DownloadServerJar(flavor, version, destDir, jarFileName string) error {
	if err := ValidateSoftware(flavor, version); err != nil {
		return err
	}
	if !rooted.Basename(jarFileName) {
		return fmt.Errorf("invalid server filename")
	}
	if IsNativeSoftware(flavor) {
		return downloadSteel(version, destDir, jarFileName)
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	root, err := rooted.Open(destDir)
	if err != nil {
		return err
	}
	defer root.Close()
	// Keep the same pinned root across both upstream attempts.
	primaryErr := downloadMCJarFiles(flavor, version, root, jarFileName)
	if primaryErr == nil {
		return nil
	}

	var downloadURL string
	err = nil

	switch flavor {
	case "paper":
		downloadURL, err = getPaperURL(version)
	case "folia":
		downloadURL, err = getFillURL("folia", version)
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
		return fmt.Errorf("failed to download %s %s (no official fallback): %w", flavor, version, primaryErr)
	}

	if err != nil {
		return fmt.Errorf("failed to resolve download URL for %s %s: %w", flavor, version, err)
	}

	if err := downloadRootFile(root, jarFileName, downloadURL); err != nil {
		return fmt.Errorf("failed to download file from %s: %w", downloadURL, err)
	}

	return acceptWorldServerEULA(root, flavor)
}

// DownloadWithMCJarFiles downloads a Minecraft server JAR from mcjarfiles.com REST API.
// Automatically follows 302 redirects to the official upstream builds.
func DownloadWithMCJarFiles(flavor, version, destDir, jarFileName string) error {
	if err := ValidateSoftware(flavor, version); err != nil {
		return err
	}
	if !rooted.Basename(jarFileName) {
		return fmt.Errorf("invalid server filename")
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	root, err := rooted.Open(destDir)
	if err != nil {
		return err
	}
	defer root.Close()
	return downloadMCJarFiles(flavor, version, root, jarFileName)
}

func downloadMCJarFiles(flavor, version string, root *rooted.Root, jarFileName string) error {
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

	// root was pinned before any remote request.

	resp, err := httpGet(endpoint)
	if err != nil {
		return fmt.Errorf("mcjarfiles download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mcjarfiles returned non-200 status: %d %s", resp.StatusCode, resp.Status)
	}

	if _, err := root.Copy(context.Background(), jarFileName, resp.Body, 1<<30, 0644); err != nil {
		return err
	}
	return acceptWorldServerEULA(root, flavor)
}

func acceptWorldServerEULA(root *rooted.Root, flavor string) error {
	if proxyconfig.Supported(flavor) {
		return nil
	}
	return root.WriteFile("eula.txt", []byte("# EULA auto-accepted by NeoPanel\neula=true\n"), 0644)
}

func downloadRootFile(root *rooted.Root, name string, url string) error {
	resp, err := httpGet(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = root.Copy(context.Background(), name, resp.Body, 1<<30, 0644)
	return err
}

// metadataJSON bounds upstream metadata independently of the larger JAR limit.
func metadataJSON(url string, out any) error {
	resp, err := httpGet(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("metadata %s returned status %d", url, resp.StatusCode)
	}
	const limit = 4 << 20
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return err
	}
	if len(data) > limit {
		return fmt.Errorf("metadata %s exceeds size limit", url)
	}
	return json.Unmarshal(data, out)
}

func wantsLatest(version string) bool {
	return strings.TrimSpace(version) == "" || strings.EqualFold(strings.TrimSpace(version), "latest")
}

// Only compare the selected project's own supported versions, never a merged
// Minecraft catalog. Velocity's normal upstream releases use a SNAPSHOT suffix.
func newestSupported(versions []string, snapshot bool) (string, error) {
	var best string
	var bestParts [4]int
	for _, version := range versions {
		base := version
		if snapshot {
			base = strings.TrimSuffix(base, "-SNAPSHOT")
		}
		parts := strings.Split(base, ".")
		if len(parts) < 2 || len(parts) > 4 {
			continue
		}
		var nums [4]int
		valid := true
		for i, part := range parts {
			if part == "" || strings.IndexFunc(part, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
				valid = false
				break
			}
			n, err := strconv.Atoi(part)
			if err != nil {
				valid = false
				break
			}
			nums[i] = n
		}
		if !valid {
			continue
		}
		newer := best == ""
		for i := range nums {
			if nums[i] != bestParts[i] {
				newer = nums[i] > bestParts[i]
				break
			}
		}
		if nums == bestParts && !strings.Contains(version, "-") {
			newer = true
		}
		if newer {
			best, bestParts = version, nums
		}
	}
	if best == "" {
		return "", fmt.Errorf("no supported release versions in upstream metadata")
	}
	return best, nil
}

func getPaperURL(version string) (string, error) {
	return getFillURL("paper", version)
}

func getFillURL(project, version string) (string, error) {
	baseURL := "https://fill.papermc.io/v3/projects/" + project
	if wantsLatest(version) {
		var payload struct {
			Versions json.RawMessage `json:"versions"`
		}
		if err := metadataJSON(baseURL, &payload); err != nil {
			return "", err
		}
		var versions []string
		if err := json.Unmarshal(payload.Versions, &versions); err != nil {
			var groups map[string][]string
			if err := json.Unmarshal(payload.Versions, &groups); err != nil {
				return "", err
			}
			for _, group := range groups {
				versions = append(versions, group...)
			}
		}
		var err error
		version, err = newestSupported(versions, project == "velocity")
		if err != nil {
			return "", fmt.Errorf("%s: %w", project, err)
		}
	}
	var builds []struct {
		ID        int `json:"id"`
		Downloads map[string]struct {
			URL string `json:"url"`
		} `json:"downloads"`
	}
	if err := metadataJSON(fmt.Sprintf("%s/versions/%s/builds", baseURL, version), &builds); err != nil {
		return "", err
	}
	if len(builds) == 0 {
		return "", fmt.Errorf("no builds found for %s %s", project, version)
	}
	latest := builds[0]
	for _, build := range builds[1:] {
		if build.ID > latest.ID {
			latest = build
		}
	}
	if url := latest.Downloads["server:default"].URL; url != "" {
		return url, nil
	}
	return "", fmt.Errorf("no server download in latest build %d for %s %s", latest.ID, project, version)
}

func getPurpurURL(version string) (string, error) {
	if wantsLatest(version) {
		var project struct {
			Versions []string `json:"versions"`
		}
		if err := metadataJSON("https://api.purpurmc.org/v2/purpur", &project); err != nil {
			return "", err
		}
		var err error
		version, err = newestSupported(project.Versions, false)
		if err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("https://api.purpurmc.org/v2/purpur/%s/latest/download", version), nil
}

func getVanillaURL(version string) (string, error) {
	var manifest struct {
		Latest struct {
			Release string `json:"release"`
		} `json:"latest"`
		Versions []struct {
			ID  string `json:"id"`
			URL string `json:"url"`
		} `json:"versions"`
	}
	if err := metadataJSON("https://launchermeta.mojang.com/mc/game/version_manifest_v2.json", &manifest); err != nil {
		return "", err
	}
	if wantsLatest(version) {
		version = manifest.Latest.Release
	}
	var detailURL string
	for _, v := range manifest.Versions {
		if version != "" && v.ID == version {
			detailURL = v.URL
			break
		}
	}
	if detailURL == "" {
		return "", fmt.Errorf("version %s not found in Mojang manifest", version)
	}
	var detail struct {
		Downloads struct {
			Server struct {
				URL string `json:"url"`
			} `json:"server"`
		} `json:"downloads"`
	}
	if err := metadataJSON(detailURL, &detail); err != nil {
		return "", err
	}
	if detail.Downloads.Server.URL == "" {
		return "", fmt.Errorf("no server download URL found for vanilla %s", version)
	}
	return detail.Downloads.Server.URL, nil
}

func stableFabricVersion(kind string) (string, error) {
	var items []struct {
		Version string `json:"version"`
		Stable  bool   `json:"stable"`
	}
	if err := metadataJSON("https://meta.fabricmc.net/v2/versions/"+kind, &items); err != nil {
		return "", err
	}
	var stable []string
	for _, item := range items {
		if item.Stable {
			stable = append(stable, item.Version)
		}
	}
	return newestSupported(stable, false)
}

func getFabricURL(version string) (string, error) {
	if wantsLatest(version) {
		var err error
		version, err = stableFabricVersion("game")
		if err != nil {
			return "", err
		}
	}
	loaderVersion, err := stableFabricVersion("loader")
	if err != nil {
		return "", err
	}
	const installerVersion = "1.0.1"
	return fmt.Sprintf("https://meta.fabricmc.net/v2/versions/loader/%s/%s/%s/server/jar",
		version, loaderVersion, installerVersion), nil
}

func getVelocityURL(version string) (string, error) {
	// Preserve the historical Minecraft-version compatibility mapping, but
	// resolve the current proxy release rather than pinning an old proxy build.
	if strings.HasPrefix(strings.TrimSpace(version), "1.") {
		version = "latest"
	}
	return getFillURL("velocity", version)
}

func getWaterfallURL(version string) (string, error) {
	return getFillURL("waterfall", version)
}

func getBungeeCordURL() (string, error) {
	return "https://ci.md-5.net/job/BungeeCord/lastSuccessfulBuild/artifact/bootstrap/target/BungeeCord.jar", nil
}

var softwareIdentifier = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+\-]{0,127}$`)

func ValidateSoftware(flavor, version string) error {
	if !softwareIdentifier.MatchString(flavor) || !softwareIdentifier.MatchString(version) {
		return fmt.Errorf("invalid software or version identifier")
	}
	switch strings.ToLower(flavor) {
	case "paper", "purpur", "leaf", "leafmc", "folia", "fabric", "forge", "neoforge", "vanilla", "velocity", "waterfall", "bungeecord", "steel":
		return nil
	}
	return fmt.Errorf("unsupported server software")
}
