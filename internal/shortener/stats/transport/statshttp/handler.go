package statshttp

import (
	"context"

	"github.com/google/uuid"
	"github.com/vladislav-koval/url-shortener/internal/platform/authorization"
	"github.com/vladislav-koval/url-shortener/internal/platform/pagination"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/middleware"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/server"
	"github.com/vladislav-koval/url-shortener/internal/shortener/stats/domain"
)

type Handler struct {
	statsService Service
}

type Service interface {
	GetClickCounts(ctx context.Context, userID uuid.UUID, pagination pagination.Pagination) (domain.LinkClickCountPage, error)
}

func NewHTTPHandler(s Service) *Handler {
	return &Handler{
		statsService: s,
	}
}

func (h *Handler) Routes(resolver authorization.Resolver, cookieSecure bool) []server.Route {
	return []server.Route{
		{
			Method:     "GET",
			Path:       "/clicks",
			Handler:    h.GetClickCounts,
			Middleware: []middleware.Middleware{middleware.CurrentUser(resolver, cookieSecure), middleware.RequireUser()},
		},
	}
}
