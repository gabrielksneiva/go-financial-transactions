package repositories

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDatabase(dbHost, dbUser, dbPassword, dbName, dbPort string) *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)
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
