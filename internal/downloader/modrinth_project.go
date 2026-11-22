package downloader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"
)

const maxModrinthVersions = 500

// These allowlisted public v2 fields deliberately exclude moderation, financial,
// payout and other private fields. Text/Markdown and external links remain
// untrusted presentation data; none of those links are fetched by the panel.
type ModrinthProject struct {
	ID                   string                 `json:"id"`
	Slug                 string                 `json:"slug"`
	Title                string                 `json:"title"`
	Description          string                 `json:"description"`
	Body                 string                 `json:"body"`
	IconURL              *string                `json:"icon_url"`
	ProjectType          string                 `json:"project_type"`
	Downloads            int64                  `json:"downloads"`
	Followers            int64                  `json:"followers"`
	Categories           []string               `json:"categories"`
	AdditionalCategories []string               `json:"additional_categories"`
	Loaders              []string               `json:"loaders"`
	GameVersions         []string               `json:"game_versions"`
	ClientSide           string                 `json:"client_side"`
	ServerSide           string                 `json:"server_side"`
	License              ModrinthLicense        `json:"license"`
	SourceURL            *string                `json:"source_url"`
	IssuesURL            *string                `json:"issues_url"`
	WikiURL              *string                `json:"wiki_url"`
	DiscordURL           *string                `json:"discord_url"`
	DonationURLs         []ModrinthDonation     `json:"donation_urls"`
	Gallery              []ModrinthGalleryImage `json:"gallery"`
	Published            string                 `json:"published"`
	Updated              string                 `json:"updated"`
	Status               string                 `json:"status"`
	Team                 string                 `json:"team"`
}

type ModrinthLicense struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	URL  *string `json:"url"`
}

type ModrinthDonation struct {
	ID       string `json:"id"`
	Platform string `json:"platform"`
	URL      string `json:"url"`
}

