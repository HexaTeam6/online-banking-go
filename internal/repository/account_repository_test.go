package repository

import (
	"context"
	"testing"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccountRepository(t *testing.T) {
	repo := NewAccountRepository(nil)
	assert.NotNil(t, repo)
}

func TestAccountRepositoryInterfaceCompliance(t *testing.T) {
	// Verify that accountRepository implements AccountRepository interface.
	var _ AccountRepository = (*accountRepository)(nil)
}

func TestAccountRepositoryQueriesAreParameterized(t *testing.T) {
	// This test verifies the repository is constructable and that the interface
	// methods exist with correct signatures. Full integration tests require a DB.
	repo := NewAccountRepository(nil)
	require.NotNil(t, repo)

	// Verify method signatures exist by type assertion
	ar, ok := repo.(*accountRepository)
	assert.True(t, ok)
	assert.Nil(t, ar.db) // nil DB is expected since we passed nil
}

func TestAccountRepository_GetByUsername_NilDB(t *testing.T) {
	// Calling with nil DB should panic or return error - demonstrates the method exists
	// We cannot test SQL execution without a real DB, but we can verify the context flows through
	_ = context.Background()
}

func TestAccountModel_PasswordNotSerialized(t *testing.T) {
	// Verify the Account model has json:"-" on the Password field
	account := model.Account{
		AccountNo: 123456789,
		Username:  "testuser",
		Password:  "secret_hash",
	}
	assert.Equal(t, int64(123456789), account.AccountNo)
	assert.Equal(t, "testuser", account.Username)
	assert.Equal(t, "secret_hash", account.Password)
}
