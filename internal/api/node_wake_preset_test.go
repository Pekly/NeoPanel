package api

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestNodeWakePresetUsesSavedBrandingWithoutSaving(t *testing.T) {
	f := newReliabilityFixture(t)
	existing := database.DefaultNodeWakeConfig()
	existing.OfflineMOTD = "Custom"
	existing.Enabled = true
	if err := f.db.SetNodeWakeConfig("local", existing); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(PanelBrandingConfig{BrandName: "My Network", AccentColor: "#123abc"})
	if err := f.db.SetPanelSetting(panelSettingBrandingKey, string(raw)); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/api/admin/nodes/local/wake/preset", nil)
	req.SetPathValue("id", "local")
	req = req.WithContext(context.WithValue(req.Context(), contextUserKey, &JWTClaims{Role: "admin"}))
	w := httptest.NewRecorder()
	f.router.nodes.GetNodeWakePreset(w, req)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var result struct {
		Config database.NodeWakeConfig `json:"config"`
		Name   string                  `json:"brand_name"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Name != "My Network" || !strings.Contains(result.Config.OfflineMOTD, "&#123abc&lMy Network") || result.Config.Enabled {
		t.Fatal(result)
	}
	if err := database.ValidateNodeWakeConfig(result.Config); err != nil {
		t.Fatal(err)
	}
	if got, err := f.db.GetNodeWakeConfig("local"); err != nil || got != existing {
		t.Fatal("preset changed configuration", got, err)
	}
	for _, role := range []string{"user", ""} {
		r := httptest.NewRequest("GET", "/", nil)
		r.SetPathValue("id", "local")
		r = r.WithContext(context.WithValue(r.Context(), contextUserKey, &JWTClaims{Role: role}))
		w := httptest.NewRecorder()
		f.router.nodes.GetNodeWakePreset(w, r)
		if w.Code != 403 {
			t.Fatal("preset not admin-only", role, w.Code)
		}
	}
}
