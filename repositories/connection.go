package repositories

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDatabase() *gorm.DB {
	dsn := "host=localhost user=user password=password dbname=finance port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&Transaction{}, &Balance{}); err != nil {
		log.Fatalf("❌ Failed to migrate models: %v", err)
	}

	fmt.Println("✅ Database connected and migrated.")
	return db
}
