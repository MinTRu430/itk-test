package wallet

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) ApplyOperation(ctx context.Context, req OperationRequest) (int64, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockService) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	args := m.Called(ctx, walletID)
	return args.Get(0).(int64), args.Error(1)
}

func TestWalletHandler(t *testing.T) {
	wid := uuid.New()
	svc := new(MockService)
	handler := NewWalletHandler(svc)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	t.Run("POST /api/v1/wallet - deposit success", func(t *testing.T) {
		svc.On("ApplyOperation", mock.Anything, OperationRequest{
			WalletID:      wid,
			OperationType: "DEPOSIT",
			Amount:        100,
		}).Return(int64(100), nil).Once()

		body := postReq{WalletID: wid.String(), OperationType: "DEPOSIT", Amount: 100}
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewReader(b))
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp balanceResp
		_ = json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, int64(100), resp.Balance)
		assert.Equal(t, wid.String(), resp.WalletID)
	})

	t.Run("POST /api/v1/wallet - invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewReader([]byte("{bad json")))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("POST /api/v1/wallet - invalid wallet ID", func(t *testing.T) {
		body := postReq{WalletID: "invalid-uuid", OperationType: "DEPOSIT", Amount: 100}
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewReader(b))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("POST /api/v1/wallet - insufficient funds", func(t *testing.T) {
		body := postReq{WalletID: wid.String(), OperationType: "WITHDRAW", Amount: 100}
		b, _ := json.Marshal(body)

		svc.On("ApplyOperation", mock.Anything, mock.Anything).Return(int64(0), ErrInsufficientFunds).Once()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewReader(b))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("POST /api/v1/wallet - other service error", func(t *testing.T) {
		body := postReq{WalletID: wid.String(), OperationType: "DEPOSIT", Amount: 100}
		b, _ := json.Marshal(body)

		svc.On("ApplyOperation", mock.Anything, mock.Anything).Return(int64(0), assert.AnError).Once()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", bytes.NewReader(b))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GET /api/v1/wallets/{id} - success", func(t *testing.T) {
		svc.On("GetBalance", mock.Anything, wid).Return(int64(500), nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+wid.String(), nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp balanceResp
		_ = json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, int64(500), resp.Balance)
		assert.Equal(t, wid.String(), resp.WalletID)
	})

	t.Run("GET /api/v1/wallets/ - missing wallet ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GET /api/v1/wallets/{id} - invalid wallet ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/invalid-uuid", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GET /api/v1/wallets/{id} - wallet not found", func(t *testing.T) {
		svc.On("GetBalance", mock.Anything, wid).Return(int64(0), ErrWalletNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+wid.String(), nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("GET /api/v1/wallets/{id} - method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/wallets/"+wid.String(), nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("GET /api/v1/wallets/{id} - wallet not found", func(t *testing.T) {
		wid := uuid.New()
		svc.On("GetBalance", mock.Anything, wid).Return(int64(0), ErrWalletNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+wid.String(), nil)
		w := httptest.NewRecorder()

		handler.GetBalance(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("POST /api/v1/wallet - method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/wallet", nil)
		w := httptest.NewRecorder()

		handler.PostOperation(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("GET /api/v1/wallets/{id} - method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/wallets/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()

		handler.GetBalance(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("GET /api/v1/wallets - wrong prefix -> 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/wrongprefix/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()

		handler.GetBalance(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

}
