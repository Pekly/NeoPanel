package api

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/downloader"
	"github.com/Pekly/NeoPanel/internal/rooted"
)

func installedTestJar(t *testing.T, descriptor, text string) []byte {
	t.Helper()
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	if descriptor != "" {
		w, err := z.Create(descriptor)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = w.Write([]byte(text)); err != nil {
			t.Fatal(err)
		}
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func installedTestHash(b []byte) string { h := sha512.Sum512(b); return hex.EncodeToString(h[:]) }
func installedTestWrite(t *testing.T, f *reliabilityFixture, name string, data []byte) {
	t.Helper()
	dir := filepath.Join(f.manager.GetWorkingDir("alpha"), "plugins")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
		t.Fatal(err)
	}
}
func installedTestVersion(t *testing.T, id, project, loader, date string, data []byte) downloader.ModrinthVersion {
	t.Helper()
	var v downloader.ModrinthVersion
	raw := fmt.Sprintf(`{"id":%q,"project_id":%q,"version_number":%q,"name":%q,"date_published":%q,"loaders":[%q],"game_versions":["different-MC"],"files":[{"filename":%q,"url":%q,"size":%d,"hashes":{"sha512":%q}}]}`, id, project, id, id, date, loader, id+".jar", "https://cdn.modrinth.com/"+id+".jar", len(data), installedTestHash(data))
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatal(err)
	}
	return v
}

type installedTestUpstream struct {
	versions       map[string]downloader.ModrinthVersion
	hashes         map[string]downloader.ModrinthVersion
	downloads      map[string][]byte
	fail           bool
	downloadFail   string
	searches       int
	downloadsCount int
	onDownload     func()
}

func installedMock(t *testing.T) *installedTestUpstream {
	t.Helper()
	m := &installedTestUpstream{versions: map[string]downloader.ModrinthVersion{}, hashes: map[string]downloader.ModrinthVersion{}, downloads: map[string][]byte{}}
	old := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = old })
	http.DefaultTransport = pluginCatalogTransport(func(r *http.Request) (*http.Response, error) {
		if _, ok := r.Context().Deadline(); !ok {
			t.Fatal("missing deadline")
		}
		if m.fail {
			return nil, errors.New("private-host-path")
		}
		var value any
		if r.URL.Host == "cdn.modrinth.com" {
			m.downloadsCount++
			if m.onDownload != nil {
				m.onDownload()
			}
			if r.URL.Path == m.downloadFail {
				return nil, errors.New("private-host-path")
			}
			data, ok := m.downloads[r.URL.Path]
			if !ok {
				t.Fatal("unexpected download", r.URL)
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(data)), Header: make(http.Header)}, nil
		}
		if r.URL.Host != "api.modrinth.com" {
			t.Fatal("unexpected host", r.URL)
		}
		switch {
		case r.URL.Path == "/v2/version_files":
			if r.Method != "POST" {
				t.Fatal(r.Method)
			}
			var req struct {
				Hashes    []string `json:"hashes"`
				Algorithm string   `json:"algorithm"`
			}
			if json.NewDecoder(r.Body).Decode(&req) != nil || req.Algorithm != "sha512" || len(req.Hashes) > 20 {
				t.Fatal("invalid hash lookup")
			}
			result := map[string]downloader.ModrinthVersion{}
			for _, hash := range req.Hashes {
				if v, ok := m.hashes[hash]; ok {
					result[hash] = v
				}
			}
			value = result
		case r.URL.Path == "/v2/search":
			m.searches++
			if r.URL.Query().Get("limit") != "5" || strings.Contains(r.URL.Query().Get("facets"), "versions:") {
				t.Fatal(r.URL)
			}
			value = map[string]any{"hits": []map[string]string{{"project_id": "P1", "slug": "example", "title": "Example", "description": "candidate only"}}}
		case strings.HasPrefix(r.URL.Path, "/v2/version/"):
			value = m.versions[strings.TrimPrefix(r.URL.Path, "/v2/version/")]
		case strings.HasSuffix(r.URL.Path, "/version"):
			if r.URL.Query().Get("offset") != "0" || r.URL.Query().Get("limit") != "50" || r.URL.Query().Has("game_versions") {
				t.Fatal("history fanout or MC gating", r.URL)
			}
			project := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v2/project/"), "/version")
			list := []downloader.ModrinthVersion{}
			for _, v := range m.versions {
				if v.ProjectID == project {
					list = append(list, v)
				}
			}
			value = list
		case strings.HasPrefix(r.URL.Path, "/v2/project/"):
			id := strings.TrimPrefix(r.URL.Path, "/v2/project/")
			value = map[string]string{"id": id, "slug": "example", "title": "Example"}
		default:
			t.Fatal("unexpected upstream path", r.URL)
		}
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(data)), Header: make(http.Header)}, nil
	})
	return m
}
func installedScan(t *testing.T, f *reliabilityFixture, query string) installedCatalog {
	t.Helper()
	w := f.request("GET", "/api/servers/alpha/plugins/catalog"+query, "", "owner")
	var page installedCatalog
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &page) != nil {
		t.Fatal(w.Code, w.Body.String())
	}
	if page.Items == nil {
		t.Fatal("null items")
	}
	return page
}
func installedUpdateBody(items ...pluginIdentityRequest) string {
	b, _ := json.Marshal(map[string]any{"confirm": true, "items": items})
	return string(b)
}

