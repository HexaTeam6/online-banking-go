package middleware

import (
	"net/http"
	"strings"

	"github.com/abdurrachmanwahed/online-banking/pkg/response"
)

// BodyLimit returns a chi-compatible middleware that restricts the request body
// size to maxBytes. If the body exceeds the limit, the request is rejected with
// a 413 Payload Too Large response. The default recommended limit is 1 MB
// (1048576 bytes).
func BodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// IsMaxBytesError checks if an error is caused by the body exceeding the
// MaxBytesReader limit. Handlers can use this to detect oversized bodies and
// return appropriate 413 responses.
func IsMaxBytesError(err error) bool {
	if err == nil {
		return false
	}
	// http.MaxBytesError was added in Go 1.19.
	// The error message includes "http: request body too large".
	return strings.Contains(err.Error(), "http: request body too large")
}

// BodyLimitWithReject returns a chi-compatible middleware that rejects oversized
// request bodies immediately by reading a single byte beyond the limit to
// trigger the MaxBytesReader error. This provides immediate 413 feedback before
// the handler reads the body.
func BodyLimitWithReject(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil && r.ContentLength > maxBytes {
				response.Error(w, http.StatusRequestEntityTooLarge, "Payload Too Large", "request body exceeds the maximum allowed size")
				return
			}

			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}

			next.ServeHTTP(w, r)
		})
	}
}
