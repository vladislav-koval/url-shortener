package service

import (
	"context"

	"github.com/vladislav-koval/url-shortener/internal/core/domain"
)

type Service struct {
	shortenerRepository ShortenerRepository
}

type ShortenerRepository interface {
	CreateShortLink(ctx context.Context, shortCode string, originalURL string) error
	GetByShortCode(ctx context.Context, code string) (domain.Link, error)
}

func NewShortenerService(shortenerRepository ShortenerRepository) *Service {
	return &Service{
		shortenerRepository: shortenerRepository,
	}
}
