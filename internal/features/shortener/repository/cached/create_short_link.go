package cached

import (
	"context"

	"github.com/vladislav-koval/url-shortener/internal/core/domain"
	"github.com/vladislav-koval/url-shortener/internal/core/logger"
	"go.uber.org/zap"
)

func (r *Repository) CreateShortLink(ctx context.Context, shortCode string, originalURL string) (domain.Link, error) {
	log := logger.FromContext(ctx)

	domainLink, err := r.mainRepository.CreateShortLink(ctx, shortCode, originalURL)
	if err != nil {
		return domain.Link{}, err
	}

	err = r.cache.Set(ctx, domainLink.ShortCode, domainLink.OriginalURL, r.cache.TTL()).Err()

	if err != nil {
		log.Error("cache set", zap.Error(err))
	}

	return domainLink, nil
}
