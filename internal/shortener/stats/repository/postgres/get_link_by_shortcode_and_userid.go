package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/domain"
)

func (repo *Repository) GetLinkByShortCodeAndUserID(ctx context.Context,
	code string,
	userID uuid.UUID,
) (domain.Link, error) {
	ctx, cancel := context.WithTimeout(ctx, repo.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT short_code, original_url, created_at
		FROM urlshortener.links
		WHERE user_id = $1
		AND short_code = $2;
	`

	row := repo.pool.QueryRow(ctx, query, userID, code)

	var model linkRow

	err := row.Scan(
		&model.shortCode,
		&model.originalUrl,
		&model.createdAt,
	)

	if err != nil {
		if errors.Is(err, pool.ErrNoRows) {
			return domain.Link{}, fmt.Errorf("link with code='%s': %w", code, apperrors.ErrNotFound)
		}

		return domain.Link{}, fmt.Errorf("scan error: %w", err)
	}

	return linkFromRow(model), nil
}
