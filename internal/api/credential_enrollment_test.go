package api

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func enrollmentPayload(t *testing.T, kind string) string {
	t.Helper()
	if kind == "api-keys" {
		return `{"name":"enrolled"}`
	}
	public, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	key, err := ssh.NewPublicKey(public)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(map[string]string{"name": "enrolled", "public_key": string(ssh.MarshalAuthorizedKey(key))})
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}

func enrollmentSource(t *testing.T, f *reliabilityFixture, name string) (string, int) {
	t.Helper()
	token := "neo_enrollment_test_" + name
	sum := sha256.Sum256([]byte(token))
	key, err := f.db.AddAPIKey(f.users[name].ID, "source", hex.EncodeToString(sum[:]), "neo_test")
	if err != nil {
		t.Fatal(err)
	}
	return token, key.ID
}

func assertEnrollment(t *testing.T, f *reliabilityFixture, kind, name string, w *httptest.ResponseRecorder, status, count int) {
	t.Helper()
	if w.Code != status {
		t.Fatalf("status=%d, want %d", w.Code, status)
	}
	var response struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK && response.Token != "" {
		t.Fatal("failed enrollment disclosed a secret")
	}
	if status == http.StatusOK && kind == "api-keys" {
		claims, err := f.router.auth.ValidateAuthToken(response.Token)
		if err != nil || claims.UserID != f.users[name].ID {
			t.Fatal("new API credential is not valid for its authenticated owner")
		}
	}
	found := 0
	if kind == "api-keys" {
		keys, err := f.db.ListAPIKeys(f.users[name].ID)
		if err != nil {
			t.Fatal(err)
		}
		for _, key := range keys {
			if key.Name == "enrolled" {
				found++
			}
		}
	} else {
		keys, err := f.db.ListSSHKeys(f.users[name].ID)
		if err != nil {
			t.Fatal(err)
		}
		found = len(keys)
	}
	if found != count {
		t.Fatalf("inserted credentials=%d, want %d", found, count)
	}
}

func TestCredentialEnrollmentDelayedBody(t *testing.T) {
	for _, kind := range []string{"api-keys", "ssh-keys"} {
		for _, mode := range []string{"jwt-reset", "api-reset", "api-revoked", "jwt-renamed"} {
			t.Run(kind+"/"+mode, func(t *testing.T) {
				f := newReliabilityFixture(t)
				user := f.users["owner"]
				token := f.tokens["owner"]
				var sourceID int
				if strings.HasPrefix(mode, "api-") {
					token, sourceID = enrollmentSource(t, f, "owner")
				}
				body := &securityBlockedBody{entered: make(chan struct{}), release: make(chan struct{}), reader: strings.NewReader(enrollmentPayload(t, kind))}
				req := httptest.NewRequest("POST", "/api/account/"+kind, body)
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				done := make(chan struct{})
				go func() { f.router.ServeHTTP(w, req); close(done) }()
				select {
				case <-body.entered:
				case <-done:
					t.Fatalf("router returned before reading the body: %d", w.Code)
				case <-time.After(5 * time.Second):
					close(body.release)
					t.Fatal("body was not reached")
				}
				var err error
				switch mode {
				case "api-revoked":
					err = f.db.DeleteAPIKey(user.ID, sourceID)
				case "jwt-renamed":
					err = f.db.AdminUpdateUser(user.ID, "renamed", "", "", "", "user", "")
					if err == nil {
						_, err = f.db.CreateUser(user.Username, "hash", "user")
					}
				default:
					err = f.db.CreatePasswordResetToken(user.ID, "reset-enrollment", time.Now().Add(time.Hour))
					if err == nil {
						err = f.db.ResetUserPassword("reset-enrollment", "newhash")
					}
				}
				close(body.release)
				select {
				case <-done:
				case <-time.After(5 * time.Second):
					t.Fatal("enrollment did not finish")
				}
				if err != nil {
					t.Fatal(err)
				}
				status, count := http.StatusUnauthorized, 0
				if mode == "api-reset" || mode == "jwt-renamed" {
					status, count = http.StatusOK, 1
				}
				assertEnrollment(t, f, kind, "owner", w, status, count)
			})
		}
	}
}

func TestCredentialEnrollmentAuthentication(t *testing.T) {
	for _, kind := range []string{"api-keys", "ssh-keys"} {
		for _, principal := range []string{"owner", "admin", "node", "anonymous", "pending"} {
			for _, apiKey := range []bool{false, true} {
				if apiKey && principal != "owner" && principal != "admin" {
					continue
				}
				t.Run(fmt.Sprintf("%s/%s/api=%v", kind, principal, apiKey), func(t *testing.T) {
					f := newReliabilityFixture(t)
					token := f.tokens[principal]
					if apiKey {
						token, _ = enrollmentSource(t, f, principal)
					}
					if principal == "node" {
						// Even a human with the machine's display name cannot inherit its authority.
						user, err := f.db.CreateUser("node-daemon", "hash", "admin")
						if err != nil {
							t.Fatal(err)
						}
						f.users["node"] = user
						token = f.cfg.Security.NodeToken
					}
					if principal == "pending" {
						var err error
						token, err = f.router.auth.Generate2FATempToken(f.users["owner"])
						if err != nil {
							t.Fatal(err)
						}
					}
					req := httptest.NewRequest("POST", "/api/account/"+kind, strings.NewReader(enrollmentPayload(t, kind)))
					if token != "" {
						req.Header.Set("Authorization", "Bearer "+token)
					}
					req.Header.Set("Content-Type", "application/json")
					w := httptest.NewRecorder()
					f.router.ServeHTTP(w, req)
					if principal == "owner" || principal == "admin" {
						assertEnrollment(t, f, kind, principal, w, http.StatusOK, 1)
					} else {
						name := "owner"
						if principal == "node" {
							name = "node"
						}
						assertEnrollment(t, f, kind, name, w, http.StatusUnauthorized, 0)
					}
				})
			}
		}
	}
}

func TestCredentialEnrollmentSSHValidation(t *testing.T) {
	f := newReliabilityFixture(t)
	for _, public := range []string{"SHA256:only-a-fingerprint", "0123456789abcdef", "ssh-ed25519 not-base64", "-----BEGIN OPENSSH PRIVATE KEY-----\ninvalid\n-----END OPENSSH PRIVATE KEY-----"} {
		body, _ := json.Marshal(map[string]string{"name": "enrolled", "public_key": public})
		w := f.request("POST", "/api/account/ssh-keys", string(body), "owner")
		assertEnrollment(t, f, "ssh-keys", "owner", w, http.StatusBadRequest, 0)
		if !strings.Contains(w.Body.String(), ".pub") {
			t.Fatal("public-key validation guidance missing")
		}
	}
}

func TestCredentialEnrollmentAPIKeyCookie(t *testing.T) {
	for _, kind := range []string{"api-keys", "ssh-keys"} {
		t.Run(kind, func(t *testing.T) {
			f := newReliabilityFixture(t)
			token, _ := enrollmentSource(t, f, "owner")
			req := httptest.NewRequest("POST", "http://panel.example/api/account/"+kind, strings.NewReader(enrollmentPayload(t, kind)))
			req.Header.Set("Origin", "http://panel.example")
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(&http.Cookie{Name: "neopanel_token", Value: token})
			w := httptest.NewRecorder()
			f.router.ServeHTTP(w, req)
			assertEnrollment(t, f, kind, "owner", w, http.StatusOK, 1)
		})
	}
}
