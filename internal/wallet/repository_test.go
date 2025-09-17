package wallet

import (
	"context"
	"os"
	"testing"
	"time"

	"itk/internal/utils"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var testRepo *DBRepo

func TestMain(m *testing.M) {

	if err := godotenv.Load("../../config.env"); err != nil {

		panic(err)
	}
	// Подключаемся к тестовой БД
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := utils.NewPostgresPool(ctx)
	if err != nil {
		panic(err)
	}

	testRepo = NewDBRepo(db)

	code := m.Run()
	db.Close()
	os.Exit(code)
}

func TestDBRepo_Integration(t *testing.T) {
	ctx := context.Background()
	wid := uuid.New()

	// -----------------
	// DEPOSIT
	// -----------------
	newBal, err := testRepo.UpdateBalanceTx(ctx, wid, "DEPOSIT", 1000)
	assert.NoError(t, err)
	assert.Equal(t, int64(1000), newBal)

	// -----------------
	// GET BALANCE
	// -----------------
	bal, err := testRepo.GetBalance(ctx, wid)
	assert.NoError(t, err)
	assert.Equal(t, int64(1000), bal)

	// -----------------
	// WITHDRAW success
	// -----------------
	newBal, err = testRepo.UpdateBalanceTx(ctx, wid, "WITHDRAW", 500)
	assert.NoError(t, err)
	assert.Equal(t, int64(500), newBal)

	// -----------------
	// WITHDRAW insufficient funds
	// -----------------
	_, err = testRepo.UpdateBalanceTx(ctx, wid, "WITHDRAW", 1000)
	assert.ErrorIs(t, err, ErrInsufficientFunds)

	// -----------------
	// GET BALANCE final
	// -----------------
	bal, err = testRepo.GetBalance(ctx, wid)
	assert.NoError(t, err)
	assert.Equal(t, int64(500), bal)
}

func TestDBRepo_ConcurrentDeposits(t *testing.T) {
	ctx := context.Background()
	wid := uuid.New()
	const n = 50
	const amount = 10

	// делаем n параллельных депозитов
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		go func() {
			_, err := testRepo.UpdateBalanceTx(ctx, wid, "DEPOSIT", amount)
			errs <- err
		}()
	}

	// ждём все горутины
	for i := 0; i < n; i++ {
		assert.NoError(t, <-errs)
	}

	// проверяем итоговый баланс
	bal, err := testRepo.GetBalance(ctx, wid)
	assert.NoError(t, err)
	assert.Equal(t, int64(n*amount), bal)
}
