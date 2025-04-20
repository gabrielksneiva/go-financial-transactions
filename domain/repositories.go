package domain

import (
	"context"
	"time"
)

var (
	DepositTransaction  = "deposit"
	WithdrawTransaction = "withdraw"
)

type User struct {
	ID            uint `gorm:"primaryKey"`
	Name          string
	Email         string `gorm:"unique"`
	Password      string
	Role          string // Ex: "user" ou "admin"
	WalletAddress string
}

type Transaction struct {
	ID            string `gorm:"type:text;primaryKey"`
	UserID        uint
	User          User
	Amount        float64
	Timestamp     time.Time
	Type          string
	WalletAddress string
	TxHash        string
	Status        string `gorm:"default:PENDING"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Balance struct {
	UserID uint `gorm:"primaryKey" json:"user_id"`
	Amount float64
}

type BlockchainTransaction struct {
	ToAddress   string `json:"to_address"`
	FromAddress string `json:"owner_address"`
	Amount      int64  `json:"amount"`
	Visible     bool   `json:"visible"`
	Timestamp   time.Time
}

type BlockchainTxResult struct {
	TxID        string
	FromAddress string
	ToAddress   string
	Amount      float64
}

type RedisClientInterface interface {
	Get(ctx context.Context, key string) (int, error)
	Set(ctx context.Context, key string, value int) error
	Incr(ctx context.Context, key string) (int, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
}

type RateLimiter interface {
	CheckTransactionRateLimit(userID uint) error
}

type TransactionRepository interface {
	Save(tx Transaction) error
	GetByUser(userID uint) ([]Transaction, error)
	GetTransactionsByUserID(userID uint) ([]Transaction, error)
	UpdateTransactionHash(txID string, txHash string) error
	UpdateTransactionStatus(txID string, status string) error // 👈 Novo método
}

type BalanceRepository interface {
	UpdateBalance(tx Transaction) error
	GetBalance(userID uint) (*Balance, error)
}

type UserRepository interface {
	Create(user User) error
	GetByEmail(email string) (*User, error)
	GetByID(id uint) (*User, error)
	Delete(email string) error
}

type BlockchainClient interface {
	SendSignedTRX(tx BlockchainTransaction, transactionID string) (*BlockchainTxResult, error)
}
