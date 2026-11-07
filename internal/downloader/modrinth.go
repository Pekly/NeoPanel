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

// SearchModrinthPlugins queries the Modrinth API for Minecraft plugins
func SearchModrinthPlugins(query string, limit int) (*ModrinthSearchResult, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	facets := `[["project_type:plugin"]]`
	apiURL := fmt.Sprintf("https://api.modrinth.com/v2/search?query=%s&facets=%s&limit=%d",
		url.QueryEscape(query), url.QueryEscape(facets), limit)

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

	// Find the best version matching loader (paper, spigot, purpur, or fabric)
	var chosenVersion *ModrinthVersion
	loaderTarget := strings.ToLower(preferredLoader)
	if loaderTarget == "purpur" {
		loaderTarget = "paper" // Purpur supports paper/spigot/bukkit plugins
	}

	for i := range versions {
		v := &versions[i]
		if len(v.Files) == 0 {
			continue
		}
		for _, l := range v.Loaders {
			lLow := strings.ToLower(l)
			if lLow == loaderTarget || (loaderTarget == "paper" && (lLow == "spigot" || lLow == "bukkit")) {
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
