package service

import (
	"context"

	"github.com/vladislav-koval/url-shortener/internal/core/domain"
)

type Service struct {
	shortenerRepository Repository
}

type Repository interface {
	CreateShortLink(ctx context.Context, shortCode string, originalURL string) error
	GetByShortCode(ctx context.Context, code string) (domain.Link, error)
}

func NewService(shortenerRepository Repository) *Service {
	return &Service{
		shortenerRepository: shortenerRepository,
	}
}
