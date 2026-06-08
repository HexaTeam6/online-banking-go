package service

import (
	"context"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/jmoiron/sqlx"
)

// --- Shared mock implementations for service tests ---

// mockAccountRepo implements repository.AccountRepository for testing.
type mockAccountRepo struct {
	usernameExistsFn func(ctx context.Context, username string) (bool, error)
	existsFn         func(ctx context.Context, accountNo int64) (bool, error)
	createFn         func(ctx context.Context, tx *sqlx.Tx, account *model.Account) (int64, error)
	getByAccountNoFn func(ctx context.Context, accountNo int64) (*model.Account, error)
	getByUsernameFn  func(ctx context.Context, username string) (*model.Account, error)
	updatePasswordFn func(ctx context.Context, accountNo int64, hashedPassword string) error
}

func (m *mockAccountRepo) GetByUsername(ctx context.Context, username string) (*model.Account, error) {
	if m.getByUsernameFn != nil {
		return m.getByUsernameFn(ctx, username)
	}
	return nil, nil
}

func (m *mockAccountRepo) GetByAccountNo(ctx context.Context, accountNo int64) (*model.Account, error) {
	if m.getByAccountNoFn != nil {
		return m.getByAccountNoFn(ctx, accountNo)
	}
	return nil, nil
}

func (m *mockAccountRepo) Exists(ctx context.Context, accountNo int64) (bool, error) {
	if m.existsFn != nil {
		return m.existsFn(ctx, accountNo)
	}
	return false, nil
}

func (m *mockAccountRepo) UsernameExists(ctx context.Context, username string) (bool, error) {
	if m.usernameExistsFn != nil {
		return m.usernameExistsFn(ctx, username)
	}
	return false, nil
}

func (m *mockAccountRepo) Create(ctx context.Context, tx *sqlx.Tx, account *model.Account) (int64, error) {
	if m.createFn != nil {
		return m.createFn(ctx, tx, account)
	}
	return account.AccountNo, nil
}

func (m *mockAccountRepo) UpdatePassword(ctx context.Context, accountNo int64, hashedPassword string) error {
	if m.updatePasswordFn != nil {
		return m.updatePasswordFn(ctx, accountNo, hashedPassword)
	}
	return nil
}

// mockCustomerRepo implements repository.CustomerRepository for testing.
type mockCustomerRepo struct {
	getByAccountNoFn func(ctx context.Context, accountNo int64) (*model.Customer, error)
	createFn         func(ctx context.Context, tx *sqlx.Tx, customer *model.Customer) error
	updateFn         func(ctx context.Context, customer *model.Customer) error
	listAllFn        func(ctx context.Context, pagination model.Pagination) ([]model.Customer, int64, error)
}

func (m *mockCustomerRepo) GetByAccountNo(ctx context.Context, accountNo int64) (*model.Customer, error) {
	if m.getByAccountNoFn != nil {
		return m.getByAccountNoFn(ctx, accountNo)
	}
	return &model.Customer{AccountNo: accountNo}, nil
}

func (m *mockCustomerRepo) Create(ctx context.Context, tx *sqlx.Tx, customer *model.Customer) error {
	if m.createFn != nil {
		return m.createFn(ctx, tx, customer)
	}
	return nil
}

func (m *mockCustomerRepo) Update(ctx context.Context, customer *model.Customer) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, customer)
	}
	return nil
}

func (m *mockCustomerRepo) ListAll(ctx context.Context, pagination model.Pagination) ([]model.Customer, int64, error) {
	if m.listAllFn != nil {
		return m.listAllFn(ctx, pagination)
	}
	return []model.Customer{}, 0, nil
}

// mockAddressRepo implements repository.AddressRepository for testing.
type mockAddressRepo struct {
	createFn         func(ctx context.Context, tx *sqlx.Tx, address *model.Address) error
	getByAccountNoFn func(ctx context.Context, accountNo int64) (*model.Address, error)
	updateFn         func(ctx context.Context, address *model.Address) error
}

func (m *mockAddressRepo) Create(ctx context.Context, tx *sqlx.Tx, address *model.Address) error {
	if m.createFn != nil {
		return m.createFn(ctx, tx, address)
	}
	return nil
}

func (m *mockAddressRepo) GetByAccountNo(ctx context.Context, accountNo int64) (*model.Address, error) {
	if m.getByAccountNoFn != nil {
		return m.getByAccountNoFn(ctx, accountNo)
	}
	return &model.Address{AccountNo: accountNo}, nil
}

