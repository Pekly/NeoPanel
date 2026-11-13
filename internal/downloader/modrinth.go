package downloader

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type ModrinthSearchResult struct {
	Hits      []ModrinthPluginHit `json:"hits"`
	TotalHits int                 `json:"total_hits"`
}

type ModrinthPluginHit struct {
	ProjectID   string   `json:"project_id"`
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Categories  []string `json:"categories"`
	Downloads   int64    `json:"downloads"`
	Follows     int64    `json:"follows"`
	IconURL     string   `json:"icon_url"`
	Author      string   `json:"author"`
	Versions    []string `json:"versions"`
}

type ModrinthVersion struct {
	ID            string   `json:"id"`
	ProjectID     string   `json:"project_id"`
	Name          string   `json:"name"`
	VersionNumber string   `json:"version_number"`
	GameVersions  []string `json:"game_versions"`
	Loaders       []string `json:"loaders"`
	Files         []struct {
		URL      string `json:"url"`
		Filename string `json:"filename"`
		Primary  bool   `json:"primary"`
		Size     int64  `json:"size"`
	} `json:"files"`
}

// IsModdedSoftware returns true when the given server software runs JVM mods (mods/) instead of Bukkit/Proxy plugins (plugins/).
func IsModdedSoftware(software string) bool {
	switch strings.ToLower(strings.TrimSpace(software)) {
	case "fabric", "forge", "neoforge":
		return true
	default:
		return false
	}
}

// SearchModrinthPlugins queries the Modrinth API for Minecraft plugins or mods with optional loader, version, sort, and project_type filters.
func SearchModrinthPlugins(query string, limit int, loader string, mcVersion string, sortBy string, projectType ...string) (*ModrinthSearchResult, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	var facetGroups []string
	loaderLow := strings.ToLower(strings.TrimSpace(loader))
	reqType := ""
	if len(projectType) > 0 {
		pt := strings.ToLower(strings.TrimSpace(projectType[0]))
		if pt == "mod" || pt == "mods" {
			reqType = "mod"
		} else if pt == "plugin" || pt == "plugins" {
			reqType = "plugin"
		}
	}
	if reqType == "" {
		if IsModdedSoftware(loaderLow) {
			reqType = "mod"
		} else {
			reqType = "plugin"
		}
	}

	if reqType == "mod" {
		facetGroups = append(facetGroups, `["project_type:mod"]`)
		if loaderLow != "" && loaderLow != "any" && loaderLow != "all" {
			if loaderLow == "fabric" || loaderLow == "forge" || loaderLow == "neoforge" || loaderLow == "quilt" {
				facetGroups = append(facetGroups, fmt.Sprintf(`["categories:%s"]`, loaderLow))
			}
		}
	} else {
		facetGroups = append(facetGroups, `["project_type:plugin"]`)
		switch loaderLow {
		case "", "any", "all", "vanilla", "custom":
			// All plugins
		case "paper", "purpur", "leaf", "leafmc", "folia":
			facetGroups = append(facetGroups, `["categories:paper","categories:spigot","categories:bukkit"]`)
		case "spigot", "bukkit":
			facetGroups = append(facetGroups, `["categories:spigot","categories:bukkit"]`)
		case "velocity", "waterfall", "bungeecord":
			facetGroups = append(facetGroups, fmt.Sprintf(`["categories:%s"]`, loaderLow))
		default:
			facetGroups = append(facetGroups, fmt.Sprintf(`["categories:%s"]`, loaderLow))
		}
	}

	verTrim := strings.TrimSpace(mcVersion)
	if verTrim != "" && !strings.EqualFold(verTrim, "any") {
		facetGroups = append(facetGroups, fmt.Sprintf(`["versions:%s"]`, verTrim))
	}

	index := "relevance"
	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "downloads", "follows", "newest", "updated", "relevance":
		index = strings.ToLower(strings.TrimSpace(sortBy))
	}

	facets := "[" + strings.Join(facetGroups, ",") + "]"
	apiURL := fmt.Sprintf("https://api.modrinth.com/v2/search?query=%s&facets=%s&index=%s&limit=%d",
		url.QueryEscape(query), url.QueryEscape(facets), url.QueryEscape(index), limit)

	resp, err := httpGet(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("modrinth search returned status %s", resp.Status)
	}

	var result ModrinthSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CompatibleModrinthVersions returns only releases explicitly supporting this server.
