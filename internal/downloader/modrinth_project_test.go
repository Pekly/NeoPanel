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
	data, err := json.Marshal(versions)
	if err != nil {
		t.Fatal(err)
	}
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
		switch r.URL.Path {
		case "/v2/project/example", "/v2/project/P1":
			return modrinthResponse(`{"id":"P1","slug":"example","title":"Example","body":"# Untrusted <script>alert(1)</script>","team":"T1","license":{"id":"MIT","name":"MIT","url":null},"private_token":"private-secret"}`), nil
		case "/v2/project/P1/version":
			return modrinthResponse(string(data)), nil
		case "/v2/team/T1/members":
			return modrinthResponse(`[{"user":{"username":"author","name":"Author","avatar_url":"https://example.invalid/avatar","email":"private-secret","payout_data":{"balance":100}},"role":"Owner","payouts_split":100,"permissions":65535,"accepted":true}]`), nil
		default:
			t.Fatalf("unexpected request %s", r.URL)
			return nil, fmt.Errorf("unexpected request")
		}
	})
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
	versions := make([]ModrinthVersion, 510)
	for i := range versions {
		loader := "fabric"
		if i < 2 {
			loader = "paper"
		}
		versions[i] = catalogVersion("v"+strconv.Itoa(i), "P1", loader, "different", time.Date(2020, 1, 1, 0, i, 0, 0, time.UTC).Format(time.RFC3339))
	}
	mockProjectCatalog(t, versions)
	detail, err := GetModrinthProjectDetail(context.Background(), "P1", "paper", "1.20")
	if err != nil {
		t.Fatal(err)
	}
	if !detail.VersionsTruncated || len(detail.Versions) != 500 || detail.Versions[499].ID != "v1" || !reflect.DeepEqual(detail.InstallableVersionIDs, []string{"v1"}) {
		t.Fatalf("latest software release hidden by cap: %+v", detail)
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
					case "/v2/project/P1/version":
						return modrinthResponse(`[]`), nil
					default:
						t.Fatalf("unexpected path: %s", r.URL.Path)
						return nil, nil
					}
				})
				_, err := GetModrinthProjectDetail(ctx, "example", "paper", "1.21")
				if err == nil || strings.Contains(err.Error(), "private-secret") {
					t.Fatal("missing or unsafe error", err)
				}
				if mode == "cancel" && !errors.Is(err, context.Canceled) {
					t.Fatal(err)
				}
				if mode == "oversize" && !strings.Contains(err.Error(), "size limit") {
					t.Fatal(err)
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
		if _, err := GetModrinthProjectDetail(context.Background(), "example", "paper", ""); err == nil {
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
