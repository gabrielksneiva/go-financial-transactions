package services

import (
	"errors"
	"time"

	d "github.com/financialkafkaconsumerproject/producer/domain"
	"github.com/financialkafkaconsumerproject/producer/producer"
	"github.com/google/uuid"
)

type WithdrawService struct {
	Repo        d.TransactionRepository
	BalanceRepo d.BalanceRepository
}

func NewWithdrawService(r d.TransactionRepository, b d.BalanceRepository) *WithdrawService {
	return &WithdrawService{
		Repo:        r,
		BalanceRepo: b,
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

	return producer.SendTransaction(tx)
}
