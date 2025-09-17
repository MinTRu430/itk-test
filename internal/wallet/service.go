package wallet

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type WalletRepo interface {
	GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error)
	UpdateBalanceTx(ctx context.Context, walletID uuid.UUID, operationType string, amount int64) (int64, error)
}

type WalletService struct {
	repo WalletRepo
}

func NewWalletService(r WalletRepo) *WalletService {
	return &WalletService{repo: r}
}

type OperationRequest struct {
	WalletID      uuid.UUID `json:"walletId"`
	OperationType string    `json:"operationType"`
	Amount        int64     `json:"amount"`
}

func (s *WalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	return s.repo.GetBalance(ctx, walletID)
}

func (s *WalletService) ApplyOperation(ctx context.Context, req OperationRequest) (int64, error) {
	if req.Amount <= 0 {
		return 0, errors.New("amount must be > 0")
	}
	if req.OperationType != "DEPOSIT" && req.OperationType != "WITHDRAW" {
		return 0, errors.New("invalid operation type")
	}
	return s.repo.UpdateBalanceTx(ctx, req.WalletID, req.OperationType, req.Amount)
}
