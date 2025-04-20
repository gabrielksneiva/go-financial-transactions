package api

import (
	"github.com/gabrielksneiva/go-financial-transactions/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type App struct {
	Fiber    *fiber.App
	Handlers *Handlers
}

func NewApp(
	depositService *services.DepositService,
	withdrawService *services.WithdrawService,
	statementService *services.StatementService,
	userService *services.UserService,
) *App {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:4000",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	handlers := NewHandlers(depositService, withdrawService, statementService, userService)

	RegisterRoutes(app, handlers)

	return &App{
		Fiber:    app,
		Handlers: handlers,
	}
}
