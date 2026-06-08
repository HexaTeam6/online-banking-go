package service

import (
	"context"
	"time"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/abdurrachmanwahed/online-banking/internal/repository"
)

// RequestService handles money request operations between accounts.
type RequestService interface {
	CreateRequest(ctx context.Context, requesterAccountNo int64, req model.CreateMoneyRequest) error
	GetReceivedRequests(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.MoneyRequest, int64, error)
	MarkAsViewed(ctx context.Context, requestID int64) error
}

// requestService is the concrete implementation of RequestService.
type requestService struct {
	requestRepo repository.RequestRepository
	accountRepo repository.AccountRepository
}

// NewRequestService creates a new RequestService with injected dependencies.
func NewRequestService(requestRepo repository.RequestRepository, accountRepo repository.AccountRepository) RequestService {
	return &requestService{
		requestRepo: requestRepo,
		accountRepo: accountRepo,
	}
}

// CreateRequest validates the money request and persists it with PENDING status.
func (s *requestService) CreateRequest(ctx context.Context, requesterAccountNo int64, req model.CreateMoneyRequest) error {
	// Validate amount limits (500-20000)
	if req.Amount < 500 || req.Amount > 20000 {
		return model.ErrLimitViolation
	}

	// Prevent self-request
	if req.ToAccount == requesterAccountNo {
		return model.ErrSelfTransfer
	}

	// Verify target account exists
	exists, err := s.accountRepo.Exists(ctx, req.ToAccount)
	if err != nil {
		return err
	}
	if !exists {
		return model.ErrAccountNotFound
	}

	// Create the money request with PENDING status
	moneyRequest := &model.MoneyRequest{
		AccountNo:   requesterAccountNo,
		ToAccount:   req.ToAccount,
		Amount:      req.Amount,
		Message:     req.Message,
		HasViewed:   false,
		Status:      "PENDING",
		RequestDate: time.Now(),
	}

	return s.requestRepo.Create(ctx, moneyRequest)
}

// GetReceivedRequests returns a paginated list of requests received by the given account.
func (s *requestService) GetReceivedRequests(ctx context.Context, accountNo int64, pagination model.Pagination) ([]model.MoneyRequest, int64, error) {
	return s.requestRepo.GetReceivedByAccountNo(ctx, accountNo, pagination)
}

// MarkAsViewed sets the hasViewed flag to true for the specified request.
func (s *requestService) MarkAsViewed(ctx context.Context, requestID int64) error {
	return s.requestRepo.MarkAsViewed(ctx, requestID)
}
