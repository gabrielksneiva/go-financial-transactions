package repositories_test

import (
	"errors"
	"testing"
	"time"

	"github.com/gabrielksneiva/go-financial-transactions/domain"
	"github.com/gabrielksneiva/go-financial-transactions/repositories"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&domain.Transaction{}, &domain.Balance{})
	assert.NoError(t, err)

	return db
}

func TestGormRepository_SaveAndGetByUser(t *testing.T) {
	db := setupTestDB(t)
	repo := repositories.NewGormRepository(db)

	tx := domain.Transaction{
		ID:        "tx-1",
		UserID:    123,
		Amount:    200.0,
		Timestamp: time.Now(),
		Type:      domain.DepositTransaction,
	}

	err := repo.Save(tx)
	assert.NoError(t, err)

	txs, err := repo.GetByUser(123)
	assert.NoError(t, err)
	assert.Len(t, txs, 1)
	assert.Equal(t, tx.ID, txs[0].ID)
	assert.Equal(t, tx.Amount, txs[0].Amount)
	assert.Equal(t, tx.Type, txs[0].Type)
}

func TestGormRepository_GetByUser_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := repositories.NewGormRepository(db)

	txs, err := repo.GetByUser(999)
	assert.NoError(t, err)
	assert.Len(t, txs, 0)
}

func TestGormRepository_UpdateBalance_InsertNewUser(t *testing.T) {
	db := setupTestDB(t)
	repo := repositories.NewGormRepository(db)

	tx := domain.Transaction{
		UserID: 1,
		Amount: 100.0,
	}

	err := repo.UpdateBalance(tx)
	assert.NoError(t, err)

	balance, err := repo.GetBalance(1)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, balance.Amount)
}

func TestGormRepository_UpdateBalance_ExistingUser(t *testing.T) {
	db := setupTestDB(t)
	repo := repositories.NewGormRepository(db)

	initial := domain.Transaction{
		UserID: 2,
		Amount: 100.0,
	}
	err := repo.UpdateBalance(initial)
	assert.NoError(t, err)

	additional := domain.Transaction{
		UserID: 2,
		Amount: 50.0,
	}
	err = repo.UpdateBalance(additional)
	assert.NoError(t, err)

	balance, err := repo.GetBalance(2)
	assert.NoError(t, err)
	assert.Equal(t, 150.0, balance.Amount)
}

func TestGormRepository_GetBalance_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := repositories.NewGormRepository(db)

	_, err := repo.GetBalance(999)
	assert.Error(t, err)
}

func TestGormRepository_Save_InvalidTransaction(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	repo := repositories.NewGormRepository(db)

	tx := domain.Transaction{
		ID:     "invalid",
		UserID: 999,
		Amount: 100,
		Type:   "deposit",
	}

	err := repo.Save(tx)
	assert.Error(t, err)
}

func TestGormRepository_Save_TableMissing(t *testing.T) {
	db := setupTestDB(t)
	repo := repositories.NewGormRepository(db)

	_ = db.Migrator().DropTable(&domain.Transaction{})
	tx := domain.Transaction{
		ID:        "invalid",
		UserID:    999,
		Amount:    100.0,
		Type:      "deposit",
		Timestamp: time.Now(),
	}
	err := repo.Save(tx)
	assert.Error(t, err)
}

func TestGormRepository_GetByUser_Error(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	repo := repositories.NewGormRepository(db)

	_, err := repo.GetByUser(1)
	assert.Error(t, err)
}

func TestGormRepository_UpdateBalance_TableMissing(t *testing.T) {
	db := setupTestDB(t)
	repo := repositories.NewGormRepository(db)
	_ = db.Migrator().DropTable(&domain.Balance{})
	tx := domain.Transaction{
		UserID: 999,
		Amount: 100.0,
	}
	err := repo.UpdateBalance(tx)
	assert.Error(t, err)
}

func TestGormRepository_UpdateBalance_ConflictUpdate(t *testing.T) {
	db := setupTestDB(t)
	repo := repositories.NewGormRepository(db)
	tx1 := domain.Transaction{
		UserID: 777,
		Amount: 100.0,
	}
	err := repo.UpdateBalance(tx1)
	assert.NoError(t, err)
	tx2 := domain.Transaction{
		UserID: 777,
		Amount: 50.0,
	}
	err = repo.UpdateBalance(tx2)
	assert.NoError(t, err)
	balance, err := repo.GetBalance(777)
	assert.NoError(t, err)
	assert.Equal(t, 150.0, balance.Amount)
}

func TestMigrate_Error_WithMock(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer sqlDB.Close()

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	assert.NoError(t, err)

	mock.ExpectQuery("SELECT (.+) FROM information_schema.tables").WillReturnError(errors.New("migration failed"))

	err = repositories.CallMigrate(db)
	assert.Error(t, err)
}
