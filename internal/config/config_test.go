package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_WithAllEnvVars(t *testing.T) {
	// Set all environment variables
	t.Setenv("DB_DSN", "user:pass@tcp(localhost:3306)/testdb")
	t.Setenv("JWT_SECRET", "supersecret")
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("CORS_ORIGINS", "http://localhost:3000")
	t.Setenv("REQUEST_TIMEOUT", "45s")

	cfg := Load()

	assert.Equal(t, "user:pass@tcp(localhost:3306)/testdb", cfg.DBDSN)
	assert.Equal(t, "supersecret", cfg.JWTSecret)
	assert.Equal(t, "9090", cfg.ServerPort)
	assert.Equal(t, "http://localhost:3000", cfg.CORSOrigins)
	assert.Equal(t, 45*time.Second, cfg.RequestTimeout)
}

func TestLoad_WithDefaults(t *testing.T) {
	// Set only required variables
	t.Setenv("DB_DSN", "user:pass@tcp(localhost:3306)/testdb")
	t.Setenv("JWT_SECRET", "supersecret")
	// Explicitly unset optional vars
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("CORS_ORIGINS")
	os.Unsetenv("REQUEST_TIMEOUT")

	cfg := Load()

	assert.Equal(t, "8080", cfg.ServerPort)
	assert.Equal(t, "*", cfg.CORSOrigins)
	assert.Equal(t, 30*time.Second, cfg.RequestTimeout)
}

func TestLoad_InvalidTimeoutUsesDefault(t *testing.T) {
	t.Setenv("DB_DSN", "user:pass@tcp(localhost:3306)/testdb")
	t.Setenv("JWT_SECRET", "supersecret")
	t.Setenv("REQUEST_TIMEOUT", "not-a-duration")

	cfg := Load()

	assert.Equal(t, 30*time.Second, cfg.RequestTimeout)
}

func TestParseTimeout_ValidDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"10s", 10 * time.Second},
		{"1m", 1 * time.Minute},
		{"500ms", 500 * time.Millisecond},
		{"2m30s", 2*time.Minute + 30*time.Second},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := parseTimeout(tc.input, 30*time.Second)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestParseTimeout_EmptyReturnsDefault(t *testing.T) {
	result := parseTimeout("", 30*time.Second)
	assert.Equal(t, 30*time.Second, result)
}

func TestParseTimeout_InvalidReturnsDefault(t *testing.T) {
	result := parseTimeout("invalid", 30*time.Second)
	assert.Equal(t, 30*time.Second, result)
}
