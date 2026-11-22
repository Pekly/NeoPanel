package downloader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func modrinthResponse(data string) *http.Response {
	return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(data))}
}

func mockModrinth(t *testing.T, f modrinthTransport) {
	t.Helper()
	old := httpClient
	t.Cleanup(func() { httpClient = old })
	httpClient = &http.Client{Transport: f}
}

func catalogVersion(id, project, loader, mc, date string) ModrinthVersion {
	var v ModrinthVersion
	raw := fmt.Sprintf(`{"id":%q,"project_id":%q,"loaders":[%q],"game_versions":[%q],"date_published":%q,"version_type":"release","changelog":"# Changes","dependencies":[{"project_id":"dependency","version_id":null,"file_name":null,"dependency_type":"required"}],"files":[{"filename":%q,"url":%q,"size":3}]}`, id, project, loader, mc, date, id+".jar", "https://cdn.modrinth.com/"+id+".jar")
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		panic(err)
	}
	return v
}

func mockProjectCatalog(t *testing.T, versions []ModrinthVersion) {
	t.Helper()
	versions = append([]ModrinthVersion(nil), versions...)
	sortModrinthVersions(versions)
	mockModrinth(t, func(r *http.Request) (*http.Response, error) {
		if _, ok := r.Context().Deadline(); !ok {
			t.Fatal("missing deadline")
		}
		if r.URL.Host == "cdn.modrinth.com" {
			return modrinthResponse("jar"), nil
		}
		if r.URL.Scheme != "https" || r.URL.Host != "api.modrinth.com" {
			t.Fatalf("unexpected URL %s", r.URL)
		}
		if strings.HasPrefix(r.URL.Path, "/v2/version/") {
			id := strings.TrimPrefix(r.URL.Path, "/v2/version/")
			for _, v := range versions {
				if v.ID == id {
					data, _ := json.Marshal(v)
					return modrinthResponse(string(data)), nil
				}
			}
			response := modrinthResponse(`{}`)
			response.StatusCode = 404
			return response, nil
		}
		switch r.URL.Path {
		case "/v2/project/example", "/v2/project/P1":
			return modrinthResponse(`{"id":"P1","slug":"example","title":"Example","body":"# Untrusted <script>alert(1)</script>","team":"T1","license":{"id":"MIT","name":"MIT","url":null},"private_token":"private-secret"}`), nil
		case "/v2/project/P1/version":
			return mockVersionPage(t, r, versions), nil
		case "/v2/team/T1/members":
			return modrinthResponse(`[{"user":{"username":"author","name":"Author","avatar_url":"https://example.invalid/avatar","email":"private-secret","payout_data":{"balance":100}},"role":"Owner","payouts_split":100,"permissions":65535,"accepted":true}]`), nil
		default:
			t.Fatalf("unexpected request %s", r.URL)
			return nil, fmt.Errorf("unexpected request")
		}
	})
}

