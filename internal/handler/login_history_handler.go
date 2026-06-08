package handler

import (
	"net/http"

	"github.com/abdurrachmanwahed/online-banking/internal/middleware"
	"github.com/abdurrachmanwahed/online-banking/internal/repository"
	"github.com/abdurrachmanwahed/online-banking/pkg/response"
)

// LoginHistoryHandler handles login history endpoints.
type LoginHistoryHandler struct {
	loginHistoryRepo repository.LoginHistoryRepository
}

// NewLoginHistoryHandler creates a new LoginHistoryHandler with the given repository dependency.
func NewLoginHistoryHandler(loginHistoryRepo repository.LoginHistoryRepository) *LoginHistoryHandler {
	return &LoginHistoryHandler{
		loginHistoryRepo: loginHistoryRepo,
	}
}

// GetLoginHistory handles GET /api/v1/login-history.
// It returns a paginated list of login sessions for the authenticated customer.
func (h *LoginHistoryHandler) GetLoginHistory(w http.ResponseWriter, r *http.Request) {
	accountNo := middleware.GetAccountNoFromContext(r.Context())
	if accountNo == 0 {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "authentication required")
		return
	}

	pagination := parsePagination(r)

	history, total, err := h.loginHistoryRepo.GetByAccountNo(r.Context(), accountNo, pagination)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal Server Error", "failed to retrieve login history")
		return
	}

	response.PaginatedJSON(w, history, int(total), pagination.Page, pagination.PageSize)
}
