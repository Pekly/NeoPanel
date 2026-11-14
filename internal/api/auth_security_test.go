package api

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func securityAuth(t *testing.T) (*database.DB, *AuthHandler) {
	t.Helper()
	db, err := database.InitDB(filepath.Join(t.TempDir(), "security.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db, NewAuthHandler(db, config.DefaultConfig())
}

func securityRequest(handler http.HandlerFunc, body string, claims *JWTClaims) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "http://panel.example/", strings.NewReader(body))
	if claims != nil {
		r = r.WithContext(context.WithValue(r.Context(), contextUserKey, claims))
	}
	w := httptest.NewRecorder()
	handler(w, r)
	return w
}

func TestSecurityResetResponseNeverDisclosesCredential(t *testing.T) {
	for _, mode := range []string{"missing", "disabled", "no-email", "failed", "success"} {
		t.Run(mode, func(t *testing.T) {
			smtp := newInvitationSMTP(t)
			h := newInvitationHarness(t, smtp)
			if err := h.db.SetPanelPublicURL("https://panel.example", "panel.example"); err != nil {
				t.Fatal(err)
			}
			if err := h.db.CreatePasswordResetToken(h.owner.ID, "previously-delivered", time.Now().Add(time.Hour)); err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "missing":
				h.db.SetPanelSetting("smtp_host", "")
			case "disabled":
				h.db.SetPanelSetting("smtp_enabled", "false")
			case "no-email":
				h.db.UpdateUserEmail(h.owner.ID, "")
			case "failed":
				smtp.fail.Store(true)
			}
			existing := securityRequest(h.router.email.RequestPasswordReset, `{"identity":"owner"}`, nil)
			unknown := securityRequest(h.router.email.RequestPasswordReset, `{"identity":"unknown"}`, nil)
			if existing.Code != 200 || unknown.Code != 200 || existing.Body.String() != unknown.Body.String() {
				t.Fatalf("distinguishable responses: %s / %s", existing.Body, unknown.Body)
			}
			if strings.Contains(existing.Body.String(), "token") || strings.Contains(existing.Body.String(), "https:") || strings.Contains(existing.Body.String(), "SMTP") {
				t.Fatal("public response discloses recovery information", existing.Body)
			}
			prior, err := h.db.GetPasswordResetToken("previously-delivered")
			if err != nil || prior.Used != (mode == "success") {
				t.Fatal("unavailable delivery changed the active recovery credential", err)
			}
			if mode == "success" {
				select {
				case message := <-smtp.messages:
					match := regexp.MustCompile(`#reset-token=([a-f0-9]{64})`).FindStringSubmatch(message)
					if len(match) != 2 {
						t.Fatal("missing email credential")
					}
					pr, err := h.db.GetPasswordResetToken(match[1])
					if err != nil || pr.Used || pr.UserID != h.owner.ID {
						t.Fatal("delivered credential not valid", err)
					}
					if strings.Contains(existing.Body.String(), match[1]) {
						t.Fatal("credential in response")
					}
				case <-time.After(time.Second):
					t.Fatal("no email")
				}
			} else {
				select {
				case <-smtp.messages:
					t.Fatal("unexpected delivery")
				default:
				}
			}
		})
	}
}

type securityBlockedBody struct {
	entered chan struct{}
	release chan struct{}
	reader  io.Reader
	once    sync.Once
}

func (b *securityBlockedBody) Read(p []byte) (int, error) {
	b.once.Do(func() { close(b.entered); <-b.release })
	return b.reader.Read(p)
}
func (b *securityBlockedBody) Close() error { return nil }

func TestSecuritySetupDelayedBodyCannotCreateSecondAdmin(t *testing.T) {
	db, auth := securityAuth(t)
	body := &securityBlockedBody{entered: make(chan struct{}), release: make(chan struct{}), reader: strings.NewReader(`{"username":"attacker","email":"attacker@example.test","password":"password"}`)}
	req := httptest.NewRequest("POST", "/api/auth/setup", body)
	w := httptest.NewRecorder()
	done := make(chan struct{})
	go func() { auth.Setup(w, req); close(done) }()
	<-body.entered
	real := securityRequest(auth.Setup, `{"username":"admin","email":"admin@example.test","password":"password"}`, nil)
	close(body.release)
	<-done
	if real.Code != 200 || w.Code != 400 {
		t.Fatalf("setup responses %d/%d: %s", real.Code, w.Code, w.Body)
	}
	users, err := db.ListUsers()
	if err != nil || len(users) != 1 || users[0].Username != "admin" || users[0].Role != "admin" {
		t.Fatal("extra admin", users, err)
	}
}

