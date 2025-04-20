package api

import (
	"github.com/gabrielksneiva/go-financial-transactions/api/middleware"
	"github.com/gabrielksneiva/go-financial-transactions/domain"
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
	Amount float64 `json:"amount"`
}

type BalanceResponse struct {
	UserID uint    `json:"user_id"`
	Amount float64 `json:"amount"`
}

type StatementResponse struct {
	UserID       uint                          `json:"user_id"`
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
	userID := c.Locals("user_id").(uint)

	var req TransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	if err := h.DepositService.Deposit(userID, req.Amount); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Deposit submitted",
	})
}

func (h *Handlers) CreateWithdrawHandler(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var req TransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	if err := h.WithdrawService.Withdraw(userID, req.Amount); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Withdrawal submitted",
	})
}

func (h *Handlers) GetBalanceHandler(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	amount, err := h.StatementService.GetBalance(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(BalanceResponse{
		UserID: userID,
		Amount: amount,
	})
}

func (h *Handlers) GetStatementHandler(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	statement, err := h.StatementService.GetStatement(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	userRetrieved, err := h.UserService.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(StatementResponse{
		UserID:       userID,
		UserEmail:    userRetrieved.Email,
		Balance:      statement.Balance,
		Transactions: services.ToTransactionDisplay(statement.Transactions),
	})
}

func (h *Handlers) RegisterHandler(c *fiber.Ctx) error {
	// Agora só name, email e password são obrigatórios
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Todos os campos são obrigatórios"})
	}

	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		// WalletAddress fica em branco
	}

	if err := h.UserService.CreateUser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created",
	})
}

func (h *Handlers) LoginHandler(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON inválido"})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Todos os campos são obrigatórios"})
	}

	user, err := h.UserService.Authenticate(req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	token, err := middleware.GenerateJWT(user)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"token": token})
}