func TestPluginInstalledCatalogHashSuggestionUnknownAndError(t *testing.T) {
	f := newReliabilityFixture(t)
	m := installedMock(t)
	old := installedTestJar(t, "plugin.yml", "name: Example\nversion: 1\n")
	next := installedTestJar(t, "plugin.yml", "name: Example\nversion: 2\n")
	custom := installedTestJar(t, "paper-plugin.yml", "name: Example\nversion: custom\n")
	unknown := installedTestJar(t, "plugin.yml", "name: SomethingElse\nversion: private\n")
	v := installedTestVersion(t, "old", "P1", "paper", "2024-01-01T00:00:00Z", old)
	m.hashes[installedTestHash(old)] = v
	m.versions["next"] = installedTestVersion(t, "next", "P1", "paper", "2025-01-01T00:00:00Z", next)
	installedTestWrite(t, f, "a-renamed.jar.disabled", old)
	installedTestWrite(t, f, "b-upload.jar", custom)
	installedTestWrite(t, f, "c-unknown.jar", unknown)
	installedTestWrite(t, f, "config.yml", []byte("untouched"))
	page := installedScan(t, f, "?limit=100")
	if len(page.Items) != 3 || page.Limit != 20 || page.Total != 3 || page.Folder != "plugins" {
		t.Fatalf("%+v", page)
	}
	a, b, c := page.Items[0], page.Items[1], page.Items[2]
	if a.Match != "hash" || a.Enabled || !a.UpdateAvailable || a.InstalledVersionID != "old" || a.Latest.ID != "next" {
		t.Fatalf("hash: %+v", a)
	}
	if b.Match != "suggested" || b.UpdateAvailable || b.Project != nil || b.Latest != nil || b.InstalledVersionID != "" || b.InstalledVersion != "custom" || len(b.Candidates) != 1 {
		t.Fatalf("suggestion: %+v", b)
	}
	if c.Match != "unknown" || c.UpdateAvailable || c.Candidates == nil || m.searches != 2 {
		t.Fatalf("unknown: %+v", c)
	}
	page = installedScan(t, f, "?offset=1&limit=1")
	if len(page.Items) != 1 || page.Items[0].Filename != b.Filename || !page.HasMore {
		t.Fatal(page)
	}
	// Newer installed release may not be automatically downgraded.
	v.DatePublished = "2026-01-01T00:00:00Z"
	m.hashes[installedTestHash(old)] = v
	if installedScan(t, f, "?limit=1").Items[0].UpdateAvailable {
		t.Fatal("automatic downgrade")
	}
	m.fail = true
	for _, item := range installedScan(t, f, "").Items {
		if item.Match != "error" || item.UpdateAvailable || strings.Contains(item.Message, "private") {
			t.Fatal(item)
		}
	}
}

