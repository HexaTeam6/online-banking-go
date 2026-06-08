package model

import (
	"fmt"
	"net/http"
)

// AppError represents a structured application error with HTTP status code,
// machine-readable code, human-readable message, and optional details.
type AppError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Status  int         `json:"-"`
	Details interface{} `json:"details,omitempty"`
}

// Error implements the error interface for AppError.
func (e *AppError) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("%s: %s (details: %v)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAppError creates a new AppError with the given code, message, and HTTP status.
func NewAppError(code string, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// WithDetails returns a copy of the AppError with the provided details attached.
func (e *AppError) WithDetails(details interface{}) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: e.Message,
		Status:  e.Status,
		Details: details,
	}
}

// Sentinel errors for common application error cases.
var (
	// ErrInvalidCredentials is returned when authentication fails due to wrong username or password.
	ErrInvalidCredentials = NewAppError("INVALID_CREDENTIALS", "Invalid credentials", http.StatusUnauthorized)

	// ErrAccountLocked is returned when an account is temporarily locked due to too many failed login attempts.
	ErrAccountLocked = NewAppError("ACCOUNT_LOCKED", "Account is temporarily locked", http.StatusForbidden)

	// ErrInsufficientBalance is returned when an account does not have enough funds for the requested operation.
	ErrInsufficientBalance = NewAppError("INSUFFICIENT_BALANCE", "Insufficient balance", http.StatusBadRequest)

	// ErrAccountNotFound is returned when the specified account does not exist in the system.
	ErrAccountNotFound = NewAppError("ACCOUNT_NOT_FOUND", "Account not found", http.StatusNotFound)

	// ErrSelfTransfer is returned when a user attempts to transfer money to their own account.
	ErrSelfTransfer = NewAppError("SELF_TRANSFER", "Cannot transfer to your own account", http.StatusBadRequest)

	// ErrDuplicateUsername is returned when a registration attempt uses an already-taken username.
	ErrDuplicateUsername = NewAppError("DUPLICATE_USERNAME", "Username already exists", http.StatusConflict)

	// ErrLimitViolation is returned when a transfer or request amount exceeds the allowed limits.
	ErrLimitViolation = NewAppError("LIMIT_VIOLATION", "Amount exceeds allowed limits", http.StatusBadRequest)
)
