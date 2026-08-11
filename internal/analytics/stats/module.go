package stats

import (
	"github.com/vladislav-koval/url-shortener/internal/analytics/stats/repository/postgres"
	"github.com/vladislav-koval/url-shortener/internal/analytics/stats/service"
	statsgrpc "github.com/vladislav-koval/url-shortener/internal/analytics/stats/transport/grpc"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool"
)

type Module struct {
	Handler *statsgrpc.Handler
}

func NewStatsModule(pool pool.Pool) *Module {
	repository := postgres.NewRepository(pool)
	svc := service.NewService(repository)
	handler := statsgrpc.NewHandler(svc)

	return &Module{
		Handler: handler,
	}
}
