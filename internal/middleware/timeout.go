package middleware

import (
	"context"
	"net/http"
	"time"
)

// Timeout returns a chi-compatible middleware that adds a context deadline to
// every incoming request. If the handler does not complete within the specified
// duration, the context is cancelled and downstream operations that respect
// context cancellation will stop.
//
// The default recommended duration is 30 seconds. The timeout value should be
// configurable via environment variable (REQUEST_TIMEOUT).
func Timeout(duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), duration)
			defer cancel()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
