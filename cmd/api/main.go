package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	"github.com/abdurrachmanwahed/online-banking/internal/config"
	"github.com/abdurrachmanwahed/online-banking/internal/handler"
	"github.com/abdurrachmanwahed/online-banking/internal/middleware"
	"github.com/abdurrachmanwahed/online-banking/internal/repository"
	"github.com/abdurrachmanwahed/online-banking/internal/security"
	"github.com/abdurrachmanwahed/online-banking/internal/service"
)

func main() {
	// Load configuration from environment variables.
	cfg := config.Load()

	// Initialize database connection with retry logic.
	db, err := connectDB(cfg.DBDSN, 3, 5*time.Second)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to database after retries")
		os.Exit(1)
	}
	defer db.Close()

	// Initialize security components.
	hasher := security.NewBcryptHasher(10)
	tokenManager := security.NewJWTManager(cfg.JWTSecret)

	// Initialize repositories.
	accountRepo := repository.NewAccountRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	addressRepo := repository.NewAddressRepository(db)
	accountTypeRepo := repository.NewAccountTypeRepository(db)
	balanceRepo := repository.NewBalanceRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	requestRepo := repository.NewRequestRepository(db)
	feedbackRepo := repository.NewFeedbackRepository(db)
	loginHistoryRepo := repository.NewLoginHistoryRepository(db)
	adminRepo := repository.NewAdminRepository(db)

	// Initialize services.
	authSvc := service.NewAuthService(accountRepo, adminRepo, loginHistoryRepo, hasher, tokenManager)
	accountSvc := service.NewAccountService(db, accountRepo, customerRepo, addressRepo, accountTypeRepo, balanceRepo, hasher)
	dashboardSvc := service.NewDashboardService(transactionRepo, balanceRepo)
	transferSvc := service.NewTransferService(accountRepo, balanceRepo, transactionRepo, db)
	requestSvc := service.NewRequestService(requestRepo, accountRepo)
	feedbackSvc := service.NewFeedbackService(feedbackRepo)
	adminSvc := service.NewAdminService(service.AdminServiceDeps{
		CustomerRepo:     customerRepo,
		BalanceRepo:      balanceRepo,
		TransactionRepo:  transactionRepo,
		RequestRepo:      requestRepo,
		FeedbackRepo:     feedbackRepo,
		LoginHistoryRepo: loginHistoryRepo,
		AccountRepo:      accountRepo,
		DB:               db,
	})

	// Initialize handlers.
	authHandler := handler.NewAuthHandler(authSvc, accountSvc)
	dashboardHandler := handler.NewDashboardHandler(dashboardSvc)
	transferHandler := handler.NewTransferHandler(transferSvc)
	requestHandler := handler.NewRequestHandler(requestSvc)
	accountHandler := handler.NewAccountHandler(accountSvc)
	feedbackHandler := handler.NewFeedbackHandler(feedbackSvc)
	healthHandler := handler.NewHealthHandler(db)
	loginHistoryHandler := handler.NewLoginHistoryHandler(loginHistoryRepo)
	adminHandler := handler.NewAdminHandler(adminSvc)

	// Set up chi router with middleware stack.
	r := chi.NewRouter()

	// Global middleware stack (order matters):
	// 1. Body limit — reject oversized payloads early
	// 2. Content-Type enforcement — require application/json for non-GET requests
	// 3. CORS — handle cross-origin requests
	// 4. Request logging — structured JSON logging
	// 5. Timeout — context-based request timeout
	// 6. Panic recovery — prevent server crashes
	r.Use(middleware.BodyLimitWithReject(1048576)) // 1 MB
	r.Use(middleware.RequireJSON)
	r.Use(middleware.CORS(cfg.CORSOrigins, "GET,POST,PUT,PATCH,DELETE,OPTIONS", "Content-Type,Authorization"))
	r.Use(middleware.RequestLogger())
	r.Use(middleware.Timeout(cfg.RequestTimeout))
	r.Use(chiMiddleware.Recoverer)

	// Rate limiter for auth endpoints: 5 requests per 60 seconds.
	rateLimiter := middleware.RateLimiter(5, 60*time.Second)

	// Public health check endpoint.
	r.Get("/health", healthHandler.Check)

	// API v1 routes.
	r.Route("/api/v1", func(r chi.Router) {
		// Public auth routes (rate limited).
		r.Group(func(r chi.Router) {
			r.Use(rateLimiter)
			r.Post("/auth/login", authHandler.CustomerLogin)
			r.Post("/auth/register", authHandler.Register)
			r.Post("/admin/auth/login", authHandler.AdminLogin)
		})

		// Protected customer routes (require JWT authentication).
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(tokenManager))

			r.Post("/auth/logout", authHandler.CustomerLogout)
			r.Get("/dashboard", dashboardHandler.GetDashboard)
			r.Get("/transactions", dashboardHandler.GetTransactions)
			r.Post("/transfers", transferHandler.QuickTransfer)
			r.Post("/requests", requestHandler.CreateRequest)
			r.Get("/requests/received", requestHandler.GetReceivedRequests)
			r.Patch("/requests/{id}/viewed", requestHandler.MarkAsViewed)
			r.Get("/profile", accountHandler.GetProfile)
			r.Put("/profile", accountHandler.UpdateProfile)
			r.Post("/feedback", feedbackHandler.Submit)
			r.Get("/login-history", loginHistoryHandler.GetLoginHistory)
		})

		// Protected admin routes (require JWT + admin role).
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(tokenManager))
			r.Use(middleware.RequireRole("admin"))

			r.Get("/admin/customers", adminHandler.ListCustomers)
			r.Post("/admin/balance-adjustment", adminHandler.AdjustBalance)
			r.Get("/admin/transactions", adminHandler.ListTransactions)
			r.Get("/admin/requests", adminHandler.ListRequests)
			r.Get("/admin/feedback", adminHandler.ListFeedback)
			r.Get("/admin/login-history", adminHandler.ListLoginHistory)
		})
	})

	// Start HTTP server.
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Graceful shutdown on interrupt signals.
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Info().Msg("shutting down server")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("server shutdown error")
		}
	}()

	log.Info().Str("addr", addr).Msg("starting HTTP server")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error().Err(err).Msg("server failed to start")
		os.Exit(1)
	}
}

// connectDB attempts to open a MySQL connection with retry logic.
// It retries up to maxRetries times with the given interval between attempts.
func connectDB(dsn string, maxRetries int, interval time.Duration) (*sqlx.DB, error) {
	var db *sqlx.DB
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		db, err = sqlx.Open("mysql", dsn)
		if err != nil {
			log.Warn().
				Int("attempt", attempt).
				Int("max_retries", maxRetries).
				Err(err).
				Msg("failed to open database connection")
			if attempt < maxRetries {
				time.Sleep(interval)
			}
			continue
		}

		// Verify connectivity with a ping.
		err = db.Ping()
		if err != nil {
			log.Warn().
				Int("attempt", attempt).
				Int("max_retries", maxRetries).
				Err(err).
				Msg("failed to ping database")
			db.Close()
			if attempt < maxRetries {
				time.Sleep(interval)
			}
			continue
		}

		log.Info().Int("attempt", attempt).Msg("database connection established")
		return db, nil
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}
