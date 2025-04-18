package services_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go-financial-transactions/domain"
	"go-financial-transactions/mocks"
	"go-financial-transactions/services"
)

// Setup helpers
func setupDepositService() (*mocks.Producer, *services.DepositService) {
	producer := new(mocks.Producer)
	service := services.NewDepositService(nil, nil, producer)
	return producer, service
}

func setupWithdrawService() (*mocks.TransactionRepository, *mocks.BalanceRepository, *mocks.Producer, *services.WithdrawService) {
	txRepo := new(mocks.TransactionRepository)
	balanceRepo := new(mocks.BalanceRepository)
	producer := new(mocks.Producer)

	service := services.NewWithdrawService(txRepo, balanceRepo, producer)
	return txRepo, balanceRepo, producer, service
}

func setupStatementService() (*mocks.TransactionRepository, *mocks.BalanceRepository, *services.StatementService) {
	txRepo := new(mocks.TransactionRepository)
	balanceRepo := new(mocks.BalanceRepository)
	service := services.NewStatementService(txRepo, balanceRepo)
	return txRepo, balanceRepo, service
}

// ----------------- Deposit Tests -----------------

func TestDepositService(t *testing.T) {
	t.Run("Deposit_Success", func(t *testing.T) {
		producer, service := setupDepositService()
		userID := "user-123"
		amount := 100.0

		producer.On("SendTransaction", mock.AnythingOfType("domain.Transaction")).Return(nil)

		err := service.Deposit(userID, amount)

		assert.NoError(t, err)
		producer.AssertCalled(t, "SendTransaction", mock.AnythingOfType("domain.Transaction"))
	})
}

// ----------------- Withdraw Tests -----------------
func TestWithdrawService(t *testing.T) {
	userID := "user123"
	amount := 50.0

	t.Run("Withdraw_Success", func(t *testing.T) {
		_, balanceRepo, producer, service := setupWithdrawService()

		// Expect balance to be enough
		balanceRepo.On("GetBalance", userID).Return(&domain.Balance{
			UserID: userID,
			Amount: 100.0,
		}, nil)

		// Expect SendTransaction to be called with any transaction
		producer.On("SendTransaction", mock.AnythingOfType("domain.Transaction")).Return(nil)

		err := service.Withdraw(userID, amount)
		assert.NoError(t, err)

		balanceRepo.AssertExpectations(t)
		producer.AssertExpectations(t)
	})

	t.Run("Withdraw_InsufficientFunds", func(t *testing.T) {
		_, balanceRepo, producer, service := setupWithdrawService()

		balanceRepo.On("GetBalance", userID).Return(&domain.Balance{
			UserID: userID,
			Amount: 20.0,
		}, nil)

		err := service.Withdraw(userID, amount)
		assert.EqualError(t, err, "insufficient funds")

		producer.AssertNotCalled(t, "SendTransaction", mock.Anything)
	})

	t.Run("Withdraw_BalanceError", func(t *testing.T) {
		_, balanceRepo, producer, service := setupWithdrawService()

		balanceRepo.On("GetBalance", userID).Return(nil, errors.New("db error"))

		err := service.Withdraw(userID, amount)
		assert.EqualError(t, err, "db error")

		producer.AssertNotCalled(t, "SendTransaction", mock.Anything)
	})
}

// ----------------- Statement Tests -----------------

func TestStatementService(t *testing.T) {
	t.Run("GetBalance_Success", func(t *testing.T) {
		_, balanceRepo, service := setupStatementService()
		userID := "user-456"

		balanceRepo.On("GetBalance", userID).Return(&domain.Balance{UserID: userID, Amount: 200.0}, nil)

		amount, err := service.GetBalance(userID)

		assert.NoError(t, err)
		assert.Equal(t, 200.0, amount)
	})

	t.Run("GetTransactions_Success", func(t *testing.T) {
		txRepo, _, service := setupStatementService()
		userID := "user-789"
		mockTxs := []domain.Transaction{
			{ID: "tx1", UserID: userID, Amount: 100.0},
			{ID: "tx2", UserID: userID, Amount: -50.0},
		}

		txRepo.On("GetByUser", userID).Return(mockTxs, nil)

		txs, err := service.GetTransactions(userID)

		assert.NoError(t, err)
		assert.Equal(t, mockTxs, txs)
	})

	t.Run("GetStatement_Success", func(t *testing.T) {
		txRepo, balanceRepo, service := setupStatementService()
		userID := "user-999"
		mockTxs := []domain.Transaction{
			{ID: "tx1", UserID: userID, Amount: 150.0},
		}
		mockBalance := &domain.Balance{UserID: userID, Amount: 150.0}

		balanceRepo.On("GetBalance", userID).Return(mockBalance, nil)
		txRepo.On("GetByUser", userID).Return(mockTxs, nil)

		statement, err := service.GetStatement(userID)

		assert.NoError(t, err)
		assert.Equal(t, 150.0, statement.Balance)
		assert.Equal(t, mockTxs, statement.Transactions)
	})
}
