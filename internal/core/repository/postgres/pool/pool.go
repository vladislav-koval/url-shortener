package pool

import (
	"context"
	"time"
)

type Pool interface {
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
	Exec(ctx context.Context, sql string, arguments ...any) (CommandTag, error)
	Close()
	OpTimeout() time.Duration
}

type Rows interface {
	Close()
	Next() bool
	Scan(dest ...any) error
	Err() error
}

type Row interface {
	Scan(dest ...any) error
}

type CommandTag interface {
	RowsAffected() int64
}