func TestSecurityConcurrentSetupOnlyOneAdmin(t *testing.T) {
	db, auth := securityAuth(t)
	var wg sync.WaitGroup
	codes := make(chan int, 8)
	start := make(chan struct{})
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			w := securityRequest(auth.Setup, fmt.Sprintf(`{"username":"admin%d","email":"admin%d@example.test","password":"password"}`, i, i), nil)
			codes <- w.Code
		}(i)
	}
	close(start)
	wg.Wait()
	close(codes)
	successes := 0
	for code := range codes {
		if code == 200 {
			successes++
		} else if code != 400 {
			t.Fatalf("unexpected status %d", code)
		}
	}
	n, err := db.GetUserCount()
	if err != nil || n != 1 || successes != 1 {
		t.Fatalf("users=%d successes=%d err=%v", n, successes, err)
	}
}

func TestSecurityForwardedHeadersCannotBypassQuota(t *testing.T) {
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP", "CF-Connecting-IP"} {
		for _, spoof := range []string{"127.0.0.1", "::1", "198.51.100.1"} {
			t.Run(header+spoof, func(t *testing.T) {
				lim := NewIPRateLimiter(60, 3)
				for i := 0; i < 2; i++ {
					r := httptest.NewRequest("POST", "/api/auth/login", nil)
					r.RemoteAddr = "203.0.113.25:9999"
					r.Header.Set(header, spoof)
					if extractClientIP(r) != "203.0.113.25" {
						t.Fatal("trusted attacker header")
					}
					if lim.Allow(r) != (i == 0) {
						t.Fatal("quota bypass")
					}
				}
			})
		}
	}
	lim := NewIPRateLimiter(60, 3)
	for i := 0; i < 10; i++ {
		r := httptest.NewRequest("POST", "/api/auth/password-reset/request", nil)
		r.RemoteAddr = "203.0.113.25:9999"
		r.Header.Set("X-Forwarded-For", fmt.Sprintf("198.51.100.%d", i))
		if lim.Allow(r) != (i == 0) {
			t.Fatal("rotating headers bypass quota")
		}
	}
	for _, peer := range []string{"127.0.0.1:10", "[::1]:10"} {
		lim := NewIPRateLimiter(60, 3)
		r := httptest.NewRequest("POST", "/api/auth/login", nil)
		r.RemoteAddr = peer
		if !lim.Allow(r) || lim.Allow(r) {
			t.Fatal("loopback exempt")
		}
	}
}

func TestSecurityTrustedProxyRightToLeftAndRouterIsolation(t *testing.T) {
	policy := newClientIPPolicy([]string{"127.0.0.1/32", "10.0.0.0/8"})
	r := httptest.NewRequest("POST", "/api/auth/login", nil)
	r.RemoteAddr = "127.0.0.1:80"
	r.Header.Set("X-Forwarded-For", "127.0.0.1, 203.0.113.25, 10.1.2.3")
	if got := extractClientIP(policy.request(r)); got != "203.0.113.25" {
		t.Fatal("spoof-left accepted", got)
	}
	if got := extractClientIP(newClientIPPolicy(nil).request(r)); got != "127.0.0.1" {
		t.Fatal("global trust leaked", got)
	}
	for _, value := range []string{strings.Repeat("1", 4097), strings.Repeat("10.1.2.3,", 33), "invalid", "203.0.113.25, invalid"} {
		r.Header.Set("X-Forwarded-For", value)
		if policy.resolve(r) != "127.0.0.1" {
			t.Fatal("bad chain accepted")
		}
	}
	r.Header.Del("X-Forwarded-For")
	r.Header.Set("X-Real-IP", "1.1.1.1")
	r.Header.Set("CF-Connecting-IP", "1.1.1.1")
	if policy.resolve(r) != "127.0.0.1" {
		t.Fatal("non-XFF header accepted")
	}
}

