package wallet

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInsufficientFunds = errors.New("insufficient funds")
var ErrWalletNotFound = errors.New("wallet not found")

type DBRepo struct {
	db *pgxpool.Pool
}

func NewDBRepo(db *pgxpool.Pool) *DBRepo {
	return &DBRepo{db: db}
}

func (r *DBRepo) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	var bal int64
	err := r.db.QueryRow(ctx, "SELECT amount FROM wallets WHERE id=$1", walletID).Scan(&bal)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return 0, err
		}
		return 0, ErrWalletNotFound
	}
	return bal, nil
}

func (r *DBRepo) UpdateBalanceTx(ctx context.Context, walletID uuid.UUID, operationType string, amount int64) (int64, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var currentBalance int64
	err = tx.QueryRow(ctx, "SELECT amount FROM wallets WHERE id=$1 FOR UPDATE", walletID).Scan(&currentBalance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if operationType == "DEPOSIT" {
				err = tx.QueryRow(ctx, `
					INSERT INTO wallets (id, amount) VALUES ($1, $2) RETURNING amount
				`, walletID, amount).Scan(&currentBalance)
				if err != nil {
					return 0, err
				}
				if err = tx.Commit(ctx); err != nil {
					return 0, err
				}
				return currentBalance, nil
			}
			return 0, ErrWalletNotFound
		}
		return 0, err
	}

	var newBalance int64
	switch operationType {
	case "DEPOSIT":
		newBalance = currentBalance + amount
	case "WITHDRAW":
		if currentBalance < amount {
			return 0, ErrInsufficientFunds
		}
		newBalance = currentBalance - amount
	default:
		return 0, errors.New("unknown operation")
	}

	_, err = tx.Exec(ctx, "UPDATE wallets SET amount=$1 WHERE id=$2", newBalance, walletID)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}

	return newBalance, nil
}
