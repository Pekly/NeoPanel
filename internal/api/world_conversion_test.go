package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
)

func TestWorldConversionHandlerPermissionsAndDisabledReads(t *testing.T) {
	f := newReliabilityFixture(t)
	h := newWorldConversionHandler(f.cfg, f.db, f.manager)
	invoke := func(fn http.HandlerFunc, user, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "/", strings.NewReader(body))
		r.SetPathValue("id", "alpha")
		r.SetPathValue("jobId", "missing")
		if user != "" {
			u := f.users[user]
			r = r.WithContext(context.WithValue(r.Context(), contextUserKey, &JWTClaims{UserID: u.ID, Role: u.Role}))
		}
		w := httptest.NewRecorder()
		fn(w, r)
		return w
	}
	for _, fn := range []http.HandlerFunc{h.Start, h.Cancel, h.Replace} {
		if w := invoke(fn, "subuser", `{}`); w.Code != 403 {
			t.Fatal("view-only mutation", w.Code, w.Body.String())
		}
		if w := invoke(fn, "stranger", `{}`); w.Code != 403 && w.Code != 404 {
			t.Fatal("stranger mutation", w.Code)
		}
	}
	if e := f.db.SetFeaturePolicy("", database.FeaturePolicy{"world_conversion": false}); e != nil {
		t.Fatal(e)
	}
	if w := invoke(h.Start, "owner", `{"input":"world","format":"JAVA_1_20_5","acknowledge_loss":true}`); w.Code != 403 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := invoke(h.Replace, "owner", `{}`); w.Code != 403 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := invoke(h.Status, "owner", ``); w.Code != 200 || !strings.Contains(w.Body.String(), `"enabled":false`) {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := invoke(h.Cancel, "owner", ``); w.Code != 404 {
		t.Fatal("disabled cancellation was blanket gated", w.Code, w.Body.String())
	}
}
func TestWorldConversionDecodeClosed(t *testing.T) {
	for _, body := range []string{`{"command":"java"}`, `{} {}`, strings.Repeat("x", 5000)} {
		r := httptest.NewRequest("POST", "/", strings.NewReader(body))
		w := httptest.NewRecorder()
		var dst struct {
			Input string `json:"input"`
		}
		if decodeWorldConversion(w, r, &dst) {
			t.Fatal("accepted invalid request", body[:2])
		}
	}
}

func TestWorldConversionRoutesPermissionsAndDisabledRecovery(t *testing.T) {
	f := newReliabilityFixture(t)
	if err := f.db.SetFeaturePolicy("", database.FeaturePolicy{"world_conversion": false}); err != nil {
		t.Fatal(err)
	}
	for _, user := range []string{"admin", "owner", "subuser"} {
		w := f.request("GET", "/api/servers/alpha/world-conversion", "", user)
		if w.Code != 200 || !strings.Contains(w.Body.String(), `"enabled":false`) {
			t.Fatalf("status %s: %d %s", user, w.Code, w.Body.String())
		}
		if user != "admin" && strings.Contains(w.Body.String(), "systemd") {
			t.Fatal("member received operator diagnostics")
		}
	}
	for _, user := range []string{"admin", "owner"} {
		for _, path := range []string{"world-conversion", "world-conversion/missing/replace"} {
			w := f.request("POST", "/api/servers/alpha/"+path, `{}`, user)
			if w.Code != 403 || !strings.Contains(w.Body.String(), "disabled") {
				t.Fatalf("disabled %s: %d %s", path, w.Code, w.Body.String())
			}
		}
		if w := f.request("DELETE", "/api/servers/alpha/world-conversion/missing", "", user); w.Code != 404 {
			t.Fatal("cancel gated", w.Code, w.Body.String())
		}
	}
	for _, tc := range []struct{ method, suffix, want string }{
		{"GET", "world-conversion", "files.view"},
		{"POST", "world-conversion", "files.view+files.edit"},
		{"DELETE", "world-conversion/job", "files.view+files.edit"},
		{"POST", "world-conversion/job/replace", "files.view+files.edit+backup.manage"},
	} {
		if got := serverActionPermission(tc.method, tc.suffix); got != tc.want {
			t.Fatalf("%s %s: %s", tc.method, tc.suffix, got)
		}
	}
}
