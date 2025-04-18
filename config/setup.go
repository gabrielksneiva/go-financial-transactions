// setup.go (novo arquivo)
package config

import (
	"context"
	"fmt"
	"go-financial-transactions/api"
	d "go-financial-transactions/domain"
	"go-financial-transactions/producer"
	"go-financial-transactions/repositories"
	s "go-financial-transactions/services"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type AppResources struct {
	DB            *gorm.DB
	KafkaWriter   *producer.KafkaWriter
	API           *api.App
	TransactionCh chan d.Transaction
	CancelFunc    context.CancelFunc
	Context       context.Context
	Config        Config
}

func LoadConfig() Config {
	_ = godotenv.Load()
	return Config{
		APIPort:      getEnv("API_PORT", "8080"),
		KafkaBroker:  os.Getenv("KAFKA_BROKER"),
		KafkaTopic:   os.Getenv("KAFKA_TOPIC"),
		KafkaGroupID: os.Getenv("KAFKA_GROUP_ID"),
		DBHost:       os.Getenv("DB_HOST"),
		DBPort:       os.Getenv("DB_PORT"),
		DBUser:       os.Getenv("DB_USER"),
		DBPassword:   os.Getenv("DB_PASSWORD"),
		DBName:       os.Getenv("DB_NAME"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func SetupApplication() *AppResources {
	fmt.Println("🚀 Initializing dependencies...")

	ctx, cancel := context.WithCancel(context.Background())
	cfg := LoadConfig()

	db := repositories.InitDatabase(cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)
	if db == nil {
		panic("❌ Failed to connect to database")
	}

	kafkaWriter := producer.NewKafkaWriter(cfg.KafkaBroker, cfg.KafkaTopic)

	repo := repositories.NewGormRepository(db)
	deposit := s.NewDepositService(repo, repo, kafkaWriter)
	withdraw := s.NewWithdrawService(repo, repo, kafkaWriter)
	statement := s.NewStatementService(repo, repo)

	apiApp := api.NewApp(deposit, withdraw, statement)

	transactions := make(chan d.Transaction, 100)

	return &AppResources{
		DB:            db,
		KafkaWriter:   kafkaWriter,
		API:           apiApp,
		TransactionCh: transactions,
		Context:       ctx,
		CancelFunc:    cancel,
		Config:        cfg,
	}
}
