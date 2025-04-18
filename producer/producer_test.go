package producer_test

import (
	"errors"
	"testing"

	"go-financial-transactions/domain"
	"go-financial-transactions/mocks"
	"go-financial-transactions/producer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestKafkaProducer_SendTransaction(t *testing.T) {
	writerMock := new(mocks.WriterInterface)
	writerMock.
		On("WriteMessages", mock.Anything, mock.Anything).
		Return(nil)

	prod := producer.NewKafkaWriterWithMock(writerMock)

	tx := domain.Transaction{
		ID:     "tx-123",
		UserID: "user-456",
		Amount: 100,
		Type:   "deposit",
	}

	err := prod.SendTransaction(tx)
	assert.NoError(t, err)
	writerMock.AssertExpectations(t)
}

func TestKafkaProducer_SendTransaction_Error(t *testing.T) {
	writerMock := new(mocks.WriterInterface)
	writerMock.
		On("WriteMessages", mock.Anything, mock.Anything).
		Return(errors.New("fail"))

	prod := producer.NewKafkaWriterWithMock(writerMock)

	tx := domain.Transaction{
		ID:     "tx-123",
		UserID: "user-456",
		Amount: 100,
		Type:   "deposit",
	}

	err := prod.SendTransaction(tx)
	assert.Error(t, err)
	assert.EqualError(t, err, "fail")
	writerMock.AssertExpectations(t)
}

func TestKafkaWriter_Close(t *testing.T) {
	writerMock := new(mocks.WriterInterface)
	writerMock.On("Close").Return(nil)

	prod := producer.NewKafkaWriterWithMock(writerMock)
	err := prod.Close()

	assert.NoError(t, err)
	writerMock.AssertExpectations(t)
}

func TestKafkaWriter_Close_Error(t *testing.T) {
	writerMock := new(mocks.WriterInterface)
	writerMock.On("Close").Return(errors.New("close fail"))

	prod := producer.NewKafkaWriterWithMock(writerMock)
	err := prod.Close()

	assert.Error(t, err)
	assert.EqualError(t, err, "close fail")
	writerMock.AssertExpectations(t)
}

func TestNewKafkaWriter(t *testing.T) {
	prod := producer.NewKafkaWriter("localhost:9092", "transactions")
	assert.NotNil(t, prod)
}
