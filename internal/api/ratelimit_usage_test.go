package api

import (
	"net/http"
	"testing"
	"time"
)

func TestRateLimitAllowsNormalNavigationAndPolling(t *testing.T) {
	f := newReliabilityFixture(t)
	f.router.globalLimiter = NewRouter(f.cfg, f.db, f.manager, nil).globalLimiter
	// Several quick page loads and their background status checks used to drain
	// the 20-request burst. They must not consume the sensitive login allowance.
	for range 120 {
		for _, path := range []string{"/api/servers", "/api/auth/status"} {
			w := f.request(http.MethodGet, path, "", "owner")
			if w.Code != http.StatusOK {
				t.Fatalf("navigation %s: %d %s", path, w.Code, w.Body.String())
			}
		}
	}
	if len(f.router.authLimiter.authBuckets) != 0 {
		t.Fatal("normal reads consumed login allowance")
	}
	if w := f.request(http.MethodGet, "/api/servers", "", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated access: %d", w.Code)
	}
}

func TestRateLimitStillProtectsSensitiveRoutes(t *testing.T) {
	paths := []string{
		"/api/auth/login", "/api/auth/login/2fa",
		"/api/auth/password-reset/request", "/api/auth/password-reset/confirm",
		"/api/auth/passkeys/login/begin", "/api/auth/passkeys/login/finish",
		"/api/auth/invitations/inspect", "/api/auth/invitations/accept",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			f := newReliabilityFixture(t)
			f.router.globalLimiter = NewRouter(f.cfg, f.db, f.manager, nil).globalLimiter
			if f.router.authLimiter.authRateSec != 0.25 || f.router.authLimiter.authBurst != 5 {
				t.Fatal("sensitive allowance changed")
			}
			for range 5 {
				w := f.request(http.MethodPost, path, `{`, "")
				if w.Code == http.StatusTooManyRequests {
					t.Fatal("initial sensitive request denied")
				}
			}
			// Freeze refill so this check does not depend on the test machine's speed.
			for _, bucket := range f.router.authLimiter.authBuckets {
				bucket.lastRefill = time.Now().Add(time.Minute)
			}
			w := f.request(http.MethodPost, path, `{`, "")
			if w.Code != http.StatusTooManyRequests || w.Header().Get("Retry-After") == "" {
				t.Fatalf("sensitive route unprotected: %d %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestRateLimitGeneralQuotaStillEnforcedAndRefills(t *testing.T) {
	f := newReliabilityFixture(t)
	f.router.globalLimiter = NewRouter(f.cfg, f.db, f.manager, nil).globalLimiter
	lim := f.router.globalLimiter
	if lim.ratePerSec != 10 || lim.burstCapacity != 200 {
		t.Fatal("unexpected general allowance")
	}
	if w := f.request(http.MethodGet, "/api/servers", "", "owner"); w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	for _, bucket := range lim.buckets {
		bucket.tokens = 0
		bucket.lastRefill = time.Now().Add(time.Minute)
	}
	w := f.request(http.MethodGet, "/api/servers", "", "owner")
	if w.Code != http.StatusTooManyRequests || w.Header().Get("Retry-After") == "" {
		t.Fatal("general quota bypassed", w.Code)
	}
	for _, bucket := range lim.buckets {
		bucket.tokens = 0
		bucket.lastRefill = time.Now().Add(-6 * time.Second)
	}
	for range 60 {
		if w := f.request(http.MethodGet, "/api/servers", "", "owner"); w.Code != http.StatusOK {
			t.Fatal("refill too slow", w.Code)
		}
	}
}
