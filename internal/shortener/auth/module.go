package auth

import (
	"github.com/vladislav-koval/url-shortener/internal/platform/authorization"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool"
	"github.com/vladislav-koval/url-shortener/internal/platform/repository/redis"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/identity/google"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/repository/postgres"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/repository/sessionstorage"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/service"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/session"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/transport/authhttp"
)

type Module struct {
	Handler         *authhttp.Handler
	SessionResolver authorization.Resolver
}

func NewModule(pool pool.Pool, cache cache.Pool, cookieSecure bool) *Module {
	httpCfg := authhttp.NewConfigMust()

	userRepository := postgres.NewUserRepository(pool)

	sessionRepository := sessionstorage.NewRepository(cache)
	sessionService := session.NewService(sessionRepository, httpCfg.SessionTTL)

	googleCfg := google.NewConfigMust()

	identityProvider := google.NewProvider(googleCfg)

	authService := service.NewService(
		identityProvider,
		userRepository,
		sessionService,
	)

	authHTTPHandler := authhttp.NewHTTPHandler(authService, httpCfg, cookieSecure)

	return &Module{
		Handler:         authHTTPHandler,
		SessionResolver: sessionService.Resolve,
	}
}
