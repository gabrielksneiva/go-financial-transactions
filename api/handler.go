package api

import (
	"github.com/financialkafkaconsumerproject/producer/services"
	"github.com/gofiber/fiber/v2"
)

type Handlers struct {
	DepositService   *services.DepositService
	WithdrawService  *services.WithdrawService
	StatementService *services.StatementService
}

type TransactionRequest struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}

type BalanceResponse struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}

type StatementResponse struct {
	UserID       string                        `json:"user_id"`
	Balance      float64                       `json:"balance"`
	Transactions []services.TransactionDisplay `json:"transactions"`
}

func NewHandlers(
	deposit *services.DepositService,
	withdraw *services.WithdrawService,
	statement *services.StatementService,
) *Handlers {
	return &Handlers{
		DepositService:   deposit,
		WithdrawService:  withdraw,
		StatementService: statement,
	}
}

func (h *Handlers) DepositHandler(c *fiber.Ctx) error {
	var req TransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	if err := h.DepositService.Deposit(req.UserID, req.Amount); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Deposit submitted",
	})
}

func (h *Handlers) WithdrawHandler(c *fiber.Ctx) error {
	var req TransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	if err := h.WithdrawService.Withdraw(req.UserID, req.Amount); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Withdrawal submitted",
	})
}

func (h *Handlers) BalanceHandler(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	amount, err := h.StatementService.GetBalance(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(BalanceResponse{
		UserID: userID,
		Amount: amount,
	})
}

func (h *Handlers) StatementHandler(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	statement, err := h.StatementService.GetStatement(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(StatementResponse{
		UserID:       userID,
		Balance:      statement.Balance,
		Transactions: services.ToTransactionDisplay(statement.Transactions),
	})
}
