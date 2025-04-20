// main.go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gabrielksneiva/go-financial-transactions/client"
	"github.com/gabrielksneiva/go-financial-transactions/config"
	"github.com/gabrielksneiva/go-financial-transactions/consumer"
	"github.com/gabrielksneiva/go-financial-transactions/domain"
	"github.com/gabrielksneiva/go-financial-transactions/frontend"
	"github.com/gabrielksneiva/go-financial-transactions/repositories"
	"github.com/gabrielksneiva/go-financial-transactions/workers"
	"github.com/gofiber/fiber/v2"
)

func RunApp() {
	fmt.Println("🚀 Starting application...")

	// 1) Carrega config e app (API já vem com app.API.Fiber)
	cfg := config.LoadConfig()
	app := config.SetupApplication()

	// 2) Cria o servidor FE manualmente (para podermos chamar Shutdown depois)
	feApp := fiber.New()
	frontend.SetupRoutes(feApp)

	// 3) Canal de sinais + contexto de cancelamento
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 4) WaitGroup para aguardar Listen() retornarem
	var wg sync.WaitGroup
	wg.Add(2)

	// 5) Start API
	go func() {
		defer wg.Done()
		if err := app.API.Fiber.Listen(":" + app.Config.APIPort); err != nil {
			fmt.Printf("⚠️ Failed to start API server: %v\n", err)
		}
	}()

	// 6) Start Front-end
	go func() {
		defer wg.Done()
		if err := feApp.Listen(":" + cfg.FrontendPort); err != nil {
			fmt.Printf("⚠️ Failed to start Front-end server: %v\n", err)
		}
	}()

	// 7) Start consumer & workers
	transactions := make(chan domain.Transaction, 100)
	go consumer.InitConsumer(ctx, transactions, cfg.KafkaBroker, cfg.KafkaTopic, cfg.KafkaGroupID)
	tronClient := client.NewTronClient()
	repo := repositories.NewGormRepository(app.DB)
	go workers.Worker(ctx, 4, transactions, app.DB, tronClient, repo)

	// 8) Aguarda sinal de interrupção
	<-quit
	fmt.Println("\n🛑 Interrupt signal received, shutting down...")

	// 9) Cancela contexto (para parar consumer & workers)
	cancel()

	// 10) Chama Shutdown para ambos os Fiber apps (com timeout)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := app.API.Fiber.ShutdownWithContext(shutdownCtx); err != nil {
		fmt.Printf("❌ Erro ao encerrar API: %v\n", err)
	}
	if err := feApp.ShutdownWithContext(shutdownCtx); err != nil {
		fmt.Printf("❌ Erro ao encerrar Front-end: %v\n", err)
	}

	// 11) Fecha conexões de Kafka, DB, etc.
	app.KafkaWriter.Close()

	// 12) Aguarda Listen() retornarem
	wg.Wait()
	fmt.Println("✔️ Gracefully shut down.")
}

func main() {
	RunApp()
	fmt.Println("✔️ Application exited.")
}
