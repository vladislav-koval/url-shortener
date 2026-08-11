package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/vladislav-koval/url-shortener/internal/platform/pagination"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/domain"
)

func (s *Service) GetClickCounts(ctx context.Context, userID uuid.UUID, pagination pagination.Pagination) (domain.LinkClickCountPage, error) {
	shortCodes, totalCount, err := s.repository.GetShortCodesByUserID(ctx, userID, pagination)

	if err != nil {
		return domain.LinkClickCountPage{}, fmt.Errorf("get short codes: %w", err)
	}

	if len(shortCodes) == 0 {
		return domain.LinkClickCountPage{Total: totalCount}, nil
	}

	clickCounts, err := s.grpcClient.GetClickCounts(ctx, shortCodes)
	if err != nil {
		return domain.LinkClickCountPage{}, fmt.Errorf("get click counts: %w", err)
	}

	return domain.LinkClickCountPage{
		Items: clickCounts,
		Total: totalCount,
	}, nil
}
