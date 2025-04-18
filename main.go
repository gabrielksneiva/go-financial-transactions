package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/financialkafkaconsumerproject/producer/api"
	"github.com/financialkafkaconsumerproject/producer/consumer"
	d "github.com/financialkafkaconsumerproject/producer/domain"
	"github.com/financialkafkaconsumerproject/producer/producer"
	"github.com/financialkafkaconsumerproject/producer/repositories"
	s "github.com/financialkafkaconsumerproject/producer/services"
	"github.com/financialkafkaconsumerproject/producer/workers"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🚀 Starting application...")

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

	// Repositories and services
	repo := repositories.NewGormRepository(db)
	depositService := s.NewDepositService(repo, repo)
	withdrawService := s.NewWithdrawService(repo, repo)
	statementService := s.NewStatementService(repo, repo)

	// Create and start Fiber app
	apiApp := api.NewApp(depositService, withdrawService, statementService)

	go func() {
		if err := apiApp.Fiber.Listen(":8080"); err != nil {
			fmt.Printf("⚠️ Failed to start API server: %v\n", err)
		}
	}()

	// Graceful shutdown setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	done := make(chan struct{})
	transactions := make(chan d.Transaction, 100)

	// Start Kafka producer, consumer, and worker pool
	go producer.InitProducer(ctx, kafkaBroker, kafkaTopic, kafkaGroupID)
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
