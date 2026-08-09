package service

import (
	"context"

	"github.com/vladislav-koval/url-shortener/internal/analytics/stats/domain"
)

type Service struct {
	repository Repository
}

type Repository interface {
	CountByShortCodes(ctx context.Context, shortCodes []string) ([]domain.LinkClickCount, error)
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}
