package postgres

import (
	"context"
	"fmt"

	"github.com/vladislav-koval/url-shortener/internal/analytics/stats/domain"
)

func (r *Repository) CountByShortCodes(ctx context.Context, shortCodes []string) ([]domain.LinkClickCount, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		WITH requested AS (
			SELECT DISTINCT unnest($1::text[]) AS short_code
		)
		SELECT
			requested.short_code,
			COUNT(clicks.short_code) AS click_count
		FROM requested
		LEFT JOIN urlshortener.clicks AS clicks
			ON clicks.short_code = requested.short_code
		GROUP BY requested.short_code;
	`

	rows, err := r.pool.Query(ctx, query, shortCodes)
	if err != nil {
		return nil, fmt.Errorf("query link click count: %w", err)
	}
	defer rows.Close()

	stats := make([]statRow, 0, len(shortCodes))

	for rows.Next() {
		var row statRow

		if err := rows.Scan(&row.shortCode, &row.clickCount); err != nil {
			return nil, fmt.Errorf("scan link click count: %w", err)
		}

		stats = append(stats, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("next rows: %w", err)
	}

	linkClickCountDomains := rowsLinkClickCountToDomain(stats)
	fmt.Println(linkClickCountDomains)
	return linkClickCountDomains, nil
}
