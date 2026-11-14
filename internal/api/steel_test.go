package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/downloader"
)

type steelAPITransport func(*http.Request) (*http.Response, error)

func (f steelAPITransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func mockSteelDownloads(t *testing.T) *bool {
	t.Helper()
	old := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = old })
	fail := false
	http.DefaultTransport = steelAPITransport(func(r *http.Request) (*http.Response, error) {
		code, data := 200, "binary"
		switch r.URL.Host {
		case "api.github.com":
			data = `{"tag_name":"v1+mc26.3","assets":[{"name":"steel-linux-musl","browser_download_url":"https://github.com/Steel-Foundation/SteelMC/releases/download/v1/steel-linux-musl","size":6},{"name":"steel-windows.exe","browser_download_url":"https://github.com/Steel-Foundation/SteelMC/releases/download/v1/steel-windows.exe","size":6},{"name":"steel-macos","browser_download_url":"https://github.com/Steel-Foundation/SteelMC/releases/download/v1/steel-macos","size":6}]}`
			if r.URL.RawQuery != "" {
				data = "[" + data + "]"
			}
		case "raw.githubusercontent.com":
			data = "[server]\nserver_port = 25565\nmotd = \"Steel\"\n"
		case "github.com":
		default:
			t.Fatalf("unexpected download URL: %s", r.URL)
		}
		if fail {
			code = 503
		}
		return &http.Response{StatusCode: code, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(data))}, nil
	})
	return &fail
}

func TestSteelCreateAndJavaOnlyGuards(t *testing.T) {
	if err := downloader.ValidateSteelPlatform(); err != nil {
		t.Skip(err)
	}
	f := newReliabilityFixture(t)
	mockSteelDownloads(t)
	w := f.request("POST", "/api/servers", `{"name":"Steel test","type":"steel","auto_download":true,"port":25581,"java_path":"missing-java","java_args":"-invalid"}`, "admin")
	if w.Code != http.StatusCreated {
		t.Fatal(w.Code, w.Body.String())
	}
	var created ServerWithStats
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	srv := created.Server
	if srv.Type != "steel" || srv.Version != "latest" || srv.JarFile != downloader.ServerFileName("steel") || srv.JavaArgs != "" || srv.JavaPath != "" || srv.MemoryMinMB != 0 {
		t.Fatalf("bad native record: %+v", srv)
	}
	dir := f.manager.GetWorkingDir(srv.ID)
	for _, file := range []string{"eula.txt", "server.properties"} {
		if _, err := os.Stat(filepath.Join(dir, file)); !os.IsNotExist(err) {
			t.Fatalf("created %s", file)
		}
	}
	config, _ := os.ReadFile(filepath.Join(dir, "config", "config.toml"))
	if !strings.Contains(string(config), "server_port = 25581") {
		t.Fatal(string(config))
	}
	for _, route := range []string{"GET properties", "POST properties", "GET configs", "POST configs", "GET plugins", "POST plugins/install", "GET plugins/versions"} {
		parts := strings.Split(route, " ")
		response := f.request(parts[0], "/api/servers/"+srv.ID+"/"+parts[1], "{}", "admin")
		if response.Code != 400 || !strings.Contains(response.Body.String(), "Steel") {
			t.Fatalf("%s: %d %s", route, response.Code, response.Body.String())
		}
	}
	if w := f.request("GET", "/api/steel/versions", "", ""); w.Code != 401 {
		t.Fatal(w.Code)
	}
	if w := f.request("GET", "/api/steel/versions", "", "owner"); w.Code != 200 || !strings.Contains(w.Body.String(), "v1+mc26.3") {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestSteelSwitchAndReinstallPreservesNativeData(t *testing.T) {
	if err := downloader.ValidateSteelPlatform(); err != nil {
		t.Skip(err)
	}
	f := newReliabilityFixture(t)
	fail := mockSteelDownloads(t)
	srv, _ := f.db.GetServer("alpha")
	srv.Port = 25582
	if err := f.db.UpdateServer(srv); err != nil {
		t.Fatal(err)
	}
	w := f.request("POST", "/api/servers/alpha/software", `{"software":"steel","version":"v1+mc26.3"}`, "owner")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	srv, _ = f.db.GetServer("alpha")
	if srv.Type != "steel" || srv.JarFile != downloader.ServerFileName("steel") || srv.JavaArgs != "" {
		t.Fatalf("%+v", srv)
	}
	dir := f.manager.GetWorkingDir("alpha")
	save := filepath.Join(dir, "saves", "world-data")
	if err := os.MkdirAll(filepath.Dir(save), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(save, []byte("precious-world"), 0644); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(dir, "config", "config.toml")
	config := "[server]\nserver_port = 25582\nmotd = \"custom\"\n"
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatal(err)
	}
	*fail = true
	w = f.request("POST", "/api/servers/alpha/reinstall", `{"software":"steel","version":"missing","preserve_world":false,"preserve_config":false}`, "owner")
	if w.Code != 502 {
		t.Fatal(w.Code, w.Body.String())
	}
	if got, _ := os.ReadFile(save); string(got) != "precious-world" {
		t.Fatal("failed download deleted save")
	}
	if got, _ := os.ReadFile(configPath); string(got) != config {
		t.Fatal("failed download deleted config")
	}
	*fail = false
	w = f.request("POST", "/api/servers/alpha/reinstall", `{"software":"steel","version":"latest","preserve_world":true,"preserve_config":true}`, "owner")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if got, _ := os.ReadFile(save); string(got) != "precious-world" {
		t.Fatal("reinstall deleted save")
	}
	if got, _ := os.ReadFile(configPath); string(got) != config {
		t.Fatal("reinstall changed config")
	}
}

func TestSteelClientProvisionDefaults(t *testing.T) {
	f := newReliabilityFixture(t)
	w := f.request("POST", "/api/client-api/v1/provision", `{"client_email":"owner@example.test","client_username":"owner","server_type":"steel","server_name":"Native provision"}`, "admin")
	if w.Code != http.StatusCreated {
		t.Fatal(w.Code, w.Body.String())
	}
	var response struct {
		ID string `json:"server_id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	srv, err := f.db.GetServer(response.ID)
	if err != nil {
		t.Fatal(err)
	}
	if srv.Type != "steel" || srv.Version != "latest" || srv.JarFile != downloader.ServerFileName("steel") || srv.JavaArgs != "" || srv.MemoryMinMB != 0 {
		t.Fatalf("bad Steel provision: %+v", srv)
	}
	for _, file := range []string{"server.properties", "eula.txt"} {
		if _, err := os.Stat(filepath.Join(f.manager.GetWorkingDir(srv.ID), file)); !os.IsNotExist(err) {
			t.Fatalf("provision wrote %s", file)
		}
	}
}
