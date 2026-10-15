package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/vladislav-koval/url-shortener/internal/platform/pagination"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/domain"
)

//go:generate mockgen -source=./service.go -destination=mocks/mock_service.go -package=mocks
type Service struct {
	repository Repository
	grpcClient GRPCClient
}

type Repository interface {
	GetLinksByUserID(ctx context.Context, userID uuid.UUID, p pagination.Pagination) ([]domain.Link, int, error)
	GetLinkByShortCodeAndUserID(ctx context.Context, shortCode string, userID uuid.UUID) (domain.Link, error)
}

type GRPCClient interface {
	GetClickCounts(ctx context.Context, shortCodes []string) ([]domain.LinkClickCount, error)
	GetLinkAnalytics(ctx context.Context, query GetLinkAnalyticsQuery) (GetLinkAnalyticsResult, error)
}

func NewService(repository Repository, grpcClient GRPCClient) *Service {
	return &Service{
		repository: repository,
		grpcClient: grpcClient,
	}
}
