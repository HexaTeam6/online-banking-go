package handler

import (
	"encoding/json"
	"net/http"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/abdurrachmanwahed/online-banking/internal/service"
	"github.com/abdurrachmanwahed/online-banking/internal/validator"
	"github.com/abdurrachmanwahed/online-banking/pkg/response"
)

// AdminHandler handles HTTP requests for admin operations including
// customer management, balance adjustments, and viewing system-wide data.
type AdminHandler struct {
	adminService service.AdminService
}

// NewAdminHandler creates a new AdminHandler with the given AdminService dependency.
func NewAdminHandler(svc service.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: svc,
	}
}

// ListCustomers handles GET /api/v1/admin/customers.
// Returns a paginated list of all customer profiles.
func (h *AdminHandler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	pagination := parsePagination(r)

	customers, total, err := h.adminService.ListCustomers(r.Context(), pagination)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal Server Error", "failed to list customers")
		return
	}

	response.PaginatedJSON(w, customers, int(total), pagination.Page, pagination.PageSize)
}

// AdjustBalance handles POST /api/v1/admin/balance-adjustment.
// Performs a credit or debit operation on a customer account.
func (h *AdminHandler) AdjustBalance(w http.ResponseWriter, r *http.Request) {
	var req model.BalanceAdjustmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Bad Request", "invalid request body")
		return
	}

	if appErr := validator.ValidateStruct(&req); appErr != nil {
		response.ValidationError(w, appErr.Message, appErr.Details)
		return
	}

	err := h.adminService.AdjustBalance(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSON(w, map[string]string{"message": "balance adjusted successfully"})
}

// ListTransactions handles GET /api/v1/admin/transactions.
// Returns a paginated list of all transactions across all accounts.
func (h *AdminHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	pagination := parsePagination(r)

	transactions, total, err := h.adminService.ListTransactions(r.Context(), pagination)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal Server Error", "failed to list transactions")
		return
	}

	response.PaginatedJSON(w, transactions, int(total), pagination.Page, pagination.PageSize)
}

// ListRequests handles GET /api/v1/admin/requests.
// Returns a paginated list of all money requests across all accounts.
func (h *AdminHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	pagination := parsePagination(r)

	requests, total, err := h.adminService.ListRequests(r.Context(), pagination)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal Server Error", "failed to list requests")
		return
	}

	response.PaginatedJSON(w, requests, int(total), pagination.Page, pagination.PageSize)
}

// ListFeedback handles GET /api/v1/admin/feedback.
// Returns a paginated list of all customer feedback entries.
func (h *AdminHandler) ListFeedback(w http.ResponseWriter, r *http.Request) {
	pagination := parsePagination(r)

	feedback, total, err := h.adminService.ListFeedback(r.Context(), pagination)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal Server Error", "failed to list feedback")
		return
	}

	response.PaginatedJSON(w, feedback, int(total), pagination.Page, pagination.PageSize)
}

// ListLoginHistory handles GET /api/v1/admin/login-history.
// Returns a paginated list of all login/logout records.
func (h *AdminHandler) ListLoginHistory(w http.ResponseWriter, r *http.Request) {
	pagination := parsePagination(r)

	history, total, err := h.adminService.ListLoginHistory(r.Context(), pagination)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal Server Error", "failed to list login history")
		return
	}

	response.PaginatedJSON(w, history, int(total), pagination.Page, pagination.PageSize)
}