func mockVersionPage(t *testing.T, r *http.Request, versions []ModrinthVersion) *http.Response {
	t.Helper()
	q := r.URL.Query()
	limit, e1 := strconv.Atoi(q.Get("limit"))
	offset, e2 := strconv.Atoi(q.Get("offset"))
	if e1 != nil || e2 != nil || limit != modrinthVersionPageSize || offset < 0 || offset >= maxModrinthVersions || q.Has("game_versions") {
		t.Fatalf("unbounded or MC-filtered request: %s", r.URL)
	}
	var loaders []string
	if q.Has("loaders") {
		if err := json.Unmarshal([]byte(q.Get("loaders")), &loaders); err != nil {
			t.Fatal(err)
		}
	}
	filtered := make([]ModrinthVersion, 0)
	for _, v := range versions {
		matched := len(loaders) == 0
		for _, l := range loaders {
			for _, vl := range v.Loaders {
				matched = matched || l == vl
			}
		}
		if matched {
			filtered = append(filtered, v)
		}
	}
	sortModrinthVersions(filtered)
	if offset > len(filtered) {
		offset = len(filtered)
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	data, _ := json.Marshal(filtered[offset:end])
	return modrinthResponse(string(data))
}

func TestModrinthLatestAndExplicitIgnoreMinecraftVersion(t *testing.T) {
	versions := []ModrinthVersion{
		catalogVersion("old", "P1", "paper", "1.20.1", "2024-01-01T00:00:00Z"),
		catalogVersion("foreign", "P2", "paper", "1.20.1", "2026-01-01T00:00:00Z"),
		catalogVersion("new", "P1", "paper", "1.22", "2025-01-01T00:00:00Z"),
		catalogVersion("fabric", "P1", "fabric", "1.20.1", "2026-01-01T00:00:00Z"),
	}
	mockProjectCatalog(t, versions)
	got, err := CompatibleModrinthVersions("example", "paper", "1.20.1")
	if err != nil || len(got) != 2 || got[0].ID != "new" || got[1].ID != "old" {
		t.Fatalf("versions: %+v %v", got, err)
	}
	for _, tc := range []struct{ version, want string }{{"", "new.jar"}, {"new", "new.jar"}, {"old", "old.jar"}, {"foreign", ""}, {"fabric", ""}, {"missing", ""}, {"../new", ""}} {
		t.Run("version_"+tc.version, func(t *testing.T) {
			dir := t.TempDir()
			name, err := DownloadModrinthPlugin("example", dir, "paper", "1.19", tc.version)
			if tc.want == "" {
				if err == nil {
					t.Fatalf("accepted %s", tc.version)
				}
				entries, _ := os.ReadDir(dir)
				if len(entries) != 0 {
					t.Fatalf("rejected install left files: %v", entries)
				}
			} else {
				if err != nil || name != tc.want {
					t.Fatalf("install %q: %q %v", tc.version, name, err)
				}
				data, err := os.ReadFile(filepath.Join(dir, name))
				if err != nil || string(data) != "jar" {
					t.Fatalf("download content: %q %v", data, err)
				}
			}
		})
	}
}

func TestModrinthLoaderMappings(t *testing.T) {
	versions := []ModrinthVersion{}
	for _, loader := range []string{"paper", "spigot", "bukkit", "folia", "bungeecord", "velocity", "fabric", "forge", "neoforge"} {
		versions = append(versions, catalogVersion(loader, "P1", loader, "different", "2025-01-01T00:00:00Z"))
	}
	for _, tc := range []struct{ loader, want string }{
		{"paper", "paper,spigot,bukkit"}, {"purpur", "paper,spigot,bukkit"}, {"leaf", "paper,spigot,bukkit"}, {"leafmc", "paper,spigot,bukkit"},
		{"spigot", "spigot,bukkit"}, {"bukkit", "bukkit"}, {"folia", "folia"}, {"waterfall", "bungeecord"}, {"velocity", "velocity"},
		{"fabric", "fabric"}, {"forge", "forge"}, {"neoforge", "neoforge"}, {"vanilla", ""},
	} {
		var ids []string
		for _, v := range filterCompatibleVersions(versions, tc.loader, "") {
			ids = append(ids, v.ID)
		}
		if strings.Join(ids, ",") != tc.want {
			t.Errorf("%s: %v", tc.loader, ids)
		}
	}
}

func TestModrinthProjectDetailPublicFieldsAndAllVersions(t *testing.T) {
	versions := []ModrinthVersion{
		catalogVersion("old", "P1", "paper", "1.20.1", "2024-01-01T00:00:00Z"),
		catalogVersion("new", "P1", "paper", "1.22", "2025-01-01T00:00:00Z"),
		catalogVersion("fabric", "P1", "fabric", "1.22", "2025-02-01T00:00:00Z"),
		catalogVersion("unsafe", "P1", "paper", "1.22", "2025-03-01T00:00:00Z"),
	}
	versions[3].Files[0].Filename = "../escape.jar"
	mockProjectCatalog(t, versions)
	detail, err := GetModrinthProjectDetail(context.Background(), "example", "paper", "1.20.1")
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, v := range detail.Versions {
		ids = append(ids, v.ID)
	}
	if !reflect.DeepEqual(ids, []string{"unsafe", "fabric", "new", "old"}) {
		t.Fatal(ids)
	}
	if !reflect.DeepEqual(detail.InstallableVersionIDs, []string{"new", "old"}) {
		t.Fatal(detail.InstallableVersionIDs)
	}
	if detail.Loader != "paper" || detail.MinecraftVersion != "1.20.1" || detail.Project.Body == "" || detail.Project.License.ID != "MIT" || detail.Members[0].User.Username != "author" {
		t.Fatalf("incomplete detail: %+v", detail)
	}
	if detail.Versions[0].Changelog != "# Changes" || detail.Versions[0].VersionType != "release" || len(detail.Versions[0].Dependencies) != 1 {
		t.Fatal("missing version metadata")
	}
	encoded, _ := json.Marshal(detail)
	for _, private := range []string{"private-secret", "payout", "permissions", "accepted", "email", "private_token"} {
		if strings.Contains(string(encoded), private) {
			t.Fatalf("leaked private data %s", private)
		}
	}
	// JSON encoding is safe for embedding, but clients must sanitize Markdown/links.
	if strings.Contains(string(encoded), "<script>") {
		t.Fatal("unescaped HTML in JSON")
	}
}

func TestModrinthProjectDetailCapAfterSorting(t *testing.T) {
	versions := make([]ModrinthVersion, 510)
	for i := range versions {
		versions[i] = catalogVersion("v"+strconv.Itoa(i), "P1", "paper", "different", time.Date(2020, 1, 1, 0, i, 0, 0, time.UTC).Format(time.RFC3339))
	}
	mockProjectCatalog(t, versions)
	detail, err := GetModrinthProjectDetail(context.Background(), "P1", "paper", "")
	if err != nil || len(detail.Versions) != 500 || len(detail.InstallableVersionIDs) != 500 || detail.Versions[0].ID != "v509" || detail.Versions[499].ID != "v10" {
		t.Fatalf("bad cap: %+v %v", detail, err)
	}
}

func TestModrinthProjectCapKeepsLatestSoftwareRelease(t *testing.T) {
	versions := make([]ModrinthVersion, 3736)
	for i := range versions {
		loader := "fabric"
		if i < 2 {
			loader = "paper"
		}
		versions[i] = catalogVersion("v"+strconv.Itoa(i), "P1", loader, "different", time.Date(2020, 1, 1, 0, i, 0, 0, time.UTC).Format(time.RFC3339))
		versions[i].Changelog = strings.Repeat("release notes ", 160)
	}
	fullCatalog, _ := json.Marshal(versions)
	if len(fullCatalog) <= 4<<20 {
		t.Fatal("fixture does not reproduce oversized catalog")
	}
	mockProjectCatalog(t, versions)
	transport := httpClient.Transport
	calls := 0
	httpClient.Transport = modrinthTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		return transport.RoundTrip(r)
	})
	detail, err := GetModrinthProjectDetail(context.Background(), "P1", "paper", "1.20")
	if err != nil {
		t.Fatal(err)
	}
	if !detail.VersionsTruncated || len(detail.Versions) != 500 || detail.Versions[499].ID != "v1" || !reflect.DeepEqual(detail.InstallableVersionIDs, []string{"v1"}) {
		t.Fatalf("latest software release hidden by cap: %+v", detail)
	}
	if calls != 13 || detail.LatestVersionStatus != "available" || detail.LatestVersionID != "v1" || len(detail.Warnings) != 0 {
		t.Fatalf("not bounded: calls=%d status=%s warnings=%v", calls, detail.LatestVersionStatus, detail.Warnings)
	}
	for _, tc := range []struct{ id, want string }{{"", "v1.jar"}, {"v0", "v0.jar"}, {"v3735", ""}} {
		calls = 0
		name, err := DownloadModrinthPlugin("P1", t.TempDir(), "paper", "unrelated-mc", tc.id)
		if tc.want == "" {
			if err == nil || calls != 2 {
				t.Fatalf("invalid explicit selection accepted: %q %v calls=%d", name, err, calls)
			}
		} else if err != nil || name != tc.want || calls != 3 {
			t.Fatalf("selection: %q %v calls=%d", name, err, calls)
		}
	}
}

