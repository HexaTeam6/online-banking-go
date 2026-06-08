// Package middleware provides HTTP middleware components for the banking API.
package middleware

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/abdurrachmanwahed/online-banking/pkg/response"
)

// visitor holds a rate limiter and the last time it was accessed.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// rateLimiterStore manages per-IP rate limiters with cleanup of stale entries.
type rateLimiterStore struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    rate.Limit
	burst    int
}

// newRateLimiterStore creates a new store and starts background cleanup.
func newRateLimiterStore(requestsPerWindow int, windowDuration time.Duration) *rateLimiterStore {
	// rate.Limit is events per second. For requestsPerWindow in windowDuration:
	// interval = windowDuration / requestsPerWindow
	interval := windowDuration / time.Duration(requestsPerWindow)
	limit := rate.Every(interval)

	store := &rateLimiterStore{
		visitors: make(map[string]*visitor),
		limit:    limit,
		burst:    requestsPerWindow,
	}

	// Start background cleanup goroutine to remove stale entries.
	go store.cleanupLoop(windowDuration)

	return store
}

// getLimiter returns the rate limiter for the given IP, creating one if needed.
func (s *rateLimiterStore) getLimiter(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, exists := s.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(s.limit, s.burst)
		s.visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// cleanupLoop periodically removes visitors that haven't been seen
// for longer than the window duration.
func (s *rateLimiterStore) cleanupLoop(windowDuration time.Duration) {
	// Clean up every window duration.
	ticker := time.NewTicker(windowDuration)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		for ip, v := range s.visitors {
			if time.Since(v.lastSeen) > windowDuration {
				delete(s.visitors, ip)
			}
		}
		s.mu.Unlock()
	}
}

// RateLimiter returns a chi-compatible middleware that enforces per-IP rate limiting.
// It allows requestsPerWindow requests within windowDuration per client IP address.
// When the limit is exceeded, it returns HTTP 429 with a Retry-After header.
func RateLimiter(requestsPerWindow int, windowDuration time.Duration) func(http.Handler) http.Handler {
	store := newRateLimiterStore(requestsPerWindow, windowDuration)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractClientIP(r)
			limiter := store.getLimiter(ip)

			if !limiter.Allow() {
				// Calculate Retry-After: time until next token is available.
				reservation := limiter.Reserve()
				delay := reservation.Delay()
				reservation.Cancel()

				retryAfter := int(math.Ceil(delay.Seconds()))
				if retryAfter < 1 {
					retryAfter = 1
				}

				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				response.Error(w, http.StatusTooManyRequests, "Too Many Requests", "Rate limit exceeded. Please try again later.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractClientIP determines the client IP from the request.
// It checks X-Forwarded-For, X-Real-IP headers, then falls back to RemoteAddr.
func extractClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (first IP in the chain is the client).
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
		// Take the first one.
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	// Check X-Real-IP header.
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr, stripping the port.
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// RemoteAddr might not have a port.
		return r.RemoteAddr
	}
	return ip
}
