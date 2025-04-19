package services

import (
	"errors"
	"time"

	"github.com/gabrielksneiva/go-financial-transactions/domain"
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
	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	if err := s.RateLimiter.CheckTransactionRateLimit(userID); err != nil {
		return err
	}

	tx := domain.Transaction{
		ID:        uuid.New().String(),
		UserID:    userID,
		User:      domain.User{ID: userID},
		Amount:    amount,
		Timestamp: time.Now(),
		Type:      "deposit",
	}

	return s.producer.SendTransaction(tx)
}
