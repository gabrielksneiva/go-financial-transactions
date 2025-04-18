package api

import (
	"github.com/financialkafkaconsumerproject/producer/services"
	"github.com/gofiber/fiber/v2"
)

type App struct {
	Fiber    *fiber.App
	Handlers *Handlers
}

func NewApp(
	depositService *services.DepositService,
	withdrawService *services.WithdrawService,
	statementService *services.StatementService,
) *App {
	app := fiber.New()

	handlers := NewHandlers(depositService, withdrawService, statementService)

	RegisterRoutes(app, handlers)

	return &App{
		Fiber:    app,
		Handlers: handlers,
	}
}
