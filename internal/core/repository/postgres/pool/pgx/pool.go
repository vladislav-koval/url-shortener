package pgx

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vladislav-koval/url-shortener/internal/core/repository/postgres/pool"
)

type Pool struct {
	*pgxpool.Pool
	opTimeout time.Duration
}

func NewPool(ctx context.Context, config Config) (*Pool, error) {
	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	pgxconfig, err := pgxpool.ParseConfig(connectionString)

	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx config: %w", err)
	}

	pgxPool, err := pgxpool.NewWithConfig(ctx, pgxconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	err = pgxPool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping pgx pool: %w", err)
	}

	return &Pool{
		Pool:      pgxPool,
		opTimeout: config.Timeout,
	}, nil
}

func (p *Pool) Query(
	ctx context.Context,
	sql string,
	args ...any,
) (pool.Rows, error) {
	rows, err := p.Pool.Query(ctx, sql, args...)

	if err != nil {
		return nil, mapErrors(err)
	}

	return pgxRows{rows}, nil
}

func (p *Pool) QueryRow(
	ctx context.Context,
	sql string,
	args ...any,
) pool.Row {
	row := p.Pool.QueryRow(ctx, sql, args...)

	return pgxRow{row}
}
func (p *Pool) Exec(
	ctx context.Context,
	sql string,
	arguments ...any,
) (pool.CommandTag, error) {
	ct, err := p.Pool.Exec(ctx, sql, arguments...)

	if err != nil {
		return nil, mapErrors(err)
	}

	return commandTag{ct}, nil
}

func (p *Pool) OpTimeout() time.Duration {
	return p.opTimeout
}
