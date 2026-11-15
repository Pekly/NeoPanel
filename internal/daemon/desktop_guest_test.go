package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Pekly/NeoPanel/internal/api"
	"github.com/Pekly/NeoPanel/internal/database"
	"net/http"
	"strings"
	"testing"
)

func TestDesktopGuestSessionIsNativeOnly(t *testing.T) {
	cfg := testConfig(t)
	svc, err := StartDesktop(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Shutdown(context.Background())
	token := svc.TakeDesktopSession()
	if token == "" || svc.TakeDesktopSession() != "" {
		t.Fatal("session must transfer exactly once")
	}
	h := api.NewAuthHandler(svc.db, cfg)
	claims, err := h.ValidateToken(token)
	if err != nil || claims.Username != "Guest" || claims.Scope != "desktop_guest" {
		t.Fatalf("claims: %v %v", claims, err)
	}
	for _, authenticated := range []bool{false, true} {
		req, _ := http.NewRequest("GET", svc.URL()+"/api/servers", nil)
		if authenticated {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		response, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		want := http.StatusUnauthorized
		if authenticated {
			want = http.StatusOK
		}
		if response.StatusCode != want {
			t.Fatalf("authenticated=%v status=%d", authenticated, response.StatusCode)
		}
	}
	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/api/admin/users/%d", svc.URL(), claims.UserID), strings.NewReader(`{"username":"Guest","email":"owner@example.test","role":"admin","password":"private-test-password"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("set password: %d", response.StatusCode)
	}
	response, err = http.Post(svc.URL()+"/api/auth/login", "application/json", strings.NewReader(`{"username":"Guest","password":"private-test-password"}`))
	if err != nil {
		t.Fatal(err)
	}
	var login struct {
		Token string `json:"token"`
	}
	err = json.NewDecoder(response.Body).Decode(&login)
	response.Body.Close()
	if err != nil || response.StatusCode != http.StatusOK || login.Token == "" {
		t.Fatalf("password login: %d %v", response.StatusCode, err)
	}
	passwordClaims, err := h.ValidateToken(login.Token)
	if err != nil || passwordClaims.Scope != "" {
		t.Fatalf("password login retained Guest scope: %v", err)
	}
	if _, err := h.ValidateToken(token); err == nil {
		t.Fatal("password change did not revoke token")
	}
	if err := svc.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	svc2, err := StartDesktop(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer svc2.Shutdown(context.Background())
	if svc2.TakeDesktopSession() != "" {
		t.Fatal("password-protected account was bypassed")
	}
}
func TestDesktopDoesNotBypassExistingAndCLIStillNeedsSetup(t *testing.T) {
	cfg := testConfig(t)
	svc, err := Start(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if svc.TakeDesktopSession() != "" {
		t.Fatal("CLI received guest")
	}
	count, _ := svc.db.GetUserCount()
	if count != 0 {
		t.Fatal("CLI created guest")
	}
	if err := svc.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	db, err := database.InitDB(cfg.Paths.DBPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateUser("existing", "hash", "admin"); err != nil {
		t.Fatal(err)
	}
	db.Close()
	svc, err = StartDesktop(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Shutdown(context.Background())
	if svc.TakeDesktopSession() != "" {
		t.Fatal("existing account bypassed")
	}
}
