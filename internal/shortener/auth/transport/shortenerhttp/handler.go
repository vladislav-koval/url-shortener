package shortenerhttp

import (
	"context"
	"time"

	"golang.org/x/oauth2"

	"github.com/vladislav-koval/url-shortener/internal/platform/transport/http/server"
)

type Handler struct {
	authService         Service
	googleOAuth         *oauth2.Config
	cookieSecure        bool
	successRedirectURL  string
	sessionCookieMaxAge time.Duration
}

type Service interface {
	LoginWithGoogle(ctx context.Context, code string, verifier string) (sessionToken string, err error)
}

func NewHTTPHandler(authService Service, googleOAuth *oauth2.Config, cfg Config) *Handler {
	return &Handler{
		authService:         authService,
		googleOAuth:         googleOAuth,
		cookieSecure:        cfg.CookieSecure,
		successRedirectURL:  cfg.SuccessRedirectURL,
		sessionCookieMaxAge: cfg.SessionCookieTTL,
	}
}

func (h *Handler) Routes() []server.Route {
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
	}
}
