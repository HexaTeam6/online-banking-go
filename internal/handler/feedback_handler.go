package handler

import (
	"encoding/json"
	"net/http"

	"github.com/abdurrachmanwahed/online-banking/internal/middleware"
	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/abdurrachmanwahed/online-banking/internal/service"
	"github.com/abdurrachmanwahed/online-banking/internal/validator"
	"github.com/abdurrachmanwahed/online-banking/pkg/response"
)

// FeedbackHandler handles customer feedback endpoints.
type FeedbackHandler struct {
	feedbackService service.FeedbackService
}

// NewFeedbackHandler creates a new FeedbackHandler with the given service dependency.
func NewFeedbackHandler(feedbackService service.FeedbackService) *FeedbackHandler {
	return &FeedbackHandler{
		feedbackService: feedbackService,
	}
}

// Submit handles POST /api/v1/feedback.
// It validates and stores customer feedback.
func (h *FeedbackHandler) Submit(w http.ResponseWriter, r *http.Request) {
	accountNo := middleware.GetAccountNoFromContext(r.Context())
	if accountNo == 0 {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "authentication required")
		return
	}

	var req model.FeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Bad Request", "invalid request body")
		return
	}

	// Validate input
	if appErr := validator.ValidateStruct(&req); appErr != nil {
		response.ValidationError(w, appErr.Message, appErr.Details)
		return
	}

	if err := h.feedbackService.Submit(r.Context(), accountNo, req); err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSONWithStatus(w, http.StatusCreated, map[string]string{"message": "feedback submitted successfully"})
}
