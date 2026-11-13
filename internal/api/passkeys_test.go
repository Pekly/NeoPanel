package api

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/fxamacker/cbor/v2"
	"github.com/go-webauthn/webauthn/webauthn"
	"golang.org/x/crypto/bcrypt"
)

const passkeyTestOrigin = "http://localhost:8123"

func newPasskeyTest(t *testing.T) (*PasskeyHandler, *database.User, *database.User) {
	t.Helper()
	db, err := database.InitDB(filepath.Join(t.TempDir(), "passkeys.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	cfg := config.DefaultConfig()
	cfg.Server.Mode = "standalone"
	cfg.Server.PublicURL = passkeyTestOrigin
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	a, err := db.CreateUser("alice", string(hash), "user")
	if err != nil {
		t.Fatal(err)
	}
	b, err := db.CreateUser("bob", string(hash), "user")
	if err != nil {
		t.Fatal(err)
	}
	return NewPasskeyHandler(NewAuthHandler(db, cfg)), a, b
}

func passkeyRequest(t *testing.T, handler http.HandlerFunc, user *database.User, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, passkeyTestOrigin+"/api/test", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", passkeyTestOrigin)
	if user != nil {
		req = req.WithContext(context.WithValue(req.Context(), contextUserKey, &JWTClaims{UserID: user.ID, Username: user.Username, Role: user.Role}))
	}
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

type passkeyBeginResult struct {
	FlowID  string `json:"flow_id"`
	Options struct {
		PublicKey struct {
			Challenge string `json:"challenge"`
			RPID      string `json:"rpId"`
			User      struct {
				ID string `json:"id"`
			} `json:"user"`
			AuthenticatorSelection struct {
				ResidentKey        string `json:"residentKey"`
				UserVerification   string `json:"userVerification"`
				RequireResidentKey bool   `json:"requireResidentKey"`
			} `json:"authenticatorSelection"`
			UserVerification string        `json:"userVerification"`
			Attestation      string        `json:"attestation"`
			AllowCredentials []interface{} `json:"allowCredentials"`
		} `json:"publicKey"`
	} `json:"options"`
}

func passkeyBegin(t *testing.T, h *PasskeyHandler, user *database.User) passkeyBeginResult {
	t.Helper()
	handler := h.LoginBegin
	body := map[string]string{}
	if user != nil {
		handler = h.RegisterBegin
		body = map[string]string{"label": "Laptop", "password": "correct-password"}
	}
	w := passkeyRequest(t, handler, user, body)
	if w.Code != 200 {
		t.Fatalf("begin %d: %s", w.Code, w.Body.String())
	}
	var result passkeyBeginResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.FlowID == "" || result.Options.PublicKey.Challenge == "" {
		t.Fatalf("invalid begin: %s", w.Body.String())
	}
	return result
}

// These fixtures implement a test authenticator, not server verification. Each
// assertion is genuinely ES256-signed, so tests exercise the library's complete
// clientData/authenticatorData/COSE/signature validation and our actual handlers.
type passkeyAuthenticator struct {
	key    *ecdsa.PrivateKey
	id     []byte
	handle string
}

func newPasskeyAuthenticator(t *testing.T) *passkeyAuthenticator {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	id := make([]byte, 32)
	if _, err := rand.Read(id); err != nil {
		t.Fatal(err)
	}
	return &passkeyAuthenticator{key: key, id: id}
}
func passkeyJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func passkeyCBOR(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, e := cbor.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func passkeyB64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }
func passkeyAuthData(rp string, flags byte, count uint32) []byte {
	hash := sha256.Sum256([]byte(rp))
	data := append([]byte{}, hash[:]...)
	data = append(data, flags, 0, 0, 0, 0)
	binary.BigEndian.PutUint32(data[33:], count)
	return data
}
func (a *passkeyAuthenticator) registration(t *testing.T, b passkeyBeginResult, origin, rp string, flags byte) map[string]interface{} {
	t.Helper()
	a.handle = b.Options.PublicKey.User.ID
	client := passkeyJSON(t, map[string]interface{}{"type": "webauthn.create", "challenge": b.Options.PublicKey.Challenge, "origin": origin, "crossOrigin": false})
	cose := passkeyCBOR(t, map[int]interface{}{1: 2, 3: -7, -1: 1, -2: a.key.X.FillBytes(make([]byte, 32)), -3: a.key.Y.FillBytes(make([]byte, 32))})
	data := passkeyAuthData(rp, flags|0x40, 0)
	data = append(data, make([]byte, 16)...)
	data = append(data, byte(len(a.id)>>8), byte(len(a.id)))
	data = append(data, a.id...)
	data = append(data, cose...)
	att := passkeyCBOR(t, map[string]interface{}{"fmt": "none", "authData": data, "attStmt": map[string]interface{}{}})
	return map[string]interface{}{"id": passkeyB64(a.id), "rawId": passkeyB64(a.id), "type": "public-key", "response": map[string]interface{}{"clientDataJSON": passkeyB64(client), "attestationObject": passkeyB64(att), "transports": []string{"internal"}}, "clientExtensionResults": map[string]interface{}{}}
}
func (a *passkeyAuthenticator) assertion(t *testing.T, b passkeyBeginResult, origin, rp, handle string, flags byte, count uint32) map[string]interface{} {
	t.Helper()
	client := passkeyJSON(t, map[string]interface{}{"type": "webauthn.get", "challenge": b.Options.PublicKey.Challenge, "origin": origin, "crossOrigin": false})
	data := passkeyAuthData(rp, flags, count)
	hash := sha256.Sum256(client)
	signed := append(append([]byte{}, data...), hash[:]...)
	digest := sha256.Sum256(signed)
	sig, err := ecdsa.SignASN1(rand.Reader, a.key, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	return map[string]interface{}{"id": passkeyB64(a.id), "rawId": passkeyB64(a.id), "type": "public-key", "response": map[string]interface{}{"clientDataJSON": passkeyB64(client), "authenticatorData": passkeyB64(data), "signature": passkeyB64(sig), "userHandle": handle}, "clientExtensionResults": map[string]interface{}{}}
}
func registerPasskey(t *testing.T, h *PasskeyHandler, user *database.User, flags byte) (*passkeyAuthenticator, string) {
	t.Helper()
	a := newPasskeyAuthenticator(t)
	b := passkeyBegin(t, h, user)
	selection := b.Options.PublicKey.AuthenticatorSelection
	if selection.ResidentKey != "required" || !selection.RequireResidentKey || selection.UserVerification != "required" || b.Options.PublicKey.Attestation != "none" {
		t.Fatal("unsafe creation options")
	}
	credential := a.registration(t, b, passkeyTestOrigin, "localhost", flags)
	w := passkeyRequest(t, h.RegisterFinish, user, map[string]interface{}{"flow_id": b.FlowID, "credential": credential})
	if w.Code != 200 {
		t.Fatalf("registration %d: %s", w.Code, w.Body.String())
	}
	var result struct {
		Passkey database.Passkey `json:"passkey"`
	}
	if json.Unmarshal(w.Body.Bytes(), &result) != nil || result.Passkey.ID == "" {
		t.Fatal("missing metadata")
	}
	return a, result.Passkey.ID
}
func loginPasskey(t *testing.T, h *PasskeyHandler, a *passkeyAuthenticator, flags byte, count uint32) *httptest.ResponseRecorder {
	t.Helper()
	b := passkeyBegin(t, h, nil)
	if b.Options.PublicKey.UserVerification != "required" || len(b.Options.PublicKey.AllowCredentials) != 0 || b.Options.PublicKey.RPID != "localhost" {
		t.Fatal("unsafe discoverable options")
	}
	return passkeyRequest(t, h.LoginFinish, nil, map[string]interface{}{"flow_id": b.FlowID, "credential": a.assertion(t, b, passkeyTestOrigin, "localhost", a.handle, flags, count)})
}

func TestPasskeySignedRegistrationLoginReplayAndCounter(t *testing.T) {
	h, user, _ := newPasskeyTest(t)
	a, _ := registerPasskey(t, h, user, 0x05)
	b := passkeyBegin(t, h, nil)
	payload := map[string]interface{}{"flow_id": b.FlowID, "credential": a.assertion(t, b, passkeyTestOrigin, "localhost", a.handle, 0x05, 1)}
	w := passkeyRequest(t, h.LoginFinish, nil, payload)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"token"`) || len(w.Result().Cookies()) != 1 {
		t.Fatalf("login: %d %s", w.Code, w.Body.String())
	}
	var result struct {
		Token string `json:"token"`
	}
	json.Unmarshal(w.Body.Bytes(), &result)
	claims, err := h.auth.ValidateToken(result.Token)
	if err != nil || claims.UserID != user.ID {
		t.Fatal("wrong login user", err)
	}
	if replay := passkeyRequest(t, h.LoginFinish, nil, payload); replay.Code != 401 {
		t.Fatal("replay accepted")
	}
	if rollback := loginPasskey(t, h, a, 0x05, 1); rollback.Code != 401 {
		t.Fatal("counter rollback accepted")
	}
	if good := loginPasskey(t, h, a, 0x05, 2); good.Code != 200 {
		t.Fatal("valid next counter rejected", good.Body.String())
	}
	keys, err := h.auth.db.ListPasskeys(user.ID)
	if err != nil || len(keys) != 1 || keys[0].LastUsedAt == nil {
		t.Fatal("missing last-used metadata")
	}
	var stored webauthn.Credential
	if json.Unmarshal([]byte(keys[0].CredentialJSON), &stored) != nil || stored.Authenticator.SignCount != 2 || len(stored.Transport) != 1 {
		t.Fatal("credential metadata not persisted")
	}
}

func TestPasskeySyncedZeroCounterBypassesTOTPButPasswordRequiresIt(t *testing.T) {
	h, user, _ := newPasskeyTest(t)
	a, _ := registerPasskey(t, h, user, 0x1d) // UP, UV, BE, BS
	for i := 0; i < 2; i++ {
		if w := loginPasskey(t, h, a, 0x1d, 0); w.Code != 200 {
			t.Fatal("synced zero counter rejected", w.Body.String())
		}
	}
	if err := h.auth.db.UpdateUserTwoFactor(user.ID, true, "secret", ""); err != nil {
		t.Fatal(err)
	}
	if unverified := loginPasskey(t, h, a, 0x19, 0); unverified.Code != http.StatusUnauthorized {
		t.Fatal("passkey without user verification bypassed 2FA", unverified.Body.String())
	}
	w := loginPasskey(t, h, a, 0x1d, 0)
	var response struct {
		Required  bool   `json:"two_factor_required"`
		TempToken string `json:"temp_token"`
		Token     string `json:"token"`
	}
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &response) != nil || response.Required || response.TempToken != "" || response.Token == "" || len(w.Result().Cookies()) != 1 {
		t.Fatal("verified passkey did not complete sign-in", w.Body.String())
	}
	if claims, err := h.auth.ValidateToken(response.Token); err != nil || claims.UserID != user.ID {
		t.Fatal("passkey did not issue the correct full session", err)
	}
	password := passkeyRequest(t, h.auth.Login, nil, map[string]string{"username": user.Username, "password": "correct-password"})
	response = struct {
		Required  bool   `json:"two_factor_required"`
		TempToken string `json:"temp_token"`
		Token     string `json:"token"`
	}{}
	if password.Code != http.StatusOK || json.Unmarshal(password.Body.Bytes(), &response) != nil || !response.Required || response.TempToken == "" || response.Token != "" || len(password.Result().Cookies()) != 0 {
		t.Fatal("password sign-in bypassed 2FA", password.Body.String())
	}
	if _, err := h.auth.ValidateToken(response.TempToken); err == nil {
		t.Fatal("2FA token grants a full session")
	}
	if _, err := h.auth.Validate2FATempToken(response.TempToken); err != nil {
		t.Fatal(err)
	}
	fresh, err := h.auth.db.GetUserByID(user.ID)
	if err != nil || !fresh.TwoFactorEnabled {
		t.Fatal("passkey sign-in changed the account's 2FA setting", err)
	}
}

func TestPasskeyRegistrationValidationAndOwnership(t *testing.T) {
	for _, name := range []string{"no_uv", "invalid_backup_state", "wrong_origin", "origin_path", "wrong_rp", "wrong_challenge", "other_user", "password_changed", "expired", "wrong_action", "deleted_user"} {
		t.Run(name, func(t *testing.T) {
			h, user, other := newPasskeyTest(t)
			b := passkeyBegin(t, h, user)
			a := newPasskeyAuthenticator(t)
			origin, rp, flags := passkeyTestOrigin, "localhost", byte(0x05)
			fixture := b
			finishUser := user
			switch name {
			case "no_uv":
				flags = 0x01
			case "invalid_backup_state":
				flags = 0x15
			case "wrong_origin":
				origin = "http://localhost:8124"
			case "origin_path":
				origin = passkeyTestOrigin + "/evil"
			case "wrong_rp":
				rp = "other.test"
			case "wrong_challenge":
				fixture.Options.PublicKey.Challenge = "wrong"
			case "other_user":
				finishUser = other
			case "password_changed":
				if err := h.auth.db.AdminUpdateUser(user.ID, user.Username, "", "", "", "user", "new-hash"); err != nil {
					t.Fatal(err)
				}
			case "expired":
				f := h.flows[b.FlowID]
				f.expires = time.Now().Add(-time.Second)
				h.flows[b.FlowID] = f
			case "wrong_action":
				f := h.flows[b.FlowID]
				f.action = "login"
				h.flows[b.FlowID] = f
			case "deleted_user":
				if err := h.auth.db.DeleteUser(user.ID); err != nil {
					t.Fatal(err)
				}
			}
			w := passkeyRequest(t, h.RegisterFinish, finishUser, map[string]interface{}{"flow_id": b.FlowID, "credential": a.registration(t, fixture, origin, rp, flags)})
			if w.Code == 200 {
				t.Fatal("invalid registration accepted")
			}
			keys, err := h.auth.db.ListPasskeys(user.ID)
			if err != nil || len(keys) != 0 {
				t.Fatal("invalid credential persisted")
			}
			if name != "deleted_user" {
				if _, ok := h.flows[b.FlowID]; ok {
					t.Fatal("failed verification did not consume flow")
				}
			}
		})
	}
}

func TestPasskeySignedAssertionFailures(t *testing.T) {
	h, user, other := newPasskeyTest(t)
	a, _ := registerPasskey(t, h, user, 0x05)
	otherHandle, err := h.auth.db.GetOrCreateWebAuthnHandle(other.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"no_uv", "backup_eligibility_changed", "invalid_backup_state", "wrong_origin", "wrong_rp", "wrong_challenge", "wrong_handle", "missing_handle", "wrong_signature", "cross_origin", "wrong_type", "expired"} {
		t.Run(name, func(t *testing.T) {
			b := passkeyBegin(t, h, nil)
			fixture := b
			origin, rp, handle, flags := passkeyTestOrigin, "localhost", a.handle, byte(0x05)
			switch name {
			case "no_uv":
				flags = 0x01
			case "backup_eligibility_changed":
				flags = 0x0d
			case "invalid_backup_state":
				flags = 0x15
			case "wrong_origin":
				origin = "http://localhost:8124"
			case "wrong_rp":
				rp = "other.test"
			case "wrong_challenge":
				fixture.Options.PublicKey.Challenge = "wrong"
			case "wrong_handle":
				handle = passkeyB64(otherHandle)
			case "missing_handle":
				handle = ""
			case "expired":
				f := h.flows[b.FlowID]
				f.expires = time.Now().Add(-time.Second)
				h.flows[b.FlowID] = f
			}
			credential := a.assertion(t, fixture, origin, rp, handle, flags, 1)
			response := credential["response"].(map[string]interface{})
			if name == "wrong_signature" {
				response["signature"] = passkeyB64([]byte("invalid"))
			}
			if name == "cross_origin" || name == "wrong_type" {
				typ := "webauthn.get"
				if name == "wrong_type" {
					typ = "webauthn.create"
				}
				response["clientDataJSON"] = passkeyB64(passkeyJSON(t, map[string]interface{}{"type": typ, "challenge": fixture.Options.PublicKey.Challenge, "origin": origin, "crossOrigin": name == "cross_origin"}))
			}
			w := passkeyRequest(t, h.LoginFinish, nil, map[string]interface{}{"flow_id": b.FlowID, "credential": credential})
			if w.Code != 401 || !strings.Contains(w.Body.String(), errPasskeyVerification.Error()) || len(w.Result().Cookies()) != 0 {
				t.Fatalf("invalid assertion: %d %s", w.Code, w.Body.String())
			}
		})
	}
	if w := loginPasskey(t, h, a, 0x05, 1); w.Code != 200 {
		t.Fatal("failures advanced counter", w.Body.String())
	}
	if err := h.auth.db.DeleteUser(user.ID); err != nil {
		t.Fatal(err)
	}
	if w := loginPasskey(t, h, a, 0x05, 2); w.Code != 401 {
		t.Fatal("deleted account signed in")
	}
}

func TestPasskeyPasswordProofRemovalCollisionAndMaintenance(t *testing.T) {
	h, user, other := newPasskeyTest(t)
	for _, password := range []string{"", "wrong"} {
		if w := passkeyRequest(t, h.RegisterBegin, user, map[string]string{"password": password}); w.Code != 401 {
			t.Fatal("password proof not required")
		}
	}
	if len(h.flows) != 0 {
		t.Fatal("bad proof allocated flow")
	}
	a, id := registerPasskey(t, h, user, 0x05)
	b := passkeyBegin(t, h, other)
	if w := passkeyRequest(t, h.RegisterFinish, other, map[string]interface{}{"flow_id": b.FlowID, "credential": a.registration(t, b, passkeyTestOrigin, "localhost", 0x05)}); w.Code == 200 {
		t.Fatal("credential reassigned")
	}
	// Restore authenticator's original handle after constructing collision fixture.
	handle, _ := h.auth.db.GetOrCreateWebAuthnHandle(user.ID)
	a.handle = passkeyB64(handle)
	remove := func(user *database.User, password string) *httptest.ResponseRecorder {
		return passkeyRequest(t, func(w http.ResponseWriter, r *http.Request) { r.SetPathValue("passkeyId", id); h.Delete(w, r) }, user, map[string]string{"password": password})
	}
	if w := remove(other, "correct-password"); w.Code != 404 {
		t.Fatal("removed another account key")
	}
	if w := remove(user, "wrong"); w.Code != 401 {
		t.Fatal("removed without proof")
	}
	if err := h.auth.db.SetPanelSetting("branding_json", `{"maintenance_enabled":true}`); err != nil {
		t.Fatal(err)
	}
	if w := loginPasskey(t, h, a, 0x05, 1); w.Code != 503 {
		t.Fatal("maintenance bypassed", w.Body.String())
	}
	if w := remove(user, "correct-password"); w.Code != 200 {
		t.Fatal("own removal failed", w.Body.String())
	}
	if w := loginPasskey(t, h, a, 0x05, 2); w.Code != 401 {
		t.Fatal("removed credential signed in")
	}
}

func TestPasskeyOriginModeCSRFAndTurnstileGuards(t *testing.T) {
	h, user, _ := newPasskeyTest(t)
	for _, raw := range []string{"http://panel.example.com", "https://user:pass@panel.test", "https://panel.test/path", "https://panel.test?query", "https://*.test", "null", "https://panel.test:99999"} {
		if _, _, err := canonicalPasskeyOrigin(raw); err == nil {
			t.Errorf("invalid origin accepted: %s", raw)
		}
	}
	for _, raw := range []string{passkeyTestOrigin, "https://panel.example.com", "https://panel.example.com:8443"} {
		if _, _, err := canonicalPasskeyOrigin(raw); err != nil {
			t.Errorf("valid origin rejected: %s: %v", raw, err)
		}
	}
	for _, name := range []string{"no_origin", "cross_origin", "cross_site", "wrong_type", "insecure", "node", "disabled", "turnstile", "oversize", "trailing_json", "configured_host"} {
		t.Run(name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.Server.Mode = "standalone"
			cfg.Server.PublicURL = passkeyTestOrigin
			h.auth.cfg = cfg
			body := "{}"
			if name == "oversize" {
				body = `{"x":"` + strings.Repeat("a", maxPasskeyBody) + `"}`
			}
			if name == "trailing_json" {
				body = "{} {}"
			}
			req := httptest.NewRequest("POST", passkeyTestOrigin+"/api/auth/passkeys/login/begin", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Origin", passkeyTestOrigin)
			switch name {
			case "no_origin":
				req.Header.Del("Origin")
			case "cross_origin":
				req.Header.Set("Origin", "https://evil.test")
			case "cross_site":
				req.Header.Set("Sec-Fetch-Site", "cross-site")
			case "wrong_type":
				req.Header.Set("Content-Type", "text/plain")
			case "insecure":
				cfg.Server.PublicURL = "http://panel.test"
				req.Host = "panel.test"
				req.Header.Set("Origin", "http://panel.test")
				req.Header.Set("X-Forwarded-Proto", "https")
			case "node":
				cfg.Server.Role = "node"
			case "disabled":
				cfg.Security.PasskeysDisabled = true
			case "turnstile":
				cfg.Security.Turnstile = config.TurnstileConfig{Enabled: true, SiteKey: "site", SecretKey: "secret"}
			case "configured_host":
				cfg.Server.PublicURL = "https://panel.test"
				cfg.Server.Mode = "behind_proxy"
			}
			w := httptest.NewRecorder()
			h.LoginBegin(w, req)
			if w.Code == 200 {
				t.Fatal("guard bypassed")
			}
		})
	}
	h.auth.cfg = config.DefaultConfig()
	h.auth.cfg.Server.Mode = "standalone"
	h.auth.cfg.Server.PublicURL = passkeyTestOrigin
	req := httptest.NewRequest("GET", passkeyTestOrigin+"/api/account/passkeys", nil)
	req = req.WithContext(context.WithValue(req.Context(), contextUserKey, &JWTClaims{UserID: user.ID}))
	h.auth.cfg.Security.PasskeysDisabled = true
	w := httptest.NewRecorder()
	h.List(w, req)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"available":false`) || !strings.Contains(w.Body.String(), `"reason"`) {
		t.Fatal("disabled availability missing", w.Body.String())
	}
	h.auth.cfg.Security.PasskeysDisabled = false
	registerPasskey(t, h, user, 0x05)
	req = httptest.NewRequest("GET", "https://different.test/api/account/passkeys", nil)
	if _, _, err := h.relyingParty(req); err == nil {
		t.Fatal("pinned RP changed")
	}
	// Behind-proxy mode uses configured policy, never attacker-supplied forwarded host.
	h2, _, _ := newPasskeyTest(t)
	h2.auth.cfg.Server.Mode = "behind_proxy"
	h2.auth.cfg.Server.PublicURL = "https://panel.test"
	req = httptest.NewRequest("POST", "http://panel.test/api/test", nil)
	req.Header.Set("X-Forwarded-Host", "evil.test")
	req.Header.Set("X-Forwarded-Proto", "http")
	wa, origin, err := h2.relyingParty(req)
	if err != nil || origin != "https://panel.test" || wa.Config.RPID != "panel.test" {
		t.Fatal("proxy origin handling", err)
	}
}

func TestPasskeyFlowsBoundedAtomicAndBound(t *testing.T) {
	h, _, _ := newPasskeyTest(t)
	base := passkeyFlow{action: "register", userID: 7, origin: passkeyTestOrigin, rpID: "localhost"}
	for _, name := range []string{"action", "user", "origin", "rp"} {
		id, err := h.putFlow(base)
		if err != nil {
			t.Fatal(err)
		}
		action, user, origin, rp := base.action, base.userID, base.origin, base.rpID
		switch name {
		case "action":
			action = "login"
		case "user":
			user = 8
		case "origin":
			origin = "http://localhost:8124"
		case "rp":
			rp = "other.test"
		}
		if _, err := h.consumeFlow(id, action, user, origin, rp); err == nil {
			t.Fatalf("%s not bound", name)
		}
		if _, err := h.consumeFlow(id, base.action, base.userID, base.origin, base.rpID); err == nil {
			t.Fatal("failed attempt reusable")
		}
	}
	id, err := h.putFlow(base)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	var successes atomic.Int32
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := h.consumeFlow(id, base.action, base.userID, base.origin, base.rpID); err == nil {
				successes.Add(1)
			}
		}()
	}
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatalf("flow consumed %d times", successes.Load())
	}
	for i := 0; i < maxPasskeyFlows; i++ {
		h.flows[fmt.Sprint(i)] = passkeyFlow{expires: time.Now().Add(time.Minute)}
	}
	if _, err := h.putFlow(base); err == nil {
		t.Fatal("unbounded flow store")
	}
	h.flows["0"] = passkeyFlow{expires: time.Now().Add(-time.Second)}
	if _, err := h.putFlow(base); err != nil {
		t.Fatal("expired flow not reclaimed")
	}
}

