package shortener

import (
	"github.com/vladislav-koval/url-shortener/internal/core/repository/postgres/pool/pgx"
	"github.com/vladislav-koval/url-shortener/internal/core/repository/redis/goredis"
	"github.com/vladislav-koval/url-shortener/internal/features/shortener/repository/cached"
	"github.com/vladislav-koval/url-shortener/internal/features/shortener/repository/postgres"
	"github.com/vladislav-koval/url-shortener/internal/features/shortener/service"
	"github.com/vladislav-koval/url-shortener/internal/features/shortener/transport/shortenerhttp"
)

type Module struct {
	Handler *shortenerhttp.Handler
}

func NewModule(pool *pgx.Pool, cache *goredis.Redis) *Module {
	shortenerRepository := cached.NewRepository(
		cache,
		cached.NewConfigMust(),
		postgres.NewRepository(pool),
	)

	shortenerService := service.NewService(shortenerRepository)
	shortenerHTTPHandler := shortenerhttp.NewHTTPHandler(shortenerService)

	return &Module{
		Handler: shortenerHTTPHandler,
	}
}
