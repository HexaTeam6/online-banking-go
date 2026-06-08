package config

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config holds all application configuration values loaded from environment variables.
type Config struct {
	// DBDSN is the database connection string (required).
	DBDSN string

	// JWTSecret is the secret key used for signing JWT tokens (required).
	JWTSecret string

	// ServerPort is the port the HTTP server listens on (default: "8080").
	ServerPort string

	// CORSOrigins defines allowed CORS origins (default: "*").
	CORSOrigins string

	// RequestTimeout is the default timeout for HTTP requests (default: 30s).
	RequestTimeout time.Duration
}

// Load reads configuration from environment variables.
// It exits with a non-zero status code if required variables (DB_DSN, JWT_SECRET)
// are missing or empty.
func Load() *Config {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	var missingVars []string

	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		missingVars = append(missingVars, "DB_DSN")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		missingVars = append(missingVars, "JWT_SECRET")
	}

	if len(missingVars) > 0 {
		for _, v := range missingVars {
			log.Error().Str("variable", v).Msg("required environment variable is missing")
		}
		os.Exit(1)
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "*"
	}

	requestTimeout := parseTimeout(os.Getenv("REQUEST_TIMEOUT"), 30*time.Second)

	return &Config{
		DBDSN:          dbDSN,
		JWTSecret:      jwtSecret,
		ServerPort:     serverPort,
		CORSOrigins:    corsOrigins,
		RequestTimeout: requestTimeout,
	}
}

// parseTimeout attempts to parse a duration string.
// Returns the default value if the input is empty or invalid.
func parseTimeout(value string, defaultVal time.Duration) time.Duration {
	if value == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		log.Warn().
			Str("value", value).
			Str("default", defaultVal.String()).
			Msg("invalid REQUEST_TIMEOUT value, using default")
		return defaultVal
	}
	return d
}
