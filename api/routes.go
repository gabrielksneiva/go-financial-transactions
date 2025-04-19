package api

import (
	"github.com/gabrielksneiva/go-financial-transactions/api/middleware"
	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App, h *Handlers) {

	app.Post("/login", h.LoginHandler)
	app.Post("/register", h.RegisterHandler)

	app.Use(middleware.JWTMiddleware())

	api := app.Group("/api")
	api.Post("/deposit", h.CreateDepositHandler)
	api.Post("/withdraw", h.CreateWithdrawHandler)
	api.Get("/balance/:user_id", h.GetBalanceHandler)
	api.Get("/statement/:user_id", h.GetStatementHandler)
}
