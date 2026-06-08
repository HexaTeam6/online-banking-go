package validator

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/go-playground/validator/v10"
)

// validate is the singleton validator instance used across the application.
var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register a custom tag name function that uses the json tag for field names.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		if name == "" {
			return fld.Name
		}
		return name
	})
}

// FieldError represents a single field validation error.
type FieldError struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

// ValidateStruct validates a struct against its validate tags and returns
// a *model.AppError with field-specific error details if validation fails.
// Returns nil if validation passes.
func ValidateStruct(s interface{}) *model.AppError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return model.NewAppError("VALIDATION_ERROR", "Validation failed", http.StatusBadRequest)
	}

	details := make([]FieldError, 0, len(validationErrors))
	for _, fe := range validationErrors {
		details = append(details, FieldError{
			Field:  fe.Field(),
			Reason: buildReason(fe),
		})
	}

	return model.NewAppError("VALIDATION_ERROR", "Validation failed", http.StatusBadRequest).
		WithDetails(details)
}

// buildReason maps a validator.FieldError to a human-readable description.
func buildReason(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "field is required"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "email":
		return "must be a valid email address"
	case "numeric":
		return "must contain only numeric characters"
	case "alphanum":
		return "must contain only alphanumeric characters"
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	case "len":
		return fmt.Sprintf("must be exactly %s characters", fe.Param())
	case "gte":
		return fmt.Sprintf("must be at least %s", fe.Param())
	case "lte":
		return fmt.Sprintf("must be at most %s", fe.Param())
	case "gt":
		return fmt.Sprintf("must be greater than %s", fe.Param())
	case "lt":
		return fmt.Sprintf("must be less than %s", fe.Param())
	default:
		return fmt.Sprintf("failed on '%s' validation", fe.Tag())
	}
}
