package downloader

import (
	"encoding/json"
	"fmt"
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

// DownloadModrinthPlugin downloads the best matching plugin jar into destPluginsDir
func DownloadModrinthPlugin(projectIDOrSlug string, destPluginsDir string, preferredLoader string) (string, error) {
	if err := os.MkdirAll(destPluginsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create plugins directory: %w", err)
	}

	apiURL := fmt.Sprintf("https://api.modrinth.com/v2/project/%s/version", url.PathEscape(projectIDOrSlug))
	resp, err := httpGet(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("modrinth versions returned status %s", resp.Status)
	}

	var versions []ModrinthVersion
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return "", err
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no versions found for plugin %s", projectIDOrSlug)
	}

	// Find the best version matching loader (paper, spigot, purpur, leaf, folia, fabric, forge, neoforge, velocity, etc.)
	var chosenVersion *ModrinthVersion
	loaderTarget := strings.ToLower(strings.TrimSpace(preferredLoader))
	if loaderTarget == "purpur" || loaderTarget == "leaf" || loaderTarget == "leafmc" || loaderTarget == "folia" {
		loaderTarget = "paper"
	}

	for i := range versions {
		v := &versions[i]
		if len(v.Files) == 0 {
			continue
		}
		for _, l := range v.Loaders {
			lLow := strings.ToLower(l)
			if lLow == loaderTarget ||
				(loaderTarget == "paper" && (lLow == "spigot" || lLow == "bukkit" || lLow == "purpur")) {
				chosenVersion = v
				break
			}
		}
		if chosenVersion != nil {
			break
		}
	}

	// Fallback to latest available version if specific loader wasn't explicitly tagged
	if chosenVersion == nil {
		chosenVersion = &versions[0]
	}

	if len(chosenVersion.Files) == 0 {
		return "", fmt.Errorf("no files in chosen version for %s", projectIDOrSlug)
	}

	// Pick primary file or first file
	chosenFile := chosenVersion.Files[0]
	for _, f := range chosenVersion.Files {
		if f.Primary {
			chosenFile = f
			break
		}
	}

	destPath := filepath.Join(destPluginsDir, chosenFile.Filename)
	if err := downloadFile(destPath, chosenFile.URL); err != nil {
		return "", fmt.Errorf("failed to download plugin file: %w", err)
	}

	return chosenFile.Filename, nil
}
