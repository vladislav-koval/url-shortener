package cached

import (
	"context"
	"time"

	cache "github.com/vladislav-koval/url-shortener/internal/platform/repository/redis"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/domain"
)

//go:generate mockgen -source=./repository.go -destination=mocks/mock_repository.go -package=mocks
type UnderlyingRepository interface {
	CreateShortLink(ctx context.Context, shortCode string, originalURL string) (domain.Link, error)
	GetByShortCode(ctx context.Context, code string) (domain.Link, error)
}

type Repository struct {
	cache          cache.Pool
	ttl            time.Duration
	mainRepository UnderlyingRepository
}

func NewRepository(cache cache.Pool, cfg Config, mainRepository UnderlyingRepository) *Repository {
	return &Repository{
		cache:          cache,
		ttl:            cfg.TTL,
		mainRepository: mainRepository,
	}
}

const keyPrefix = "shortener:url:"

func cacheKey(shortCode string) string {
	return keyPrefix + shortCode
}

func (r *Repository) setCache(ctx context.Context, shortCode string, originalURL string) error {
	return r.cache.Set(ctx, cacheKey(shortCode), originalURL, r.ttl).Err()
}

func (r *Repository) getCache(ctx context.Context, shortCode string) (string, error) {
	return r.cache.Get(ctx, cacheKey(shortCode)).Result()
}
