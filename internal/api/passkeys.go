package api

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

const passkeyFlowTTL = 5 * time.Minute
const maxPasskeyFlows = 1024
const maxPasskeyBody = 64 << 10

var errPasskeyVerification = errors.New("Passkey verification failed; please start again")

// Flows are deliberately process-local: restarting safely invalidates all pending
// ceremonies. The bounded map holds no private keys; consumption is atomic and
// precedes verification, including failed, wrong-action and wrong-user attempts.
type passkeyFlow struct {
	action       string
	userID       int
	passwordHash string
	label        string
	origin       string
	rpID         string
	expires      time.Time
	session      webauthn.SessionData
}

type PasskeyHandler struct {
	auth  *AuthHandler
	mu    sync.Mutex
	flows map[string]passkeyFlow
}

func NewPasskeyHandler(auth *AuthHandler) *PasskeyHandler {
	return &PasskeyHandler{auth: auth, flows: make(map[string]passkeyFlow)}
}

func (h *PasskeyHandler) putFlow(f passkeyFlow) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	id := base64.RawURLEncoding.EncodeToString(b)
	h.mu.Lock()
	defer h.mu.Unlock()
	now := time.Now()
	for k, v := range h.flows {
		if !now.Before(v.expires) {
			delete(h.flows, k)
		}
	}
	if len(h.flows) >= maxPasskeyFlows {
		return "", errors.New("Too many pending passkey requests; try again shortly")
	}
	f.expires = now.Add(passkeyFlowTTL)
	f.session.Expires = f.expires
	h.flows[id] = f
	return id, nil
}

func (h *PasskeyHandler) consumeFlow(id, action string, userID int, origin, rpID string) (passkeyFlow, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	f, ok := h.flows[id]
	delete(h.flows, id)
	if !ok || !time.Now().Before(f.expires) || f.action != action || f.userID != userID || f.origin != origin || f.rpID != rpID {
		return passkeyFlow{}, errPasskeyVerification
	}
	return f, nil
}

type passkeyUser struct {
	user        *database.User
	handle      []byte
	credentials []webauthn.Credential
}

func (u *passkeyUser) WebAuthnID() []byte                         { return u.handle }
func (u *passkeyUser) WebAuthnName() string                       { return u.user.Username }
func (u *passkeyUser) WebAuthnDisplayName() string                { return u.user.Username }
func (u *passkeyUser) WebAuthnIcon() string                       { return "" }
func (u *passkeyUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }

func (h *PasskeyHandler) loadUser(id int, rpID string) (*passkeyUser, []database.Passkey, error) {
	user, err := h.auth.db.GetUserByID(id)
	if err != nil {
		return nil, nil, err
	}
	handle, err := h.auth.db.GetOrCreateWebAuthnHandle(id)
	if err != nil {
		return nil, nil, err
	}
	keys, err := h.auth.db.ListPasskeys(id)
	if err != nil {
		return nil, nil, err
	}
	u := &passkeyUser{user: user, handle: handle}
	for _, p := range keys {
		if p.RPID != rpID {
			return nil, nil, errPasskeyVerification
		}
		var c webauthn.Credential
		if err := json.Unmarshal([]byte(p.CredentialJSON), &c); err != nil || !bytes.Equal(c.ID, p.CredentialID) {
			return nil, nil, errPasskeyVerification
		}
		u.credentials = append(u.credentials, c)
	}
	return u, keys, nil
}