func TestModrinthSearchPaginationAndFacets(t *testing.T) {
	for _, tc := range []struct{ limit, offset, wantLimit, wantOffset int }{{24, 24, 24, 24}, {100, 20000, 50, 10000}, {-1, -10, 20, 0}} {
		mockModrinth(t, func(r *http.Request) (*http.Response, error) {
			if r.URL.Host != "api.modrinth.com" || r.URL.Path != "/v2/search" {
				t.Fatal(r.URL)
			}
			q := r.URL.Query()
			if q.Get("offset") != strconv.Itoa(tc.wantOffset) || q.Get("limit") != strconv.Itoa(tc.wantLimit) {
				t.Fatal(q)
			}
			var facets [][]string
			if err := json.Unmarshal([]byte(q.Get("facets")), &facets); err != nil {
				t.Fatal("invalid facets", err)
			}
			if !reflect.DeepEqual(facets, [][]string{{"project_type:plugin"}, {"categories:folia"}, {"versions:1.\"injected"}}) {
				t.Fatal(facets)
			}
			return modrinthResponse(`{"hits":[],"total_hits":72,"offset":999999,"limit":99999}`), nil
		})
		result, err := SearchModrinthPluginsPage(context.Background(), "query", tc.limit, tc.offset, "folia", `1."injected`, "downloads")
		if err != nil || result.Offset != tc.wantOffset || result.Limit != tc.wantLimit || result.TotalHits != 72 {
			t.Fatalf("page: %+v %v", result, err)
		}
	}
}

