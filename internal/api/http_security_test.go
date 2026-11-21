package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSecurityCookieWritesRejectSiblingOrigin(t *testing.T) {
	f := newReliabilityFixture(t)
	for _, tc := range []struct {
		origin, site string
		bearer       bool
		want         int
	}{
		{"https://evil.example.com", "same-site", false, 403},
		{"https://panel.example.com:8443", "same-site", false, 403},
		{"null", "cross-site", false, 403},
		{"", "", false, 403},
		{"https://panel.example.com", "same-origin", false, 200},
		{"", "", true, 200},
	} {
		req := httptest.NewRequest("POST", "https://panel.example.com/api/servers/alpha/files/create", strings.NewReader(`{"path":"created","is_dir":true}`))
		if tc.bearer {
			req.Header.Set("Authorization", "Bearer "+f.tokens["owner"])
		} else {
			req.AddCookie(&http.Cookie{Name: "neopanel_token", Value: f.tokens["owner"]})
		}
		if tc.origin != "" {
			req.Header.Set("Origin", tc.origin)
		}
		if tc.site != "" {
			req.Header.Set("Sec-Fetch-Site", tc.site)
		}
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()
		f.router.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Fatal(tc, w.Code, w.Body.String())
		}
	}
}
func TestSecurityQueryTokensCannotAuthorizeWrites(t *testing.T) {
	f := newReliabilityFixture(t)
	r := httptest.NewRequest("POST", "/api/servers/alpha/files/create?token="+f.tokens["owner"], strings.NewReader(`{"path":"query-write","is_dir":true}`))
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal(w.Code, w.Body.String())
	}
	r = httptest.NewRequest("GET", "/api/servers/alpha?token="+f.tokens["owner"], nil)
	w = httptest.NewRecorder()
	f.router.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
}
func TestSecuritySessionCookieTransportAndLegacyCleanup(t *testing.T) {
	f := newReliabilityFixture(t)
	for _, tc := range []struct {
		name, url, peer, proto string
		trusted, public        bool
		secure                 bool
	}{
		{"tls", "https://panel.example.com/", "203.0.113.2:1", "", false, false, true},
		{"http", "http://panel.example.com/", "203.0.113.2:1", "", false, false, false},
		{"spoof", "http://panel.example.com/", "203.0.113.2:1", "https", false, false, false},
		{"trusted", "http://panel.example.com/", "127.0.0.1:1", "https", true, false, true},
		{"canonical", "http://panel.example.com/", "203.0.113.2:1", "", false, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f.cfg.Security.TrustedProxies = nil
			f.cfg.Server.PublicURL = ""
			if tc.trusted {
				f.cfg.Security.TrustedProxies = []string{"127.0.0.1/32"}
			}
			if tc.public {
				f.cfg.Server.PublicURL = "https://panel.example.com"
			}
			r := httptest.NewRequest("POST", tc.url, nil)
			r.RemoteAddr = tc.peer
			r.AddCookie(&http.Cookie{Name: "jwt", Value: "legacy-cookie"})
			if tc.proto != "" {
				r.Header.Set("X-Forwarded-Proto", tc.proto)
			}
			w := httptest.NewRecorder()
			f.router.auth.setSessionCookie(w, r, "test-value")
			cookies := w.Result().Cookies()
			found, cleared := false, false
			for _, c := range cookies {
				if c.Name == "neopanel_token" {
					found = true
					if c.Secure != tc.secure || !c.HttpOnly || c.SameSite != http.SameSiteLaxMode || c.Path != "/" {
						t.Fatal(c)
					}
				}
				if c.Name == "jwt" && c.MaxAge < 0 {
					cleared = true
				}
			}
			if !found || !cleared {
				t.Fatal(cookies)
			}
		})
	}
}
func TestSecurityControlBodyBoundsAndUploadExceptions(t *testing.T) {
	for _, tc := range []struct {
		path string
		want int64
	}{
		{"/api/servers/alpha", controlBodyLimit},
		{"/api/nodes/worker/proxy/api/servers/alpha", controlBodyLimit},
		{"/api/servers/alpha/files/upload", 200 << 20},
		{"/api/nodes/worker/proxy/api/servers/alpha/files/content", 6*fileEditorLimit + (64 << 10)},
		{"/api/account/avatar", avatarRequestLimit},
		{"/api/admin/users/1/avatar", avatarRequestLimit},
		{"/api/admin/branding/upload", (12 << 20) + 4096},
		{"/api/admin/mods/upload", (20 << 20) + 4096},
		{"/api/servers/alpha/databases/one/query", maxServerSQLRequestBytes},
		{"/api/evil/files/upload", controlBodyLimit},
	} {
		r := httptest.NewRequest("POST", tc.path, nil)
		r.Header.Set("Content-Type", "multipart/form-data")
		if got := apiRequestBodyLimit(r); got != tc.want {
			t.Fatal(tc.path, got, tc.want)
		}
	}
	f := newReliabilityFixture(t)
	payload := `{"name":"` + strings.Repeat("a", int(controlBodyLimit)) + `"}`
	for _, chunked := range []bool{false, true} {
		r := httptest.NewRequest("PUT", "/api/servers/alpha", strings.NewReader(payload))
		if chunked {
			r.ContentLength = -1
		}
		r.Header.Set("Authorization", "Bearer "+f.tokens["owner"])
		r.Header.Set("Content-Type", "multipart/form-data")
		w := httptest.NewRecorder()
		f.router.ServeHTTP(w, r)
		if w.Code != 413 && w.Code != 400 {
			t.Fatal("unbounded control body", chunked, w.Code)
		}
		s, _ := f.db.GetServer("alpha")
		if s.Name != "alpha" {
			t.Fatal("oversized body mutated server")
		}
	}
	r := httptest.NewRequest("POST", "/api/servers/alpha/files/content", strings.NewReader(strings.Repeat("x", int(controlBodyLimit)+1)))
	w := httptest.NewRecorder()
	if !limitAPIRequestBody(w, r) {
		t.Fatal("editor allowance lost")
	}
	if _, err := io.ReadAll(r.Body); err != nil {
		t.Fatal(err)
	}
}
func TestSecurityRateLimiterStateIsBounded(t *testing.T) {
	lim := NewIPRateLimiter(600, 15)
	lim.maxBuckets = 2
	request := func(ip, path string) bool {
		r := httptest.NewRequest("GET", path, nil)
		r.RemoteAddr = ip + ":1234"
		return lim.Allow(r)
	}
	if !request("192.0.2.1", "/api/test") || !request("192.0.2.2", "/api/test") || request("192.0.2.3", "/api/test") {
		t.Fatal("bucket cap not enforced")
	}
	if !request("192.0.2.1", "/api/test") || len(lim.buckets) != 2 {
		t.Fatal("active client lost quota state")
	}
	for _, b := range lim.buckets {
		b.lastRefill = time.Now().Add(-11 * time.Minute)
	}
	lim.lastCleanup = time.Now().Add(-6 * time.Minute)
	if !request("192.0.2.3", "/api/test") || len(lim.buckets) != 1 {
		t.Fatal("stale buckets not reclaimed")
	}
}
func TestSecurityBrowserHeaders(t *testing.T) {
	f := newReliabilityFixture(t)
	for _, scheme := range []string{"http", "https"} {
		r := httptest.NewRequest("GET", scheme+"://panel.example.com/api/auth/status", nil)
		w := httptest.NewRecorder()
		f.router.ServeHTTP(w, r)
		if w.Header().Get("Referrer-Policy") != "no-referrer" || !strings.Contains(w.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'") || w.Header().Get("Permissions-Policy") == "" {
			t.Fatal(w.Header())
		}
		if (w.Header().Get("Strict-Transport-Security") != "") != (scheme == "https") {
			t.Fatal(scheme, w.Header())
		}
	}
}

