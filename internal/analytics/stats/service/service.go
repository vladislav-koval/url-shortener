package service

import (
	"context"
	"time"

	"github.com/vladislav-koval/url-shortener/internal/analytics/stats/domain"
)

//go:generate mockgen -source=./service.go -destination=mocks/mock_service.go -package=mocks
type Service struct {
	repository Repository
}

type Repository interface {
	CountByShortCodes(ctx context.Context, shortCodes []string) ([]domain.LinkClickCount, error)
	GetCities(ctx context.Context, shortCode string, from time.Time, to time.Time) (Breakdown, error)
	GetCountries(ctx context.Context, shortCode string, from time.Time, to time.Time) (Breakdown, error)
	GetDevices(ctx context.Context, shortCode string, from time.Time, to time.Time) (Breakdown, error)
	GetOperatingSystems(ctx context.Context, shortCode string, from time.Time, to time.Time) (Breakdown, error)
	GetBrowsers(ctx context.Context, shortCode string, from time.Time, to time.Time) (Breakdown, error)
	GetReferers(ctx context.Context, shortCode string, from time.Time, to time.Time) (Breakdown, error)
	GetTimeline(ctx context.Context, shortCode string, from time.Time, to time.Time, duration time.Duration) ([]TimelinePoint, error)
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}
