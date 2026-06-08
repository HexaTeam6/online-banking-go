package repository

import (
	"context"
	"fmt"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/jmoiron/sqlx"
)

// TransactionRepository defines the interface for transaction data access.
type TransactionRepository interface {
	Create(ctx context.Context, tx *sqlx.Tx, transaction *model.Transaction) error
	GetByAccountNo(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.Transaction, int64, error)
	GetAll(ctx context.Context, pagination model.Pagination) ([]model.Transaction, int64, error)
	GetSummary(ctx context.Context, accountNo int64) (*model.TransactionSummary, error)
	GetMonthlySummary(ctx context.Context, accountNo int64, year int, month int) (*model.TransactionSummary, error)
}

// transactionRepository is the sqlx implementation of TransactionRepository.
type transactionRepository struct {
	db *sqlx.DB
}

// NewTransactionRepository creates a new TransactionRepository instance.
func NewTransactionRepository(db *sqlx.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

// Create inserts a new transaction record within a database transaction.
func (r *transactionRepository) Create(ctx context.Context, tx *sqlx.Tx, transaction *model.Transaction) error {
	query := `
		INSERT INTO tbl_transaction (trans_date, amount, trans_type, purpose, to_account, account_no, account_bal)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := tx.ExecContext(ctx, query,
		transaction.TransDate,
		transaction.Amount,
		transaction.TransType,
		transaction.Purpose,
		transaction.ToAccount,
		transaction.AccountNo,
		transaction.AccountBal,
	)
	return err
}

// GetByAccountNo retrieves paginated transactions for a specific account, ordered by date descending.
func (r *transactionRepository) GetByAccountNo(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.Transaction, int64, error) {
	var totalCount int64
	countQuery := `SELECT COUNT(*) FROM tbl_transaction WHERE account_no = ?`
	if err := r.db.GetContext(ctx, &totalCount, countQuery, accountNo); err != nil {
		return nil, 0, fmt.Errorf("counting transactions: %w", err)
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	query := `
		SELECT trans_id, trans_date, amount, trans_type, purpose, to_account, account_no, account_bal
		FROM tbl_transaction
		WHERE account_no = ?
		ORDER BY trans_date DESC
		LIMIT ? OFFSET ?`

	var transactions []model.Transaction
	if err := r.db.SelectContext(ctx, &transactions, query, accountNo, pagination.PageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("fetching transactions: %w", err)
	}

	if transactions == nil {
		transactions = []model.Transaction{}
	}

	return transactions, totalCount, nil
}

// GetAll retrieves paginated transactions across all accounts, ordered by date descending.
func (r *transactionRepository) GetAll(ctx context.Context, pagination model.Pagination) ([]model.Transaction, int64, error) {
	var totalCount int64
	countQuery := `SELECT COUNT(*) FROM tbl_transaction`
	if err := r.db.GetContext(ctx, &totalCount, countQuery); err != nil {
		return nil, 0, fmt.Errorf("counting transactions: %w", err)
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	query := `
		SELECT trans_id, trans_date, amount, trans_type, purpose, to_account, account_no, account_bal
		FROM tbl_transaction
		ORDER BY trans_date DESC
		LIMIT ? OFFSET ?`

	var transactions []model.Transaction
	if err := r.db.SelectContext(ctx, &transactions, query, pagination.PageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("fetching transactions: %w", err)
	}

	if transactions == nil {
		transactions = []model.Transaction{}
	}

	return transactions, totalCount, nil
}

// GetSummary retrieves all-time transaction statistics for a specific account.
func (r *transactionRepository) GetSummary(ctx context.Context, accountNo int64) (*model.TransactionSummary, error) {
	query := `
		SELECT
			COUNT(*) AS transaction_count,
			COALESCE(SUM(CASE WHEN trans_type = 'CREDIT' THEN amount ELSE 0 END), 0) AS total_credit,
			COALESCE(SUM(CASE WHEN trans_type = 'DEBIT' THEN amount ELSE 0 END), 0) AS total_debit
		FROM tbl_transaction
		WHERE account_no = ?`

	var summary model.TransactionSummary
	if err := r.db.GetContext(ctx, &summary, query, accountNo); err != nil {
		return nil, fmt.Errorf("fetching transaction summary: %w", err)
	}

	return &summary, nil
}

// GetMonthlySummary retrieves transaction statistics for a specific account within a given month and year.
func (r *transactionRepository) GetMonthlySummary(ctx context.Context, accountNo int64, year int, month int) (*model.TransactionSummary, error) {
	query := `
		SELECT
			COUNT(*) AS transaction_count,
			COALESCE(SUM(CASE WHEN trans_type = 'CREDIT' THEN amount ELSE 0 END), 0) AS total_credit,
			COALESCE(SUM(CASE WHEN trans_type = 'DEBIT' THEN amount ELSE 0 END), 0) AS total_debit
		FROM tbl_transaction
		WHERE account_no = ? AND YEAR(trans_date) = ? AND MONTH(trans_date) = ?`

	var summary model.TransactionSummary
	if err := r.db.GetContext(ctx, &summary, query, accountNo, year, month); err != nil {
		return nil, fmt.Errorf("fetching monthly transaction summary: %w", err)
	}

	return &summary, nil
}
