package api

import (
	"net/http/httptest"
	"testing"
)

func TestInvitationFlowUsesAuthenticationRateBucket(t *testing.T) {
	limiter := NewIPRateLimiter(1, 15)
	paths := []string{"/api/auth/invitations/inspect", "/api/auth/invitations/accept", "/api/auth/login", "/api/auth/invitations/inspect", "/api/auth/invitations/accept"}
	for _, path := range paths {
		req := httptest.NewRequest("POST", path, nil)
		req.RemoteAddr = "198.51.100.42:12345"
		if !limiter.Allow(req) {
			t.Fatalf("normal invitation/authentication flow rate-limited at %s", path)
		}
	}
	req := httptest.NewRequest("POST", "/api/auth/invitations/accept", nil)
	req.RemoteAddr = "198.51.100.42:12345"
	if limiter.Allow(req) {
		t.Fatal("invitation brute-force burst was not limited")
	}
}
