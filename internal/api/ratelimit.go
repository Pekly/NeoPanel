package api

import (
	"context"
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

type clientIPContextKey struct{}

// clientIPPolicy is immutable and local to a router. No forwarding header is
// trusted unless the socket peer is explicitly configured as a trusted proxy.
type clientIPPolicy struct{ trusted []*net.IPNet }

func newClientIPPolicy(proxies []string) *clientIPPolicy {
	p := &clientIPPolicy{}
	for _, value := range proxies {
		value = strings.TrimSpace(value)
		if ip := net.ParseIP(value); ip != nil {
			bits := 128
			if ip.To4() != nil {
				ip = ip.To4()
				bits = 32
			}
			p.trusted = append(p.trusted, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
		} else if _, block, err := net.ParseCIDR(value); err == nil {
			p.trusted = append(p.trusted, block)
		}
	}
	return p
}

func (p *clientIPPolicy) trusts(ip net.IP) bool {
	if p != nil {
		for _, block := range p.trusted {
			if block.Contains(ip) {
				return true
			}
		}
	}
	return false
}

func peerIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.String()
	}
	return host
}

func (p *clientIPPolicy) resolve(r *http.Request) string {
	peer := peerIP(r)
	if !p.trusts(net.ParseIP(peer)) {
		return peer
	}
	// Bound allocation and parsing, including repeated header lines.
	values := r.Header.Values("X-Forwarded-For")
	size := 0
	for _, value := range values {
		size += len(value) + 1
		if size > 4096 {
			return peer
		}
	}
	if size == 0 {
		return peer
	}
	hops := strings.Split(strings.Join(values, ","), ",")
	if len(hops) > 32 {
		return peer
	}
	current := peer
	for i := len(hops) - 1; i >= 0; i-- {
		if !p.trusts(net.ParseIP(current)) {
			return current
		}
		ip := net.ParseIP(strings.TrimSpace(hops[i]))
		if ip == nil {
			return peer
		}
		current = ip.String()
	}
	return current
}

func (p *clientIPPolicy) request(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), clientIPContextKey{}, p.resolve(r)))
}

// Audit, Turnstile and all limiters share the same router-resolved identity.
// Direct handler calls use only the socket peer.
func extractClientIP(r *http.Request) string {
	if ip, ok := r.Context().Value(clientIPContextKey{}).(string); ok {
		return ip
	}
	return peerIP(r)
}

// Allow checks if the request from the client IP should be permitted.
func (lim *IPRateLimiter) Allow(r *http.Request) bool {
	ip := extractClientIP(r)

	now := time.Now()
	isAuth := strings.HasPrefix(r.URL.Path, "/api/auth/")

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
