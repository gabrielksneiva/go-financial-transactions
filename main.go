package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go-financial-transactions/api"
	"go-financial-transactions/consumer"
	d "go-financial-transactions/domain"
	"go-financial-transactions/producer"
	"go-financial-transactions/repositories"
	s "go-financial-transactions/services"
	"go-financial-transactions/workers"

	"github.com/joho/godotenv"
)

func Run() {
	fmt.Println("🚀 Starting application...")

	// Graceful shutdown setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️ Error loading .env file")
	}

	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "8080"
	}
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	kafkaGroupID := os.Getenv("KAFKA_GROUP_ID")

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Initialize database
	db := repositories.InitDatabase(dbHost, dbUser, dbPassword, dbName, dbPort)
	if db == nil {
		panic("Database connection error")
	}

	// Initialize Kafka writer (producer)
	kafkaWriter := producer.NewKafkaWriter(kafkaBroker, kafkaTopic)
	defer kafkaWriter.Close()

	// Repositories and services
	repo := repositories.NewGormRepository(db)
	depositService := s.NewDepositService(repo, repo, kafkaWriter)
	withdrawService := s.NewWithdrawService(repo, repo, kafkaWriter)
	statementService := s.NewStatementService(repo, repo)

	// Create and start Fiber app
	apiApp := api.NewApp(depositService, withdrawService, statementService)

	go func() {
		if err := apiApp.Fiber.Listen(":" + apiPort); err != nil {
			fmt.Printf("⚠️ Failed to start API server: %v\n", err)
		}
	}()

	// Signal handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	done := make(chan struct{})
	transactions := make(chan d.Transaction, 100)

	// Start Kafka consumer and worker pool
	go consumer.InitConsumer(ctx, transactions, kafkaBroker, kafkaTopic, kafkaGroupID)
	go workers.StartWorkers(ctx, transactions, 4, db)

	// Shutdown handler
	go func() {
		<-sigs
		fmt.Println("\n🛑 Interrupt signal received, shutting down...")
		cancel()
		close(done)
	}()

	<-done
	fmt.Println("✔️ Gracefully shut down.")
}

func main() {
	Run()
}
