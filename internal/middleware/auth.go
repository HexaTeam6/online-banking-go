// Package middleware provides HTTP middleware for the online banking API.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/abdurrachmanwahed/online-banking/internal/security"
	"github.com/abdurrachmanwahed/online-banking/pkg/response"
)

// contextKey is an unexported type for context keys to avoid collisions.
type contextKey string

const (
	// contextKeyAccountNo is the context key for the authenticated account number.
	contextKeyAccountNo contextKey = "account_no"
	// contextKeyRole is the context key for the authenticated user's role.
	contextKeyRole contextKey = "role"
	// contextKeyToken is the context key for the raw JWT token string.
	contextKeyToken contextKey = "token"
)

// Authenticate returns a middleware that validates JWT tokens from the
// Authorization header. On success, it injects the token claims (AccountNo,
// Role) and the raw token string into the request context. On failure, it
// responds with 401 Unauthorized.
func Authenticate(tokenManager security.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, "Unauthorized", "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.Error(w, http.StatusUnauthorized, "Unauthorized", "invalid authorization header format")
				return
			}

			tokenString := strings.TrimSpace(parts[1])
			if tokenString == "" {
				response.Error(w, http.StatusUnauthorized, "Unauthorized", "missing token")
				return
			}

			claims, err := tokenManager.Validate(tokenString)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "Unauthorized", "invalid or expired token")
				return
			}

			// Inject claims and token into the request context.
			ctx := r.Context()
			ctx = context.WithValue(ctx, contextKeyAccountNo, claims.AccountNo)
			ctx = context.WithValue(ctx, contextKeyRole, claims.Role)
			ctx = context.WithValue(ctx, contextKeyToken, tokenString)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole returns a middleware that checks whether the authenticated user
// has the specified role. It must be applied after Authenticate. If the user's
// role does not match, it responds with 403 Forbidden.
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(contextKeyRole).(string)
			if !ok || userRole == "" {
				response.Error(w, http.StatusUnauthorized, "Unauthorized", "authentication required")
				return
			}

			if userRole != role {
				response.Error(w, http.StatusForbidden, "Forbidden", "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetAccountNoFromContext extracts the authenticated account number from the
// request context. Returns 0 if the value is not present.
func GetAccountNoFromContext(ctx context.Context) int64 {
	accountNo, _ := ctx.Value(contextKeyAccountNo).(int64)
	return accountNo
}

// GetRoleFromContext extracts the authenticated user's role from the request
// context. Returns an empty string if the value is not present.
func GetRoleFromContext(ctx context.Context) string {
	role, _ := ctx.Value(contextKeyRole).(string)
	return role
}

// GetTokenFromContext extracts the raw JWT token string from the request
// context. Returns an empty string if the value is not present.
func GetTokenFromContext(ctx context.Context) string {
	token, _ := ctx.Value(contextKeyToken).(string)
	return token
}
