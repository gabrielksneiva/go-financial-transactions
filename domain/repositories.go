package domain

import (
	"time"
)

var (
	DepositTransaction  = "deposit"
	WithdrawTransaction = "withdraw"
)

type Transaction struct {
	ID        string
	UserID    string
	Amount    float64
	Timestamp time.Time
	Type      string
}

type Balance struct {
	UserID string
	Amount float64
}

type TransactionRepository interface {
	Save(tx Transaction) error
	GetByUser(userID string) ([]Transaction, error)
}

type BalanceRepository interface {
	UpdateBalance(tx Transaction) error
	GetBalance(userID string) (*Balance, error)
}
