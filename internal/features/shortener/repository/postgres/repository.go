package postgres

import "github.com/vladislav-koval/url-shortener/internal/core/repository/postgres/pool"

type Repository struct {
	pool pool.Pool
}

func NewShortenerRepository(pool pool.Pool) *Repository {
	return &Repository{
		pool,
	}
}
