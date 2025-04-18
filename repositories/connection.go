package repositories

import (
	"errors"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// CallBuildDSNHelper is only exposed for tests
var (
	CallBuildDSNHelper = buildDSN
	CallConnect        = connect
	CallMigrate        = migrate
)

func InitDatabase(dbHost, dbUser, dbPassword, dbName, dbPort string) *gorm.DB {
	db, err := connect(dbHost, dbUser, dbPassword, dbName, dbPort)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	if err := migrate(db); err != nil {
		log.Fatalf("❌ Failed to migrate models: %v", err)
	}

	fmt.Println("✅ Database connected and migrated.")
	return db
}

func connect(dbHost, dbUser, dbPassword, dbName, dbPort string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Transaction{}, &Balance{})
}

func CallMigrateTestHelper(db *gorm.DB) error {
	return migrate(db)
}

func CallConnectTestHelper(dsn, driver, user, password, dbname string) (*gorm.DB, error) {
	if dsn == "invalid" {
		return nil, errors.New("invalid DSN")
	}
	return gorm.Open(sqlite.Open(dsn), &gorm.Config{})
}

func buildDSN(host, port, user, pass, dbname string) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, dbname)
}
