package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const turnstileSiteverifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type turnstileVerifyResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
	Action      string   `json:"action"`
}

func extractClientIP(r *http.Request) string {
	if cfIP := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); cfIP != "" {
		return cfIP
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

func generateIdempotencyKey() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// VerifyTurnstileToken validates a client cf-turnstile-response token against Cloudflare's siteverify API.
func VerifyTurnstileToken(secretKey, token string, r *http.Request) error {
	secretKey = strings.TrimSpace(secretKey)
	if secretKey == "" {
		return nil
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("Please complete the Cloudflare Turnstile security verification")
	}

	form := url.Values{}
	form.Set("secret", secretKey)
	form.Set("response", token)
	if r != nil {
		if ip := extractClientIP(r); ip != "" {
			form.Set("remoteip", ip)
		}
	}
	form.Set("idempotency_key", generateIdempotencyKey())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm(turnstileSiteverifyURL, form)
	if err != nil {
		return fmt.Errorf("failed to reach Cloudflare Turnstile verification service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Cloudflare Turnstile verification returned HTTP %d", resp.StatusCode)
	}

	var outcome turnstileVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&outcome); err != nil {
		return fmt.Errorf("invalid response from Cloudflare Turnstile: %w", err)
	}

	if !outcome.Success {
		for _, code := range outcome.ErrorCodes {
			switch code {
			case "timeout-or-duplicate":
				return errors.New("Security verification expired or already used. Please try again.")
			case "invalid-input-response":
				return errors.New("Invalid Cloudflare Turnstile verification token. Please try again.")
			case "invalid-input-secret", "missing-input-secret":
				return errors.New("Server Turnstile secret key is misconfigured. Contact administrator.")
			}
		}
		return errors.New("Cloudflare Turnstile verification failed. Please try again.")
	}

	return nil
}

