package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/proxyconfig"
	"github.com/pelletier/go-toml/v2"
)

const proxyURL = "/api/servers/alpha/proxy-config"

func setProxyFixture(t *testing.T, f *reliabilityFixture, kind string) *database.Server {
	t.Helper()
	srv, err := f.db.GetServer("alpha")
	if err != nil {
		t.Fatal(err)
	}
	srv.Type, srv.Port = kind, 25588
	if err := f.db.UpdateServer(srv); err != nil {
		t.Fatal(err)
	}
	return srv
}

func proxyViewFromResponse(t *testing.T, w *httptest.ResponseRecorder) proxyConfigView {
	t.Helper()
	if w.Code != 200 && w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var view proxyConfigView
	if err := json.Unmarshal(w.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	return view
}

func usableProxyConfig(config proxyconfig.Settings) proxyconfig.Settings {
	config.Backends = []proxyconfig.Backend{{Name: "lobby", Address: "127.0.0.1:25566", MOTD: "Lobby"}, {Name: "survival", Address: "play.example.com:25567", MOTD: "Survival"}}
	config.Try = []string{"lobby", "survival"}
	config.ForcedHosts = []proxyconfig.ForcedHost{{Host: "survival.example.com", Servers: []string{"survival"}}}
	return config
}

func proxySaveBody(t *testing.T, view proxyConfigView, confirm bool) string {
	t.Helper()
	data, err := json.Marshal(map[string]any{"revision": view.Revision, "config": view.Config, "confirm_offline": confirm})
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func mockProxyDownloads(t *testing.T) {
	t.Helper()
	old := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = old })
	http.DefaultTransport = steelAPITransport(func(r *http.Request) (*http.Response, error) {
		data := "jar-fixture"
		switch r.URL.Host {
		case "mcjarfiles.com", "fill-data.papermc.io", "ci.md-5.net":
		case "fill.papermc.io":
			if strings.HasSuffix(r.URL.Path, "/builds") {
				data = `[{"id":578,"downloads":{"server:default":{"url":"https://fill-data.papermc.io/fixture.jar"}}}]`
			} else {
				data = `{"versions":["1.19","1.20"]}`
			}
		default:
			t.Fatalf("unexpected external request: %s", r.URL)
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(data))}, nil
	})
}

func TestProxyCreateUsesRealConfigAndNoWorldDefaults(t *testing.T) {
	for _, kind := range []string{"velocity", "bungeecord", "waterfall"} {
		t.Run(kind, func(t *testing.T) {
			f := newReliabilityFixture(t)
			mockProxyDownloads(t)
			body, _ := json.Marshal(map[string]any{"name": "Proxy", "type": kind, "port": 25588, "auto_download": true})
			w := f.request("POST", "/api/servers", string(body), "admin")
			if w.Code != 201 {
				t.Fatal(w.Code, w.Body.String())
			}
			var result ServerWithStats
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Version != "latest" || result.Port != 25588 {
				t.Fatal(result.Server)
			}
			dir := f.manager.GetWorkingDir(result.ID)
			for _, name := range []string{"server.properties", "eula.txt"} {
				if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
					t.Fatal("world config created", name, err)
				}
			}
			if data, err := os.ReadFile(filepath.Join(dir, result.JarFile)); err != nil || string(data) != "jar-fixture" {
				t.Fatal(string(data), err)
			}
			view, err := proxyconfig.Read(dir, kind, 25588)
			if err != nil || !view.Exists {
				t.Fatal(view, err)
			}
			if !view.Config.OnlineMode || len(view.Config.Backends) != 0 {
				t.Fatal("unsafe/invented defaults", view)
			}
		})
	}
}

