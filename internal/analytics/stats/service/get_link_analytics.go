package service

import (
	"context"
	"fmt"
	"time"

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

func (g Granularity) Duration() time.Duration {
	switch g {
	case Granularity5Minutes:
		return 5 * time.Minute
	case Granularity15Minutes:
		return 15 * time.Minute
	case Granularity30Minutes:
		return 30 * time.Minute
	case Granularity1Hour:
		return time.Hour
	case Granularity6Hours:
		return 6 * time.Hour
	case Granularity12Hours:
		return 12 * time.Hour
	case Granularity1Day:
		return 24 * time.Hour
	case Granularity7Days:
		return 7 * 24 * time.Hour
	default:
		return 0
	}
}

func ResolveGranularity(
	g Granularity,
	from, to time.Time,
) Granularity {
	if g != GranularityAuto {
		return g
	}

	duration := to.Sub(from)

	granularities := []Granularity{
		Granularity5Minutes,
		Granularity15Minutes,
		Granularity30Minutes,
		Granularity1Hour,
		Granularity6Hours,
		Granularity12Hours,
		Granularity1Day,
		Granularity7Days,
	}

	for _, candidate := range granularities {
		points := duration / candidate.Duration()

		if points <= 200 {
			return candidate
		}
	}

	return Granularity7Days
}

type GetLinkAnalyticsQuery struct {
	ShortCode   string
	From        time.Time
	To          time.Time
	Granularity Granularity
}

func (q GetLinkAnalyticsQuery) Validate() error {
	if q.ShortCode == "" {
		return fmt.Errorf("short code is required: %w", apperrors.ErrInvalidArgument)
	}

	if !q.From.Before(q.To) {
		return fmt.Errorf("from must be before to: %w", apperrors.ErrInvalidArgument)
	}

	if err := q.validateGranularity(); err != nil {
		return err
	}

	return nil
}

func (q GetLinkAnalyticsQuery) validateGranularity() error {
	if q.Granularity == GranularityAuto {
		return nil
	}

	duration := q.Granularity.Duration()
	if duration <= 0 {
		return fmt.Errorf("invalid granularity: %w", apperrors.ErrInvalidArgument)
	}

	const maxTimelinePoints = 500

	points := q.To.Sub(q.From) / duration
	if points > maxTimelinePoints {
		return fmt.Errorf(
			"granularity is too small for selected range: %w",
			apperrors.ErrInvalidArgument,
		)
	}

	return nil
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

type TimelinePoint struct {
	BucketStart time.Time
	Clicks      int
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

func (b Breakdown) TotalClicks() int {
	var total int

	for _, item := range b.Items {
		total += item.Clicks
	}

	return total
}

type BreakdownItem struct {
	Type   BreakdownItemType
	Value  string
	Clicks int
}

func (s *Service) GetLinkAnalytics(ctx context.Context, query GetLinkAnalyticsQuery) (GetLinkAnalyticsResult, error) {
	if err := query.Validate(); err != nil {
		return GetLinkAnalyticsResult{}, err
	}

	cities, err := s.repository.GetCities(ctx, query.ShortCode, query.From, query.To)
	if err != nil {
		return GetLinkAnalyticsResult{}, err
	}

	countries, err := s.repository.GetCountries(ctx, query.ShortCode, query.From, query.To)
	if err != nil {
		return GetLinkAnalyticsResult{}, err
	}

	devices, err := s.repository.GetDevices(ctx, query.ShortCode, query.From, query.To)
	if err != nil {
		return GetLinkAnalyticsResult{}, err
	}

	operatingSystems, err := s.repository.GetOperatingSystems(ctx, query.ShortCode, query.From, query.To)
	if err != nil {
		return GetLinkAnalyticsResult{}, err
	}

	browsers, err := s.repository.GetBrowsers(ctx, query.ShortCode, query.From, query.To)
	if err != nil {
		return GetLinkAnalyticsResult{}, err
	}

	referrers, err := s.repository.GetReferers(ctx, query.ShortCode, query.From, query.To)
	if err != nil {
		return GetLinkAnalyticsResult{}, err
	}

	granularity := ResolveGranularity(query.Granularity, query.From, query.To)
	timeline, err := s.repository.GetTimeline(
		ctx,
		query.ShortCode,
		query.From,
		query.To,
		granularity.Duration(),
	)
	if err != nil {
		return GetLinkAnalyticsResult{}, err
	}

	return GetLinkAnalyticsResult{
		Cities:           cities,
		Countries:        countries,
		Devices:          devices,
		OperatingSystems: operatingSystems,
		Browsers:         browsers,
		Referrers:        referrers,
		Timeline:         timeline,
		TotalClicks:      cities.TotalClicks(),
	}, nil
}
