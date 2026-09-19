package commons

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type KafkaReader struct {
	reader *kafka.Reader
}

func NewKafkaReader(brokers []string, topic, groupID string) *KafkaReader {
	return &KafkaReader{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
	}
}

func (k *KafkaReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return k.reader.ReadMessage(ctx)
}

func (k *KafkaReader) Close() error {
	return k.reader.Close()
}

type KafkaWriter struct {
	writer *kafka.Writer
}

func NewKafkaWriter(brokers []string, topic string) *KafkaWriter {
	return &KafkaWriter{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (k *KafkaWriter) WriteMessage(ctx context.Context, key, value []byte) error {
	return k.writer.WriteMessages(ctx, kafka.Message{Key: key, Value: value})
}

func (k *KafkaWriter) Close() error {
	return k.writer.Close()
}
