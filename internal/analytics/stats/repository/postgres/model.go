package postgres

import "github.com/vladislav-koval/url-shortener/internal/analytics/stats/domain"

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