func TestPluginInstalledCatalogNonmutatingAndLease(t *testing.T) {
	f := newReliabilityFixture(t)
	page := installedScan(t, f, "")
	if page.Total != 0 {
		t.Fatal(page)
	}
	if _, err := os.Stat(filepath.Join(f.manager.GetWorkingDir("alpha"), "plugins")); !os.IsNotExist(err) {
		t.Fatal("scan created folder")
	}
	release, err := rooted.TryExclusiveMutation(f.manager.GetWorkingDir("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	w := f.request("GET", "/api/servers/alpha/plugins/catalog", "", "owner")
	release()
	if w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestPluginIdentifyHashBoundAndNoAutomaticSuggestionUpdate(t *testing.T) {
	f := newReliabilityFixture(t)
	m := installedMock(t)
	original := installedTestJar(t, "plugin.yml", "name: Example\nversion: custom")
	target := installedTestJar(t, "plugin.yml", "name: Example\nversion: next")
	installedTestWrite(t, f, "uploaded.jar", original)
	m.versions["next"] = installedTestVersion(t, "next", "P1", "paper", "2025-01-01T00:00:00Z", target)
	item := pluginIdentityRequest{"uploaded.jar", installedTestHash(original), "P1", "next"}
	if w := f.request("POST", "/api/servers/alpha/plugins/update", installedUpdateBody(item), "owner"); w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	identify := func(hash, project string) int {
		body := fmt.Sprintf(`{"filename":"uploaded.jar","sha512":%q,"project_id":%q,"confirm":true}`, hash, project)
		return f.request("POST", "/api/servers/alpha/plugins/identify", body, "owner").Code
	}
	if code := identify(strings.Repeat("a", 128), "P1"); code != 409 {
		t.Fatal(code)
	}
	if code := identify(item.SHA512, "P1"); code != 200 {
		t.Fatal(code)
	}
	p := installedScan(t, f, "").Items[0]
	if p.Match != "confirmed" || !p.UpdateAvailable || p.InstalledVersionID != "" || !strings.Contains(p.Message, "replaces") {
		t.Fatal(p)
	}
	installedTestWrite(t, f, "uploaded.jar", target)
	if p := installedScan(t, f, "").Items[0]; p.Match != "suggested" || p.UpdateAvailable {
		t.Fatal("confirmation survived hash change", p)
	}
	if got, err := f.db.ConfirmedPluginSource("beta", "uploaded.jar", item.SHA512); err != nil || got != "" {
		t.Fatal("cross-server confirmation", got, err)
	}
	if got, err := f.db.ConfirmedPluginSource("alpha", "renamed.jar", item.SHA512); err != nil || got != "" {
		t.Fatal("cross-file confirmation", got, err)
	}
}

func TestPluginBulkUpdateSelectedOnlyDisabledBackupsPartialAndPreflight(t *testing.T) {
	f := newReliabilityFixture(t)
	m := installedMock(t)
	old := installedTestJar(t, "plugin.yml", "name: Example\nversion: old")
	next := installedTestJar(t, "plugin.yml", "name: Example\nversion: next")
	oldV := installedTestVersion(t, "old", "P1", "paper", "2024-01-01T00:00:00Z", old)
	nextV := installedTestVersion(t, "next", "P1", "paper", "2025-01-01T00:00:00Z", next)
	m.hashes[installedTestHash(old)] = oldV
	m.versions["next"] = nextV
	m.downloads["/next.jar"] = next
	for _, name := range []string{"selected.jar.disabled", "not-selected.jar", "second.jar"} {
		installedTestWrite(t, f, name, old)
	}
	installedTestWrite(t, f, "config.yml", []byte("keep secrets"))
	item := pluginIdentityRequest{"selected.jar.disabled", installedTestHash(old), "P1", "next"}
	for _, bad := range []pluginIdentityRequest{{"second.jar", strings.Repeat("a", 128), "P1", "next"}, {"second.jar", item.SHA512, "wrong", "next"}, {"second.jar", item.SHA512, "P1", "fabric"}} {
		m.versions["fabric"] = installedTestVersion(t, "fabric", "P1", "fabric", "2025-01-01T00:00:00Z", next)
		w := f.request("POST", "/api/servers/alpha/plugins/update", installedUpdateBody(item, bad), "owner")
		if w.Code != 409 && w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
		if m.downloadsCount != 0 {
			t.Fatal("download before all preflight")
		}
	}
	release, err := rooted.TryMutation(f.manager.GetWorkingDir("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	if w := f.request("POST", "/api/servers/alpha/plugins/update", installedUpdateBody(item), "owner"); w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	release()
	m.onDownload = func() {
		if release, err := rooted.TryMutation(f.manager.GetWorkingDir("alpha")); err == nil {
			release()
			t.Fatal("managed mutation bypassed update")
		}
		if w := f.request("GET", "/api/servers/alpha/plugins/catalog", "", "owner"); w.Code != 409 {
			t.Fatal("scan overlapped update", w.Code)
		}
	}
	badV := nextV
	badV.ID = "broken"
	badV.Files = append(badV.Files[:0:0], badV.Files...)
	badV.Files[0].URL = "https://cdn.modrinth.com/broken.jar"
	m.versions["broken"] = badV
	m.downloadFail = "/broken.jar"
	second := pluginIdentityRequest{"second.jar", item.SHA512, "P1", "broken"}
	w := f.request("POST", "/api/servers/alpha/plugins/update", installedUpdateBody(item, second), "owner")
	var result pluginUpdateResponse
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &result) != nil {
		t.Fatal(w.Code, w.Body.String())
	}
	if !result.RestartRequired || result.Results[0].Status != "updated" || result.Results[1].Status != "failed" {
		t.Fatal(w.Body.String())
	}
	root := f.manager.GetWorkingDir("alpha")
	for name, want := range map[string][]byte{"selected.jar.disabled": next, "second.jar": old, "not-selected.jar": old, "config.yml": []byte("keep secrets")} {
		got, err := os.ReadFile(filepath.Join(root, "plugins", name))
		if err != nil || !bytes.Equal(got, want) {
			t.Fatal(name, err)
		}
	}
	backup := result.Results[0].BackupPath
	if !strings.HasPrefix(backup, "plugin-backups/") || filepath.IsAbs(backup) || strings.Contains(w.Body.String(), root) {
		t.Fatal("unsafe backup path", w.Body.String())
	}
	if got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(backup))); err != nil || !bytes.Equal(got, old) {
		t.Fatal("backup lost", err)
	}
	if _, err := os.Stat(filepath.Join(root, "plugins", "next.jar")); !os.IsNotExist(err) {
		t.Fatal("duplicate enabled release")
	}
	if release, err := rooted.TryExclusiveMutation(root); err != nil {
		t.Fatal("gate leaked", err)
	} else {
		release()
	}
	if w := f.request("POST", "/api/servers/alpha/plugins/update", installedUpdateBody(item), "owner"); w.Code != 409 {
		t.Fatal("stale retry", w.Code)
	}
}

func TestPluginPublicationRollback(t *testing.T) {
	for _, failRollback := range []bool{false, true} {
		t.Run(fmt.Sprint(failRollback), func(t *testing.T) {
			dir := t.TempDir()
			root, err := rooted.Open(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			if err := root.WriteFile("live", []byte("original"), 0600); err != nil {
				t.Fatal(err)
			}
			link := func(a, b string) error {
				if failRollback && a == "backup" {
					return errors.New("injected")
				}
				return root.Link(a, b)
			}
			recovery, err := publishPluginUpdate("live", "staged", "backup", link, root.Remove, func(string, string) error { return errors.New("injected") })
			if err == nil || recovery != failRollback {
				t.Fatal(recovery, err)
			}
			if b, e := root.ReadFile("backup"); e != nil || string(b) != "original" {
				t.Fatal("lost backup", e)
			}
			if !failRollback {
				if b, e := root.ReadFile("live"); e != nil || string(b) != "original" {
					t.Fatal("rollback failed", e)
				}
			}
		})
	}
}

func TestPluginDescriptorBoundsAndFallback(t *testing.T) {
	for _, tc := range []struct {
		name, body, want string
		bad              bool
	}{
		{"plugin.yml", "name: Example\nversion: 1.2", "Example", false},
		{"velocity-plugin.json", `{"id":"velocity-example","version":"1"}`, "velocity-example", false},
		{"fabric.mod.json", `{"name":"Fabric Example","version":"1"}`, "Fabric Example", false},
		{"META-INF/mods.toml", "[[mods]]\nmodId='example'\nversion='1'", "example", false},
		{"plugin.yml", "name: " + strings.Repeat("x", pluginDescriptorLimit), "", true},
		{"../plugin.yml", "name: escaped", "", false},
		{"plugin.yml", "name: &x [*x, *x]", "", false},
	} {
		t.Run(tc.name+tc.want, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "test.jar"), installedTestJar(t, tc.name, tc.body), 0600); err != nil {
				t.Fatal(err)
			}
			root, err := rooted.Open(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			name, _, err := pluginDescriptor(root, "test.jar")
			if (err != nil) != tc.bad || name != tc.want {
				t.Fatal(name, err)
			}
		})
	}
	if pluginFilenameQuery("Example-v1.2-build42.jar.disabled") != "Example" {
		t.Fatal("bad filename fallback")
	}
	dir := t.TempDir()
	root, err := rooted.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	f, err := root.OpenFile("big.jar", os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Truncate(pluginFileLimit + 1); err != nil {
		t.Fatal(err)
	}
	f.Close()
	budget := pluginBatchLimit
	if _, _, err := pluginDigest(context.Background(), root, "big.jar", &budget); err == nil {
		t.Fatal("oversize hashed")
	}
	if budget != pluginBatchLimit {
		t.Fatal("oversize consumed hash budget")
	}
}
