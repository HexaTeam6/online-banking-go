// Package handler provides HTTP handlers for the online banking API.
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/abdurrachmanwahed/online-banking/internal/middleware"
	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/abdurrachmanwahed/online-banking/internal/service"
	"github.com/abdurrachmanwahed/online-banking/pkg/response"
)

// DashboardHandler handles dashboard and transaction history endpoints.
type DashboardHandler struct {
	dashboardService service.DashboardService
}

// NewDashboardHandler creates a new DashboardHandler with the given DashboardService.
func NewDashboardHandler(svc service.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: svc,
	}
}

// GetDashboard handles GET /api/v1/dashboard.
// It returns the all-time transaction summary, current month summary, and current balance
// for the authenticated customer.
func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	accountNo := middleware.GetAccountNoFromContext(r.Context())
	if accountNo == 0 {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "authentication required")
		return
	}

	dashboard, err := h.dashboardService.GetDashboard(r.Context(), accountNo)
	if err != nil {
		var appErr *model.AppError
		if errors.As(err, &appErr) {
			response.Error(w, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal Server Error", "failed to retrieve dashboard data")
		return
	}

	response.JSON(w, dashboard)
}

// GetTransactions handles GET /api/v1/transactions.
// It returns a paginated list of transactions for the authenticated customer,
// ordered by date descending. Query params: page (default 1), page_size (default 10, max 50).
func (h *DashboardHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	accountNo := middleware.GetAccountNoFromContext(r.Context())
	if accountNo == 0 {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "authentication required")
		return
	}

	// Parse pagination parameters from query string.
	page := parseIntParam(r, "page", 1)
	pageSize := parseIntParam(r, "page_size", 10)

	// Enforce constraints.
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	pagination := model.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	transactions, totalCount, err := h.dashboardService.GetTransactionHistory(r.Context(), accountNo, pagination)
	if err != nil {
		var appErr *model.AppError
		if errors.As(err, &appErr) {
			response.Error(w, appErr.Status, appErr.Code, appErr.Message)
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal Server Error", "failed to retrieve transaction history")
		return
	}

	// Ensure we never return nil slice in JSON (return [] instead of null).
	if transactions == nil {
		transactions = []model.Transaction{}
	}

	response.PaginatedJSON(w, transactions, int(totalCount), page, pageSize)
}

// parseIntParam extracts an integer query parameter from the request.
// Returns the default value if the parameter is missing or invalid.
func parseIntParam(r *http.Request, key string, defaultVal int) int {
	valStr := r.URL.Query().Get(key)
	if valStr == "" {
		return defaultVal
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}

	return val
}
