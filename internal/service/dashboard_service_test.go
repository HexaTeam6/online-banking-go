package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTransactionRepository is a test double for TransactionRepository.
type mockTransactionRepository struct {
	getSummaryFn        func(ctx context.Context, accountNo int64) (*model.TransactionSummary, error)
	getMonthlySummaryFn func(ctx context.Context, accountNo int64, year int, month int) (*model.TransactionSummary, error)
	getByAccountNoFn    func(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.Transaction, int64, error)
}

func (m *mockTransactionRepository) Create(ctx context.Context, tx *sqlx.Tx, transaction *model.Transaction) error {
	return nil
}

func (m *mockTransactionRepository) GetByAccountNo(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.Transaction, int64, error) {
	if m.getByAccountNoFn != nil {
		return m.getByAccountNoFn(ctx, accountNo, pagination)
	}
	return nil, 0, nil
}

func (m *mockTransactionRepository) GetAll(ctx context.Context, pagination model.Pagination) ([]model.Transaction, int64, error) {
	return nil, 0, nil
}

func (m *mockTransactionRepository) GetSummary(ctx context.Context, accountNo int64) (*model.TransactionSummary, error) {
	if m.getSummaryFn != nil {
		return m.getSummaryFn(ctx, accountNo)
	}
	return nil, nil
}

func (m *mockTransactionRepository) GetMonthlySummary(ctx context.Context, accountNo int64, year int, month int) (*model.TransactionSummary, error) {
	if m.getMonthlySummaryFn != nil {
		return m.getMonthlySummaryFn(ctx, accountNo, year, month)
	}
	return nil, nil
}

// mockBalanceRepository is a test double for BalanceRepository.
type mockBalanceRepository struct {
	getByAccountNoFn func(ctx context.Context, accountNo int64) (int64, error)
}

func (m *mockBalanceRepository) GetByAccountNo(ctx context.Context, accountNo int64) (int64, error) {
	if m.getByAccountNoFn != nil {
		return m.getByAccountNoFn(ctx, accountNo)
	}
	return 0, nil
}

func (m *mockBalanceRepository) Credit(ctx context.Context, tx *sqlx.Tx, accountNo int64, amount int64) (int64, error) {
	return 0, nil
}

func (m *mockBalanceRepository) Debit(ctx context.Context, tx *sqlx.Tx, accountNo int64, amount int64) (int64, error) {
	return 0, nil
}

func (m *mockBalanceRepository) Create(ctx context.Context, tx *sqlx.Tx, accountNo int64, accountType string, balance int64) error {
	return nil
}

func TestGetDashboard_Success(t *testing.T) {
	transactionRepo := &mockTransactionRepository{
		getSummaryFn: func(ctx context.Context, accountNo int64) (*model.TransactionSummary, error) {
			return &model.TransactionSummary{
				TransactionCount: 10,
				TotalCredit:      50000,
				TotalDebit:       30000,
			}, nil
		},
		getMonthlySummaryFn: func(ctx context.Context, accountNo int64, year int, month int) (*model.TransactionSummary, error) {
			return &model.TransactionSummary{
				TransactionCount: 3,
				TotalCredit:      15000,
				TotalDebit:       5000,
			}, nil
		},
	}

	balanceRepo := &mockBalanceRepository{
		getByAccountNoFn: func(ctx context.Context, accountNo int64) (int64, error) {
			return 20000, nil
		},
	}

	svc := NewDashboardService(transactionRepo, balanceRepo)

	result, err := svc.GetDashboard(context.Background(), 123456789)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(10), result.AllTime.TransactionCount)
	assert.Equal(t, int64(50000), result.AllTime.TotalCredit)
	assert.Equal(t, int64(30000), result.AllTime.TotalDebit)
	assert.Equal(t, int64(3), result.CurrentMonth.TransactionCount)
	assert.Equal(t, int64(15000), result.CurrentMonth.TotalCredit)
	assert.Equal(t, int64(5000), result.CurrentMonth.TotalDebit)
	assert.Equal(t, int64(20000), result.CurrentBalance)
}

func TestGetDashboard_NoTransactions(t *testing.T) {
	transactionRepo := &mockTransactionRepository{
		getSummaryFn: func(ctx context.Context, accountNo int64) (*model.TransactionSummary, error) {
			return &model.TransactionSummary{
				TransactionCount: 0,
				TotalCredit:      0,
				TotalDebit:       0,
			}, nil
		},
		getMonthlySummaryFn: func(ctx context.Context, accountNo int64, year int, month int) (*model.TransactionSummary, error) {
			return &model.TransactionSummary{
				TransactionCount: 0,
				TotalCredit:      0,
				TotalDebit:       0,
			}, nil
		},
	}

	balanceRepo := &mockBalanceRepository{
		getByAccountNoFn: func(ctx context.Context, accountNo int64) (int64, error) {
			return 0, nil
		},
	}

	svc := NewDashboardService(transactionRepo, balanceRepo)

	result, err := svc.GetDashboard(context.Background(), 123456789)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(0), result.AllTime.TransactionCount)
	assert.Equal(t, int64(0), result.AllTime.TotalCredit)
	assert.Equal(t, int64(0), result.AllTime.TotalDebit)
	assert.Equal(t, int64(0), result.CurrentMonth.TransactionCount)
	assert.Equal(t, int64(0), result.CurrentBalance)
}

