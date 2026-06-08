package handler

import (
	"net/http"

	"github.com/jmoiron/sqlx"
)

// HealthHandler handles the health check endpoint.
type HealthHandler struct {
	db *sqlx.DB
}

// NewHealthHandler creates a new HealthHandler with the given database dependency.
func NewHealthHandler(db *sqlx.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// healthResponse represents the health check JSON response.
type healthResponse struct {
	Status string `json:"status"`
}

// Check handles GET /health.
// It verifies database connectivity and returns 200 with {"status":"healthy"}
// when the database is reachable, or 503 with {"status":"unhealthy"} when it is not.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	err := h.db.PingContext(r.Context())

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"unhealthy"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"healthy"}`))
}
