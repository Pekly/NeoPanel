package api

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestPluginMetadataOnlyIdentityWithoutUpdateOrSearchRequests(t *testing.T) {
	f := newReliabilityFixture(t)
	m := installedMock(t)
	known := installedTestJar(t, "plugin.yml", "name: EmbeddedName\nversion: 1\n")
	custom := installedTestJar(t, "plugin.yml", "name: CustomName\nversion: patched\n")
	unknown := installedTestJar(t, "plugin.yml", "name: PrivateTool\nversion: local\n")
	m.hashes[installedTestHash(known)] = installedTestVersion(t, "old", "P1", "paper", "2024-01-01T00:00:00Z", known)
	installedTestWrite(t, f, "a-renamed.jar", known)
	installedTestWrite(t, f, "b-copy.jar.disabled", known)
	installedTestWrite(t, f, "c-custom.jar", custom)
	installedTestWrite(t, f, "d-private.jar", unknown)
	if err := f.db.ConfirmPluginSource("alpha", "c-custom.jar", installedTestHash(custom), "P1"); err != nil {
		t.Fatal(err)
	}
	transport := http.DefaultTransport
	projects, hashes := 0, 0
	http.DefaultTransport = pluginCatalogTransport(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/v2/version_files":
			hashes++
			return transport.RoundTrip(r)
		case "/v2/project/P1":
			projects++
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"id":"P1","slug":"real-plugin","title":"Real Plugin","icon_url":"https://cdn.modrinth.com/icon.png"}`))}, nil
		default:
			t.Fatal("automatic display must not fetch releases, search names, or download files", r.URL)
			return nil, nil
		}
	})
	page := installedScan(t, f, "?metadata_only=true")
	if len(page.Items) != 4 || projects != 1 || hashes != 1 {
		t.Fatal(page, projects, hashes)
	}
	for i, item := range page.Items {
		if item.Latest != nil || item.UpdateAvailable || len(item.Candidates) != 0 || item.Candidates == nil {
			t.Fatal("display data authorized update or suggested source", item)
		}
		if i < 3 && (item.Name != "Real Plugin" || item.Project == nil || item.Project.IconURL != "https://cdn.modrinth.com/icon.png") {
			t.Fatal(item)
		}
	}
	if page.Items[0].Match != "hash" || page.Items[0].InstalledVersionID != "old" || page.Items[1].Enabled {
		t.Fatal(page.Items[:2])
	}
	if page.Items[2].Match != "confirmed" || page.Items[2].InstalledVersionID != "" || page.Items[2].InstalledVersion != "patched" {
		t.Fatal(page.Items[2])
	}
	if page.Items[3].Match != "unknown" || page.Items[3].Name != "PrivateTool" || page.Items[3].Project != nil {
		t.Fatal(page.Items[3])
	}
	if w := f.request("GET", "/api/servers/alpha/plugins/catalog?metadata_only=true", "", "stranger"); w.Code != 403 {
		t.Fatal(w.Code)
	}
	if w := f.request("GET", "/api/servers/alpha/plugins/catalog?metadata_only=invalid", "", "owner"); w.Code != 400 {
		t.Fatal(w.Code)
	}
}

func TestPluginMetadataOnlyProviderFailureKeepsSafeFilenameFallback(t *testing.T) {
	f := newReliabilityFixture(t)
	m := installedMock(t)
	installedTestWrite(t, f, "Example-v1.2.jar", installedTestJar(t, "plugin.yml", "name: Example"))
	m.fail = true
	page := installedScan(t, f, "?metadata_only=true")
	item := page.Items[0]
	if item.Match != "error" || item.Name != "Example" || item.Project != nil || item.Latest != nil || item.UpdateAvailable || strings.Contains(item.Message, "private") {
		t.Fatal(item)
	}
}
