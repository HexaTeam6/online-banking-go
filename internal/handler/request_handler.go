package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/abdurrachmanwahed/online-banking/internal/middleware"
	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/abdurrachmanwahed/online-banking/internal/service"
	"github.com/abdurrachmanwahed/online-banking/internal/validator"
	"github.com/abdurrachmanwahed/online-banking/pkg/response"
)

// RequestHandler handles money request HTTP endpoints.
type RequestHandler struct {
	requestService service.RequestService
}

// NewRequestHandler creates a new RequestHandler with the given RequestService.
func NewRequestHandler(svc service.RequestService) *RequestHandler {
	return &RequestHandler{
		requestService: svc,
	}
}

// CreateRequest handles POST /api/v1/requests.
// It parses the CreateMoneyRequest body, validates it, and delegates to RequestService.
func (h *RequestHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	accountNo := middleware.GetAccountNoFromContext(r.Context())
	if accountNo == 0 {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "authentication required")
		return
	}

	var req model.CreateMoneyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Bad Request", "invalid request body")
		return
	}

	if appErr := validator.ValidateStruct(&req); appErr != nil {
		response.ValidationError(w, appErr.Message, appErr.Details)
		return
	}

	if err := h.requestService.CreateRequest(r.Context(), accountNo, req); err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSONWithStatus(w, http.StatusCreated, nil)
}

// GetReceivedRequests handles GET /api/v1/requests/received.
// It retrieves paginated money requests received by the authenticated user.
func (h *RequestHandler) GetReceivedRequests(w http.ResponseWriter, r *http.Request) {
	accountNo := middleware.GetAccountNoFromContext(r.Context())
	if accountNo == 0 {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "authentication required")
		return
	}

	pagination := parsePagination(r)

	requests, totalCount, err := h.requestService.GetReceivedRequests(r.Context(), accountNo, pagination)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	if requests == nil {
		requests = []model.MoneyRequest{}
	}

	response.PaginatedJSON(w, requests, int(totalCount), pagination.Page, pagination.PageSize)
}

// MarkAsViewed handles PATCH /api/v1/requests/{id}/viewed.
// It marks the specified request as viewed.
func (h *RequestHandler) MarkAsViewed(w http.ResponseWriter, r *http.Request) {
	accountNo := middleware.GetAccountNoFromContext(r.Context())
	if accountNo == 0 {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "authentication required")
		return
	}

	idParam := chi.URLParam(r, "id")
	requestID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || requestID <= 0 {
		response.Error(w, http.StatusBadRequest, "Bad Request", "invalid request ID")
		return
	}

	if err := h.requestService.MarkAsViewed(r.Context(), requestID); err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSON(w, nil)
}
