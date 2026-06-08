package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixedTime() time.Time {
	return time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
}

func setupFixedTimestamp() func() {
	original := nowFunc
	nowFunc = fixedTime
	return func() { nowFunc = original }
}

func TestJSON(t *testing.T) {
	cleanup := setupFixedTimestamp()
	defer cleanup()

	w := httptest.NewRecorder()
	data := map[string]string{"name": "John"}

	JSON(w, data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "2024-06-15T10:30:00Z", resp.Meta.Timestamp)
	assert.Nil(t, resp.Meta.Pagination)

	// Verify data is present
	dataMap, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "John", dataMap["name"])
}

func TestJSONWithStatus(t *testing.T) {
	cleanup := setupFixedTimestamp()
	defer cleanup()

	w := httptest.NewRecorder()
	data := map[string]int{"id": 42}

	JSONWithStatus(w, http.StatusCreated, data)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "2024-06-15T10:30:00Z", resp.Meta.Timestamp)
}

func TestError(t *testing.T) {
	cleanup := setupFixedTimestamp()
	defer cleanup()

	w := httptest.NewRecorder()

	Error(w, http.StatusNotFound, "Not Found", "Resource not found")

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "Not Found", resp.Error.Code)
	assert.Equal(t, "Resource not found", resp.Error.Message)
	assert.Nil(t, resp.Error.Details)
}

func TestErrorMessageTruncation(t *testing.T) {
	cleanup := setupFixedTimestamp()
	defer cleanup()

	w := httptest.NewRecorder()

	// Create a message longer than 500 chars
	longMessage := make([]byte, 600)
	for i := range longMessage {
		longMessage[i] = 'a'
	}

	Error(w, http.StatusInternalServerError, "Internal Server Error", string(longMessage))

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Len(t, resp.Error.Message, 500)
}

func TestValidationError(t *testing.T) {
	cleanup := setupFixedTimestamp()
	defer cleanup()

	w := httptest.NewRecorder()
	details := []map[string]string{
		{"field": "email", "reason": "invalid email format"},
		{"field": "password", "reason": "must be at least 8 characters"},
	}

	ValidationError(w, "Validation failed", details)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "Bad Request", resp.Error.Code)
	assert.Equal(t, "Validation failed", resp.Error.Message)
	assert.NotNil(t, resp.Error.Details)

	// Verify details structure
	detailsList, ok := resp.Error.Details.([]interface{})
	require.True(t, ok)
	assert.Len(t, detailsList, 2)
}

func TestPaginatedJSON(t *testing.T) {
	cleanup := setupFixedTimestamp()
	defer cleanup()

	w := httptest.NewRecorder()
	data := []string{"item1", "item2", "item3"}

	PaginatedJSON(w, data, 25, 2, 10)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "2024-06-15T10:30:00Z", resp.Meta.Timestamp)
	require.NotNil(t, resp.Meta.Pagination)
	assert.Equal(t, 25, resp.Meta.Pagination.TotalCount)
	assert.Equal(t, 2, resp.Meta.Pagination.Page)
	assert.Equal(t, 10, resp.Meta.Pagination.PageSize)
	assert.Equal(t, 3, resp.Meta.Pagination.TotalPages)
}

func TestPaginatedJSONEmptyPage(t *testing.T) {
	cleanup := setupFixedTimestamp()
	defer cleanup()

	w := httptest.NewRecorder()
	data := []string{}

	PaginatedJSON(w, data, 0, 1, 20)

	var resp SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	require.NotNil(t, resp.Meta.Pagination)
	assert.Equal(t, 0, resp.Meta.Pagination.TotalCount)
	assert.Equal(t, 1, resp.Meta.Pagination.Page)
	assert.Equal(t, 20, resp.Meta.Pagination.PageSize)
	assert.Equal(t, 0, resp.Meta.Pagination.TotalPages)
}

func TestPaginatedJSONTotalPagesCalculation(t *testing.T) {
	tests := []struct {
		name          string
		totalCount    int
		pageSize      int
		expectedPages int
	}{
		{"exact division", 20, 10, 2},
		{"with remainder", 21, 10, 3},
		{"single page", 5, 10, 1},
		{"zero items", 0, 10, 0},
		{"zero page size", 10, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			PaginatedJSON(w, []string{}, tt.totalCount, 1, tt.pageSize)

			var resp SuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedPages, resp.Meta.Pagination.TotalPages)
		})
	}
}

func TestTimestampISO8601Format(t *testing.T) {
	cleanup := setupFixedTimestamp()
	defer cleanup()

	w := httptest.NewRecorder()
	JSON(w, nil)

	var resp SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Verify the timestamp is a valid RFC3339 (ISO 8601) value
	_, parseErr := time.Parse(time.RFC3339, resp.Meta.Timestamp)
	assert.NoError(t, parseErr)
}

func TestErrorResponseNoDetails(t *testing.T) {
	cleanup := setupFixedTimestamp()
	defer cleanup()

	w := httptest.NewRecorder()
	Error(w, http.StatusUnauthorized, "Unauthorized", "Invalid credentials")

	body := w.Body.String()
	// "details" should not appear when it's nil (omitempty)
	assert.NotContains(t, body, "details")
}

func TestValidationErrorWithDetails(t *testing.T) {
	cleanup := setupFixedTimestamp()
	defer cleanup()

	w := httptest.NewRecorder()
	ValidationError(w, "Validation failed", []string{"field1 is required"})

	body := w.Body.String()
	// "details" should appear when present
	assert.Contains(t, body, "details")
}
