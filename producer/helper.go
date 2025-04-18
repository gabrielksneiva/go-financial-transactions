package producer

// NewKafkaWriterWithMock is used only in unit tests to inject a mock writer
func NewKafkaWriterWithMock(w WriterInterface) *KafkaWriter {
	return &KafkaWriter{writer: w}
}
