package postgres

import (
	"time"

	"github.com/google/uuid"

	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/domain"
)

type userRow struct {
	ID            uuid.UUID
	GoogleSub     string
	Email         string
	EmailVerified bool
	Name          string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func userFromRow(row userRow) domain.User {
	return domain.User{
		ID:            row.ID,
		GoogleSub:     row.GoogleSub,
		Email:         row.Email,
		EmailVerified: row.EmailVerified,
		Name:          row.Name,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}
