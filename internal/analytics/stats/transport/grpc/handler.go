package statsgrpc

import (
	"context"

	analyticsv1 "github.com/vladislav-koval/url-shortener/api/gen/analytics/v1"
	"github.com/vladislav-koval/url-shortener/internal/analytics/stats/domain"
	"google.golang.org/grpc"
)

type Handler struct {
	analyticsv1.UnimplementedAnalyticsServiceServer
	service Service
}

type Service interface {
	GetLinkClickCount(ctx context.Context, shortCodes []string) ([]domain.LinkClickCount, error)
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Register(registrar grpc.ServiceRegistrar) {
	analyticsv1.RegisterAnalyticsServiceServer(
		registrar,
		h,
	)
}