type ModrinthGalleryImage struct {
	URL         string  `json:"url"`
	Featured    bool    `json:"featured"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Created     string  `json:"created"`
	Ordering    int     `json:"ordering"`
}

type ModrinthMember struct {
	User struct {
		Username  string  `json:"username"`
		Name      *string `json:"name"`
		AvatarURL *string `json:"avatar_url"`
	} `json:"user"`
	Role string `json:"role"`
}

type ModrinthDependency struct {
	VersionID      *string `json:"version_id"`
	ProjectID      *string `json:"project_id"`
	FileName       *string `json:"file_name"`
	DependencyType string  `json:"dependency_type"`
}

type ModrinthProjectDetail struct {
	Project               ModrinthProject   `json:"project"`
	Members               []ModrinthMember  `json:"members"`
	Versions              []ModrinthVersion `json:"versions"`
	InstallableVersionIDs []string          `json:"installable_version_ids"`
	Loader                string            `json:"loader"`
	MinecraftVersion      string            `json:"minecraft_version"`
	VersionsTruncated     bool              `json:"versions_truncated"`
}

// ValidModrinthIdentifier rejects separators, escapes, queries and Unicode before
// any path is constructed, including identifiers supplied by upstream metadata.
func ValidModrinthIdentifier(s string) bool {
	if len(s) == 0 || len(s) > 128 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

func BoundModrinthPage(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	if offset > 10000 {
		offset = 10000
	}
	return limit, offset
}

func modrinthGet(ctx context.Context, endpoint string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	// Do not follow untrusted redirect locations, even from the fixed upstream.
	client := *httpClient
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return client.Do(req)
}

func modrinthRequestError(ctx context.Context, kind string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("Modrinth %s canceled or timed out; retry the request: %w", kind, err)
	}
	return fmt.Errorf("Modrinth %s unavailable; check connectivity and retry later", kind)
}

func modrinthMetadata(ctx context.Context, path string, out any) error {
	resp, err := modrinthGet(ctx, "https://api.modrinth.com/v2"+path)
	if err != nil {
		return modrinthRequestError(ctx, "metadata")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Modrinth project data not found; check the project identifier")
		}
		return fmt.Errorf("Modrinth metadata returned HTTP %d; retry later", resp.StatusCode)
	}
	const limit = 4 << 20
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return modrinthRequestError(ctx, "metadata")
	}
	if len(data) > limit {
		return fmt.Errorf("Modrinth metadata exceeds 4 MiB size limit")
	}
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return fmt.Errorf("Modrinth returned empty metadata; retry later")
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("Modrinth returned invalid metadata; retry later")
	}
	return nil
}

func fetchModrinthProject(ctx context.Context, identifier string) (*ModrinthProject, error) {
	if !ValidModrinthIdentifier(identifier) {
		return nil, fmt.Errorf("invalid Modrinth project identifier (use 1-128 ASCII letters, digits, underscores or hyphens)")
	}
	var project ModrinthProject
	if err := modrinthMetadata(ctx, "/project/"+identifier, &project); err != nil {
		return nil, err
	}
	if !ValidModrinthIdentifier(project.ID) || (project.ID != identifier && project.Slug != identifier) {
		return nil, fmt.Errorf("Modrinth returned an invalid or mismatched project")
	}
	return &project, nil
}

func sortModrinthVersions(versions []ModrinthVersion) {
	sort.SliceStable(versions, func(i, j int) bool {
		a, _ := time.Parse(time.RFC3339Nano, versions[i].DatePublished)
		b, _ := time.Parse(time.RFC3339Nano, versions[j].DatePublished)
		return a.After(b)
	})
}

func fetchModrinthVersions(ctx context.Context, projectID string) ([]ModrinthVersion, error) {
	var versions []ModrinthVersion
	if err := modrinthMetadata(ctx, "/project/"+projectID+"/version", &versions); err != nil {
		return nil, err
	}
	// The endpoint is project-scoped, but still validate membership explicitly.
	result := make([]ModrinthVersion, 0, len(versions))
	for _, v := range versions {
		if v.ProjectID == projectID && ValidModrinthIdentifier(v.ID) {
			result = append(result, v)
		}
	}
	sortModrinthVersions(result)
	return result, nil
}

func safeModrinthJar(name, rawURL string, size int64) bool {
	if !validPluginJarName(name) || size < 0 || size > 200<<20 {
		return false
	}
	u, err := url.Parse(rawURL)
	return err == nil && u.Scheme == "https" && u.Host == "cdn.modrinth.com" && u.User == nil && u.Fragment == ""
}

// GetModrinthProjectDetail uses only three sequential, fixed-host metadata
// requests and a single overall deadline: no per-version fanout or global cache.
func GetModrinthProjectDetail(ctx context.Context, identifier, loader, mcVersion string) (*ModrinthProjectDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	project, err := fetchModrinthProject(ctx, identifier)
	if err != nil {
		return nil, err
	}
	versions, err := fetchModrinthVersions(ctx, project.ID)
	if err != nil {
		return nil, err
	}
	truncated := len(versions) > maxModrinthVersions
	if truncated {
		eligible := filterCompatibleVersions(versions, loader, mcVersion)
		versions = versions[:maxModrinthVersions]
		// Always keep the latest actual-software release reachable, even if
		// hundreds of newer releases target other loaders.
		if len(eligible) > 0 {
			found := false
			for _, v := range versions {
				found = found || v.ID == eligible[0].ID
			}
			if !found {
				versions[len(versions)-1] = eligible[0]
			}
		}
	}
	members := []ModrinthMember{}
	if !ValidModrinthIdentifier(project.Team) {
		return nil, fmt.Errorf("Modrinth returned an invalid team identifier")
	}
	if err := modrinthMetadata(ctx, "/team/"+project.Team+"/members", &members); err != nil {
		return nil, err
	}
	if members == nil {
		members = []ModrinthMember{}
	}
	ids := make([]string, 0)
	for _, v := range filterCompatibleVersions(versions, loader, mcVersion) {
		ids = append(ids, v.ID)
	}
	return &ModrinthProjectDetail{Project: *project, Members: members, Versions: versions, InstallableVersionIDs: ids, Loader: loader, MinecraftVersion: mcVersion, VersionsTruncated: truncated}, nil
}
