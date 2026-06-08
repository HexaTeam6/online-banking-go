// Package handler provides HTTP handlers for the online banking API.
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

// AccountHandler handles customer profile endpoints.
type AccountHandler struct {
	accountService service.AccountService
}

// NewAccountHandler creates a new AccountHandler with the given service dependency.
func NewAccountHandler(accountService service.AccountService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

// ProfileResponse is the response DTO returned by the profile endpoint.
type ProfileResponse struct {
	AccountNo   int64  `json:"account_no"`
	FullName    string `json:"full_name"`
	Gender      string `json:"gender"`
	BirthDate   string `json:"birth_date"`
	Mobile      string `json:"mobile"`
	Email       string `json:"email"`
	HomeAddress string `json:"home_address"`
	City        string `json:"city"`
	State       string `json:"state"`
	Pincode     int    `json:"pincode"`
	AccountType string `json:"account_type"`
}

// GetProfile handles GET /api/v1/profile.
// It returns the authenticated customer's profile information.
func (h *AccountHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	accountNo := middleware.GetAccountNoFromContext(r.Context())
	if accountNo == 0 {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "authentication required")
		return
	}

	customer, address, accountType, err := h.accountService.GetProfile(r.Context(), accountNo)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	profile := ProfileResponse{
		AccountNo:   customer.AccountNo,
		FullName:    customer.FullName,
		Gender:      customer.Gender,
		BirthDate:   customer.BirthDate,
		Mobile:      customer.Mobile,
		Email:       customer.Email,
		HomeAddress: address.HomeAddress,
		City:        address.City,
		State:       address.State,
		Pincode:     address.Pincode,
		AccountType: accountType,
	}

	response.JSON(w, profile)
}

// UpdateProfile handles PUT /api/v1/profile.
// It applies partial updates to the authenticated customer's profile.
func (h *AccountHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	accountNo := middleware.GetAccountNoFromContext(r.Context())
	if accountNo == 0 {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "authentication required")
		return
	}

	var req model.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Bad Request", "invalid request body")
		return
	}

	// Validate input
	if appErr := validator.ValidateStruct(&req); appErr != nil {
		response.ValidationError(w, appErr.Message, appErr.Details)
		return
	}

	if err := h.accountService.UpdateProfile(r.Context(), accountNo, req); err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSON(w, map[string]string{"message": "profile updated successfully"})
}
