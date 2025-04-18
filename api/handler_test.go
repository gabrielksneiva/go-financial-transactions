// api/handler_test.go
package api_test

import (
	"bytes"
	"errors"
	"go-financial-transactions/api"
	"go-financial-transactions/domain"
	"go-financial-transactions/mocks"
	"go-financial-transactions/services"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestApp() (*fiber.App, *mocks.Producer, *mocks.TransactionRepository, *mocks.BalanceRepository) {
	txRepo := new(mocks.TransactionRepository)
	balanceRepo := new(mocks.BalanceRepository)
	producer := new(mocks.Producer)

	depositService := services.NewDepositService(txRepo, balanceRepo, producer)
	withdrawService := services.NewWithdrawService(txRepo, balanceRepo, producer)
	statementService := services.NewStatementService(txRepo, balanceRepo)

	appStruct := api.NewApp(depositService, withdrawService, statementService)

	return appStruct.Fiber, producer, txRepo, balanceRepo
}

func TestDepositHandler_Success(t *testing.T) {
	app, producerMock, _, _ := setupTestApp()

	// Espera que o serviço deposite sem erro
	producerMock.
		On("SendTransaction", mock.Anything).
		Return(nil)

	body := []byte(`{"user_id":"user-123","amount":100.0}`)
	req := httptest.NewRequest(http.MethodPost, "/deposit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusAccepted, resp.StatusCode)

	producerMock.AssertExpectations(t)
}

func TestWithdrawHandler_InsufficientFunds(t *testing.T) {
	app, _, _, balanceRepoMock := setupTestApp()

	// Simula saldo insuficiente
	balanceRepoMock.On("GetBalance", "user-456").Return(&domain.Balance{
		UserID: "user-456", Amount: 50.0,
	}, nil)

	body := []byte(`{"user_id":"user-456","amount":100.0}`)
	req := httptest.NewRequest(http.MethodPost, "/withdraw", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestBalanceHandler_Success(t *testing.T) {
	app, _, _, balanceRepoMock := setupTestApp()

	balanceRepoMock.On("GetBalance", "user-123").Return(&domain.Balance{
		UserID: "user-123", Amount: 150.0,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/balance/user-123", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestStatementHandler_Success(t *testing.T) {
	app, _, txRepoMock, balanceRepoMock := setupTestApp()

	userID := "user-789"

	txRepoMock.On("GetByUser", userID).Return([]domain.Transaction{
		{ID: "tx1", UserID: userID, Amount: 50.0, Type: "deposit"},
	}, nil)

	balanceRepoMock.On("GetBalance", userID).Return(&domain.Balance{
		UserID: userID, Amount: 50.0,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/statement/"+userID, nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestDepositHandler_InvalidJSON(t *testing.T) {
	app, _, _, _ := setupTestApp()

	body := []byte(`invalid json`)
	req := httptest.NewRequest(http.MethodPost, "/deposit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestDepositHandler_InternalError(t *testing.T) {
	app, producerMock, _, _ := setupTestApp()

	producerMock.On("SendTransaction", mock.Anything).Return(errors.New("kafka error"))

	body := []byte(`{"user_id":"user-err","amount":50.0}`)
	req := httptest.NewRequest(http.MethodPost, "/deposit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestWithdrawHandler_InvalidJSON(t *testing.T) {
	app, _, _, _ := setupTestApp()

	body := []byte(`not-a-json`)
	req := httptest.NewRequest(http.MethodPost, "/withdraw", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestWithdrawHandler_DBError(t *testing.T) {
	app, _, _, balanceRepo := setupTestApp()

	balanceRepo.On("GetBalance", "user-db-error").Return(nil, errors.New("db error"))

	body := []byte(`{"user_id":"user-db-error","amount":20.0}`)
	req := httptest.NewRequest(http.MethodPost, "/withdraw", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestBalanceHandler_NotFound(t *testing.T) {
	app, _, _, balanceRepo := setupTestApp()

	balanceRepo.On("GetBalance", "not-found").Return(nil, errors.New("not found"))

	req := httptest.NewRequest(http.MethodGet, "/balance/not-found", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestStatementHandler_NotFound(t *testing.T) {
	app, _, _, balanceRepo := setupTestApp()

	balanceRepo.On("GetBalance", "missing-user").Return(nil, errors.New("user not found"))

	req := httptest.NewRequest(http.MethodGet, "/statement/missing-user", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}
