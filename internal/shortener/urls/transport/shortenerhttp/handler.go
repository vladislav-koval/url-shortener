package shortenerhttp

import (
	"context"

	"github.com/google/uuid"
	"github.com/vladislav-koval/url-shortener/internal/platform/authorization"
	"github.com/vladislav-koval/url-shortener/internal/platform/geo"
	"github.com/vladislav-koval/url-shortener/internal/platform/messaging/events"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/middleware"
	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/server"
	"github.com/vladislav-koval/url-shortener/internal/shortener/urls/domain"
)

type Handler struct {
	shortenerService Service
	geoResolver      geo.Resolver
}

type Service interface {
	CreateShortLink(ctx context.Context, originalURL string, userID *uuid.UUID) (domain.Link, error)
	ResolveShortLink(ctx context.Context, code string, event events.ClickEvent) (string, error)
}

func NewHTTPHandler(shortenerService Service, geoResolver geo.Resolver) *Handler {
	return &Handler{
		shortenerService: shortenerService,
		geoResolver:      geoResolver,
	}
}

func (h *Handler) RedirectRoute() []server.Route {
	return []server.Route{
		{
			Method:  "GET",
			Path:    "/{code}",
			Handler: h.Redirect,
		},
	}
}

func (h *Handler) APIRoutes(resolver authorization.Resolver, cookieSecure bool) []server.Route {
	return []server.Route{
		{
			Method:     "POST",
			Path:       "/link",
			Handler:    h.CreateShortLink,
			Middleware: []middleware.Middleware{middleware.CurrentUser(resolver, cookieSecure)},
		},
	}
}
