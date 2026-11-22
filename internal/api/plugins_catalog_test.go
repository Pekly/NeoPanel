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
		{"", 24, 0, false}, {"limit=50&offset=10000", 50, 10000, false}, {"limit=999&offset=10001", 50, 10000, false},
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
				if r.URL.Query().Get("offset") != "48" || r.URL.Query().Get("limit") != "50" {
					t.Fatal(r.URL)
				}
				data = `{"hits":[{"project_id":"P1","slug":"example"}],"total_hits":123}`
			case "/v2/project/example":
				data = `{"id":"P1","slug":"example","team":"T1","body":"# Example","license":{"id":"MIT"}}`
			case "/v2/project/P1/version":
				data = `[
    {"id":"old","project_id":"P1","date_published":"2024-01-01T00:00:00Z","loaders":["paper"],"game_versions":["1.21"],"files":[{"filename":"old.jar","url":"https://cdn.modrinth.com/old.jar","size":3}]},
    {"id":"new","project_id":"P1","date_published":"2025-01-01T00:00:00Z","version_type":"release","changelog":"# New","dependencies":[],"loaders":["paper"],"game_versions":["1.22"],"files":[{"filename":"new.jar","url":"https://cdn.modrinth.com/new.jar","size":3}]},
    {"id":"fabric","project_id":"P1","date_published":"2026-01-01T00:00:00Z","loaders":["fabric"],"game_versions":["1.21"],"files":[{"filename":"mod.jar","url":"https://cdn.modrinth.com/mod.jar","size":3}]},
    {"id":"foreign","project_id":"P2","loaders":["paper"],"files":[{"filename":"foreign.jar","url":"https://cdn.modrinth.com/foreign.jar","size":3}]}
   ]`
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
		if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil || page.Offset != 48 || page.Limit != 50 || page.TotalHits != 123 || len(page.Hits) != 1 {
			t.Fatalf("page %+v %v", page, err)
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
		if w.Code != 502 || strings.Contains(w.Body.String(), "private-secret") || !strings.Contains(w.Body.String(), "retry") {
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
