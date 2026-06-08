package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/jmoiron/sqlx"
)

// AccountRepository defines the interface for account data access operations.
type AccountRepository interface {
	GetByUsername(ctx context.Context, username string) (*model.Account, error)
	GetByAccountNo(ctx context.Context, accountNo int64) (*model.Account, error)
	Exists(ctx context.Context, accountNo int64) (bool, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
	Create(ctx context.Context, tx *sqlx.Tx, account *model.Account) (int64, error)
	UpdatePassword(ctx context.Context, accountNo int64, hashedPassword string) error
}

// accountRepository is the sqlx-based implementation of AccountRepository.
type accountRepository struct {
	db *sqlx.DB
}

// NewAccountRepository creates a new AccountRepository backed by sqlx.
func NewAccountRepository(db *sqlx.DB) AccountRepository {
	return &accountRepository{db: db}
}

// GetByUsername retrieves an account by its username.
func (r *accountRepository) GetByUsername(ctx context.Context, username string) (*model.Account, error) {
	var account model.Account
	query := `SELECT account_no, username, password FROM tbl_account WHERE username = ?`
	err := r.db.GetContext(ctx, &account, query, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

// GetByAccountNo retrieves an account by its account number.
func (r *accountRepository) GetByAccountNo(ctx context.Context, accountNo int64) (*model.Account, error) {
	var account model.Account
	query := `SELECT account_no, username, password FROM tbl_account WHERE account_no = ?`
	err := r.db.GetContext(ctx, &account, query, accountNo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

// Exists checks whether an account with the given account number exists.
func (r *accountRepository) Exists(ctx context.Context, accountNo int64) (bool, error) {
	var count int
	query := `SELECT COUNT(1) FROM tbl_account WHERE account_no = ?`
	err := r.db.GetContext(ctx, &count, query, accountNo)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UsernameExists checks whether an account with the given username exists.
func (r *accountRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	var count int
	query := `SELECT COUNT(1) FROM tbl_account WHERE username = ?`
	err := r.db.GetContext(ctx, &count, query, username)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Create inserts a new account within the provided transaction.
// Returns the generated account number.
func (r *accountRepository) Create(ctx context.Context, tx *sqlx.Tx, account *model.Account) (int64, error) {
	query := `INSERT INTO tbl_account (account_no, username, password) VALUES (?, ?, ?)`
	_, err := tx.ExecContext(ctx, query, account.AccountNo, account.Username, account.Password)
	if err != nil {
		return 0, err
	}
	return account.AccountNo, nil
}

// UpdatePassword updates the password hash for the given account number.
func (r *accountRepository) UpdatePassword(ctx context.Context, accountNo int64, hashedPassword string) error {
	query := `UPDATE tbl_account SET password = ? WHERE account_no = ?`
	result, err := r.db.ExecContext(ctx, query, hashedPassword, accountNo)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return model.ErrAccountNotFound
	}
	return nil
}
