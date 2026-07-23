package postgres

import (
	"context"
	"fmt"

	"github.com/vladislav-koval/url-shortener/internal/core/apperrors"
	"github.com/vladislav-koval/url-shortener/internal/core/messaging/events"
)

func (r *Repository) SaveClicks(ctx context.Context, events []events.ClickEvent) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	numFields := 4
	query := "INSERT INTO urlshortener.clicks (id, short_code, ip_address, clicked_at) VALUES "

	args := make([]any, 0, len(events)*numFields)
	for i, event := range events {
		if i > 0 {
			query += ","
		}
		idx := i * numFields
		query += fmt.Sprintf("($%d, $%d, $%d, $%d)", idx+1, idx+2, idx+3, idx+4)

		args = append(args, event.ID, event.ShortCode, event.IP, event.ClickedAt)
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
