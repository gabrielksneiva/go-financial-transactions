// consumer/reader.go
package consumer

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type KafkaReader interface {
	ReadMessage(context.Context) (kafka.Message, error)
	Close() error
}
