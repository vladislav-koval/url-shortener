package stats

import (
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/client"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/repository/postgres"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/service"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/transport/statshttp"
	"google.golang.org/grpc"
)

type Module struct {
	Handler *statshttp.Handler
}

func NewModule(pool pool.Pool, grpcConn grpc.ClientConnInterface) *Module {
	repository := postgres.NewRepository(pool)

	grpcClient := client.NewClient(grpcConn)

	svc := service.NewService(repository, grpcClient)
	handler := statshttp.NewHTTPHandler(svc)

	return &Module{
		Handler: handler,
	}
}
