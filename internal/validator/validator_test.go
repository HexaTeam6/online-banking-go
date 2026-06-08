package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testLoginRequest mimics a login DTO with validation tags.
type testLoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=64"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

// testFeedbackRequest mimics a feedback DTO with validation tags.
type testFeedbackRequest struct {
	Feedback string `json:"feedback" validate:"required,min=1,max=1000"`
	Hearts   int    `json:"hearts" validate:"required,min=1,max=5"`
}

// testTransferRequest mimics a transfer DTO with validation tags.
type testTransferRequest struct {
	ToAccount int64  `json:"to_account" validate:"required"`
	Amount    int64  `json:"amount" validate:"required,min=500,max=20000"`
	Purpose   string `json:"purpose" validate:"required,min=1,max=100"`
}

func TestValidateStruct_ValidInput_ReturnsNil(t *testing.T) {
	req := testLoginRequest{
		Username: "johndoe",
		Password: "securepass123",
	}

	err := ValidateStruct(req)
	assert.Nil(t, err)
}

func TestValidateStruct_MissingRequiredFields_ReturnsFieldErrors(t *testing.T) {
	req := testLoginRequest{
		Username: "",
		Password: "",
	}

	appErr := ValidateStruct(req)
	require.NotNil(t, appErr)
	assert.Equal(t, "VALIDATION_ERROR", appErr.Code)
	assert.Equal(t, "Validation failed", appErr.Message)
	assert.Equal(t, 400, appErr.Status)

	details, ok := appErr.Details.([]FieldError)
	require.True(t, ok)
	assert.Len(t, details, 2)

	fieldNames := make(map[string]string)
	for _, d := range details {
		fieldNames[d.Field] = d.Reason
	}

	assert.Equal(t, "field is required", fieldNames["username"])
	assert.Equal(t, "field is required", fieldNames["password"])
}

func TestValidateStruct_MinLengthViolation_ReturnsCorrectReason(t *testing.T) {
	req := testLoginRequest{
		Username: "ab",            // min 3
		Password: "securepass123", // valid
	}

	appErr := ValidateStruct(req)
	require.NotNil(t, appErr)

	details, ok := appErr.Details.([]FieldError)
	require.True(t, ok)
	assert.Len(t, details, 1)
	assert.Equal(t, "username", details[0].Field)
	assert.Equal(t, "must be at least 3 characters", details[0].Reason)
}

func TestValidateStruct_MaxLengthViolation_ReturnsCorrectReason(t *testing.T) {
	req := testFeedbackRequest{
		Feedback: "valid feedback",
		Hearts:   10, // max is 5
	}

	appErr := ValidateStruct(req)
	require.NotNil(t, appErr)

	details, ok := appErr.Details.([]FieldError)
	require.True(t, ok)
	assert.Len(t, details, 1)
	assert.Equal(t, "hearts", details[0].Field)
	assert.Equal(t, "must be at most 5 characters", details[0].Reason)
}

func TestValidateStruct_TransferAmountBelowMin_ReturnsError(t *testing.T) {
	req := testTransferRequest{
		ToAccount: 123456789,
		Amount:    100, // min is 500
		Purpose:   "test transfer",
	}

	appErr := ValidateStruct(req)
	require.NotNil(t, appErr)

	details, ok := appErr.Details.([]FieldError)
	require.True(t, ok)
	assert.Len(t, details, 1)
	assert.Equal(t, "amount", details[0].Field)
	assert.Equal(t, "must be at least 500 characters", details[0].Reason)
}

func TestValidateStruct_MultipleViolations_ReturnsAllErrors(t *testing.T) {
	req := testTransferRequest{
		ToAccount: 0,  // required (zero value fails)
		Amount:    0,  // required (zero value fails)
		Purpose:   "", // required
	}

	appErr := ValidateStruct(req)
	require.NotNil(t, appErr)

	details, ok := appErr.Details.([]FieldError)
	require.True(t, ok)
	assert.Len(t, details, 3)

	fieldNames := make(map[string]bool)
	for _, d := range details {
		fieldNames[d.Field] = true
	}

	assert.True(t, fieldNames["to_account"])
	assert.True(t, fieldNames["amount"])
	assert.True(t, fieldNames["purpose"])
}

func TestValidateStruct_ValidFeedback_ReturnsNil(t *testing.T) {
	req := testFeedbackRequest{
		Feedback: "Great service!",
		Hearts:   5,
	}

	err := ValidateStruct(req)
	assert.Nil(t, err)
}

func TestValidateStruct_UsesJsonTagForFieldName(t *testing.T) {
	// This tests that the validator uses the JSON tag name, not the Go field name.
	type request struct {
		FirstName string `json:"first_name" validate:"required"`
	}

	req := request{FirstName: ""}

	appErr := ValidateStruct(req)
	require.NotNil(t, appErr)

	details, ok := appErr.Details.([]FieldError)
	require.True(t, ok)
	assert.Len(t, details, 1)
	assert.Equal(t, "first_name", details[0].Field)
}
