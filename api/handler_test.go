package api_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gabrielksneiva/go-financial-transactions/api"
	"github.com/gabrielksneiva/go-financial-transactions/domain"
	"github.com/gabrielksneiva/go-financial-transactions/mocks"
	"github.com/gabrielksneiva/go-financial-transactions/services"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestApp() (*fiber.App, *mocks.Producer, *mocks.TransactionRepository, *mocks.BalanceRepository, *mocks.UserRepository) {
	txRepo := new(mocks.TransactionRepository)
	balanceRepo := new(mocks.BalanceRepository)
	userRepo := new(mocks.UserRepository)
	producer := new(mocks.Producer)

	depositService := services.NewDepositService(txRepo, balanceRepo, producer)
	withdrawService := services.NewWithdrawService(txRepo, balanceRepo, producer)
	statementService := services.NewStatementService(txRepo, balanceRepo)
	userService := services.NewUserService(userRepo)

	appStruct := api.NewApp(depositService, withdrawService, statementService, userService)

	return appStruct.Fiber, producer, txRepo, balanceRepo, userRepo
}

func TestStatementHandler_Success(t *testing.T) {
	app, _, txRepoMock, balanceRepoMock, userRepoMock := setupTestApp()

	userID := uint(789)

	// Mock para o GetByID
	userRepoMock.On("GetByID", userID).Return(&domain.User{ID: userID}, nil)

	txRepoMock.On("GetByUser", userID).Return([]domain.Transaction{
		{ID: "tx1", UserID: userID, Amount: 50.0, Type: "deposit"},
	}, nil)

	balanceRepoMock.On("GetBalance", userID).Return(&domain.Balance{
		UserID: userID, Amount: 50.0,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/statement/789", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Verificando as expectativas
	userRepoMock.AssertExpectations(t)
	txRepoMock.AssertExpectations(t)
	balanceRepoMock.AssertExpectations(t)
}

func TestWithdrawHandler_InsufficientFunds(t *testing.T) {
	app, _, _, balanceRepoMock, _ := setupTestApp()

	// Simula saldo insuficiente
	balanceRepoMock.On("GetBalance", uint(456)).Return(&domain.Balance{
		UserID: 456, Amount: 50.0,
	}, nil)

	body := []byte(`{"amount":100.0}`)
	req := httptest.NewRequest(http.MethodPost, "/api/withdraw", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestBalanceHandler_Success(t *testing.T) {
	app, _, _, balanceRepoMock, _ := setupTestApp()

	balanceRepoMock.On("GetBalance", uint(123)).Return(&domain.Balance{
		UserID: 123, Amount: 150.0,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/balance/123", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestDepositHandler_InvalidJSON(t *testing.T) {
	app, _, _, _, _ := setupTestApp()

	body := []byte(`invalid json`)
	req := httptest.NewRequest(http.MethodPost, "/api/deposit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestDepositHandler_InternalError(t *testing.T) {
	app, producerMock, _, _, _ := setupTestApp()

	producerMock.On("SendTransaction", mock.Anything).Return(errors.New("kafka error"))

	body := []byte(`{"amount":50.0}`)
	req := httptest.NewRequest(http.MethodPost, "/api/deposit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestWithdrawHandler_InvalidJSON(t *testing.T) {
	app, _, _, _, _ := setupTestApp()

	body := []byte(`not-a-json`)
	req := httptest.NewRequest(http.MethodPost, "/api/withdraw", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestWithdrawHandler_DBError(t *testing.T) {
	app, _, _, balanceRepo, _ := setupTestApp()

	balanceRepo.On("GetBalance", uint(999)).Return(nil, errors.New("db error"))

	body := []byte(`{"amount":20.0}`)
	req := httptest.NewRequest(http.MethodPost, "/api/withdraw", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestBalanceHandler_NotFound(t *testing.T) {
	app, _, _, balanceRepo, _ := setupTestApp()

	balanceRepo.On("GetBalance", uint(404)).Return(nil, errors.New("not found"))

	req := httptest.NewRequest(http.MethodGet, "/api/balance/404", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestStatementHandler_NotFound(t *testing.T) {
	app, _, _, txRepoMock, _ := setupTestApp()

	userID := uint(789)

	// Simulando um erro ao buscar transações ou balanço
	txRepoMock.On("GetByUser", userID).Return(nil, errors.New("transactions not found"))
	txRepoMock.On("GetBalance", userID).Return(nil, errors.New("user not found"))

	req := httptest.NewRequest(http.MethodGet, "/api/statement/789", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}
