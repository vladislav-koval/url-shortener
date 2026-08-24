package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/vladislav-koval/url-shortener/internal/platform/pagination"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/domain"
)

func (s *Service) GetLinkStats(ctx context.Context, userID uuid.UUID, pagination pagination.Pagination) (domain.LinkStatsPage, error) {
	links, totalCount, err := s.repository.GetLinksByUserID(ctx, userID, pagination)

	if err != nil {
		return domain.LinkStatsPage{}, fmt.Errorf("get links: %w", err)
	}

	if len(links) == 0 {
		return domain.LinkStatsPage{Total: totalCount}, nil
	}

	shortCodes := make([]string, 0, len(links))

	for _, link := range links {
		shortCodes = append(shortCodes, link.ShortCode)
	}

	stats, err := s.grpcClient.GetClickCounts(ctx, shortCodes)
	if err != nil {
		return domain.LinkStatsPage{}, fmt.Errorf("get click counts: %w", err)
	}

	clickCounts := make(map[string]int, len(stats))

	for _, stat := range stats {
		clickCounts[stat.ShortCode] = stat.ClickCount
	}

	items := make([]domain.LinkItem, 0, len(links))

	for _, link := range links {
		items = append(items, domain.LinkItem{
			ShortCode:   link.ShortCode,
			OriginalURL: link.OriginalURL,
			CreatedAt:   link.CreatedAt,
			ClickCount:  clickCounts[link.ShortCode],
		})
	}

	return domain.LinkStatsPage{
		Items: items,
		Total: totalCount,
	}, nil
}
