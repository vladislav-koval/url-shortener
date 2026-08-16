package urls

import (
	"github.com/vladislav-koval/url-shortener/internal/platform/geo"
	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"github.com/vladislav-koval/url-shortener/internal/platform/messaging/gokafka"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/redis"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/producer"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/repository/cached"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/repository/postgres"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/service"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/transport/shortenerhttp"
)

type Module struct {
	Handler *shortenerhttp.Handler
}

func NewModule(pool pool.Pool, cache cache.Pool, clickWriter gokafka.Writer, log *logger.Logger, geoResolver geo.Resolver) *Module {
	shortenerRepository := cached.NewRepository(
		cache,
		cached.NewConfigMust(),
		postgres.NewRepository(pool),
	)

	clickRecorder := producer.NewProducer(clickWriter, log)

	shortenerService := service.NewService(shortenerRepository, clickRecorder, service.NewConfigMust())
	shortenerHTTPHandler := shortenerhttp.NewHTTPHandler(shortenerService, geoResolver)

	return &Module{
		Handler: shortenerHTTPHandler,
	}
}
