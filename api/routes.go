package api

import (
	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App, h *Handlers) {
	app.Post("/deposit", h.DepositHandler)
	app.Post("/withdraw", h.WithdrawHandler)
	app.Get("/balance/:user_id", h.BalanceHandler)
	app.Get("/statement/:user_id", h.StatementHandler)
}
