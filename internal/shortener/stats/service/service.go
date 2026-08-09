package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/vladislav-koval/url-shortener/internal/platform/pagination"
)

type Service struct {
	repository Repository
}

type Repository interface {
	GetShortCodesByUserID(ctx context.Context, userID uuid.UUID, p pagination.Pagination) ([]string, int, error)
}

func NewService(r Repository) *Service {
	return &Service{
		repository: r,
	}
}