// canonicalPasskeyOrigin accepts only a serialized origin, not a URL path,
// credentials, fragment, wildcard or opaque origin. No forwarded header is used.
func canonicalPasskeyOrigin(raw string) (string, string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.User != nil || u.Opaque != "" || u.Host == "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || (u.Path != "" && u.Path != "/") || u.RawPath != "" {
		return "", "", errors.New("Configure server.public_url as the exact panel origin without a path")
	}
	host := strings.ToLower(u.Hostname())
	if host == "" || strings.HasSuffix(host, ".") {
		return "", "", errors.New("Invalid passkey hostname")
	}
	local := host == "localhost" || host == "127.0.0.1" || host == "::1"
	if u.Scheme != "https" && !(u.Scheme == "http" && local) {
		return "", "", errors.New("Passkeys require HTTPS (HTTP is allowed only on localhost)")
	}
	if ip := net.ParseIP(host); ip != nil && !local {
		return "", "", errors.New("Passkeys require a stable panel hostname")
	}
	if net.ParseIP(host) == nil {
		for _, label := range strings.Split(host, ".") {
			if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
				return "", "", errors.New("Invalid passkey hostname")
			}
			for _, c := range label {
				if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
					return "", "", errors.New("Invalid passkey hostname")
				}
			}
		}
	}
	port := u.Port()
	if port != "" {
		n, e := strconv.Atoi(port)
		if e != nil || n < 1 || n > 65535 {
			return "", "", errors.New("Invalid passkey port")
		}
	}
	if (u.Scheme == "https" && port == "443") || (u.Scheme == "http" && port == "80") {
		port = ""
	}
	authority := host
	if strings.Contains(host, ":") {
		authority = "[" + host + "]"
	}
	if port != "" {
		authority = net.JoinHostPort(host, port)
	}
	return u.Scheme + "://" + authority, host, nil
}

func (h *PasskeyHandler) relyingParty(r *http.Request) (*webauthn.WebAuthn, string, error) {
	h.auth.mu.RLock()
	cfg := h.auth.cfg
	publicURL, mode := "", ""
	if cfg != nil {
		if cfg.IsNode() {
			h.auth.mu.RUnlock()
			return nil, "", errors.New("Passkeys are managed on the Master Panel, not worker nodes")
		}
		if cfg.Security.PasskeysDisabled {
			h.auth.mu.RUnlock()
			return nil, "", errors.New("Passkeys are disabled by the administrator")
		}
		publicURL, mode = cfg.Server.PublicURL, cfg.Server.Mode
	}
	h.auth.mu.RUnlock()
	// Request headers are not trusted configuration: an ordinary user must not
	// choose the panel-wide RP hostname by making the first registration.
	if strings.TrimSpace(publicURL) == "" {
		return nil, "", errors.New("Ask an administrator to configure server.public_url before using passkeys")
	}
	requestHost, err := url.Parse("http://" + r.Host)
	if err != nil || requestHost.Host != r.Host || requestHost.User != nil {
		return nil, "", errors.New("Invalid panel host")
	}
	origin, rpID, err := canonicalPasskeyOrigin(strings.TrimSpace(publicURL))
	if err != nil {
		return nil, "", err
	}
	// A configured canonical URL must not make a request on an unrelated Host valid.
	expected, _ := url.Parse(origin)
	hostOrigin, _, err := canonicalPasskeyOrigin(expected.Scheme + "://" + r.Host)
	if err != nil || hostOrigin != origin {
		return nil, "", errors.New("Use the configured public panel URL for passkeys")
	}
	if expected.Scheme == "https" && r.TLS == nil && mode != "behind_proxy" {
		return nil, "", errors.New("Passkeys require HTTPS or an explicitly configured HTTPS reverse proxy")
	}
	pinned, err := h.auth.db.PasskeyRPID()
	if err != nil {
		return nil, "", errors.New("Passkey storage is unavailable")
	}
	if pinned != "" && pinned != rpID {
		return nil, "", errors.New("Panel hostname changed; use the original passkey panel hostname")
	}
	resident := true
	wa, err := webauthn.New(&webauthn.Config{
		RPID: rpID, RPDisplayName: "NeoPanel", RPOrigins: []string{origin}, AttestationPreference: protocol.PreferNoAttestation,
		AuthenticatorSelection: protocol.AuthenticatorSelection{ResidentKey: protocol.ResidentKeyRequirementRequired, RequireResidentKey: &resident, UserVerification: protocol.VerificationRequired},
		Timeouts:               webauthn.TimeoutsConfig{Login: webauthn.TimeoutConfig{Enforce: true, Timeout: passkeyFlowTTL}, Registration: webauthn.TimeoutConfig{Enforce: true, Timeout: passkeyFlowTTL}},
	})
	if err != nil {
		return nil, "", errors.New("Passkey relying-party configuration is invalid")
	}
	return wa, origin, nil
}

func passkeyError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

