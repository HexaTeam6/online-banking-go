package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

// AccountTypeRepository handles account type data access.
type AccountTypeRepository interface {
	Create(ctx context.Context, tx *sqlx.Tx, accountNo int64, accountType string) error
	GetByAccountNo(ctx context.Context, accountNo int64) (string, error)
}

// accountTypeRepository is the sqlx implementation of AccountTypeRepository.
type accountTypeRepository struct {
	db *sqlx.DB
}

// NewAccountTypeRepository creates a new AccountTypeRepository backed by sqlx.
func NewAccountTypeRepository(db *sqlx.DB) AccountTypeRepository {
	return &accountTypeRepository{db: db}
}

// Create inserts a new account type record within a transaction.
func (r *accountTypeRepository) Create(ctx context.Context, tx *sqlx.Tx, accountNo int64, accountType string) error {
	query := `INSERT INTO tbl_account_type (account_no, account_type) VALUES (?, ?)`
	_, err := tx.ExecContext(ctx, query, accountNo, accountType)
	return err
}

// GetByAccountNo retrieves the account type for a given account number.
func (r *accountTypeRepository) GetByAccountNo(ctx context.Context, accountNo int64) (string, error) {
	var accountType string
	query := `SELECT account_type FROM tbl_account_type WHERE account_no = ?`
	err := r.db.GetContext(ctx, &accountType, query, accountNo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return accountType, nil
}
