package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/abdurrachmanwahed/online-banking/internal/security"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Auth-specific mock implementations with unique names ---

type authMockAdminRepo struct {
	getByIDFn    func(ctx context.Context, adminID int64) (*model.Admin, error)
	getByEmailFn func(ctx context.Context, email string) (*model.Admin, error)
}

func (m *authMockAdminRepo) GetByID(ctx context.Context, adminID int64) (*model.Admin, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, adminID)
	}
	return nil, nil
}

func (m *authMockAdminRepo) GetByEmail(ctx context.Context, email string) (*model.Admin, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, nil
}

type authMockLoginHistoryRepo struct {
	recordLoginFn        func(ctx context.Context, accountNo int64, ipAddress string) (int64, error)
	recordLogoutFn       func(ctx context.Context, tokenID int64) error
	getByAccountNoFn     func(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.LoginHistory, int64, error)
	getAllFn             func(ctx context.Context, pagination model.Pagination) ([]model.LoginHistory, int64, error)
	getLatestByAccountFn func(ctx context.Context, accountNo int64) (*model.LoginHistory, error)
}

func (m *authMockLoginHistoryRepo) RecordLogin(ctx context.Context, accountNo int64, ipAddress string) (int64, error) {
	if m.recordLoginFn != nil {
		return m.recordLoginFn(ctx, accountNo, ipAddress)
	}
	return 1, nil
}

func (m *authMockLoginHistoryRepo) RecordLogout(ctx context.Context, tokenID int64) error {
	if m.recordLogoutFn != nil {
		return m.recordLogoutFn(ctx, tokenID)
	}
	return nil
}

func (m *authMockLoginHistoryRepo) GetByAccountNo(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.LoginHistory, int64, error) {
	if m.getByAccountNoFn != nil {
		return m.getByAccountNoFn(ctx, accountNo, pagination)
	}
	return nil, 0, nil
}

func (m *authMockLoginHistoryRepo) GetAll(ctx context.Context, pagination model.Pagination) ([]model.LoginHistory, int64, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx, pagination)
	}
	return nil, 0, nil
}

func (m *authMockLoginHistoryRepo) GetLatestByAccountNo(ctx context.Context, accountNo int64) (*model.LoginHistory, error) {
	if m.getLatestByAccountFn != nil {
		return m.getLatestByAccountFn(ctx, accountNo)
	}
	return nil, nil
}

type authMockTokenManager struct {
	generateFn   func(claims security.TokenClaims) (string, error)
	validateFn   func(tokenString string) (*security.TokenClaims, error)
	invalidateFn func(tokenString string) error
}

func (m *authMockTokenManager) Generate(claims security.TokenClaims) (string, error) {
	if m.generateFn != nil {
		return m.generateFn(claims)
	}
	return "mock-jwt-token", nil
}

func (m *authMockTokenManager) Validate(tokenString string) (*security.TokenClaims, error) {
	if m.validateFn != nil {
		return m.validateFn(tokenString)
	}
	return nil, nil
}

func (m *authMockTokenManager) Invalidate(tokenString string) error {
	if m.invalidateFn != nil {
		return m.invalidateFn(tokenString)
	}
	return nil
}

type authMockHasher struct {
	hashFn    func(password string) (string, error)
	compareFn func(hashedPassword, plainPassword string) error
}

func (m *authMockHasher) Hash(password string) (string, error) {
	if m.hashFn != nil {
		return m.hashFn(password)
	}
	return "$2a$10$hashedpassword", nil
}

func (m *authMockHasher) Compare(hashedPassword, plainPassword string) error {
	if m.compareFn != nil {
		return m.compareFn(hashedPassword, plainPassword)
	}
	return nil
}

type authMockAccountRepo struct {
	getByUsernameFn  func(ctx context.Context, username string) (*model.Account, error)
	getByAccountNoFn func(ctx context.Context, accountNo int64) (*model.Account, error)
	existsFn         func(ctx context.Context, accountNo int64) (bool, error)
	usernameExistsFn func(ctx context.Context, username string) (bool, error)
}

func (m *authMockAccountRepo) GetByUsername(ctx context.Context, username string) (*model.Account, error) {
	if m.getByUsernameFn != nil {
		return m.getByUsernameFn(ctx, username)
	}
	return nil, nil
}

func (m *authMockAccountRepo) GetByAccountNo(ctx context.Context, accountNo int64) (*model.Account, error) {
	if m.getByAccountNoFn != nil {
		return m.getByAccountNoFn(ctx, accountNo)
	}
	return nil, nil
}

