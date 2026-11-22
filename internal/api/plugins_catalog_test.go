package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/downloader"
)

type pluginCatalogTransport func(*http.Request) (*http.Response, error)

func (f pluginCatalogTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestPluginSearchPageBounds(t *testing.T) {
	for _, tc := range []struct {
		query         string
		limit, offset int
		bad           bool
	}{
		{"", 24, 0, false}, {"limit=50&offset=10000", 50, 10000, false}, {"limit=999&offset=10001", 100, 10000, false},
		{"limit=-2&offset=-3", 24, 0, false}, {"limit=1&offset=24", 1, 24, false}, {"offset=abc", 0, 0, true}, {"limit=999999999999999999999999", 0, 0, true},
	} {
		req := httptest.NewRequest("GET", "/api/plugins/search?"+tc.query, nil)
		limit, offset, err := pluginSearchPage(req)
		if tc.bad {
			if err == nil {
				t.Fatal("accepted invalid page", tc.query)
			}
			w := httptest.NewRecorder()
			(&PluginHandler{}).Search(w, req)
			if w.Code != 400 {
				t.Fatal(w.Code)
			}
		} else if err != nil || limit != tc.limit || offset != tc.offset {
			t.Fatalf("%s: %d %d %v", tc.query, limit, offset, err)
		}
	}
}

func TestPluginCatalogAPIContract(t *testing.T) {
	f := newReliabilityFixture(t)
	oldTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = oldTransport })
	calls := 0
	fail := false
	http.DefaultTransport = pluginCatalogTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if fail {
			return nil, fmt.Errorf("private-secret-host")
		}
		if _, ok := r.Context().Deadline(); !ok {
			t.Fatal("missing upstream deadline")
		}
		data := ""
		if r.URL.Host == "cdn.modrinth.com" {
			data = "jar"
		} else {
			if r.URL.Host != "api.modrinth.com" {
				t.Fatal("unexpected host", r.URL)
			}
			switch r.URL.Path {
			case "/v2/search":
				if r.URL.Query().Get("offset") != "48" || r.URL.Query().Get("limit") != "99" {
					t.Fatal(r.URL)
				}
				data = `{"hits":[{"project_id":"P1","slug":"example","featured_gallery":"https://cdn.modrinth.com/banner.webp","gallery":["https://cdn.modrinth.com/gallery.webp"]}],"total_hits":123}`
			case "/v2/project/example":
				data = `{"id":"P1","slug":"example","team":"T1","body":"# Example","license":{"id":"MIT"}}`
			case "/v2/project/P1/version", "/v2/version/old", "/v2/version/fabric", "/v2/version/foreign":
				data = `[
    {"id":"old","project_id":"P1","date_published":"2024-01-01T00:00:00Z","loaders":["paper"],"game_versions":["1.21"],"files":[{"filename":"old.jar","url":"https://cdn.modrinth.com/old.jar","size":3}]},
    {"id":"new","project_id":"P1","date_published":"2025-01-01T00:00:00Z","version_type":"release","changelog":"# New","dependencies":[],"loaders":["paper"],"game_versions":["1.22"],"files":[{"filename":"new.jar","url":"https://cdn.modrinth.com/new.jar","size":3}]},
    {"id":"fabric","project_id":"P1","date_published":"2026-01-01T00:00:00Z","loaders":["fabric"],"game_versions":["1.21"],"files":[{"filename":"mod.jar","url":"https://cdn.modrinth.com/mod.jar","size":3}]},
    {"id":"foreign","project_id":"P2","loaders":["paper"],"files":[{"filename":"foreign.jar","url":"https://cdn.modrinth.com/foreign.jar","size":3}]}
   ]`
				if strings.HasPrefix(r.URL.Path, "/v2/version/") {
					var versions []downloader.ModrinthVersion
					if err := json.Unmarshal([]byte(data), &versions); err != nil {
						t.Fatal(err)
					}
					for _, v := range versions {
						if v.ID == strings.TrimPrefix(r.URL.Path, "/v2/version/") {
							encoded, _ := json.Marshal(v)
							data = string(encoded)
							break
						}
					}
				} else if r.URL.Query().Get("limit") != "50" || r.URL.Query().Get("offset") != "0" || r.URL.Query().Has("game_versions") {
					t.Fatal("unbounded version request", r.URL)
				}
			case "/v2/team/T1/members":
				data = `[{"user":{"username":"author","name":null,"avatar_url":null,"email":"private-secret"},"role":"Owner","payouts_split":100}]`
			default:
				t.Fatal("unexpected path", r.URL)
			}
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(data))}, nil
	})
	const base = "/api/servers/alpha/plugins/"
	if w := f.request("GET", "/api/plugins/search?offset=48&limit=99", "", "owner"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	} else {
		var page downloader.ModrinthSearchResult
		if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil || page.Offset != 48 || page.Limit != 99 || page.TotalHits != 123 || len(page.Hits) != 1 {
			t.Fatalf("page %+v %v", page, err)
		}
		if page.Hits[0].FeaturedGallery != "https://cdn.modrinth.com/banner.webp" || len(page.Hits[0].Gallery) != 1 {
			t.Fatal("missing search artwork", page.Hits[0])
		}
	}
	for _, suffix := range []string{"project", "versions"} {
		for _, tc := range []struct {
			user   string
			status int
		}{{"", 401}, {"stranger", 403}, {"subuser", 200}, {"owner", 200}} {
			w := f.request("GET", base+suffix+"?slug=example", "", tc.user)
			if w.Code != tc.status {
				t.Fatalf("%s %s: %d %s", suffix, tc.user, w.Code, w.Body.String())
			}
			if tc.status != 200 {
				continue
			}
			var result struct {
				Versions              []downloader.ModrinthVersion `json:"versions"`
				Loader                string                       `json:"loader"`
				MinecraftVersion      string                       `json:"minecraft_version"`
				InstallableVersionIDs []string                     `json:"installable_version_ids"`
				LatestVersionStatus   string                       `json:"latest_version_status"`
				LatestVersionID       string                       `json:"latest_version_id"`
				Warnings              []downloader.ModrinthWarning `json:"warnings"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Loader != "paper" || result.MinecraftVersion != "1.21" {
				t.Fatal("lost server metadata", result)
			}
			if suffix == "versions" && (len(result.Versions) != 2 || result.Versions[0].ID != "new") {
				t.Fatal("wrong eligible versions", result)
			}
			if suffix == "project" && (len(result.Versions) != 3 || result.Versions[0].ID != "fabric" || strings.Join(result.InstallableVersionIDs, ",") != "new,old") {
				t.Fatal("wrong detail", result)
			}
			if suffix == "project" && (result.LatestVersionStatus != "available" || result.LatestVersionID != "new" || len(result.Warnings) != 0) {
				t.Fatal("wrong latest contract", result)
			}
			if strings.Contains(w.Body.String(), "private-secret") || strings.Contains(w.Body.String(), "payout") {
				t.Fatal("private fields exposed")
			}
		}
		before := calls
		for _, slug := range []string{"", "a/b", "a%2fb", "../bad", "a?x=y", "é", strings.Repeat("a", 129)} {
			if w := f.request("GET", base+suffix+"?slug="+url.QueryEscape(slug), "", "owner"); w.Code != 400 {
				t.Fatal(w.Code, w.Body.String())
			}
		}
		if calls != before {
			t.Fatal("invalid identifier reached upstream")
		}
		if serverActionPermission("GET", "plugins/"+suffix) != "files.view" || featureForRoute("GET", base+suffix) != "plugins" {
			t.Fatal("plugin policy mapping differs")
		}
	}
	for _, id := range []string{"fabric", "foreign"} {
		if w := f.request("POST", base+"install", `{"slug":"example","version_id":"`+id+`"}`, "owner"); w.Code != 400 {
			t.Fatal("invalid install", w.Code, w.Body.String())
		}
	}
	for _, body := range []string{`{"slug":"example"}`, `{"slug":"example","version_id":"old"}`} {
		w := f.request("POST", base+"install", body, "owner")
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	for _, name := range []string{"old.jar", "new.jar"} {
		data, err := os.ReadFile(filepath.Join(f.manager.GetWorkingDir("alpha"), "plugins", name))
		if err != nil || string(data) != "jar" {
			t.Fatalf("%s: %s %v", name, data, err)
		}
	}
	// No auto-start/restart is invoked: the installed server remains unstarted.
	if w := f.request("GET", "/api/servers/alpha", "", "owner"); w.Code != 200 || strings.Contains(w.Body.String(), `"status":"running"`) {
		t.Fatal(w.Code, w.Body.String())
	}
	fail = true
	for _, path := range []string{base + "project?slug=example", base + "versions?slug=example", "/api/plugins/search?offset=48&limit=99"} {
		w := f.request("GET", path, "", "owner")
		if w.Code != 502 || strings.Contains(w.Body.String(), "private-secret") || !strings.Contains(w.Body.String(), "retry") || !strings.Contains(w.Body.String(), `"code":"plugin_catalog_unavailable"`) {
			t.Fatal("unsafe upstream error", w.Code, w.Body.String())
		}
	}
	if err := f.db.SetFeaturePolicy("alpha", database.FeaturePolicy{"plugins": false}); err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"project", "versions"} {
		if w := f.request("GET", base+suffix+"?slug=example", "", "owner"); w.Code != 403 {
			t.Fatal("feature bypass", w.Code, w.Body.String())
		}
	}
}

func TestPluginProjectRequestContext(t *testing.T) {
	f := newReliabilityFixture(t)
	old := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = old })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	http.DefaultTransport = pluginCatalogTransport(func(r *http.Request) (*http.Response, error) {
		cancel()
		<-r.Context().Done()
		return nil, r.Context().Err()
	})
	req := httptest.NewRequest("GET", "/api/servers/alpha/plugins/project?slug=example", nil).WithContext(ctx)
	req.Header.Set("Authorization", "Bearer "+f.tokens["owner"])
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	if w.Code != 502 || !strings.Contains(w.Body.String(), "canceled") {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestPluginProjectWorkerRouting(t *testing.T) {
	master := newReliabilityFixture(t)
	worker := newReliabilityFixture(t)
	workerHTTP := httptest.NewServer(worker.router)
	defer workerHTTP.Close()
	if err := master.db.CreateNode(&database.ClusterNode{ID: "remote", Name: "remote", BaseURL: workerHTTP.URL, APIToken: worker.cfg.Security.NodeToken}); err != nil {
		t.Fatal(err)
	}
	old := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = old })
	workerURL, _ := url.Parse(workerHTTP.URL)
	http.DefaultTransport = pluginCatalogTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host == workerURL.Host {
			return old.RoundTrip(r)
		}
		if r.URL.Host != "api.modrinth.com" {
			t.Fatal("unexpected upstream", r.URL)
		}
		data := ""
		switch r.URL.Path {
		case "/v2/project/example":
			data = `{"id":"P1","slug":"example","team":"T1"}`
		case "/v2/project/P1/version", "/v2/team/T1/members":
			data = `[]`
		default:
			t.Fatal("unexpected path", r.URL)
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(data))}, nil
	})
	const path = "/api/nodes/remote/proxy/api/servers/alpha/plugins/project?slug=example"
	for _, tc := range []struct {
		user   string
		status int
	}{{"owner", 200}, {"subuser", 200}, {"stranger", 403}, {"", 401}} {
		w := master.request("GET", path, "", tc.user)
		if w.Code != tc.status {
			t.Fatalf("%s: %d %s", tc.user, w.Code, w.Body.String())
		}
		if w.Code == 200 && !strings.Contains(w.Body.String(), `"installable_version_ids":[]`) {
			t.Fatal(w.Body.String())
		}
	}
}

func TestPluginProjectPartialDataContract(t *testing.T) {
	f := newReliabilityFixture(t)
	old := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = old })
	http.DefaultTransport = pluginCatalogTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "api.modrinth.com" {
			t.Fatal(r.URL)
		}
		data, status := "private-secret", 503
		if r.URL.Path == "/v2/project/example" {
			data, status = `{"id":"P1","slug":"example","team":"T1","body":"# Available description"}`, 200
		}
		return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(data))}, nil
	})
	w := f.request("GET", "/api/servers/alpha/plugins/project?slug=example", "", "owner")
	var result downloader.ModrinthProjectDetail
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if w.Code != 200 || result.Project.Body != "# Available description" || result.LatestVersionStatus != "unknown" || result.LatestVersionID != "" || len(result.Warnings) != 3 || !result.VersionsTruncated {
		t.Fatalf("partial-data contract: %d %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "private-secret") || !strings.Contains(w.Body.String(), `"installable_version_ids":[]`) {
		t.Fatal("unsafe or ambiguous partial response", w.Body.String())
	}
	for _, warning := range result.Warnings {
		if warning.Code != "upstream_unavailable" || !strings.Contains(warning.Message, "503") {
			t.Fatal(warning)
		}
	}
}

func TestPluginSelectionPreflightDoesNotDependOnHistory(t *testing.T) {
	f := newReliabilityFixture(t)
	old := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = old })
	historyCalls := 0
	var version downloader.ModrinthVersion
	if err := json.Unmarshal([]byte(`{"id":"chosen","project_id":"P1","loaders":["paper"],"date_published":"2026-01-01T00:00:00Z","files":[{"filename":"voice.jar","url":"https://cdn.modrinth.com/voice.jar","size":1}]}`), &version); err != nil {
		t.Fatal(err)
	}
	http.DefaultTransport = pluginCatalogTransport(func(r *http.Request) (*http.Response, error) {
		data, status := []byte(`{}`), 200
		switch r.URL.Path {
		case "/v2/project/example":
			data = []byte(`{"id":"P1","slug":"example","team":"T1"}`)
		case "/v2/version/chosen":
			data, _ = json.Marshal(version)
		case "/v2/project/P1/version":
			if r.URL.Query().Get("offset") != "0" {
				historyCalls++
				status = 503
				break
			}
			page := make([]downloader.ModrinthVersion, 50)
			for i := range page {
				page[i] = version
				if i > 0 {
					page[i].ID = fmt.Sprintf("older%d", i)
					page[i].DatePublished = "2025-01-01T00:00:00Z"
				}
			}
			data, _ = json.Marshal(page)
		default:
			t.Fatalf("unexpected metadata request %s", r.URL)
		}
		return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(string(data)))}, nil
	})
	for _, query := range []string{"latest=true", "version_id=chosen"} {
		w := f.request("GET", "/api/servers/alpha/plugins/versions?slug=example&"+query, "", "owner")
		var result struct {
			Versions []downloader.ModrinthVersion `json:"versions"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if w.Code != 200 || len(result.Versions) != 1 || result.Versions[0].ID != "chosen" {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if historyCalls != 0 {
		t.Fatal("selection fetched unrelated history")
	}
	w := f.request("GET", "/api/servers/alpha/plugins/versions?slug=example", "", "owner")
	if w.Code != 502 || historyCalls != 1 {
		t.Fatal("fixture did not exercise failed later history", w.Code, historyCalls)
	}
}
