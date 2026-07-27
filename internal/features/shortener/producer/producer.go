package producer

import (
	"encoding/json"

	"github.com/vladislav-koval/url-shortener/internal/core/logger"
	"github.com/vladislav-koval/url-shortener/internal/core/messaging/events"
	"github.com/vladislav-koval/url-shortener/internal/core/messaging/gokafka"
	"go.uber.org/zap"
)

type Producer struct {
	writer gokafka.Writer
	log    *logger.Logger
}

func NewProducer(writer gokafka.Writer, log *logger.Logger) *Producer {
	return &Producer{writer: writer, log: log}
}

func (p *Producer) RecordClick(clickEvent events.ClickEvent) {
	value, err := json.Marshal(clickEvent)
	if err != nil {
		p.log.Error("marshal click event", zap.Error(err))
		return
	}

	p.writer.AsyncWriteMessage(gokafka.Message{Key: []byte(clickEvent.ShortCode), Value: value})
}