func (m *mockAddressRepo) Update(ctx context.Context, address *model.Address) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, address)
	}
	return nil
}

// mockAccountTypeRepo implements repository.AccountTypeRepository for testing.
type mockAccountTypeRepo struct {
	createFn         func(ctx context.Context, tx *sqlx.Tx, accountNo int64, accountType string) error
	getByAccountNoFn func(ctx context.Context, accountNo int64) (string, error)
}

func (m *mockAccountTypeRepo) Create(ctx context.Context, tx *sqlx.Tx, accountNo int64, accountType string) error {
	if m.createFn != nil {
		return m.createFn(ctx, tx, accountNo, accountType)
	}
	return nil
}

func (m *mockAccountTypeRepo) GetByAccountNo(ctx context.Context, accountNo int64) (string, error) {
	if m.getByAccountNoFn != nil {
		return m.getByAccountNoFn(ctx, accountNo)
	}
	return "SAVING", nil
}

// mockBalanceRepo implements repository.BalanceRepository for testing.
type mockBalanceRepo struct {
	getByAccountNoFn func(ctx context.Context, accountNo int64) (int64, error)
	creditFn         func(ctx context.Context, tx *sqlx.Tx, accountNo int64, amount int64) (int64, error)
	debitFn          func(ctx context.Context, tx *sqlx.Tx, accountNo int64, amount int64) (int64, error)
	createFn         func(ctx context.Context, tx *sqlx.Tx, accountNo int64, accountType string, balance int64) error
}

func (m *mockBalanceRepo) GetByAccountNo(ctx context.Context, accountNo int64) (int64, error) {
	if m.getByAccountNoFn != nil {
		return m.getByAccountNoFn(ctx, accountNo)
	}
	return 0, nil
}

func (m *mockBalanceRepo) Credit(ctx context.Context, tx *sqlx.Tx, accountNo int64, amount int64) (int64, error) {
	if m.creditFn != nil {
		return m.creditFn(ctx, tx, accountNo, amount)
	}
	return amount, nil
}

func (m *mockBalanceRepo) Debit(ctx context.Context, tx *sqlx.Tx, accountNo int64, amount int64) (int64, error) {
	if m.debitFn != nil {
		return m.debitFn(ctx, tx, accountNo, amount)
	}
	return 0, nil
}

func (m *mockBalanceRepo) Create(ctx context.Context, tx *sqlx.Tx, accountNo int64, accountType string, balance int64) error {
	if m.createFn != nil {
		return m.createFn(ctx, tx, accountNo, accountType, balance)
	}
	return nil
}

// mockTransactionRepo implements repository.TransactionRepository for testing.
type mockTransactionRepo struct {
	createFn            func(ctx context.Context, tx *sqlx.Tx, transaction *model.Transaction) error
	getAllFn            func(ctx context.Context, pagination model.Pagination) ([]model.Transaction, int64, error)
	getByAccountNoFn    func(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.Transaction, int64, error)
	getSummaryFn        func(ctx context.Context, accountNo int64) (*model.TransactionSummary, error)
	getMonthlySummaryFn func(ctx context.Context, accountNo int64, year int, month int) (*model.TransactionSummary, error)
}

func (m *mockTransactionRepo) Create(ctx context.Context, tx *sqlx.Tx, transaction *model.Transaction) error {
	if m.createFn != nil {
		return m.createFn(ctx, tx, transaction)
	}
	return nil
}

func (m *mockTransactionRepo) GetByAccountNo(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.Transaction, int64, error) {
	if m.getByAccountNoFn != nil {
		return m.getByAccountNoFn(ctx, accountNo, pagination)
	}
	return []model.Transaction{}, 0, nil
}

func (m *mockTransactionRepo) GetAll(ctx context.Context, pagination model.Pagination) ([]model.Transaction, int64, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx, pagination)
	}
	return []model.Transaction{}, 0, nil
}

func (m *mockTransactionRepo) GetSummary(ctx context.Context, accountNo int64) (*model.TransactionSummary, error) {
	if m.getSummaryFn != nil {
		return m.getSummaryFn(ctx, accountNo)
	}
	return &model.TransactionSummary{}, nil
}

func (m *mockTransactionRepo) GetMonthlySummary(ctx context.Context, accountNo int64, year int, month int) (*model.TransactionSummary, error) {
	if m.getMonthlySummaryFn != nil {
		return m.getMonthlySummaryFn(ctx, accountNo, year, month)
	}
	return &model.TransactionSummary{}, nil
}

