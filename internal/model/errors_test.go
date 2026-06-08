package model

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppError_ImplementsErrorInterface(t *testing.T) {
	var err error = NewAppError("TEST", "test message", http.StatusBadRequest)
	assert.NotNil(t, err)
	assert.Equal(t, "TEST: test message", err.Error())
}

func TestAppError_ErrorWithoutDetails(t *testing.T) {
	appErr := NewAppError("CODE", "something went wrong", http.StatusInternalServerError)
	assert.Equal(t, "CODE: something went wrong", appErr.Error())
}

func TestAppError_ErrorWithDetails(t *testing.T) {
	appErr := NewAppError("VALIDATION", "validation failed", http.StatusBadRequest)
	detailed := appErr.WithDetails(map[string]string{"field": "username"})
	assert.Contains(t, detailed.Error(), "VALIDATION: validation failed")
	assert.Contains(t, detailed.Error(), "details:")
}

func TestNewAppError_SetsAllFields(t *testing.T) {
	appErr := NewAppError("MY_CODE", "my message", http.StatusNotFound)
	assert.Equal(t, "MY_CODE", appErr.Code)
	assert.Equal(t, "my message", appErr.Message)
	assert.Equal(t, http.StatusNotFound, appErr.Status)
	assert.Nil(t, appErr.Details)
}

func TestWithDetails_ReturnsNewInstance(t *testing.T) {
	original := NewAppError("CODE", "msg", http.StatusBadRequest)
	detailed := original.WithDetails([]string{"field1", "field2"})

	// Original should remain unchanged
	assert.Nil(t, original.Details)
	// New instance has details
	require.NotNil(t, detailed.Details)
	assert.Equal(t, []string{"field1", "field2"}, detailed.Details)
	// Preserves code, message, status
	assert.Equal(t, original.Code, detailed.Code)
	assert.Equal(t, original.Message, detailed.Message)
	assert.Equal(t, original.Status, detailed.Status)
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name           string
		err            *AppError
		expectedCode   string
		expectedMsg    string
		expectedStatus int
	}{
		{
			name:           "ErrInvalidCredentials",
			err:            ErrInvalidCredentials,
			expectedCode:   "INVALID_CREDENTIALS",
			expectedMsg:    "Invalid credentials",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "ErrAccountLocked",
			err:            ErrAccountLocked,
			expectedCode:   "ACCOUNT_LOCKED",
			expectedMsg:    "Account is temporarily locked",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "ErrInsufficientBalance",
			err:            ErrInsufficientBalance,
			expectedCode:   "INSUFFICIENT_BALANCE",
			expectedMsg:    "Insufficient balance",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ErrAccountNotFound",
			err:            ErrAccountNotFound,
			expectedCode:   "ACCOUNT_NOT_FOUND",
			expectedMsg:    "Account not found",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "ErrSelfTransfer",
			err:            ErrSelfTransfer,
			expectedCode:   "SELF_TRANSFER",
			expectedMsg:    "Cannot transfer to your own account",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ErrDuplicateUsername",
			err:            ErrDuplicateUsername,
			expectedCode:   "DUPLICATE_USERNAME",
			expectedMsg:    "Username already exists",
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "ErrLimitViolation",
			err:            ErrLimitViolation,
			expectedCode:   "LIMIT_VIOLATION",
			expectedMsg:    "Amount exceeds allowed limits",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedCode, tt.err.Code)
			assert.Equal(t, tt.expectedMsg, tt.err.Message)
			assert.Equal(t, tt.expectedStatus, tt.err.Status)
			assert.Nil(t, tt.err.Details)
		})
	}
}
