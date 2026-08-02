package cached

import (
	"context"

	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/domain"
	"go.uber.org/zap"
)

func (r *Repository) CreateShortLink(ctx context.Context, shortCode string, originalURL string) (domain.Link, error) {
	log := logger.FromContext(ctx)

	domainLink, err := r.mainRepository.CreateShortLink(ctx, shortCode, originalURL)
	if err != nil {
		return domain.Link{}, err
	}

	if err := r.setCache(ctx, domainLink.ShortCode, domainLink.OriginalURL); err != nil {
		log.Error("cache set", zap.Error(err))
	}

	return domainLink, nil
}
