package segmentio

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/vladislav-koval/url-shortener/internal/core/logger"
	"go.uber.org/zap"
)

type Writer struct {
	writer *kafka.Writer
}

func NewWriter(
	cfg Config,
	topic string,
	batchSize int,
	batchTimeout time.Duration,
	log *logger.Logger,
) *Writer {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Topic:                  topic,
		RequiredAcks:           kafka.RequireOne,
		Balancer:               &kafka.RoundRobin{},
		BatchTimeout:           batchTimeout,
		BatchSize:              batchSize,
		AllowAutoTopicCreation: false,

		/*
			Асинхронный режим, WriteMessages ставит сообщение в очередь и мгновенно возвращает управление
			не дожидаясь отдачи пакета брокеку.
			Проблема может быть в том, что очередь сообщений безгранична и при сбое брокера очередь может заполнится
			и вызвать out of memory.
		*/
		Async: true,

		Completion: func(messages []kafka.Message, err error) {
			if err != nil {
				log.Error("kafka async write failed", zap.Int("messages", len(messages)), zap.Error(err))
			}
		},

		Logger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			log.Debug(fmt.Sprintf(msg, args...))
		}),

		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			errMsg := fmt.Sprintf(msg, args...)
			log.Error("kafka write error", zap.String("details", errMsg))
		}),
	}

	return &Writer{
		writer: writer,
	}
}

func (w *Writer) WriteMessage(ctx context.Context, key, value []byte) error {
	if err := w.writer.WriteMessages(ctx, kafka.Message{Key: key, Value: value}); err != nil {
		return fmt.Errorf("write kafka message: %w", err)
	}

	return nil
}

func (w *Writer) Close() error {
	return w.writer.Close()
}
