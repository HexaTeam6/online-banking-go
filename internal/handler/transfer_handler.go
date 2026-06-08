package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/abdurrachmanwahed/online-banking/internal/middleware"
	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/abdurrachmanwahed/online-banking/internal/service"
	"github.com/abdurrachmanwahed/online-banking/internal/validator"
	"github.com/abdurrachmanwahed/online-banking/pkg/response"
)

// TransferHandler handles HTTP requests for money transfer operations.
type TransferHandler struct {
	transferService service.TransferService
}

// NewTransferHandler creates a new TransferHandler with the given TransferService.
func NewTransferHandler(svc service.TransferService) *TransferHandler {
	return &TransferHandler{
		transferService: svc,
	}
}

// QuickTransfer handles POST /api/v1/transfers requests.
// It parses the transfer request, extracts the sender account from the context,
// validates input, and delegates to the TransferService.
func (h *TransferHandler) QuickTransfer(w http.ResponseWriter, r *http.Request) {
	var req model.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Bad Request", "invalid request body")
		return
	}

	if appErr := validator.ValidateStruct(&req); appErr != nil {
		response.ValidationError(w, appErr.Message, appErr.Details)
		return
	}

	senderAccountNo := middleware.GetAccountNoFromContext(r.Context())
	if senderAccountNo == 0 {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "authentication required")
		return
	}

	err := h.transferService.QuickTransfer(r.Context(), senderAccountNo, req)
	if err != nil {
		var appErr *model.AppError
		if errors.As(err, &appErr) {
			response.Error(w, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal Server Error", "an unexpected error occurred")
		return
	}

	response.JSON(w, map[string]string{"message": "transfer completed successfully"})
}
