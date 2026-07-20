package recorder

import (
	"context"
	"encoding/json"

	"github.com/vladislav-koval/url-shortener/internal/core/logger"
	"github.com/vladislav-koval/url-shortener/internal/core/messaging/events"
	"github.com/vladislav-koval/url-shortener/internal/core/messaging/gokafka"
	"go.uber.org/zap"
)

type Recorder struct {
	writer gokafka.Writer
}

func NewRecorder(writer gokafka.Writer) *Recorder {
	return &Recorder{writer: writer}
}

func (r *Recorder) RecordClick(ctx context.Context, clickEvent events.ClickEvent) {
	log := logger.FromContext(ctx)

	value, err := json.Marshal(clickEvent)
	if err != nil {
		log.Error("marshal click event", zap.Error(err))
		return
	}

	if err := r.writer.WriteMessage(ctx, []byte(clickEvent.ShortCode), value); err != nil {
		log.Error("failed to write click event", zap.Error(err))
	}
}
