// producer/writer.go
package producer

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type WriterInterface interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}
