package postgres

import "github.com/vladislav-koval/url-shortener/internal/core/domain"

type linkRow struct {
	OriginalURL string
	ShortCode   string
}

func linkFromRow(row linkRow) domain.Link {
	return domain.NewLink(row.ShortCode, row.OriginalURL)
}