func TestProxyConfigInitializeAndBuilderRoundTrip(t *testing.T) {
	for _, kind := range []string{"velocity", "bungeecord", "waterfall"} {
		t.Run(kind, func(t *testing.T) {
			f := newReliabilityFixture(t)
			setProxyFixture(t, f, kind)
			view := proxyViewFromResponse(t, f.request("GET", proxyURL, "", "owner"))
			if view.Exists || view.Ready || !view.CanEdit || view.Port != 25588 {
				t.Fatal(view)
			}
			if files, _ := os.ReadDir(f.manager.GetWorkingDir("alpha")); len(files) != 0 {
				t.Fatal("GET created files")
			}
			view = proxyViewFromResponse(t, f.request("POST", proxyURL+"/initialize", "{}", "owner"))
			before, err := os.ReadFile(filepath.Join(f.manager.GetWorkingDir("alpha"), view.File))
			if err != nil {
				t.Fatal(err)
			}
			view = proxyViewFromResponse(t, f.request("POST", proxyURL+"/initialize", "{}", "owner"))
			after, _ := os.ReadFile(filepath.Join(f.manager.GetWorkingDir("alpha"), view.File))
			if string(before) != string(after) {
				t.Fatal("initialize replaced existing config")
			}
			view.Config = usableProxyConfig(view.Config)
			view.Config.MOTD = "My network"
			saved := proxyViewFromResponse(t, f.request("PUT", proxyURL, proxySaveBody(t, view, false), "owner"))
			if !saved.Ready || saved.Running || saved.Revision == view.Revision || saved.Config.MOTD != "My network" || len(saved.Config.Backends) != 2 {
				t.Fatal(saved)
			}
			if w := f.request("PUT", proxyURL, proxySaveBody(t, view, false), "owner"); w.Code != 409 {
				t.Fatal("stale save accepted", w.Code, w.Body.String())
			}
			for _, route := range []string{"properties", "configs"} {
				for _, method := range []string{"GET", "POST"} {
					w := f.request(method, "/api/servers/alpha/"+route, `{"motd":"wrong file"}`, "owner")
					if w.Code != 400 {
						t.Fatal(route, method, w.Code, w.Body.String())
					}
				}
			}
			if _, err := os.Stat(filepath.Join(f.manager.GetWorkingDir("alpha"), "server.properties")); !os.IsNotExist(err) {
				t.Fatal("proxy wrote world properties", err)
			}
			if got := readServerMOTD(f.manager.GetWorkingDir("alpha"), kind); got != "My network" {
				t.Fatal(got)
			}
		})
	}
}

func TestProxyPermissionsAndExplicitSecretRead(t *testing.T) {
	f := newReliabilityFixture(t)
	setProxyFixture(t, f, "velocity")
	for _, route := range []struct{ method, path string }{{"GET", proxyURL}, {"POST", proxyURL + "/initialize"}, {"PUT", proxyURL}, {"POST", proxyURL + "/secret"}} {
		for _, user := range []string{"", "stranger", "subuser"} {
			want := 403
			if user == "" {
				want = 401
			}
			w := f.request(route.method, route.path, "{}", user)
			if w.Code != want {
				t.Fatal(user, route, w.Code, w.Body.String())
			}
		}
	}
	view := proxyViewFromResponse(t, f.request("POST", proxyURL+"/initialize", "{}", "owner"))
	secret, err := proxyconfig.ReadSecret(f.manager.GetWorkingDir("alpha"), "velocity")
	if err != nil {
		t.Fatal(err)
	}
	if w := f.request("GET", proxyURL, "", "owner"); strings.Contains(w.Body.String(), secret) {
		t.Fatal("ordinary settings exposed secret")
	}
	subs, _ := f.db.ListServerSubusers("alpha")
	for _, perms := range [][]string{{"settings.view"}, {"settings.view", "settings.edit"}, {"settings.view", "settings.edit", "files.view"}} {
		if err := f.db.UpdateServerSubuser("alpha", subs[0].ID, perms); err != nil {
			t.Fatal(err)
		}
		v := proxyViewFromResponse(t, f.request("GET", proxyURL, "", "subuser"))
		if v.CanViewSecret != (len(perms) == 3) || v.CanEdit != (len(perms) > 1) {
			t.Fatal(v)
		}
		w := f.request("POST", proxyURL+"/secret", "{}", "subuser")
		want := 403
		if len(perms) == 3 {
			want = 200
		}
		if w.Code != want {
			t.Fatal(perms, w.Code, w.Body.String())
		}
		if want == 200 && (!strings.Contains(w.Body.String(), secret) || w.Header().Get("Cache-Control") != "private, no-store") {
			t.Fatal(w.Body.String(), w.Header())
		}
	}
	view.Config = usableProxyConfig(view.Config)
	view.Config.OnlineMode = false
	if w := f.request("PUT", proxyURL, proxySaveBody(t, view, false), "owner"); w.Code != 400 {
		t.Fatal("offline mode needs confirmation", w.Code)
	}
	proxyViewFromResponse(t, f.request("PUT", proxyURL, proxySaveBody(t, view, true), "owner"))
}

