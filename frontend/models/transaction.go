package models

// Transaction espelha o JSON que sai da /api/transactions
type Transaction struct {
	ID     string  `json:"id"`
	Type   string  `json:"type"`
	Amount float64 `json:"amount"`
	Status string  `json:"status"`
}