func TestPasskeyAuthRateBucket(t *testing.T) {
	lim := NewIPRateLimiter(1, 15)
	for i, path := range []string{"/api/auth/passkeys/login/begin", "/api/auth/passkeys/login/finish", "/api/auth/passkeys/login/begin", "/api/auth/passkeys/login/finish", "/api/auth/passkeys/login/begin", "/api/auth/passkeys/login/finish"} {
		req := httptest.NewRequest("POST", path, nil)
		req.RemoteAddr = "198.51.100.10:3456"
		if allowed := lim.Allow(req); allowed != (i < 5) {
			t.Fatalf("request %d used wrong bucket", i)
		}
	}
}

func TestPasskeysRequireConfiguredOriginBeforeFirstRegistration(t *testing.T) {
	h, user, _ := newPasskeyTest(t)
	h.auth.cfg.Server.PublicURL = ""
	for _, origin := range []string{"https://attacker.example", passkeyTestOrigin} {
		body := strings.NewReader(`{"label":"Untrusted first key","password":"correct-password"}`)
		req := httptest.NewRequest(http.MethodPost, origin+"/api/account/passkeys/register/begin", body)
		req.Header.Set("Origin", origin)
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), contextUserKey, &JWTClaims{UserID: user.ID, Username: user.Username, Role: user.Role}))
		response := httptest.NewRecorder()
		h.RegisterBegin(response, req)
		if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "Public panel URL") {
			t.Fatalf("unconfigured origin %s accepted: %d %s", origin, response.Code, response.Body.String())
		}
	}
	if pinned, err := h.auth.db.PasskeyRPID(); err != nil || pinned != "" {
		t.Fatalf("untrusted request pinned RP: %q %v", pinned, err)
	}
}
