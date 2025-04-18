package domain

import (
	"time"
)

var (
	DepositTransaction  = "deposit"
	WithdrawTransaction = "withdraw"
)

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"size:255;not null"`
	Password string `gorm:"size:255;not null"`
	Email    string `gorm:"size:255;unique;not null"`
}

type Transaction struct {
	ID        string `gorm:"primaryKey"`
	UserID    uint   `gorm:"primaryKey" json:"user_id"` // Alterado de string para uint
	User      User   `gorm:"foreignKey:UserID;references:ID"`
	Amount    float64
	Timestamp time.Time
	Type      string
}

type Balance struct {
	UserID uint `gorm:"primaryKey" json:"user_id"` // Alterado de string para uint
	Amount float64
}

type TransactionRepository interface {
	Save(tx Transaction) error
	GetByUser(userID uint) ([]Transaction, error)
}

type BalanceRepository interface {
	UpdateBalance(tx Transaction) error
	GetBalance(userID uint) (*Balance, error)
}

type UserRepository interface {
	Create(user User) error
	GetByEmail(email string) (*User, error)
	GetByID(id uint) (*User, error)
}
