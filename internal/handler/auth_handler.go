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

// AuthHandler handles authentication-related HTTP endpoints.
type AuthHandler struct {
	authService    service.AuthService
	accountService service.AccountService
}

// NewAuthHandler creates a new AuthHandler with all required dependencies.
func NewAuthHandler(authSvc service.AuthService, accountSvc service.AccountService) *AuthHandler {
	return &AuthHandler{
		authService:    authSvc,
		accountService: accountSvc,
	}
}

// customerLoginResponse extends LoginResponse with the token ID for logout tracking.
type customerLoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	TokenID   int64  `json:"token_id"`
}

// logoutRequest is the request body for customer logout.
type logoutRequest struct {
	TokenID int64 `json:"token_id" validate:"required"`
}

// CustomerLogin handles POST /api/v1/auth/login.
// It authenticates a customer and returns a JWT token in a success envelope.
func (h *AuthHandler) CustomerLogin(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Bad Request", "invalid request body")
		return
	}

	// Validate input
	if appErr := validator.ValidateStruct(&req); appErr != nil {
		response.ValidationError(w, appErr.Message, appErr.Details)
		return
	}

	// Extract client IP
	ipAddress := extractClientIP(r)

	// Call service
	loginResp, tokenID, err := h.authService.CustomerLogin(r.Context(), req, ipAddress)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	// Return token with tokenID for logout tracking
	resp := customerLoginResponse{
		Token:     loginResp.Token,
		ExpiresAt: loginResp.ExpiresAt,
		TokenID:   tokenID,
	}
	response.JSON(w, resp)
}

// CustomerLogout handles POST /api/v1/auth/logout.
// It invalidates the current token and records the logout timestamp.
func (h *AuthHandler) CustomerLogout(w http.ResponseWriter, r *http.Request) {
	// Extract token from context (set by auth middleware)
	tokenString := middleware.GetTokenFromContext(r.Context())
	if tokenString == "" {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", "missing token")
		return
	}

	// Parse the logout request to get the tokenID
	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Bad Request", "invalid request body")
		return
	}

	if appErr := validator.ValidateStruct(&req); appErr != nil {
		response.ValidationError(w, appErr.Message, appErr.Details)
		return
	}

	// Call service
	if err := h.authService.CustomerLogout(r.Context(), tokenString, req.TokenID); err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSON(w, map[string]string{"message": "logged out successfully"})
}

// Register handles POST /api/v1/auth/register.
// It creates a new customer account and returns 201 Created.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Bad Request", "invalid request body")
		return
	}

	// Validate input
	if appErr := validator.ValidateStruct(&req); appErr != nil {
		response.ValidationError(w, appErr.Message, appErr.Details)
		return
	}

	// Call service
	account, err := h.accountService.Register(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	// Return 201 with account info (excluding password)
	resp := map[string]interface{}{
		"account_no": account.AccountNo,
		"username":   account.Username,
	}
	response.JSONWithStatus(w, http.StatusCreated, resp)
}

// AdminLogin handles POST /api/v1/admin/auth/login.
// It authenticates an admin and returns a JWT token with admin role claims.
func (h *AuthHandler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Bad Request", "invalid request body")
		return
	}

	// Validate input
	if appErr := validator.ValidateStruct(&req); appErr != nil {
		response.ValidationError(w, appErr.Message, appErr.Details)
		return
	}

	// Call service
	loginResp, err := h.authService.AdminLogin(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSON(w, loginResp)
}
