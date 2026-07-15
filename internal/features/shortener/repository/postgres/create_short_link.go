package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/vladislav-koval/url-shortener/internal/core/errors"
	"github.com/vladislav-koval/url-shortener/internal/core/repository/postgres/pool"
)

func (r *Repository) CreateShortLink(ctx context.Context, shortCode string, originalURL string) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO urlshortener.links (short_code, original_url)
		VALUES ($1, $2);
	`

	_, err := r.pool.Exec(ctx, query, shortCode, originalURL)
	if err != nil {
		if errors.Is(err, pool.ErrUniqueViolation) {
			return fmt.Errorf("short code %q already taken: %w", shortCode, apperrors.ErrConflict)
		}

		return fmt.Errorf("insert link: %w", err)
	}

	return nil
}