func TestSecurityEnable2FACannotReplaceFactor(t *testing.T) {
	db, auth := securityAuth(t)
	user, err := db.CreateUser("alice", "hash", "user")
	if err != nil {
		t.Fatal(err)
	}
	account := NewAccountHandler(db, auth)
	claims := &JWTClaims{UserID: user.ID, Username: user.Username}
	secret, _ := GenerateTOTPSecret()
	other, _ := GenerateTOTPSecret()
	payload := func(secret string) string {
		return fmt.Sprintf(`{"secret":%q,"code":%q}`, secret, calculateTOTP(secret, time.Now().Unix()/30))
	}
	// The stale enrollment has already loaded the unenrolled user before the real enrollment.
	body := &securityBlockedBody{entered: make(chan struct{}), release: make(chan struct{}), reader: strings.NewReader(payload(other))}
	req := httptest.NewRequest("POST", "/api/account/2fa/enable", body).WithContext(context.WithValue(context.Background(), contextUserKey, claims))
	stale := httptest.NewRecorder()
	done := make(chan struct{})
	go func() { account.Enable2FA(stale, req); close(done) }()
	<-body.entered
	first := securityRequest(account.Enable2FA, payload(secret), claims)
	saved, _ := db.GetUserByID(user.ID)
	close(body.release)
	<-done
	second := securityRequest(account.Enable2FA, payload(other), claims)
	if first.Code != 200 || stale.Code != 409 || second.Code != 409 {
		t.Fatalf("statuses %d/%d/%d", first.Code, stale.Code, second.Code)
	}
	current, _ := db.GetUserByID(user.ID)
	if !current.TwoFactorEnabled || current.TwoFactorSecret != secret || current.RecoveryCodes != saved.RecoveryCodes {
		t.Fatal("factor replaced")
	}
	denied := securityRequest(account.Disable2FA, `{"password":"wrong","code":"wrong"}`, claims)
	if denied.Code != 401 {
		t.Fatal("disable bypass")
	}
}

func TestSecurityPasswordMutationsRevokeSessions(t *testing.T) {
	for _, mode := range []string{"self", "admin", "reset"} {
		t.Run(mode, func(t *testing.T) {
			db, auth := securityAuth(t)
			hash, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.MinCost)
			user, err := db.CreateUserWithDetails("alice", "alice@example.test", string(hash), "admin", "")
			if err != nil {
				t.Fatal(err)
			}
			old, _ := auth.GenerateToken(user)
			pending, _ := auth.Generate2FATempToken(user)
			claims, _ := auth.ValidateToken(old)
			key := "neo_security_test"
			sum := sha256.Sum256([]byte(key))
			if _, err := db.AddAPIKey(user.ID, "test", hex.EncodeToString(sum[:]), "neo_test"); err != nil {
				t.Fatal(err)
			}
			account := NewAccountHandler(db, auth)
			var w *httptest.ResponseRecorder
			switch mode {
			case "self":
				w = securityRequest(account.UpdatePassword, `{"current_password":"oldpassword","new_password":"newpassword"}`, claims)
			case "admin":
				mux := http.NewServeMux()
				mux.HandleFunc("PUT /users/{id}", auth.AdminUpdateUser)
				r := httptest.NewRequest("PUT", fmt.Sprintf("/users/%d", user.ID), strings.NewReader(`{"username":"alice","email":"alice@example.test","role":"admin","password":"newpassword"}`))
				w = httptest.NewRecorder()
				mux.ServeHTTP(w, r)
			case "reset":
				if err := db.CreatePasswordResetToken(user.ID, "recovery", time.Now().Add(time.Hour)); err != nil {
					t.Fatal(err)
				}
				email := NewEmailHandler(db, auth.cfg, auth)
				w = securityRequest(email.ConfirmPasswordReset, `{"token":"recovery","new_password":"newpassword"}`, nil)
			}
			if w.Code != 200 {
				t.Fatalf("change failed: %d %s", w.Code, w.Body)
			}
			if _, err := auth.ValidateToken(old); err == nil {
				t.Fatal("old session valid")
			}
			if _, err := auth.ValidateAuthToken(old); err == nil {
				t.Fatal("old auth session valid")
			}
			if _, err := auth.Validate2FATempToken(pending); err == nil {
				t.Fatal("pending session valid")
			}
			if _, err := auth.ValidateAuthToken(key); err != nil {
				t.Fatal("API credential silently revoked", err)
			}
			// Also exercise the stale middleware-to-handler window, not just middleware rejection.
			renew := securityRequest(account.UpdateProfile, `{"username":"alice","email":"alice@example.test"}`, claims)
			if renew.Code != 401 {
				t.Fatal("stale JWT renewed", renew.Body)
			}
			login := securityRequest(auth.Login, `{"username":"alice","password":"newpassword"}`, nil)
			var response struct{ Token string }
			json.Unmarshal(login.Body.Bytes(), &response)
			if login.Code != 200 || response.Token == "" {
				t.Fatal("new login failed", login.Body)
			}
			fresh, err := auth.ValidateToken(response.Token)
			if err != nil || fresh.AuthVersion != 1 {
				t.Fatal("new login version", err)
			}
			renew = securityRequest(account.UpdateProfile, `{"username":"alice","email":"alice@example.test"}`, fresh)
			json.Unmarshal(renew.Body.Bytes(), &response)
			if renew.Code != 200 {
				t.Fatal("fresh renewal failed", renew.Body)
			}
			if _, err := auth.ValidateToken(response.Token); err != nil {
				t.Fatal("renewal used wrong version", err)
			}
		})
	}
}

