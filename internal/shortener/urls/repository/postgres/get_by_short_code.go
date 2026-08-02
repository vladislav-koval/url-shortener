package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/domain"
)

func (r *Repository) GetByShortCode(ctx context.Context, code string) (domain.Link, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT short_code, original_url FROM urlshortener.links WHERE short_code = $1;
	`

	row := r.pool.QueryRow(ctx, query, code)

	var model linkRow

	err := row.Scan(
		&model.ShortCode,
		&model.OriginalURL,
	)

	if err != nil {
		if errors.Is(err, pool.ErrNoRows) {
			return domain.Link{}, fmt.Errorf("url with code='%s': %w", code, apperrors.ErrNotFound)
		}

		return domain.Link{}, fmt.Errorf("scan error: %w", err)
	}

	return linkFromRow(model), nil
}
