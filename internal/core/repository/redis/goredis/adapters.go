package goredis

import (
	"errors"

	"github.com/redis/go-redis/v9"
	cache "github.com/vladislav-koval/url-shortener/internal/core/repository/redis"
)

type goredisStringCmd struct {
	*redis.StringCmd
}

func (cmd goredisStringCmd) Result() (string, error) {
	str, err := cmd.StringCmd.Result()

	if err != nil {
		return "", mapErrors(err)
	}

	return str, nil
}

type goredisStatusCmd struct {
	*redis.StatusCmd
}

func mapErrors(err error) error {
	if errors.Is(err, redis.Nil) {
		return cache.ErrNotFound
	}

	return err
}
