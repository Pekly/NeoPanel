package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// The body barrier is reached only after router authentication and credential
// capture. Revocation completes before any request payload reaches the handler.
func TestAccountCredentialDelayedBody(t *testing.T) {
	for _, kind := range []string{"profile", "totp"} {
		for _, mode := range []string{"jwt-normal", "api-normal", "jwt-reset", "api-reset", "api-revoked", "api-cookie-revoked", "existing-factor"} {
			if kind == "profile" && mode == "existing-factor" {
				continue
			}
			t.Run(kind+"/"+mode, func(t *testing.T) {
				f := newReliabilityFixture(t)
				original := f.users["owner"]
				token := f.tokens["owner"]
				var sourceID int
				if strings.HasPrefix(mode, "api-") {
					token, sourceID = enrollmentSource(t, f, "owner")
					// A different surviving key must not authorize the deleted key.
					if _, err := f.db.AddAPIKey(original.ID, "other", "other-hash", "other"); err != nil {
						t.Fatal(err)
					}
				}
				method, path := "PUT", "/api/account/profile"
				payload := `{"username":"changed","first_name":"First","last_name":"Last","email":"changed@example.test","avatar_url":"/changed.png"}`
				secret := ""
				if kind == "totp" {
					var err error
					secret, err = GenerateTOTPSecret()
					if err != nil {
						t.Fatal(err)
					}
					method, path = "POST", "/api/account/2fa/enable"
					payload = fmt.Sprintf(`{"secret":%q,"code":%q}`, secret, calculateTOTP(secret, time.Now().Unix()/30))
				}
				body := &securityBlockedBody{entered: make(chan struct{}), release: make(chan struct{}), reader: strings.NewReader(payload)}
				req := httptest.NewRequest(method, "http://panel.example"+path, body)
				req.Header.Set("Content-Type", "application/json")
				if mode == "api-cookie-revoked" {
					req.Header.Set("Origin", "http://panel.example")
					req.AddCookie(&http.Cookie{Name: "neopanel_token", Value: token})
				} else {
					req.Header.Set("Authorization", "Bearer "+token)
				}
				w := httptest.NewRecorder()
				done := make(chan struct{})
				go func() { f.router.ServeHTTP(w, req); close(done) }()
				select {
				case <-body.entered:
				case <-done:
					t.Fatalf("returned before body barrier: %d", w.Code)
				case <-time.After(5 * time.Second):
					close(body.release)
					t.Fatal("body barrier not reached")
				}
				var err error
				status := http.StatusOK
				switch mode {
				case "jwt-reset", "api-reset":
					err = f.db.CreatePasswordResetToken(original.ID, "account-reset", time.Now().Add(time.Hour))
					if err == nil {
						err = f.db.ResetUserPassword("account-reset", "newhash")
					}
					if mode == "jwt-reset" {
						status = http.StatusUnauthorized
					}
				case "api-revoked", "api-cookie-revoked":
					err = f.db.DeleteAPIKey(original.ID, sourceID)
					status = http.StatusUnauthorized
				case "existing-factor":
					err = f.db.EnableUserTwoFactor(original.ID, "existing-secret", "existing-codes")
					status = http.StatusConflict
				}
				close(body.release)
				select {
				case <-done:
				case <-time.After(5 * time.Second):
					t.Fatal("request did not finish")
				}
				if err != nil {
					t.Fatal(err)
				}
				if w.Code != status {
					t.Fatalf("status=%d, want %d", w.Code, status)
				}
				var response struct {
					Token         string   `json:"token"`
					RecoveryCodes []string `json:"recovery_codes"`
				}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatal(err)
				}
				saved, err := f.db.GetUserByID(original.ID)
				if err != nil {
					t.Fatal(err)
				}
				if status != http.StatusOK {
					if response.Token != "" || len(response.RecoveryCodes) != 0 || len(w.Result().Cookies()) != 0 {
						t.Fatal("failed mutation disclosed credentials or issued cookie")
					}
					if saved.Username != original.Username || saved.FirstName != original.FirstName || saved.LastName != original.LastName || saved.Email != original.Email || saved.AvatarURL != original.AvatarURL {
						t.Fatal("failed request changed profile")
					}
					if mode == "existing-factor" {
						if !saved.TwoFactorEnabled || saved.TwoFactorSecret != "existing-secret" || saved.RecoveryCodes != "existing-codes" {
							t.Fatal("existing factor replaced")
						}
					} else if saved.TwoFactorEnabled || saved.TwoFactorSecret != "" || saved.RecoveryCodes != "" {
						t.Fatal("revoked credential enrolled a factor")
					}
					return
				}
				if kind == "totp" {
					if !saved.TwoFactorEnabled || saved.TwoFactorSecret != secret || len(response.RecoveryCodes) != 8 || saved.RecoveryCodes != strings.Join(response.RecoveryCodes, ",") {
						t.Fatal("factor enrollment did not persist returned recovery codes")
					}
				} else {
					if saved.Username != "changed" || saved.FirstName != "First" || saved.LastName != "Last" || saved.Email != "changed@example.test" || saved.AvatarURL != "/changed.png" {
						t.Fatal("profile update not persisted")
					}
					claims, err := f.router.auth.ValidateToken(response.Token)
					if err != nil || claims.UserID != original.ID || claims.AuthVersion != saved.AuthVersion || claims.Username != saved.Username {
						t.Fatal("profile token does not match committed user")
					}
					cookies := w.Result().Cookies()
					if len(cookies) != 1 || cookies[0].Value != response.Token {
						t.Fatal("profile session cookie missing or incorrect")
					}
				}
			})
		}
	}
}
