// main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gabrielksneiva/go-financial-transactions/config"
	"github.com/gabrielksneiva/go-financial-transactions/consumer"
	"github.com/gabrielksneiva/go-financial-transactions/domain"
	"github.com/gabrielksneiva/go-financial-transactions/workers"
	"github.com/joho/godotenv"
)

func RunApp() {
	fmt.Println("🚀 Starting application...")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar o arquivo .env")
	}

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

	// Start consumer & workers
	go consumer.InitConsumer(ctx, transactions, cfg.KafkaBroker, cfg.KafkaTopic, cfg.KafkaGroupID)
	go workers.StartWorkers(ctx, transactions, 4, app.DB)

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