func TestModrinthIdentifierValidationBeforeNetwork(t *testing.T) {
	mockModrinth(t, func(*http.Request) (*http.Response, error) {
		t.Fatal("network reached for invalid identifier")
		return nil, nil
	})
	for _, id := range []string{"", "..", "../project", "a/b", "a%2fb", "a?x=y", "a#fragment", "a\\b", "https://evil.invalid", " a", "é", strings.Repeat("a", 129)} {
		if _, err := GetModrinthProjectDetail(context.Background(), id, "paper", "1.21"); err == nil {
			t.Errorf("accepted %q", id)
		}
		if _, err := CompatibleModrinthVersions(id, "paper", "1.21"); err == nil {
			t.Errorf("versions accepted %q", id)
		}
	}
	for _, id := range []string{"Abc09_-", strings.Repeat("a", 128)} {
		if !ValidModrinthIdentifier(id) {
			t.Fatal(id)
		}
	}
}

func TestModrinthDetailBoundsAndCancellationAtEachRequest(t *testing.T) {
	for _, target := range []string{"/v2/project/example", "/v2/project/P1/version", "/v2/team/T1/members"} {
		for _, mode := range []string{"oversize", "cancel", "upstream", "invalid", "redirect"} {
			t.Run(target+"/"+mode, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				mockModrinth(t, func(r *http.Request) (*http.Response, error) {
					if r.URL.Host != "api.modrinth.com" {
						t.Fatal("followed untrusted URL", r.URL)
					}
					if deadline, ok := r.Context().Deadline(); !ok || time.Until(deadline) > 31*time.Second {
						t.Fatal("missing bounded deadline")
					}
					if r.URL.Path == target {
						switch mode {
						case "oversize":
							return modrinthResponse(strings.Repeat(" ", (4<<20)+1)), nil
						case "cancel":
							cancel()
							<-r.Context().Done()
							return nil, r.Context().Err()
						case "upstream":
							return nil, fmt.Errorf("private-secret-internal-host")
						case "invalid":
							return modrinthResponse(`{"private-secret":`), nil
						case "redirect":
							response := modrinthResponse("")
							response.StatusCode = 302
							response.Header.Set("Location", "https://evil.invalid/private-secret")
							return response, nil
						}
					}
					switch r.URL.Path {
					case "/v2/project/example":
						return modrinthResponse(`{"id":"P1","slug":"example","team":"T1"}`), nil
					case "/v2/project/P1/version", "/v2/team/T1/members":
						return modrinthResponse(`[]`), nil
					default:
						t.Fatalf("unexpected path: %s", r.URL.Path)
						return nil, nil
					}
				})
				detail, err := GetModrinthProjectDetail(ctx, "example", "paper", "1.21")
				if target == "/v2/project/example" || mode == "cancel" {
					if err == nil || strings.Contains(err.Error(), "private-secret") {
						t.Fatal("missing or unsafe error", err)
					}
					if mode == "cancel" && !errors.Is(err, context.Canceled) {
						t.Fatal(err)
					}
					if mode == "oversize" && !strings.Contains(err.Error(), "size limit") {
						t.Fatal(err)
					}
				} else {
					if err != nil || detail == nil || len(detail.Warnings) == 0 {
						t.Fatalf("lost partial detail: %+v %v", detail, err)
					}
					encoded, _ := json.Marshal(detail)
					if strings.Contains(string(encoded), "private-secret") {
						t.Fatal("unsafe warning")
					}
					if target == "/v2/project/P1/version" && detail.LatestVersionStatus != "unknown" {
						t.Fatal("unknown catalog reported as absent")
					}
				}
			})
		}
	}
}