func (m *authMockAccountRepo) Exists(ctx context.Context, accountNo int64) (bool, error) {
	if m.existsFn != nil {
		return m.existsFn(ctx, accountNo)
	}
	return false, nil
}

func (m *authMockAccountRepo) UsernameExists(ctx context.Context, username string) (bool, error) {
	if m.usernameExistsFn != nil {
		return m.usernameExistsFn(ctx, username)
	}
	return false, nil
}

func (m *authMockAccountRepo) Create(_ context.Context, _ *sqlx.Tx, _ *model.Account) (int64, error) {
	return 0, nil
}

func (m *authMockAccountRepo) UpdatePassword(_ context.Context, _ int64, _ string) error {
	return nil
}

// --- Tests ---

func TestCustomerLogin_Success(t *testing.T) {
	accountRepo := &authMockAccountRepo{
		getByUsernameFn: func(_ context.Context, username string) (*model.Account, error) {
			return &model.Account{
				AccountNo: 123456789,
				Username:  username,
				Password:  "$2a$10$somehashedpassword",
			}, nil
		},
	}

	var recordedAccountNo int64
	var recordedIP string
	loginHistoryRepo := &authMockLoginHistoryRepo{
		recordLoginFn: func(_ context.Context, accountNo int64, ipAddress string) (int64, error) {
			recordedAccountNo = accountNo
			recordedIP = ipAddress
			return 42, nil
		},
	}

	var generatedRole string
	tokenMgr := &authMockTokenManager{
		generateFn: func(claims security.TokenClaims) (string, error) {
			generatedRole = claims.Role
			return "jwt-token-123", nil
		},
	}

	svc := NewAuthService(accountRepo, &authMockAdminRepo{}, loginHistoryRepo, &authMockHasher{}, tokenMgr)

	resp, tokenID, err := svc.CustomerLogin(context.Background(), model.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}, "192.168.1.1")

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "jwt-token-123", resp.Token)
	assert.Equal(t, int64(42), tokenID)
	assert.Equal(t, "customer", generatedRole)
	assert.Equal(t, int64(123456789), recordedAccountNo)
	assert.Equal(t, "192.168.1.1", recordedIP)
	assert.True(t, resp.ExpiresAt > time.Now().Unix())
}

func TestCustomerLogin_InvalidUsername(t *testing.T) {
	accountRepo := &authMockAccountRepo{
		getByUsernameFn: func(_ context.Context, _ string) (*model.Account, error) {
			return nil, nil // user not found
		},
	}

	svc := NewAuthService(accountRepo, &authMockAdminRepo{}, &authMockLoginHistoryRepo{}, &authMockHasher{}, &authMockTokenManager{})

	resp, tokenID, err := svc.CustomerLogin(context.Background(), model.LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}, "10.0.0.1")

	assert.Nil(t, resp)
	assert.Equal(t, int64(0), tokenID)
	assert.Equal(t, model.ErrInvalidCredentials, err)
}

func TestCustomerLogin_InvalidPassword(t *testing.T) {
	accountRepo := &authMockAccountRepo{
		getByUsernameFn: func(_ context.Context, _ string) (*model.Account, error) {
			return &model.Account{
				AccountNo: 123456789,
				Username:  "testuser",
				Password:  "$2a$10$somehashedpassword",
			}, nil
		},
	}

	hasher := &authMockHasher{
		compareFn: func(_, _ string) error {
			return errors.New("mismatch")
		},
	}

	svc := NewAuthService(accountRepo, &authMockAdminRepo{}, &authMockLoginHistoryRepo{}, hasher, &authMockTokenManager{})

	resp, tokenID, err := svc.CustomerLogin(context.Background(), model.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}, "10.0.0.1")

	assert.Nil(t, resp)
	assert.Equal(t, int64(0), tokenID)
	assert.Equal(t, model.ErrInvalidCredentials, err)
}

func TestCustomerLogin_AccountLocked(t *testing.T) {
	accountRepo := &authMockAccountRepo{
		getByUsernameFn: func(_ context.Context, _ string) (*model.Account, error) {
			return &model.Account{
				AccountNo: 123456789,
				Username:  "lockeduser",
				Password:  "$2a$10$somehashedpassword",
			}, nil
		},
	}

	hasher := &authMockHasher{
		compareFn: func(_, _ string) error {
			return errors.New("mismatch")
		},
	}

	svc := NewAuthService(accountRepo, &authMockAdminRepo{}, &authMockLoginHistoryRepo{}, hasher, &authMockTokenManager{})

	// Make 5 failed attempts to trigger lockout
	for i := 0; i < 5; i++ {
		_, _, _ = svc.CustomerLogin(context.Background(), model.LoginRequest{
			Username: "lockeduser",
			Password: "wrongpass1",
		}, "10.0.0.1")
	}

	// 6th attempt should be locked
	resp, tokenID, err := svc.CustomerLogin(context.Background(), model.LoginRequest{
		Username: "lockeduser",
		Password: "correctpass",
	}, "10.0.0.1")

	assert.Nil(t, resp)
	assert.Equal(t, int64(0), tokenID)
	assert.Equal(t, model.ErrAccountLocked, err)
}

