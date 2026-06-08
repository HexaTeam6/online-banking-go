# Implementation Plan: PHP to Golang Conversion

## Overview

This implementation plan converts the PHP online banking monolith into a modern Go REST API backend following clean architecture principles. The plan progresses from project scaffolding through domain models, repositories, services, handlers, middleware, migrations, and comprehensive testing. Each task builds incrementally, ensuring no orphaned code.

## Tasks

- [x] 1. Project initialization and core infrastructure
  - [x] 1.1 Initialize Go module and install dependencies
    - Run `go mod init` with the appropriate module path
    - Add all dependencies: chi, sqlx, go-sql-driver/mysql, golang-jwt/jwt/v5, x/crypto/bcrypt, go-playground/validator/v10, zerolog, x/time/rate, golang-migrate/migrate, gopter, testify, go.uber.org/mock
    - Create the directory structure: /cmd/api, /internal/config, /internal/handler, /internal/middleware, /internal/service, /internal/repository, /internal/model, /internal/validator, /internal/security, /pkg/response, /migrations, /test/integration, /test/mock
    - _Requirements: 11.7_

  - [x] 1.2 Implement configuration loader
    - Create `/internal/config/config.go` with struct for all environment variables (DB_DSN, JWT_SECRET, SERVER_PORT, CORS_ORIGINS, REQUEST_TIMEOUT)
    - Implement `Load()` function that reads from environment variables
    - Exit with non-zero status and error log if required variables (DB_DSN, JWT_SECRET) are missing
    - _Requirements: 11.1, 11.2_

  - [x] 1.3 Implement shared response envelope types
    - Create `/pkg/response/response.go` with SuccessResponse, ErrorResponse, ErrorBody, ResponseMeta, PaginationMeta structs
    - Implement helper functions: `JSON()`, `Error()`, `ValidationError()`, `PaginatedJSON()`
    - Ensure all responses include timestamp in ISO 8601 format
    - _Requirements: 13.1, 13.2, 13.3, 13.4, 13.5_

  - [x] 1.4 Implement custom AppError type and error utilities
    - Create error types in `/internal/model/errors.go` with AppError struct (Code, Message, Status, Details)
    - Define sentinel errors for common cases: ErrInvalidCredentials, ErrAccountLocked, ErrInsufficientBalance, ErrAccountNotFound, ErrSelfTransfer, ErrDuplicateUsername, ErrLimitViolation
    - _Requirements: 13.2, 13.3_

- [x] 2. Domain models and validation
  - [x] 2.1 Define domain models
    - Create `/internal/model/account.go`, `customer.go`, `transaction.go`, `request.go`, `feedback.go`, `admin.go`, `login_history.go`, `pagination.go`
    - Implement all domain structs with proper `db` and `json` tags
    - Ensure Password fields have `json:"-"` to prevent serialization
    - _Requirements: 2.1, 3.3, 4.7, 5.1, 7.1, 14.1_

  - [x] 2.2 Define request/response DTOs
    - Create `/internal/model/dto.go` with LoginRequest, RegisterRequest, TransferRequest, CreateMoneyRequest, UpdateProfileRequest, FeedbackRequest, BalanceAdjustmentRequest, DashboardResponse, TransactionSummary
    - Add `validate` struct tags for all fields with constraints from the design
    - _Requirements: 1.5, 2.4, 2.5, 2.8, 2.9, 4.2, 5.2, 5.5, 6.3, 7.2, 7.3, 10.2_

  - [x] 2.3 Implement custom validator
    - Create `/internal/validator/validator.go` with go-playground/validator instance
    - Register custom validation rules if needed (e.g., date format)
    - Implement `ValidateStruct()` that returns field-specific errors compatible with AppError
    - _Requirements: 10.2, 2.4, 2.5_

  - [ ]* 2.4 Write property test for input validation (Property 7)
    - **Property 7: Input validation rejects all constraint violations with field identification**
    - Generate random strings violating each constraint and verify rejection with correct field identification
    - **Validates: Requirements 2.4, 2.5, 2.8, 2.9, 6.3, 7.2, 7.3, 10.2**