func TestSecurityAuthBodiesBounded(t *testing.T) {
	_, auth := securityAuth(t)
	email := NewEmailHandler(auth.db, auth.cfg, auth)
	for name, handler := range map[string]http.HandlerFunc{"setup": auth.Setup, "login": auth.Login, "2fa": auth.Login2FA, "request-reset": email.RequestPasswordReset, "confirm-reset": email.ConfirmPasswordReset} {
		for _, chunked := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/chunked=%v", name, chunked), func(t *testing.T) {
				// A valid leading document must not bypass the cap via trailing whitespace.
				r := httptest.NewRequest("POST", "/api/auth/"+name, strings.NewReader(`{}`+strings.Repeat(" ", int(authBodyLimit))))
				if chunked {
					r.ContentLength = -1
					r.TransferEncoding = []string{"chunked"}
				}
				w := httptest.NewRecorder()
				handler(w, r)
				if w.Code != 400 {
					t.Fatalf("oversize body accepted: %d %s", w.Code, w.Body)
				}
			})
		}
	}
	for _, path := range []string{"/api/auth/login", "/api/auth/setup", "/api/auth/password-reset/request"} {
		r := httptest.NewRequest("POST", path, strings.NewReader(strings.Repeat("x", int(authBodyLimit)+1)))
		w := httptest.NewRecorder()
		router := &Router{mux: http.NewServeMux()}
		router.ServeHTTP(w, r)
		if w.Code != 413 {
			t.Fatal("known-length body not rejected", w.Code)
		}
	}
	for _, path := range []string{"/api/account/passkeys/register/finish", "/api/auth/passkeys/login/finish", "/api/servers/test/files/upload"} {
		body := strings.Repeat("x", int(authBodyLimit)*2)
		r := httptest.NewRequest("POST", path, strings.NewReader(body))
		w := httptest.NewRecorder()
		if !limitAuthBody(w, r) {
			t.Fatal("non-password payload wrongly capped", path)
		}
		got, err := io.ReadAll(r.Body)
		if err != nil || len(got) != len(body) {
			t.Fatal("non-password payload truncated", path, err)
		}
	}
}

