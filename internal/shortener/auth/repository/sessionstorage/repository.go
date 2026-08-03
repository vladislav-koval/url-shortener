package sessionstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/redis"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/domain"
	"go.uber.org/zap"
)

type Repository struct {
	cache cache.Pool
}

func NewRepository(cache cache.Pool) *Repository {
	return &Repository{
		cache: cache,
	}
}

const keyPrefix = "shortener:session:"

func cacheKey(tokenHash string) string {
	return keyPrefix + tokenHash
}

func (r *Repository) Save(
	ctx context.Context,
	tokenHash string,
	session domain.Session,
	ttl time.Duration,
) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshal sessionstorage: %w", err)
	}

	if err := r.cache.Set(ctx, cacheKey(tokenHash), data, ttl).Err(); err != nil {
		return fmt.Errorf("save sessionstorage in redis: %w", err)
	}

	return nil
}

func (r *Repository) Delete(
	ctx context.Context,
	tokenHash string,
) {
	log := logger.FromContext(ctx)

	err := r.cache.Del(ctx, cacheKey(tokenHash)).Err()

	if err != nil {
		log.Error("delete session token", zap.Error(err))
	}
}