- [x] 3. Security layer (password hashing and JWT)
  - [x] 3.1 Implement password hasher
    - Create `/internal/security/hasher.go` with `PasswordHasher` interface and `BcryptHasher` struct
    - Implement `Hash()` with bcrypt cost factor of 10 minimum
    - Implement `Compare()` using bcrypt.CompareHashAndPassword
    - Use constructor injection pattern: `NewBcryptHasher(cost int)`
    - _Requirements: 1.6, 2.2_

  - [ ]* 3.2 Write property test for password hashing (Property 6)
    - **Property 6: Registration password hashing**
    - Generate random passwords, hash them, and verify: hash length is 60, cost >= 10, Compare succeeds with original, Compare fails with different password
    - **Validates: Requirements 2.2, 1.6**

  - [x] 3.3 Implement JWT token manager
    - Create `/internal/security/token.go` with `TokenManager` interface and `JWTManager` struct
    - Implement `Generate()` with role-based expiry (15 min customer, 30 min admin)
    - Implement `Validate()` with expiration and signature checks
    - Implement `Invalidate()` using in-memory token blacklist
    - Include `TokenClaims` struct with AccountNo, Role, ExpiresAt
    - _Requirements: 1.1, 1.8, 8.1, 8.3, 8.6_

  - [ ]* 3.4 Write property test for JWT token (Property 1, Property 5)
    - **Property 1: Valid credentials produce correctly-scoped JWT**
    - **Property 5: Expired or invalidated tokens produce auth error**
    - Generate valid claims, verify tokens have correct role and expiry; verify expired/invalidated tokens fail validation
    - **Validates: Requirements 1.1, 1.8, 8.1, 8.3, 8.6**

- [x] 4. Repository layer
  - [x] 4.1 Implement account repository
    - Create `/internal/repository/account_repository.go` with interface and sqlx implementation
    - Implement: GetByUsername, GetByAccountNo, Exists, UsernameExists, Create (with *sqlx.Tx), UpdatePassword
    - Use parameterized queries exclusively
    - _Requirements: 1.1, 2.1, 2.3, 10.1_

  - [x] 4.2 Implement customer and address repositories
    - Create `/internal/repository/customer_repository.go` and `address_repository.go`
    - CustomerRepository: GetByAccountNo, Create (with *sqlx.Tx), Update, ListAll with pagination
    - AddressRepository: Create (with *sqlx.Tx), GetByAccountNo, Update
    - _Requirements: 2.1, 6.1, 6.2, 9.1_

  - [x] 4.3 Implement transaction and balance repositories
    - Create `/internal/repository/transaction_repository.go` and balance operations within it
    - TransactionRepository: Create (with *sqlx.Tx), GetByAccountNo, GetAll, GetSummary, GetMonthlySummary
    - BalanceRepository: GetByAccountNo, Credit (with *sqlx.Tx), Debit (with *sqlx.Tx), Create (with *sqlx.Tx)
    - All balance operations use `SELECT ... FOR UPDATE` within transactions
    - _Requirements: 3.1, 3.2, 3.3, 4.1, 4.7, 9.2, 9.5_

  - [x] 4.4 Implement request, feedback, and login history repositories
    - Create `/internal/repository/request_repository.go`, `feedback_repository.go`, `login_history_repository.go`
    - RequestRepository: Create, GetReceivedByAccountNo, MarkAsViewed, GetAll
    - FeedbackRepository: Create, GetAll
    - LoginHistoryRepository: RecordLogin, RecordLogout, GetByAccountNo, GetAll, GetLatestByAccountNo
    - _Requirements: 5.1, 5.6, 5.7, 7.1, 9.6, 9.7, 14.1, 14.2, 14.3_

  - [x] 4.5 Implement admin and account type repositories
    - Create `/internal/repository/admin_repository.go` and `account_type_repository.go`
    - AdminRepository: GetByID, GetByEmail
    - AccountTypeRepository: Create (with *sqlx.Tx), GetByAccountNo
    - _Requirements: 8.1, 2.1_

- [x] 5. Checkpoint - Ensure all repository code compiles
  - Ensure all tests pass, ask the user if questions arise.

