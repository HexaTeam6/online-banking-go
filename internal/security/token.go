package security

import (
	"errors"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenManager abstracts JWT token operations.
type TokenManager interface {
	Generate(claims TokenClaims) (string, error)
	Validate(tokenString string) (*TokenClaims, error)
	Invalidate(tokenString string) error
}

// TokenClaims holds JWT payload data.
type TokenClaims struct {
	AccountNo int64     `json:"account_no"`
	Role      string    `json:"role"` // "customer" or "admin"
	ExpiresAt time.Time `json:"expires_at"`
}

// jwtCustomClaims embeds jwt.RegisteredClaims and adds custom fields.
type jwtCustomClaims struct {
	AccountNo int64  `json:"account_no"`
	Role      string `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager implements TokenManager using golang-jwt.
type JWTManager struct {
	secret    []byte
	blacklist map[string]time.Time
	mu        sync.RWMutex
}

// NewJWTManager creates a new JWTManager with the given secret string.
func NewJWTManager(secret string) *JWTManager {
	m := &JWTManager{
		secret:    []byte(secret),
		blacklist: make(map[string]time.Time),
	}
	return m
}

// Generate creates a signed JWT token with role-based expiry.
// Customer tokens expire in 15 minutes, admin tokens expire in 30 minutes.
func (m *JWTManager) Generate(claims TokenClaims) (string, error) {
	var expiry time.Duration
	switch claims.Role {
	case "admin":
		expiry = 30 * time.Minute
	case "customer":
		expiry = 15 * time.Minute
	default:
		return "", errors.New("invalid role: must be 'customer' or 'admin'")
	}

	now := time.Now()
	expiresAt := now.Add(expiry)

	customClaims := jwtCustomClaims{
		AccountNo: claims.AccountNo,
		Role:      claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims)
	signedToken, err := token.SignedString(m.secret)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// Validate parses and validates a JWT token string.
// It checks the signature, expiration, and blacklist status.
func (m *JWTManager) Validate(tokenString string) (*TokenClaims, error) {
	// Check blacklist first
	m.mu.RLock()
	if _, blacklisted := m.blacklist[tokenString]; blacklisted {
		m.mu.RUnlock()
		return nil, errors.New("token has been invalidated")
	}
	m.mu.RUnlock()

	token, err := jwt.ParseWithClaims(tokenString, &jwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	customClaims, ok := token.Claims.(*jwtCustomClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return &TokenClaims{
		AccountNo: customClaims.AccountNo,
		Role:      customClaims.Role,
		ExpiresAt: customClaims.ExpiresAt.Time,
	}, nil
}

// Invalidate adds a token to the blacklist. The token remains blacklisted
// until it would have naturally expired (lazy cleanup on next access).
func (m *JWTManager) Invalidate(tokenString string) error {
	// Parse the token to get its expiry for cleanup purposes
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &jwtCustomClaims{})
	if err != nil {
		return errors.New("cannot parse token for invalidation")
	}

	customClaims, ok := token.Claims.(*jwtCustomClaims)
	if !ok {
		return errors.New("invalid token claims")
	}

	expiryTime := customClaims.ExpiresAt.Time

	m.mu.Lock()
	m.blacklist[tokenString] = expiryTime
	m.mu.Unlock()

	// Lazy cleanup: remove expired entries from the blacklist
	m.cleanupExpired()

	return nil
}

// cleanupExpired removes tokens from the blacklist that have passed their expiry time.
func (m *JWTManager) cleanupExpired() {
	now := time.Now()
	m.mu.Lock()
	defer m.mu.Unlock()

	for token, expiry := range m.blacklist {
		if now.After(expiry) {
			delete(m.blacklist, token)
		}
	}
}
