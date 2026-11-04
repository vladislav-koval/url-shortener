package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/vladislav-koval/url-shortener/internal/analytics/stats/service"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool"
)

func rowsToBreakdownItemRows(rows pool.Rows) ([]breakdownItemRow, error) {
	breakdownItems := make([]breakdownItemRow, 0, 7)

	for rows.Next() {
		var row breakdownItemRow

		if err := rows.Scan(&row.breakdownType, &row.value, &row.clicks); err != nil {
			return nil, fmt.Errorf("scan analytics: %w", err)
		}

		breakdownItems = append(breakdownItems, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("next rows: %w", err)
	}

	return breakdownItems, nil
}

func (r *Repository) GetCities(
	ctx context.Context,
	shortCode string,
	from time.Time,
	to time.Time,
) (service.Breakdown, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		WITH grouped AS (
		    SELECT
		        NULLIF(city, '') AS city,
		        NULLIF(country, '') AS country,
		        COUNT(*) AS clicks
		    FROM analytics.clicks
		    WHERE short_code = $1
		      AND clicked_at >= $2
		      AND clicked_at < $3
		    GROUP BY NULLIF(country, ''), NULLIF(city, '')
		),
		ranked AS (
		    SELECT
		        CASE
		            WHEN country IS NOT NULL THEN city || ', ' || country
		            ELSE city
		        END AS value,
		        clicks,
		        ROW_NUMBER() OVER (ORDER BY clicks DESC, country, city) AS rn
		    FROM grouped
		    WHERE city IS NOT NULL
		)
		SELECT type, value, clicks
		FROM (
		    SELECT
		        'value' AS type,
		        value,
		        clicks,
		        rn AS sort_order
		    FROM ranked
		    WHERE rn <= 5
		
		    UNION ALL
		
		    SELECT
		        'other',
		        '',
		        SUM(clicks),
		        6
		    FROM ranked
		    WHERE rn > 5
		    HAVING SUM(clicks) > 0
		
		    UNION ALL
		
		    SELECT
		        'unknown',
		        '',
		        SUM(clicks),
		        7
		    FROM grouped
		    WHERE city IS NULL
		    HAVING SUM(clicks) > 0
		) result
		ORDER BY sort_order;
	`

	rows, err := r.pool.Query(ctx, query, shortCode, from, to)

	if err != nil {
		return service.Breakdown{}, fmt.Errorf("query cities analytics: %w", err)
	}
	defer rows.Close()

	breakdownItems, err := rowsToBreakdownItemRows(rows)

	if err != nil {
		return service.Breakdown{}, fmt.Errorf("failed to scan cities: %w", err)
	}

	return toServiceBreakdown(breakdownItems), nil
}

func (r *Repository) GetCountries(
	ctx context.Context,
	shortCode string,
	from time.Time,
	to time.Time,
) (service.Breakdown, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		WITH grouped AS (
		    SELECT
		        NULLIF(country, '') AS value,
		        COUNT(*) AS clicks
		    FROM analytics.clicks
		    WHERE short_code = $1
		      AND clicked_at >= $2
		      AND clicked_at < $3
		    GROUP BY NULLIF(country, '')
		),
		ranked AS (
		    SELECT
		        value,
		        clicks,
		        ROW_NUMBER() OVER (ORDER BY clicks DESC, value) AS rn
		    FROM grouped
		    WHERE value IS NOT NULL
		)
		SELECT type, value, clicks
		FROM (
		    SELECT
		        'value' AS type,
		        value,
		        clicks,
		        rn AS sort_order
		    FROM ranked
		    WHERE rn <= 5

		    UNION ALL

		    SELECT
		        'other',
		        '',
		        SUM(clicks),
		        6
		    FROM ranked
		    WHERE rn > 5
		    HAVING SUM(clicks) > 0

		    UNION ALL

		    SELECT
		        'unknown',
		        '',
		        clicks,
		        7
		    FROM grouped
		    WHERE value IS NULL
		) result
		ORDER BY sort_order;
	`

	rows, err := r.pool.Query(ctx, query, shortCode, from, to)

	if err != nil {
		return service.Breakdown{}, fmt.Errorf("query countries analytics: %w", err)
	}
	defer rows.Close()

	breakdownItems, err := rowsToBreakdownItemRows(rows)

	if err != nil {
		return service.Breakdown{}, fmt.Errorf("failed to scan countries: %w", err)
	}

	return toServiceBreakdown(breakdownItems), nil
}

func (r *Repository) GetDevices(
	ctx context.Context,
	shortCode string,
	from time.Time,
	to time.Time,
) (service.Breakdown, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		WITH grouped AS (
		    SELECT
		        NULLIF(device_type, '') AS value,
		        COUNT(*) AS clicks
		    FROM analytics.clicks
		    WHERE short_code = $1
		      AND clicked_at >= $2
		      AND clicked_at < $3
		    GROUP BY NULLIF(device_type, '')
		)
		SELECT
		    CASE
		        WHEN value IS NULL THEN 'unknown'
		        ELSE 'value'
		    END AS type,
		    COALESCE(value, '') AS value,
		    clicks
		FROM grouped
		ORDER BY
		    value IS NULL,
		    clicks DESC;
	`

	rows, err := r.pool.Query(ctx, query, shortCode, from, to)
	if err != nil {
		return service.Breakdown{}, fmt.Errorf("query device_type analytics: %w", err)
	}
	defer rows.Close()

	breakdownItems, err := rowsToBreakdownItemRows(rows)
	if err != nil {
		return service.Breakdown{}, fmt.Errorf("failed to scan devices: %w", err)
	}

	return toServiceBreakdown(breakdownItems), nil
}

func (r *Repository) GetOperatingSystems(
	ctx context.Context,
	shortCode string,
	from time.Time,
	to time.Time,
) (service.Breakdown, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		WITH grouped AS (
			SELECT
				NULLIF(os, '') AS value,
				COUNT(*) AS clicks
			FROM analytics.clicks
			WHERE short_code = $1
				AND clicked_at >= $2
				AND clicked_at < $3
			GROUP BY NULLIF(os, '')
		),
		ranked AS (
		    SELECT
		        value,
		        clicks,
		        ROW_NUMBER() OVER (ORDER BY clicks DESC, value) AS rn
		    FROM grouped
		    WHERE value IS NOT NULL
		)
		SELECT type, value, clicks
		FROM (
			SELECT
		        'value' AS type,
		        value,
		        clicks,
		        rn AS sort_order
		    FROM ranked
		    WHERE rn <= 5

			UNION ALL

		    SELECT
		        'other',
		        '',
		        SUM(clicks),
		        6
		    FROM ranked
		    WHERE rn > 5
		    HAVING SUM(clicks) > 0

			UNION ALL

			SELECT
				'unknown',
				'',
				clicks,
				7
			FROM grouped
		    WHERE value IS NULL
		) result
		ORDER BY sort_order;
	`

	rows, err := r.pool.Query(ctx, query, shortCode, from, to)
	if err != nil {
		return service.Breakdown{}, fmt.Errorf("query os analytics: %w", err)
	}
	defer rows.Close()

	breakdownItems, err := rowsToBreakdownItemRows(rows)
	if err != nil {
		return service.Breakdown{}, fmt.Errorf("failed to scan os analytics: %w", err)
	}

	return toServiceBreakdown(breakdownItems), nil
}

func (r *Repository) GetBrowsers(
	ctx context.Context,
	shortCode string,
	from time.Time,
	to time.Time,
) (service.Breakdown, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		WITH grouped AS (
			SELECT
				NULLIF(browser, '') AS value,
				COUNT(*) AS clicks
			FROM analytics.clicks
			WHERE short_code = $1
				AND clicked_at >= $2
				AND clicked_at < $3
			GROUP BY NULLIF(browser, '')
		),
		ranked AS (
		    SELECT
		        value,
		        clicks,
		        ROW_NUMBER() OVER (ORDER BY clicks DESC, value) AS rn
		    FROM grouped
		    WHERE value IS NOT NULL
		)
		SELECT type, value, clicks
		FROM (
			SELECT
		        'value' AS type,
		        value,
		        clicks,
		        rn AS sort_order
		    FROM ranked
		    WHERE rn <= 5

			UNION ALL

			SELECT
		        'other',
		        '',
		        SUM(clicks),
		        6
		    FROM ranked
		    WHERE rn > 5
		    HAVING SUM(clicks) > 0

			UNION ALL

			SELECT
				'unknown',
				'',
				clicks,
				7
			FROM grouped
		    WHERE value IS NULL
		) AS result
		ORDER BY sort_order;
	`

	rows, err := r.pool.Query(ctx, query, shortCode, from, to)
	if err != nil {
		return service.Breakdown{}, fmt.Errorf("query browsers analytics: %w", err)
	}
	defer rows.Close()

	breakdownItems, err := rowsToBreakdownItemRows(rows)
	if err != nil {
		return service.Breakdown{}, fmt.Errorf("failed to scan browsers analytics: %w", err)
	}

	return toServiceBreakdown(breakdownItems), nil
}

func (r *Repository) GetReferers(
	ctx context.Context,
	shortCode string,
	from time.Time,
	to time.Time,
) (service.Breakdown, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		WITH grouped AS (
			SELECT
				NULLIF(referer, '') AS value,
				COUNT(*) AS clicks
			FROM analytics.clicks
			WHERE short_code = $1
				AND clicked_at >= $2
				AND clicked_at < $3
			GROUP BY NULLIF(referer, '')
		),
		ranked AS (
		    SELECT
		        value,
		        clicks,
		        ROW_NUMBER() OVER (ORDER BY clicks DESC, value) AS rn
		    FROM grouped
		    WHERE value IS NOT NULL
		)
		SELECT type, value, clicks
		FROM (
			SELECT
		        'value' AS type,
		        value,
		        clicks,
		        rn AS sort_order
		    FROM ranked
		    WHERE rn <= 5

			UNION ALL

			SELECT
		        'other',
		        '',
		        SUM(clicks),
		        6
		    FROM ranked
		    WHERE rn > 5
		    HAVING SUM(clicks) > 0

			UNION ALL

			SELECT
				'unknown',
				'',
				clicks,
				7
			FROM grouped
		    WHERE value IS NULL
		) AS result
		ORDER BY sort_order;
	`

	rows, err := r.pool.Query(ctx, query, shortCode, from, to)
	if err != nil {
		return service.Breakdown{}, fmt.Errorf("query referer analytics: %w", err)
	}
	defer rows.Close()

	breakdownItems, err := rowsToBreakdownItemRows(rows)
	if err != nil {
		return service.Breakdown{}, fmt.Errorf("failed to scan referer analytics: %w", err)
	}

	return toServiceBreakdown(breakdownItems), nil
}

func (r *Repository) GetTimeline(
	ctx context.Context,
	shortCode string,
	from time.Time,
	to time.Time,
	duration time.Duration,
) ([]service.TimelinePoint, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		WITH buckets AS (
		    SELECT generate_series(
		        $2::timestamptz,
		        $3::timestamptz - $4 * INTERVAL '1 microsecond',
		        $4 * INTERVAL '1 microsecond'
		    ) AS bucket_start
		),
		clicks AS (
		    SELECT
		        date_bin(
		            $4 * INTERVAL '1 microsecond',
		            clicked_at,
		            $2::timestamptz
		        ) AS bucket_start,
		        COUNT(*) AS clicks
		    FROM analytics.clicks
		    WHERE short_code = $1
		      AND clicked_at >= $2
		      AND clicked_at < $3
		    GROUP BY bucket_start
		)
		SELECT
		    b.bucket_start,
		    COALESCE(c.clicks, 0) AS clicks
		FROM buckets b
		LEFT JOIN clicks c USING (bucket_start)
		ORDER BY b.bucket_start;
	`

	rows, err := r.pool.Query(ctx, query, shortCode, from, to, duration.Microseconds())
	if err != nil {
		return []service.TimelinePoint{}, fmt.Errorf("query timeline analytics: %w", err)
	}
	defer rows.Close()

	timelineRows := make([]timelineRow, 0)

	for rows.Next() {
		var row timelineRow

		if err := rows.Scan(&row.bucketStart, &row.clicks); err != nil {
			return []service.TimelinePoint{}, fmt.Errorf("failed to scan timeline rows: %w", err)
		}

		timelineRows = append(timelineRows, row)
	}

	if err := rows.Err(); err != nil {
		return []service.TimelinePoint{}, fmt.Errorf("next timeline rows: %w", err)
	}

	return toServiceTimeline(timelineRows), nil
}