func TestModrinthUntrustedProjectIdentifiers(t *testing.T) {
	for _, raw := range []string{`{"id":"../bad","slug":"example","team":"T1"}`, `{"id":"P2","slug":"different","team":"T1"}`, `{"id":"P1","slug":"example","team":"T1%2f.."}`} {
		mockModrinth(t, func(r *http.Request) (*http.Response, error) {
			switch r.URL.Path {
			case "/v2/project/example":
				return modrinthResponse(raw), nil
			case "/v2/project/P1/version":
				return modrinthResponse(`[]`), nil
			default:
				t.Fatal("unsafe upstream identifier followed", r.URL)
				return nil, nil
			}
		})
		detail, err := GetModrinthProjectDetail(context.Background(), "example", "paper", "")
		if strings.Contains(raw, "T1%2f") {
			if err != nil || len(detail.Warnings) != 1 || detail.Warnings[0].Section != "members" {
				t.Fatalf("missing team warning: %+v %v", detail, err)
			}
		} else if err == nil {
			t.Fatal("accepted unsafe project metadata")
		}
	}
}

func TestModrinthSafeJar(t *testing.T) {
	for _, tc := range []struct {
		name, url string
		size      int64
	}{
		{"../escape.jar", "https://cdn.modrinth.com/file", 1}, {"C:escape.jar", "https://cdn.modrinth.com/file", 1}, {"bad\x00.jar", "https://cdn.modrinth.com/file", 1},
		{"plugin.zip", "https://cdn.modrinth.com/file", 1}, {"plugin.jar", "https://evil.invalid/file", 1}, {"plugin.jar", "http://cdn.modrinth.com/file", 1},
		{"plugin.jar", "https://cdn.modrinth.com.evil.invalid/file", 1}, {"plugin.jar", "https://cdn.modrinth.com@evil.invalid/file", 1},
		{"plugin.jar", "https://cdn.modrinth.com/file", -1}, {"plugin.jar", "https://cdn.modrinth.com/file", (200 << 20) + 1},
	} {
		if safeModrinthJar(tc.name, tc.url, tc.size) {
			t.Errorf("accepted %+v", tc)
		}
	}
	if !safeModrinthJar("plugin.jar", "https://cdn.modrinth.com/data/P1/versions/v1/plugin.jar", 200<<20) {
		t.Fatal("rejected valid jar")
	}
}

func TestModrinthPartialSections(t *testing.T) {
	versions := make([]ModrinthVersion, 120)
	for i := range versions {
		versions[i] = catalogVersion(fmt.Sprint("v", i), "P1", "paper", "different", time.Date(2024, 1, 1, 0, i, 0, 0, time.UTC).Format(time.RFC3339))
	}
	for _, section := range []string{"members", "latest_version", "versions"} {
		t.Run(section, func(t *testing.T) {
			mockProjectCatalog(t, versions)
			transport := httpClient.Transport
			httpClient.Transport = modrinthTransport(func(r *http.Request) (*http.Response, error) {
				q := r.URL.Query()
				fail := section == "members" && strings.Contains(r.URL.Path, "/team/") || section == "latest_version" && q.Has("loaders") || section == "versions" && !q.Has("loaders") && q.Get("offset") == "50"
				if fail {
					response := modrinthResponse("private-secret")
					response.StatusCode = 429
					return response, nil
				}
				return transport.RoundTrip(r)
			})
			detail, err := GetModrinthProjectDetail(context.Background(), "P1", "paper", "1.20")
			if err != nil || detail.Project.Body == "" || len(detail.Warnings) != 1 || detail.Warnings[0].Section != section {
				t.Fatalf("partial: %+v %v", detail, err)
			}
			if section == "versions" && (len(detail.Versions) != 50 || !detail.VersionsTruncated) {
				t.Fatal("discarded successful pages")
			}
			if section == "latest_version" {
				if detail.LatestVersionStatus != "unknown" || detail.LatestVersionID != "" || len(detail.Versions) != 120 {
					t.Fatal("false latest certainty")
				}
			} else if detail.LatestVersionStatus != "available" || detail.LatestVersionID != "v119" {
				t.Fatal("lost latest release")
			}
		})
	}
}

