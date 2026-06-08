package service

import (
	"context"
	"net/http"
	"time"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/abdurrachmanwahed/online-banking/internal/repository"
)

// FeedbackService handles customer feedback operations.
type FeedbackService interface {
	Submit(ctx context.Context, accountNo int64, req model.FeedbackRequest) error
}

// feedbackService is the implementation of FeedbackService.
type feedbackService struct {
	feedbackRepo repository.FeedbackRepository
}

// NewFeedbackService creates a new FeedbackService with the given dependencies.
func NewFeedbackService(feedbackRepo repository.FeedbackRepository) FeedbackService {
	return &feedbackService{
		feedbackRepo: feedbackRepo,
	}
}

// Submit validates the feedback request and stores it with a timestamp.
func (s *feedbackService) Submit(ctx context.Context, accountNo int64, req model.FeedbackRequest) error {
	// Service-level validation as a safety net (handler validates via struct tags)
	if len(req.Feedback) < 1 || len(req.Feedback) > 1000 {
		return model.NewAppError(
			"VALIDATION_ERROR",
			"Feedback text must be between 1 and 1000 characters",
			http.StatusBadRequest,
		)
	}

	if req.Hearts < 1 || req.Hearts > 5 {
		return model.NewAppError(
			"VALIDATION_ERROR",
			"Rating must be between 1 and 5",
			http.StatusBadRequest,
		)
	}

	feedback := &model.Feedback{
		AccountNo:    accountNo,
		FeedbackText: req.Feedback,
		Hearts:       req.Hearts,
		Time:         time.Now(),
	}

	return s.feedbackRepo.Create(ctx, feedback)
}
