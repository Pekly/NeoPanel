package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Every request must match a public-URL fixture; no test can reach the network.
func mockLatestMetadata(t *testing.T, fixtures map[string]string) *[]string {
	t.Helper()
	old := httpClient
	t.Cleanup(func() { httpClient = old })
	calls := []string{}
	httpClient = &http.Client{Timeout: time.Second, Transport: modrinthTransport(func(r *http.Request) (*http.Response, error) {
		url := r.URL.String()
		calls = append(calls, url)
		body, ok := fixtures[url]
		status := 200
		if !ok {
			if r.URL.Host != "mcjarfiles.com" {
				t.Errorf("unexpected request: %s", url)
			}
			status, body = 503, "unavailable"
		}
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	return &calls
}

func TestLatestOfficialFallbacks(t *testing.T) {
	const fill = "https://fill.papermc.io/v3/projects/"
	const fabric = "https://meta.fabricmc.net/v2/versions/"
	const jar = "https://fixtures.invalid/server.jar"
	builds := `[{"id":9,"downloads":{"server:default":{"url":"https://fixtures.invalid/old.jar"}}},{"id":11,"downloads":{"server:default":{"url":"` + jar + `"}}}]`
	for _, tc := range []struct {
		kind, version, resolved string
		fixtures                map[string]string
	}{
		{"paper", "latest", fill + "paper/versions/1.21.10/builds", map[string]string{fill + "paper": `{"versions":{"1.21":["1.21.9","1.21.10"],"1.20":["1.20.6"],"future":["26.1-pre1"]}}`}},
		{"folia", "latest", fill + "folia/versions/1.21.8/builds", map[string]string{fill + "folia": `{"versions":["1.21.8","1.20.6"]}`}},
		{"velocity", "latest", fill + "velocity/versions/3.5.0-SNAPSHOT/builds", map[string]string{fill + "velocity": `{"versions":["3.4.0-SNAPSHOT","3.5.0-SNAPSHOT"]}`}},
		{"velocity", "1.21.10", fill + "velocity/versions/3.5.0-SNAPSHOT/builds", map[string]string{fill + "velocity": `{"versions":{"3":["3.4.0-SNAPSHOT","3.5.0-SNAPSHOT"]}}`}},
		{"waterfall", "latest", fill + "waterfall/versions/1.20/builds", map[string]string{fill + "waterfall": `{"versions":["1.8","1.20","1.19"]}`}},
		{"purpur", "latest", "https://api.purpurmc.org/v2/purpur/1.21.8/latest/download", map[string]string{"https://api.purpurmc.org/v2/purpur": `{"versions":["1.21.8","1.20.6"]}`}},
		{"vanilla", "latest", "https://fixtures.invalid/release.json", map[string]string{"https://launchermeta.mojang.com/mc/game/version_manifest_v2.json": `{"latest":{"release":"1.21.10","snapshot":"26.1-snapshot"},"versions":[{"id":"26.1-snapshot","url":"https://fixtures.invalid/snapshot.json"},{"id":"1.21.10","url":"https://fixtures.invalid/release.json"}]}`}},
		{"fabric", "latest", fabric + "loader/1.21.10/0.18.1/1.0.1/server/jar", map[string]string{fabric + "game": `[{"version":"26.1","stable":false},{"version":"26.1-pre1","stable":true},{"version":"1.21.9","stable":true},{"version":"1.21.10","stable":true}]`, fabric + "loader": `[{"version":"0.19.0","stable":false},{"version":"0.17.0","stable":true},{"version":"0.18.1","stable":true}]`}},
		{"bungeecord", "latest", "https://ci.md-5.net/job/BungeeCord/lastSuccessfulBuild/artifact/bootstrap/target/BungeeCord.jar", map[string]string{}},
	} {
		t.Run(tc.kind+"/"+tc.version, func(t *testing.T) {
			if strings.HasSuffix(tc.resolved, "/builds") {
				tc.fixtures[tc.resolved] = builds
			} else if tc.kind == "vanilla" {
				tc.fixtures[tc.resolved] = `{"downloads":{"server":{"url":"` + jar + `"}}}`
			} else {
				tc.fixtures[tc.resolved] = "jar-fixture"
			}
			tc.fixtures[jar] = "jar-fixture"
			calls := mockLatestMetadata(t, tc.fixtures)
			dir := t.TempDir()
			if err := DownloadServerJar(tc.kind, tc.version, dir, "server.jar"); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(dir, "server.jar"))
			if err != nil || string(data) != "jar-fixture" {
				t.Fatalf("artifact: %q, %v", data, err)
			}
			found := false
			for _, call := range *calls {
				if call == tc.resolved {
					found = true
				}
			}
			if !found {
				t.Fatalf("did not resolve supported release: %v", *calls)
			}
		})
	}
}

func TestExplicitOfficialVersionsRemainExact(t *testing.T) {
	const fill = "https://fill.papermc.io/v3/projects/"
	const jar = "https://fixtures.invalid/exact.jar"
	for _, kind := range []string{"paper", "folia", "velocity", "waterfall", "purpur", "fabric", "vanilla"} {
		t.Run(kind, func(t *testing.T) {
			version := "1.16.5"
			if kind == "velocity" {
				version = "3.2.0-SNAPSHOT"
			}
			fixtures := map[string]string{jar: "exact"}
			switch kind {
			case "paper", "folia", "velocity", "waterfall":
				fixtures[fill+kind+"/versions/"+version+"/builds"] = `[{"id":1,"downloads":{"server:default":{"url":"` + jar + `"}}}]`
			case "purpur":
				fixtures["https://api.purpurmc.org/v2/purpur/"+version+"/latest/download"] = "exact"
			case "fabric":
				fixtures["https://meta.fabricmc.net/v2/versions/loader"] = `[{"version":"0.18.1","stable":true}]`
				fixtures["https://meta.fabricmc.net/v2/versions/loader/"+version+"/0.18.1/1.0.1/server/jar"] = "exact"
			case "vanilla":
				fixtures["https://launchermeta.mojang.com/mc/game/version_manifest_v2.json"] = `{"latest":{"release":"26.1"},"versions":[{"id":"` + version + `","url":"https://fixtures.invalid/detail"}]}`
				fixtures["https://fixtures.invalid/detail"] = `{"downloads":{"server":{"url":"` + jar + `"}}}`
			}
			mockLatestMetadata(t, fixtures)
			dir := t.TempDir()
			if err := DownloadServerJar(kind, version, dir, "server.jar"); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(dir, "server.jar"))
			if err != nil || string(data) != "exact" {
				t.Fatalf("artifact: %q, %v", data, err)
			}
		})
	}
}

func TestFillFailuresNeverInstallPinnedArtifacts(t *testing.T) {
	for _, kind := range []string{"paper", "folia", "velocity", "waterfall"} {
		for _, body := range []string{"[]", "not json", `[{"id":9,"downloads":{}}]`} {
			t.Run(kind+body, func(t *testing.T) {
				version := "1.18"
				if kind == "velocity" {
					version = "3.2.0-SNAPSHOT"
				}
				mockLatestMetadata(t, map[string]string{"https://fill.papermc.io/v3/projects/" + kind + "/versions/" + version + "/builds": body})
				dir := t.TempDir()
				if err := DownloadServerJar(kind, version, dir, "server.jar"); err == nil {
					t.Fatal("expected resolution error")
				}
				if _, err := os.Stat(filepath.Join(dir, "server.jar")); !os.IsNotExist(err) {
					t.Fatal("published an artifact", err)
				}
			})
		}
	}
}

func TestLatestMetadataFailures(t *testing.T) {
	for _, body := range []string{`{"versions":[]}`, `{"versions":["26.1-pre1"]}`, `{"versions":{}}`, `not json`, strings.Repeat(" ", (4<<20)+1)} {
		t.Run(fmt.Sprintf("size%d", len(body)), func(t *testing.T) {
			mockLatestMetadata(t, map[string]string{"https://fill.papermc.io/v3/projects/paper": body})
			if url, err := getPaperURL("latest"); err == nil || url != "" {
				t.Fatalf("got %q, %v", url, err)
			}
		})
	}
	for _, kind := range []string{"game", "loader"} {
		t.Run("fabric-no-stable-"+kind, func(t *testing.T) {
			mockLatestMetadata(t, map[string]string{"https://meta.fabricmc.net/v2/versions/" + kind: `[{"version":"26.1","stable":false}]`})
			if _, err := stableFabricVersion(kind); err == nil {
				t.Fatal("selected unstable fallback")
			}
		})
	}
}

func TestLatestPrimaryEndpointsAndFailures(t *testing.T) {
	for kind, route := range map[string]string{"paper": "servers/paper", "purpur": "servers/purpur", "leaf": "servers/leafmc", "leafmc": "servers/leafmc", "folia": "servers/folia", "fabric": "modded/fabric", "forge": "modded/forge", "neoforge": "modded/neoforge", "vanilla": "vanilla/release", "velocity": "proxies/velocity"} {
		t.Run(kind, func(t *testing.T) {
			calls := mockLatestMetadata(t, map[string]string{"https://mcjarfiles.com/api/get-latest-jar/" + route: "latest"})
			if err := DownloadServerJar(kind, "latest", t.TempDir(), "server.jar"); err != nil {
				t.Fatal(err)
			}
			if len(*calls) != 1 {
				t.Fatal(*calls)
			}
		})
	}
	for _, kind := range []string{"leaf", "leafmc", "forge", "neoforge"} {
		t.Run(kind+"-unavailable", func(t *testing.T) {
			calls := mockLatestMetadata(t, nil)
			dir := t.TempDir()
			err := DownloadServerJar(kind, "latest", dir, "server.jar")
			if err == nil || !strings.Contains(err.Error(), "503") {
				t.Fatalf("lost upstream failure: %v", err)
			}
			if len(*calls) != 1 {
				t.Fatal(*calls)
			}
			if _, err := os.Stat(filepath.Join(dir, "server.jar")); !os.IsNotExist(err) {
				t.Fatal("published an artifact", err)
			}
		})
	}
}

func TestOfficialTransportFailure(t *testing.T) {
	old := httpClient
	t.Cleanup(func() { httpClient = old })
	httpClient = &http.Client{Transport: modrinthTransport(func(r *http.Request) (*http.Response, error) { return nil, fmt.Errorf("upstream unavailable") })}
	for _, resolve := range []func(string) (string, error){getPaperURL, getPurpurURL, getVanillaURL, getFabricURL, getVelocityURL, getWaterfallURL} {
		if url, err := resolve("latest"); err == nil || url != "" {
			t.Fatalf("got %q, %v", url, err)
		}
	}
}

func TestOfficialHTTPFailures(t *testing.T) {
	old := httpClient
	t.Cleanup(func() { httpClient = old })
	httpClient = &http.Client{Transport: modrinthTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader(`{"versions":["1.20"]}`))}, nil
	})}
	for _, resolve := range []func(string) (string, error){getPaperURL, getPurpurURL, getVanillaURL, getFabricURL, getVelocityURL, getWaterfallURL} {
		if url, err := resolve("latest"); err == nil || url != "" || !strings.Contains(err.Error(), "503") {
			t.Fatalf("got %q, %v", url, err)
		}
	}
	for _, resolve := range []func(string) (string, error){getVelocityURL, getWaterfallURL} {
		if url, err := resolve("3.2.0-SNAPSHOT"); err == nil || url != "" {
			t.Fatalf("explicit failure installed a fallback: %q, %v", url, err)
		}
	}
}