func TestCustomerLogout_Success(t *testing.T) {
	var invalidatedToken string
	var loggedOutTokenID int64

	tokenMgr := &authMockTokenManager{
		invalidateFn: func(tokenString string) error {
			invalidatedToken = tokenString
			return nil
		},
	}

	loginHistoryRepo := &authMockLoginHistoryRepo{
		recordLogoutFn: func(_ context.Context, tokenID int64) error {
			loggedOutTokenID = tokenID
			return nil
		},
	}

	svc := NewAuthService(&authMockAccountRepo{}, &authMockAdminRepo{}, loginHistoryRepo, &authMockHasher{}, tokenMgr)

	err := svc.CustomerLogout(context.Background(), "some-jwt-token", 42)

	require.NoError(t, err)
	assert.Equal(t, "some-jwt-token", invalidatedToken)
	assert.Equal(t, int64(42), loggedOutTokenID)
}

func TestAdminLogin_Success(t *testing.T) {
	adminRepo := &authMockAdminRepo{
		getByEmailFn: func(_ context.Context, email string) (*model.Admin, error) {
			return &model.Admin{
				AdminID:  1,
				FullName: "Admin User",
				Email:    email,
				Password: "$2a$10$adminhashedpassword",
			}, nil
		},
	}

	var generatedRole string
	var generatedAccountNo int64
	tokenMgr := &authMockTokenManager{
		generateFn: func(claims security.TokenClaims) (string, error) {
			generatedRole = claims.Role
			generatedAccountNo = claims.AccountNo
			return "admin-jwt-token", nil
		},
	}

	svc := NewAuthService(&authMockAccountRepo{}, adminRepo, &authMockLoginHistoryRepo{}, &authMockHasher{}, tokenMgr)

	resp, err := svc.AdminLogin(context.Background(), model.LoginRequest{
		Username: "admin@bank.com",
		Password: "adminpass123",
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "admin-jwt-token", resp.Token)
	assert.Equal(t, "admin", generatedRole)
	assert.Equal(t, int64(1), generatedAccountNo)
	assert.True(t, resp.ExpiresAt > time.Now().Unix())
}

func TestAdminLogin_InvalidCredentials(t *testing.T) {
	adminRepo := &authMockAdminRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.Admin, error) {
			return nil, nil // admin not found
		},
	}

	svc := NewAuthService(&authMockAccountRepo{}, adminRepo, &authMockLoginHistoryRepo{}, &authMockHasher{}, &authMockTokenManager{})

	resp, err := svc.AdminLogin(context.Background(), model.LoginRequest{
		Username: "nonexistent@bank.com",
		Password: "somepass123",
	})

	assert.Nil(t, resp)
	assert.Equal(t, model.ErrInvalidCredentials, err)
}

func TestIsAccountLocked_NotLocked(t *testing.T) {
	svc := NewAuthService(&authMockAccountRepo{}, &authMockAdminRepo{}, &authMockLoginHistoryRepo{}, &authMockHasher{}, &authMockTokenManager{})

	locked, err := svc.IsAccountLocked("someuser")
	require.NoError(t, err)
	assert.False(t, locked)
}

func TestIsAccountLocked_LockedAfterFiveFailures(t *testing.T) {
	accountRepo := &authMockAccountRepo{
		getByUsernameFn: func(_ context.Context, _ string) (*model.Account, error) {
			return &model.Account{
				AccountNo: 123456789,
				Username:  "failuser",
				Password:  "$2a$10$somehashedpassword",
			}, nil
		},
	}

	hasher := &authMockHasher{
		compareFn: func(_, _ string) error {
			return errors.New("mismatch")
		},
	}

	svc := NewAuthService(accountRepo, &authMockAdminRepo{}, &authMockLoginHistoryRepo{}, hasher, &authMockTokenManager{})

	// Make 5 failed login attempts
	for i := 0; i < 5; i++ {
		_, _, _ = svc.CustomerLogin(context.Background(), model.LoginRequest{
			Username: "failuser",
			Password: "wrongpass",
		}, "10.0.0.1")
	}

	locked, err := svc.IsAccountLocked("failuser")
	require.NoError(t, err)
	assert.True(t, locked)
}
