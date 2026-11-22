package api

import (
	"encoding/json"
	"github.com/Pekly/NeoPanel/internal/database"
	"strings"
	"testing"
)

func TestFeaturePolicyDefaultsPermissionsAndPrecedence(t *testing.T) {
	f := newReliabilityFixture(t)
	for _, key := range database.FeatureKeys {
		if !f.db.FeatureEnabled("alpha", key) {
			t.Fatal("default disabled", key)
		}
	}
	for _, path := range []string{"/api/admin/features", "/api/servers/alpha/features", "/api/cluster/features"} {
		if w := f.request("PUT", path, `{"plugins":false}`, "owner"); w.Code != 403 {
			t.Fatal(path, w.Code, w.Body.String())
		}
	}
	if w := f.request("PUT", "/api/servers/alpha/features", `{"plugins":false}`, "admin"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if f.db.FeatureEnabled("alpha", "plugins") || !f.db.FeatureEnabled("beta", "plugins") {
		t.Fatal("server scope")
	}
	if w := f.request("PUT", "/api/admin/features", `{"plugins":false}`, "admin"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := f.request("PUT", "/api/servers/alpha/features", `{"plugins":true}`, "admin"); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if f.db.FeatureEnabled("alpha", "plugins") {
		t.Fatal("global disable bypassed")
	}
	for _, body := range []string{`{"nope":true}`, `{"plugins":null}`, `{"plugins":"false"}`, `{} {}`} {
		if w := f.request("PUT", "/api/admin/features", body, "admin"); w.Code != 400 {
			t.Fatal(body, w.Code)
		}
	}
	if w := f.request("GET", "/api/servers/alpha/features", "", "stranger"); w.Code != 403 {
		t.Fatal("membership", w.Code)
	}
	w := f.request("GET", "/api/servers/alpha/features", "", "owner")
	var view featureView
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &view) != nil || view.Effective["plugins"] {
		t.Fatal(w.Code, w.Body.String())
	}
	if err := f.db.SetPanelSetting("features:global", "bad"); err != nil {
		t.Fatal(err)
	}
	if f.db.FeatureEnabled("alpha", "plugins") {
		t.Fatal("malformed policy enabled")
	}
}

func TestDisabledFeatureRoutesAndRecovery(t *testing.T) {
	f := newReliabilityFixture(t)
	policy := database.FeaturePolicy{}
	for _, key := range database.FeatureKeys {
		policy[key] = false
	}
	if err := f.db.SetFeaturePolicy("", policy); err != nil {
		t.Fatal(err)
	}
	for _, route := range []string{"POST backups", "GET schedules", "GET sftp", "POST files/import-sftp", "GET plugins", "GET databases", "GET players", "GET proxy-config", "POST ports", "POST software", "GET subdomain", "POST power/countdown"} {
		parts := strings.SplitN(route, " ", 2)
		for _, user := range []string{"admin", "owner"} {
			w := f.request(parts[0], "/api/servers/alpha/"+parts[1], "{}", user)
			if w.Code != 403 || !strings.Contains(w.Body.String(), "disabled") {
				t.Errorf("%s %s: %d %s", user, route, w.Code, w.Body.String())
			}
		}
	}
	for _, route := range []string{"GET backups", "GET ports", "GET files", "GET features"} {
		parts := strings.SplitN(route, " ", 2)
		w := f.request(parts[0], "/api/servers/alpha/"+parts[1], "", "owner")
		if w.Code != 200 {
			t.Errorf("recovery %s: %d %s", route, w.Code, w.Body.String())
		}
	}
	for _, route := range []string{"backups/restore", "files/import-sftp/id", "world-conversion/id"} {
		method := "DELETE"
		if route == "backups/restore" {
			method = "POST"
		}
		if featureForRoute(method, "/api/servers/alpha/"+route) != "" {
			t.Fatal("recovery gated", route)
		}
	}
}
