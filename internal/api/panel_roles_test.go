package api

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pekly/NeoPanel/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func retirementAPIKey(t *testing.T, f *reliabilityFixture, user string) string {
	t.Helper()
	token := "neo_retirement_" + user
	sum := sha256.Sum256([]byte(token))
	if _, err := f.db.AddAPIKey(f.users[user].ID, "existing-key", hex.EncodeToString(sum[:]), "neo_retire"); err != nil {
		t.Fatal(err)
	}
	return token
}

// Tokens and users already exist before this simulated upgrade. Only a fixture's
// t.TempDir database is opened; never a configured or production database.
func retireLegacyFixture(t *testing.T, f *reliabilityFixture) {
	t.Helper()
	path := filepath.Join(f.cfg.Paths.DataDir, "panel.db")
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err := conn.Exec(`CREATE TABLE panel_roles(id TEXT PRIMARY KEY, name TEXT, permissions TEXT);
 CREATE TABLE panel_user_roles(user_id INTEGER PRIMARY KEY, role_id TEXT);
 INSERT INTO panel_roles VALUES('creator','Creators','["servers.create"]');
 INSERT INTO panel_user_roles SELECT id,'creator' FROM users;`); err != nil {
		t.Fatal(err)
	}
	upgraded, err := database.InitDB(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := upgraded.Close(); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := conn.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name IN ('panel_roles','panel_user_roles')`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("legacy schema survived: %d %v", count, err)
	}
}

func assertNoCustomRoleJSON(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Fatalf("identity response: %d %s", w.Code, w.Body)
	}
	for _, field := range []string{"panel_role_id", "panel_role_name", "panel_permissions"} {
		if strings.Contains(w.Body.String(), `"`+field+`"`) {
			t.Fatalf("retired field %s: %s", field, w.Body)
		}
	}
}

func TestPanelRoleRetirementIdentityResponses(t *testing.T) {
	f := newReliabilityFixture(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.db.UpdateUserPassword(f.users["stranger"].ID, string(hash)); err != nil {
		t.Fatal(err)
	}
	current, err := f.db.GetUserByID(f.users["stranger"].ID)
	if err != nil {
		t.Fatal(err)
	}
	f.tokens["stranger"], err = f.router.auth.GenerateToken(current)
	if err != nil {
		t.Fatal(err)
	}
	retireLegacyFixture(t, f)
	assertNoCustomRoleJSON(t, f.request("POST", "/api/auth/login", `{"username":"stranger","password":"password"}`, ""))
	for _, user := range []string{"stranger", "admin"} {
		assertNoCustomRoleJSON(t, f.request("GET", "/api/auth/me", "", user))
	}
	assertNoCustomRoleJSON(t, f.request("GET", "/api/admin/users", "", "admin"))
	f.tokens["machine"] = f.cfg.Security.NodeToken
	assertNoCustomRoleJSON(t, f.request("GET", "/api/auth/me", "", "machine"))
}

func TestAdminDemotionRefreshesExistingTokens(t *testing.T) {
	f := newReliabilityFixture(t)
	token := retirementAPIKey(t, f, "admin")
	staleClaims, err := f.router.auth.ValidateAuthToken(f.tokens["admin"])
	if err != nil {
		t.Fatal(err)
	}
	if err := f.db.UpdateUserRole(f.users["admin"].ID, "user"); err != nil {
		t.Fatal(err)
	}
	if err := f.db.UpdateUserProfile(f.users["admin"].ID, "renamed", "", "", ""); err != nil {
		t.Fatal(err)
	}
	for _, tok := range []string{token, f.tokens["admin"]} {
		claims, err := f.router.auth.ValidateAuthToken(tok)
		if err != nil || claims.Role != "user" || claims.Username != "renamed" {
			t.Fatalf("stale identity: %+v %v", claims, err)
		}
		req := httptest.NewRequest("POST", "/api/servers", nil)
		req = req.WithContext(context.WithValue(req.Context(), contextUserKey, claims))
		if ok, err := f.router.auth.CanCreateServer(req); ok || err != nil {
			t.Fatalf("demoted admin can create: %v %v", ok, err)
		}
	}
	// Handlers must also reject a previously authenticated admin context, not
	// merely rely on middleware refreshing the JWT on the next request.
	req := httptest.NewRequest("POST", "/api/servers", strings.NewReader(`{"name":"Denied","type":"custom"}`))
	req = req.WithContext(context.WithValue(req.Context(), contextUserKey, staleClaims))
	if ok, err := f.router.auth.CanCreateServer(req); ok || err != nil {
		t.Fatalf("stale context can create: %v %v", ok, err)
	}
	w := httptest.NewRecorder()
	f.router.auth.ListUsers(w, req)
	if w.Code != 403 {
		t.Fatalf("stale context lists users: %d %s", w.Code, w.Body)
	}
	if w := f.request("GET", "/api/admin/users", "", "admin"); w.Code != 403 {
		t.Fatalf("demoted JWT admin: %d", w.Code)
	}
}

func TestServerCreationAuthorizationFailsClosed(t *testing.T) {
	f := newReliabilityFixture(t)
	for _, claims := range []*JWTClaims{nil, {UserID: -1, Role: "admin"}, {UserID: f.users["owner"].ID, Role: "admin"}} {
		req := httptest.NewRequest("POST", "/api/servers", nil)
		req = req.WithContext(context.WithValue(req.Context(), contextUserKey, claims))
		if ok, _ := f.router.auth.CanCreateServer(req); ok {
			t.Fatalf("invalid identity accepted: %+v", claims)
		}
	}
	claims := &JWTClaims{UserID: f.users["admin"].ID, Role: "admin"}
	req := httptest.NewRequest("POST", "/api/servers", nil)
	req = req.WithContext(context.WithValue(req.Context(), contextUserKey, claims))
	if err := f.db.DeleteUser(claims.UserID); err != nil {
		t.Fatal(err)
	}
	if ok, err := f.router.auth.CanCreateServer(req); ok || err == nil {
		t.Fatalf("deleted admin accepted: %v %v", ok, err)
	}
	if _, err := f.router.auth.ValidateAuthToken(f.tokens["admin"]); err == nil {
		t.Fatal("deleted account JWT accepted")
	}
	if err := f.db.Close(); err != nil {
		t.Fatal(err)
	}
	if ok, err := f.router.auth.CanCreateServer(req); ok || err == nil {
		t.Fatalf("unavailable storage accepted: %v %v", ok, err)
	}
}
