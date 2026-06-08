package security

import (
	"golang.org/x/crypto/bcrypt"
)

const minBcryptCost = 10

// PasswordHasher abstracts password hashing operations.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, plainPassword string) error
}

// BcryptHasher implements PasswordHasher using bcrypt.
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a new BcryptHasher with the given cost factor.
// If cost is less than 10, it defaults to 10.
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost < minBcryptCost {
		cost = minBcryptCost
	}
	return &BcryptHasher{cost: cost}
}

// Hash generates a bcrypt hash for the given password.
func (h *BcryptHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Compare checks whether a plaintext password matches a bcrypt hash.
// Returns nil on match, or an error on mismatch.
func (h *BcryptHasher) Compare(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}
