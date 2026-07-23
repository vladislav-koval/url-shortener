package shortener

import (
	"github.com/vladislav-koval/url-shortener/internal/core/logger"
	"github.com/vladislav-koval/url-shortener/internal/core/messaging/gokafka"
	"github.com/vladislav-koval/url-shortener/internal/core/repository/postgres/pool"
	"github.com/vladislav-koval/url-shortener/internal/core/repository/redis"
	"github.com/vladislav-koval/url-shortener/internal/features/shortener/producer"
	"github.com/vladislav-koval/url-shortener/internal/features/shortener/repository/cached"
	"github.com/vladislav-koval/url-shortener/internal/features/shortener/repository/postgres"
	"github.com/vladislav-koval/url-shortener/internal/features/shortener/service"
	"github.com/vladislav-koval/url-shortener/internal/features/shortener/transport/shortenerhttp"
)

type Module struct {
	Handler *shortenerhttp.Handler
}

func NewModule(pool pool.Pool, cache cache.Pool, clickWriter gokafka.Writer, log *logger.Logger) *Module {
	shortenerRepository := cached.NewRepository(
		cache,
		cached.NewConfigMust(),
		postgres.NewRepository(pool),
	)

	clickRecorder := producer.NewProducer(clickWriter, log)

	shortenerService := service.NewService(shortenerRepository, clickRecorder)
	shortenerHTTPHandler := shortenerhttp.NewHTTPHandler(shortenerService)

	return &Module{
		Handler: shortenerHTTPHandler,
	}
}