func TestModrinthLatestScanBudgetIsUnknownNotAbsent(t *testing.T) {
	versions := make([]ModrinthVersion, 600)
	for i := range versions {
		versions[i] = catalogVersion(fmt.Sprint("v", i), "P1", "paper", "different", time.Date(2024, 1, 1, 0, i, 0, 0, time.UTC).Format(time.RFC3339))
		if i > 0 {
			versions[i].Files[0].Filename = "unsafe.zip"
		}
	}
	mockProjectCatalog(t, versions)
	transport := httpClient.Transport
	calls := 0
	httpClient.Transport = modrinthTransport(func(r *http.Request) (*http.Response, error) { calls++; return transport.RoundTrip(r) })
	detail, err := GetModrinthProjectDetail(context.Background(), "P1", "paper", "")
	if err != nil || detail.LatestVersionStatus != "unknown" || !detail.VersionsTruncated || len(detail.InstallableVersionIDs) != 0 || len(detail.Warnings) != 1 || calls != 22 {
		t.Fatalf("scan not bounded/unknown: calls=%d detail=%+v err=%v", calls, detail, err)
	}
	calls = 0
	if _, err := CompatibleModrinthVersions("P1", "paper", ""); err == nil || calls != 11 {
		t.Fatalf("false absence: %v calls=%d", err, calls)
	}
	calls = 0
	if _, err := DownloadModrinthPlugin("P1", t.TempDir(), "paper", "", ""); err == nil || !strings.Contains(err.Error(), "scan limit") || calls != 11 {
		t.Fatalf("unsafe latest: %v calls=%d", err, calls)
	}
	calls = 0
	if name, err := DownloadModrinthPlugin("P1", t.TempDir(), "paper", "", "v0"); err != nil || name != "v0.jar" || calls != 3 {
		t.Fatalf("explicit version unreachable: %s %v", name, err)
	}
}

func TestModrinthTotalMetadataBudget(t *testing.T) {
	calls := 0
	mockModrinth(t, func(r *http.Request) (*http.Response, error) {
		calls++
		return modrinthResponse(`[]` + strings.Repeat(" ", (4<<20)-2)), nil
	})
	ctx, cancel := modrinthMetadataContext(context.Background())
	defer cancel()
	for i := 0; i < 4; i++ {
		var out []ModrinthVersion
		if err := modrinthMetadata(ctx, "/project/P1/version?limit=50&offset=0", &out); err != nil {
			t.Fatal(err)
		}
	}
	var out []ModrinthVersion
	if err := modrinthMetadata(ctx, "/project/P1/version?limit=50&offset=0", &out); err == nil || !strings.Contains(err.Error(), "total size limit") || calls != 4 {
		t.Fatalf("total limit: %v calls=%d", err, calls)
	}
}

func TestModrinthRejectsIgnoredPagination(t *testing.T) {
	versions := make([]ModrinthVersion, 51)
	for i := range versions {
		versions[i] = catalogVersion(fmt.Sprint("v", i), "P1", "paper", "", "2025-01-01T00:00:00Z")
	}
	mockProjectCatalog(t, versions)
	transport := httpClient.Transport
	httpClient.Transport = modrinthTransport(func(r *http.Request) (*http.Response, error) {
		if strings.HasSuffix(r.URL.Path, "/version") {
			data, _ := json.Marshal(versions)
			return modrinthResponse(string(data)), nil
		}
		return transport.RoundTrip(r)
	})
	detail, err := GetModrinthProjectDetail(context.Background(), "P1", "paper", "")
	if err != nil || detail.LatestVersionStatus != "unknown" || len(detail.Warnings) != 2 || detail.Project.Body == "" {
		t.Fatalf("ignored pagination: %+v %v", detail, err)
	}
}
