package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/vladislav-koval/url-shortener/internal/analytics/clicks/domain"
	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
)

func (r *Repository) SaveClicks(ctx context.Context, clicks []domain.Click) error {
	if len(clicks) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const numFields = 9
	query := "INSERT INTO analytics.clicks (id, short_code, country, city, clicked_at, device_type, os, browser, referer) VALUES "

	args := make([]any, 0, len(clicks)*numFields)
	rowPlaceholders := make([]string, len(clicks))

	for i, click := range clicks {
		idx := i * numFields
		placeholders := make([]string, numFields)

		for j := 0; j < numFields; j++ {
			placeholders[j] = fmt.Sprintf("$%d", idx+j+1)
		}

		rowPlaceholders[i] = "(" + strings.Join(placeholders, ", ") + ")"

		args = append(args,
			click.ID,
			click.ShortCode,
			click.Country,
			click.City,
			click.ClickedAt,
			click.DeviceType,
			click.OS,
			click.Browser,
			click.Referer,
		)
	}

	query += strings.Join(rowPlaceholders, ", ") + " ON CONFLICT (id) DO NOTHING;"

	ct, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save clicks: %w", err)
	}

	rowsAffected := ct.RowsAffected()
	rowsExpected := int64(len(clicks))

	if rowsAffected < rowsExpected {
		return fmt.Errorf(
			"rows affected: %d < %d, skip duplicates %d, %w",
			rowsAffected,
			rowsExpected,
			rowsExpected-rowsAffected,
			apperrors.ErrConflict,
		)
	}

	return nil
}
