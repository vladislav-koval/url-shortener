package http

import (
	"context"

	"github.com/vladislav-koval/url-shortener/internal/core/domain"
	"github.com/vladislav-koval/url-shortener/internal/core/transport/http/server"
)

type Handler struct {
	shortenerService Service
}

type Service interface {
	CreateShortLink(ctx context.Context, originalURL string) (domain.Link, error)
	ResolveShortLink(ctx context.Context, code string) (string, error)
}

func NewHTTPHandler(shortenerService Service) *Handler {
	return &Handler{
		shortenerService: shortenerService,
	}
}

func (h *Handler) Routes() []server.Route {
	return []server.Route{
		{
			Method:  "POST",
			Path:    "/link",
			Handler: h.CreateShortLink,
		},
		{
			Method:  "GET",
			Path:    "/{code}",
			Handler: h.Redirect,
		},
	}
}
