package postgres

import (
	"time"

	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/domain"
)

type linkRow struct {
	shortCode   string
	originalUrl string
	createdAt   time.Time
}

func linksFromRow(links []linkRow) []domain.Link {
	result := make([]domain.Link, len(links))
	for i, link := range links {
		result[i] = domain.Link{
			ShortCode:   link.shortCode,
			OriginalURL: link.originalUrl,
			CreatedAt:   link.createdAt,
		}
	}

	return result
}