- [x] 6. Service layer - Authentication
  - [x] 6.1 Implement auth service
    - Create `/internal/service/auth_service.go` with AuthService interface and implementation
    - Inject AccountRepository, AdminRepository, LoginHistoryRepository, PasswordHasher, TokenManager
    - Implement CustomerLogin: validate credentials, check lockout, generate JWT, record login history with IP
    - Implement CustomerLogout: invalidate token, record logout time
    - Implement AdminLogin: validate admin credentials, generate JWT with admin role
    - Implement IsAccountLocked: check in-memory lockout map (5 failures in 15 min → 30 min lock)
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.7, 1.8, 8.1, 8.2_

  - [ ]* 6.2 Write property tests for auth service (Properties 2, 3, 4)
    - **Property 2: Invalid credentials produce generic error**
    - **Property 3: Missing authentication fields produce field-specific validation errors**
    - **Property 4: Account lockout after consecutive failures**
    - Mock repositories and hasher; verify error messages don't reveal field identity; verify lockout triggers at exactly 5 failures
    - **Validates: Requirements 1.2, 1.5, 1.7, 8.2, 8.5**

  - [ ]* 6.3 Write unit tests for auth service
    - Test successful customer login flow
    - Test successful admin login flow
    - Test logout records timestamp
    - Test login records IP address
    - Test expired token rejection
    - _Requirements: 1.1, 1.3, 1.4, 1.8, 8.1, 8.6_

- [x] 7. Service layer - Account and Dashboard
  - [x] 7.1 Implement account service
    - Create `/internal/service/account_service.go` with AccountService interface and implementation
    - Inject AccountRepository, CustomerRepository, AddressRepository, AccountTypeRepository, BalanceRepository, PasswordHasher
    - Implement Register: validate input, check username uniqueness, generate 9-digit account number, hash password, create all records in single transaction, initialize balance to 0
    - Implement GetProfile: fetch customer + address + account type
    - Implement UpdateProfile: partial updates for allowed fields only
    - _Requirements: 2.1, 2.2, 2.3, 2.6, 2.7, 2.10, 6.1, 6.2, 6.3, 6.5_

  - [ ]* 7.2 Write property tests for account service (Properties 8, 9, 23)
    - **Property 8: Account number generation produces unique 9-digit values**
    - **Property 9: New account balance initialization**
    - **Property 23: Profile update only changes allowed fields**
    - Generate random registrations, verify 9-digit uniqueness and zero balance; verify profile updates don't change account_no/username/account_type
    - **Validates: Requirements 2.6, 2.7, 6.2**

  - [x] 7.3 Implement dashboard service
    - Create `/internal/service/dashboard_service.go` with DashboardService interface and implementation
    - Inject TransactionRepository, BalanceRepository
    - Implement GetDashboard: aggregate all-time stats + current month stats + current balance
    - Implement GetTransactionHistory: paginated list ordered by date descending
    - _Requirements: 3.1, 3.2, 3.3, 3.5_

  - [ ]* 7.4 Write property tests for dashboard (Properties 10, 11, 12)
    - **Property 10: Dashboard aggregation correctness**
    - **Property 11: Monthly statistics filter correctness**
    - **Property 12: Pagination correctness**
    - Generate random transaction sets, verify sums and counts; verify pagination metadata correctness
    - **Validates: Requirements 3.1, 3.2, 3.3, 3.6, 13.4, 13.5**

- [x] 8. Service layer - Transfer
  - [x] 8.1 Implement transfer service
    - Create `/internal/service/transfer_service.go` with TransferService interface and implementation
    - Inject AccountRepository, BalanceRepository, TransactionRepository
    - Implement QuickTransfer: validate amount limits (500-20000), verify destination exists, prevent self-transfer, check sufficient balance, execute atomic debit+credit in single DB transaction, record both transaction entries
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7, 4.8, 4.9_

  - [ ]* 8.2 Write property tests for transfer service (Properties 13, 14, 15, 16, 17, 18)
    - **Property 13: Transfer atomicity and balance invariant**
    - **Property 14: Transfer amount limits enforcement**
    - **Property 15: Transfer to non-existent account rejected**
    - **Property 16: Self-transfer prevention**
    - **Property 17: Insufficient balance prevents transfer**
    - **Property 18: Invalid account number format rejected**
    - Generate random transfer scenarios; verify balance sum invariant, rejection conditions, and atomicity
    - **Validates: Requirements 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7**

  - [ ]* 8.3 Write unit tests for transfer service
    - Test successful transfer updates both balances
    - Test transaction records created for both sender and receiver
    - Test rollback on DB failure
    - _Requirements: 4.1, 4.7, 4.8_

