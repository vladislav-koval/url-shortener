package cached

import (
	cache "github.com/vladislav-koval/url-shortener/internal/core/repository/redis"
	"github.com/vladislav-koval/url-shortener/internal/features/shortener/service"
)

type Repository struct {
	cache          cache.Store
	mainRepository service.Repository
}

func NewRepository(cache cache.Store, mainRepository service.Repository) *Repository {
	return &Repository{
		cache:          cache,
		mainRepository: mainRepository,
	}
}
