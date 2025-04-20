package services

import (
	"time"

	d "github.com/gabrielksneiva/go-financial-transactions/domain"
)

type TransactionDisplay struct {
	ID            string    `json:"id"`
	Amount        float64   `json:"amount"`
	Type          string    `json:"type"`
	Status        string    `json:"status"`
	WalletAddress string    `json:"wallet"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func ToTransactionDisplay(txs []d.Transaction) []TransactionDisplay {
	result := make([]TransactionDisplay, 0, len(txs))
	for _, tx := range txs {
		result = append(result, TransactionDisplay{
			ID:        tx.ID,
			Amount:    tx.Amount,
			Type:      tx.Type,
			CreatedAt: tx.CreatedAt,
			UpdatedAt: tx.UpdatedAt,
			//			Status:        tx.Status,
			WalletAddress: tx.WalletAddress,
		})
	}
	return result
}
