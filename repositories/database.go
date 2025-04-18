package repositories

import (
	d "github.com/financialkafkaconsumerproject/producer/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db}
}

// Garantir que GormRepository implementa as interfaces
var _ d.TransactionRepository = &GormRepository{}
var _ d.BalanceRepository = &GormRepository{}

// Implementa d.TransactionRepository
func (r *GormRepository) Save(tx d.Transaction) error {
	return r.db.Create(&tx).Error
}

func (r *GormRepository) GetByUser(userID string) ([]d.Transaction, error) {
	var txs []d.Transaction
	err := r.db.Where("user_id = ?", userID).Order("timestamp desc").Find(&txs).Error
	return txs, err
}

// Implementa BalanceRepository
func (r *GormRepository) UpdateBalance(tx d.Transaction) error {

	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"amount": gorm.Expr("balances.amount + EXCLUDED.amount"),
		}),
	}).Create(&d.Balance{
		UserID: tx.UserID,
		Amount: tx.Amount,
	}).Error
}

func (r *GormRepository) GetBalance(userID string) (d.Balance, error) {
	var b d.Balance
	err := r.db.Where("user_id = ?", userID).First(&b).Error
	return b, err
}
