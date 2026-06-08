package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/jmoiron/sqlx"
)

// FeedbackRepository handles feedback data access.
type FeedbackRepository interface {
	Create(ctx context.Context, feedback *model.Feedback) error
	GetAll(ctx context.Context, pagination model.Pagination) ([]model.Feedback, int64, error)
}

// feedbackRepository is the sqlx implementation of FeedbackRepository.
type feedbackRepository struct {
	db *sqlx.DB
}

// NewFeedbackRepository creates a new FeedbackRepository instance.
func NewFeedbackRepository(db *sqlx.DB) FeedbackRepository {
	return &feedbackRepository{db: db}
}

// Create inserts a new feedback record.
func (r *feedbackRepository) Create(ctx context.Context, feedback *model.Feedback) error {
	query := `INSERT INTO tbl_feedback (account_no, feedback, hearts, time)
		VALUES (?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query,
		feedback.AccountNo,
		feedback.FeedbackText,
		feedback.Hearts,
		feedback.Time,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	feedback.FeedbackID = id

	return nil
}

// GetAll returns all feedback entries with pagination (for admin use).
func (r *feedbackRepository) GetAll(ctx context.Context, pagination model.Pagination) ([]model.Feedback, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM tbl_feedback`
	if err := r.db.GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}

	offset := (pagination.Page - 1) * pagination.PageSize

	var feedbacks []model.Feedback
	query := `SELECT feedback_id, account_no, feedback, hearts, time
		FROM tbl_feedback
		ORDER BY time DESC
		LIMIT ? OFFSET ?`

	if err := r.db.SelectContext(ctx, &feedbacks, query, pagination.PageSize, offset); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []model.Feedback{}, total, nil
		}
		return nil, 0, err
	}

	if feedbacks == nil {
		feedbacks = []model.Feedback{}
	}

	return feedbacks, total, nil
}
