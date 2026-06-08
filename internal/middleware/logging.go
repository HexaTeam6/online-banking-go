package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// responseRecorder wraps http.ResponseWriter to capture the status code
// written by downstream handlers.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader captures the status code and delegates to the underlying writer.
func (rr *responseRecorder) WriteHeader(code int) {
	if !rr.written {
		rr.statusCode = code
		rr.written = true
	}
	rr.ResponseWriter.WriteHeader(code)
}

// Write delegates to the underlying writer and sets the default 200 status
// if WriteHeader has not been called explicitly.
func (rr *responseRecorder) Write(b []byte) (int, error) {
	if !rr.written {
		rr.statusCode = http.StatusOK
		rr.written = true
	}
	return rr.ResponseWriter.Write(b)
}

// RequestLogger returns a chi-compatible middleware that logs each request
// using zerolog with structured JSON output.
//
// Each log entry includes:
//   - timestamp (ISO 8601)
//   - level (INFO for success, WARN for 4xx, ERROR for 5xx)
//   - method
//   - path
//   - status code
//   - duration in milliseconds
func RequestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rec := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rec, r)

			duration := time.Since(start)
			durationMs := float64(duration.Nanoseconds()) / float64(time.Millisecond)

			var event *zerolog.Event
			switch {
			case rec.statusCode >= 500:
				event = log.Error()
			case rec.statusCode >= 400:
				event = log.Warn()
			default:
				event = log.Info()
			}

			event.
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", rec.statusCode).
				Float64("duration_ms", durationMs).
				Msg("request completed")
		})
	}
}
