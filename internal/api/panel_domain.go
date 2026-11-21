package api

import (
	"context"
	"crypto/subtle"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"sync"
	"time"

	"github.com/Pekly/NeoPanel/internal/database"
	"github.com/Pekly/NeoPanel/internal/paneldomain"
)

type domainCaddy interface {
	Capability(context.Context) paneldomain.Capability
	Provision(context.Context, string, string) (paneldomain.ProvisionResult, error)
}
type domainChallenge struct {
	proof   string
	expires time.Time
}
type domainTicket struct {
	report      panelDomainReport
	authVersion int64
}
type PanelDomainHandler struct {
	auth          *AuthHandler
	caddy         domainCaddy
	mu            sync.Mutex
	challenges    map[string]domainChallenge
	tickets       map[int]domainTicket
	active        map[int]bool
	nextCheck     map[int]time.Time
	nextSetup     map[int]time.Time
	now           func() time.Time
	lookup        func(context.Context, string, string) ([]netip.Addr, error)
	dial          func(context.Context, string, string) (net.Conn, error)
	roots         *x509.CertPool // nil means system trust; injectable only by Go tests.
	checkInterval time.Duration
	setupInterval time.Duration
}

func NewPanelDomainHandler(auth *AuthHandler) *PanelDomainHandler {
	cfg := auth.cfg
	return &PanelDomainHandler{auth: auth, caddy: paneldomain.NewCaddyManager(cfg.Server.Host, cfg.Server.Port, !cfg.IsNode(), cfg.Server.Mode == "behind_proxy"),
		challenges: make(map[string]domainChallenge), tickets: make(map[int]domainTicket), active: make(map[int]bool), nextCheck: make(map[int]time.Time), nextSetup: make(map[int]time.Time), now: time.Now,
		lookup: net.DefaultResolver.LookupNetIP, dial: (&net.Dialer{Timeout: 5 * time.Second}).DialContext, checkInterval: 3 * time.Second, setupInterval: 30 * time.Second}
}
func decodeDomainBody(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeError(w, 400, "Provide a valid domain settings object")
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		writeError(w, 400, "Provide exactly one domain settings object")
		return false
	}
	return true
}
func (h *PanelDomainHandler) begin(w http.ResponseWriter, r *http.Request, kind string) (*JWTClaims, bool) {
	claims, _ := r.Context().Value(contextUserKey).(*JWTClaims)
	if claims == nil || claims.Role != "admin" || h.auth.cfg.IsNode() {
		writeError(w, 403, "Only panel administrators can change the panel domain")
		return nil, false
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	now := h.now()
	for key, ticket := range h.tickets {
		if !now.Before(ticket.report.ExpiresAt) {
			delete(h.tickets, key)
		}
	}
	for key, next := range h.nextCheck {
		if !now.Before(next) {
			delete(h.nextCheck, key)
		}
	}
	for key, next := range h.nextSetup {
		if !now.Before(next) {
			delete(h.nextSetup, key)
		}
	}
	if h.active[claims.UserID] || len(h.active) >= 4 {
		w.Header().Set("Retry-After", "3")
		writeError(w, 429, "A domain operation is already running. Wait for it to finish before trying again")
		return nil, false
	}
	if (kind == "check" && now.Before(h.nextCheck[claims.UserID])) || (kind == "setup" && now.Before(h.nextSetup[claims.UserID])) {
		w.Header().Set("Retry-After", "3")
		writeError(w, 429, "Please wait before retrying this domain operation")
		return nil, false
	}
	if len(h.tickets) >= 128 {
		for key := range h.tickets {
			delete(h.tickets, key)
			break
		}
	}
	if len(h.nextCheck)+len(h.nextSetup) >= 256 {
		writeError(w, 429, "Domain checks are busy. Wait a minute and retry")
		return nil, false
	}
	h.active[claims.UserID] = true
	if kind == "check" {
		h.nextCheck[claims.UserID] = now.Add(h.checkInterval)
		delete(h.tickets, claims.UserID)
	}
	if kind == "setup" {
		h.nextSetup[claims.UserID] = now.Add(h.setupInterval)
	}
	return claims, true
}
func (h *PanelDomainHandler) finish(userID int) { h.mu.Lock(); delete(h.active, userID); h.mu.Unlock() }
func (h *PanelDomainHandler) ticket(claims *JWTClaims, origin, token string) (domainTicket, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	ticket, ok := h.tickets[claims.UserID]
	return ticket, ok && token != "" && subtle.ConstantTimeCompare([]byte(ticket.report.CheckToken), []byte(token)) == 1 && ticket.report.PublicURL == origin && ticket.authVersion == claims.AuthVersion && h.now().Before(ticket.report.ExpiresAt)
}
func (h *PanelDomainHandler) clearTicket(id int) { h.mu.Lock(); delete(h.tickets, id); h.mu.Unlock() }

// Proof is public but returns only a random proof for an active, unguessable,
// short-lived challenge. It exposes neither settings nor authentication data.
func (h *PanelDomainHandler) Proof(w http.ResponseWriter, r *http.Request) {
	nonce := r.PathValue("nonce")
	h.mu.Lock()
	challenge, ok := h.challenges[nonce]
	h.mu.Unlock()
	w.Header().Set("Cache-Control", "no-store")
	if !ok || !h.now().Before(challenge.expires) {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, 200, map[string]string{"proof": challenge.proof})
}
func (h *PanelDomainHandler) Check(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.begin(w, r, "check")
	if !ok {
		return
	}
	defer h.finish(claims.UserID)
	var req struct {
		PublicURL string `json:"public_url"`
	}
	if !decodeDomainBody(w, r, &req) {
		return
	}
	origin, host, err := normalizeDomainOrigin(req.PublicURL)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	report, err := h.inspect(ctx, origin, host)
	if err != nil {
		writeDomainInspectionError(w, err, "Could not inspect panel settings. Check storage and retry")
		return
	}
	token, err := domainRandom()
	if err != nil {
		writeError(w, 500, "Could not create the verification check")
		return
	}
	report.CheckToken = token
	h.mu.Lock()
	h.tickets[claims.UserID] = domainTicket{report, claims.AuthVersion}
	h.mu.Unlock()
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, 200, report)
}
func (h *PanelDomainHandler) Apply(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.begin(w, r, "apply")
	if !ok {
		return
	}
	defer h.finish(claims.UserID)
	var req struct {
		PublicURL  string `json:"public_url"`
		CheckToken string `json:"check_token"`
	}
	if !decodeDomainBody(w, r, &req) {
		return
	}
	origin, host, err := normalizeDomainOrigin(req.PublicURL)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	ticket, ok := h.ticket(claims, origin, req.CheckToken)
	if !ok || !ticket.report.Ready {
		writeError(w, 409, "Run the domain checks successfully before applying this address. Checks expire after five minutes")
		return
	}
	// Recheck DNS, certificate, instance proof and passkeys immediately before save.
	// Old URL stays active throughout; a check is not an authorization bypass.
	h.clearTicket(claims.UserID)
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	report, err := h.inspect(ctx, origin, host)
	if err != nil {
		writeDomainInspectionError(w, err, "Could not recheck the address. The saved address is unchanged")
		return
	}
	if !report.Ready {
		writeJSON(w, 409, map[string]any{"error": "The domain checks changed. Resolve the reported issue and check again; the saved address is unchanged", "check": report})
		return
	}
	h.auth.panelURLMu.Lock()
	defer h.auth.panelURLMu.Unlock()
	current, err := h.auth.configuredPanelURL()
	if err != nil {
		writeError(w, 500, "Could not read the current panel address")
		return
	}
	if current.PublicURL != ticket.report.CurrentURL {
		writeError(w, 409, "Another administrator changed the panel address. Refresh and check again")
		return
	}
	if err := h.auth.db.SetPanelPublicURL(origin, host); err != nil {
		if errors.Is(err, database.ErrPublicURLPasskeys) {
			writeError(w, 409, err.Error())
		} else {
			writeError(w, 500, "Could not save the panel address")
		}
		return
	}
	logAudit(h.auth.db, r, "admin_panel_url", "Verified domain connectivity and passkey compatibility; updated public panel URL to "+origin)
	writeJSON(w, 200, panelURLSettings{PublicURL: origin, Source: "settings"})
}
func (h *PanelDomainHandler) SetupHTTPS(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.begin(w, r, "setup")
	if !ok {
		return
	}
	defer h.finish(claims.UserID)
	var req struct {
		PublicURL  string `json:"public_url"`
		CheckToken string `json:"check_token"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		Confirm    bool   `json:"confirm"`
	}
	if !decodeDomainBody(w, r, &req) {
		return
	}
	if !req.Confirm || claims.UserID <= 0 || claims.APIKey {
		writeError(w, 400, "Sign in as a panel administrator and explicitly confirm HTTPS setup")
		return
	}
	user, err := h.auth.db.GetUserByID(claims.UserID)
	if err != nil || user == nil || user.Role != "admin" || !h.auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		writeError(w, 403, "Confirm with your current administrator password")
		return
	}
	req.Password = ""
	origin, host, err := normalizeDomainOrigin(req.PublicURL)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	u, _ := url.Parse(origin)
	if u.Scheme != "https" || u.Port() != "" || domainLocal(host) {
		writeError(w, 400, "Automatic Let's Encrypt setup requires a public hostname using standard HTTPS port 443")
		return
	}
	ticket, ok := h.ticket(claims, origin, req.CheckToken)
	if !ok || !ticket.report.DNSReady {
		writeError(w, 409, "Check this domain's DNS before setting up HTTPS")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
	defer cancel()
	capability := h.caddy.Capability(ctx)
	if !capability.Available {
		writeError(w, 409, capability.Reason)
		return
	}
	// Only prepare a domain on the same public addresses as this recognized
	// Caddy installation. This is not a generic proxy editor for arbitrary hosts.
	target, err := h.resolve(ctx, host)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	known := make(map[netip.Addr]bool)
	for _, existing := range capability.Hosts {
		ips, err := h.resolve(ctx, existing)
		if err == nil {
			for _, ip := range ips {
				known[ip] = true
			}
		}
	}
	for _, ip := range target {
		if !known[ip] {
			writeError(w, 409, "The new domain's DNS does not match this installation's recognized Caddy host. Point all A/AAAA records at this panel, or configure your external proxy manually")
			return
		}
	}
	h.clearTicket(claims.UserID)
	logAudit(h.auth.db, r, "admin_panel_https", "Requested additive Let's Encrypt setup for "+host)
	result, err := h.caddy.Provision(ctx, host, req.Email)
	if err != nil {
		logAudit(h.auth.db, r, "admin_panel_https_failed", "HTTPS setup did not finish for "+host)
		writeError(w, 502, "HTTPS setup did not finish: "+err.Error()+". The saved panel address has not changed")
		return
	}
	logAudit(h.auth.db, r, "admin_panel_https_prepared", "Prepared HTTPS configuration for "+host+"; certificate and domain verification still required")
	writeJSON(w, 200, map[string]any{"changed": result.Changed, "hostname": result.Hostname, "message": result.Message, "needs_verification": true})
}

func writeDomainInspectionError(w http.ResponseWriter, err error, message string) {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		writeError(w, http.StatusGatewayTimeout, "Domain verification timed out or was cancelled. Wait for any running HTTPS setup, then check again. The saved address is unchanged")
		return
	}
	writeError(w, http.StatusInternalServerError, message)
}