func CompatibleModrinthVersions(project, loader, mcVersion string) ([]ModrinthVersion, error) {
	if strings.TrimSpace(project) == "" || strings.TrimSpace(mcVersion) == "" {
		return nil, fmt.Errorf("project and server Minecraft version are required")
	}
	apiURL := fmt.Sprintf("https://api.modrinth.com/v2/project/%s/version", url.PathEscape(project))
	resp, err := httpGet(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Modrinth returned status %s", resp.Status)
	}
	var versions []ModrinthVersion
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, err
	}
	return filterCompatibleVersions(versions, loader, mcVersion), nil
}

func filterCompatibleVersions(versions []ModrinthVersion, loader, mcVersion string) []ModrinthVersion {
	loader = strings.ToLower(strings.TrimSpace(loader))
	switch loader {
	case "purpur", "leaf", "leafmc":
		loader = "paper"
	case "waterfall":
		loader = "bungeecord"
	}
	result := make([]ModrinthVersion, 0)
	for _, v := range versions {
		versionOK, loaderOK, jarOK := false, false, false
		for _, ver := range v.GameVersions {
			if ver == strings.TrimSpace(mcVersion) {
				versionOK = true
			}
		}
		for _, l := range v.Loaders {
			l = strings.ToLower(l)
			if l == loader || (loader == "paper" && (l == "spigot" || l == "bukkit")) || (loader == "spigot" && l == "bukkit") {
				loaderOK = true
			}
		}
		for _, f := range v.Files {
			if validPluginJarName(f.Filename) {
				jarOK = true
			}
		}
		// Folia must be explicitly tagged; Paper compatibility alone is not sufficient.
		if versionOK && loaderOK && jarOK {
			result = append(result, v)
		}
	}
	return result
}

func validPluginJarName(name string) bool {
	return name != "" && !strings.ContainsAny(name, "/\\:") && !strings.Contains(name, "..") && strings.HasSuffix(strings.ToLower(name), ".jar")
}

// DownloadModrinthPlugin revalidates the selected release against the server, then
// publishes a completed JAR without replacing an existing enabled or disabled file.
func DownloadModrinthPlugin(project, destDir, loader, mcVersion, versionID string) (string, error) {
	versions, err := CompatibleModrinthVersions(project, loader, mcVersion)
	if err != nil {
		return "", err
	}
	var chosen *ModrinthVersion
	for i := range versions {
		if versionID == "" || versions[i].ID == versionID {
			chosen = &versions[i]
			break
		}
	}
	if chosen == nil {
		return "", fmt.Errorf("no compatible release for %s on Minecraft %s (%s)", project, mcVersion, loader)
	}
	fileIndex := -1
	for i, f := range chosen.Files {
		if validPluginJarName(f.Filename) && (fileIndex < 0 || f.Primary) {
			fileIndex = i
			if f.Primary {
				break
			}
		}
	}
	if fileIndex < 0 {
		return "", fmt.Errorf("compatible release has no safe JAR file")
	}
	f := chosen.Files[fileIndex]
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}
	dest := filepath.Join(destDir, f.Filename)
	for _, path := range []string{dest, dest + ".disabled"} {
		if _, err := os.Lstat(path); err == nil {
			return "", fmt.Errorf("%s already exists; remove it explicitly before reinstalling", filepath.Base(path))
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}
	resp, err := httpGet(f.URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned status %s", resp.Status)
	}
	tmp, err := os.CreateTemp(destDir, ".neopanel-plugin-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	n, copyErr := io.Copy(tmp, resp.Body)
	closeErr := tmp.Close()
	if copyErr != nil {
		return "", copyErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	if f.Size > 0 && n != f.Size {
		return "", fmt.Errorf("incomplete plugin download: expected %d bytes, received %d", f.Size, n)
	}
	if err := os.Chmod(tmp.Name(), 0644); err != nil {
		return "", err
	}
	// Link is atomic and, unlike Rename on Unix, never overwrites the destination.
	if err := os.Link(tmp.Name(), dest); err != nil {
		return "", fmt.Errorf("cannot publish plugin (file may already exist): %w", err)
	}
	return f.Filename, nil
}
