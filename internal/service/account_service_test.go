package service

import (
	"context"
	"errors"
	"testing"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test helpers ---

func validRegisterRequest() model.RegisterRequest {
	return model.RegisterRequest{
		FirstName:   "John",
		LastName:    "Doe",
		Gender:      "Male",
		BirthDate:   "1990-01-01",
		Mobile:      "1234567890",
		Email:       "john@example.com",
		Address:     "123 Main St",
		City:        "Springfield",
		State:       "IL",
		ZipCode:     "62701",
		Username:    "johndoe",
		Password:    "securepassword123",
		AccountType: "SAVING",
	}
}

func newTestAccountService(
	accountRepo *mockAccountRepo,
	customerRepo *mockCustomerRepo,
	addressRepo *mockAddressRepo,
	accountTypeRepo *mockAccountTypeRepo,
	balanceRepo *mockBalanceRepo,
	hasher *mockHasher,
) AccountService {
	return &accountService{
		db:              nil,
		accountRepo:     accountRepo,
		customerRepo:    customerRepo,
		addressRepo:     addressRepo,
		accountTypeRepo: accountTypeRepo,
		balanceRepo:     balanceRepo,
		hasher:          hasher,
	}
}

// --- Tests ---

func TestRegister_ValidationError(t *testing.T) {
	svc := newTestAccountService(
		&mockAccountRepo{},
		&mockCustomerRepo{},
		&mockAddressRepo{},
		&mockAccountTypeRepo{},
		&mockBalanceRepo{},
		&mockHasher{},
	)

	// Missing required fields
	req := model.RegisterRequest{}
	_, err := svc.Register(context.Background(), req)

	require.Error(t, err)
	var appErr *model.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "VALIDATION_ERROR", appErr.Code)
}

func TestRegister_DuplicateUsername(t *testing.T) {
	accountRepo := &mockAccountRepo{
		usernameExistsFn: func(_ context.Context, _ string) (bool, error) {
			return true, nil
		},
	}

	svc := newTestAccountService(
		accountRepo,
		&mockCustomerRepo{},
		&mockAddressRepo{},
		&mockAccountTypeRepo{},
		&mockBalanceRepo{},
		&mockHasher{},
	)

	req := validRegisterRequest()
	_, err := svc.Register(context.Background(), req)

	require.Error(t, err)
	assert.Equal(t, model.ErrDuplicateUsername, err)
}

func TestRegister_HashError(t *testing.T) {
	accountRepo := &mockAccountRepo{
		usernameExistsFn: func(_ context.Context, _ string) (bool, error) {
			return false, nil
		},
		existsFn: func(_ context.Context, _ int64) (bool, error) {
			return false, nil
		},
	}

	hasher := &mockHasher{
		hashFn: func(_ string) (string, error) {
			return "", errors.New("hash error")
		},
	}

	svc := newTestAccountService(
		accountRepo,
		&mockCustomerRepo{},
		&mockAddressRepo{},
		&mockAccountTypeRepo{},
		&mockBalanceRepo{},
		hasher,
	)

	req := validRegisterRequest()
	_, err := svc.Register(context.Background(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "hashing password")
}

func TestGetProfile_Success(t *testing.T) {
	customerRepo := &mockCustomerRepo{
		getByAccountNoFn: func(_ context.Context, accountNo int64) (*model.Customer, error) {
			return &model.Customer{
				AccountNo: accountNo,
				FullName:  "John Doe",
				Gender:    "Male",
				Email:     "john@example.com",
			}, nil
		},
	}

	addressRepo := &mockAddressRepo{
		getByAccountNoFn: func(_ context.Context, accountNo int64) (*model.Address, error) {
			return &model.Address{
				AccountNo:   accountNo,
				HomeAddress: "123 Main St",
				City:        "Springfield",
			}, nil
		},
	}

	accountTypeRepo := &mockAccountTypeRepo{
		getByAccountNoFn: func(_ context.Context, _ int64) (string, error) {
			return "SAVING", nil
		},
	}

	svc := newTestAccountService(
		&mockAccountRepo{},
		customerRepo,
		addressRepo,
		accountTypeRepo,
		&mockBalanceRepo{},
		&mockHasher{},
	)

	customer, address, accountType, err := svc.GetProfile(context.Background(), 123456789)

	require.NoError(t, err)
	assert.Equal(t, "John Doe", customer.FullName)
	assert.Equal(t, "123 Main St", address.HomeAddress)
	assert.Equal(t, "SAVING", accountType)
}

func TestGetProfile_CustomerNotFound(t *testing.T) {
	customerRepo := &mockCustomerRepo{
		getByAccountNoFn: func(_ context.Context, _ int64) (*model.Customer, error) {
			return nil, errors.New("customer not found")
		},
	}

	svc := newTestAccountService(
		&mockAccountRepo{},
		customerRepo,
		&mockAddressRepo{},
		&mockAccountTypeRepo{},
		&mockBalanceRepo{},
		&mockHasher{},
	)

	_, _, _, err := svc.GetProfile(context.Background(), 123456789)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetching customer profile")
}

func TestUpdateProfile_Success(t *testing.T) {
	fullName := "Jane Doe"
	email := "jane@example.com"

	var updatedCustomer *model.Customer
	customerRepo := &mockCustomerRepo{
		getByAccountNoFn: func(_ context.Context, accountNo int64) (*model.Customer, error) {
			return &model.Customer{
				AccountNo: accountNo,
				FullName:  "John Doe",
				Email:     "john@example.com",
				Gender:    "Male",
				BirthDate: "1990-01-01",
				Mobile:    "1234567890",
			}, nil
		},
		updateFn: func(_ context.Context, customer *model.Customer) error {
			updatedCustomer = customer
			return nil
		},
	}

	svc := newTestAccountService(
		&mockAccountRepo{},
		customerRepo,
		&mockAddressRepo{},
		&mockAccountTypeRepo{},
		&mockBalanceRepo{},
		&mockHasher{},
	)

	req := model.UpdateProfileRequest{
		FullName: &fullName,
		Email:    &email,
	}

	err := svc.UpdateProfile(context.Background(), 123456789, req)

	require.NoError(t, err)
	require.NotNil(t, updatedCustomer)
	assert.Equal(t, "Jane Doe", updatedCustomer.FullName)
	assert.Equal(t, "jane@example.com", updatedCustomer.Email)
	// Unchanged fields should remain the same
	assert.Equal(t, "Male", updatedCustomer.Gender)
	assert.Equal(t, "1234567890", updatedCustomer.Mobile)
}

func TestUpdateProfile_NoFieldsProvided(t *testing.T) {
	svc := newTestAccountService(
		&mockAccountRepo{},
		&mockCustomerRepo{},
		&mockAddressRepo{},
		&mockAccountTypeRepo{},
		&mockBalanceRepo{},
		&mockHasher{},
	)

	req := model.UpdateProfileRequest{}
	err := svc.UpdateProfile(context.Background(), 123456789, req)

	require.Error(t, err)
	var appErr *model.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "VALIDATION_ERROR", appErr.Code)
	assert.Contains(t, appErr.Message, "At least one field must be provided")
}

