package cached

import (
	"context"
	"errors"

	"github.com/vladislav-koval/url-shortener/internal/core/domain"
	"github.com/vladislav-koval/url-shortener/internal/core/logger"
	cache "github.com/vladislav-koval/url-shortener/internal/core/repository/redis"
	"go.uber.org/zap"
)

func (r *Repository) GetByShortCode(ctx context.Context, code string) (domain.Link, error) {
	log := logger.FromContext(ctx)

	originalURL, err := r.getCache(ctx, code)

	if err == nil {
		return domain.NewLink(code, originalURL), nil
	}

	if !errors.Is(err, cache.ErrNotFound) {
		log.Error("read from cache", zap.Error(err))
	}

	domainLink, err := r.mainRepository.GetByShortCode(ctx, code)
	if err != nil {
		return domain.Link{}, err
	}

	if err := r.setCache(ctx, code, domainLink.OriginalURL); err != nil {
		log.Error("cache set", zap.Error(err))
	}

	return domainLink, nil
}