func TestGetDashboard_SummaryError(t *testing.T) {
	transactionRepo := &mockTransactionRepository{
		getSummaryFn: func(ctx context.Context, accountNo int64) (*model.TransactionSummary, error) {
			return nil, errors.New("database connection error")
		},
	}

	balanceRepo := &mockBalanceRepository{}

	svc := NewDashboardService(transactionRepo, balanceRepo)

	result, err := svc.GetDashboard(context.Background(), 123456789)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "fetching all-time summary")
}

func TestGetDashboard_MonthlySummaryError(t *testing.T) {
	transactionRepo := &mockTransactionRepository{
		getSummaryFn: func(ctx context.Context, accountNo int64) (*model.TransactionSummary, error) {
			return &model.TransactionSummary{}, nil
		},
		getMonthlySummaryFn: func(ctx context.Context, accountNo int64, year int, month int) (*model.TransactionSummary, error) {
			return nil, errors.New("query timeout")
		},
	}

	balanceRepo := &mockBalanceRepository{}

	svc := NewDashboardService(transactionRepo, balanceRepo)

	result, err := svc.GetDashboard(context.Background(), 123456789)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "fetching monthly summary")
}

func TestGetDashboard_BalanceError(t *testing.T) {
	transactionRepo := &mockTransactionRepository{
		getSummaryFn: func(ctx context.Context, accountNo int64) (*model.TransactionSummary, error) {
			return &model.TransactionSummary{}, nil
		},
		getMonthlySummaryFn: func(ctx context.Context, accountNo int64, year int, month int) (*model.TransactionSummary, error) {
			return &model.TransactionSummary{}, nil
		},
	}

	balanceRepo := &mockBalanceRepository{
		getByAccountNoFn: func(ctx context.Context, accountNo int64) (int64, error) {
			return 0, errors.New("account not found")
		},
	}

	svc := NewDashboardService(transactionRepo, balanceRepo)

	result, err := svc.GetDashboard(context.Background(), 123456789)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "fetching balance")
}

func TestGetTransactionHistory_Success(t *testing.T) {
	expectedTransactions := []model.Transaction{
		{
			TransID:    1,
			TransDate:  time.Now(),
			Amount:     5000,
			TransType:  "CREDIT",
			Purpose:    "Transfer from friend",
			ToAccount:  987654321,
			AccountNo:  123456789,
			AccountBal: 25000,
		},
		{
			TransID:    2,
			TransDate:  time.Now().Add(-24 * time.Hour),
			Amount:     2000,
			TransType:  "DEBIT",
			Purpose:    "Bill payment",
			ToAccount:  111222333,
			AccountNo:  123456789,
			AccountBal: 20000,
		},
	}

	transactionRepo := &mockTransactionRepository{
		getByAccountNoFn: func(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.Transaction, int64, error) {
			assert.Equal(t, int64(123456789), accountNo)
			assert.Equal(t, 1, pagination.Page)
			assert.Equal(t, 10, pagination.PageSize)
			return expectedTransactions, 2, nil
		},
	}

	balanceRepo := &mockBalanceRepository{}

	svc := NewDashboardService(transactionRepo, balanceRepo)

	transactions, totalCount, err := svc.GetTransactionHistory(context.Background(), 123456789, model.Pagination{Page: 1, PageSize: 10})

	require.NoError(t, err)
	assert.Equal(t, int64(2), totalCount)
	assert.Len(t, transactions, 2)
	assert.Equal(t, expectedTransactions[0].TransID, transactions[0].TransID)
	assert.Equal(t, expectedTransactions[1].TransID, transactions[1].TransID)
}

func TestGetTransactionHistory_EmptyResult(t *testing.T) {
	transactionRepo := &mockTransactionRepository{
		getByAccountNoFn: func(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.Transaction, int64, error) {
			return []model.Transaction{}, 0, nil
		},
	}

	balanceRepo := &mockBalanceRepository{}

	svc := NewDashboardService(transactionRepo, balanceRepo)

	transactions, totalCount, err := svc.GetTransactionHistory(context.Background(), 123456789, model.Pagination{Page: 1, PageSize: 10})

	require.NoError(t, err)
	assert.Equal(t, int64(0), totalCount)
	assert.Empty(t, transactions)
}

func TestGetTransactionHistory_RepositoryError(t *testing.T) {
	transactionRepo := &mockTransactionRepository{
		getByAccountNoFn: func(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.Transaction, int64, error) {
			return nil, 0, errors.New("database error")
		},
	}

	balanceRepo := &mockBalanceRepository{}

	svc := NewDashboardService(transactionRepo, balanceRepo)

	transactions, totalCount, err := svc.GetTransactionHistory(context.Background(), 123456789, model.Pagination{Page: 1, PageSize: 10})

	require.Error(t, err)
	assert.Nil(t, transactions)
	assert.Equal(t, int64(0), totalCount)
	assert.Contains(t, err.Error(), "fetching transaction history")
}
