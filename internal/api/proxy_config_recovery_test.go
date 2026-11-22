package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProxyConfigRecoveryContract(t *testing.T) {
	f := newReliabilityFixture(t)
	setProxyFixture(t, f, "velocity")
	dir := f.manager.GetWorkingDir("alpha")
	raw := "config-version = '2.8'\nshow-max-players = -1\nonline-mode = true\nplayer-info-forwarding-mode = 'none'\n[servers]\nlobby = '127.0.0.1:25566'\ntry = ['lobby']\n[forced-hosts]\n'play.example.com' = 'lobby'\n[packet-limiter]\ninterval = 7\n"
	file := filepath.Join(dir, "velocity.toml")
	if err := os.WriteFile(file, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "forwarding.secret"), []byte("unchanged-secret"), 0600); err != nil {
		t.Fatal(err)
	}
	view := proxyViewFromResponse(t, f.request("GET", proxyURL, "", "owner"))
	if view.Ready || view.ValidationError == "" || view.Config.MaxPlayers != -1 || view.Config.ForwardingMode != "NONE" || len(view.Config.ForcedHosts[0].Servers) != 1 {
		t.Fatal(view)
	}
	// Existing invalid editable values do not defeat authentication confirmation.
	view.Config.MaxPlayers = 0
	view.Config.OnlineMode = false
	if w := f.request("PUT", proxyURL, proxySaveBody(t, view, false), "owner"); w.Code != 400 {
		t.Fatal(w.Code, w.Body.String())
	}
	if got, _ := os.ReadFile(file); string(got) != raw {
		t.Fatal("unconfirmed save wrote config")
	}
	saved := proxyViewFromResponse(t, f.request("PUT", proxyURL, proxySaveBody(t, view, true), "owner"))
	if !saved.Ready || saved.ValidationError != "" || saved.Config.MaxPlayers != 0 {
		t.Fatal(saved)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "forwarding.secret")); string(got) != "unchanged-secret" {
		t.Fatal("secret changed")
	}
	// Simulate Velocity rewriting a root setting; stale builder writes still fail.
	data, _ := os.ReadFile(file)
	if err := os.WriteFile(file, append([]byte("sample-players-in-ping = true\n"), data...), 0600); err != nil {
		t.Fatal(err)
	}
	w := f.request("PUT", proxyURL, proxySaveBody(t, saved, true), "owner")
	if w.Code != 409 || !strings.Contains(w.Body.String(), `"code":"revision_conflict"`) {
		t.Fatal(w.Code, w.Body.String())
	}
}

func TestProxyConfigFailureCategories(t *testing.T) {
	for _, tc := range []struct {
		name, content, code string
		directory           bool
		status              int
	}{
		{"syntax", "motd = 'private-unclosed", "unsupported_config", false, 400},
		{"wrong type", "online-mode = 'private-value'", "unsupported_config", false, 400},
		{"secret path", "forwarding-secret-file = '../private-secret'", "unsafe_secret", false, 400},
		{"oversize", strings.Repeat("#", (1<<20)+1), "file_unavailable", false, 409},
		{"not regular", "", "file_unavailable", true, 409},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newReliabilityFixture(t)
			setProxyFixture(t, f, "velocity")
			file := filepath.Join(f.manager.GetWorkingDir("alpha"), "velocity.toml")
			var err error
			if tc.directory {
				err = os.Mkdir(file, 0700)
			} else {
				err = os.WriteFile(file, []byte(tc.content), 0600)
			}
			if err != nil {
				t.Fatal(err)
			}
			w := f.request("GET", proxyURL, "", "owner")
			var body map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if w.Code != tc.status || body["code"] != tc.code || strings.Contains(w.Body.String(), "private-") {
				t.Fatal(w.Code, w.Body.String())
			}
		})
	}
}
