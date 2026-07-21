package service

import (
	"context"
	"fmt"

	"github.com/vladislav-koval/url-shortener/internal/core/messaging/events"
)

func (s *Service) ResolveShortLink(ctx context.Context, code string, event events.ClickEvent) (string, error) {
	link, err := s.shortenerRepository.GetByShortCode(ctx, code)
	if err != nil {
		return "", fmt.Errorf("get link from repo: %w", err)
	}

	s.clickRecorder.RecordClick(ctx, event)

	return link.OriginalURL, nil
}
