package api

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Pekly/NeoPanel/internal/config"
)

const controlBodyLimit int64 = 1 << 20

func safeHTTPMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

// Route-specific allowances cannot be bypassed by claiming a multipart content
// type. The worker proxy receives the same limit as the matching local route.
func apiRequestBodyLimit(r *http.Request) int64 {
	p := r.URL.Path
	if strings.HasPrefix(p, "/api/nodes/") {
		parts := strings.Split(strings.TrimPrefix(p, "/api/nodes/"), "/")
		if len(parts) > 2 && parts[1] == "proxy" {
			p = "/" + strings.Join(parts[2:], "/")
		}
	}
	if r.Method == http.MethodPost {
		if p == "/api/account/avatar" {
			return avatarRequestLimit
		}
		if p == "/api/admin/branding/upload" {
			return (12 << 20) + 4096
		}
		if p == "/api/admin/mods/upload" {
			return (20 << 20) + 4096
		}
		parts := strings.Split(strings.Trim(p, "/"), "/")
		if len(parts) == 5 && parts[0] == "api" && parts[1] == "admin" && parts[2] == "users" && parts[4] == "avatar" {
			return avatarRequestLimit
		}
		if len(parts) >= 4 && parts[0] == "api" && parts[1] == "servers" {
			if len(parts) == 4 && parts[3] == "icon" {
				return 10 << 20
			}
			if len(parts) == 5 && parts[3] == "files" {
				if parts[4] == "upload" {
					return 200 << 20
				}
				if parts[4] == "content" {
					return 6*fileEditorLimit + (64 << 10)
				}
			}
			if len(parts) == 6 && parts[3] == "databases" && parts[5] == "query" {
				return maxServerSQLRequestBytes
			}
		}
	}
	return controlBodyLimit
}

func limitAPIRequestBody(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(r.URL.Path, "/api/") {
		return true
	}
	limit := apiRequestBodyLimit(r)
	if r.ContentLength > limit {
		writeError(w, http.StatusRequestEntityTooLarge, "Request body exceeds the endpoint limit")
		return false
	}
	if r.Body != nil {
		r.Body = http.MaxBytesReader(w, r.Body, limit)
	}
	return true
}

func trustedRequestPeer(r *http.Request, cfg *config.Config) bool {
	return cfg != nil && newClientIPPolicy(cfg.Security.TrustedProxies).trusts(net.ParseIP(peerIP(r)))
}
func secureTransport(r *http.Request, cfg *config.Config) bool {
	if r.TLS != nil {
		return true
	}
	values := r.Header.Values("X-Forwarded-Proto")
	return trustedRequestPeer(r, cfg) && len(values) == 1 && values[0] == "https"
}

func serializedOrigin(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.User != nil || u.Opaque != "" || u.Host == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", false
	}
	host := strings.ToLower(u.Hostname())
	port := u.Port()
	if host == "" || strings.HasSuffix(u.Host, ":") {
		return "", false
	}
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return "", false
	}
	return u.Scheme + "://" + net.JoinHostPort(host, strconv.Itoa(n)), true
}

func (h *AuthHandler) sameOriginRequest(r *http.Request) bool {
	origins := r.Header.Values("Origin")
	if len(origins) != 1 {
		return false
	}
	origin, ok := serializedOrigin(origins[0])
	if !ok {
		return false
	}
	if site := r.Header.Get("Sec-Fetch-Site"); site != "" && site != "same-origin" {
		return false
	}
	h.mu.RLock()
	cfg := h.cfg
	h.mu.RUnlock()
	scheme := "http"
	if secureTransport(r, cfg) {
		scheme = "https"
	}
	target, ok := serializedOrigin(scheme + "://" + r.Host)
	if ok && origin == target {
		return true
	}
	// Canonical HTTPS origins also cover an edge that preserves Host while
	// terminating TLS. Rewritten backend hosts require an explicitly trusted peer.
	// Arbitrary X-Forwarded-Host values never become browser authority.
	if h.db != nil {
		settings, err := h.panelURL()
		if err == nil {
			canonical, valid := serializedOrigin(settings.PublicURL)
			if valid && origin == canonical {
				public, _ := url.Parse(settings.PublicURL)
				canonicalHost, hostOK := serializedOrigin(public.Scheme + "://" + r.Host)
				return (hostOK && canonicalHost == canonical) || trustedRequestPeer(r, cfg)
			}
		}
	}
	return false
}

func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, r *http.Request, token string) {
	h.mu.RLock()
	cfg, hours := h.cfg, h.expHours
	h.mu.RUnlock()
	secure := secureTransport(r, cfg)
	if h.db != nil {
		if settings, err := h.panelURL(); err == nil && strings.HasPrefix(settings.PublicURL, "https://") {
			secure = true
		}
	}
	if hours <= 0 {
		hours = 72
	}
	http.SetCookie(w, &http.Cookie{Name: "neopanel_token", Value: token, Path: "/", Expires: time.Now().Add(time.Duration(hours) * time.Hour), HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode})
	// Profile updates used to issue an unused duplicate JWT cookie.
	if _, err := r.Cookie("jwt"); err == nil {
		http.SetCookie(w, &http.Cookie{Name: "jwt", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode})
	}
}

func (r *Router) responseSecurityHeaders(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "0")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=(), usb=()")
	// Do not claim an inline-script lockdown: this SPA and trusted extensions use
	// inline handlers. Specialized file/asset handlers can set a stronger sandbox.
	w.Header().Set("Content-Security-Policy", "base-uri 'self'; frame-ancestors 'none'")
	if secureTransport(req, r.cfg) {
		w.Header().Set("Strict-Transport-Security", "max-age=15552000")
	}
	if strings.HasPrefix(req.URL.Path, "/api/") {
		w.Header().Set("Cache-Control", "private, no-store")
	}
}