- [x] 9. Service layer - Request, Feedback, Admin
  - [x] 9.1 Implement request service
    - Create `/internal/service/request_service.go` with RequestService interface and implementation
    - Inject RequestRepository, AccountRepository
    - Implement CreateRequest: validate amount limits, verify target account exists, prevent self-request, create with PENDING status
    - Implement GetReceivedRequests: paginated list
    - Implement MarkAsViewed: set hasViewed flag
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7_

  - [ ]* 9.2 Write property tests for request service (Properties 19, 20, 21, 22)
    - **Property 19: Money request creation with valid inputs**
    - **Property 20: Money request amount limits enforcement**
    - **Property 21: Money request self-request prevention**
    - **Property 22: Money request message length enforcement**
    - Generate random request scenarios; verify status is PENDING, rejection on invalid amounts/self-request/message length
    - **Validates: Requirements 5.1, 5.2, 5.4, 5.5**

  - [x] 9.3 Implement feedback service
    - Create `/internal/service/feedback_service.go` with FeedbackService interface and implementation
    - Inject FeedbackRepository
    - Implement Submit: validate text length (1-1000) and rating (1-5), store with timestamp
    - _Requirements: 7.1, 7.2, 7.3_

  - [ ]* 9.4 Write property test for feedback service (Property 24)
    - **Property 24: Feedback storage round-trip**
    - Generate random valid feedback, submit, verify stored data matches input exactly
    - **Validates: Requirements 7.1**

  - [x] 9.5 Implement admin service
    - Create `/internal/service/admin_service.go` with AdminService interface and implementation
    - Inject CustomerRepository, BalanceRepository, TransactionRepository, RequestRepository, FeedbackRepository, LoginHistoryRepository, AccountRepository
    - Implement ListCustomers, AdjustBalance (credit/debit with transaction record), ListTransactions, ListRequests, ListFeedback, ListLoginHistory
    - AdjustBalance: verify account exists, check balance for debit, execute atomically
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6, 9.7, 9.8_

  - [ ]* 9.6 Write property test for admin balance adjustment (Property 25)
    - **Property 25: Admin balance adjustment correctness**
    - Generate random adjustment scenarios; verify resulting balance = previous ± amount; verify debit rejection when insufficient
    - **Validates: Requirements 9.2, 9.3, 9.4**

- [x] 10. Checkpoint - Ensure all service layer tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 11. Middleware layer
  - [ ] 11.1 Implement JWT authentication middleware
    - Create `/internal/middleware/auth.go`
    - Extract Bearer token from Authorization header
    - Validate token using TokenManager, inject claims into request context
    - Return 401 for missing/invalid/expired tokens
    - Implement role-checking middleware (RequireRole) that returns 403 for insufficient permissions
    - _Requirements: 1.8, 3.4, 4.9, 5.8, 6.4, 7.4, 8.4, 8.6_

  - [ ] 11.2 Implement rate limiting middleware
    - Create `/internal/middleware/ratelimit.go`
    - Use `golang.org/x/time/rate` with per-IP token bucket
    - Configure 5 requests per 60-second window for auth endpoints
    - Return 429 with Retry-After header when limit exceeded
    - _Requirements: 10.4, 10.5_

  - [ ]* 11.3 Write property test for rate limiting (Property 26)
    - **Property 26: Rate limiting enforcement**
    - Simulate N requests from same IP; verify first 5 pass and subsequent get 429 with Retry-After
    - **Validates: Requirements 10.4, 10.5**

  - [ ] 11.4 Implement CORS, content-type, body-limit, logging, and timeout middleware
    - Create `/internal/middleware/cors.go`: configurable CORS headers from environment variables
    - Create `/internal/middleware/content_type.go`: reject non-JSON requests with 415
    - Create `/internal/middleware/body_limit.go`: reject bodies > 1MB with 413
    - Create `/internal/middleware/logging.go`: structured JSON logging with zerolog (method, path, status, duration)
    - Create `/internal/middleware/timeout.go`: context timeout (default 30s, configurable)
    - _Requirements: 10.3, 10.6, 10.7, 11.3, 11.6_

  - [ ]* 11.5 Write property test for structured logging (Property 27)
    - **Property 27: Structured log completeness**
    - Generate random HTTP requests, verify log entries contain all required fields (timestamp, level, method, path, status, duration_ms)
    - **Validates: Requirements 11.3**

  - [ ]* 11.6 Write unit tests for middleware
    - Test auth middleware returns 401 on missing token
    - Test role middleware returns 403 for wrong role
    - Test content-type middleware returns 415 for non-JSON
    - Test body-limit middleware returns 413 for oversized body
    - Test CORS headers present in response
    - _Requirements: 8.4, 10.3, 10.6, 10.7_