func TestSecurityProfileDelayedBodyCannotOutliveReset(t *testing.T) {
	for _, apiKey := range []bool{false, true} {
		t.Run(fmt.Sprintf("api-key=%v", apiKey), func(t *testing.T) {
			db, auth := securityAuth(t)
			user, err := db.CreateUserWithDetails("alice", "alice@example.test", "hash", "user", "")
			if err != nil {
				t.Fatal(err)
			}
			claims := &JWTClaims{UserID: user.ID, Username: user.Username, AuthVersion: user.AuthVersion, APIKey: apiKey}
			account := NewAccountHandler(db, auth)
			body := &securityBlockedBody{entered: make(chan struct{}), release: make(chan struct{}), reader: strings.NewReader(`{"username":"changed","first_name":"First","last_name":"Last","email":"changed@example.test","avatar_url":"/changed.png"}`)}
			req := httptest.NewRequest("PUT", "/api/account/profile", body).WithContext(context.WithValue(context.Background(), contextUserKey, claims))
			w := httptest.NewRecorder()
			done := make(chan struct{})
			go func() { account.UpdateProfile(w, req); close(done) }()
			<-body.entered // The handler has loaded the old version; the body is still blocked.
			err = db.CreatePasswordResetToken(user.ID, "reset-profile", time.Now().Add(time.Hour))
			if err == nil {
				err = db.ResetUserPassword("reset-profile", "newhash")
			}
			close(body.release)
			<-done
			if err != nil {
				t.Fatal(err)
			}
			saved, err := db.GetUserByID(user.ID)
			if err != nil {
				t.Fatal(err)
			}
			var response struct{ Token string }
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if apiKey {
				if w.Code != 200 || saved.Email != "changed@example.test" || saved.AvatarURL != "/changed.png" {
					t.Fatal("password reset revoked API-key mutation", w.Code, w.Body)
				}
			} else {
				if w.Code != 401 || response.Token != "" || len(w.Result().Cookies()) != 0 {
					t.Fatal("revoked session renewed", w.Code, w.Body)
				}
				if saved.Username != user.Username || saved.FirstName != user.FirstName || saved.LastName != user.LastName || saved.Email != user.Email || saved.AvatarURL != user.AvatarURL {
					t.Fatal("revoked session changed profile", saved)
				}
			}
			if response.Token != "" {
				if _, err := auth.ValidateToken(response.Token); err == nil {
					t.Fatal("delayed request minted upgraded session")
				}
			}
		})
	}
}

func TestSecurityRecoveryLoginRequiresCommittedConsumption(t *testing.T) {
	for _, failure := range []string{"none", "stale", "error"} {
		t.Run(failure, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "recovery-login.db")
			db, err := database.InitDB(path)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			auth := NewAuthHandler(db, config.DefaultConfig())
			user, err := db.CreateUser("alice", "hash", "user")
			if err != nil {
				t.Fatal(err)
			}
			if err := db.EnableUserTwoFactor(user.ID, "JBSWY3DPEHPK3PXP", "aaaa-bbbb,cccc-dddd"); err != nil {
				t.Fatal(err)
			}
			token, err := auth.Generate2FATempToken(user)
			if err != nil {
				t.Fatal(err)
			}
			if failure != "none" {
				conn, err := sql.Open("sqlite", path)
				if err != nil {
					t.Fatal(err)
				}
				defer conn.Close()
				raise := "IGNORE" // Simulate zero affected rows after successful snapshot verification.
				if failure == "error" {
					raise = "ABORT, 'write failed'"
				}
				if _, err := conn.Exec(`CREATE TRIGGER reject_recovery BEFORE UPDATE OF recovery_codes ON users BEGIN SELECT RAISE(` + raise + `); END`); err != nil {
					t.Fatal(err)
				}
			}
			payload := fmt.Sprintf(`{"temp_token":%q,"code":"aaaa-bbbb"}`, token)
			w := securityRequest(auth.Login2FA, payload, nil)
			var response struct{ Token string }
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			saved, err := db.GetUserByID(user.ID)
			if err != nil {
				t.Fatal(err)
			}
			if failure == "none" {
				if w.Code != 200 || response.Token == "" || len(w.Result().Cookies()) != 1 || saved.RecoveryCodes != "cccc-dddd" {
					t.Fatal("committed recovery login failed", w.Code, w.Body)
				}
				if _, err := auth.ValidateToken(response.Token); err != nil {
					t.Fatal(err)
				}
				replay := securityRequest(auth.Login2FA, payload, nil)
				if replay.Code != 401 || len(replay.Result().Cookies()) != 0 || strings.Contains(replay.Body.String(), `"token"`) {
					t.Fatal("recovery code replayed", replay.Code, replay.Body)
				}
			} else {
				if w.Code != 401 || response.Token != "" || len(w.Result().Cookies()) != 0 || saved.RecoveryCodes != "aaaa-bbbb,cccc-dddd" {
					t.Fatal("uncommitted recovery issued session", w.Code, w.Body)
				}
			}
		})
	}
}
