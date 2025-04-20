package services_test

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/gabrielksneiva/go-financial-transactions/domain"
	"github.com/gabrielksneiva/go-financial-transactions/mocks"
	"github.com/gabrielksneiva/go-financial-transactions/services"
)

// Setup helpers
func setupDepositService() (*mocks.Producer, *mocks.RateLimiter, *services.DepositService) {
	producer := new(mocks.Producer)
	rateLimiter := new(mocks.RateLimiter)

	service := services.NewDepositService(nil, nil, producer, rateLimiter)
	return producer, rateLimiter, service
}

func setupWithdrawService() (*mocks.TransactionRepository, *mocks.BalanceRepository, *mocks.Producer, *mocks.RateLimiter, *services.WithdrawService) {
	txRepo := new(mocks.TransactionRepository)
	balanceRepo := new(mocks.BalanceRepository)
	producer := new(mocks.Producer)
	rateLimiter := new(mocks.RateLimiter)

	service := services.NewWithdrawService(txRepo, balanceRepo, producer, rateLimiter)
	return txRepo, balanceRepo, producer, rateLimiter, service
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
		producer, rateLimiter, service := setupDepositService()
		userID := uint(123)
		amount := 100.0

		// Configuração do mock para RateLimiter
		rateLimiter.On("CheckTransactionRateLimit", userID).Return(nil)

		// Configuração do mock para Producer
		producer.On("SendTransaction", mock.AnythingOfType("domain.Transaction")).Return(nil)

		err := service.Deposit(userID, amount)

		// Verificações
		assert.NoError(t, err)
		producer.AssertCalled(t, "SendTransaction", mock.AnythingOfType("domain.Transaction"))
		rateLimiter.AssertCalled(t, "CheckTransactionRateLimit", userID)
	})

	t.Run("Deposit_RateLimiterError", func(t *testing.T) {
		producer, rateLimiter, service := setupDepositService()
		userID := uint(1)
		amount := 100.0

		rateLimiter.On("CheckTransactionRateLimit", userID).Return(errors.New("rate limit exceeded"))

		err := service.Deposit(userID, amount)
		assert.EqualError(t, err, "rate limit exceeded")
		producer.AssertNotCalled(t, "SendTransaction", mock.Anything)
	})

	t.Run("Deposit_RateLimitExceeded", func(t *testing.T) {
		producer, rateLimiter, service := setupDepositService()
		userID := uint(123)
		amount := 100.0

		rateLimiter.On("CheckTransactionRateLimit", userID).
			Return(errors.New("rate limit exceeded"))

		err := service.Deposit(userID, amount)
		assert.EqualError(t, err, "rate limit exceeded")
		producer.AssertNotCalled(t, "SendTransaction", mock.Anything)
	})

	t.Run("Deposit_ProducerError", func(t *testing.T) {
		producer, rateLimiter, service := setupDepositService()
		userID := uint(2)
		amount := 150.0

		rateLimiter.On("CheckTransactionRateLimit", userID).Return(nil)
		producer.On("SendTransaction", mock.AnythingOfType("domain.Transaction")).Return(errors.New("kafka down"))

		err := service.Deposit(userID, amount)
		assert.EqualError(t, err, "kafka down")
		producer.AssertCalled(t, "SendTransaction", mock.Anything)
	})

	t.Run("Deposit_ProducerFails", func(t *testing.T) {
		producer, rateLimiter, service := setupDepositService()
		userID := uint(123)
		amount := 100.0

		rateLimiter.On("CheckTransactionRateLimit", userID).Return(nil)
		producer.On("SendTransaction", mock.AnythingOfType("domain.Transaction")).Return(errors.New("kafka down"))

		err := service.Deposit(userID, amount)
		assert.EqualError(t, err, "kafka down")
	})

	t.Run("Deposit_InvalidAmount", func(t *testing.T) {
		producer, rateLimiter, service := setupDepositService()
		userID := uint(10)

		err := service.Deposit(userID, 0)
		assert.EqualError(t, err, "amount must be greater than zero")

		producer.AssertNotCalled(t, "SendTransaction", mock.Anything)
		rateLimiter.AssertNotCalled(t, "CheckTransactionRateLimit", userID)
	})

}

