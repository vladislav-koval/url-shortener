package service

import (
	"context"

	"github.com/vladislav-koval/url-shortener/internal/analytics/clicks/domain"
)

//go:generate mockgen -source=./service.go -destination=mocks/mock_service.go -package=mocks
type Repository interface {
	SaveClicks(ctx context.Context, events []domain.Click) error
}
type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}
