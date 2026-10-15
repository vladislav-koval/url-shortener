package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
)

type Granularity string

const (
	GranularityAuto      Granularity = "auto"
	Granularity5Minutes  Granularity = "5m"
	Granularity15Minutes Granularity = "15m"
	Granularity30Minutes Granularity = "30m"
	Granularity1Hour     Granularity = "1h"
	Granularity6Hours    Granularity = "6h"
	Granularity12Hours   Granularity = "12h"
	Granularity1Day      Granularity = "1d"
	Granularity7Days     Granularity = "7d"
)

type GetLinkAnalyticsQuery struct {
	ShortCode   string
	From        time.Time
	To          time.Time
	Granularity Granularity
}
type BreakdownItemType string

const (
	BreakdownItemValue   BreakdownItemType = "value"
	BreakdownItemOther   BreakdownItemType = "other"
	BreakdownItemUnknown BreakdownItemType = "unknown"
)

type Breakdown struct {
	Items []BreakdownItem
}

type BreakdownItem struct {
	Type   BreakdownItemType
	Value  string
	Clicks int
}

type TimelinePoint struct {
	BucketStart time.Time
	Clicks      int
}

type GetLinkAnalyticsResult struct {
	TotalClicks      int
	Timeline         []TimelinePoint
	Countries        Breakdown
	Cities           Breakdown
	Devices          Breakdown
	OperatingSystems Breakdown
	Browsers         Breakdown
	Referrers        Breakdown
}

func (s *Service) GetLinkAnalytics(
	ctx context.Context,
	query GetLinkAnalyticsQuery,
	userID uuid.UUID,
) (GetLinkAnalyticsResult, error) {
	_, err := s.repository.GetLinkByShortCodeAndUserID(ctx, query.ShortCode, userID)

	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return GetLinkAnalyticsResult{}, fmt.Errorf("link not found: %w", err)
		}

		return GetLinkAnalyticsResult{}, fmt.Errorf("get link analytics: %w", err)
	}

	result, err := s.grpcClient.GetLinkAnalytics(ctx, query)
	if err != nil {
		return GetLinkAnalyticsResult{}, fmt.Errorf("get link analytics: %w", err)
	}

	return result, nil
}
