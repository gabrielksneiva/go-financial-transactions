package services

import (
	"time"

	d "github.com/gabrielksneiva/go-financial-transactions/domain"
	p "github.com/gabrielksneiva/go-financial-transactions/producer"

	"github.com/google/uuid"
)

type DepositService struct {
	Repo        d.TransactionRepository
	BalanceRepo d.BalanceRepository
	producer    p.Producer
	RateLimiter d.RateLimiter
}

func NewDepositService(r d.TransactionRepository, b d.BalanceRepository, p p.Producer, rate d.RateLimiter) *DepositService {
	return &DepositService{
		Repo:        r,
		BalanceRepo: b,
		producer:    p,
		RateLimiter: rate,
	}
}

func (s *DepositService) Deposit(userID uint, amount float64) error {
	if err := s.RateLimiter.CheckTransactionRateLimit(userID); err != nil {
		return err
	}

	tx := d.Transaction{
		ID:        uuid.New().String(),
		UserID:    userID,
		Amount:    amount,
		Timestamp: time.Now(),
		Type:      "deposit",
	}

	return s.producer.SendTransaction(tx)
}
