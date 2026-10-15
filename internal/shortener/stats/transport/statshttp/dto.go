package statshttp

import (
	"time"

	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/domain"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/service"
)

type ClickCount struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	ClickCount  int    `json:"click_count"`
	CreatedAt   string `json:"created_at"`
}

type ClickCountResponse struct {
	Items []ClickCount `json:"items"`
	Total int          `json:"total"`
}

func domainToResponse(page domain.LinkStatsPage) ClickCountResponse {
	items := make([]ClickCount, len(page.Items))

	for i, item := range page.Items {
		items[i] = ClickCount{
			ShortCode:   item.ShortCode,
			OriginalURL: item.OriginalURL,
			ClickCount:  item.ClickCount,
			CreatedAt:   item.CreatedAt.Format(time.RFC3339),
		}
	}

	return ClickCountResponse{
		Items: items,
		Total: page.Total,
	}
}

type GetLinkAnalyticsResponse struct {
	TotalClicks      int             `json:"total_clicks"`
	Timeline         []TimelinePoint `json:"timeline"`
	Countries        Breakdown       `json:"countries"`
	Cities           Breakdown       `json:"cities"`
	Devices          Breakdown       `json:"devices"`
	OperatingSystems Breakdown       `json:"operating_systems"`
	Browsers         Breakdown       `json:"browsers"`
	Referrers        Breakdown       `json:"referrers"`
}

type TimelinePoint struct {
	BucketStart time.Time `json:"bucket_start"`
	Clicks      int       `json:"clicks"`
}

type Breakdown struct {
	Items []BreakdownItem `json:"items"`
}

type BreakdownItem struct {
	Type   string `json:"type"`
	Value  string `json:"value"`
	Clicks int    `json:"clicks"`
}

func serviceToResponse(
	analytics service.GetLinkAnalyticsResult,
) GetLinkAnalyticsResponse {
	return GetLinkAnalyticsResponse{
		TotalClicks:      analytics.TotalClicks,
		Timeline:         timelineToResponse(analytics.Timeline),
		Countries:        breakdownToResponse(analytics.Countries),
		Cities:           breakdownToResponse(analytics.Cities),
		Devices:          breakdownToResponse(analytics.Devices),
		OperatingSystems: breakdownToResponse(analytics.OperatingSystems),
		Browsers:         breakdownToResponse(analytics.Browsers),
		Referrers:        breakdownToResponse(analytics.Referrers),
	}
}

func timelineToResponse(
	timeline []service.TimelinePoint,
) []TimelinePoint {
	result := make([]TimelinePoint, 0, len(timeline))

	for _, point := range timeline {
		result = append(result, TimelinePoint{
			BucketStart: point.BucketStart,
			Clicks:      point.Clicks,
		})
	}

	return result
}

func breakdownToResponse(
	breakdown service.Breakdown,
) Breakdown {
	items := make([]BreakdownItem, 0, len(breakdown.Items))

	for _, item := range breakdown.Items {
		items = append(items, BreakdownItem{
			Type:   string(item.Type),
			Value:  item.Value,
			Clicks: item.Clicks,
		})
	}

	return Breakdown{
		Items: items,
	}
}
