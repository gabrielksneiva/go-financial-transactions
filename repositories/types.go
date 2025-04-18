package repositories

import "time"

type Transaction struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index;not null" json:"user_id"`
	Amount    float64   `json:"amount"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
}

type Balance struct {
	UserID string  `gorm:"primaryKey" json:"user_id"`
	Amount float64 `json:"amount"`
}