// Mutations require JSON plus an exact same-origin browser Origin, including
// port. Requiring Origin also prevents cookie-authenticated login/registration CSRF.
func (h *PasskeyHandler) ceremony(w http.ResponseWriter, r *http.Request) (*webauthn.WebAuthn, string, bool) {
	w.Header().Set("Cache-Control", "no-store")
	wa, origin, err := h.relyingParty(r)
	if err != nil {
		passkeyError(w, http.StatusServiceUnavailable, err)
		return nil, "", false
	}
	if r.Header.Get("Origin") != origin || (r.Header.Get("Sec-Fetch-Site") != "" && r.Header.Get("Sec-Fetch-Site") != "same-origin") {
		passkeyError(w, http.StatusForbidden, errors.New("Passkeys require a same-origin request"))
		return nil, "", false
	}
	typ, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || typ != "application/json" {
		passkeyError(w, http.StatusUnsupportedMediaType, errors.New("JSON request required"))
		return nil, "", false
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxPasskeyBody)
	return wa, origin, true
}

func decodePasskeyBody(r *http.Request, v interface{}) error {
	d := json.NewDecoder(r.Body)
	if err := d.Decode(v); err != nil {
		return err
	}
	var extra interface{}
	if err := d.Decode(&extra); err != io.EOF {
		return errors.New("Invalid request body")
	}
	return nil
}

func (h *PasskeyHandler) currentUser(r *http.Request) (*database.User, error) {
	c, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if c == nil || c.UserID <= 0 || c.Scope != "" {
		return nil, errPasskeyVerification
	}
	return h.auth.db.GetUserByID(c.UserID)
}

func (h *PasskeyHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	user, err := h.currentUser(r)
	if err != nil {
		passkeyError(w, http.StatusUnauthorized, errPasskeyVerification)
		return
	}
	keys, err := h.auth.db.ListPasskeys(user.ID)
	if err != nil {
		passkeyError(w, http.StatusInternalServerError, errors.New("Passkey storage is unavailable"))
		return
	}
	wa, _, err := h.relyingParty(r)
	result := map[string]interface{}{"passkeys": keys, "available": err == nil, "rp_id": ""}
	if err != nil {
		result["reason"] = err.Error()
	} else {
		result["rp_id"] = wa.Config.RPID
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *PasskeyHandler) RegisterBegin(w http.ResponseWriter, r *http.Request) {
	wa, origin, ok := h.ceremony(w, r)
	if !ok {
		return
	}
	var req struct {
		Label    string `json:"label"`
		Password string `json:"password"`
	}
	if decodePasskeyBody(r, &req) != nil {
		passkeyError(w, http.StatusBadRequest, errors.New("Invalid request body"))
		return
	}
	user, err := h.currentUser(r)
	if err != nil || len(req.Password) > 1024 || !h.auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		passkeyError(w, http.StatusUnauthorized, errors.New("Current password verification failed"))
		return
	}
	req.Label = strings.TrimSpace(req.Label)
	if req.Label == "" {
		req.Label = "Passkey"
	}
	if len(req.Label) > 80 {
		passkeyError(w, http.StatusBadRequest, errors.New("Passkey label must be at most 80 bytes"))
		return
	}
	u, keys, err := h.loadUser(user.ID, wa.Config.RPID)
	if err != nil {
		passkeyError(w, http.StatusInternalServerError, errPasskeyVerification)
		return
	}
	if len(keys) >= database.MaxPasskeysPerUser {
		passkeyError(w, http.StatusBadRequest, errors.New("Maximum of 10 passkeys per account reached"))
		return
	}
	exclusions := make([]protocol.CredentialDescriptor, 0, len(u.credentials))
	for _, c := range u.credentials {
		exclusions = append(exclusions, c.Descriptor())
	}
	options, session, err := wa.BeginRegistration(u, webauthn.WithExclusions(exclusions))
	if err != nil {
		passkeyError(w, http.StatusInternalServerError, errPasskeyVerification)
		return
	}
	flow, err := h.putFlow(passkeyFlow{action: "register", userID: user.ID, passwordHash: user.PasswordHash, label: req.Label, origin: origin, rpID: wa.Config.RPID, session: *session})
	if err != nil {
		passkeyError(w, http.StatusTooManyRequests, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"flow_id": flow, "options": options})
}

type passkeyFinishRequest struct {
	FlowID     string          `json:"flow_id"`
	Credential json.RawMessage `json:"credential"`
}

