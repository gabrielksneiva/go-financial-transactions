package services

import (
	"time"

	d "go-financial-transactions/domain"
)

type TransactionDisplay struct {
	ID        string    `json:"id"`
	Amount    float64   `json:"amount"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

func ToTransactionDisplay(txs []d.Transaction) []TransactionDisplay {
	result := make([]TransactionDisplay, 0, len(txs))
	for _, tx := range txs {
		result = append(result, TransactionDisplay{
			ID:        tx.ID,
			Amount:    tx.Amount,
			Type:      tx.Type,
			Timestamp: tx.Timestamp,
		})
	}
	return result
}
