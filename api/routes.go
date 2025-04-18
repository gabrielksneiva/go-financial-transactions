package api

import (
	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App, h *Handlers) {
	app.Post("/deposit", h.CreateDepositHandler)
	app.Post("/withdraw", h.CreateWithdrawHandler)
	app.Post("/user", h.CreateUsersHandler)
	app.Get("/balance/:user_id", h.GetBalanceHandler)
	app.Get("/statement/:user_id", h.GetStatementHandler)
}
