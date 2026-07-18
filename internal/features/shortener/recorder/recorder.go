package recorder

import (
	"context"
	"encoding/json"
	"time"

	"github.com/vladislav-koval/url-shortener/internal/core/logger"
	"github.com/vladislav-koval/url-shortener/internal/core/messaging/events"
	"github.com/vladislav-koval/url-shortener/internal/core/messaging/kafka"
	"go.uber.org/zap"
)

const writeTimeout = 3 * time.Second

type Recorder struct {
	writer kafka.Writer
}

func NewRecorder(writer kafka.Writer) *Recorder {
	return &Recorder{writer: writer}
}

func (r *Recorder) RecordClick(ctx context.Context, clickEvent events.ClickEvent) {
	log := logger.FromContext(ctx)

	value, err := json.Marshal(clickEvent)
	if err != nil {
		log.Error("marshal click event", zap.Error(err))
		return
	}

	writeCtx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	if err := r.writer.WriteMessage(writeCtx, []byte(clickEvent.ShortCode), value); err != nil {
		log.Error("failed to write click event", zap.Error(err))
	}
}
