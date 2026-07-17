package cache

import (
	"context"
	"time"
)

type Store interface {
	Get(ctx context.Context, key string) StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) StatusCmd
	Close() error

	TTL() time.Duration
}

type StringCmd interface {
	Result() (string, error)
}
type StatusCmd interface {
	Err() error
}
