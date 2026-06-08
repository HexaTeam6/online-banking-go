package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// BalanceRepository defines the interface for account balance operations.
type BalanceRepository interface {
	GetByAccountNo(ctx context.Context, accountNo int64) (int64, error)
	Credit(ctx context.Context, tx *sqlx.Tx, accountNo int64, amount int64) (int64, error)
	Debit(ctx context.Context, tx *sqlx.Tx, accountNo int64, amount int64) (int64, error)
	Create(ctx context.Context, tx *sqlx.Tx, accountNo int64, accountType string, balance int64) error
}

// balanceRepository is the sqlx implementation of BalanceRepository.
type balanceRepository struct {
	db *sqlx.DB
}

// NewBalanceRepository creates a new BalanceRepository instance.
func NewBalanceRepository(db *sqlx.DB) BalanceRepository {
	return &balanceRepository{db: db}
}

// GetByAccountNo retrieves the current balance for a specific account.
func (r *balanceRepository) GetByAccountNo(ctx context.Context, accountNo int64) (int64, error) {
	query := `SELECT balance FROM tbl_balance WHERE account_no = ?`

	var balance int64
	if err := r.db.GetContext(ctx, &balance, query, accountNo); err != nil {
		return 0, fmt.Errorf("fetching balance: %w", err)
	}

	return balance, nil
}

// Credit adds the specified amount to the account balance within a database transaction.
// Uses SELECT ... FOR UPDATE to lock the row and prevent concurrent modifications.
// Returns the new balance after crediting.
func (r *balanceRepository) Credit(ctx context.Context, tx *sqlx.Tx, accountNo int64, amount int64) (int64, error) {
	// Lock the row for update to prevent concurrent balance modifications
	selectQuery := `SELECT balance FROM tbl_balance WHERE account_no = ? FOR UPDATE`
	var currentBalance int64
	if err := tx.GetContext(ctx, &currentBalance, selectQuery, accountNo); err != nil {
		return 0, fmt.Errorf("locking balance for credit: %w", err)
	}

	newBalance := currentBalance + amount

	updateQuery := `UPDATE tbl_balance SET balance = ? WHERE account_no = ?`
	if _, err := tx.ExecContext(ctx, updateQuery, newBalance, accountNo); err != nil {
		return 0, fmt.Errorf("crediting balance: %w", err)
	}

	return newBalance, nil
}

// Debit subtracts the specified amount from the account balance within a database transaction.
// Uses SELECT ... FOR UPDATE to lock the row and prevent concurrent modifications.
// Returns the new balance after debiting.
func (r *balanceRepository) Debit(ctx context.Context, tx *sqlx.Tx, accountNo int64, amount int64) (int64, error) {
	// Lock the row for update to prevent concurrent balance modifications
	selectQuery := `SELECT balance FROM tbl_balance WHERE account_no = ? FOR UPDATE`
	var currentBalance int64
	if err := tx.GetContext(ctx, &currentBalance, selectQuery, accountNo); err != nil {
		return 0, fmt.Errorf("locking balance for debit: %w", err)
	}

	newBalance := currentBalance - amount

	updateQuery := `UPDATE tbl_balance SET balance = ? WHERE account_no = ?`
	if _, err := tx.ExecContext(ctx, updateQuery, newBalance, accountNo); err != nil {
		return 0, fmt.Errorf("debiting balance: %w", err)
	}

	return newBalance, nil
}

// Create inserts a new balance record for a new account within a database transaction.
func (r *balanceRepository) Create(ctx context.Context, tx *sqlx.Tx, accountNo int64, accountType string, balance int64) error {
	query := `INSERT INTO tbl_balance (account_no, account_type, balance) VALUES (?, ?, ?)`

	if _, err := tx.ExecContext(ctx, query, accountNo, accountType, balance); err != nil {
		return fmt.Errorf("creating balance record: %w", err)
	}

	return nil
}
