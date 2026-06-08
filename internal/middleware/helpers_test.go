package middleware

import "net/http"

// dummyHandler returns a simple handler that responds with 200 OK.
// It is used as the downstream handler in middleware tests.
func dummyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
