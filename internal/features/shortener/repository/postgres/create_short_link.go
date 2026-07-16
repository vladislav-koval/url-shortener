package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/vladislav-koval/url-shortener/internal/core/domain"
	"github.com/vladislav-koval/url-shortener/internal/core/errors"
	"github.com/vladislav-koval/url-shortener/internal/core/repository/postgres/pool"
)

func (r *Repository) CreateShortLink(ctx context.Context, shortCode string, originalURL string) (domain.Link, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO urlshortener.links (short_code, original_url)
		VALUES ($1, $2)
		RETURNING short_code, original_url;
	`

	row := r.pool.QueryRow(ctx, query, shortCode, originalURL)

	var link linkRow

	err := row.Scan(
		&link.ShortCode,
		&link.OriginalURL,
	)

	if err != nil {
		if errors.Is(err, pool.ErrUniqueViolation) {
			return domain.Link{}, fmt.Errorf("short code %q already taken: %w", shortCode, apperrors.ErrConflict)
		}

		return domain.Link{}, fmt.Errorf("insert link: %w", err)
	}

	return linkFromRow(link), nil
}
