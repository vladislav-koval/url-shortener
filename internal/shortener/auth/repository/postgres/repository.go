package postgres

import "github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool"

type Repository struct {
	pool pool.Pool
}

func NewUserRepository(pool pool.Pool) *Repository {
	return &Repository{
		pool,
	}
}
