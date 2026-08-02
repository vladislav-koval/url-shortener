package urls

import (
	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"github.com/vladislav-koval/url-shortener/internal/platform/messaging/gokafka"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/redis"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/producer"
	cached2 "github.com/vladislav-koval/url-shortener/internal/shortener/urls/repository/cached"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/repository/postgres"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/service"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/transport/shortenerhttp"
)

type Module struct {
	Handler *shortenerhttp.Handler
}

func NewModule(pool pool.Pool, cache cache.Pool, clickWriter gokafka.Writer, log *logger.Logger) *Module {
	shortenerRepository := cached2.NewRepository(
		cache,
		cached2.NewConfigMust(),
		postgres.NewRepository(pool),
	)

	clickRecorder := producer.NewProducer(clickWriter, log)

	shortenerService := service.NewService(shortenerRepository, clickRecorder)
	shortenerHTTPHandler := shortenerhttp.NewHTTPHandler(shortenerService)

	return &Module{
		Handler: shortenerHTTPHandler,
	}
}
