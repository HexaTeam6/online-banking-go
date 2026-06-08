package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/jmoiron/sqlx"
)

// LoginHistoryRepository handles login tracking data access.
type LoginHistoryRepository interface {
	RecordLogin(ctx context.Context, accountNo int64, ipAddress string) (int64, error)
	RecordLogout(ctx context.Context, tokenID int64) error
	GetByAccountNo(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.LoginHistory, int64, error)
	GetAll(ctx context.Context, pagination model.Pagination) ([]model.LoginHistory, int64, error)
	GetLatestByAccountNo(ctx context.Context, accountNo int64) (*model.LoginHistory, error)
}

// loginHistoryRepository is the sqlx implementation of LoginHistoryRepository.
type loginHistoryRepository struct {
	db *sqlx.DB
}

// NewLoginHistoryRepository creates a new LoginHistoryRepository instance.
func NewLoginHistoryRepository(db *sqlx.DB) LoginHistoryRepository {
	return &loginHistoryRepository{db: db}
}

// RecordLogin inserts a new login record and returns the token ID.
func (r *loginHistoryRepository) RecordLogin(ctx context.Context, accountNo int64, ipAddress string) (int64, error) {
	query := `INSERT INTO tbl_login_history (account_no, login_time, ip_address)
		VALUES (?, NOW(), ?)`

	result, err := r.db.ExecContext(ctx, query, accountNo, ipAddress)
	if err != nil {
		return 0, err
	}

	tokenID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return tokenID, nil
}

// RecordLogout updates the logout timestamp for a login session.
func (r *loginHistoryRepository) RecordLogout(ctx context.Context, tokenID int64) error {
	query := `UPDATE tbl_login_history SET logout_time = NOW() WHERE token_id = ?`
	_, err := r.db.ExecContext(ctx, query, tokenID)
	return err
}

// GetByAccountNo returns paginated login history for a specific account.
func (r *loginHistoryRepository) GetByAccountNo(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.LoginHistory, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM tbl_login_history WHERE account_no = ?`
	if err := r.db.GetContext(ctx, &total, countQuery, accountNo); err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.PageSize

	var history []model.LoginHistory
	query := `SELECT token_id, account_no, login_time, logout_time, ip_address
		FROM tbl_login_history
		WHERE account_no = ?
		ORDER BY login_time DESC
		LIMIT ? OFFSET ?`

	if err := r.db.SelectContext(ctx, &history, query, accountNo, pagination.PageSize, offset); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []model.LoginHistory{}, total, nil
		}
		return nil, 0, err
	}

	if history == nil {
		history = []model.LoginHistory{}
	}

	return history, total, nil
}

// GetAll returns all login history records with pagination (for admin use).
func (r *loginHistoryRepository) GetAll(ctx context.Context, pagination model.Pagination) ([]model.LoginHistory, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM tbl_login_history`
	if err := r.db.GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.PageSize

	var history []model.LoginHistory
	query := `SELECT token_id, account_no, login_time, logout_time, ip_address
		FROM tbl_login_history
		ORDER BY login_time DESC
		LIMIT ? OFFSET ?`

	if err := r.db.SelectContext(ctx, &history, query, pagination.PageSize, offset); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []model.LoginHistory{}, total, nil
		}
		return nil, 0, err
	}

	if history == nil {
		history = []model.LoginHistory{}
	}

	return history, total, nil
}

// GetLatestByAccountNo returns the most recent login record for an account.
func (r *loginHistoryRepository) GetLatestByAccountNo(ctx context.Context, accountNo int64) (*model.LoginHistory, error) {
	var entry model.LoginHistory
	query := `SELECT token_id, account_no, login_time, logout_time, ip_address
		FROM tbl_login_history
		WHERE account_no = ?
		ORDER BY login_time DESC
		LIMIT 1`

	if err := r.db.GetContext(ctx, &entry, query, accountNo); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &entry, nil
}
