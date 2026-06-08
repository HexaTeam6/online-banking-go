package service

import (
	"context"
	"errors"
	"testing"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockReqSvcRequestRepo is a test double for repository.RequestRepository.
type mockReqSvcRequestRepo struct {
	createFn           func(ctx context.Context, request *model.MoneyRequest) error
	getReceivedByAccFn func(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.MoneyRequest, int64, error)
	markAsViewedFn     func(ctx context.Context, requestID int64) error
	getAllFn           func(ctx context.Context, pagination model.Pagination) ([]model.MoneyRequest, int64, error)
}

func (m *mockReqSvcRequestRepo) Create(ctx context.Context, request *model.MoneyRequest) error {
	if m.createFn != nil {
		return m.createFn(ctx, request)
	}
	return nil
}

func (m *mockReqSvcRequestRepo) GetReceivedByAccountNo(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.MoneyRequest, int64, error) {
	if m.getReceivedByAccFn != nil {
		return m.getReceivedByAccFn(ctx, accountNo, pagination)
	}
	return []model.MoneyRequest{}, 0, nil
}

func (m *mockReqSvcRequestRepo) MarkAsViewed(ctx context.Context, requestID int64) error {
	if m.markAsViewedFn != nil {
		return m.markAsViewedFn(ctx, requestID)
	}
	return nil
}

func (m *mockReqSvcRequestRepo) GetAll(ctx context.Context, pagination model.Pagination) ([]model.MoneyRequest, int64, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx, pagination)
	}
	return []model.MoneyRequest{}, 0, nil
}

// mockReqSvcAccountRepo is a test double for repository.AccountRepository.
type mockReqSvcAccountRepo struct {
	existsFn func(ctx context.Context, accountNo int64) (bool, error)
}

func (m *mockReqSvcAccountRepo) GetByUsername(_ context.Context, _ string) (*model.Account, error) {
	return nil, nil
}

func (m *mockReqSvcAccountRepo) GetByAccountNo(_ context.Context, _ int64) (*model.Account, error) {
	return nil, nil
}

func (m *mockReqSvcAccountRepo) Exists(ctx context.Context, accountNo int64) (bool, error) {
	if m.existsFn != nil {
		return m.existsFn(ctx, accountNo)
	}
	return true, nil
}

