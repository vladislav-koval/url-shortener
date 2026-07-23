package consumer

import (
	"context"
	"encoding/json"
	"errors"
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
}

func NewClickConsumer(reader gokafka.Reader, repository ClickRepository, log *logger.Logger, cfg Config) *ClickConsumer {
	return &ClickConsumer{
		reader:       reader,
		repository:   repository,
		log:          log,
		batchSize:    cfg.BatchSize,
		batchTimeout: cfg.BatchTimeout,
	}
}

func (p *ClickConsumer) Start(ctx context.Context) {
	msgChannel := make(chan gokafka.Message, p.batchSize)

	go func() {
		defer close(msgChannel)
		for {
			msg, err := p.reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(ctx.Err(), context.Canceled) {
					return
				}
				p.log.Error("failed to fetch message from kafka", zap.Error(err))
				continue
			}
			msgChannel <- msg
		}
	}()

	batch := make([]events.ClickEvent, 0, p.batchSize)
	kafkaMessages := make([]gokafka.Message, 0, p.batchSize)

	ticker := time.NewTicker(p.batchTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				p.flush(context.Background(), batch, kafkaMessages)
			}
			return

		case <-ticker.C:
			if len(batch) > 0 {
				p.flush(ctx, batch, kafkaMessages)
				batch = batch[:0]
				kafkaMessages = kafkaMessages[:0]
				ticker.Reset(p.batchTimeout)
			}

		case msg, ok := <-msgChannel:
			if !ok {
				if len(batch) > 0 {
					p.flush(context.Background(), batch, kafkaMessages)
				}
				return
			}

			var event events.ClickEvent

			if err := json.Unmarshal(msg.Value, &event); err != nil {
				p.log.Error("failed to deserialize value", zap.Error(err))
				_ = p.reader.CommitMessages(ctx, msg)
				continue
			}

			if err := event.Validate(); err != nil {
				p.log.Error("failed to validate event", zap.Error(err))
				_ = p.reader.CommitMessages(ctx, msg)
				continue
			}

			batch = append(batch, event)
			kafkaMessages = append(kafkaMessages, msg)

			if len(batch) >= p.batchSize {
				p.flush(ctx, batch, kafkaMessages)

				batch = batch[:0]
				kafkaMessages = kafkaMessages[:0]
				ticker.Reset(p.batchTimeout)
			}
		}
	}
}

func (p *ClickConsumer) flush(ctx context.Context, events []events.ClickEvent, msgs []gokafka.Message) {
	if len(events) == 0 {
		return
	}

	var err error

	for attempt := 0; attempt < maxSaveAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(saveRetryBackoff):
			case <-ctx.Done():
				p.log.Warn("giving up retrying batch save: shutting down", zap.Int("count", len(events)))
				return
			}
		}

		err = p.repository.SaveClicks(ctx, events)

		if err == nil || errors.Is(err, apperrors.ErrConflict) {
			break
		}

		p.log.Warn("failed to save batch, retrying",
			zap.Int("attempt", attempt+1),
			zap.Int("count", len(events)),
			zap.Error(err),
		)
	}

	if err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			p.log.Warn("some clicks in batch were duplicates, skipped", zap.Error(err))
		} else {
			p.log.Error("failed to save batch after retries, offsets not committed — will be redelivered by kafka",
				zap.Int("count", len(events)),
				zap.Error(err),
			)
			return
		}
	}

	if err := p.reader.CommitMessages(ctx, msgs...); err != nil {
		p.log.Error("failed to commit batch offsets to kafka", zap.Error(err))
	} else {
		p.log.Debug("successfully flushed batch to postgres", zap.Int("count", len(events)))
	}
}
