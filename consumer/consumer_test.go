package consumer_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/gabrielksneiva/go-financial-transactions/consumer"
	"github.com/gabrielksneiva/go-financial-transactions/domain"
	"github.com/gabrielksneiva/go-financial-transactions/mocks"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInitConsumerWithReader_Success(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan domain.Transaction, 1)

	tx := domain.Transaction{
		ID:     "tx-123",
		UserID: "user-456",
		Amount: 200.0,
		Type:   "deposit",
	}

	data, _ := json.Marshal(tx)

	readerMock := new(mocks.KafkaReader)
	var wg sync.WaitGroup
	wg.Add(1)

	readerMock.On("ReadMessage", mock.Anything).Once().Return(kafka.Message{Value: data}, nil)
	readerMock.On("ReadMessage", mock.Anything).Return(kafka.Message{}, context.Canceled)
	readerMock.On("Close").Return(nil)

	go func() {
		defer wg.Done()
		consumer.InitConsumerWithReader(ctx, ch, readerMock)
	}()

	select {
	case received := <-ch:
		assert.Equal(t, tx.ID, received.ID)
		assert.Equal(t, tx.UserID, received.UserID)
		assert.Equal(t, tx.Amount, received.Amount)
		cancel()
	case <-time.After(time.Second):
		t.Fatal("timeout esperando transação")
	}

	wg.Wait()
	readerMock.AssertExpectations(t)
}

func TestInitConsumerWithReader_InvalidJSON(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan domain.Transaction, 1)
	readerMock := new(mocks.KafkaReader)
	var wg sync.WaitGroup
	wg.Add(1)

	readerMock.On("ReadMessage", mock.Anything).Once().Return(kafka.Message{Value: []byte("invalid-json")}, nil)
	readerMock.On("ReadMessage", mock.Anything).Return(kafka.Message{}, context.Canceled)
	readerMock.On("Close").Return(nil)

	go func() {
		defer wg.Done()
		consumer.InitConsumerWithReader(ctx, ch, readerMock)
	}()

	time.Sleep(200 * time.Millisecond)
	cancel()
	wg.Wait()

	// canal não deve receber nada, pois o JSON era inválido
	select {
	case <-ch:
		t.Fatal("esperava canal vazio")
	default:
	}

	readerMock.AssertExpectations(t)
}

func TestInitConsumerWithReader_ReadError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan domain.Transaction, 1)
	readerMock := new(mocks.KafkaReader)
	var wg sync.WaitGroup
	wg.Add(1)

	readerMock.On("ReadMessage", mock.Anything).Once().Return(kafka.Message{}, errors.New("read error"))
	readerMock.On("ReadMessage", mock.Anything).Return(kafka.Message{}, context.Canceled)
	readerMock.On("Close").Return(nil)

	go func() {
		defer wg.Done()
		consumer.InitConsumerWithReader(ctx, ch, readerMock)
	}()

	time.Sleep(200 * time.Millisecond)
	cancel()
	wg.Wait()

	readerMock.AssertExpectations(t)
}
