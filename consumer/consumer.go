package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	d "github.com/gabrielksneiva/go-financial-transactions/domain"

	"github.com/segmentio/kafka-go"
)

// consumer/consumer.go
func InitConsumerWithReader(ctx context.Context, ch chan<- d.Transaction, reader KafkaReader) {
	defer reader.Close()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("📥 Consumer encerrado.")
			return
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Erro ao ler mensagem: %v", err)
				continue
			}

			var tx d.Transaction
			if err := json.Unmarshal(msg.Value, &tx); err != nil {
				log.Printf("Erro ao deserializar JSON: %v", err)
				continue
			}

			ch <- tx
		}
	}
}

// consumer/consumer.go
func InitConsumer(ctx context.Context, ch chan<- d.Transaction, kafkaBroker, kafkaTopic, kafkaGroupID string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaBroker},
		Topic:   kafkaTopic,
		GroupID: kafkaGroupID,
	})
	InitConsumerWithReader(ctx, ch, reader)
}