func (m *mockReqSvcAccountRepo) UsernameExists(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (m *mockReqSvcAccountRepo) Create(_ context.Context, _ *sqlx.Tx, _ *model.Account) (int64, error) {
	return 0, nil
}

func (m *mockReqSvcAccountRepo) UpdatePassword(_ context.Context, _ int64, _ string) error {
	return nil
}

func TestCreateRequest_Success(t *testing.T) {
	var createdRequest *model.MoneyRequest

	reqRepo := &mockReqSvcRequestRepo{
		createFn: func(_ context.Context, request *model.MoneyRequest) error {
			createdRequest = request
			return nil
		},
	}
	accRepo := &mockReqSvcAccountRepo{
		existsFn: func(_ context.Context, _ int64) (bool, error) {
			return true, nil
		},
	}

	svc := NewRequestService(reqRepo, accRepo)

	err := svc.CreateRequest(context.Background(), 100000001, model.CreateMoneyRequest{
		ToAccount: 100000002,
		Amount:    1000,
		Message:   "Please pay me back",
	})

	require.NoError(t, err)
	assert.NotNil(t, createdRequest)
	assert.Equal(t, int64(100000001), createdRequest.AccountNo)
	assert.Equal(t, int64(100000002), createdRequest.ToAccount)
	assert.Equal(t, int64(1000), createdRequest.Amount)
	assert.Equal(t, "Please pay me back", createdRequest.Message)
	assert.Equal(t, "PENDING", createdRequest.Status)
	assert.False(t, createdRequest.HasViewed)
}

func TestCreateRequest_AmountBelowMinimum(t *testing.T) {
	reqRepo := &mockReqSvcRequestRepo{}
	accRepo := &mockReqSvcAccountRepo{}

	svc := NewRequestService(reqRepo, accRepo)

	err := svc.CreateRequest(context.Background(), 100000001, model.CreateMoneyRequest{
		ToAccount: 100000002,
		Amount:    499,
		Message:   "Too little",
	})

	require.Error(t, err)
	assert.Equal(t, model.ErrLimitViolation, err)
}

func TestCreateRequest_AmountAboveMaximum(t *testing.T) {
	reqRepo := &mockReqSvcRequestRepo{}
	accRepo := &mockReqSvcAccountRepo{}

	svc := NewRequestService(reqRepo, accRepo)

	err := svc.CreateRequest(context.Background(), 100000001, model.CreateMoneyRequest{
		ToAccount: 100000002,
		Amount:    20001,
		Message:   "Too much",
	})

	require.Error(t, err)
	assert.Equal(t, model.ErrLimitViolation, err)
}

func TestCreateRequest_AmountAtMinimumBoundary(t *testing.T) {
	reqRepo := &mockReqSvcRequestRepo{
		createFn: func(_ context.Context, _ *model.MoneyRequest) error {
			return nil
		},
	}
	accRepo := &mockReqSvcAccountRepo{
		existsFn: func(_ context.Context, _ int64) (bool, error) {
			return true, nil
		},
	}

	svc := NewRequestService(reqRepo, accRepo)

	err := svc.CreateRequest(context.Background(), 100000001, model.CreateMoneyRequest{
		ToAccount: 100000002,
		Amount:    500,
		Message:   "Minimum allowed",
	})

	require.NoError(t, err)
}

func TestCreateRequest_AmountAtMaximumBoundary(t *testing.T) {
	reqRepo := &mockReqSvcRequestRepo{
		createFn: func(_ context.Context, _ *model.MoneyRequest) error {
			return nil
		},
	}
	accRepo := &mockReqSvcAccountRepo{
		existsFn: func(_ context.Context, _ int64) (bool, error) {
			return true, nil
		},
	}

	svc := NewRequestService(reqRepo, accRepo)

	err := svc.CreateRequest(context.Background(), 100000001, model.CreateMoneyRequest{
		ToAccount: 100000002,
		Amount:    20000,
		Message:   "Maximum allowed",
	})

	require.NoError(t, err)
}

func TestCreateRequest_SelfRequest(t *testing.T) {
	reqRepo := &mockReqSvcRequestRepo{}
	accRepo := &mockReqSvcAccountRepo{}

	svc := NewRequestService(reqRepo, accRepo)

	err := svc.CreateRequest(context.Background(), 100000001, model.CreateMoneyRequest{
		ToAccount: 100000001,
		Amount:    1000,
		Message:   "Self request",
	})

	require.Error(t, err)
	assert.Equal(t, model.ErrSelfTransfer, err)
}

func TestCreateRequest_TargetAccountNotFound(t *testing.T) {
	reqRepo := &mockReqSvcRequestRepo{}
	accRepo := &mockReqSvcAccountRepo{
		existsFn: func(_ context.Context, _ int64) (bool, error) {
			return false, nil
		},
	}

	svc := NewRequestService(reqRepo, accRepo)

	err := svc.CreateRequest(context.Background(), 100000001, model.CreateMoneyRequest{
		ToAccount: 999999999,
		Amount:    1000,
		Message:   "Non-existent account",
	})

	require.Error(t, err)
	assert.Equal(t, model.ErrAccountNotFound, err)
}

func TestCreateRequest_RepositoryError(t *testing.T) {
	reqRepo := &mockReqSvcRequestRepo{}
	accRepo := &mockReqSvcAccountRepo{
		existsFn: func(_ context.Context, _ int64) (bool, error) {
			return false, errors.New("database connection lost")
		},
	}

	svc := NewRequestService(reqRepo, accRepo)

	err := svc.CreateRequest(context.Background(), 100000001, model.CreateMoneyRequest{
		ToAccount: 100000002,
		Amount:    1000,
		Message:   "DB error",
	})

	require.Error(t, err)
	assert.NotEqual(t, model.ErrAccountNotFound, err)
	assert.Contains(t, err.Error(), "database connection lost")
}

func TestGetReceivedRequests_Success(t *testing.T) {
	expectedRequests := []model.MoneyRequest{
		{RequestID: 1, AccountNo: 100000002, ToAccount: 100000001, Amount: 1000, Status: "PENDING"},
		{RequestID: 2, AccountNo: 100000003, ToAccount: 100000001, Amount: 500, Status: "PENDING"},
	}

	reqRepo := &mockReqSvcRequestRepo{
		getReceivedByAccFn: func(_ context.Context, accountNo int64, pagination model.Pagination) ([]model.MoneyRequest, int64, error) {
			assert.Equal(t, int64(100000001), accountNo)
			assert.Equal(t, 1, pagination.Page)
			assert.Equal(t, 10, pagination.PageSize)
			return expectedRequests, 2, nil
		},
	}
	accRepo := &mockReqSvcAccountRepo{}

	svc := NewRequestService(reqRepo, accRepo)

	requests, total, err := svc.GetReceivedRequests(context.Background(), 100000001, model.Pagination{Page: 1, PageSize: 10})

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, requests, 2)
	assert.Equal(t, expectedRequests, requests)
}

func TestMarkAsViewed_Success(t *testing.T) {
	var viewedID int64

	reqRepo := &mockReqSvcRequestRepo{
		markAsViewedFn: func(_ context.Context, requestID int64) error {
			viewedID = requestID
			return nil
		},
	}
	accRepo := &mockReqSvcAccountRepo{}

	svc := NewRequestService(reqRepo, accRepo)

	err := svc.MarkAsViewed(context.Background(), 42)

	require.NoError(t, err)
	assert.Equal(t, int64(42), viewedID)
}

func TestMarkAsViewed_RepositoryError(t *testing.T) {
	reqRepo := &mockReqSvcRequestRepo{
		markAsViewedFn: func(_ context.Context, _ int64) error {
			return errors.New("update failed")
		},
	}
	accRepo := &mockReqSvcAccountRepo{}

	svc := NewRequestService(reqRepo, accRepo)

	err := svc.MarkAsViewed(context.Background(), 42)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")
}
