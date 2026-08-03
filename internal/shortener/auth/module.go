package auth

import (
	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"

	"github.com/vladislav-koval/url-shortener/internal/platform/repository/postgres/pool"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/identity/google"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/repository/postgres"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/service"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/session"
	"github.com/vladislav-koval/url-shortener/internal/shortener/auth/transport/shortenerhttp"
)

type Module struct {
	Handler *shortenerhttp.Handler
}

func NewModule(pool pool.Pool) *Module {
	httpCfg := shortenerhttp.NewConfigMust()
	sessionCfg := session.NewConfigMust()

	googleOAuth := &oauth2.Config{
		ClientID:     httpCfg.GoogleClientID,
		ClientSecret: httpCfg.GoogleClientSecret,
		RedirectURL:  httpCfg.GoogleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     googleoauth.Endpoint,
	}

	authRepository := postgres.NewRepository(pool)
	identityProvider := google.NewProvider(googleOAuth)
	sessionIssuer := session.NewIssuer(sessionCfg)

	authService := service.NewService(authRepository, identityProvider, sessionIssuer)
	authHTTPHandler := shortenerhttp.NewHTTPHandler(authService, googleOAuth, httpCfg)

	return &Module{
		Handler: authHTTPHandler,
	}
}
