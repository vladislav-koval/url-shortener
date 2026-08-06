package postgres

import (
	"github.com/google/uuid"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/domain"
)

type linkRow struct {
	OriginalURL string
	ShortCode   string
	UserID      *uuid.UUID
}

func linkFromRow(row linkRow) domain.Link {
	return domain.NewLink(row.ShortCode, row.OriginalURL, row.UserID)
}
