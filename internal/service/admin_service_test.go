package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAdminRequestRepo is a test double for RequestRepository used in admin service tests.
type mockAdminRequestRepo struct {
	createFn           func(ctx context.Context, request *model.MoneyRequest) error
	getReceivedByAccFn func(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.MoneyRequest, int64, error)
	markAsViewedFn     func(ctx context.Context, requestID int64) error
	getAllFn           func(ctx context.Context, pagination model.Pagination) ([]model.MoneyRequest, int64, error)
}

func (m *mockAdminRequestRepo) Create(ctx context.Context, request *model.MoneyRequest) error {
	if m.createFn != nil {
		return m.createFn(ctx, request)
	}
	return nil
}

func (m *mockAdminRequestRepo) GetReceivedByAccountNo(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.MoneyRequest, int64, error) {
	if m.getReceivedByAccFn != nil {
		return m.getReceivedByAccFn(ctx, accountNo, pagination)
	}
	return []model.MoneyRequest{}, 0, nil
}

func (m *mockAdminRequestRepo) MarkAsViewed(ctx context.Context, requestID int64) error {
	if m.markAsViewedFn != nil {
		return m.markAsViewedFn(ctx, requestID)
	}
	return nil
}

func (m *mockAdminRequestRepo) GetAll(ctx context.Context, pagination model.Pagination) ([]model.MoneyRequest, int64, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx, pagination)
	}
	return []model.MoneyRequest{}, 0, nil
}

// mockAdminFeedbackRepo is a test double for FeedbackRepository used in admin service tests.
type mockAdminFeedbackRepo struct {
	createFn func(ctx context.Context, feedback *model.Feedback) error
	getAllFn func(ctx context.Context, pagination model.Pagination) ([]model.Feedback, int64, error)
}

func (m *mockAdminFeedbackRepo) Create(ctx context.Context, feedback *model.Feedback) error {
	if m.createFn != nil {
		return m.createFn(ctx, feedback)
	}
	return nil
}

func (m *mockAdminFeedbackRepo) GetAll(ctx context.Context, pagination model.Pagination) ([]model.Feedback, int64, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx, pagination)
	}
	return []model.Feedback{}, 0, nil
}

// --- Tests ---

func TestAdminService_ListCustomers(t *testing.T) {
	expectedCustomers := []model.Customer{
		{AccountNo: 123456789, FullName: "John Doe", Email: "john@example.com"},
		{AccountNo: 987654321, FullName: "Jane Smith", Email: "jane@example.com"},
	}

	customerRepo := &mockCustomerRepo{
		listAllFn: func(ctx context.Context, pagination model.Pagination) ([]model.Customer, int64, error) {
			assert.Equal(t, 1, pagination.Page)
			assert.Equal(t, 20, pagination.PageSize)
			return expectedCustomers, 2, nil
		},
	}

	svc := NewAdminService(AdminServiceDeps{
		CustomerRepo:     customerRepo,
		BalanceRepo:      &mockBalanceRepo{},
		TransactionRepo:  &mockTransactionRepo{},
		RequestRepo:      &mockAdminRequestRepo{},
		FeedbackRepo:     &mockAdminFeedbackRepo{},
		LoginHistoryRepo: &mockLoginHistoryRepo{},
		AccountRepo:      &mockAccountRepo{},
		DB:               nil,
	})

	customers, total, err := svc.ListCustomers(context.Background(), model.Pagination{Page: 1, PageSize: 20})

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, expectedCustomers, customers)
}

func TestAdminService_ListCustomers_Error(t *testing.T) {
	customerRepo := &mockCustomerRepo{
		listAllFn: func(ctx context.Context, pagination model.Pagination) ([]model.Customer, int64, error) {
			return nil, 0, errors.New("db connection failed")
		},
	}

	svc := NewAdminService(AdminServiceDeps{
		CustomerRepo:     customerRepo,
		BalanceRepo:      &mockBalanceRepo{},
		TransactionRepo:  &mockTransactionRepo{},
		RequestRepo:      &mockAdminRequestRepo{},
		FeedbackRepo:     &mockAdminFeedbackRepo{},
		LoginHistoryRepo: &mockLoginHistoryRepo{},
		AccountRepo:      &mockAccountRepo{},
		DB:               nil,
	})

	_, _, err := svc.ListCustomers(context.Background(), model.Pagination{Page: 1, PageSize: 20})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "list customers")
}

