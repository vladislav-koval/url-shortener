package sessionstorage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/vladislav-koval/url-shortener/internal/platform/apperrors"
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
		return fmt.Errorf("marshal session: %w", err)
	}

	if err := r.cache.Set(ctx, cacheKey(tokenHash), data, ttl).Err(); err != nil {
		return fmt.Errorf("save session in cache: %w", err)
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

func (r *Repository) Get(ctx context.Context, tokenHash string) (domain.Session, error) {
	data, err := r.cache.Get(ctx, cacheKey(tokenHash)).Result()

	if errors.Is(err, cache.ErrNotFound) {
		return domain.Session{}, fmt.Errorf("session not found: %w", apperrors.ErrUnauthenticated)
	}

	if err != nil {
		return domain.Session{}, fmt.Errorf("get session from cache: %w", err)
	}

	var session domain.Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return domain.Session{}, fmt.Errorf("unmarshal session: %w", err)
	}

	return session, nil
}