// The compatible library predates crossOrigin/topOrigin checks and normalizes
// client origins. Enforce exact origin and top-level-only ceremonies explicitly,
// while delegating all authenticator parsing and cryptographic checks to it.
func passkeyClientOrigin(credential json.RawMessage, origin string) bool {
	var c struct {
		Response struct {
			ClientData string `json:"clientDataJSON"`
		} `json:"response"`
	}
	if json.Unmarshal(credential, &c) != nil {
		return false
	}
	data, err := base64.RawURLEncoding.DecodeString(c.Response.ClientData)
	if err != nil {
		return false
	}
	var client struct {
		Origin      string `json:"origin"`
		CrossOrigin bool   `json:"crossOrigin"`
		TopOrigin   string `json:"topOrigin"`
	}
	return json.Unmarshal(data, &client) == nil && client.Origin == origin && !client.CrossOrigin && client.TopOrigin == ""
}

func (h *PasskeyHandler) RegisterFinish(w http.ResponseWriter, r *http.Request) {
	wa, origin, ok := h.ceremony(w, r)
	if !ok {
		return
	}
	var req passkeyFinishRequest
	if decodePasskeyBody(r, &req) != nil {
		passkeyError(w, http.StatusBadRequest, errPasskeyVerification)
		return
	}
	user, err := h.currentUser(r)
	if err != nil {
		passkeyError(w, http.StatusUnauthorized, errPasskeyVerification)
		return
	}
	f, err := h.consumeFlow(req.FlowID, "register", user.ID, origin, wa.Config.RPID)
	if err != nil || f.passwordHash != user.PasswordHash || !passkeyClientOrigin(req.Credential, origin) {
		passkeyError(w, http.StatusBadRequest, errPasskeyVerification)
		return
	}
	u, _, err := h.loadUser(user.ID, wa.Config.RPID)
	if err != nil {
		passkeyError(w, http.StatusBadRequest, errPasskeyVerification)
		return
	}
	parsed, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(req.Credential))
	if err != nil {
		passkeyError(w, http.StatusBadRequest, errPasskeyVerification)
		return
	}
	credential, err := wa.CreateCredential(u, f.session, parsed)
	if err != nil || !bytes.Equal(parsed.RawID, credential.ID) || (credential.Flags.BackupState && !credential.Flags.BackupEligible) {
		passkeyError(w, http.StatusBadRequest, errPasskeyVerification)
		return
	}
	data, err := json.Marshal(credential)
	if err != nil {
		passkeyError(w, http.StatusBadRequest, errPasskeyVerification)
		return
	}
	p, err := h.auth.db.AddPasskey(user.ID, f.passwordHash, wa.Config.RPID, f.label, credential.ID, string(data))
	if err != nil {
		passkeyError(w, http.StatusBadRequest, errPasskeyVerification)
		return
	}
	logAuditUser(h.auth.db, user.ID, r, "account_passkey_add", "Registered a passkey")
	writeJSON(w, http.StatusOK, map[string]interface{}{"message": "Passkey registered", "passkey": p})
}

func (h *PasskeyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	_, _, ok := h.ceremony(w, r)
	if !ok {
		return
	}
	var req struct {
		Password string `json:"password"`
	}
	if decodePasskeyBody(r, &req) != nil {
		passkeyError(w, http.StatusBadRequest, errPasskeyVerification)
		return
	}
	user, err := h.currentUser(r)
	if err != nil || len(req.Password) > 1024 || !h.auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		passkeyError(w, http.StatusUnauthorized, errors.New("Current password verification failed"))
		return
	}
	if h.auth.db.DeletePasskey(user.ID, getURLParam(r, "passkeyId"), user.PasswordHash) != nil {
		passkeyError(w, http.StatusNotFound, errors.New("Passkey not found"))
		return
	}
	logAuditUser(h.auth.db, user.ID, r, "account_passkey_delete", "Removed a passkey")
	writeJSON(w, http.StatusOK, map[string]string{"message": "Passkey removed"})
}

