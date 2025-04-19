package api_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gabrielksneiva/go-financial-transactions/api"
	"github.com/gabrielksneiva/go-financial-transactions/domain"
	"github.com/gabrielksneiva/go-financial-transactions/mocks"
	"github.com/gabrielksneiva/go-financial-transactions/services"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestApp() (*fiber.App, *mocks.Producer, *mocks.TransactionRepository, *mocks.BalanceRepository, *mocks.UserRepository, *mocks.RateLimiter) {
	txRepo := new(mocks.TransactionRepository)
	balanceRepo := new(mocks.BalanceRepository)
	userRepo := new(mocks.UserRepository)
	rateLimiter := new(mocks.RateLimiter)
	producer := new(mocks.Producer)

	depositService := services.NewDepositService(txRepo, balanceRepo, producer, rateLimiter)
	withdrawService := services.NewWithdrawService(txRepo, balanceRepo, producer, rateLimiter)
	statementService := services.NewStatementService(txRepo, balanceRepo)
	userService := services.NewUserService(userRepo)

	appStruct := api.NewApp(depositService, withdrawService, statementService, userService)

	return appStruct.Fiber, producer, txRepo, balanceRepo, userRepo, rateLimiter
}

func generateTestJWT(userID uint) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 1).Unix(),
	})

	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	return tokenString
}

func TestStatementHandler_Success(t *testing.T) {
	app, _, txRepoMock, balanceRepoMock, userRepoMock, _ := setupTestApp()

	userID := uint(789)

	userRepoMock.On("GetByID", userID).Return(&domain.User{ID: userID}, nil)
	txRepoMock.On("GetByUser", userID).Return([]domain.Transaction{
		{ID: "tx1", UserID: userID, Amount: 50.0, Type: "deposit"},
	}, nil)
	balanceRepoMock.On("GetBalance", userID).Return(&domain.Balance{
		UserID: userID, Amount: 50.0,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/statement/789", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestJWT(userID))

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	userRepoMock.AssertExpectations(t)
	txRepoMock.AssertExpectations(t)
	balanceRepoMock.AssertExpectations(t)
}

func TestWithdrawHandler_InsufficientFunds(t *testing.T) {
	app, _, _, balanceRepoMock, _, rateLimiterMock := setupTestApp()

	userID := uint(456)

	balanceRepoMock.On("GetBalance", userID).Return(&domain.Balance{
		UserID: userID, Amount: 50.0,
	}, nil)

	rateLimiterMock.On("CheckTransactionRateLimit", mock.AnythingOfType("uint")).Return(nil)

	body := []byte(`{"amount":100.0}`)
	req := httptest.NewRequest(http.MethodPost, "/api/withdraw", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+generateTestJWT(userID))

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestBalanceHandler_Success(t *testing.T) {
	app, _, _, balanceRepoMock, _, rateLimiterMock := setupTestApp()

	userID := uint(123)

	balanceRepoMock.On("GetBalance", userID).Return(&domain.Balance{
		UserID: userID, Amount: 150.0,
	}, nil)

	rateLimiterMock.On("CheckTransactionRateLimit", mock.AnythingOfType("uint")).Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/balance/123", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestJWT(userID))

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestDepositHandler_InvalidJSON(t *testing.T) {
	app, _, _, _, _, _ := setupTestApp()

	userID := uint(111)

	body := []byte(`invalid json`)
	req := httptest.NewRequest(http.MethodPost, "/api/deposit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+generateTestJWT(userID))

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestDepositHandler_InternalError(t *testing.T) {
	app, producerMock, _, _, _, rateLimiterMock := setupTestApp()

	userID := uint(222)

	producerMock.On("SendTransaction", mock.Anything).Return(errors.New("kafka error"))
	rateLimiterMock.On("CheckTransactionRateLimit", mock.AnythingOfType("uint")).Return(nil)

	body := []byte(`{"amount":50.0}`)
	req := httptest.NewRequest(http.MethodPost, "/api/deposit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+generateTestJWT(userID))

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestWithdrawHandler_InvalidJSON(t *testing.T) {
	app, _, _, _, _, _ := setupTestApp()

	userID := uint(333)

	body := []byte(`not-a-json`)
	req := httptest.NewRequest(http.MethodPost, "/api/withdraw", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+generateTestJWT(userID))

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestWithdrawHandler_DBError(t *testing.T) {
	app, _, _, balanceRepo, _, rateLimiterMock := setupTestApp()

	userID := uint(999)

	balanceRepo.On("GetBalance", userID).Return(nil, errors.New("db error"))
	rateLimiterMock.On("CheckTransactionRateLimit", mock.AnythingOfType("uint")).Return(nil)

	body := []byte(`{"amount":20.0}`)
	req := httptest.NewRequest(http.MethodPost, "/api/withdraw", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+generateTestJWT(userID))

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestBalanceHandler_NotFound(t *testing.T) {
	app, _, _, balanceRepo, _, _ := setupTestApp()

	userID := uint(404)

	balanceRepo.On("GetBalance", userID).Return(nil, errors.New("not found"))

	req := httptest.NewRequest(http.MethodGet, "/api/balance/404", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestJWT(userID))

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestStatementHandler_NotFound(t *testing.T) {
	app, _, txRepoMock, balanceRepo, userRepoMock, rateLimiterMock := setupTestApp()

	userID := uint(789)

	userRepoMock.On("GetByID", userID).Return(nil, errors.New("user not found"))
	txRepoMock.On("GetByUser", userID).Return(nil, errors.New("transactions not found"))
	balanceRepo.On("GetBalance", mock.AnythingOfType("uint")).Return(nil, errors.New("balance not found"))
	rateLimiterMock.On("CheckTransactionRateLimit", mock.AnythingOfType("uint")).Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/statement/789", nil)
	req.Header.Set("Authorization", "Bearer "+generateTestJWT(userID))

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestDepositHandler_Unauthorized(t *testing.T) {
	app, _, _, _, _, _ := setupTestApp()

	body := []byte(`{"amount":50.0}`)
	req := httptest.NewRequest(http.MethodPost, "/api/deposit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestCreateUser_Success(t *testing.T) {
	app, _, _, _, userRepo, _ := setupTestApp()

	userRepo.On("Create", mock.AnythingOfType("domain.User")).Return(nil)

	body := []byte(`{"name":"John","email":"john@example.com","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestCreateUser_InvalidJSON(t *testing.T) {
	app, _, _, _, _, _ := setupTestApp()

	body := []byte(`invalid json`)
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
