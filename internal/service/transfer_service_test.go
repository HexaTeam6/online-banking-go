package service

import (
	"context"
	"errors"
	"testing"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestQuickTransfer_AmountBelowMinimum(t *testing.T) {
	svc := NewTransferService(&mockAccountRepo{}, &mockBalanceRepo{}, &mockTransactionRepo{}, nil)

	err := svc.QuickTransfer(context.Background(), 123456789, model.TransferRequest{
		ToAccount: 987654321,
		Amount:    499,
		Purpose:   "test",
	})

	assert.ErrorIs(t, err, model.ErrLimitViolation)
}

func TestQuickTransfer_AmountAboveMaximum(t *testing.T) {
	svc := NewTransferService(&mockAccountRepo{}, &mockBalanceRepo{}, &mockTransactionRepo{}, nil)

	err := svc.QuickTransfer(context.Background(), 123456789, model.TransferRequest{
		ToAccount: 987654321,
		Amount:    20001,
		Purpose:   "test",
	})

	assert.ErrorIs(t, err, model.ErrLimitViolation)
}

func TestQuickTransfer_SelfTransfer(t *testing.T) {
	svc := NewTransferService(&mockAccountRepo{}, &mockBalanceRepo{}, &mockTransactionRepo{}, nil)

	err := svc.QuickTransfer(context.Background(), 123456789, model.TransferRequest{
		ToAccount: 123456789,
		Amount:    1000,
		Purpose:   "test",
	})

	assert.ErrorIs(t, err, model.ErrSelfTransfer)
}

func TestQuickTransfer_DestinationNotFound(t *testing.T) {
	accountRepo := &mockAccountRepo{
		existsFn: func(_ context.Context, _ int64) (bool, error) {
			return false, nil
		},
	}
	svc := NewTransferService(accountRepo, &mockBalanceRepo{}, &mockTransactionRepo{}, nil)

	err := svc.QuickTransfer(context.Background(), 123456789, model.TransferRequest{
		ToAccount: 987654321,
		Amount:    1000,
		Purpose:   "test",
	})

	assert.ErrorIs(t, err, model.ErrAccountNotFound)
}

func TestQuickTransfer_AccountRepoError(t *testing.T) {
	accountRepo := &mockAccountRepo{
		existsFn: func(_ context.Context, _ int64) (bool, error) {
			return false, errors.New("db connection error")
		},
	}
	svc := NewTransferService(accountRepo, &mockBalanceRepo{}, &mockTransactionRepo{}, nil)

	err := svc.QuickTransfer(context.Background(), 123456789, model.TransferRequest{
		ToAccount: 987654321,
		Amount:    1000,
		Purpose:   "test",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checking destination account")
}

func TestQuickTransfer_AmountAtBoundaries(t *testing.T) {
	accountRepo := &mockAccountRepo{
		existsFn: func(_ context.Context, _ int64) (bool, error) {
			return false, nil
		},
	}
	svc := NewTransferService(accountRepo, &mockBalanceRepo{}, &mockTransactionRepo{}, nil)

	// Exactly 500 - should pass limit check but fail on non-existent account
	err := svc.QuickTransfer(context.Background(), 123456789, model.TransferRequest{
		ToAccount: 987654321,
		Amount:    500,
		Purpose:   "test",
	})
	assert.ErrorIs(t, err, model.ErrAccountNotFound)

	// Exactly 20000 - should pass limit check but fail on non-existent account
	err = svc.QuickTransfer(context.Background(), 123456789, model.TransferRequest{
		ToAccount: 987654321,
		Amount:    20000,
		Purpose:   "test",
	})
	assert.ErrorIs(t, err, model.ErrAccountNotFound)
}
