package segmentio

import "github.com/segmentio/kafka-go"

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
