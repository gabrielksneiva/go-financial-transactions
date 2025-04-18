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
)

func main() {
	fmt.Println("🚀 Starting application...")

	// Initialize database
	db := repositories.InitDatabase()

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
	go producer.InitProducer(ctx)
	go consumer.InitConsumer(ctx, transactions)
	go workers.StartWorkers(ctx, transactions, 4, repo, repo)

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