// mockRequestRepo implements repository.RequestRepository for testing.
type mockRequestRepo struct {
	createFn               func(ctx context.Context, request *model.MoneyRequest) error
	getReceivedByAccountFn func(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.MoneyRequest, int64, error)
	markAsViewedFn         func(ctx context.Context, requestID int64) error
	getAllFn               func(ctx context.Context, pagination model.Pagination) ([]model.MoneyRequest, int64, error)
}

func (m *mockRequestRepo) Create(ctx context.Context, request *model.MoneyRequest) error {
	if m.createFn != nil {
		return m.createFn(ctx, request)
	}
	return nil
}

func (m *mockRequestRepo) GetReceivedByAccountNo(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.MoneyRequest, int64, error) {
	if m.getReceivedByAccountFn != nil {
		return m.getReceivedByAccountFn(ctx, accountNo, pagination)
	}
	return []model.MoneyRequest{}, 0, nil
}

func (m *mockRequestRepo) MarkAsViewed(ctx context.Context, requestID int64) error {
	if m.markAsViewedFn != nil {
		return m.markAsViewedFn(ctx, requestID)
	}
	return nil
}

func (m *mockRequestRepo) GetAll(ctx context.Context, pagination model.Pagination) ([]model.MoneyRequest, int64, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx, pagination)
	}
	return []model.MoneyRequest{}, 0, nil
}

// mockFeedbackRepo implements repository.FeedbackRepository for testing.
type mockFeedbackRepo struct {
	createFn func(ctx context.Context, feedback *model.Feedback) error
	getAllFn func(ctx context.Context, pagination model.Pagination) ([]model.Feedback, int64, error)
}

func (m *mockFeedbackRepo) Create(ctx context.Context, feedback *model.Feedback) error {
	if m.createFn != nil {
		return m.createFn(ctx, feedback)
	}
	return nil
}

func (m *mockFeedbackRepo) GetAll(ctx context.Context, pagination model.Pagination) ([]model.Feedback, int64, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx, pagination)
	}
	return []model.Feedback{}, 0, nil
}

// mockLoginHistoryRepo implements repository.LoginHistoryRepository for testing.
type mockLoginHistoryRepo struct {
	recordLoginFn        func(ctx context.Context, accountNo int64, ipAddress string) (int64, error)
	recordLogoutFn       func(ctx context.Context, tokenID int64) error
	getByAccountNoFn     func(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.LoginHistory, int64, error)
	getAllFn             func(ctx context.Context, pagination model.Pagination) ([]model.LoginHistory, int64, error)
	getLatestByAccountFn func(ctx context.Context, accountNo int64) (*model.LoginHistory, error)
}

func (m *mockLoginHistoryRepo) RecordLogin(ctx context.Context, accountNo int64, ipAddress string) (int64, error) {
	if m.recordLoginFn != nil {
		return m.recordLoginFn(ctx, accountNo, ipAddress)
	}
	return 1, nil
}

func (m *mockLoginHistoryRepo) RecordLogout(ctx context.Context, tokenID int64) error {
	if m.recordLogoutFn != nil {
		return m.recordLogoutFn(ctx, tokenID)
	}
	return nil
}

func (m *mockLoginHistoryRepo) GetByAccountNo(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.LoginHistory, int64, error) {
	if m.getByAccountNoFn != nil {
		return m.getByAccountNoFn(ctx, accountNo, pagination)
	}
	return []model.LoginHistory{}, 0, nil
}

func (m *mockLoginHistoryRepo) GetAll(ctx context.Context, pagination model.Pagination) ([]model.LoginHistory, int64, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx, pagination)
	}
	return []model.LoginHistory{}, 0, nil
}

func (m *mockLoginHistoryRepo) GetLatestByAccountNo(ctx context.Context, accountNo int64) (*model.LoginHistory, error) {
	if m.getLatestByAccountFn != nil {
		return m.getLatestByAccountFn(ctx, accountNo)
	}
	return nil, nil
}

// mockHasher implements security.PasswordHasher for testing.
type mockHasher struct {
	hashFn    func(password string) (string, error)
	compareFn func(hashedPassword, plainPassword string) error
}

func (m *mockHasher) Hash(password string) (string, error) {
	if m.hashFn != nil {
		return m.hashFn(password)
	}
	return "$2a$10$hashedpassword", nil
}

func (m *mockHasher) Compare(hashedPassword, plainPassword string) error {
	if m.compareFn != nil {
		return m.compareFn(hashedPassword, plainPassword)
	}
	return nil
}
