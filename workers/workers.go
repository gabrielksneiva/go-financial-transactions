package workers

import (
	"context"
	"fmt"
	"log"

	d "github.com/financialkafkaconsumerproject/producer/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// StartWorkers initializes 'count' workers using domain repositories
func StartWorkers(ctx context.Context, ch <-chan d.Transaction, count int, db *gorm.DB) {
	for i := 0; i < count; i++ {
		go worker(ctx, ch, i, db)
	}
}

func worker(
	ctx context.Context,
	ch <-chan d.Transaction,
	workerID int,
	db *gorm.DB,
) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("🛑 Worker %d stopped.\n", workerID)
			return

		case tx := <-ch:
			fmt.Printf("👷 Worker %d processing transaction %s (%.2f)\n", workerID, tx.ID, tx.Amount)

			err := db.Transaction(func(txDB *gorm.DB) error {
				var balance d.Balance

				// Lock user's balance row
				if err := txDB.Clauses(clause.Locking{Strength: "UPDATE"}).
					Where("user_id = ?", tx.UserID).
					FirstOrCreate(&balance, d.Balance{UserID: tx.UserID}).Error; err != nil {
					return err
				}

				newBalance := balance.Amount + tx.Amount
				if newBalance < 0 {
					return fmt.Errorf("insufficient funds for user %s", tx.UserID)
				}

				if err := txDB.Model(&d.Balance{}).
					Where("user_id = ?", balance.UserID).
					Update("amount", newBalance).Error; err != nil {
					return err
				}

				if err := txDB.Create(&tx).Error; err != nil {
					return err
				}

				return nil
			})

			if err != nil {
				log.Printf("❌ Worker %d failed to process transaction %s: %v", workerID, tx.ID, err)
			}
		}
	}
}