func TestAdminService_ListTransactions(t *testing.T) {
	expectedTx := []model.Transaction{
		{TransID: 1, Amount: 1000, TransType: "CREDIT", Purpose: "Test", AccountNo: 123456789},
	}

	transactionRepo := &mockTransactionRepo{
		getAllFn: func(ctx context.Context, pagination model.Pagination) ([]model.Transaction, int64, error) {
			return expectedTx, 1, nil
		},
	}

	svc := NewAdminService(AdminServiceDeps{
		CustomerRepo:     &mockCustomerRepo{},
		BalanceRepo:      &mockBalanceRepo{},
		TransactionRepo:  transactionRepo,
		RequestRepo:      &mockAdminRequestRepo{},
		FeedbackRepo:     &mockAdminFeedbackRepo{},
		LoginHistoryRepo: &mockLoginHistoryRepo{},
		AccountRepo:      &mockAccountRepo{},
		DB:               nil,
	})

	transactions, total, err := svc.ListTransactions(context.Background(), model.Pagination{Page: 1, PageSize: 20})

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, expectedTx, transactions)
}

func TestAdminService_ListRequests(t *testing.T) {
	now := time.Now()
	expectedReqs := []model.MoneyRequest{
		{RequestID: 1, AccountNo: 123, ToAccount: 456, Amount: 1000, Status: "PENDING", RequestDate: now},
	}

	requestRepo := &mockAdminRequestRepo{
		getAllFn: func(ctx context.Context, pagination model.Pagination) ([]model.MoneyRequest, int64, error) {
			return expectedReqs, 1, nil
		},
	}

	svc := NewAdminService(AdminServiceDeps{
		CustomerRepo:     &mockCustomerRepo{},
		BalanceRepo:      &mockBalanceRepo{},
		TransactionRepo:  &mockTransactionRepo{},
		RequestRepo:      requestRepo,
		FeedbackRepo:     &mockAdminFeedbackRepo{},
		LoginHistoryRepo: &mockLoginHistoryRepo{},
		AccountRepo:      &mockAccountRepo{},
		DB:               nil,
	})

	requests, total, err := svc.ListRequests(context.Background(), model.Pagination{Page: 1, PageSize: 20})

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, expectedReqs, requests)
}

func TestAdminService_ListFeedback(t *testing.T) {
	now := time.Now()
	expectedFb := []model.Feedback{
		{FeedbackID: 1, AccountNo: 123, FeedbackText: "Great service", Hearts: 5, Time: now},
	}

	feedbackRepo := &mockAdminFeedbackRepo{
		getAllFn: func(ctx context.Context, pagination model.Pagination) ([]model.Feedback, int64, error) {
			return expectedFb, 1, nil
		},
	}

	svc := NewAdminService(AdminServiceDeps{
		CustomerRepo:     &mockCustomerRepo{},
		BalanceRepo:      &mockBalanceRepo{},
		TransactionRepo:  &mockTransactionRepo{},
		RequestRepo:      &mockAdminRequestRepo{},
		FeedbackRepo:     feedbackRepo,
		LoginHistoryRepo: &mockLoginHistoryRepo{},
		AccountRepo:      &mockAccountRepo{},
		DB:               nil,
	})

	feedback, total, err := svc.ListFeedback(context.Background(), model.Pagination{Page: 1, PageSize: 20})

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, expectedFb, feedback)
}

func TestAdminService_ListLoginHistory(t *testing.T) {
	now := time.Now()
	expectedHistory := []model.LoginHistory{
		{TokenID: 1, AccountNo: 123, LoginTime: now, IPAddress: "192.168.1.1"},
	}

	loginHistoryRepo := &mockLoginHistoryRepo{
		getAllFn: func(ctx context.Context, pagination model.Pagination) ([]model.LoginHistory, int64, error) {
			return expectedHistory, 1, nil
		},
	}

	svc := NewAdminService(AdminServiceDeps{
		CustomerRepo:     &mockCustomerRepo{},
		BalanceRepo:      &mockBalanceRepo{},
		TransactionRepo:  &mockTransactionRepo{},
		RequestRepo:      &mockAdminRequestRepo{},
		FeedbackRepo:     &mockAdminFeedbackRepo{},
		LoginHistoryRepo: loginHistoryRepo,
		AccountRepo:      &mockAccountRepo{},
		DB:               nil,
	})

	history, total, err := svc.ListLoginHistory(context.Background(), model.Pagination{Page: 1, PageSize: 20})

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, expectedHistory, history)
}

