package producer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	d "github.com/financialkafkaconsumerproject/producer/domain"
	"github.com/segmentio/kafka-go"
)

var writer *kafka.Writer

func InitProducer(ctx context.Context) {
	writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{"localhost:9092"},
		Topic:        "transacoes",
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: int(kafka.RequireAll),
	})
}

func SendTransaction(tx d.Transaction) error {
	payload, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(tx.UserID),
		Value: payload,
		Time:  time.Now(),
	}

	log.Printf("📤 Enviando transação %s para o Kafka", tx.ID)
	return writer.WriteMessages(context.Background(), msg)
}
