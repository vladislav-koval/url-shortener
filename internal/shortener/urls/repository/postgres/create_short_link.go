package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/domain"
)

func (r *Repository) CreateShortLink(ctx context.Context, shortCode string, originalURL string, userID *uuid.UUID) (domain.Link, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO urlshortener.links (short_code, original_url, user_id)
		VALUES ($1, $2, $3)
		RETURNING short_code, original_url, user_id;
	`

	row := r.pool.QueryRow(ctx, query, shortCode, originalURL, userID)

	var link linkRow

	err := row.Scan(
		&link.ShortCode,
		&link.OriginalURL,
		&link.UserID,
	)

	if err != nil {
		if errors.Is(err, pool.ErrUniqueViolation) {
			return domain.Link{}, fmt.Errorf("short code %q already taken: %w", shortCode, apperrors.ErrConflict)
		}

		return domain.Link{}, fmt.Errorf("insert link: %w", err)
	}

	return linkFromRow(link), nil
}
