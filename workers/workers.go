package workers

import (
	"context"
	"fmt"
	"log"
	"os"

	d "github.com/gabrielksneiva/go-financial-transactions/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func StartWorkers(ctx context.Context, ch <-chan d.Transaction, count int, db *gorm.DB, blockchain d.BlockchainClient, repo d.TransactionRepository) {
	for i := 0; i < count; i++ {
		go worker(ctx, ch, i, db, blockchain, repo)
	}
}

func worker(
	ctx context.Context,
	ch <-chan d.Transaction,
	workerID int,
	db *gorm.DB,
	b d.BlockchainClient,
	repo d.TransactionRepository,
) {
	log.Printf("🚧 Worker %d iniciado", workerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("🛑 Worker %d encerrado via contexto", workerID)
			return

		case tx := <-ch:
			log.Printf("📥 Worker %d recebeu transação %s (%.2f)", workerID, tx.ID, tx.Amount)

			err := db.Transaction(func(txDB *gorm.DB) error {
				log.Printf("🔒 Worker %d bloqueando saldo do usuário %d...", workerID, tx.UserID)
				var balance d.Balance

				if err := txDB.Clauses(clause.Locking{Strength: "UPDATE"}).
					Where("user_id = ?", tx.UserID).
					FirstOrCreate(&balance, d.Balance{UserID: tx.UserID}).Error; err != nil {
					log.Printf("❌ Worker %d: erro ao buscar/criar saldo: %v", workerID, err)
					return err
				}

				var newBalance float64
				if tx.Type == "withdraw" {
					newBalance = balance.Amount - tx.Amount
					log.Printf("💸 Worker %d: saldo atual %.2f → novo saldo %.2f (saque)", workerID, balance.Amount, newBalance)

					if newBalance < 0 {
						log.Printf("⛔ Worker %d: fundos insuficientes para usuário %d", workerID, tx.UserID)
						return fmt.Errorf("insufficient funds for user %d", tx.UserID)
					}
				} else {
					newBalance = balance.Amount + tx.Amount
					log.Printf("💰 Worker %d: saldo atual %.2f → novo saldo %.2f (depósito)", workerID, balance.Amount, newBalance)
				}

				if err := txDB.Model(&d.Balance{}).
					Where("user_id = ?", tx.UserID).
					Update("amount", newBalance).Error; err != nil {
					log.Printf("❌ Worker %d: erro ao atualizar saldo: %v", workerID, err)
					return err
				}

				tx.Status = "PENDING"
				if err := txDB.Create(&tx).Error; err != nil {
					log.Printf("❌ Worker %d: erro ao salvar transação no banco: %v", workerID, err)
					return err
				}

				log.Printf("✅ Worker %d: transação %s persistida com sucesso", workerID, tx.ID)
				return nil
			})

			if err != nil {
				log.Printf("❌ Worker %d falhou ao processar transação %s: %v", workerID, tx.ID, err)
				continue
			}

			// Processa saque
			if tx.Type == "withdraw" {
				log.Printf("🚀 Worker %d iniciando envio de TRX para usuário %d", workerID, tx.UserID)

				var user d.User
				if err := db.First(&user, tx.UserID).Error; err != nil {
					log.Printf("❌ Worker %d: erro ao buscar usuário %d: %v", workerID, tx.UserID, err)
					continue
				}

				if user.WalletAddress == "" {
					log.Printf("⚠️ Worker %d: usuário %d sem endereço TRON", workerID, tx.UserID)
					continue
				}

				// Marcar como PENDING
				if err := repo.UpdateTransactionStatus(tx.ID, "PENDING"); err != nil {
					log.Printf("⚠️ Worker %d: erro ao atualizar status para PENDING: %v", workerID, err)
					continue
				}

				txOut := d.BlockchainTransaction{
					FromAddress: os.Getenv("TRON_FROM_ADDR"),
					ToAddress:   user.WalletAddress,
					Amount:      int64(tx.Amount * 1e6),
					Visible:     true,
				}

				log.Printf("📤 Worker %d: enviando TRX para %s...", workerID, user.WalletAddress)

				result, err := b.SendSignedTRX(txOut, tx.ID)
				if err != nil {
					log.Printf("❌ Worker %d: erro ao enviar TRX: %v", workerID, err)
					// ❗ Aqui você pode marcar como "FAILED", se quiser adicionar esse status
					if err := repo.UpdateTransactionStatus(tx.ID, "FAILED"); err != nil {
						log.Printf("⚠️ Worker %d: erro ao atualizar status para FAILED: %v", workerID, err)
					} else {
						log.Printf("📝 Worker %d: status da transação %s atualizado para FAILED", workerID, tx.ID)
					}
					continue
				}

				log.Printf("✅ Worker %d: transação blockchain enviada com sucesso | txID: %s", workerID, result.TxID)

				// Atualiza o hash da transação
				if err := repo.UpdateTransactionHash(tx.ID, result.TxID); err != nil {
					log.Printf("⚠️ Worker %d: erro ao salvar hash da transação %s: %v", workerID, tx.ID, err)
				} else {
					log.Printf("📝 Worker %d: hash da transação %s atualizado no banco", workerID, tx.ID)
				}

				// Marcar como COMPLETED
				if err := repo.UpdateTransactionStatus(tx.ID, "COMPLETED"); err != nil {
					log.Printf("⚠️ Worker %d: erro ao atualizar status para COMPLETED: %v", workerID, err)
				}
			}

		}
	}
}
