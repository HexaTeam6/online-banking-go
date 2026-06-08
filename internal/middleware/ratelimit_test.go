package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/abdurrachmanwahed/online-banking/pkg/response"
)

func TestRateLimiter_AllowsRequestsWithinLimit(t *testing.T) {
	// Configure: 5 requests per 60s window.
	mw := RateLimiter(5, 60*time.Second)
	handler := mw(okHandler())

	// First 5 requests from same IP should pass.
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "Request %d should succeed", i+1)
	}
}

func TestRateLimiter_RejectsExcessRequests(t *testing.T) {
	// Configure: 5 requests per 60s window.
	mw := RateLimiter(5, 60*time.Second)
	handler := mw(okHandler())

	// Exhaust the 5 allowed requests.
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		req.RemoteAddr = "10.0.0.1:5000"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	}

	// 6th request should be rate limited.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = "10.0.0.1:5000"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)

	// Verify Retry-After header is present and positive.
	retryAfter := rec.Header().Get("Retry-After")
	assert.NotEmpty(t, retryAfter, "Retry-After header should be set")

	// Verify JSON error response.
	var errResp response.ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "Too Many Requests", errResp.Error.Code)
	assert.Contains(t, errResp.Error.Message, "Rate limit exceeded")
}

func TestRateLimiter_DifferentIPsHaveSeparateLimits(t *testing.T) {
	mw := RateLimiter(2, 60*time.Second)
	handler := mw(okHandler())

	// IP 1: exhaust limit.
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		req.RemoteAddr = "1.1.1.1:1000"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// IP 1: should be rate limited.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = "1.1.1.1:1000"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)

	// IP 2: should still be allowed.
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = "2.2.2.2:2000"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestExtractClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 70.41.3.18, 150.172.238.178")

	ip := extractClientIP(req)
	assert.Equal(t, "203.0.113.50", ip)
}

func TestExtractClientIP_XForwardedForSingle(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50")

	ip := extractClientIP(req)
	assert.Equal(t, "203.0.113.50", ip)
}

func TestExtractClientIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "198.51.100.42")

	ip := extractClientIP(req)
	assert.Equal(t, "198.51.100.42", ip)
}

func TestExtractClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.0.1:8080"

	ip := extractClientIP(req)
	assert.Equal(t, "192.168.0.1", ip)
}

func TestExtractClientIP_RemoteAddrWithoutPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.0.1"

	ip := extractClientIP(req)
	assert.Equal(t, "192.168.0.1", ip)
}

func TestExtractClientIP_XForwardedForTakesPrecedence(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "10.10.10.10")
	req.Header.Set("X-Real-IP", "20.20.20.20")
	req.RemoteAddr = "30.30.30.30:1234"

	ip := extractClientIP(req)
	assert.Equal(t, "10.10.10.10", ip)
}

func TestRateLimiter_RetryAfterHeaderIsPositiveInteger(t *testing.T) {
	mw := RateLimiter(1, 60*time.Second)
	handler := mw(okHandler())

	// First request passes.
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = "5.5.5.5:999"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Second request should be rate limited with Retry-After.
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = "5.5.5.5:999"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)

	retryAfter := rec.Header().Get("Retry-After")
	assert.NotEmpty(t, retryAfter)
	// Retry-After should be a positive integer string.
	assert.Regexp(t, `^\d+$`, retryAfter)
}
