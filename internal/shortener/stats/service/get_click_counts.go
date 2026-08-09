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

	//clickCounts, err := grpcClient.GetClickCounts(shortCodes)
	// if err != nil { do something }
	// return clickCounts
}
