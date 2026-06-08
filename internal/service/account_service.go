package service

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/abdurrachmanwahed/online-banking/internal/repository"
	"github.com/abdurrachmanwahed/online-banking/internal/security"
	"github.com/abdurrachmanwahed/online-banking/internal/validator"
	"github.com/jmoiron/sqlx"
)

// AccountService handles registration and profile operations.
type AccountService interface {
	Register(ctx context.Context, req model.RegisterRequest) (*model.Account, error)
	GetProfile(ctx context.Context, accountNo int64) (*model.Customer, *model.Address, string, error)
	UpdateProfile(ctx context.Context, accountNo int64, req model.UpdateProfileRequest) error
}

// accountService is the concrete implementation of AccountService.
type accountService struct {
	db              *sqlx.DB
	accountRepo     repository.AccountRepository
	customerRepo    repository.CustomerRepository
	addressRepo     repository.AddressRepository
	accountTypeRepo repository.AccountTypeRepository
	balanceRepo     repository.BalanceRepository
	hasher          security.PasswordHasher
}

// NewAccountService creates a new AccountService with all required dependencies.
func NewAccountService(
	db *sqlx.DB,
	accountRepo repository.AccountRepository,
	customerRepo repository.CustomerRepository,
	addressRepo repository.AddressRepository,
	accountTypeRepo repository.AccountTypeRepository,
	balanceRepo repository.BalanceRepository,
	hasher security.PasswordHasher,
) AccountService {
	return &accountService{
		db:              db,
		accountRepo:     accountRepo,
		customerRepo:    customerRepo,
		addressRepo:     addressRepo,
		accountTypeRepo: accountTypeRepo,
		balanceRepo:     balanceRepo,
		hasher:          hasher,
	}
}

