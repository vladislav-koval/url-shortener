package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/vladislav-koval/url-shortener/internal/core/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/core/logger"
	"github.com/vladislav-koval/url-shortener/internal/core/messaging/events"
	"github.com/vladislav-koval/url-shortener/internal/core/messaging/gokafka"
	"go.uber.org/zap"
)

type ClickRepository interface {
	SaveClicks(ctx context.Context, events []events.ClickEvent) error
}

const (
	maxSaveAttempts  = 3
	saveRetryBackoff = 500 * time.Millisecond
)

type ClickConsumer struct {
	reader       gokafka.Reader
	repository   ClickRepository
	log          *logger.Logger
	batchSize    int
	batchTimeout time.Duration

	fetchCtx    context.Context
	cancelFetch context.CancelFunc

	shutdownCtx  context.Context
	shutdownOnce sync.Once
}

func NewClickConsumer(
	reader gokafka.Reader,
	repository ClickRepository,
	log *logger.Logger,
	cfg Config,
) *ClickConsumer {
	fetchCtx, cancelFetch := context.WithCancel(context.Background())

	return &ClickConsumer{
		reader:       reader,
		repository:   repository,
		log:          log,
		batchSize:    cfg.BatchSize,
		batchTimeout: cfg.BatchTimeout,
		fetchCtx:     fetchCtx,
		cancelFetch:  cancelFetch,
		shutdownCtx:  context.Background(),
	}
}

func (p *ClickConsumer) Run() error {
	msgChannel := make(chan gokafka.Message, p.batchSize)

	var fetching sync.WaitGroup
	fetching.Add(1)

	go func() {
		defer fetching.Done()
		defer close(msgChannel)

		for {
			msg, err := p.reader.FetchMessage(p.fetchCtx)
			if err != nil {
				if p.fetchCtx.Err() != nil {
					return
				}

				p.log.Error(
					"failed to fetch message from kafka",
					zap.Error(err),
				)
				continue
			}

			select {
			case msgChannel <- msg:
			case <-p.fetchCtx.Done():
				return
			}
		}
	}()

	defer fetching.Wait()

	batch := make([]events.ClickEvent, 0, p.batchSize)
	kafkaMessages := make([]gokafka.Message, 0, p.batchSize)

	ticker := time.NewTicker(p.batchTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if len(batch) == 0 {
				continue
			}

			p.flush(p.currentOperationCtx(), batch, kafkaMessages)

			batch = batch[:0]
			kafkaMessages = kafkaMessages[:0]

		case msg, ok := <-msgChannel:
			if !ok {
				p.shutdownFlush(batch, kafkaMessages)
				p.log.Warn("analytics consumer stopped")
				return nil
			}

			var event events.ClickEvent

			if err := json.Unmarshal(msg.Value, &event); err != nil {
				p.log.Error("failed to deserialize value", zap.Error(err))
				_ = p.reader.CommitMessages(p.currentOperationCtx(), msg)
				continue
			}

			if err := event.Validate(); err != nil {
				p.log.Error("failed to validate event", zap.Error(err))
				_ = p.reader.CommitMessages(p.currentOperationCtx(), msg)
				continue
			}

			batch = append(batch, event)
			kafkaMessages = append(kafkaMessages, msg)

			if len(batch) < p.batchSize {
				continue
			}

			p.flush(p.currentOperationCtx(), batch, kafkaMessages)

			batch = batch[:0]
			kafkaMessages = kafkaMessages[:0]
			ticker.Reset(p.batchTimeout)
		}
	}
}

func (p *ClickConsumer) Shutdown(ctx context.Context) {
	p.shutdownOnce.Do(func() {
		p.log.Warn("shutdown analytics consumer")

		p.shutdownCtx = ctx
		p.cancelFetch()
	})
}

func (p *ClickConsumer) currentOperationCtx() context.Context {
	if p.fetchCtx.Err() == nil {
		return p.fetchCtx
	}

	return p.shutdownCtx
}

func (p *ClickConsumer) shutdownFlush(
	clickEvents []events.ClickEvent,
	messages []gokafka.Message,
) {
	if len(clickEvents) == 0 {
		return
	}

	p.flush(p.shutdownCtx, clickEvents, messages)
}

func (p *ClickConsumer) flush(
	ctx context.Context,
	clickEvents []events.ClickEvent,
	messages []gokafka.Message,
) {
	if len(clickEvents) == 0 {
		return
	}

	var err error

	for attempt := 0; attempt < maxSaveAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(saveRetryBackoff):
			case <-ctx.Done():
				p.log.Warn("giving up retrying batch save: shutting down", zap.Int("count", len(clickEvents)))
				return
			}
		}

		err = p.repository.SaveClicks(ctx, clickEvents)

		if err == nil || errors.Is(err, apperrors.ErrConflict) {
			break
		}

		p.log.Warn(
			"failed to save batch, retrying",
			zap.Int("attempt", attempt+1),
			zap.Int("count", len(clickEvents)),
			zap.Error(err),
		)
	}

	if err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			p.log.Warn(
				"some clicks in batch were duplicates, skipped",
				zap.Error(err),
			)
		} else {
			p.log.Error(
				"failed to save batch after retries, offsets not committed — will be redelivered by kafka",
				zap.Int("count", len(clickEvents)),
				zap.Error(err),
			)
			return
		}
	}

	if err := p.reader.CommitMessages(ctx, messages...); err != nil {
		p.log.Error(
			"failed to commit batch offsets to kafka",
			zap.Error(err),
		)
		return
	}

	p.log.Debug(
		"successfully flushed batch to postgres",
		zap.Int("count", len(clickEvents)),
	)
}