// ----------------- Withdraw Tests -----------------
func TestWithdrawService(t *testing.T) {
	userID := uint(123)
	amount := 50.0

	t.Run("Withdraw_Success", func(t *testing.T) {
		_, balanceRepo, producer, rateLimiter, service := setupWithdrawService()

		balanceRepo.On("GetBalance", userID).
			Return(&domain.Balance{UserID: userID, Amount: 100.0}, nil)

		producer.On("SendTransaction", mock.AnythingOfType("domain.Transaction")).
			Return(nil)

		// Configuração do mock para RateLimiter
		rateLimiter.On("CheckTransactionRateLimit", userID).Return(nil)

		err := service.Withdraw(userID, amount)
		assert.NoError(t, err)

		balanceRepo.AssertExpectations(t)
		producer.AssertExpectations(t)
		rateLimiter.AssertCalled(t, "CheckTransactionRateLimit", userID)
	})

	t.Run("Withdraw_InsufficientFunds", func(t *testing.T) {
		_, balanceRepo, producer, rateLimiter, service := setupWithdrawService()

		balanceRepo.On("GetBalance", userID).Return(&domain.Balance{
			UserID: userID,
			Amount: 20.0,
		}, nil)

		// Configuração do mock para RateLimiter
		rateLimiter.On("CheckTransactionRateLimit", userID).Return(nil)

		err := service.Withdraw(userID, amount)
		assert.EqualError(t, err, "insufficient funds")

		producer.AssertNotCalled(t, "SendTransaction", mock.Anything)
		rateLimiter.AssertCalled(t, "CheckTransactionRateLimit", userID)
	})

	t.Run("Withdraw_BalanceError", func(t *testing.T) {
		_, balanceRepo, producer, rateLimiter, service := setupWithdrawService()

		balanceRepo.On("GetBalance", userID).Return(nil, errors.New("db error"))

		// Configuração do mock para RateLimiter
		rateLimiter.On("CheckTransactionRateLimit", userID).Return(nil)

		err := service.Withdraw(userID, amount)
		assert.EqualError(t, err, "db error")

		producer.AssertNotCalled(t, "SendTransaction", mock.Anything)
		rateLimiter.AssertCalled(t, "CheckTransactionRateLimit", userID)
	})

	t.Run("Withdraw_RateLimiterError", func(t *testing.T) {
		_, balanceRepo, producer, rateLimiter, service := setupWithdrawService()

		rateLimiter.On("CheckTransactionRateLimit", userID).Return(errors.New("rate limit exceeded"))

		err := service.Withdraw(userID, amount)
		assert.EqualError(t, err, "rate limit exceeded")

		balanceRepo.AssertNotCalled(t, "GetBalance", mock.Anything)
		producer.AssertNotCalled(t, "SendTransaction", mock.Anything)
	})

	t.Run("Withdraw_ProducerError", func(t *testing.T) {
		_, balanceRepo, producer, rateLimiter, service := setupWithdrawService()

		rateLimiter.On("CheckTransactionRateLimit", userID).Return(nil)
		balanceRepo.On("GetBalance", userID).Return(&domain.Balance{UserID: userID, Amount: 100.0}, nil)
		producer.On("SendTransaction", mock.Anything).Return(errors.New("kafka fail"))

		err := service.Withdraw(userID, amount)
		assert.EqualError(t, err, "kafka fail")

		producer.AssertCalled(t, "SendTransaction", mock.Anything)
	})

}

// ----------------- Statement Tests -----------------

