// main.go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gabrielksneiva/go-financial-transactions/client"
	"github.com/gabrielksneiva/go-financial-transactions/config"
	"github.com/gabrielksneiva/go-financial-transactions/consumer"
	"github.com/gabrielksneiva/go-financial-transactions/domain"
	"github.com/gabrielksneiva/go-financial-transactions/repositories"
	"github.com/gabrielksneiva/go-financial-transactions/workers"
)

func RunApp() {
	fmt.Println("🚀 Starting application...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.LoadConfig()
	app := config.SetupApplication()

	defer app.KafkaWriter.Close()

	// Start HTTP server
	go func() {
		if err := app.API.Fiber.Listen(":" + app.Config.APIPort); err != nil {
			fmt.Printf("⚠️ Failed to start API server: %v\n", err)
		}
	}()

	// Setup signal handler
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	done := make(chan struct{})
	transactions := make(chan domain.Transaction, 100)

	tronClient := client.NewTronClient()

	// poller.StartPollingTRXTransactions(ctx, cfg.TronWallet, 30*time.Second, app.DB)
	repo := repositories.NewGormRepository(app.DB)

	// Start consumer & workers
	go consumer.InitConsumer(ctx, transactions, cfg.KafkaBroker, cfg.KafkaTopic, cfg.KafkaGroupID)
	go workers.Worker(ctx, 4, transactions, app.DB, tronClient, repo)

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
	RunApp()
	fmt.Println("✔️ Application exited.")
}
