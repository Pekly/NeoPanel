package api

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAppearanceBootstrap(t *testing.T) {
	f := newReliabilityFixture(t)
	b := defaultPanelBranding()
	b.AccentColor = "#ef3456"
	b.BgDeep = "#18090d"
	b.DefaultTheme = "oled"
	b.CustomCSS = "sensitive-custom-css"
	b.LogoURL = "https://example.com/private-brand-url"
	raw, _ := json.Marshal(b)
	if err := f.db.SetPanelSetting(panelSettingBrandingKey, string(raw)); err != nil {
		t.Fatal(err)
	}
	w := f.request("GET", "/api/branding/appearance.js", "", "")
	if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), "application/javascript") || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal(w.Code, w.Header(), w.Body.String())
	}
	body := strings.TrimSuffix(strings.TrimPrefix(w.Body.String(), "window.NPInitialAppearance="), ";\n")
	var values map[string]any
	if err := json.Unmarshal([]byte(body), &values); err != nil {
		t.Fatal(err, body)
	}
	if values["accent_color"] != "#ef3456" || values["default_theme"] != "oled" || len(values) != 12 {
		t.Fatal(values)
	}
	if strings.Contains(w.Body.String(), "sensitive") || strings.Contains(w.Body.String(), "private-brand") {
		t.Fatal("appearance leaked non-palette fields")
	}
	b.AccentColor = "#123456"
	raw, _ = json.Marshal(b)
	if err := f.db.SetPanelSetting(panelSettingBrandingKey, string(raw)); err != nil {
		t.Fatal(err)
	}
	if w = f.request("GET", "/api/branding/appearance.js", "", ""); !strings.Contains(w.Body.String(), "#123456") {
		t.Fatal("stale bootstrap", w.Body.String())
	}
}

func TestNodeNameOrdinaryUserClusterMetadata(t *testing.T) {
	f := newReliabilityFixture(t)
	if w := f.request("PATCH", "/api/nodes/local/name", `{"name":"Home host"}`, "admin"); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	w := f.request("GET", "/api/cluster/servers", "", "owner")
	var cluster struct {
		Servers []struct {
			NodeName string `json:"node_name"`
		} `json:"servers"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &cluster); err != nil || w.Code != 200 || len(cluster.Servers) == 0 {
		t.Fatal(w.Code, w.Body.String(), err)
	}
	for _, server := range cluster.Servers {
		if server.NodeName != "Home host" {
			t.Fatal(server)
		}
	}
}
