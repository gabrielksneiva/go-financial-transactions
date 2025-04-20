// frontend/router.go
package frontend

import (
	"fmt"
	"os"

	"github.com/gabrielksneiva/go-financial-transactions/api/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

func SetupRoutes(app *fiber.App) {
	// Proxy de todas as chamadas /api/* para o back-end
	app.Use("/api/*", func(c *fiber.Ctx) error {
		apiPort := os.Getenv("API_PORT")
		target := fmt.Sprintf("http://localhost:%s%s", apiPort, c.OriginalURL())
		return proxy.Do(c, target)
	})

	// Recursos estáticos e páginas
	app.Static("/static", "./frontend/static")
	app.Get("/login", LoginPage)
	app.Get("/register", RegisterPage)
	app.Get("/dashboard", middleware.JWTProtected(), Dashboard)
	app.Get("/dashboard/extract", middleware.JWTProtected(), TransactionExtractPartial)
}