- [ ] 12. Handler layer
  - [ ] 12.1 Implement auth handlers
    - Create `/internal/handler/auth_handler.go`
    - POST /api/v1/auth/login: parse LoginRequest, call AuthService.CustomerLogin, return token in success envelope
    - POST /api/v1/auth/logout: extract claims from context, call AuthService.CustomerLogout
    - POST /api/v1/auth/register: parse RegisterRequest, call AccountService.Register, return 201
    - POST /api/v1/admin/auth/login: parse LoginRequest, call AuthService.AdminLogin, return token
    - _Requirements: 1.1, 1.2, 1.4, 2.1, 8.1_

  - [ ] 12.2 Implement dashboard and transaction handlers
    - Create `/internal/handler/dashboard_handler.go`
    - GET /api/v1/dashboard: call DashboardService.GetDashboard
    - GET /api/v1/transactions: parse pagination params, call DashboardService.GetTransactionHistory
    - _Requirements: 3.1, 3.2, 3.3_

  - [ ] 12.3 Implement transfer handler
    - Create `/internal/handler/transfer_handler.go`
    - POST /api/v1/transfers: parse TransferRequest, extract sender from context, call TransferService.QuickTransfer
    - _Requirements: 4.1_

  - [ ] 12.4 Implement request handlers
    - Create `/internal/handler/request_handler.go`
    - POST /api/v1/requests: parse CreateMoneyRequest, call RequestService.CreateRequest
    - GET /api/v1/requests/received: call RequestService.GetReceivedRequests
    - PATCH /api/v1/requests/{id}/viewed: call RequestService.MarkAsViewed
    - _Requirements: 5.1, 5.6, 5.7_

  - [ ] 12.5 Implement profile, feedback, login-history handlers
    - Create `/internal/handler/account_handler.go`: GET /api/v1/profile, PUT /api/v1/profile
    - Create `/internal/handler/feedback_handler.go`: POST /api/v1/feedback
    - Create `/internal/handler/health_handler.go`: GET /health (verify DB connectivity, return 200/503)
    - Create `/internal/handler/login_history_handler.go`: GET /api/v1/login-history
    - _Requirements: 6.1, 6.2, 7.1, 11.4, 14.1_

  - [ ] 12.6 Implement admin handlers
    - Create `/internal/handler/admin_handler.go`
    - GET /api/v1/admin/customers: call AdminService.ListCustomers
    - POST /api/v1/admin/balance-adjustment: call AdminService.AdjustBalance
    - GET /api/v1/admin/transactions: call AdminService.ListTransactions
    - GET /api/v1/admin/requests: call AdminService.ListRequests
    - GET /api/v1/admin/feedback: call AdminService.ListFeedback
    - GET /api/v1/admin/login-history: call AdminService.ListLoginHistory
    - _Requirements: 9.1, 9.2, 9.5, 9.6, 9.7, 9.8_

  - [ ]* 12.7 Write property tests for response format (Properties 28, 29)
    - **Property 28: Success response envelope format**
    - **Property 29: Error response envelope format**
    - Generate random handler responses; verify success has "data"+"meta" with timestamp; verify error has "error" with code+message (max 500 chars)+details only for multi-field validation
    - **Validates: Requirements 13.1, 13.2**

  - [ ]* 12.8 Write unit tests for handlers
    - Test each endpoint returns correct status codes (200, 201, 400, 401, 403, 404, 409, 413, 415, 429, 500)
    - Test pagination metadata in paginated responses
    - Use httptest.ResponseRecorder and mocked services
    - _Requirements: 13.3, 13.4, 13.5_

- [ ] 13. Router setup and application wiring
  - [ ] 13.1 Implement main.go with DI wiring and router setup
    - Create `/cmd/api/main.go`
    - Load config, initialize DB connection with retry logic (3 attempts, 5s interval)
    - Initialize all repositories, services, handlers with constructor injection
    - Set up chi router with middleware stack: body limit → content-type → CORS → logging → timeout → panic recovery
    - Mount auth endpoints (public), mount protected customer routes with JWT middleware, mount admin routes with JWT + RequireRole("admin")
    - Start HTTP server on configured port
    - _Requirements: 11.1, 11.2, 11.5, 11.7_

  - [ ]* 13.2 Write unit test for router configuration
    - Verify all 21 endpoints are registered
    - Verify middleware ordering
    - Verify protected routes return 401 without token
    - _Requirements: 11.7_

