package wallet

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepo нужен для юнит-тестов WalletService
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	args := m.Called(ctx, walletID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepo) UpdateBalanceTx(ctx context.Context, walletID uuid.UUID, operationType string, amount int64) (int64, error) {
	args := m.Called(ctx, walletID, operationType, amount)
	return args.Get(0).(int64), args.Error(1)
}

func TestWalletService_ApplyOperation(t *testing.T) {
	repo := new(MockRepo)
	svc := NewWalletService(repo)
	ctx := context.Background()
	wid := uuid.New()

	t.Run("deposit success", func(t *testing.T) {
		repo.On("UpdateBalanceTx", ctx, wid, "DEPOSIT", int64(100)).Return(int64(100), nil).Once()
		bal, err := svc.ApplyOperation(ctx, OperationRequest{
			WalletID:      wid,
			OperationType: "DEPOSIT",
			Amount:        100,
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(100), bal)
	})

	t.Run("withdraw success", func(t *testing.T) {
		repo.On("UpdateBalanceTx", ctx, wid, "WITHDRAW", int64(50)).Return(int64(50), nil).Once()
		bal, err := svc.ApplyOperation(ctx, OperationRequest{
			WalletID:      wid,
			OperationType: "WITHDRAW",
			Amount:        50,
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(50), bal)
	})

	t.Run("withdraw insufficient funds", func(t *testing.T) {
		repo.On("UpdateBalanceTx", ctx, wid, "WITHDRAW", int64(100)).Return(int64(0), ErrInsufficientFunds).Once()
		_, err := svc.ApplyOperation(ctx, OperationRequest{
			WalletID:      wid,
			OperationType: "WITHDRAW",
			Amount:        100,
		})
		assert.ErrorIs(t, err, ErrInsufficientFunds)
	})

	t.Run("invalid operation type", func(t *testing.T) {
		_, err := svc.ApplyOperation(ctx, OperationRequest{
			WalletID:      wid,
			OperationType: "UNKNOWN",
			Amount:        10,
		})
		assert.Error(t, err)
	})

	t.Run("amount <= 0", func(t *testing.T) {
		_, err := svc.ApplyOperation(ctx, OperationRequest{
			WalletID:      wid,
			OperationType: "DEPOSIT",
			Amount:        0,
		})
		assert.Error(t, err)
	})
}

func TestWalletService_GetBalance(t *testing.T) {
	repo := new(MockRepo)
	svc := NewWalletService(repo)
	ctx := context.Background()
	wid := uuid.New()

	t.Run("balance found", func(t *testing.T) {
		repo.On("GetBalance", ctx, wid).Return(int64(500), nil).Once()
		bal, err := svc.GetBalance(ctx, wid)
		assert.NoError(t, err)
		assert.Equal(t, int64(500), bal)
	})

	t.Run("wallet not found", func(t *testing.T) {
		repo.On("GetBalance", ctx, wid).Return(int64(0), ErrWalletNotFound).Once()
		_, err := svc.GetBalance(ctx, wid)
		assert.ErrorIs(t, err, ErrWalletNotFound)
	})
}
