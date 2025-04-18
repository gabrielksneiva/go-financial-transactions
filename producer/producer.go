// producer.go
package producer

import (
	"context"
	"encoding/json"
	"go-financial-transactions/domain"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type Producer interface {
	SendTransaction(tx domain.Transaction) error
	Close() error
}

type KafkaWriter struct {
	writer WriterInterface
}

func NewKafkaWriter(broker, topic string) *KafkaWriter {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{broker},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})

	return &KafkaWriter{
		writer: w, // agora *kafka.Writer implementa WriterInterface corretamente
	}
}

func (k *KafkaWriter) SendTransaction(tx domain.Transaction) error {
	data, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(uuid.New().String()),
		Value: data,
	}

	return k.writer.WriteMessages(context.Background(), msg)
}

func (k *KafkaWriter) Close() error {
	return k.writer.Close()
}
