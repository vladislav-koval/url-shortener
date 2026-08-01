package clicks

import (
	"github.com/vladislav-koval/url-shortener/internal/analytics/clicks/consumer"
	"github.com/vladislav-koval/url-shortener/internal/analytics/clicks/repository/postgres"
	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"github.com/vladislav-koval/url-shortener/internal/platform/messaging/gokafka"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool"
)

type Module struct {
	Consumers []*consumer.ClickConsumer
}

// NewModule По одному консьюмеру на каждый переданный reader
func NewModule(pool pool.Pool, readers []gokafka.Reader, log *logger.Logger, cfg consumer.Config) *Module {
	repository := postgres.NewRepository(pool)

	consumers := make([]*consumer.ClickConsumer, len(readers))
	for i, reader := range readers {
		consumers[i] = consumer.NewClickConsumer(reader, repository, log, cfg)
	}

	return &Module{
		Consumers: consumers,
	}
}
