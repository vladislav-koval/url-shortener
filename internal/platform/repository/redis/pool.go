package cache

import (
	"context"
	"time"
)

//go:generate mockgen -source=./pool.go -destination=mocks/mock_pool.go -package=mocks
type Pool interface {
	Get(ctx context.Context, key string) StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) StatusCmd
	Close() error
}

type StringCmd interface {
	Result() (string, error)
}

type StatusCmd interface {
	Err() error
}
