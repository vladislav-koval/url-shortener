package shortenerhttp

import (
	"context"

	"github.com/vladislav-koval/url-shortener/internal/platform/authorization"
	"github.com/vladislav-koval/url-shortener/internal/platform/messaging/events"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/middleware"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/server"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/domain"
)

type Handler struct {
	shortenerService Service
}

type Service interface {
	CreateShortLink(ctx context.Context, originalURL string) (domain.Link, error)
	ResolveShortLink(ctx context.Context, code string, event events.ClickEvent) (string, error)
}

func NewHTTPHandler(shortenerService Service) *Handler {
	return &Handler{
		shortenerService: shortenerService,
	}
}

func (h *Handler) Routes(resolver authorization.Resolver) []server.Route {
	return []server.Route{
		{
			Method:     "POST",
			Path:       "/link",
			Handler:    h.CreateShortLink,
			Middleware: []middleware.Middleware{middleware.CurrentUser(resolver)},
		},
		{
			Method:  "GET",
			Path:    "/{code}",
			Handler: h.Redirect,
		},
	}
}
