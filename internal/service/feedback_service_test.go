package service

import (
	"context"
	"errors"
	"testing"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockFeedbackRepository is a test double for FeedbackRepository.
type mockFeedbackRepository struct {
	createFn func(ctx context.Context, feedback *model.Feedback) error
	created  *model.Feedback
}

func (m *mockFeedbackRepository) Create(ctx context.Context, feedback *model.Feedback) error {
	m.created = feedback
	if m.createFn != nil {
		return m.createFn(ctx, feedback)
	}
	return nil
}

func (m *mockFeedbackRepository) GetAll(ctx context.Context, pagination model.Pagination) ([]model.Feedback, int64, error) {
	return nil, 0, nil
}

func TestFeedbackService_Submit_Success(t *testing.T) {
	repo := &mockFeedbackRepository{}
	svc := NewFeedbackService(repo)

	req := model.FeedbackRequest{
		Feedback: "Great banking service!",
		Hearts:   5,
	}

	err := svc.Submit(context.Background(), 123456789, req)
	require.NoError(t, err)

	assert.Equal(t, int64(123456789), repo.created.AccountNo)
	assert.Equal(t, "Great banking service!", repo.created.FeedbackText)
	assert.Equal(t, 5, repo.created.Hearts)
	assert.False(t, repo.created.Time.IsZero())
}

func TestFeedbackService_Submit_EmptyFeedback(t *testing.T) {
	repo := &mockFeedbackRepository{}
	svc := NewFeedbackService(repo)

	req := model.FeedbackRequest{
		Feedback: "",
		Hearts:   3,
	}

	err := svc.Submit(context.Background(), 123456789, req)
	require.Error(t, err)

	var appErr *model.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "VALIDATION_ERROR", appErr.Code)
	assert.Contains(t, appErr.Message, "1 and 1000 characters")
}

func TestFeedbackService_Submit_FeedbackTooLong(t *testing.T) {
	repo := &mockFeedbackRepository{}
	svc := NewFeedbackService(repo)

	longText := make([]byte, 1001)
	for i := range longText {
		longText[i] = 'a'
	}

	req := model.FeedbackRequest{
		Feedback: string(longText),
		Hearts:   3,
	}

	err := svc.Submit(context.Background(), 123456789, req)
	require.Error(t, err)

	var appErr *model.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "VALIDATION_ERROR", appErr.Code)
}

func TestFeedbackService_Submit_RatingTooLow(t *testing.T) {
	repo := &mockFeedbackRepository{}
	svc := NewFeedbackService(repo)

	req := model.FeedbackRequest{
		Feedback: "Some feedback",
		Hearts:   0,
	}

	err := svc.Submit(context.Background(), 123456789, req)
	require.Error(t, err)

	var appErr *model.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "VALIDATION_ERROR", appErr.Code)
	assert.Contains(t, appErr.Message, "1 and 5")
}

func TestFeedbackService_Submit_RatingTooHigh(t *testing.T) {
	repo := &mockFeedbackRepository{}
	svc := NewFeedbackService(repo)

	req := model.FeedbackRequest{
		Feedback: "Some feedback",
		Hearts:   6,
	}

	err := svc.Submit(context.Background(), 123456789, req)
	require.Error(t, err)

	var appErr *model.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "VALIDATION_ERROR", appErr.Code)
}

func TestFeedbackService_Submit_RepositoryError(t *testing.T) {
	repo := &mockFeedbackRepository{
		createFn: func(ctx context.Context, feedback *model.Feedback) error {
			return errors.New("database connection error")
		},
	}
	svc := NewFeedbackService(repo)

	req := model.FeedbackRequest{
		Feedback: "Good service",
		Hearts:   4,
	}

	err := svc.Submit(context.Background(), 123456789, req)
	require.Error(t, err)
	assert.Equal(t, "database connection error", err.Error())
}

func makeBytes(n int, ch byte) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = ch
	}
	return b
}

func TestFeedbackService_Submit_BoundaryValues(t *testing.T) {
	tests := []struct {
		name    string
		req     model.FeedbackRequest
		wantErr bool
	}{
		{
			name:    "min valid feedback (1 char) and min rating (1)",
			req:     model.FeedbackRequest{Feedback: "x", Hearts: 1},
			wantErr: false,
		},
		{
			name:    "max valid rating (5)",
			req:     model.FeedbackRequest{Feedback: "Good", Hearts: 5},
			wantErr: false,
		},
		{
			name:    "exactly 1000 chars feedback",
			req:     model.FeedbackRequest{Feedback: string(makeBytes(1000, 'a')), Hearts: 3},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockFeedbackRepository{}
			svc := NewFeedbackService(repo)

			err := svc.Submit(context.Background(), 123456789, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
