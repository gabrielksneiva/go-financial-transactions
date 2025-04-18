package services

import (
	"time"

	d "go-financial-transactions/domain"
	p "go-financial-transactions/producer"

	"github.com/google/uuid"
)

type DepositService struct {
	Repo        d.TransactionRepository
	BalanceRepo d.BalanceRepository
	producer    p.Producer
}

func NewDepositService(r d.TransactionRepository, b d.BalanceRepository, p p.Producer) *DepositService {
	return &DepositService{
		Repo:        r,
		BalanceRepo: b,
		producer:    p,
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

	return s.producer.SendTransaction(tx)
}