func TestAdminService_AdjustBalance_AccountNotFound(t *testing.T) {
	accountRepo := &mockAccountRepo{
		existsFn: func(ctx context.Context, accountNo int64) (bool, error) {
			return false, nil
		},
	}

	svc := NewAdminService(AdminServiceDeps{
		CustomerRepo:     &mockCustomerRepo{},
		BalanceRepo:      &mockBalanceRepo{},
		TransactionRepo:  &mockTransactionRepo{},
		RequestRepo:      &mockAdminRequestRepo{},
		FeedbackRepo:     &mockAdminFeedbackRepo{},
		LoginHistoryRepo: &mockLoginHistoryRepo{},
		AccountRepo:      accountRepo,
		DB:               nil,
	})

	err := svc.AdjustBalance(context.Background(), model.BalanceAdjustmentRequest{
		AccountNo: 999999999,
		Operation: "credit",
		Amount:    1000,
		Purpose:   "Test credit",
	})

	require.Error(t, err)
	assert.Equal(t, model.ErrAccountNotFound, err)
}

func TestAdminService_AdjustBalance_DebitInsufficientBalance(t *testing.T) {
	accountRepo := &mockAccountRepo{
		existsFn: func(ctx context.Context, accountNo int64) (bool, error) {
			return true, nil
		},
	}
	balanceRepo := &mockBalanceRepo{
		getByAccountNoFn: func(ctx context.Context, accountNo int64) (int64, error) {
			return 500, nil // only 500 available
		},
	}

	svc := NewAdminService(AdminServiceDeps{
		CustomerRepo:     &mockCustomerRepo{},
		BalanceRepo:      balanceRepo,
		TransactionRepo:  &mockTransactionRepo{},
		RequestRepo:      &mockAdminRequestRepo{},
		FeedbackRepo:     &mockAdminFeedbackRepo{},
		LoginHistoryRepo: &mockLoginHistoryRepo{},
		AccountRepo:      accountRepo,
		DB:               nil,
	})

	err := svc.AdjustBalance(context.Background(), model.BalanceAdjustmentRequest{
		AccountNo: 123456789,
		Operation: "debit",
		Amount:    1000, // trying to debit more than available
		Purpose:   "Test debit",
	})

	require.Error(t, err)
	assert.Equal(t, model.ErrInsufficientBalance, err)
}

func TestAdminService_AdjustBalance_AccountExistsCheckError(t *testing.T) {
	accountRepo := &mockAccountRepo{
		existsFn: func(ctx context.Context, accountNo int64) (bool, error) {
			return false, errors.New("database error")
		},
	}

	svc := NewAdminService(AdminServiceDeps{
		CustomerRepo:     &mockCustomerRepo{},
		BalanceRepo:      &mockBalanceRepo{},
		TransactionRepo:  &mockTransactionRepo{},
		RequestRepo:      &mockAdminRequestRepo{},
		FeedbackRepo:     &mockAdminFeedbackRepo{},
		LoginHistoryRepo: &mockLoginHistoryRepo{},
		AccountRepo:      accountRepo,
		DB:               nil,
	})

	err := svc.AdjustBalance(context.Background(), model.BalanceAdjustmentRequest{
		AccountNo: 123456789,
		Operation: "credit",
		Amount:    1000,
		Purpose:   "Test",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "check account existence")
}

func TestAdminService_AdjustBalance_DebitBalanceCheckError(t *testing.T) {
	accountRepo := &mockAccountRepo{
		existsFn: func(ctx context.Context, accountNo int64) (bool, error) {
			return true, nil
		},
	}
	balanceRepo := &mockBalanceRepo{
		getByAccountNoFn: func(ctx context.Context, accountNo int64) (int64, error) {
			return 0, errors.New("balance fetch error")
		},
	}

	svc := NewAdminService(AdminServiceDeps{
		CustomerRepo:     &mockCustomerRepo{},
		BalanceRepo:      balanceRepo,
		TransactionRepo:  &mockTransactionRepo{},
		RequestRepo:      &mockAdminRequestRepo{},
		FeedbackRepo:     &mockAdminFeedbackRepo{},
		LoginHistoryRepo: &mockLoginHistoryRepo{},
		AccountRepo:      accountRepo,
		DB:               nil,
	})

	err := svc.AdjustBalance(context.Background(), model.BalanceAdjustmentRequest{
		AccountNo: 123456789,
		Operation: "debit",
		Amount:    1000,
		Purpose:   "Test debit",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "get balance for debit check")
}
