package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestNewBcryptHasher_CostEnforcement(t *testing.T) {
	tests := []struct {
		name         string
		inputCost    int
		expectedCost int
	}{
		{"cost below minimum defaults to 10", 4, minBcryptCost},
		{"cost of 0 defaults to 10", 0, minBcryptCost},
		{"cost of 9 defaults to 10", 9, minBcryptCost},
		{"cost of 10 stays at 10", 10, 10},
		{"cost of 12 stays at 12", 12, 12},
		{"negative cost defaults to 10", -1, minBcryptCost},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := NewBcryptHasher(tt.inputCost)
			assert.Equal(t, tt.expectedCost, hasher.cost)
		})
	}
}

func TestBcryptHasher_Hash_ProducesValidHash(t *testing.T) {
	hasher := NewBcryptHasher(10)
	password := "securePassword123"

	hash, err := hasher.Hash(password)

	require.NoError(t, err)
	// bcrypt hashes are always 60 characters
	assert.Len(t, hash, 60)
}

func TestBcryptHasher_Hash_CostIsAtLeast10(t *testing.T) {
	hasher := NewBcryptHasher(10)
	password := "testPassword123"

	hash, err := hasher.Hash(password)
	require.NoError(t, err)

	cost, err := bcrypt.Cost([]byte(hash))
	require.NoError(t, err)
	assert.GreaterOrEqual(t, cost, 10)
}

func TestBcryptHasher_Compare_SucceedsWithCorrectPassword(t *testing.T) {
	hasher := NewBcryptHasher(10)
	password := "correctPassword123"

	hash, err := hasher.Hash(password)
	require.NoError(t, err)

	err = hasher.Compare(hash, password)
	assert.NoError(t, err)
}

func TestBcryptHasher_Compare_FailsWithWrongPassword(t *testing.T) {
	hasher := NewBcryptHasher(10)
	password := "correctPassword123"

	hash, err := hasher.Hash(password)
	require.NoError(t, err)

	err = hasher.Compare(hash, "wrongPassword456")
	assert.Error(t, err)
}

func TestBcryptHasher_Hash_DifferentPasswordsProduceDifferentHashes(t *testing.T) {
	hasher := NewBcryptHasher(10)

	hash1, err := hasher.Hash("password1")
	require.NoError(t, err)

	hash2, err := hasher.Hash("password2")
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2)
}

func TestBcryptHasher_Hash_SamePasswordProducesDifferentHashes(t *testing.T) {
	hasher := NewBcryptHasher(10)
	password := "samePassword"

	hash1, err := hasher.Hash(password)
	require.NoError(t, err)

	hash2, err := hasher.Hash(password)
	require.NoError(t, err)

	// bcrypt uses random salt, so same password produces different hashes
	assert.NotEqual(t, hash1, hash2)

	// But both should still compare successfully
	assert.NoError(t, hasher.Compare(hash1, password))
	assert.NoError(t, hasher.Compare(hash2, password))
}

func TestBcryptHasher_ImplementsPasswordHasherInterface(t *testing.T) {
	var _ PasswordHasher = (*BcryptHasher)(nil)
}
