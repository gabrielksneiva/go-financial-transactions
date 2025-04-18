package api

import (
	"strconv"

	"github.com/gabrielksneiva/go-financial-transactions/services"

	"github.com/gofiber/fiber/v2"
)

type Handlers struct {
	DepositService   *services.DepositService
	WithdrawService  *services.WithdrawService
	StatementService *services.StatementService
	UserService      *services.UserService
}

type TransactionRequest struct {
	UserID uint    `json:"user_id"`
	Amount float64 `json:"amount"`
}

type BalanceResponse struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}

type StatementResponse struct {
	UserID       string                        `json:"user_id"`
	UserEmail    string                        `json:"user_email"`
	Balance      float64                       `json:"balance"`
	Transactions []services.TransactionDisplay `json:"transactions"`
}

func NewHandlers(
	deposit *services.DepositService,
	withdraw *services.WithdrawService,
	statement *services.StatementService,
	user *services.UserService,
) *Handlers {
	return &Handlers{
		DepositService:   deposit,
		WithdrawService:  withdraw,
		StatementService: statement,
		UserService:      user,
	}
}

func (h *Handlers) CreateDepositHandler(c *fiber.Ctx) error {
	var req TransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	// Passa userIDUint como uint para o serviço
	if err := h.DepositService.Deposit(req.UserID, req.Amount); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Deposit submitted",
	})
}

func (h *Handlers) CreateWithdrawHandler(c *fiber.Ctx) error {
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

func (h *Handlers) GetBalanceHandler(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	// Converte userID para uint
	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}
	amount, err := h.StatementService.GetBalance(uint(userIDUint))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(BalanceResponse{
		UserID: userID,
		Amount: amount,
	})
}

func (h *Handlers) GetStatementHandler(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	statement, err := h.StatementService.GetStatement(uint(userIDUint))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := h.UserService.GetUserByID(uint(userIDUint))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(StatementResponse{
		UserID:       userID,
		UserEmail:    user.Email,
		Balance:      statement.Balance,
		Transactions: services.ToTransactionDisplay(statement.Transactions),
	})
}

func (h *Handlers) CreateUsersHandler(c *fiber.Ctx) error {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	if err := h.UserService.CreateUser(req.Name, req.Email, req.Password); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created",
	})
}
