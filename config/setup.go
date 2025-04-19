// setup.go (novo arquivo)
package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gabrielksneiva/go-financial-transactions/api"
	d "github.com/gabrielksneiva/go-financial-transactions/domain"
	"github.com/gabrielksneiva/go-financial-transactions/producer"
	"github.com/gabrielksneiva/go-financial-transactions/repositories"
	"github.com/gabrielksneiva/go-financial-transactions/services"
	s "github.com/gabrielksneiva/go-financial-transactions/services"

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
	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatalf("❌ Erro ao converter REDIS_DB para inteiro: %v", err)
	}
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
		RedisHost:    os.Getenv("REDIS_HOST"),
		RedisDB:      redisDB,
		JwtSecret:    os.Getenv("JWT_SECRET"),
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

	redisClient := repositories.InitRedis(cfg.RedisHost, cfg.RedisDB)
	rateLimiter := repositories.NewRedisRateLimiter(redisClient)

	kafkaWriter := producer.NewKafkaWriter(cfg.KafkaBroker, cfg.KafkaTopic)

	repo := repositories.NewGormRepository(db)
	deposit := s.NewDepositService(repo, repo, kafkaWriter, rateLimiter)
	withdraw := s.NewWithdrawService(repo, repo, kafkaWriter, rateLimiter)
	statement := s.NewStatementService(repo, repo)
	userService := services.NewUserService(repo)

	apiApp := api.NewApp(deposit, withdraw, statement, userService)

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
