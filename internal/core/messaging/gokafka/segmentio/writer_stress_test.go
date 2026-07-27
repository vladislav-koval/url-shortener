package segmentio

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/vladislav-koval/url-shortener/internal/core/logger"
	"github.com/vladislav-koval/url-shortener/internal/core/messaging/gokafka"
	"go.uber.org/zap"
)

// Эмпирическая проверка: конкурентные AsyncWriteMessage во время Shutdown() не
// должны паниковать ни разу, независимо от того, успевает ли реальный брокер
// принять сообщения (брокера тут нет вообще, адрес заведомо недоступен).
func TestWriter_NoPanicOnConcurrentShutdownAndSend(t *testing.T) {
	for attempt := 0; attempt < 50; attempt++ {
		w := NewWriter(
			Config{Brokers: []string{"10.255.255.1:9092"}},
			WriterConfig{
				Topic:         "test-topic",
				BatchSize:     10,
				QueueSize:     5,
				FlushInterval: 10 * time.Millisecond,
				WriteTimeout:  100 * time.Millisecond,
			},
			&logger.Logger{Logger: zap.NewNop()},
		)

		var wg sync.WaitGroup
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("AsyncWriteMessage panicked: %v", r)
					}
				}()
				w.AsyncWriteMessage(gokafka.Message{Key: []byte("k"), Value: []byte("v")})
			}()
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()
			_ = w.Shutdown(ctx)
		}()

		wg.Wait()
	}
}
