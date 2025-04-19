package middleware_test

import (
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gabrielksneiva/go-financial-transactions/api/middleware"
	"github.com/gabrielksneiva/go-financial-transactions/domain"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func setup() string {
	// Definindo a variável de ambiente para o JWT_SECRET
	if secret := os.Getenv("JWT_SECRET"); secret == "" {
		os.Setenv("JWT_SECRET", "secret") // Garantir que o JWT_SECRET esteja configurado
		return secret
	}
	return os.Getenv("JWT_SECRET")
}

func TestGenerateJWT_Success(t *testing.T) {
	// Setup
	secret := setup()

	user := &domain.User{
		ID:    1,
		Email: "test@example.com",
		Role:  "user",
	}

	token, err := middleware.GenerateJWT(user)

	// Validações
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verifica a estrutura do token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Valida se está usando o método de assinatura correto
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			t.Errorf("expected signing method to be HMAC")
		}
		return []byte(secret), nil
	})
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, float64(user.ID), claims["user_id"])
	assert.Equal(t, user.Email, claims["email"])
	assert.Equal(t, user.Role, claims["role"])
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	app := fiber.New()

	app.Get("/protected", middleware.JWTMiddleware(), func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		assert.Equal(t, uint(1), userID)
		return c.SendString("Success")
	})

	// Gerando o token válido
	user := &domain.User{
		ID:    1,
		Email: "test@example.com",
		Role:  "user",
	}
	token, err := middleware.GenerateJWT(user)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGenerateJWT_InvalidSecret(t *testing.T) {
	// Setup com variável de ambiente faltando
	os.Unsetenv("JWT_SECRET")

	user := &domain.User{
		ID:    1,
		Email: "test@example.com",
		Role:  "user",
	}

	_, err := middleware.GenerateJWT(user)
	assert.Error(t, err)
}

func TestJWTMiddleware_MissingAuthorizationHeader(t *testing.T) {
	app := fiber.New()

	app.Get("/protected", middleware.JWTMiddleware(), func(c *fiber.Ctx) error {
		return c.SendString("Success")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestJWTMiddleware_InvalidAuthorizationFormat(t *testing.T) {
	app := fiber.New()

	app.Get("/protected", middleware.JWTMiddleware(), func(c *fiber.Ctx) error {
		return c.SendString("Success")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidHeader")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	app := fiber.New()

	app.Get("/protected", middleware.JWTMiddleware(), func(c *fiber.Ctx) error {
		return c.SendString("Success")
	})

	// Token inválido
	invalidToken := "Bearer invalidtoken"
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", invalidToken)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestJWTMiddleware_ExpiredToken(t *testing.T) {
	app := fiber.New()

	app.Get("/protected", middleware.JWTMiddleware(), func(c *fiber.Ctx) error {
		return c.SendString("Success")
	})

	// Gerar um token expirado
	claims := jwt.MapClaims{
		"user_id": 1,
		"exp":     time.Now().Add(-time.Hour).Unix(), // Token expirado
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte("secret"))
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signedToken)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestJWTMiddleware_MissingUserIDClaim(t *testing.T) {
	app := fiber.New()

	app.Get("/protected", middleware.JWTMiddleware(), func(c *fiber.Ctx) error {
		return c.SendString("Success")
	})

	// Gerar um token sem o "user_id" na claim
	claims := jwt.MapClaims{
		"email": "test@example.com",
		"role":  "user",
		"exp":   time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte("secret"))
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signedToken)

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
