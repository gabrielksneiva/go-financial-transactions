package workers_test

import (
	"context"
	"testing"
	"time"

	"github.com/gabrielksneiva/go-financial-transactions/domain"
	"github.com/gabrielksneiva/go-financial-transactions/mocks"
	"github.com/gabrielksneiva/go-financial-transactions/workers"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	cleanup := func() {
		sqlDB.Close()
	}

	return db, mock, cleanup
}

func TestWorker_ProcessTransaction_Success(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	tx := domain.Transaction{
		ID:        "tx-1",
		UserID:    1,
		Amount:    100.0,
		Type:      "deposit",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FROM "balances"`).
		WithArgs(tx.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "amount"}).
			AddRow(tx.UserID, 0.0))

	mock.ExpectExec(`UPDATE "balances"`).
		WithArgs(100.0, tx.UserID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`INSERT INTO "transactions"`).
		WithArgs(tx.ID, tx.UserID, tx.Amount, sqlmock.AnyArg(), tx.Type, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	ch := make(chan domain.Transaction, 1)
	ch <- tx

	blockchainMock := new(mocks.BlockchainClient)
	repoMock := new(mocks.TransactionRepository)

	ctx, cancel := context.WithCancel(context.Background())
	go workers.StartWorkers(ctx, ch, 1, db, blockchainMock, repoMock)
	time.Sleep(200 * time.Millisecond)
	cancel()
}

func TestWorker_InsufficientFunds(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	blockchainMock := new(mocks.BlockchainClient)
	repoMock := new(mocks.TransactionRepository)

	tx := domain.Transaction{
		ID:        "tx-2",
		UserID:    2,
		Amount:    -100.0,
		Type:      "withdraw",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FROM "balances"`).
		WithArgs(tx.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "amount"}).
			AddRow(tx.UserID, 50.0))
	mock.ExpectRollback()

	ch := make(chan domain.Transaction, 1)
	ch <- tx

	ctx, cancel := context.WithCancel(context.Background())
	go workers.StartWorkers(ctx, ch, 1, db, blockchainMock, repoMock)
	time.Sleep(200 * time.Millisecond)
	cancel()
}

func TestWorker_ErrorOnInsert(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	blockchainMock := new(mocks.BlockchainClient)
	repoMock := new(mocks.TransactionRepository)

	tx := domain.Transaction{
		ID:        "tx-3",
		UserID:    3,
		Amount:    50.0,
		Type:      "deposit",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FROM "balances"`).
		WithArgs(tx.UserID, tx.UserID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "amount"}).
			AddRow(tx.UserID, 100.0))

	mock.ExpectExec(`UPDATE "balances"`).
		WithArgs(150.0, tx.UserID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`INSERT INTO "transactions"`).
		WithArgs(tx.ID, tx.UserID, tx.Amount, sqlmock.AnyArg(), tx.Type, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(assert.AnError)

	mock.ExpectRollback()

	ch := make(chan domain.Transaction, 1)
	ch <- tx

	ctx, cancel := context.WithCancel(context.Background())
	go workers.StartWorkers(ctx, ch, 1, db, blockchainMock, repoMock)
	time.Sleep(200 * time.Millisecond)
	cancel()
}