func TestSecurityCookieOriginBehindProxy(t *testing.T) {
	f := newReliabilityFixture(t)
	f.cfg.Server.PublicURL = "https://panel.example.com"
	f.cfg.Security.TrustedProxies = []string{"127.0.0.1/32"}
	for _, tc := range []struct {
		name, host, peer, origin, proto string
		want                            bool
	}{
		{"canonical preserved host", "panel.example.com", "192.0.2.1:1234", "https://panel.example.com", "", true},
		{"trusted rewritten host", "localhost:8080", "127.0.0.1:1234", "https://panel.example.com", "https", true},
		{"untrusted rewritten host", "localhost:8080", "192.0.2.1:1234", "https://panel.example.com", "https", false},
		{"trusted noncanonical https", "other.example.com", "127.0.0.1:1234", "https://other.example.com", "https", true},
		{"spoofed proto", "other.example.com", "192.0.2.1:1234", "https://other.example.com", "https", false},
		{"sibling origin", "panel.example.com", "127.0.0.1:1234", "https://tenant.example.com", "https", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "http://"+tc.host+"/api/account/ssh-keys", nil)
			r.RemoteAddr = tc.peer
			r.Header.Set("Origin", tc.origin)
			r.Header.Set("Sec-Fetch-Site", "same-origin")
			r.Header.Set("X-Forwarded-Proto", tc.proto)
			r.Header.Set("X-Forwarded-Host", "tenant.example.com")
			if got := f.router.auth.sameOriginRequest(r); got != tc.want {
				t.Fatalf("sameOriginRequest = %v, want %v", got, tc.want)
			}
		})
	}
}
