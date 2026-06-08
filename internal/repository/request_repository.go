package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/jmoiron/sqlx"
)

// RequestRepository handles money request operations.
type RequestRepository interface {
	Create(ctx context.Context, request *model.MoneyRequest) error
	GetReceivedByAccountNo(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.MoneyRequest, int64, error)
	MarkAsViewed(ctx context.Context, requestID int64) error
	GetAll(ctx context.Context, pagination model.Pagination) ([]model.MoneyRequest, int64, error)
}

// requestRepository is the sqlx implementation of RequestRepository.
type requestRepository struct {
	db *sqlx.DB
}

// NewRequestRepository creates a new RequestRepository instance.
func NewRequestRepository(db *sqlx.DB) RequestRepository {
	return &requestRepository{db: db}
}

// Create inserts a new money request record.
func (r *requestRepository) Create(ctx context.Context, request *model.MoneyRequest) error {
	query := `INSERT INTO tbl_requests (account_no, to_account, amount, message, hasViewed, status, request_date)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query,
		request.AccountNo,
		request.ToAccount,
		request.Amount,
		request.Message,
		request.HasViewed,
		request.Status,
		request.RequestDate,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	request.RequestID = id

	return nil
}

// GetReceivedByAccountNo returns paginated money requests received by an account.
func (r *requestRepository) GetReceivedByAccountNo(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.MoneyRequest, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM tbl_requests WHERE to_account = ?`
	if err := r.db.GetContext(ctx, &total, countQuery, accountNo); err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.PageSize

	var requests []model.MoneyRequest
	query := `SELECT request_id, account_no, to_account, amount, message, hasViewed, status, request_date
		FROM tbl_requests
		WHERE to_account = ?
		ORDER BY request_date DESC
		LIMIT ? OFFSET ?`

	if err := r.db.SelectContext(ctx, &requests, query, accountNo, pagination.PageSize, offset); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []model.MoneyRequest{}, total, nil
		}
		return nil, 0, err
	}

	if requests == nil {
		requests = []model.MoneyRequest{}
	}

	return requests, total, nil
}

// MarkAsViewed sets the hasViewed flag to true for a request.
func (r *requestRepository) MarkAsViewed(ctx context.Context, requestID int64) error {
	query := `UPDATE tbl_requests SET hasViewed = TRUE WHERE request_id = ?`
	_, err := r.db.ExecContext(ctx, query, requestID)
	return err
}

// GetAll returns all money requests with pagination (for admin use).
func (r *requestRepository) GetAll(ctx context.Context, pagination model.Pagination) ([]model.MoneyRequest, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM tbl_requests`
	if err := r.db.GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.PageSize

	var requests []model.MoneyRequest
	query := `SELECT request_id, account_no, to_account, amount, message, hasViewed, status, request_date
		FROM tbl_requests
		ORDER BY request_date DESC
		LIMIT ? OFFSET ?`

	if err := r.db.SelectContext(ctx, &requests, query, pagination.PageSize, offset); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []model.MoneyRequest{}, total, nil
		}
		return nil, 0, err
	}

	if requests == nil {
		requests = []model.MoneyRequest{}
	}

	return requests, total, nil
}
