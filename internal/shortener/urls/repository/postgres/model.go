package postgres

import (
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/domain"
)

type linkRow struct {
	OriginalURL string
	ShortCode   string
}

func linkFromRow(row linkRow) domain.Link {
	return domain.NewLink(row.ShortCode, row.OriginalURL)
}
