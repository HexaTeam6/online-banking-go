package handler

import (
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/abdurrachmanwahed/online-banking/pkg/response"
)

// parsePagination extracts pagination parameters from query string.
// Defaults: page=1, page_size=20. Maximum page_size=100.
func parsePagination(r *http.Request) model.Pagination {
	page := 1
	pageSize := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	// Enforce maximum page size
	if pageSize > 100 {
		pageSize = 100
	}

	return model.Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

// handleServiceError maps service-layer errors to appropriate HTTP responses.
// It checks if the error is an *AppError and returns the appropriate status code
// and error body. If the AppError has details (e.g., validation errors), it uses
// the ValidationError response format.
func handleServiceError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*model.AppError); ok {
		if appErr.Details != nil {
			response.ValidationError(w, appErr.Message, appErr.Details)
			return
		}
		response.Error(w, appErr.Status, appErr.Code, appErr.Message)
		return
	}
	// Default to internal server error for unexpected errors
	response.Error(w, http.StatusInternalServerError, "Internal Server Error", "an unexpected error occurred")
}

// extractClientIP parses the client IP address from the request,
// checking X-Forwarded-For, X-Real-IP headers, then falling back to RemoteAddr.
func extractClientIP(r *http.Request) string {
	// Check X-Forwarded-For (may contain comma-separated list of IPs)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// The first IP in the list is the original client
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr (may include port)
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// RemoteAddr might not have a port
		return r.RemoteAddr
	}
	return ip
}
