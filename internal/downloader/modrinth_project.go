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

const (
	maxModrinthVersions      = 500
	modrinthVersionPageSize  = 50
	maxModrinthMetadataBytes = 16 << 20
)

// All metadata requests for an operation are sequential and share this budget.
type modrinthBudgetKey struct{}
type modrinthBudget struct{ remaining int64 }

func modrinthMetadataContext(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	return ModrinthMetadataBudgetContext(ctx), cancel
}

// ModrinthMetadataBudgetContext shares a sequential 16 MiB budget across nested calls.
func ModrinthMetadataBudgetContext(ctx context.Context) context.Context {
	if ctx.Value(modrinthBudgetKey{}) != nil {
		return ctx
	}
	return context.WithValue(ctx, modrinthBudgetKey{}, &modrinthBudget{remaining: maxModrinthMetadataBytes})
}

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

type ModrinthWarning struct {
	Section string `json:"section"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ModrinthProjectDetail struct {
	Project               ModrinthProject   `json:"project"`
	Members               []ModrinthMember  `json:"members"`
	Versions              []ModrinthVersion `json:"versions"`
	InstallableVersionIDs []string          `json:"installable_version_ids"`
	Loader                string            `json:"loader"`
	MinecraftVersion      string            `json:"minecraft_version"`
	VersionsTruncated     bool              `json:"versions_truncated"`
	// Status distinguishes a confirmed absence from an unavailable catalog.
	LatestVersionStatus string            `json:"latest_version_status"` // available, none, unknown
	LatestVersionID     string            `json:"latest_version_id,omitempty"`
	Warnings            []ModrinthWarning `json:"warnings"`
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
	return modrinthMetadataRequest(ctx, http.MethodGet, path, nil, out)
}

func modrinthMetadataRequest(ctx context.Context, method, path string, body io.Reader, out any) error {
	if ctx.Err() != nil {
		return modrinthRequestError(ctx, "metadata")
	}
	budget, _ := ctx.Value(modrinthBudgetKey{}).(*modrinthBudget)
	if budget != nil && budget.remaining <= 0 {
		return fmt.Errorf("Modrinth metadata exceeds total size limit; retry later")
	}
	req, err := http.NewRequestWithContext(ctx, method, "https://api.modrinth.com/v2"+path, body)
	if err != nil {
		return modrinthRequestError(ctx, "metadata")
	}
	req.Header.Set("User-Agent", userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := *httpClient
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
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
	limit := int64(4 << 20)
	if budget != nil && budget.remaining < limit {
		limit = budget.remaining
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if budget != nil {
		budget.remaining -= int64(len(data))
	}
	if err != nil {
		return modrinthRequestError(ctx, "metadata")
	}
	if int64(len(data)) > limit {
		return fmt.Errorf("Modrinth metadata exceeds size limit (4 MiB per response, 16 MiB total)")
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

// v2 supports limit/offset, with newest publication first. Never request the
// unbounded project/version catalog: popular multi-loader projects exceed 4 MiB.
func fetchModrinthVersionPage(ctx context.Context, projectID string, loaders []string, offset int) ([]ModrinthVersion, int, error) {
	q := url.Values{"limit": {fmt.Sprint(modrinthVersionPageSize)}, "offset": {fmt.Sprint(offset)}}
	if len(loaders) > 0 {
		encoded, _ := json.Marshal(loaders)
		q.Set("loaders", string(encoded))
	}
	var versions []ModrinthVersion
	if err := modrinthMetadata(ctx, "/project/"+projectID+"/version?"+q.Encode(), &versions); err != nil {
		return nil, 0, err
	}
	if len(versions) > modrinthVersionPageSize {
		return nil, 0, fmt.Errorf("Modrinth returned an unbounded version page; retry later")
	}
	result := make([]ModrinthVersion, 0, len(versions))
	for _, v := range versions {
		if v.ProjectID == projectID && ValidModrinthIdentifier(v.ID) {
			result = append(result, v)
		}
	}
	sortModrinthVersions(result)
	return result, len(versions), nil
}

// At most ten pages; partial results are only used by the detail endpoint.
func fetchModrinthVersions(ctx context.Context, projectID string, loaders []string) ([]ModrinthVersion, bool, error) {
	result := make([]ModrinthVersion, 0)
	seen := make(map[string]bool)
	for offset := 0; offset < maxModrinthVersions; offset += modrinthVersionPageSize {
		page, count, err := fetchModrinthVersionPage(ctx, projectID, loaders, offset)
		if err != nil {
			return result, true, err
		}
		for _, v := range page {
			if seen[v.ID] {
				return result, true, fmt.Errorf("Modrinth repeated a version page; retry later")
			}
			seen[v.ID] = true
			result = append(result, v)
		}
		sortModrinthVersions(result)
		if count < modrinthVersionPageSize {
			return result, false, nil
		}
	}
	return result, true, nil
}

// Loader filtering happens upstream before pagination, with NO Minecraft filter.
// Thus 500+ newer releases for other loaders cannot hide the latest usable JAR.
func latestModrinthVersion(ctx context.Context, projectID, loader string) (*ModrinthVersion, error) {
	for offset := 0; offset < maxModrinthVersions; offset += modrinthVersionPageSize {
		page, count, err := fetchModrinthVersionPage(ctx, projectID, modrinthLoaders(loader), offset)
		if err != nil {
			return nil, err
		}
		eligible := filterCompatibleVersions(page, loader, "")
		if len(eligible) > 0 {
			return &eligible[0], nil
		}
		if count < modrinthVersionPageSize {
			return nil, nil
		}
	}
	return nil, fmt.Errorf("Modrinth latest release could not be determined within the version scan limit; choose an explicit version or retry later")
}

func selectedModrinthVersion(ctx context.Context, projectID, loader, versionID string) (*ModrinthVersion, error) {
	var version ModrinthVersion
	if err := modrinthMetadata(ctx, "/version/"+versionID, &version); err != nil {
		return nil, err
	}
	if version.ID != versionID || version.ProjectID != projectID || len(filterCompatibleVersions([]ModrinthVersion{version}, loader, "")) == 0 {
		return nil, fmt.Errorf("selected Modrinth version does not belong to this project or support this server software")
	}
	return &version, nil
}

func safeModrinthJar(name, rawURL string, size int64) bool {
	if !validPluginJarName(name) || size < 0 || size > 200<<20 {
		return false
	}
	u, err := url.Parse(rawURL)
	return err == nil && u.Scheme == "https" && u.Host == "cdn.modrinth.com" && u.User == nil && u.Fragment == ""
}

// Optional section failures preserve the available project description. All
// requests share a deadline/byte budget; no fanout, unbounded list or global cache.
func GetModrinthProjectDetail(ctx context.Context, identifier, loader, mcVersion string) (*ModrinthProjectDetail, error) {
	requestCtx := ctx
	ctx, cancel := modrinthMetadataContext(ctx)
	defer cancel()
	project, err := fetchModrinthProject(ctx, identifier)
	if err != nil {
		return nil, err
	}
	detail := &ModrinthProjectDetail{Project: *project, Members: []ModrinthMember{}, Versions: []ModrinthVersion{}, InstallableVersionIDs: []string{}, Loader: loader, MinecraftVersion: mcVersion, LatestVersionStatus: "unknown", Warnings: []ModrinthWarning{}}
	warn := func(section string, err error) {
		detail.Warnings = append(detail.Warnings, ModrinthWarning{Section: section, Code: "upstream_unavailable", Message: err.Error()})
	}
	// Fetch the inexpensive optional team before the catalog can spend the budget.
	if !ValidModrinthIdentifier(project.Team) {
		warn("members", fmt.Errorf("Modrinth returned an invalid team identifier"))
	} else if err := modrinthMetadata(ctx, "/team/"+project.Team+"/members", &detail.Members); err != nil {
		detail.Members = []ModrinthMember{}
		warn("members", err)
	}
	if detail.Members == nil {
		detail.Members = []ModrinthMember{}
	}
	// Resolve latest independently first, so even a broken display page cannot
	// turn an unknown catalog into a false "no installable releases" result.
	latest, latestErr := latestModrinthVersion(ctx, project.ID, loader)
	if latestErr != nil {
		warn("latest_version", latestErr)
	} else if latest == nil {
		detail.LatestVersionStatus = "none"
	} else {
		detail.LatestVersionStatus, detail.LatestVersionID = "available", latest.ID
	}
	detail.Versions, detail.VersionsTruncated, err = fetchModrinthVersions(ctx, project.ID, nil)
	if err != nil {
		warn("versions", err)
	}
	if latest != nil {
		found := false
		for _, v := range detail.Versions {
			found = found || v.ID == latest.ID
		}
		if !found {
			if len(detail.Versions) == maxModrinthVersions {
				detail.Versions = detail.Versions[:maxModrinthVersions-1]
			}
			detail.Versions = append(detail.Versions, *latest)
			sortModrinthVersions(detail.Versions)
		}
	}
	for _, v := range filterCompatibleVersions(detail.Versions, loader, mcVersion) {
		detail.InstallableVersionIDs = append(detail.InstallableVersionIDs, v.ID)
	}
	if requestCtx.Err() != nil {
		return nil, modrinthRequestError(requestCtx, "metadata")
	}
	return detail, nil
}