- [ ] 14. Database migrations
  - [ ] 14.1 Create migration files
    - Create `/migrations/000001_initial_schema.up.sql`: full schema from bank_db.sql with proper types
    - Create `/migrations/000001_initial_schema.down.sql`: DROP TABLE statements
    - Create `/migrations/000002_alter_password_columns.up.sql`: ALTER password columns to VARCHAR(72)
    - Create `/migrations/000002_alter_password_columns.down.sql`: revert to CHAR(25)
    - Create `/migrations/000003_add_foreign_keys.up.sql`: FK constraints on all tables referencing tbl_account
    - Create `/migrations/000003_add_foreign_keys.down.sql`: DROP FOREIGN KEY statements
    - Create `/migrations/000004_add_indexes.up.sql`: indexes on transaction.account_no, transaction.trans_date, login_history.account_no, feedback.account_no
    - Create `/migrations/000004_add_indexes.down.sql`: DROP INDEX statements
    - Create `/migrations/000005_add_ip_address_and_hash_passwords.up.sql`: add ip_address column + hash existing plaintext passwords
    - Create `/migrations/000005_add_ip_address_and_hash_passwords.down.sql`: remove ip_address column (passwords cannot be unhashed)
    - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5, 12.6, 12.7_

  - [ ]* 14.2 Write property tests for migration concepts (Properties 30, 31)
    - **Property 30: Migration idempotence**
    - **Property 31: Migration password hashing**
    - Verify running migrations twice produces no errors or duplicate changes; verify plaintext passwords < 60 chars get hashed, passwords already 60 chars are unchanged
    - **Validates: Requirements 12.4, 12.7**

- [ ] 15. Checkpoint - Full build verification
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 16. Integration tests and login history
  - [ ]* 16.1 Write integration tests for registration flow
    - Test full registration creates records in all 5 tables atomically
    - Test rollback on failure leaves no partial records
    - Test duplicate username returns 409
    - _Requirements: 2.1, 2.3, 2.10_

  - [ ]* 16.2 Write integration tests for transfer flow
    - Test successful transfer updates both balances correctly
    - Test atomicity: failure mid-transfer leaves balances unchanged
    - Test balance sum invariant across transfers
    - _Requirements: 4.1, 4.7, 4.8_

  - [ ]* 16.3 Write property test for login history (Property 32)
    - **Property 32: Active session null logout**
    - Verify sessions without explicit logout have null logout_time; verify logout sets timestamp
    - **Validates: Requirements 14.4**

  - [ ]* 16.4 Write integration tests for migration execution
    - Test each migration up/down independently
    - Test full migration sequence preserves all existing data (row count verification)
    - Test re-running migrations skips applied ones
    - _Requirements: 12.5, 12.6, 12.7_

- [ ] 17. Generate mocks and finalize test infrastructure
  - [ ] 17.1 Generate mock implementations
    - Run `mockgen` for all repository interfaces, PasswordHasher, TokenManager
    - Place generated mocks in `/test/mock/`
    - Ensure all service tests use generated mocks
    - _Requirements: 11.7_

- [ ] 18. Final checkpoint - Ensure all tests pass and build succeeds
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties from the design document (32 properties total)
- Unit tests validate specific examples and edge cases
- Integration tests require a MySQL test instance (or testcontainers)
- All code follows SOLID principles with constructor-based dependency injection
- Mocks are generated using go.uber.org/mock/mockgen for all interfaces
- gopter property tests use minimum 100 iterations with fixed seed for CI reproducibility

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1"] },
    { "id": 1, "tasks": ["1.2", "1.3", "1.4"] },
    { "id": 2, "tasks": ["2.1", "2.2", "2.3", "3.1", "3.3"] },
    { "id": 3, "tasks": ["2.4", "3.2", "3.4"] },
    { "id": 4, "tasks": ["4.1", "4.2", "4.3", "4.4", "4.5"] },
    { "id": 5, "tasks": ["6.1", "7.1", "7.3", "8.1", "9.1", "9.3", "9.5"] },
    { "id": 6, "tasks": ["6.2", "6.3", "7.2", "7.4", "8.2", "8.3", "9.2", "9.4", "9.6"] },
    { "id": 7, "tasks": ["11.1", "11.2", "11.4"] },
    { "id": 8, "tasks": ["11.3", "11.5", "11.6"] },
    { "id": 9, "tasks": ["12.1", "12.2", "12.3", "12.4", "12.5", "12.6"] },
    { "id": 10, "tasks": ["12.7", "12.8", "13.1"] },
    { "id": 11, "tasks": ["13.2", "14.1"] },
    { "id": 12, "tasks": ["14.2", "17.1"] },
    { "id": 13, "tasks": ["16.1", "16.2", "16.3", "16.4"] }
  ]
}
```
