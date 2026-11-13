package api

import (
	"strings"
	"testing"
)

func TestProvisioningExamplesNeverExposePrivilegedCredentials(t *testing.T) {
	f := newReliabilityFixture(t)
	for _, user := range []string{"", "owner", "subuser"} {
		want := 403
		if user == "" {
			want = 401
		}
		if w := f.request("GET", "/api/client-api/v1/code-samples", "", user); w.Code != want {
			t.Fatal(user, w.Code, w.Body.String())
		}
		if w := f.request("POST", "/api/client-api/v1/provision", `{}`, user); w.Code != 401 {
			t.Fatal("ordinary session can provision", user, w.Code, w.Body.String())
		}
	}
	w := f.request("GET", "/api/client-api/v1/code-samples", "", "admin")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "YOUR_ADMIN_API_KEY") {
		t.Fatal(w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), f.cfg.Security.NodeToken) || strings.Contains(w.Body.String(), f.cfg.Security.JWTSecret) {
		t.Fatal("code examples leaked privileged credentials")
	}
}
