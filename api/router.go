package api

import (
	"github.com/gabrielksneiva/go-financial-transactions/api/middleware"
	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App, h *Handlers) {

	app.Post("/api/login", h.LoginHandler)
	app.Post("/api/register", h.RegisterHandler)

	api := app.Group("/api", middleware.JWTMiddleware())
	api.Post("/deposit", h.CreateDepositHandler)
	api.Post("/withdraw", h.CreateWithdrawHandler)
	api.Get("/balance/:user_id", h.GetBalanceHandler)
	api.Get("/statement/:user_id", h.GetStatementHandler)
}
