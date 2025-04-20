package middleware

import (
	"fmt"
	"os"
	"strings"
	"time"

	d "github.com/gabrielksneiva/go-financial-transactions/domain"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(user *d.User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET is not set")
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func JWTProtected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenStr := ""

		// 1. Primeiro tenta pegar do header
		auth := c.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			tokenStr = strings.TrimPrefix(auth, "Bearer ")
		}

		// 2. Se não encontrar no header, tenta pegar do cookie
		if tokenStr == "" {
			tokenStr = c.Cookies("token")
		}

		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing or invalid token",
			})
		}

		// ✅ Use MapClaims sem ponteiro
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// 🛠️ Converte user_id para uint com verificação
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or missing user_id in token",
			})
		}

		// ✅ Salva como uint para evitar cast nos handlers
		c.Locals("user_id", uint(userIDFloat))
		c.Locals("email", claims["email"])

		return c.Next()
	}
}
