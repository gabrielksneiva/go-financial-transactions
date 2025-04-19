//go:build integration
// +build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/gabrielksneiva/go-financial-transactions/api"
	"github.com/gabrielksneiva/go-financial-transactions/config"
	"github.com/gabrielksneiva/go-financial-transactions/domain"
	"github.com/gabrielksneiva/go-financial-transactions/repositories"
	"github.com/gabrielksneiva/go-financial-transactions/services"
	"github.com/gabrielksneiva/go-financial-transactions/workers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type FakeRateLimiter struct{}

func (f *FakeRateLimiter) Allow(userID uint) (bool, error) {
	return true, nil
}

func (f *FakeRateLimiter) CheckTransactionRateLimit(userID uint) error {
	return nil
}

type ChannelWriter struct {
	Ch chan domain.Transaction
}

func (cw *ChannelWriter) SendTransaction(tx domain.Transaction) error {
	fmt.Printf("Producing fake message: %+v\n", tx)
	cw.Ch <- tx
	return nil
}

func (cw *ChannelWriter) Close() error {
	return nil
}

func setupTestApp(txChannel chan domain.Transaction) (*api.App, *repositories.GormRepository) {
	cfg := config.LoadConfig()
	db := repositories.InitDatabase(cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)
	repo := repositories.NewGormRepository(db)

	// Usa o fake writer que escreve no canal em vez do Kafka real
	fakeWriter := &ChannelWriter{Ch: txChannel}
	fakeLimiter := &FakeRateLimiter{}

	depositSvc := services.NewDepositService(repo, repo, fakeWriter, fakeLimiter)
	withdrawSvc := services.NewWithdrawService(repo, repo, fakeWriter, fakeLimiter)
	statementSvc := services.NewStatementService(repo, repo)
	userSvc := services.NewUserService(repo)

	app := api.NewApp(depositSvc, withdrawSvc, statementSvc, userSvc)

	return app, repo
}

func cleanupUserByEmail(repo *repositories.GormRepository, email string) error {
	user, err := repo.GetByEmail(email)
	if err != nil || user == nil {
		return nil // Nada para limpar
	}
	if err := repo.DeleteTransactionsByUserID(user.ID); err != nil {
		return err
	}
	return repo.Delete(email)
}

func TestIntegration_DepositFlow(t *testing.T) {
	txChannel := make(chan domain.Transaction, 10)
	app, repo := setupTestApp(txChannel)

	ctx := t.Context()

	go workers.StartWorkers(ctx, txChannel, 1, repo.GetDB())

	// Clean up any leftover data from previous runs
	user, _ := repo.GetByEmail("test@example.com")
	if user != nil {
		// Deleta transações primeiro, respeitando a FK
		_ = repo.DeleteTransactionsByUserID(user.ID)
		_ = repo.Delete("test@example.com")
	}

	// 1. Register user
	registerBody := map[string]string{
		"name":     "Test User",
		"email":    "test@example.com",
		"password": "password",
	}

	bodyBytes, _ := json.Marshal(registerBody)
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Fiber.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, res.StatusCode)

	// 2. Login user
	loginBody := map[string]string{
		"email":    "test@example.com",
		"password": "password",
	}
	bodyBytes, _ = json.Marshal(loginBody)
	req = httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	res, err = app.Fiber.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var loginResponse map[string]string
	err = json.NewDecoder(res.Body).Decode(&loginResponse)
	require.NoError(t, err)
	token := loginResponse["token"]
	assert.NotEmpty(t, token)

	// 3. Deposit
	depositBody := map[string]float64{
		"amount": 100,
	}
	bodyBytes, _ = json.Marshal(depositBody)
	req = httptest.NewRequest(http.MethodPost, "/api/deposit", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err = app.Fiber.Test(req)
	require.NoError(t, err)
	require.NotNil(t, res)

	body, _ := io.ReadAll(res.Body)
	t.Logf("Status: %d", res.StatusCode)
	t.Logf("Body: %s", string(body))
	assert.Equal(t, http.StatusAccepted, res.StatusCode)

	// 4. Check if transaction is saved in DB
	user, err = repo.GetByEmail("test@example.com")
	assert.NoError(t, err)

	require.Eventually(t, func() bool {
		txs, err := repo.GetTransactionsByUserID(user.ID)
		if err != nil || len(txs) == 0 {
			return false
		}
		return true
	}, 5*time.Second, 100*time.Millisecond)

	txs, err := repo.GetTransactionsByUserID(user.ID)
	require.NoError(t, err)
	require.Len(t, txs, 1)
	assert.Equal(t, 100.0, txs[0].Amount)
	assert.Equal(t, "deposit", txs[0].Type)

}

