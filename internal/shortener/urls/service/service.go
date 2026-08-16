package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/vladislav-koval/url-shortener/internal/platform/messaging/events"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/domain"
)

//go:generate mockgen -source=./service.go -destination=mocks/mock_service.go -package=mocks
type Service struct {
	shortenerRepository Repository
	clickRecorder       ClickRecorder
	bannedDomains       []string
}

type Repository interface {
	CreateShortLink(ctx context.Context, shortCode string, originalURL string, userID *uuid.UUID) (domain.Link, error)
	GetByShortCode(ctx context.Context, code string) (domain.Link, error)
}

type ClickRecorder interface {
	RecordClick(clickEvent events.ClickEvent)
}

func NewService(shortenerRepository Repository, clickRecorder ClickRecorder, cfg Config) *Service {
	return &Service{
		shortenerRepository: shortenerRepository,
		clickRecorder:       clickRecorder,
		bannedDomains:       cfg.BannedDomains,
	}
}
