package services

import (
	"errors"
	"time"

	d "github.com/gabrielksneiva/go-financial-transactions/domain"
	p "github.com/gabrielksneiva/go-financial-transactions/producer"

	"github.com/google/uuid"
)

type WithdrawService struct {
	Repo        d.TransactionRepository
	BalanceRepo d.BalanceRepository
	producer    p.Producer
	RateLimiter d.RateLimiter
}

func NewWithdrawService(r d.TransactionRepository, b d.BalanceRepository, p p.Producer, rate d.RateLimiter) *WithdrawService {
	return &WithdrawService{
		Repo:        r,
		BalanceRepo: b,
		producer:    p,
		RateLimiter: rate,
	}
}

func (s *WithdrawService) Withdraw(userID uint, amount float64) error {
	if err := s.RateLimiter.CheckTransactionRateLimit(userID); err != nil {
		return err
	}

	bal, err := s.BalanceRepo.GetBalance(userID)
	if err != nil {
		return err
	}

	if bal.Amount < amount {
		return errors.New("insufficient funds")
	}

	tx := d.Transaction{
		ID:        uuid.New().String(),
		UserID:    userID,
		User:      d.User{ID: userID},
		Amount:    -amount,
		Timestamp: time.Now(),
		Type:      "withdrawal",
	}

	return s.producer.SendTransaction(tx)
}