func TestProxyConfigRejectsPortOverrideAndBusyServer(t *testing.T) {
	f := newReliabilityFixture(t)
	setProxyFixture(t, f, "velocity")
	view := proxyViewFromResponse(t, f.request("POST", proxyURL+"/initialize", "{}", "owner"))
	view.Config = usableProxyConfig(view.Config)
	body := proxySaveBody(t, view, false)
	for _, invalid := range []string{"null", body + "{}", strings.Replace(body, `"motd":`, `"bind":"0.0.0.0:1","motd":`, 1), strings.Repeat(" ", 256<<10+1)} {
		w := f.request("PUT", proxyURL, invalid, "owner")
		if w.Code != 400 && w.Code != 413 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	release, err := f.manager.BeginServerOperation("alpha")
	if err != nil {
		t.Fatal(err)
	}
	w := f.request("PUT", proxyURL, body, "owner")
	release()
	if w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	before := proxyViewFromResponse(t, f.request("GET", proxyURL, "", "owner"))
	if before.Revision != view.Revision {
		t.Fatal("rejected request wrote config")
	}
}

func TestProxyConfigWorkerRouteAndPolicyPort(t *testing.T) {
	master, worker := newReliabilityFixture(t), newReliabilityFixture(t)
	setProxyFixture(t, worker, "velocity")
	remote := httptest.NewServer(worker.router)
	defer remote.Close()
	if err := master.db.CreateNode(&database.ClusterNode{ID: "remote", Name: "worker", BaseURL: remote.URL, APIToken: worker.cfg.Security.NodeToken, AllowUserPorts: false}); err != nil {
		t.Fatal(err)
	}
	url := "/api/nodes/remote/proxy" + proxyURL
	view := proxyViewFromResponse(t, master.request("POST", url+"/initialize", "{}", "owner"))
	view.Config = usableProxyConfig(view.Config)
	saved := proxyViewFromResponse(t, master.request("PUT", url, proxySaveBody(t, view, false), "owner"))
	if saved.Port != 25588 || !saved.Ready {
		t.Fatal(saved)
	}
	if files, _ := os.ReadDir(master.manager.GetWorkingDir("alpha")); len(files) != 0 {
		t.Fatal("proxy config written on main panel")
	}
	if w := master.request("PUT", "/api/nodes/remote/proxy/api/servers/alpha", `{"port":25599}`, "owner"); w.Code != 403 {
		t.Fatal("node port restriction bypassed", w.Code, w.Body.String())
	}
}

func TestProxyReinstallPreservesCustomSecretAndConfig(t *testing.T) {
	f := newReliabilityFixture(t)
	setProxyFixture(t, f, "velocity")
	mockProxyDownloads(t)
	view := proxyViewFromResponse(t, f.request("POST", proxyURL+"/initialize", "{}", "owner"))
	view.Config = usableProxyConfig(view.Config)
	view = proxyViewFromResponse(t, f.request("PUT", proxyURL, proxySaveBody(t, view, false), "owner"))
	dir := f.manager.GetWorkingDir("alpha")
	file := filepath.Join(dir, "velocity.toml")
	data, _ := os.ReadFile(file)
	var document map[string]any
	if err := toml.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	document["forwarding-secret-file"] = "keys/network.key"
	document["custom-plugin-setting"] = "keep-me"
	data, err := toml.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "keys"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "keys", "network.key"), []byte("keep-this-secret"), 0600); err != nil {
		t.Fatal(err)
	}
	w := f.request("POST", "/api/servers/alpha/reinstall", `{"software":"velocity","version":"latest","preserve_config":true}`, "owner")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if secret, err := proxyconfig.ReadSecret(dir, "velocity"); err != nil || secret != "keep-this-secret" {
		t.Fatal(secret, err)
	}
	data, _ = os.ReadFile(file)
	if !strings.Contains(string(data), "keep-me") {
		t.Fatal("custom config lost", string(data))
	}
	if _, err := os.Stat(filepath.Join(dir, "server.properties")); !os.IsNotExist(err) {
		t.Fatal("world config created")
	}
}
