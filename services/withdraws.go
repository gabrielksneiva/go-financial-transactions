package services

import (
	"errors"
	"time"

	d "go-financial-transactions/domain"
	p "go-financial-transactions/producer"

	"github.com/google/uuid"
)

type WithdrawService struct {
	Repo        d.TransactionRepository
	BalanceRepo d.BalanceRepository
	producer    p.Producer
}

func NewWithdrawService(r d.TransactionRepository, b d.BalanceRepository, p p.Producer) *WithdrawService {
	return &WithdrawService{
		Repo:        r,
		BalanceRepo: b,
		producer:    p,
	}
}

func (s *WithdrawService) Withdraw(userID string, amount float64) error {
	balance, err := s.BalanceRepo.GetBalance(userID)
	if err != nil {
		return err
	}

	if balance.Amount < amount {
		return errors.New("insufficient funds")
	}

	tx := d.Transaction{
		ID:        uuid.New().String(),
		UserID:    userID,
		Amount:    -amount,
		Timestamp: time.Now(),
		Type:      "withdraw",
	}

	return s.producer.SendTransaction(tx)
}
