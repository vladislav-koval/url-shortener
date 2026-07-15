package shortener

import (
	"github.com/vladislav-koval/url-shortener/internal/core/repository/postgres/pool/pgx"
	"github.com/vladislav-koval/url-shortener/internal/features/shortener/repository/postgres"
	"github.com/vladislav-koval/url-shortener/internal/features/shortener/service"
	"github.com/vladislav-koval/url-shortener/internal/features/shortener/transport/http"
)

type Module struct {
	Handler *http.Handler
}

func NewModule(pool *pgx.Pool) *Module {
	shortenerRepository := postgres.NewRepository(pool)
	shortenerService := service.NewService(shortenerRepository)
	shortenerHTTPHandler := http.NewHTTPHandler(shortenerService)

	return &Module{
		Handler: shortenerHTTPHandler,
	}
}
