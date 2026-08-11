package service

import (
	"context"

	"github.com/vladislav-koval/url-shortener/internal/analytics/stats/domain"
)

//go:generate mockgen -source=./service.go -destination=mocks/mock_service.go -package=mocks
type Service struct {
	repository Repository
}

type Repository interface {
	CountByShortCodes(ctx context.Context, shortCodes []string) ([]domain.LinkClickCount, error)
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}
