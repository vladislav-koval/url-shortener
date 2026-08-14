package postgres

import (
	"context"
	"fmt"

	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/platform/messaging/events"
)

func (r *Repository) SaveClicks(ctx context.Context, events []events.ClickEvent) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	numFields := 5
	query := "INSERT INTO analytics.clicks (id, short_code, country, city, clicked_at) VALUES "

	args := make([]any, 0, len(events)*numFields)
	for i, event := range events {
		if i > 0 {
			query += ","
		}
		idx := i * numFields
		query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", idx+1, idx+2, idx+3, idx+4, idx+5)

		args = append(args, event.ID, event.ShortCode, event.CountryCode, event.City, event.ClickedAt)
	}

	query += " ON CONFLICT (id) DO NOTHING;"

	ct, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save clicks: %w", err)
	}

	rowsAffected := ct.RowsAffected()
	rowsExpected := int64(len(events))

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
