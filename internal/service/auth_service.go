package service

import (
	"context"
	"sync"
	"time"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/abdurrachmanwahed/online-banking/internal/repository"
	"github.com/abdurrachmanwahed/online-banking/internal/security"
)

const (
	maxFailedAttempts = 5
	lockoutWindow     = 15 * time.Minute
	lockoutDuration   = 30 * time.Minute
)

// AuthService handles authentication operations.
type AuthService interface {
	CustomerLogin(ctx context.Context, req model.LoginRequest, ipAddress string) (*model.LoginResponse, int64, error)
	CustomerLogout(ctx context.Context, tokenString string, tokenID int64) error
	AdminLogin(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error)
	IsAccountLocked(username string) (bool, error)
}

// failedAttempt tracks a single failed login attempt timestamp.
type failedAttempt struct {
	timestamp time.Time
}

// lockoutEntry tracks failed attempts and lockout state for a username.
type lockoutEntry struct {
	attempts []failedAttempt
	lockedAt *time.Time
}

// authService is the concrete implementation of AuthService.
type authService struct {
	accountRepo      repository.AccountRepository
	adminRepo        repository.AdminRepository
	loginHistoryRepo repository.LoginHistoryRepository
	hasher           security.PasswordHasher
	tokenManager     security.TokenManager

	mu       sync.Mutex
	lockouts map[string]*lockoutEntry
}

// NewAuthService creates a new AuthService with all required dependencies injected.
func NewAuthService(
	accountRepo repository.AccountRepository,
	adminRepo repository.AdminRepository,
	loginHistoryRepo repository.LoginHistoryRepository,
	hasher security.PasswordHasher,
	tokenManager security.TokenManager,
) AuthService {
	return &authService{
		accountRepo:      accountRepo,
		adminRepo:        adminRepo,
		loginHistoryRepo: loginHistoryRepo,
		hasher:           hasher,
		tokenManager:     tokenManager,
		lockouts:         make(map[string]*lockoutEntry),
	}
}

// CustomerLogin validates customer credentials, checks lockout, generates a JWT,
// and records login history with the client IP address.
// Returns the login response, the token ID (for logout tracking), and any error.
func (s *authService) CustomerLogin(ctx context.Context, req model.LoginRequest, ipAddress string) (*model.LoginResponse, int64, error) {
	// Check if account is locked
	locked, err := s.IsAccountLocked(req.Username)
	if err != nil {
		return nil, 0, err
	}
	if locked {
		return nil, 0, model.ErrAccountLocked
	}

	// Retrieve account by username
	account, err := s.accountRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, 0, err
	}
	if account == nil {
		// Record failed attempt even for non-existent username to prevent enumeration
		s.recordFailedAttempt(req.Username)
		return nil, 0, model.ErrInvalidCredentials
	}

	// Verify password
	if err := s.hasher.Compare(account.Password, req.Password); err != nil {
		s.recordFailedAttempt(req.Username)
		return nil, 0, model.ErrInvalidCredentials
	}

	// Successful login — clear any failed attempts
	s.clearFailedAttempts(req.Username)

	// Generate JWT token
	claims := security.TokenClaims{
		AccountNo: account.AccountNo,
		Role:      "customer",
	}
	token, err := s.tokenManager.Generate(claims)
	if err != nil {
		return nil, 0, err
	}

	// Record login history
	tokenID, err := s.loginHistoryRepo.RecordLogin(ctx, account.AccountNo, ipAddress)
	if err != nil {
		return nil, 0, err
	}

	// Calculate expiry for response
	expiresAt := time.Now().Add(15 * time.Minute).Unix()

	response := &model.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}

	return response, tokenID, nil
}

// CustomerLogout invalidates the current token and records the logout time.
func (s *authService) CustomerLogout(ctx context.Context, tokenString string, tokenID int64) error {
	// Invalidate the token
	if err := s.tokenManager.Invalidate(tokenString); err != nil {
		return err
	}

	// Record logout time in login history
	if err := s.loginHistoryRepo.RecordLogout(ctx, tokenID); err != nil {
		return err
	}

	return nil
}

// AdminLogin validates admin credentials and generates a JWT with admin role claims.
func (s *authService) AdminLogin(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error) {
	// For admin login, the username field contains the email
	admin, err := s.adminRepo.GetByEmail(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, model.ErrInvalidCredentials
	}

	// Verify password
	if err := s.hasher.Compare(admin.Password, req.Password); err != nil {
		return nil, model.ErrInvalidCredentials
	}

	// Generate JWT token with admin role
	claims := security.TokenClaims{
		AccountNo: admin.AdminID,
		Role:      "admin",
	}
	token, err := s.tokenManager.Generate(claims)
	if err != nil {
		return nil, err
	}

	// Calculate expiry for response (30 minutes for admin)
	expiresAt := time.Now().Add(30 * time.Minute).Unix()

	response := &model.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}

	return response, nil
}

// IsAccountLocked checks whether the given username is currently locked out
// due to too many failed login attempts within the lockout window.
// Lockout rule: 5 failures within 15 minutes triggers a 30-minute lock.
func (s *authService) IsAccountLocked(username string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.lockouts[username]
	if !exists {
		return false, nil
	}

	// Check if account is currently in a locked state
	if entry.lockedAt != nil {
		elapsed := time.Since(*entry.lockedAt)
		if elapsed < lockoutDuration {
			return true, nil
		}
		// Lock has expired, reset the entry
		delete(s.lockouts, username)
		return false, nil
	}

	return false, nil
}

// recordFailedAttempt tracks a failed login attempt for the given username.
// If the number of recent failures reaches the threshold, the account is locked.
func (s *authService) recordFailedAttempt(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	entry, exists := s.lockouts[username]
	if !exists {
		entry = &lockoutEntry{
			attempts: make([]failedAttempt, 0),
		}
		s.lockouts[username] = entry
	}

	// If already locked, no need to record more
	if entry.lockedAt != nil {
		return
	}

	// Add the new failed attempt
	entry.attempts = append(entry.attempts, failedAttempt{timestamp: now})

	// Filter attempts to only those within the lockout window
	cutoff := now.Add(-lockoutWindow)
	recentAttempts := make([]failedAttempt, 0)
	for _, attempt := range entry.attempts {
		if attempt.timestamp.After(cutoff) {
			recentAttempts = append(recentAttempts, attempt)
		}
	}
	entry.attempts = recentAttempts

	// Check if we've hit the threshold
	if len(entry.attempts) >= maxFailedAttempts {
		entry.lockedAt = &now
	}
}

// clearFailedAttempts removes all tracked failed attempts for a username.
func (s *authService) clearFailedAttempts(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.lockouts, username)
}
