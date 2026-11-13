package api

import (
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
)

func savePanelURL(f *reliabilityFixture, origin string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]string{"public_url": origin})
	return f.request("PUT", "/api/admin/panel-url", string(body), "admin")
}

func TestPanelURLSettings(t *testing.T) {
	f := newReliabilityFixture(t)
	for _, method := range []string{"GET", "PUT"} {
		for _, user := range []string{"", "owner", "subuser"} {
			want := 403
			if user == "" {
				want = 401
			}
			if w := f.request(method, "/api/admin/panel-url", `{"public_url":"https://attacker.example"}`, user); w.Code != want {
				t.Fatalf("%s %s: %d %s", method, user, w.Code, w.Body.String())
			}
		}
	}
	if w := f.request("GET", "/api/admin/panel-url", "", "admin"); w.Code != 200 || !strings.Contains(w.Body.String(), `"source":"unset"`) {
		t.Fatal(w.Code, w.Body.String())
	}
	f.cfg.Server.PublicURL = "https://config.example/"
	if w := f.request("GET", "/api/admin/panel-url", "", "admin"); !strings.Contains(w.Body.String(), `"source":"config"`) {
		t.Fatal(w.Body.String())
	}
	if w := savePanelURL(f, " https://PANEL.example:443/ "); w.Code != 200 || !strings.Contains(w.Body.String(), `"public_url":"https://panel.example"`) {
		t.Fatal(w.Code, w.Body.String())
	}
	if f.cfg.Server.PublicURL != "https://config.example/" {
		t.Fatal("startup configuration was mutated")
	}
	if w := f.request("GET", "/api/admin/panel-url", "", "admin"); w.Code != 200 || !strings.Contains(w.Body.String(), `"source":"settings"`) {
		t.Fatal(w.Code, w.Body.String())
	}
	// An independent connection sees the persisted override, not a handler cache.
	reopened, err := database.InitDB(filepath.Join(f.cfg.Paths.DataDir, "panel.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if raw, err := reopened.GetPanelSetting(database.PanelPublicURLKey); err != nil || raw != "https://panel.example" {
		t.Fatal(raw, err)
	}
	for _, raw := range []string{"", "http://panel.example", "https://panel.example/path", "https://panel.example?x=1", "https://panel.example#", "https://user@panel.example", "https://*.example", "https://panel.example:", "https://panel.example:99999", "https://192.0.2.1", "javascript:alert(1)", "https://panel.example\\evil"} {
		if w := savePanelURL(f, raw); w.Code != 400 {
			t.Errorf("accepted %q: %d %s", raw, w.Code, w.Body.String())
		}
	}
	for _, body := range []string{`{}`, `null`, `{"public_url":123}`, `{"public_url":"https://panel.example","unknown":true}`, `{"public_url":"https://panel.example"} {}`, strings.Repeat("x", 5000)} {
		if w := f.request("PUT", "/api/admin/panel-url", body, "admin"); w.Code != 400 {
			t.Errorf("accepted malformed payload: %d", w.Code)
		}
	}
	if raw, _ := f.db.GetPanelSetting(database.PanelPublicURLKey); raw != "https://panel.example" {
		t.Fatal("failed save changed URL", raw)
	}
}

func TestPanelURLConsumers(t *testing.T) {
	f := newReliabilityFixture(t)
	passkeys := NewPasskeyHandler(f.router.auth)
	f.cfg.Server.Mode = "behind_proxy"
	f.cfg.Server.PublicURL = "https://config.example"
	if err := f.db.SetPanelSetting(panelSettingDaemonKey, `{"fqdn":"game.example"}`); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest("GET", "https://untrusted.example/", nil)
	request.Header.Set("X-Forwarded-Host", "spoof.example")
	request.Header.Set("X-Forwarded-Proto", "http")
	if origin, err := f.router.invitations.origin(request); err != nil || origin != "https://config.example" {
		t.Fatal(origin, err)
	}
	if w := savePanelURL(f, "https://panel.example"); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if origin, err := f.router.invitations.origin(request); err != nil || origin != "https://panel.example" {
		t.Fatal(origin, err)
	}
	if origin, err := f.router.email.auth.panelLinkOrigin(request); err != nil || origin != "https://panel.example" {
		t.Fatal(origin, err)
	}
	if _, _, err := passkeys.relyingParty(request); err == nil {
		t.Fatal("accepted unrelated request host")
	}
	request = httptest.NewRequest("GET", "https://panel.example/", nil)
	if wa, origin, err := passkeys.relyingParty(request); err != nil || origin != "https://panel.example" || wa.Config.RPID != "panel.example" {
		t.Fatal(origin, err)
	}
	if w := savePanelURL(f, "http://localhost:8123"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if _, _, err := passkeys.relyingParty(httptest.NewRequest("GET", "http://localhost:8123/", nil)); err != nil {
		t.Fatal(err)
	}
}

func TestPanelURLProtectsPasskeysAndAllowsMigration(t *testing.T) {
	f := newReliabilityFixture(t)
	user := f.users["owner"]
	key, err := f.db.AddPasskey(user.ID, user.PasswordHash, "old.example", "Test", []byte("credential"), "{}")
	if err != nil {
		t.Fatal(err)
	}
	if w := savePanelURL(f, "https://old.example:8443"); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w := savePanelURL(f, "https://new.example"); w.Code != 400 || !strings.Contains(w.Body.String(), "Remove all") {
		t.Fatal(w.Code, w.Body.String())
	}
	if raw, _ := f.db.GetPanelSetting(database.PanelPublicURLKey); raw != "https://old.example:8443" {
		t.Fatal("blocked save changed URL", raw)
	}
	if err := f.db.DeletePasskey(user.ID, key.ID, user.PasswordHash); err != nil {
		t.Fatal(err)
	}
	if w := savePanelURL(f, "https://new.example"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if _, err := f.db.AddPasskey(user.ID, user.PasswordHash, "old.example", "Stale flow", []byte("old"), "{}"); err == nil {
		t.Fatal("old in-flight registration accepted")
	}
	if _, err := f.db.AddPasskey(user.ID, user.PasswordHash, "new.example", "New", []byte("new"), "{}"); err != nil {
		t.Fatal(err)
	}
}

func TestPanelURLConcurrentRegistration(t *testing.T) {
	f := newReliabilityFixture(t)
	user := f.users["owner"]
	// Either the old key wins and migration is blocked, or migration wins and the
	// old registration is blocked. Never commit both inconsistent operations.
	var wg sync.WaitGroup
	var saveErr, keyErr error
	wg.Add(2)
	go func() { defer wg.Done(); saveErr = f.db.SetPanelPublicURL("https://new.example", "new.example") }()
	go func() {
		defer wg.Done()
		_, keyErr = f.db.AddPasskey(user.ID, user.PasswordHash, "old.example", "Racing", []byte("race"), "{}")
	}()
	wg.Wait()
	if saveErr == nil && keyErr == nil {
		t.Fatal("hostname changed with incompatible credential")
	}
	if saveErr != nil && keyErr != nil {
		t.Fatal("neither operation succeeded", saveErr, keyErr)
	}
}

func TestPanelURLRepairsInvalidStartupConfig(t *testing.T) {
	f := newReliabilityFixture(t)
	f.cfg.Server.PublicURL = "https://legacy.example/wrong-path"
	if w := f.request("GET", "/api/admin/panel-url", "", "admin"); w.Code != 200 || !strings.Contains(w.Body.String(), "wrong-path") {
		t.Fatal("administrator cannot load invalid configuration for repair", w.Code, w.Body.String())
	}
	if _, err := f.router.auth.panelURL(); err == nil {
		t.Fatal("invalid configuration accepted by URL consumers")
	}
	if w := savePanelURL(f, "https://repaired.example"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if settings, err := f.router.auth.panelURL(); err != nil || settings.PublicURL != "https://repaired.example" {
		t.Fatal(settings, err)
	}
}
