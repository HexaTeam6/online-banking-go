package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abdurrachmanwahed/online-banking/internal/handler"
	"github.com/abdurrachmanwahed/online-banking/internal/middleware"
	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/abdurrachmanwahed/online-banking/internal/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTransferService is a test double for service.TransferService.
type mockTransferService struct {
	quickTransferFn func(ctx context.Context, senderAccountNo int64, req model.TransferRequest) error
}

func (m *mockTransferService) QuickTransfer(ctx context.Context, senderAccountNo int64, req model.TransferRequest) error {
	return m.quickTransferFn(ctx, senderAccountNo, req)
}

const testJWTSecret = "test-secret-key-for-handler-tests"

// authenticatedRequest creates a request with a valid JWT token for the given account number.
func authenticatedRequest(t *testing.T, method, url string, body []byte, accountNo int64) *http.Request {
	t.Helper()
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, url, bytes.NewBuffer(body))
	} else {
		req = httptest.NewRequest(method, url, nil)
	}
	req.Header.Set("Content-Type", "application/json")

	tm := security.NewJWTManager(testJWTSecret)
	token, err := tm.Generate(security.TokenClaims{
		AccountNo: accountNo,
		Role:      "customer",
	})
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

// serveWithAuth wraps the handler function with the Authenticate middleware.
func serveWithAuth(handlerFn http.HandlerFunc) http.Handler {
	tm := security.NewJWTManager(testJWTSecret)
	return middleware.Authenticate(tm)(http.HandlerFunc(handlerFn))
}

func TestQuickTransfer_Success(t *testing.T) {
	svc := &mockTransferService{
		quickTransferFn: func(_ context.Context, senderAccountNo int64, req model.TransferRequest) error {
			assert.Equal(t, int64(123456789), senderAccountNo)
			assert.Equal(t, int64(987654321), req.ToAccount)
			assert.Equal(t, int64(1000), req.Amount)
			assert.Equal(t, "payment", req.Purpose)
			return nil
		},
	}
	h := handler.NewTransferHandler(svc)

	body := []byte(`{"to_account":987654321,"amount":1000,"purpose":"payment"}`)
	req := authenticatedRequest(t, http.MethodPost, "/api/v1/transfers", body, 123456789)

	rr := httptest.NewRecorder()
	serveWithAuth(h.QuickTransfer).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "transfer completed successfully", data["message"])
}

func TestQuickTransfer_InvalidJSON(t *testing.T) {
	svc := &mockTransferService{}
	h := handler.NewTransferHandler(svc)

	body := []byte(`{invalid`)
	req := authenticatedRequest(t, http.MethodPost, "/api/v1/transfers", body, 123456789)

	rr := httptest.NewRecorder()
	serveWithAuth(h.QuickTransfer).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestQuickTransfer_ValidationError(t *testing.T) {
	svc := &mockTransferService{}
	h := handler.NewTransferHandler(svc)

	// Amount below minimum (500)
	body := []byte(`{"to_account":987654321,"amount":100,"purpose":"test"}`)
	req := authenticatedRequest(t, http.MethodPost, "/api/v1/transfers", body, 123456789)

	rr := httptest.NewRecorder()
	serveWithAuth(h.QuickTransfer).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestQuickTransfer_NoAuth(t *testing.T) {
	svc := &mockTransferService{}
	h := handler.NewTransferHandler(svc)

	body := []byte(`{"to_account":987654321,"amount":1000,"purpose":"payment"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transfers", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header

	rr := httptest.NewRecorder()
	serveWithAuth(h.QuickTransfer).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestQuickTransfer_LimitViolation(t *testing.T) {
	svc := &mockTransferService{
		quickTransferFn: func(_ context.Context, _ int64, _ model.TransferRequest) error {
			return model.ErrLimitViolation
		},
	}
	h := handler.NewTransferHandler(svc)

	body := []byte(`{"to_account":987654321,"amount":1000,"purpose":"test"}`)
	req := authenticatedRequest(t, http.MethodPost, "/api/v1/transfers", body, 123456789)

	rr := httptest.NewRecorder()
	serveWithAuth(h.QuickTransfer).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestQuickTransfer_SelfTransfer(t *testing.T) {
	svc := &mockTransferService{
		quickTransferFn: func(_ context.Context, _ int64, _ model.TransferRequest) error {
			return model.ErrSelfTransfer
		},
	}
	h := handler.NewTransferHandler(svc)

	body := []byte(`{"to_account":123456789,"amount":1000,"purpose":"test"}`)
	req := authenticatedRequest(t, http.MethodPost, "/api/v1/transfers", body, 123456789)

	rr := httptest.NewRecorder()
	serveWithAuth(h.QuickTransfer).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestQuickTransfer_AccountNotFound(t *testing.T) {
	svc := &mockTransferService{
		quickTransferFn: func(_ context.Context, _ int64, _ model.TransferRequest) error {
			return model.ErrAccountNotFound
		},
	}
	h := handler.NewTransferHandler(svc)

	body := []byte(`{"to_account":999999999,"amount":1000,"purpose":"test"}`)
	req := authenticatedRequest(t, http.MethodPost, "/api/v1/transfers", body, 123456789)

	rr := httptest.NewRecorder()
	serveWithAuth(h.QuickTransfer).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestQuickTransfer_InsufficientBalance(t *testing.T) {
	svc := &mockTransferService{
		quickTransferFn: func(_ context.Context, _ int64, _ model.TransferRequest) error {
			return model.ErrInsufficientBalance
		},
	}
	h := handler.NewTransferHandler(svc)

	body := []byte(`{"to_account":987654321,"amount":1000,"purpose":"test"}`)
	req := authenticatedRequest(t, http.MethodPost, "/api/v1/transfers", body, 123456789)

	rr := httptest.NewRecorder()
	serveWithAuth(h.QuickTransfer).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
