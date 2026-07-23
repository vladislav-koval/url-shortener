package goredis

import (
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	cache "github.com/vladislav-koval/url-shortener/internal/core/repository/redis"
)

func TestGoredisStringCmd_Result(t *testing.T) {
	testCases := []struct {
		name    string
		realCmd *redis.StringCmd
		check   func(t *testing.T, val string, err error)
	}{
		{
			name:    "maps redis.Nil to cache.ErrNotFound",
			realCmd: redis.NewStringResult("", redis.Nil),
			check: func(t *testing.T, val string, err error) {
				assert.ErrorIs(t, err, cache.ErrNotFound)
				assert.Empty(t, val)
			},
		},
		{
			name:    "passes through a value on success",
			realCmd: redis.NewStringResult("http://google.com", nil),
			check: func(t *testing.T, val string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "http://google.com", val)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			cmd := goredisStringCmd{tt.realCmd}

			val, err := cmd.Result()

			tt.check(t, val, err)
		})
	}
}

func TestMapErrors(t *testing.T) {
	testCases := []struct {
		name      string
		actualErr error
		check     func(t *testing.T, err error)
	}{
		{
			name:      "maps redis.Nil to cache.ErrNotFound",
			actualErr: redis.Nil,
			check: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, cache.ErrNotFound)
			},
		},
		{
			name:      "passes another error through unchanged",
			actualErr: errors.New("connection refused"),
			check: func(t *testing.T, err error) {
				assert.NotErrorIs(t, err, cache.ErrNotFound)
				assert.EqualError(t, err, "connection refused")
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := mapErrors(tt.actualErr)

			tt.check(t, err)
		})
	}
}
