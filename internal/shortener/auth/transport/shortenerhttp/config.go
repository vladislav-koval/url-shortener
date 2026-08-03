package shortenerhttp

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	CookieSecure       bool
	SuccessRedirectURL string
	SessionCookieTTL   time.Duration
}

type googleOAuthConfig struct {
	ClientID     string `envconfig:"GOOGLE_CLIENT_ID" required:"true"`
	ClientSecret string `envconfig:"GOOGLE_CLIENT_SECRET" required:"true"`
	RedirectURL  string `envconfig:"GOOGLE_REDIRECT_URL" required:"true"`
}

type httpOnlyConfig struct {
	CookieSecure       bool   `envconfig:"COOKIE_SECURE" required:"true"`
	SuccessRedirectURL string `envconfig:"SUCCESS_REDIRECT_URL" required:"true"`
}

// sessionTTLConfig — то же AUTH_SESSION_TTL, что читает session.Config, тем же
// приёмом, что и общий KAFKA_TOPIC для producer/consumer: один env-ключ, два
// независимых NewConfig(), значение не может разъехаться между тем, что реально
// подписывает Issuer, и тем, на сколько живёт cookie в браузере.
type sessionTTLConfig struct {
	TTL time.Duration `envconfig:"TTL" required:"true"`
}

func NewConfig() (Config, error) {
	var google googleOAuthConfig
	if err := envconfig.Process("AUTH", &google); err != nil {
		return Config{}, fmt.Errorf("failed to process env google oauth config: %w", err)
	}

	var httpCfg httpOnlyConfig
	if err := envconfig.Process("AUTH", &httpCfg); err != nil {
		return Config{}, fmt.Errorf("failed to process env auth http config: %w", err)
	}

	var sessionCfg sessionTTLConfig
	if err := envconfig.Process("AUTH_SESSION", &sessionCfg); err != nil {
		return Config{}, fmt.Errorf("failed to process shared auth session config: %w", err)
	}

	return Config{
		GoogleClientID:     google.ClientID,
		GoogleClientSecret: google.ClientSecret,
		GoogleRedirectURL:  google.RedirectURL,
		CookieSecure:       httpCfg.CookieSecure,
		SuccessRedirectURL: httpCfg.SuccessRedirectURL,
		SessionCookieTTL:   sessionCfg.TTL,
	}, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get auth http config: %w", err))
	}

	return config
}
