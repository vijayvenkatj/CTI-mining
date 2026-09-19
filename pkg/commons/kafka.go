package commons

import (
	"context"
	"strings"

	"github.com/segmentio/kafka-go"
)

type KafkaReader struct {
	reader *kafka.Reader
}

func NewKafkaReader(ctx context.Context, brokers []string, topic, groupID string) (*KafkaReader, error) {
	if err := EnsureTopics(ctx, brokers[0], topic); err != nil {
		return nil, err
	}

	return &KafkaReader{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
	}, nil
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

func NewKafkaWriter(ctx context.Context, brokers []string, topic string) (*KafkaWriter, error) {
	if err := EnsureTopics(ctx, brokers[0], topic); err != nil {
		return nil, err
	}

	return &KafkaWriter{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}, nil
}

func (k *KafkaWriter) WriteMessage(ctx context.Context, key, value []byte) error {
	return k.writer.WriteMessages(ctx, kafka.Message{Key: key, Value: value})
}

func (k *KafkaWriter) Close() error {
	return k.writer.Close()
}

func EnsureTopics(ctx context.Context, broker string, topics ...string) error {
	conn, err := kafka.DialContext(ctx, "tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()

	for _, topic := range topics {
		err := conn.CreateTopics(kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "topic already exists") {
			return err
		}
	}

	return nil
}
