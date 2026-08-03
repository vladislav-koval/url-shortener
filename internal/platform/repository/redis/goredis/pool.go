package goredis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	cache "github.com/vladislav-koval/url-shortener/internal/platform/repository/redis"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(ctx context.Context, config Config) (*Redis, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &Redis{client: rdb}, nil
}

func (c *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) cache.StatusCmd {
	cmd := c.client.Set(ctx, key, value, expiration)
	return goredisStatusCmd{cmd}
}

func (c *Redis) Get(ctx context.Context, key string) cache.StringCmd {
	cmd := c.client.Get(ctx, key)
	return goredisStringCmd{cmd}
}

func (c *Redis) Del(ctx context.Context, keys ...string) cache.IntCmd {
	cmd := c.client.Del(ctx, keys...)
	return goredisIntCmd{cmd}
}

func (c *Redis) Close() error {
	return c.client.Close()
}
