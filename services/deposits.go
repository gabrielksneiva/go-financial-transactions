package services

import (
	"time"

	d "github.com/financialkafkaconsumerproject/producer/domain"
	"github.com/financialkafkaconsumerproject/producer/producer"
	"github.com/google/uuid"
)

type DepositService struct {
	Repo        d.TransactionRepository
	BalanceRepo d.BalanceRepository
}

func NewDepositService(r d.TransactionRepository, b d.BalanceRepository) *DepositService {
	return &DepositService{
		Repo:        r,
		BalanceRepo: b,
	}
}

func (s *DepositService) Deposit(userID string, amount float64) error {
	tx := d.Transaction{
		ID:        uuid.New().String(),
		UserID:    userID,
		Amount:    amount,
		Timestamp: time.Now(),
		Type:      "deposit",
	}

	return producer.SendTransaction(tx)
}