func TestStatementService(t *testing.T) {
	t.Run("GetBalance_Success", func(t *testing.T) {
		_, balanceRepo, service := setupStatementService()
		userID := uint(456)

		balanceRepo.On("GetBalance", userID).Return(&domain.Balance{UserID: userID, Amount: 200.0}, nil)

		amount, err := service.GetBalance(userID)

		assert.NoError(t, err)
		assert.Equal(t, 200.0, amount)
	})

	t.Run("GetTransactions_Success", func(t *testing.T) {
		txRepo, _, service := setupStatementService()
		userID := uint(789)
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
		userID := uint(999)
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

	t.Run("GetBalance_Error", func(t *testing.T) {
		_, balanceRepo, service := setupStatementService()
		userID := uint(555)

		balanceRepo.On("GetBalance", userID).Return(nil, errors.New("db error"))

		_, err := service.GetBalance(userID)
		assert.EqualError(t, err, "db error")
	})

	t.Run("GetTransactions_Error", func(t *testing.T) {
		txRepo, _, service := setupStatementService()
		userID := uint(666)

		txRepo.On("GetByUser", userID).Return(nil, errors.New("db fail"))

		_, err := service.GetTransactions(userID)
		assert.EqualError(t, err, "db fail")
	})

	t.Run("GetStatement_BalanceError", func(t *testing.T) {
		_, balanceRepo, service := setupStatementService()
		userID := uint(777)

		balanceRepo.On("GetBalance", userID).Return(nil, errors.New("no balance"))

		_, err := service.GetStatement(userID)
		assert.EqualError(t, err, "no balance")
	})

	t.Run("GetStatement_TransactionsError", func(t *testing.T) {
		txRepo, balanceRepo, service := setupStatementService()
		userID := uint(888)

		balanceRepo.On("GetBalance", userID).Return(&domain.Balance{UserID: userID, Amount: 100}, nil)
		txRepo.On("GetByUser", userID).Return(nil, errors.New("tx error"))

		_, err := service.GetStatement(userID)
		assert.EqualError(t, err, "tx error")
	})

	t.Run("GetStatement_TxError", func(t *testing.T) {
		txRepo, balanceRepo, service := setupStatementService()
		userID := uint(1)

		balanceRepo.On("GetBalance", userID).Return(&domain.Balance{UserID: userID, Amount: 0}, nil)
		txRepo.On("GetByUser", userID).Return(nil, errors.New("tx fail"))

		_, err := service.GetStatement(userID)
		assert.EqualError(t, err, "tx fail")
	})

}

func TestUserService_CreateUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		repo := new(mocks.UserRepository) // novo mock
		service := services.NewUserService(repo)
		user := &domain.User{Email: "test@example.com", Password: "password123"}

		repo.On("Create", mock.Anything).Return(nil)

		err := service.CreateUser(user)
		assert.NoError(t, err)
	})

	t.Run("DuplicateEmail", func(t *testing.T) {
		repo := new(mocks.UserRepository) // novo mock
		service := services.NewUserService(repo)

		pgErr := &pgconn.PgError{Code: "23505"}
		repo.On("Create", mock.Anything).Return(pgErr)

		user := &domain.User{Email: "test@example.com", Password: "password123"}

		err := service.CreateUser(user)
		assert.EqualError(t, err, "e-mail já cadastrado")
	})

	t.Run("OtherCreateError", func(t *testing.T) {
		repo := new(mocks.UserRepository) // novo mock
		service := services.NewUserService(repo)

		repo.On("Create", mock.Anything).Return(errors.New("db error"))

		user := &domain.User{Email: "test@example.com", Password: "password123"}

		err := service.CreateUser(user)
		assert.EqualError(t, err, "db error")
	})
}

func TestUserService_Authenticate(t *testing.T) {
	repo := new(mocks.UserRepository)
	service := services.NewUserService(repo)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("secret"), 14)
	user := &domain.User{Email: "test", Password: string(hashed)}

	t.Run("Success", func(t *testing.T) {
		repo.On("GetByEmail", "test").Return(user, nil)
		u, err := service.Authenticate("test", "secret")
		assert.NoError(t, err)
		assert.Equal(t, user.Email, u.Email)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		repo.On("GetByEmail", "test2").Return(nil, errors.New("not found"))
		_, err := service.Authenticate("test2", "secret")
		assert.EqualError(t, err, "usuário não encontrado")
	})

	t.Run("WrongPassword", func(t *testing.T) {
		repo.On("GetByEmail", "test3").Return(&domain.User{Password: string(hashed)}, nil)
		_, err := service.Authenticate("test3", "wrong")
		assert.EqualError(t, err, "senha inválida")
	})
}
