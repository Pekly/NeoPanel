package downloader

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/rooted"
)

type ModrinthSearchResult struct {
	Hits      []ModrinthPluginHit `json:"hits"`
	TotalHits int                 `json:"total_hits"`
	Offset    int                 `json:"offset"`
	Limit     int                 `json:"limit"`
}

type ModrinthPluginHit struct {
	ProjectID       string   `json:"project_id"`
	Slug            string   `json:"slug"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Categories      []string `json:"categories"`
	Downloads       int64    `json:"downloads"`
	Follows         int64    `json:"follows"`
	IconURL         string   `json:"icon_url"`
	Author          string   `json:"author"`
	Versions        []string `json:"versions"`
	FeaturedGallery string   `json:"featured_gallery"`
	Gallery         []string `json:"gallery"`
}

type ModrinthVersion struct {
	ID            string               `json:"id"`
	ProjectID     string               `json:"project_id"`
	Name          string               `json:"name"`
	VersionNumber string               `json:"version_number"`
	GameVersions  []string             `json:"game_versions"`
	Loaders       []string             `json:"loaders"`
	DatePublished string               `json:"date_published"`
	VersionType   string               `json:"version_type"`
	Changelog     string               `json:"changelog"`
	Dependencies  []ModrinthDependency `json:"dependencies"`
	Files         []struct {
		URL      string            `json:"url"`
		Filename string            `json:"filename"`
		Primary  bool              `json:"primary"`
		Size     int64             `json:"size"`
		Hashes   map[string]string `json:"hashes"`
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
	return SearchModrinthPluginsPage(context.Background(), query, limit, 0, loader, mcVersion, sortBy, projectType...)
}

// SearchModrinthPluginsPage keeps pagination bounded independently of the upstream.
func SearchModrinthPluginsPage(ctx context.Context, query string, limit, offset int, loader, mcVersion, sortBy string, projectType ...string) (*ModrinthSearchResult, error) {
	limit, offset = BoundModrinthPage(limit, offset)
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	facet := func(value string) string { b, _ := json.Marshal([]string{value}); return string(b) }

	var facetGroups []string
	loaderLow := strings.ToLower(strings.TrimSpace(loader))
	reqType := ""
	if len(projectType) > 0 {
		pt := strings.ToLower(strings.TrimSpace(projectType[0]))
		if pt == "mod" || pt == "mods" {
			reqType = "mod"
		} else if pt == "modpack" || pt == "modpacks" {
			reqType = "modpack"
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

	if reqType == "modpack" {
		facetGroups = append(facetGroups, `["project_type:modpack"]`)
		if loaderLow != "" && loaderLow != "any" && loaderLow != "all" {
			if loaderLow == "fabric" || loaderLow == "forge" || loaderLow == "neoforge" || loaderLow == "quilt" {
				facetGroups = append(facetGroups, fmt.Sprintf(`["categories:%s"]`, loaderLow))
			}
		}
	} else if reqType == "mod" {
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
		case "paper", "purpur", "leaf", "leafmc":
			facetGroups = append(facetGroups, `["categories:paper","categories:spigot","categories:bukkit"]`)
		case "spigot", "bukkit":
			facetGroups = append(facetGroups, `["categories:spigot","categories:bukkit"]`)
		case "waterfall":
			facetGroups = append(facetGroups, facet("categories:bungeecord"))
		default:
			facetGroups = append(facetGroups, facet("categories:"+loaderLow))
		}
	}

	verTrim := strings.TrimSpace(mcVersion)
	if verTrim != "" && !strings.EqualFold(verTrim, "any") {
		facetGroups = append(facetGroups, facet("versions:"+verTrim))
	}

	index := "relevance"
	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "downloads", "follows", "newest", "updated", "relevance":
		index = strings.ToLower(strings.TrimSpace(sortBy))
	}

	facets := "[" + strings.Join(facetGroups, ",") + "]"
	apiPath := fmt.Sprintf("/search?query=%s&facets=%s&index=%s&limit=%d&offset=%d",
		url.QueryEscape(query), url.QueryEscape(facets), url.QueryEscape(index), limit, offset)

	var result ModrinthSearchResult
	if err := modrinthMetadata(ctx, apiPath, &result); err != nil {
		return nil, err
	}

	result.Limit, result.Offset = limit, offset
	if len(result.Hits) > limit {
		result.Hits = result.Hits[:limit]
	}
	if result.Hits == nil {
		result.Hits = []ModrinthPluginHit{}
	}
	for i := range result.Hits {
		// Search already carries artwork: avoid a details request for every card.
		if len(result.Hits[i].Gallery) > 8 {
			result.Hits[i].Gallery = result.Hits[i].Gallery[:8]
		}
		if result.Hits[i].Gallery == nil {
			result.Hits[i].Gallery = []string{}
		}
	}
	return &result, nil
}

// CompatibleModrinthVersions returns releases supporting the actual server loader.
// Minecraft versions are informational, not an installation restriction.
func GetModrinthProjectVersions(ctx context.Context, identifier string) ([]ModrinthVersion, error) {
	ctx, cancel := modrinthMetadataContext(ctx)
	defer cancel()
	p, err := fetchModrinthProject(ctx, identifier)
	if err != nil {
		return nil, err
	}
	versions, _, err := fetchModrinthVersions(ctx, p.ID, nil)
	return versions, err
}

func CompatibleModrinthVersions(project, loader, mcVersion string) ([]ModrinthVersion, error) {
	return CompatibleModrinthVersionsContext(context.Background(), project, loader, mcVersion)
}

func CompatibleModrinthVersionsContext(ctx context.Context, project, loader, mcVersion string) ([]ModrinthVersion, error) {
	ctx, cancel := modrinthMetadataContext(ctx)
	defer cancel()
	p, err := fetchModrinthProject(ctx, project)
	if err != nil {
		return nil, err
	}
	versions, truncated, err := fetchModrinthVersions(ctx, p.ID, modrinthLoaders(loader))
	if err != nil {
		return nil, err
	}
	eligible := filterCompatibleVersions(versions, loader, mcVersion)
	if truncated && len(eligible) == 0 {
		return nil, fmt.Errorf("Modrinth eligible releases unknown within the version scan limit; choose an explicit version or retry later")
	}
	return eligible, nil
}

// Keep upstream filtering aligned with the local installation policy.
func modrinthLoaders(loader string) []string {
	loader = strings.ToLower(strings.TrimSpace(loader))
	switch loader {
	case "paper", "purpur", "leaf", "leafmc":
		return []string{"paper", "spigot", "bukkit"}
	case "spigot":
		return []string{"spigot", "bukkit"}
	case "waterfall":
		return []string{"bungeecord"}
	default:
		return []string{loader}
	}
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
		loaderOK, jarOK := false, false
		for _, l := range v.Loaders {
			l = strings.ToLower(l)
			if l == loader || (loader == "paper" && (l == "spigot" || l == "bukkit")) || (loader == "spigot" && l == "bukkit") {
				loaderOK = true
			}
		}
		for _, f := range v.Files {
			if safeModrinthJar(f.Filename, f.URL, f.Size) {
				jarOK = true
			}
		}
		// Folia must be explicitly tagged; Paper compatibility alone is not sufficient.
		if loaderOK && jarOK {
			result = append(result, v)
		}
	}
	sortModrinthVersions(result)
	return result
}

func validPluginJarName(name string) bool {
	if name == "" || strings.ContainsAny(name, "/\\:") || strings.Contains(name, "..") || !strings.HasSuffix(strings.ToLower(name), ".jar") {
		return false
	}
	for _, c := range name {
		if c < 32 || c == 127 {
			return false
		}
	}
	return true
}

// DownloadModrinthPlugin revalidates the selected release against the server, then
// publishes a completed JAR without replacing an existing enabled or disabled file.
func DownloadModrinthPlugin(project, destDir, loader, mcVersion, versionID string) (string, error) {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}
	root, err := rooted.Open(destDir)
	if err != nil {
		return "", err
	}
	defer root.Close()
	return DownloadModrinthPluginRoot(project, root, loader, mcVersion, versionID)
}

func DownloadModrinthPluginRoot(project string, root *rooted.Root, loader, mcVersion, versionID string) (string, error) {
	return DownloadModrinthPluginRootContext(context.Background(), project, root, loader, mcVersion, versionID)
}

// ResolveModrinthVersionContext resolves only the requested installation target,
// without making unrelated history pages a prerequisite for confirmation.
func ResolveModrinthVersionContext(ctx context.Context, project, loader, versionID string) (*ModrinthVersion, error) {
	if versionID != "" && !ValidModrinthIdentifier(versionID) {
		return nil, fmt.Errorf("invalid Modrinth version identifier")
	}
	ctx, cancel := modrinthMetadataContext(ctx)
	defer cancel()
	p, err := fetchModrinthProject(ctx, project)
	if err != nil {
		return nil, err
	}
	if versionID != "" {
		return selectedModrinthVersion(ctx, p.ID, loader, versionID)
	}
	return latestModrinthVersion(ctx, p.ID, loader)
}

func DownloadModrinthPluginRootContext(ctx context.Context, project string, root *rooted.Root, loader, mcVersion, versionID string) (string, error) {
	if versionID != "" && !ValidModrinthIdentifier(versionID) {
		return "", fmt.Errorf("invalid Modrinth version identifier")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	chosen, err := ResolveModrinthVersionContext(ctx, project, loader, versionID)
	if err != nil {
		return "", err
	}
	if chosen == nil {
		return "", fmt.Errorf("no installable release for this project and server software (%s); choose a version supporting this loader", loader)
	}
	fileIndex := -1
	for i, f := range chosen.Files {
		if safeModrinthJar(f.Filename, f.URL, f.Size) && (fileIndex < 0 || f.Primary) {
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
	dest := f.Filename
	for _, path := range []string{dest, dest + ".disabled"} {
		if _, err := root.Lstat(path); err == nil {
			return "", fmt.Errorf("%s already exists; remove it explicitly before reinstalling", filepath.Base(path))
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}
	resp, err := modrinthGet(ctx, f.URL)
	if err != nil {
		return "", modrinthRequestError(ctx, "download")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Modrinth download returned HTTP %d; retry later", resp.StatusCode)
	}
	dir, stage, err := root.TempDir()
	if err != nil {
		return "", err
	}
	defer root.RemoveAll(dir)
	defer stage.Close()
	tmp, err := stage.OpenFile("plugin", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return "", err
	}
	n, copyErr := rooted.Copy(ctx, tmp, resp.Body, 200<<20)
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
	// Link is atomic and, unlike Rename on Unix, never overwrites the destination.
	if err := root.Link(filepath.Join(dir, "plugin"), dest); err != nil {
		return "", fmt.Errorf("cannot publish plugin (file may already exist): %w", err)
	}
	return f.Filename, nil
}