// Register creates a new customer account with all associated records within a single transaction.
func (s *accountService) Register(ctx context.Context, req model.RegisterRequest) (*model.Account, error) {
	// Validate input
	if appErr := validator.ValidateStruct(&req); appErr != nil {
		return nil, appErr
	}

	// Check username uniqueness
	exists, err := s.accountRepo.UsernameExists(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("checking username existence: %w", err)
	}
	if exists {
		return nil, model.ErrDuplicateUsername
	}

	// Generate unique 9-digit account number (100000000–999999999)
	accountNo, err := s.generateUniqueAccountNo(ctx)
	if err != nil {
		return nil, fmt.Errorf("generating account number: %w", err)
	}

	// Hash password
	hashedPassword, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	// Begin transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Create Account record
	account := &model.Account{
		AccountNo: accountNo,
		Username:  req.Username,
		Password:  hashedPassword,
	}
	_, err = s.accountRepo.Create(ctx, tx, account)
	if err != nil {
		return nil, fmt.Errorf("creating account: %w", err)
	}

	// Create Customer record
	fullName := req.FirstName + " " + req.LastName
	customer := &model.Customer{
		AccountNo: accountNo,
		FullName:  fullName,
		Gender:    req.Gender,
		BirthDate: req.BirthDate,
		Mobile:    req.Mobile,
		Email:     req.Email,
	}
	err = s.customerRepo.Create(ctx, tx, customer)
	if err != nil {
		return nil, fmt.Errorf("creating customer: %w", err)
	}

	// Create Address record
	pincode, _ := strconv.Atoi(req.ZipCode)
	address := &model.Address{
		AccountNo:   accountNo,
		HomeAddress: req.Address,
		City:        req.City,
		State:       req.State,
		Pincode:     pincode,
	}
	err = s.addressRepo.Create(ctx, tx, address)
	if err != nil {
		return nil, fmt.Errorf("creating address: %w", err)
	}

	// Create AccountType record
	err = s.accountTypeRepo.Create(ctx, tx, accountNo, req.AccountType)
	if err != nil {
		return nil, fmt.Errorf("creating account type: %w", err)
	}

	// Create Balance record initialized to 0
	err = s.balanceRepo.Create(ctx, tx, accountNo, req.AccountType, 0)
	if err != nil {
		return nil, fmt.Errorf("creating balance: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	return account, nil
}

// GetProfile returns the customer profile, address, and account type for the given account number.
func (s *accountService) GetProfile(ctx context.Context, accountNo int64) (*model.Customer, *model.Address, string, error) {
	customer, err := s.customerRepo.GetByAccountNo(ctx, accountNo)
	if err != nil {
		return nil, nil, "", fmt.Errorf("fetching customer profile: %w", err)
	}

	address, err := s.addressRepo.GetByAccountNo(ctx, accountNo)
	if err != nil {
		return nil, nil, "", fmt.Errorf("fetching address: %w", err)
	}

	accountType, err := s.accountTypeRepo.GetByAccountNo(ctx, accountNo)
	if err != nil {
		return nil, nil, "", fmt.Errorf("fetching account type: %w", err)
	}

	return customer, address, accountType, nil
}

// UpdateProfile applies partial updates to the customer profile and address.
// Only fields present (non-nil pointers) in the request are updated.
func (s *accountService) UpdateProfile(ctx context.Context, accountNo int64, req model.UpdateProfileRequest) error {
	// Validate input
	if appErr := validator.ValidateStruct(&req); appErr != nil {
		return appErr
	}

	// Check that at least one field is provided
	if !hasAnyField(req) {
		return model.NewAppError("VALIDATION_ERROR", "At least one field must be provided", http.StatusBadRequest)
	}

	// Update customer fields
	if err := s.updateCustomerFields(ctx, accountNo, req); err != nil {
		return err
	}

	// Update address fields if any are present
	if hasAddressField(req) {
		if err := s.updateAddressFields(ctx, accountNo, req); err != nil {
			return err
		}
	}

	return nil
}

// updateCustomerFields fetches the current customer and applies partial updates.
func (s *accountService) updateCustomerFields(ctx context.Context, accountNo int64, req model.UpdateProfileRequest) error {
	customer, err := s.customerRepo.GetByAccountNo(ctx, accountNo)
	if err != nil {
		return fmt.Errorf("fetching current customer: %w", err)
	}

	applyCustomerUpdates(customer, req)

	if err := s.customerRepo.Update(ctx, customer); err != nil {
		return fmt.Errorf("updating customer: %w", err)
	}
	return nil
}

// applyCustomerUpdates applies non-nil fields from the request to the customer model.
func applyCustomerUpdates(customer *model.Customer, req model.UpdateProfileRequest) {
	if req.FullName != nil {
		customer.FullName = *req.FullName
	}
	if req.Gender != nil {
		customer.Gender = *req.Gender
	}
	if req.BirthDate != nil {
		customer.BirthDate = *req.BirthDate
	}
	if req.Mobile != nil {
		customer.Mobile = *req.Mobile
	}
	if req.Email != nil {
		customer.Email = *req.Email
	}
}

// updateAddressFields fetches the current address and applies partial updates.
func (s *accountService) updateAddressFields(ctx context.Context, accountNo int64, req model.UpdateProfileRequest) error {
	address, err := s.addressRepo.GetByAccountNo(ctx, accountNo)
	if err != nil {
		return fmt.Errorf("fetching current address: %w", err)
	}

	applyAddressUpdates(address, req)

	if err := s.addressRepo.Update(ctx, address); err != nil {
		return fmt.Errorf("updating address: %w", err)
	}
	return nil
}

// applyAddressUpdates applies non-nil address fields from the request to the address model.
func applyAddressUpdates(address *model.Address, req model.UpdateProfileRequest) {
	if req.Address != nil {
		address.HomeAddress = *req.Address
	}
	if req.City != nil {
		address.City = *req.City
	}
	if req.State != nil {
		address.State = *req.State
	}
	if req.ZipCode != nil {
		pincode, _ := strconv.Atoi(*req.ZipCode)
		address.Pincode = pincode
	}
}

// generateUniqueAccountNo generates a unique 9-digit account number (100000000–999999999).
// It attempts up to 10 times to find a unique number before returning an error.
func (s *accountService) generateUniqueAccountNo(ctx context.Context) (int64, error) {
	const maxAttempts = 10
	for i := 0; i < maxAttempts; i++ {
		// Generate random 9-digit number: 100000000 to 999999999
		accountNo := int64(rand.Intn(900000000) + 100000000)

		exists, err := s.accountRepo.Exists(ctx, accountNo)
		if err != nil {
			return 0, err
		}
		if !exists {
			return accountNo, nil
		}
	}
	return 0, fmt.Errorf("failed to generate unique account number after %d attempts", maxAttempts)
}

// hasAnyField checks if at least one field in UpdateProfileRequest is non-nil.
func hasAnyField(req model.UpdateProfileRequest) bool {
	return req.FullName != nil ||
		req.Gender != nil ||
		req.BirthDate != nil ||
		req.Mobile != nil ||
		req.Email != nil ||
		req.Address != nil ||
		req.City != nil ||
		req.State != nil ||
		req.ZipCode != nil
}

// hasAddressField checks if any address-related field in UpdateProfileRequest is non-nil.
func hasAddressField(req model.UpdateProfileRequest) bool {
	return req.Address != nil ||
		req.City != nil ||
		req.State != nil ||
		req.ZipCode != nil
}
