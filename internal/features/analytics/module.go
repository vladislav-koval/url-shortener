package analytics

import (
	"context"

	"github.com/vladislav-koval/url-shortener/internal/core/logger"
	"github.com/vladislav-koval/url-shortener/internal/core/messaging/gokafka"
	"github.com/vladislav-koval/url-shortener/internal/core/repository/postgres/pool"
	"github.com/vladislav-koval/url-shortener/internal/features/analytics/consumer"
	"github.com/vladislav-koval/url-shortener/internal/features/analytics/repository/postgres"
)

func StartConsumer(
	ctx context.Context,
	pool pool.Pool,
	reader gokafka.Reader,
	log *logger.Logger,
	count int,
) {
	repository := postgres.NewRepository(pool)
	cfg := consumer.NewConfigMust()
	clickConsumer := consumer.NewClickConsumer(reader, repository, log, cfg)

	for i := 0; i < count; i++ {
		go clickConsumer.Start(ctx)
	}
}
