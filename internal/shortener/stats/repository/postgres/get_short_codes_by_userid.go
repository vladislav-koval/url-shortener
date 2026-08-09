package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (repo *Repository) GetShortCodesByUserID(userID uuid.UUID, limit, offset int) ([]string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), repo.pool.OpTimeout())
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

	query := `SELECT short_code
				FROM urlshortener.links
				WHERE user_id = $1
				ORDER BY short_code ASC
				LIMIT $2
				OFFSET $3;
	`

	rows, err := repo.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("select short codes: %w", err)
	}
	defer rows.Close()

	shortCodes := make([]string, limit)

	for rows.Next() {
		var shortCode string
		err := rows.Scan(&shortCode)
		if err != nil {
			return nil, 0, fmt.Errorf("scan short code: %w", err)
		}

		shortCodes = append(shortCodes, shortCode)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("next rows: %w", err)
	}

	return shortCodes, totalCount, nil
}
