package service

import (
	"context"
	"fmt"
	"time"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/abdurrachmanwahed/online-banking/internal/repository"
	"github.com/jmoiron/sqlx"
)

// TransferService handles money transfers between accounts.
type TransferService interface {
	QuickTransfer(ctx context.Context, senderAccountNo int64, req model.TransferRequest) error
}

// transferService is the implementation of TransferService.
type transferService struct {
	accountRepo     repository.AccountRepository
	balanceRepo     repository.BalanceRepository
	transactionRepo repository.TransactionRepository
	db              *sqlx.DB
}

// NewTransferService creates a new TransferService with the required dependencies.
func NewTransferService(
	accountRepo repository.AccountRepository,
	balanceRepo repository.BalanceRepository,
	transactionRepo repository.TransactionRepository,
	db *sqlx.DB,
) TransferService {
	return &transferService{
		accountRepo:     accountRepo,
		balanceRepo:     balanceRepo,
		transactionRepo: transactionRepo,
		db:              db,
	}
}

// QuickTransfer executes an atomic money transfer between two accounts.
// It validates amount limits, verifies the destination account, prevents self-transfer,
// checks sufficient balance, and executes debit+credit in a single database transaction.
func (s *transferService) QuickTransfer(ctx context.Context, senderAccountNo int64, req model.TransferRequest) (err error) {
	// 1. Validate amount limits (500-20000)
	if req.Amount < 500 || req.Amount > 20000 {
		return model.ErrLimitViolation
	}

	// 2. Validate destination is not the sender (self-transfer)
	if req.ToAccount == senderAccountNo {
		return model.ErrSelfTransfer
	}

	// 3. Check destination account exists
	exists, err := s.accountRepo.Exists(ctx, req.ToAccount)
	if err != nil {
		return fmt.Errorf("checking destination account: %w", err)
	}
	if !exists {
		return model.ErrAccountNotFound
	}

	// 4. Begin database transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 5. Debit sender account
	senderNewBalance, err := s.balanceRepo.Debit(ctx, tx, senderAccountNo, req.Amount)
	if err != nil {
		return fmt.Errorf("debiting sender: %w", err)
	}

	// 6. Check for insufficient balance (debit went negative)
	if senderNewBalance < 0 {
		return model.ErrInsufficientBalance
	}

	// 7. Credit receiver account
	receiverNewBalance, err := s.balanceRepo.Credit(ctx, tx, req.ToAccount, req.Amount)
	if err != nil {
		return fmt.Errorf("crediting receiver: %w", err)
	}

	// 8. Create DEBIT transaction record for sender
	now := time.Now()
	senderTransaction := &model.Transaction{
		TransDate:  now,
		Amount:     req.Amount,
		TransType:  "DEBIT",
		Purpose:    req.Purpose,
		ToAccount:  req.ToAccount,
		AccountNo:  senderAccountNo,
		AccountBal: senderNewBalance,
	}
	if err = s.transactionRepo.Create(ctx, tx, senderTransaction); err != nil {
		return fmt.Errorf("recording sender transaction: %w", err)
	}

	// 9. Create CREDIT transaction record for receiver
	receiverTransaction := &model.Transaction{
		TransDate:  now,
		Amount:     req.Amount,
		TransType:  "CREDIT",
		Purpose:    req.Purpose,
		ToAccount:  senderAccountNo,
		AccountNo:  req.ToAccount,
		AccountBal: receiverNewBalance,
	}
	if err = s.transactionRepo.Create(ctx, tx, receiverTransaction); err != nil {
		return fmt.Errorf("recording receiver transaction: %w", err)
	}

	// 10. Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("committing transfer transaction: %w", err)
	}

	return nil
}
