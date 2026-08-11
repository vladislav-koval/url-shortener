package service

import (
	"context"
	"fmt"

	"github.com/vladislav-koval/url-shortener/internal/analytics/stats/domain"
)

func (s *Service) GetLinkClickCount(ctx context.Context, shortCodes []string) ([]domain.LinkClickCount, error) {
	linkClickCounts, err := s.repository.CountByShortCodes(ctx, shortCodes)
	if err != nil {
		return nil, fmt.Errorf("error getting link click counts: %w", err)
	}

	return linkClickCounts, nil
}
