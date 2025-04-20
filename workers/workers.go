package workers

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	d "github.com/gabrielksneiva/go-financial-transactions/domain"
)

const (
	TypeDeposit  = "deposit"
	TypeWithdraw = "withdraw"
	TypeRefund   = "refund"

	StatusPending   = "PENDING"
	StatusCompleted = "COMPLETED"
	StatusFailed    = "FAILED"
)

func Worker(ctx context.Context, id int, jobs <-chan d.Transaction, db *gorm.DB, b d.BlockchainClient, repo d.TransactionRepository) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("🛑 Worker %d encerrado", id)
			return

		case tx := <-jobs:
			processTransaction(tx, id, db, b, repo)
		}
	}
}

func processTransaction(tx d.Transaction, workerID int, db *gorm.DB, b d.BlockchainClient, repo d.TransactionRepository) {
	log.Printf("📥 Worker %d recebeu transação %s (%.2f)", workerID, tx.ID, tx.Amount)

	err := db.Transaction(func(txDB *gorm.DB) error {
		return handleTransactionDB(txDB, &tx, workerID)
	})

	if err != nil {
		log.Printf("❌ Worker %d falhou ao processar transação %s: %v", workerID, tx.ID, err)
		return
	}

	if tx.Type == TypeWithdraw {
		handleWithdrawal(tx, workerID, db, b, repo)
	}
}

func handleTransactionDB(txDB *gorm.DB, tx *d.Transaction, workerID int) error {
	var balance d.Balance

	if err := txDB.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", tx.UserID).
		FirstOrCreate(&balance, d.Balance{UserID: tx.UserID}).Error; err != nil {
		log.Printf("❌ Worker %d: erro ao buscar/criar saldo: %v", workerID, err)
		return err
	}

	var newBalance float64
	switch tx.Type {
	case TypeWithdraw:
		newBalance = balance.Amount - tx.Amount
		if newBalance < 0 {
			log.Printf("⛔ Worker %d: fundos insuficientes para usuário %d", workerID, tx.UserID)
			return fmt.Errorf("insufficient funds for user %d", tx.UserID)
		}
		log.Printf("💸 Worker %d: saldo atual %.2f → novo saldo %.2f (saque)", workerID, balance.Amount, newBalance)
	case TypeDeposit:
		newBalance = balance.Amount + tx.Amount
		log.Printf("💰 Worker %d: saldo atual %.2f → novo saldo %.2f (depósito)", workerID, balance.Amount, newBalance)
	}

	if err := txDB.Model(&d.Balance{}).
		Where("user_id = ?", tx.UserID).
		Update("amount", newBalance).Error; err != nil {
		log.Printf("❌ Worker %d: erro ao atualizar saldo: %v", workerID, err)
		return err
	}

	tx.Status = map[string]string{
		TypeDeposit:  StatusCompleted,
		TypeWithdraw: StatusPending,
	}[tx.Type]

	if err := txDB.Create(tx).Error; err != nil {
		log.Printf("❌ Worker %d: erro ao salvar transação: %v", workerID, err)
		return err
	}

	return nil
}

func handleWithdrawal(tx d.Transaction, workerID int, db *gorm.DB, b d.BlockchainClient, repo d.TransactionRepository) {
	var user d.User
	if err := db.First(&user, tx.UserID).Error; err != nil {
		log.Printf("❌ Worker %d: erro ao buscar usuário: %v", workerID, err)
		return
	}

	if user.WalletAddress == "" {
		log.Printf("⚠️ Worker %d: usuário %d sem endereço TRON", workerID, tx.UserID)
		return
	}

	if err := repo.UpdateTransactionStatus(tx.ID, StatusPending); err != nil {
		log.Printf("⚠️ Worker %d: erro ao atualizar status para PENDING: %v", workerID, err)
		return
	}

	txOut := d.BlockchainTransaction{
		FromAddress: os.Getenv("TRON_FROM_ADDR"),
		ToAddress:   user.WalletAddress,
		Amount:      int64(tx.Amount * 1e6),
		Visible:     true,
	}

	result, err := b.SendSignedTRX(txOut, tx.ID)
	if err != nil {
		log.Printf("❌ Worker %d: erro ao enviar TRX: %v", workerID, err)
		handleFailedTransaction(tx, workerID, db, repo)
		return
	}

	log.Printf("✅ Worker %d: transação enviada com sucesso | txID: %s", workerID, result.TxID)

	if err := repo.UpdateTransactionHash(tx.ID, result.TxID); err != nil {
		log.Printf("⚠️ Worker %d: erro ao atualizar hash: %v", workerID, err)
	}

	if err := repo.UpdateTransactionStatus(tx.ID, StatusCompleted); err != nil {
		log.Printf("⚠️ Worker %d: erro ao atualizar status para COMPLETED: %v", workerID, err)
	}
}

func handleFailedTransaction(tx d.Transaction, workerID int, db *gorm.DB, repo d.TransactionRepository) {
	if err := repo.UpdateTransactionStatus(tx.ID, StatusFailed); err != nil {
		log.Printf("⚠️ Worker %d: falha ao marcar transação como FAILED: %v", workerID, err)
	}

	err := db.Transaction(func(txDB *gorm.DB) error {
		var balance d.Balance
		if err := txDB.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", tx.UserID).
			First(&balance).Error; err != nil {
			return err
		}

		newBalance := balance.Amount + tx.Amount
		if err := txDB.Model(&d.Balance{}).
			Where("user_id = ?", tx.UserID).
			Update("amount", newBalance).Error; err != nil {
			return err
		}

		refundTx := d.Transaction{
			ID:     uuid.New().String(),
			UserID: tx.UserID,
			Amount: tx.Amount,
			Type:   TypeRefund,
			Status: StatusCompleted,
		}

		return txDB.Create(&refundTx).Error
	})

	if err != nil {
		log.Printf("⚠️ Worker %d: erro ao processar estorno: %v", workerID, err)
	} else {
		log.Printf("✅ Worker %d: estorno concluído para usuário %d", workerID, tx.UserID)
	}
}
