package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/vladislav-koval/url-shortener/internal/platform/pagination"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/domain"
)

func (repo *Repository) GetLinksByUserID(ctx context.Context, userID uuid.UUID, p pagination.Pagination) ([]domain.Link, int, error) {
	ctx, cancel := context.WithTimeout(ctx, repo.pool.OpTimeout())
	defer cancel()

	queryCount := `
		SELECT COUNT(*)
		FROM urlshortener.links
		WHERE user_id = $1
	`

	row := repo.pool.QueryRow(ctx, queryCount, userID)
	var totalCount int
	err := row.Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("scan error: %w", err)
	}

	query := `SELECT short_code, original_url, created_at
				FROM urlshortener.links
				WHERE user_id = $1
				ORDER BY created_at DESC
				LIMIT $2
				OFFSET $3;
	`

	rows, err := repo.pool.Query(ctx, query, userID, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("select short codes: %w", err)
	}
	defer rows.Close()

	links := make([]linkRow, 0, p.Limit)

	for rows.Next() {
		var link linkRow
		err := rows.Scan(&link.shortCode, &link.originalUrl, &link.createdAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scan link: %w", err)
		}

		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("next rows: %w", err)
	}

	return linksFromRow(links), totalCount, nil
}
