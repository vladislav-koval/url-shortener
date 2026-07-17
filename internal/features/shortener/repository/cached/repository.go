package cached

import (
	"context"
	"time"

	"github.com/vladislav-koval/url-shortener/internal/core/domain"
	cache "github.com/vladislav-koval/url-shortener/internal/core/repository/redis"
)

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
