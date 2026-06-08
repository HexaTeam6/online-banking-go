package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/abdurrachmanwahed/online-banking/internal/repository"
	"github.com/jmoiron/sqlx"
)

// AdminService handles admin operations including listing resources and balance adjustments.
type AdminService interface {
	ListCustomers(ctx context.Context, pagination model.Pagination) ([]model.Customer, int64, error)
	AdjustBalance(ctx context.Context, req model.BalanceAdjustmentRequest) error
	ListTransactions(ctx context.Context, pagination model.Pagination) ([]model.Transaction, int64, error)
	ListRequests(ctx context.Context, pagination model.Pagination) ([]model.MoneyRequest, int64, error)
	ListFeedback(ctx context.Context, pagination model.Pagination) ([]model.Feedback, int64, error)
	ListLoginHistory(ctx context.Context, pagination model.Pagination) ([]model.LoginHistory, int64, error)
}

// AdminServiceDeps holds all dependencies required by AdminService.
type AdminServiceDeps struct {
	CustomerRepo     repository.CustomerRepository
	BalanceRepo      repository.BalanceRepository
	TransactionRepo  repository.TransactionRepository
	RequestRepo      repository.RequestRepository
	FeedbackRepo     repository.FeedbackRepository
	LoginHistoryRepo repository.LoginHistoryRepository
	AccountRepo      repository.AccountRepository
	DB               *sqlx.DB
}

// adminService is the concrete implementation of AdminService.
type adminService struct {
	customerRepo     repository.CustomerRepository
	balanceRepo      repository.BalanceRepository
	transactionRepo  repository.TransactionRepository
	requestRepo      repository.RequestRepository
	feedbackRepo     repository.FeedbackRepository
	loginHistoryRepo repository.LoginHistoryRepository
	accountRepo      repository.AccountRepository
	db               *sqlx.DB
}

// NewAdminService creates a new AdminService with all required dependencies injected.
func NewAdminService(deps AdminServiceDeps) AdminService {
	return &adminService{
		customerRepo:     deps.CustomerRepo,
		balanceRepo:      deps.BalanceRepo,
		transactionRepo:  deps.TransactionRepo,
		requestRepo:      deps.RequestRepo,
		feedbackRepo:     deps.FeedbackRepo,
		loginHistoryRepo: deps.LoginHistoryRepo,
		accountRepo:      deps.AccountRepo,
		db:               deps.DB,
	}
}

// ListCustomers returns a paginated list of all customers.
func (s *adminService) ListCustomers(ctx context.Context, pagination model.Pagination) ([]model.Customer, int64, error) {
	customers, total, err := s.customerRepo.ListAll(ctx, pagination)
	if err != nil {
		return nil, 0, fmt.Errorf("list customers: %w", err)
	}
	return customers, total, nil
}

// AdjustBalance performs a credit or debit operation on the specified account atomically.
// It verifies the account exists, checks sufficient balance for debits, updates the balance,
// and creates a transaction record within a single database transaction.
func (s *adminService) AdjustBalance(ctx context.Context, req model.BalanceAdjustmentRequest) error {
	// Verify account exists
	exists, err := s.accountRepo.Exists(ctx, req.AccountNo)
	if err != nil {
		return fmt.Errorf("check account existence: %w", err)
	}
	if !exists {
		return model.ErrAccountNotFound
	}

	// For debit operations, verify sufficient balance before starting the transaction
	operation := strings.ToLower(req.Operation)
	if operation == "debit" {
		if err := s.verifyDebitBalance(ctx, req.AccountNo, req.Amount); err != nil {
			return err
		}
	}

	// Execute the balance adjustment atomically
	return s.executeAdjustment(ctx, req, operation)
}

// verifyDebitBalance checks that the account has sufficient funds for a debit operation.
func (s *adminService) verifyDebitBalance(ctx context.Context, accountNo int64, amount int64) error {
	currentBalance, err := s.balanceRepo.GetByAccountNo(ctx, accountNo)
	if err != nil {
		return fmt.Errorf("get balance for debit check: %w", err)
	}
	if currentBalance < amount {
		return model.ErrInsufficientBalance
	}
	return nil
}

// executeAdjustment performs the balance update and transaction record creation within a DB transaction.
func (s *adminService) executeAdjustment(ctx context.Context, req model.BalanceAdjustmentRequest, operation string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Apply balance operation
	var newBalance int64
	if operation == "credit" {
		newBalance, err = s.balanceRepo.Credit(ctx, tx, req.AccountNo, req.Amount)
	} else {
		newBalance, err = s.balanceRepo.Debit(ctx, tx, req.AccountNo, req.Amount)
	}
	if err != nil {
		return fmt.Errorf("%s balance: %w", operation, err)
	}

	// Create transaction record
	transaction := &model.Transaction{
		TransDate:  time.Now(),
		Amount:     req.Amount,
		TransType:  strings.ToUpper(operation),
		Purpose:    req.Purpose,
		ToAccount:  req.AccountNo,
		AccountNo:  req.AccountNo,
		AccountBal: newBalance,
	}

	if createErr := s.transactionRepo.Create(ctx, tx, transaction); createErr != nil {
		err = fmt.Errorf("create transaction record: %w", createErr)
		return err
	}

	// Commit
	if commitErr := tx.Commit(); commitErr != nil {
		err = fmt.Errorf("commit transaction: %w", commitErr)
		return err
	}

	return nil
}

// ListTransactions returns a paginated list of all transactions across all accounts.
func (s *adminService) ListTransactions(ctx context.Context, pagination model.Pagination) ([]model.Transaction, int64, error) {
	transactions, total, err := s.transactionRepo.GetAll(ctx, pagination)
	if err != nil {
		return nil, 0, fmt.Errorf("list transactions: %w", err)
	}
	return transactions, total, nil
}

// ListRequests returns a paginated list of all money requests across all accounts.
func (s *adminService) ListRequests(ctx context.Context, pagination model.Pagination) ([]model.MoneyRequest, int64, error) {
	requests, total, err := s.requestRepo.GetAll(ctx, pagination)
	if err != nil {
		return nil, 0, fmt.Errorf("list requests: %w", err)
	}
	return requests, total, nil
}

// ListFeedback returns a paginated list of all customer feedback entries.
func (s *adminService) ListFeedback(ctx context.Context, pagination model.Pagination) ([]model.Feedback, int64, error) {
	feedback, total, err := s.feedbackRepo.GetAll(ctx, pagination)
	if err != nil {
		return nil, 0, fmt.Errorf("list feedback: %w", err)
	}
	return feedback, total, nil
}

// ListLoginHistory returns a paginated list of all login/logout records.
func (s *adminService) ListLoginHistory(ctx context.Context, pagination model.Pagination) ([]model.LoginHistory, int64, error) {
	history, total, err := s.loginHistoryRepo.GetAll(ctx, pagination)
	if err != nil {
		return nil, 0, fmt.Errorf("list login history: %w", err)
	}
	return history, total, nil
}
