package security

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTManager_Generate_ProducesValidToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key")

	claims := TokenClaims{
		AccountNo: 123456789,
		Role:      "customer",
	}

	tokenString, err := manager.Generate(claims)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Validate the generated token
	result, err := manager.Validate(tokenString)
	require.NoError(t, err)
	assert.Equal(t, int64(123456789), result.AccountNo)
	assert.Equal(t, "customer", result.Role)
}

func TestJWTManager_Generate_CustomerExpiryIs15Min(t *testing.T) {
	manager := NewJWTManager("test-secret-key")

	claims := TokenClaims{
		AccountNo: 123456789,
		Role:      "customer",
	}

	tokenString, err := manager.Generate(claims)
	require.NoError(t, err)

	result, err := manager.Validate(tokenString)
	require.NoError(t, err)

	// Expiry should be approximately 15 minutes from now
	expectedExpiry := time.Now().Add(15 * time.Minute)
	assert.WithinDuration(t, expectedExpiry, result.ExpiresAt, 5*time.Second)
}

func TestJWTManager_Generate_AdminExpiryIs30Min(t *testing.T) {
	manager := NewJWTManager("test-secret-key")

	claims := TokenClaims{
		AccountNo: 1,
		Role:      "admin",
	}

	tokenString, err := manager.Generate(claims)
	require.NoError(t, err)

	result, err := manager.Validate(tokenString)
	require.NoError(t, err)

	// Expiry should be approximately 30 minutes from now
	expectedExpiry := time.Now().Add(30 * time.Minute)
	assert.WithinDuration(t, expectedExpiry, result.ExpiresAt, 5*time.Second)
}

func TestJWTManager_Generate_InvalidRoleReturnsError(t *testing.T) {
	manager := NewJWTManager("test-secret-key")

	claims := TokenClaims{
		AccountNo: 123456789,
		Role:      "invalid",
	}

	_, err := manager.Generate(claims)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role")
}

func TestJWTManager_Validate_ReturnsCorrectClaims(t *testing.T) {
	manager := NewJWTManager("test-secret-key")

	claims := TokenClaims{
		AccountNo: 987654321,
		Role:      "admin",
	}

	tokenString, err := manager.Generate(claims)
	require.NoError(t, err)

	result, err := manager.Validate(tokenString)
	require.NoError(t, err)
	assert.Equal(t, int64(987654321), result.AccountNo)
	assert.Equal(t, "admin", result.Role)
	assert.False(t, result.ExpiresAt.IsZero())
}

func TestJWTManager_Validate_ExpiredTokenRejected(t *testing.T) {
	manager := NewJWTManager("test-secret-key")

	// Create a token that's already expired
	now := time.Now()
	customClaims := jwtCustomClaims{
		AccountNo: 123456789,
		Role:      "customer",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims)
	tokenString, err := token.SignedString([]byte("test-secret-key"))
	require.NoError(t, err)

	_, err = manager.Validate(tokenString)
	assert.Error(t, err)
}

func TestJWTManager_Validate_InvalidatedTokenRejected(t *testing.T) {
	manager := NewJWTManager("test-secret-key")

	claims := TokenClaims{
		AccountNo: 123456789,
		Role:      "customer",
	}

	tokenString, err := manager.Generate(claims)
	require.NoError(t, err)

	// Validate before invalidation should succeed
	_, err = manager.Validate(tokenString)
	require.NoError(t, err)

	// Invalidate the token
	err = manager.Invalidate(tokenString)
	require.NoError(t, err)

	// Validate after invalidation should fail
	_, err = manager.Validate(tokenString)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalidated")
}

func TestJWTManager_Validate_WrongSignatureRejected(t *testing.T) {
	manager1 := NewJWTManager("secret-one")
	manager2 := NewJWTManager("secret-two")

	claims := TokenClaims{
		AccountNo: 123456789,
		Role:      "customer",
	}

	// Generate with one secret
	tokenString, err := manager1.Generate(claims)
	require.NoError(t, err)

	// Validate with different secret
	_, err = manager2.Validate(tokenString)
	assert.Error(t, err)
}

func TestJWTManager_Validate_MalformedTokenRejected(t *testing.T) {
	manager := NewJWTManager("test-secret-key")

	_, err := manager.Validate("not.a.valid.token")
	assert.Error(t, err)
}

func TestJWTManager_Invalidate_MalformedTokenReturnsError(t *testing.T) {
	manager := NewJWTManager("test-secret-key")

	err := manager.Invalidate("completely-invalid-garbage")
	assert.Error(t, err)
}

func TestJWTManager_CleanupExpired_RemovesExpiredEntries(t *testing.T) {
	manager := NewJWTManager("test-secret-key")

	// Manually add an expired entry to the blacklist
	manager.mu.Lock()
	manager.blacklist["expired-token"] = time.Now().Add(-1 * time.Hour)
	manager.mu.Unlock()

	// Trigger cleanup
	manager.cleanupExpired()

	manager.mu.RLock()
	_, exists := manager.blacklist["expired-token"]
	manager.mu.RUnlock()

	assert.False(t, exists, "expired token should have been cleaned up")
}
