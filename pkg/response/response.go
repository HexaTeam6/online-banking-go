// Package response provides shared response envelope types and helper functions
// for consistent API responses across all handlers.
package response

import (
	"encoding/json"
	"math"
	"net/http"
	"time"
)

// SuccessResponse wraps all successful API responses.
type SuccessResponse struct {
	Data interface{}  `json:"data"`
	Meta ResponseMeta `json:"meta"`
}

// ErrorResponse wraps all error API responses.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody contains error details.
type ErrorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ResponseMeta contains timestamp and optional pagination.
type ResponseMeta struct {
	Timestamp  string          `json:"timestamp"`
	Pagination *PaginationMeta `json:"pagination,omitempty"`
}

// PaginationMeta holds pagination metadata for paginated responses.
type PaginationMeta struct {
	TotalCount int `json:"total_count"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalPages int `json:"total_pages"`
}

// nowFunc is a package-level variable to allow overriding in tests.
var nowFunc = func() time.Time {
	return time.Now().UTC()
}

// timestamp returns the current time in ISO 8601 format.
func timestamp() string {
	return nowFunc().Format(time.RFC3339)
}

// JSON writes a success response with HTTP 200 and the given data payload.
func JSON(w http.ResponseWriter, data interface{}) {
	resp := SuccessResponse{
		Data: data,
		Meta: ResponseMeta{
			Timestamp: timestamp(),
		},
	}
	writeJSON(w, http.StatusOK, resp)
}

// JSONWithStatus writes a success response with the given HTTP status code and data payload.
func JSONWithStatus(w http.ResponseWriter, statusCode int, data interface{}) {
	resp := SuccessResponse{
		Data: data,
		Meta: ResponseMeta{
			Timestamp: timestamp(),
		},
	}
	writeJSON(w, statusCode, resp)
}

// Error writes an error response with the given HTTP status code, error code, and message.
func Error(w http.ResponseWriter, statusCode int, code string, message string) {
	resp := ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: truncateMessage(message, 500),
		},
	}
	writeJSON(w, statusCode, resp)
}

// ValidationError writes a 400 Bad Request response with validation details.
// The details parameter typically contains a slice of field-level validation errors.
func ValidationError(w http.ResponseWriter, message string, details interface{}) {
	resp := ErrorResponse{
		Error: ErrorBody{
			Code:    "Bad Request",
			Message: truncateMessage(message, 500),
			Details: details,
		},
	}
	writeJSON(w, http.StatusBadRequest, resp)
}

// PaginatedJSON writes a success response with HTTP 200, the given data payload,
// and pagination metadata.
func PaginatedJSON(w http.ResponseWriter, data interface{}, totalCount int, page int, pageSize int) {
	totalPages := 0
	if pageSize > 0 {
		totalPages = int(math.Ceil(float64(totalCount) / float64(pageSize)))
	}

	resp := SuccessResponse{
		Data: data,
		Meta: ResponseMeta{
			Timestamp: timestamp(),
			Pagination: &PaginationMeta{
				TotalCount: totalCount,
				Page:       page,
				PageSize:   pageSize,
				TotalPages: totalPages,
			},
		},
	}
	writeJSON(w, http.StatusOK, resp)
}

// writeJSON serializes the payload to JSON and writes it to the response writer.
func writeJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// If encoding fails, attempt to write a minimal error.
		// This is a last-resort fallback.
		http.Error(w, `{"error":{"code":"Internal Server Error","message":"failed to encode response"}}`, http.StatusInternalServerError)
	}
}

// truncateMessage ensures the message does not exceed maxLen characters.
func truncateMessage(message string, maxLen int) string {
	if len(message) <= maxLen {
		return message
	}
	return message[:maxLen]
}