func TestUpdateProfile_ValidationError(t *testing.T) {
	invalidEmail := "not-an-email"
	svc := newTestAccountService(
		&mockAccountRepo{},
		&mockCustomerRepo{},
		&mockAddressRepo{},
		&mockAccountTypeRepo{},
		&mockBalanceRepo{},
		&mockHasher{},
	)

	req := model.UpdateProfileRequest{
		Email: &invalidEmail,
	}

	err := svc.UpdateProfile(context.Background(), 123456789, req)

	require.Error(t, err)
	var appErr *model.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "VALIDATION_ERROR", appErr.Code)
}

func TestUpdateProfile_WithAddressFields(t *testing.T) {
	city := "New York"
	zipCode := "10001"

	var updatedAddress *model.Address
	addressRepo := &mockAddressRepo{
		getByAccountNoFn: func(_ context.Context, accountNo int64) (*model.Address, error) {
			return &model.Address{
				AccountNo:   accountNo,
				HomeAddress: "123 Main St",
				City:        "Springfield",
				State:       "IL",
				Pincode:     62701,
			}, nil
		},
		updateFn: func(_ context.Context, address *model.Address) error {
			updatedAddress = address
			return nil
		},
	}

	customerRepo := &mockCustomerRepo{
		getByAccountNoFn: func(_ context.Context, accountNo int64) (*model.Customer, error) {
			return &model.Customer{
				AccountNo: accountNo,
				FullName:  "John",
				Gender:    "Male",
				BirthDate: "1990-01-01",
				Mobile:    "1234567890",
				Email:     "j@e.com",
			}, nil
		},
		updateFn: func(_ context.Context, _ *model.Customer) error {
			return nil
		},
	}

	svc := newTestAccountService(
		&mockAccountRepo{},
		customerRepo,
		addressRepo,
		&mockAccountTypeRepo{},
		&mockBalanceRepo{},
		&mockHasher{},
	)

	req := model.UpdateProfileRequest{
		City:    &city,
		ZipCode: &zipCode,
	}

	err := svc.UpdateProfile(context.Background(), 123456789, req)

	require.NoError(t, err)
	require.NotNil(t, updatedAddress)
	assert.Equal(t, "New York", updatedAddress.City)
	assert.Equal(t, 10001, updatedAddress.Pincode)
	// Unchanged address fields
	assert.Equal(t, "123 Main St", updatedAddress.HomeAddress)
	assert.Equal(t, "IL", updatedAddress.State)
}

func TestGenerateUniqueAccountNo(t *testing.T) {
	callCount := 0
	accountRepo := &mockAccountRepo{
		existsFn: func(_ context.Context, _ int64) (bool, error) {
			callCount++
			// First call returns true (collision), second returns false (unique)
			if callCount == 1 {
				return true, nil
			}
			return false, nil
		},
	}

	svc := &accountService{
		accountRepo: accountRepo,
	}

	accountNo, err := svc.generateUniqueAccountNo(context.Background())

	require.NoError(t, err)
	assert.GreaterOrEqual(t, accountNo, int64(100000000))
	assert.LessOrEqual(t, accountNo, int64(999999999))
	assert.Equal(t, 2, callCount)
}

func TestGenerateUniqueAccountNo_AllCollisions(t *testing.T) {
	accountRepo := &mockAccountRepo{
		existsFn: func(_ context.Context, _ int64) (bool, error) {
			return true, nil // always collision
		},
	}

	svc := &accountService{
		accountRepo: accountRepo,
	}

	_, err := svc.generateUniqueAccountNo(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate unique account number")
}
