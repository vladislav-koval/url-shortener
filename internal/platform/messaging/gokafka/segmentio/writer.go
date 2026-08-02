package segmentio

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"github.com/vladislav-koval/url-shortener/internal/platform/messaging/gokafka"
	"go.uber.org/zap"
)

const dropReportInterval = 30 * time.Second

type Writer struct {
	writer *kafka.Writer
	log    *logger.Logger

	queue         chan kafka.Message
	batchSize     int
	flushInterval time.Duration

	dropped atomic.Uint64
	wg      sync.WaitGroup

	stop chan struct{}

	drainCtx     context.Context
	shutdownOnce sync.Once
	shutdownErr  error
}

type WriterConfig struct {
	Topic     string
	BatchSize int
	// QueueSize — потолок памяти под неотправленные сообщения. Переполнение
	// очереди означает потерю сообщения
	QueueSize int
	// FlushInterval — период принудительной отправки недособранного батча.
	FlushInterval time.Duration
	WriteTimeout  time.Duration
}

func NewWriter(cfg Config, writerCfg WriterConfig, log *logger.Logger) *Writer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        writerCfg.Topic,
		RequiredAcks: kafka.RequireOne,
		Balancer:     &kafka.RoundRobin{},

		BatchTimeout: 10 * time.Millisecond,

		WriteTimeout:           writerCfg.WriteTimeout,
		BatchSize:              writerCfg.BatchSize,
		AllowAutoTopicCreation: false,

		MaxAttempts: 2,

		Logger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			log.Debug(fmt.Sprintf(msg, args...))
		}),

		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			errMsg := fmt.Sprintf(msg, args...)
			log.Error("kafka write error", zap.String("details", errMsg))
		}),
	}

	w := &Writer{
		writer:        writer,
		log:           log,
		queue:         make(chan kafka.Message, writerCfg.QueueSize),
		batchSize:     writerCfg.BatchSize,
		flushInterval: writerCfg.FlushInterval,
		stop:          make(chan struct{}),
	}

	w.wg.Add(2)
	go w.run()
	go w.reportDropped()
	return w
}

func (w *Writer) reportDropped() {
	defer w.wg.Done()

	defer func() {
		if p := recover(); p != nil {
			w.log.RecoverPanic("kafka writer report dropped", p)
		}
	}()

	ticker := time.NewTicker(dropReportInterval)
	defer ticker.Stop()

	var lastReported uint64

	for {
		select {
		case <-ticker.C:
			current := w.dropped.Load()
			if delta := current - lastReported; delta > 0 {
				w.log.Warn("dropped events since last report", zap.Uint64("count", delta))
				lastReported = current
			}

		case <-w.stop:
			return
		}
	}
}

func (w *Writer) run() {
	defer w.wg.Done()

	batch := make([]kafka.Message, 0, w.batchSize)

	defer func() {
		if p := recover(); p != nil {
			w.dropped.Add(uint64(len(batch)))
			w.log.RecoverPanic("kafka writer run loop", p, zap.Int("messages", len(batch)))
		}
	}()

	ticker := time.NewTicker(w.flushInterval)
	defer ticker.Stop()

	flush := func(ctx context.Context) {
		if len(batch) == 0 {
			return
		}

		err := w.writer.WriteMessages(ctx, batch...)
		if err != nil {

			failed := len(batch)

			if writeErrs, ok := errors.AsType[kafka.WriteErrors](err); ok {
				failed = writeErrs.Count() // точное число реально не доставленных, а не весь батч
			}
			w.dropped.Add(uint64(failed))
			w.log.Error("failed to send batch to kafka", zap.Int("messages", len(batch)), zap.Int("failed", failed), zap.Error(err))
		}

		batch = batch[:0]
	}

	for {
		select {
		case message := <-w.queue:
			batch = append(batch, message)

			if len(batch) >= w.batchSize {
				flush(context.Background())
			}

		case <-ticker.C:
			flush(context.Background())

		case <-w.stop:
			for {
				select {
				case message := <-w.queue:
					batch = append(batch, message)

					if len(batch) >= w.batchSize {
						flush(w.drainCtx)
					}
				default:
					flush(w.drainCtx)
					return
				}
			}
		}
	}
}

func (w *Writer) AsyncWriteMessage(message gokafka.Message) bool {
	select {
	case <-w.stop:
		w.dropped.Add(1)
		return false
	default:
	}

	select {
	case w.queue <- kafka.Message{Key: message.Key, Value: message.Value}:
		return true
	default:
		w.dropped.Add(1)
		return false
	}
}

func (w *Writer) Shutdown(ctx context.Context) error {
	w.shutdownOnce.Do(func() {
		w.log.Warn("shutdown kafka writer")

		w.drainCtx = ctx
		close(w.stop)
		w.wg.Wait()
		if dropped := w.dropped.Load(); dropped > 0 {
			w.log.Warn("kafka writer closed with dropped events", zap.Uint64("total_dropped", dropped))
		}
		w.shutdownErr = w.writer.Close()

		if w.shutdownErr == nil {
			w.log.Warn("kafka writer stopped")
		}
	})
	return w.shutdownErr
}
