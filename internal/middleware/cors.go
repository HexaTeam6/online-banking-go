package middleware

import (
	"net/http"
	"strings"
)

// CORS returns a chi-compatible middleware that sets Access-Control-Allow-Origin,
// Access-Control-Allow-Methods, and Access-Control-Allow-Headers response headers
// based on the provided configuration strings. Each parameter is a comma-separated
// list of allowed values (e.g., "GET,POST,PUT" for methods).
//
// Preflight (OPTIONS) requests are handled automatically with a 204 No Content
// response.
func CORS(allowedOrigins, allowedMethods, allowedHeaders string) func(http.Handler) http.Handler {
	origins := normalizeCSV(allowedOrigins)
	methods := normalizeCSV(allowedMethods)
	headers := normalizeCSV(allowedHeaders)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origins)
			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Access-Control-Allow-Headers", headers)

			// Handle preflight requests.
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// normalizeCSV trims whitespace from each item in a comma-separated string
// and returns them joined back with ", ".
func normalizeCSV(value string) string {
	parts := strings.Split(value, ",")
	trimmed := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s != "" {
			trimmed = append(trimmed, s)
		}
	}
	return strings.Join(trimmed, ", ")
}
