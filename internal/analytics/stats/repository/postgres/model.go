package postgres

import (
	"time"

	"github.com/vladislav-koval/url-shortener/internal/analytics/stats/domain"
	"github.com/vladislav-koval/url-shortener/internal/analytics/stats/service"
)

type statRow struct {
	shortCode  string
	clickCount int
}

func rowsLinkClickCountToDomain(rows []statRow) []domain.LinkClickCount {
	domainLinkClickCount := make([]domain.LinkClickCount, len(rows))

	for i, row := range rows {
		domainLinkClickCount[i] = domain.LinkClickCount{ShortCode: row.shortCode, ClickCount: row.clickCount}
	}

	return domainLinkClickCount
}

type breakdownItemRow struct {
	breakdownType service.BreakdownItemType
	value         string
	clicks        int
}

type timelineRow struct {
	bucketStart time.Time
	clicks      int
}

func toServiceBreakdown(rows []breakdownItemRow) service.Breakdown {
	breakdownItems := make([]service.BreakdownItem, len(rows))

	for i, row := range rows {
		breakdownItems[i] = service.BreakdownItem{
			Type:   row.breakdownType,
			Value:  row.value,
			Clicks: row.clicks,
		}
	}

	return service.Breakdown{
		Items: breakdownItems,
	}
}

func toServiceTimeline(rows []timelineRow) []service.TimelinePoint {
	timelinePoints := make([]service.TimelinePoint, len(rows))

	for i, row := range rows {
		timelinePoints[i] = service.TimelinePoint{
			BucketStart: row.bucketStart,
			Clicks:      row.clicks,
		}
	}

	return timelinePoints
}
