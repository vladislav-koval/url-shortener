package service

import (
	"context"

	"github.com/vladislav-koval/url-shortener/internal/core/domain"
	"github.com/vladislav-koval/url-shortener/internal/core/messaging/events"
)

//go:generate mockgen -source=./service.go -destination=mocks/mock_service.go -package=mocks
type Service struct {
	shortenerRepository Repository
	clickRecorder       ClickRecorder
}

type Repository interface {
	CreateShortLink(ctx context.Context, shortCode string, originalURL string) (domain.Link, error)
	GetByShortCode(ctx context.Context, code string) (domain.Link, error)
}

type ClickRecorder interface {
	RecordClick(clickEvent events.ClickEvent)
}

func NewService(shortenerRepository Repository, clickRecorder ClickRecorder) *Service {
	return &Service{
		shortenerRepository: shortenerRepository,
		clickRecorder:       clickRecorder,
	}
}
