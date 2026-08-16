package authhttp

import (
	"context"
	"time"

	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/server"
)

type Handler struct {
	authService  Service
	cookieSecure bool
	frontendURL  string
	cookieTTL    time.Duration
}

type Service interface {
	LoginWithGoogle(ctx context.Context, code string, verifier string) (sessionToken string, err error)
	AuthCodeURL(state, verifier string) string
	Logout(ctx context.Context, rawToken string)
}

func NewHTTPHandler(authService Service, cfg Config, cookieSecure bool) *Handler {
	return &Handler{
		authService:  authService,
		cookieSecure: cookieSecure,
		frontendURL:  cfg.FrontendURL,
		cookieTTL:    cfg.SessionTTL,
	}
}

func (h *Handler) APIRoutes() []server.Route {
	return []server.Route{
		{
			Method:  "GET",
			Path:    "/auth/google/login",
			Handler: h.GoogleLogin,
		},
		{
			Method:  "GET",
			Path:    "/auth/google/callback",
			Handler: h.GoogleCallback,
		},
		{
			Method:  "POST",
			Path:    "/auth/logout",
			Handler: h.Logout,
		},
	}
}
