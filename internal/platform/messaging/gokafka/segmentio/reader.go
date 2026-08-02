package segmentio

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
	"github.com/vladislav-koval/url-shortener/internal/platform/messaging/gokafka"
)

type Reader struct {
	reader *kafka.Reader
}

func NewReader(cfg Config, topic string, groupID string) *Reader {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Brokers,
		GroupID: groupID,
		Topic:   topic,
	})

	return &Reader{
		reader: reader,
	}
}

func (r *Reader) FetchMessage(ctx context.Context) (gokafka.Message, error) {
	km, err := r.reader.FetchMessage(ctx)

	if err != nil {
		return gokafka.Message{}, fmt.Errorf("failed to fetch message: %w", err)
	}

	return gokafka.Message{
		Key:       km.Key,
		Value:     km.Value,
		Topic:     km.Topic,
		Partition: km.Partition,
		Offset:    km.Offset,
	}, nil
}

func (r *Reader) CommitMessages(ctx context.Context, msgs ...gokafka.Message) error {
	kafkaMsgs := make([]kafka.Message, len(msgs))

	for i, m := range msgs {
		kafkaMsgs[i] = kafka.Message{
			Topic:     m.Topic,
			Partition: m.Partition,
			Offset:    m.Offset,
		}
	}

	return r.reader.CommitMessages(ctx, kafkaMsgs...)
}

func (r *Reader) Close() error {
	return r.reader.Close()
}
