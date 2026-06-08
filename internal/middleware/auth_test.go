package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/abdurrachmanwahed/online-banking/internal/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTokenManager is a test double for the TokenManager interface.
type mockTokenManager struct {
	validateFn func(tokenString string) (*security.TokenClaims, error)
}

func (m *mockTokenManager) Generate(_ security.TokenClaims) (string, error) {
	return "", nil
}

func (m *mockTokenManager) Validate(tokenString string) (*security.TokenClaims, error) {
	return m.validateFn(tokenString)
}

func (m *mockTokenManager) Invalidate(_ string) error {
	return nil
}

// okHandler is used to verify the middleware passes the request through.
func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

func TestAuthenticate_MissingAuthorizationHeader(t *testing.T) {
	tm := &mockTokenManager{}
	handler := Authenticate(tm)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "missing authorization header")
}

func TestAuthenticate_InvalidHeaderFormat(t *testing.T) {
	tm := &mockTokenManager{}
	handler := Authenticate(tm)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic abc123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid authorization header format")
}

func TestAuthenticate_MissingToken(t *testing.T) {
	tm := &mockTokenManager{}
	handler := Authenticate(tm)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer ")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "missing token")
}

func TestAuthenticate_InvalidToken(t *testing.T) {
	tm := &mockTokenManager{
		validateFn: func(_ string) (*security.TokenClaims, error) {
			return nil, errors.New("token expired")
		},
	}
	handler := Authenticate(tm)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid or expired token")
}

func TestAuthenticate_ValidToken(t *testing.T) {
	expectedClaims := &security.TokenClaims{
		AccountNo: 123456789,
		Role:      "customer",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	tm := &mockTokenManager{
		validateFn: func(tokenString string) (*security.TokenClaims, error) {
			assert.Equal(t, "valid-token", tokenString)
			return expectedClaims, nil
		},
	}

	// Use a handler that verifies context values.
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountNo := GetAccountNoFromContext(r.Context())
		role := GetRoleFromContext(r.Context())
		token := GetTokenFromContext(r.Context())

		assert.Equal(t, int64(123456789), accountNo)
		assert.Equal(t, "customer", role)
		assert.Equal(t, "valid-token", token)

		w.WriteHeader(http.StatusOK)
	})

	handler := Authenticate(tm)(inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthenticate_BearerCaseInsensitive(t *testing.T) {
	expectedClaims := &security.TokenClaims{
		AccountNo: 100000001,
		Role:      "admin",
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	tm := &mockTokenManager{
		validateFn: func(_ string) (*security.TokenClaims, error) {
			return expectedClaims, nil
		},
	}

	handler := Authenticate(tm)(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "bearer my-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireRole_MatchingRole(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Simulate authenticated context with role "admin".
	tm := &mockTokenManager{
		validateFn: func(_ string) (*security.TokenClaims, error) {
			return &security.TokenClaims{
				AccountNo: 100000001,
				Role:      "admin",
				ExpiresAt: time.Now().Add(30 * time.Minute),
			}, nil
		},
	}

	handler := Authenticate(tm)(RequireRole("admin")(inner))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-admin-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireRole_MismatchedRole(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Simulate authenticated context with role "customer" trying to access admin route.
	tm := &mockTokenManager{
		validateFn: func(_ string) (*security.TokenClaims, error) {
			return &security.TokenClaims{
				AccountNo: 123456789,
				Role:      "customer",
				ExpiresAt: time.Now().Add(15 * time.Minute),
			}, nil
		},
	}

	handler := Authenticate(tm)(RequireRole("admin")(inner))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-customer-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "insufficient permissions")
}

func TestRequireRole_NoRoleInContext(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// RequireRole without Authenticate middleware — no role in context.
	handler := RequireRole("admin")(inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "authentication required")
}

func TestGetAccountNoFromContext_NoValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	accountNo := GetAccountNoFromContext(req.Context())
	assert.Equal(t, int64(0), accountNo)
}

func TestGetRoleFromContext_NoValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	role := GetRoleFromContext(req.Context())
	assert.Equal(t, "", role)
}

func TestGetTokenFromContext_NoValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	token := GetTokenFromContext(req.Context())
	assert.Equal(t, "", token)
}
