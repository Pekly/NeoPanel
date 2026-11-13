package api

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// clientBucket tracks token-bucket rate limit state for a single client IP.
type clientBucket struct {
	tokens     float64
	lastRefill time.Time
}

// IPRateLimiter provides a high-performance in-memory token bucket rate limiter
// using pure Go standard library with zero external dependencies.
type IPRateLimiter struct {
	mu            sync.RWMutex
	buckets       map[string]*clientBucket
	ratePerSec    float64
	burstCapacity float64
	authBuckets   map[string]*clientBucket
	authRateSec   float64
	authBurst     float64
	lastCleanup   time.Time
}

// NewIPRateLimiter creates a rate limiter.
// defaultLimit: requests per minute allowed for general endpoints (e.g. 180 req/min).
// authLimit: requests per minute allowed for sensitive auth endpoints (e.g. 15 req/min).
func NewIPRateLimiter(defaultLimit float64, authLimit float64) *IPRateLimiter {
	if defaultLimit <= 0 {
		defaultLimit = 180 // 3 req/sec continuous
	}
	if authLimit <= 0 {
		authLimit = 15 // thwarts brute-force login attempts
	}

	return &IPRateLimiter{
		buckets:       make(map[string]*clientBucket),
		ratePerSec:    defaultLimit / 60.0,
		burstCapacity: defaultLimit / 3.0, // burst capacity (e.g. 60 requests)
		authBuckets:   make(map[string]*clientBucket),
		authRateSec:   authLimit / 60.0,
		authBurst:     authLimit / 3.0, // auth burst (e.g. 5 requests)
		lastCleanup:   time.Now(),
	}
}

// extractClientIP extracts the real client IP, respecting trusted proxy headers.
func extractClientIP(r *http.Request) string {
	// Check standard proxy headers
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		ip := strings.TrimSpace(xri)
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		ip := strings.TrimSpace(cfip)
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && net.ParseIP(host) != nil {
		return host
	}
	return r.RemoteAddr
}

// Allow checks if the request from the client IP should be permitted.
func (lim *IPRateLimiter) Allow(r *http.Request) bool {
	ip := extractClientIP(r)

	// Always permit local loopback and inter-daemon node traffic
	if ip == "127.0.0.1" || ip == "::1" || ip == "localhost" {
		return true
	}

	now := time.Now()
	isAuth := strings.HasPrefix(r.URL.Path, "/api/auth/login") || strings.HasPrefix(r.URL.Path, "/api/auth/setup") || strings.HasPrefix(r.URL.Path, "/api/auth/invitations/")

	lim.mu.Lock()
	defer lim.mu.Unlock()

	// Periodic cleanup of stale IP buckets every 5 minutes
	if now.Sub(lim.lastCleanup) > 5*time.Minute {
		for k, b := range lim.buckets {
			if now.Sub(b.lastRefill) > 10*time.Minute {
				delete(lim.buckets, k)
			}
		}
		for k, b := range lim.authBuckets {
			if now.Sub(b.lastRefill) > 10*time.Minute {
				delete(lim.authBuckets, k)
			}
		}
		lim.lastCleanup = now
	}

	targetMap := lim.buckets
	rate := lim.ratePerSec
	burst := lim.burstCapacity

	if isAuth {
		targetMap = lim.authBuckets
		rate = lim.authRateSec
		burst = lim.authBurst
	}

	bucket, exists := targetMap[ip]
	if !exists {
		targetMap[ip] = &clientBucket{
			tokens:     burst - 1,
			lastRefill: now,
		}
		return true
	}

	// Refill tokens according to elapsed time
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	bucket.tokens += elapsed * rate
	if bucket.tokens > burst {
		bucket.tokens = burst
	}
	bucket.lastRefill = now

	if bucket.tokens >= 1.0 {
		bucket.tokens -= 1.0
		return true
	}

	return false
}

// Middleware returns an HTTP middleware handler that rejects rate-limited requests.
func (lim *IPRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only rate-limit API routes, let static assets through quickly
		if strings.HasPrefix(r.URL.Path, "/api/") {
			if !lim.Allow(r) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "5")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"Too many requests. Please slow down and try again shortly."}`))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