func TestIntegration_WithdrawFlow(t *testing.T) {
	txChannel := make(chan domain.Transaction, 10)
	app, repo := setupTestApp(txChannel)

	ctx := t.Context()
	go workers.StartWorkers(ctx, txChannel, 1, repo.GetDB())

	// Cleanup
	user, _ := repo.GetByEmail("withdraw@example.com")
	if user != nil {
		_ = repo.DeleteTransactionsByUserID(user.ID)
		_ = repo.Delete("withdraw@example.com")
	}

	// 1. Register user
	registerBody := map[string]string{
		"name":     "Withdraw User",
		"email":    "withdraw@example.com",
		"password": "password",
	}
	bodyBytes, _ := json.Marshal(registerBody)
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Fiber.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode)

	// 2. Login user
	loginBody := map[string]string{
		"email":    "withdraw@example.com",
		"password": "password",
	}
	bodyBytes, _ = json.Marshal(loginBody)
	req = httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	res, err = app.Fiber.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	var loginResponse map[string]string
	err = json.NewDecoder(res.Body).Decode(&loginResponse)
	require.NoError(t, err)
	token := loginResponse["token"]
	require.NotEmpty(t, token)

	// 3. Deposit R$ 1000
	depositBody := map[string]float64{
		"amount": 1000,
	}
	bodyBytes, _ = json.Marshal(depositBody)
	req = httptest.NewRequest(http.MethodPost, "/api/deposit", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err = app.Fiber.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, res.StatusCode)

	// Aguarda processamento
	user, _ = repo.GetByEmail("withdraw@example.com")
	require.Eventually(t, func() bool {
		txs, err := repo.GetTransactionsByUserID(user.ID)
		if err != nil || len(txs) == 0 {
			return false
		}
		return true
	}, 5*time.Second, 100*time.Millisecond)

	// 4. Withdraw R$ 200
	withdrawBody := map[string]float64{
		"amount": 200,
	}
	bodyBytes, _ = json.Marshal(withdrawBody)
	req = httptest.NewRequest(http.MethodPost, "/api/withdraw", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err = app.Fiber.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, res.StatusCode)

	// Aguarda processamento
	require.Eventually(t, func() bool {
		balance, err := repo.GetBalance(user.ID)
		if err != nil {
			return false
		}
		return balance.Amount == 800
	}, 5*time.Second, 100*time.Millisecond)

	balance, err := repo.GetBalance(user.ID)
	require.NoError(t, err)
	assert.Equal(t, 800.0, balance.Amount)
	assert.Equal(t, user.ID, balance.UserID)
	assert.Equal(t, 800.0, balance.Amount)

	// 5. Tenta sacar R$ 1000 (deve falhar por saldo insuficiente)
	overdraftBody := map[string]float64{
		"amount": 1000,
	}
	bodyBytes, _ = json.Marshal(overdraftBody)
	req = httptest.NewRequest(http.MethodPost, "/api/withdraw", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err = app.Fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestIntegration_StatementFlow(t *testing.T) {
	txChannel := make(chan domain.Transaction, 10)
	app, repo := setupTestApp(txChannel)

	ctx := t.Context()
	go workers.StartWorkers(ctx, txChannel, 1, repo.GetDB())

	// Cleanup antes do registro
	_ = cleanupUserByEmail(repo, "statement@example.com")

	// 1. Register user
	registerBody := map[string]string{
		"name":     "Statement User",
		"email":    "statement@example.com",
		"password": "password",
	}
	bodyBytes, _ := json.Marshal(registerBody)
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Fiber.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode)

	// 2. Login user
	loginBody := map[string]string{
		"email":    "statement@example.com",
		"password": "password",
	}
	bodyBytes, _ = json.Marshal(loginBody)
	req = httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	res, err = app.Fiber.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	var loginResponse map[string]string
	err = json.NewDecoder(res.Body).Decode(&loginResponse)
	require.NoError(t, err)
	token := loginResponse["token"]
	require.NotEmpty(t, token)

	var user *domain.User
	user, err = repo.GetByEmail("statement@example.com")
	require.NoError(t, err)

	// 3. Deposit R$ 500
	depositBody := map[string]float64{
		"amount": 500,
	}
	bodyBytes, _ = json.Marshal(depositBody)
	req = httptest.NewRequest(http.MethodPost, "/api/deposit", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err = app.Fiber.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, res.StatusCode)

	require.Eventually(t, func() bool {
		balance, err := repo.GetBalance(user.ID)
		return err == nil && balance.Amount >= 500
	}, 5*time.Second, 100*time.Millisecond)

	// 4. Withdraw R$ 200
	withdrawBody := map[string]float64{
		"amount": 200,
	}
	bodyBytes, _ = json.Marshal(withdrawBody)
	req = httptest.NewRequest(http.MethodPost, "/api/withdraw", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err = app.Fiber.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, res.StatusCode)

	require.Eventually(t, func() bool {
		txs, err := repo.GetTransactionsByUserID(user.ID)
		return err == nil && len(txs) >= 2
	}, 5*time.Second, 100*time.Millisecond)

	// 5. Get statement
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/statement/%d", user.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, err = app.Fiber.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	var statementResponse map[string]any
	err = json.NewDecoder(res.Body).Decode(&statementResponse)
	require.NoError(t, err)

	txs := statementResponse["transactions"].([]interface{})
	require.Len(t, txs, 2)

	sort.Slice(txs, func(i, j int) bool {
		tx1 := txs[i].(map[string]interface{})
		tx2 := txs[j].(map[string]interface{})
		return tx1["timestamp"].(string) < tx2["timestamp"].(string)
	})

	tx0 := txs[0].(map[string]interface{})
	tx1 := txs[1].(map[string]interface{})

	assert.Equal(t, 500.0, tx0["amount"].(float64))
	assert.Equal(t, "deposit", tx0["type"].(string))
	assert.Equal(t, -200.0, tx1["amount"].(float64))
	assert.Equal(t, "withdrawal", tx1["type"].(string))

}
