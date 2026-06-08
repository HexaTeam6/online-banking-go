package middleware

import (
	"net/http"
	"strings"

	"github.com/abdurrachmanwahed/online-banking/pkg/response"
)

// RequireJSON is a chi-compatible middleware that enforces Content-Type
// application/json on requests that carry a body (POST, PUT, PATCH).
// Requests with other methods or a matching Content-Type header pass through
// unchanged. Non-JSON requests receive a 415 Unsupported Media Type response.
func RequireJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only enforce on methods that typically carry a request body.
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			ct := r.Header.Get("Content-Type")
			if !strings.HasPrefix(ct, "application/json") {
				response.Error(w, http.StatusUnsupportedMediaType, "Unsupported Media Type", "Content-Type must be application/json")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