func (h *PasskeyHandler) LoginBegin(w http.ResponseWriter, r *http.Request) {
	wa, origin, ok := h.ceremony(w, r)
	if !ok {
		return
	}
	var req struct {
		TurnstileToken string `json:"turnstile_token"`
		TurnstileResp  string `json:"cf-turnstile-response"`
	}
	if decodePasskeyBody(r, &req) != nil {
		passkeyError(w, http.StatusBadRequest, errors.New("Invalid request body"))
		return
	}
	if active, _, secret := h.auth.isTurnstileActive(); active {
		token := req.TurnstileToken
		if token == "" {
			token = req.TurnstileResp
		}
		if err := VerifyTurnstileToken(secret, token, r); err != nil {
			passkeyError(w, http.StatusBadRequest, err)
			return
		}
	}
	options, session, err := wa.BeginDiscoverableLogin(webauthn.WithUserVerification(protocol.VerificationRequired))
	if err != nil {
		passkeyError(w, http.StatusInternalServerError, errPasskeyVerification)
		return
	}
	flow, err := h.putFlow(passkeyFlow{action: "login", origin: origin, rpID: wa.Config.RPID, session: *session})
	if err != nil {
		passkeyError(w, http.StatusTooManyRequests, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"flow_id": flow, "options": options})
}

func (h *PasskeyHandler) LoginFinish(w http.ResponseWriter, r *http.Request) {
	wa, origin, ok := h.ceremony(w, r)
	if !ok {
		return
	}
	var req passkeyFinishRequest
	if decodePasskeyBody(r, &req) != nil {
		passkeyError(w, http.StatusBadRequest, errPasskeyVerification)
		return
	}
	f, err := h.consumeFlow(req.FlowID, "login", 0, origin, wa.Config.RPID)
	if err != nil || !passkeyClientOrigin(req.Credential, origin) {
		passkeyError(w, http.StatusUnauthorized, errPasskeyVerification)
		return
	}
	parsed, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(req.Credential))
	if err != nil {
		passkeyError(w, http.StatusUnauthorized, errPasskeyVerification)
		return
	}
	var matched database.Passkey
	var owner *passkeyUser
	credential, err := wa.ValidateDiscoverableLogin(func(rawID, handle []byte) (webauthn.User, error) {
		if len(handle) != 32 || len(rawID) == 0 || len(rawID) > 1024 {
			return nil, errPasskeyVerification
		}
		id, e := h.auth.db.PasskeyUserByHandle(handle)
		if e != nil {
			return nil, errPasskeyVerification
		}
		u, keys, e := h.loadUser(id, wa.Config.RPID)
		if e != nil || !bytes.Equal(u.handle, handle) {
			return nil, errPasskeyVerification
		}
		for _, p := range keys {
			if bytes.Equal(p.CredentialID, rawID) {
				matched = p
				owner = u
				return u, nil
			}
		}
		return nil, errPasskeyVerification
	}, f.session, parsed)
	if err != nil || owner == nil || credential == nil || !credential.Flags.UserVerified || credential.Authenticator.CloneWarning {
		passkeyError(w, http.StatusUnauthorized, errPasskeyVerification)
		return
	}
	// Backup eligibility is immutable; a backup state is valid only for an
	// eligible credential. The compatible library does not enforce this policy.
	var previous webauthn.Credential
	if json.Unmarshal([]byte(matched.CredentialJSON), &previous) != nil ||
		previous.Flags.BackupEligible != credential.Flags.BackupEligible ||
		(credential.Flags.BackupState && !credential.Flags.BackupEligible) {
		passkeyError(w, http.StatusUnauthorized, errPasskeyVerification)
		return
	}
	data, err := json.Marshal(credential)
	if err != nil || h.auth.db.UpdatePasskey(matched, string(data)) != nil {
		passkeyError(w, http.StatusUnauthorized, errPasskeyVerification)
		return
	}
	// Recheck the account before issuing a session; do not change its password-login 2FA setting.
	user, err := h.auth.db.GetUserByID(owner.user.ID)
	if err != nil || user.PasswordHash != owner.user.PasswordHash {
		passkeyError(w, http.StatusUnauthorized, errPasskeyVerification)
		return
	}
	if LoadBrandingConfig(h.auth.db).MaintenanceEnabled && user.Role != "admin" {
		passkeyError(w, http.StatusServiceUnavailable, errors.New("Panel is currently under maintenance"))
		return
	}
	// The verified, user-verifying passkey replaces password plus TOTP for this sign-in.
	h.auth.completeVerifiedLogin(w, r, user)
}
