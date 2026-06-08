package service

import (
	"context"
	"fmt"
	"time"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/abdurrachmanwahed/online-banking/internal/repository"
)

// DashboardService defines the interface for dashboard data aggregation.
type DashboardService interface {
	GetDashboard(ctx context.Context, accountNo int64) (*model.DashboardResponse, error)
	GetTransactionHistory(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.Transaction, int64, error)
}

// dashboardService is the concrete implementation of DashboardService.
type dashboardService struct {
	transactionRepo repository.TransactionRepository
	balanceRepo     repository.BalanceRepository
}

// NewDashboardService creates a new DashboardService with injected dependencies.
func NewDashboardService(
	transactionRepo repository.TransactionRepository,
	balanceRepo repository.BalanceRepository,
) DashboardService {
	return &dashboardService{
		transactionRepo: transactionRepo,
		balanceRepo:     balanceRepo,
	}
}

// GetDashboard aggregates all-time transaction stats, current month stats, and the current balance
// for the given account number.
func (s *dashboardService) GetDashboard(ctx context.Context, accountNo int64) (*model.DashboardResponse, error) {
	// Get all-time transaction summary
	allTimeSummary, err := s.transactionRepo.GetSummary(ctx, accountNo)
	if err != nil {
		return nil, fmt.Errorf("fetching all-time summary: %w", err)
	}

	// Get current month transaction summary
	now := time.Now()
	monthlySummary, err := s.transactionRepo.GetMonthlySummary(ctx, accountNo, now.Year(), int(now.Month()))
	if err != nil {
		return nil, fmt.Errorf("fetching monthly summary: %w", err)
	}

	// Get current balance
	balance, err := s.balanceRepo.GetByAccountNo(ctx, accountNo)
	if err != nil {
		return nil, fmt.Errorf("fetching balance: %w", err)
	}

	return &model.DashboardResponse{
		AllTime:        *allTimeSummary,
		CurrentMonth:   *monthlySummary,
		CurrentBalance: balance,
	}, nil
}

// GetTransactionHistory returns a paginated list of transactions for the given account,
// ordered by date descending.
func (s *dashboardService) GetTransactionHistory(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.Transaction, int64, error) {
	transactions, totalCount, err := s.transactionRepo.GetByAccountNo(ctx, accountNo, pagination)
	if err != nil {
		return nil, 0, fmt.Errorf("fetching transaction history: %w", err)
	}

	return transactions, totalCount, nil
}
