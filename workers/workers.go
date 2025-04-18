package workers

import (
	"context"
	"fmt"
	"log"

	d "github.com/financialkafkaconsumerproject/producer/domain"
)

// StartWorkers initializes 'count' workers with shared context and transaction channel
func StartWorkers(ctx context.Context, ch <-chan d.Transaction, count int, txRepo d.TransactionRepository, balanceRepo d.BalanceRepository) {
	for i := 0; i < count; i++ {
		go worker(ctx, ch, i, txRepo, balanceRepo)
	}
}

func worker(ctx context.Context, ch <-chan d.Transaction, workerID int, txRepo d.TransactionRepository, balanceRepo d.BalanceRepository) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("🛑 Worker %d stopped.\n", workerID)
			return
		case tx := <-ch:
			fmt.Printf("👷 Worker %d processing transaction %s (%.2f)\n", workerID, tx.ID, tx.Amount)

			// Save transaction
			if err := txRepo.Save(tx); err != nil {
				log.Printf("❌ Worker %d: failed to save transaction: %v", workerID, err)
				continue
			}

			// Update balance
			if err := balanceRepo.UpdateBalance(tx); err != nil {
				log.Printf("❌ Worker %d: failed to update balance: %v", workerID, err)
			}
		}
	}
}
