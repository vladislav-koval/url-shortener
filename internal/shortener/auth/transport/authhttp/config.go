package authhttp

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	CookieSecure       bool
	SuccessRedirectURL string
	SessionTTL         time.Duration
}

type httpOnlyConfig struct {
	CookieSecure       bool   `envconfig:"COOKIE_SECURE" required:"true"`
	SuccessRedirectURL string `envconfig:"SUCCESS_REDIRECT_URL" required:"true"`
}

type sessionTTLConfig struct {
	TTL time.Duration `envconfig:"TTL" required:"true"`
}

func NewConfig() (Config, error) {
	var httpCfg httpOnlyConfig
	if err := envconfig.Process("AUTH", &httpCfg); err != nil {
		return Config{}, fmt.Errorf("failed to process env auth http config: %w", err)
	}

	var sessionCfg sessionTTLConfig
	if err := envconfig.Process("AUTH_SESSION", &sessionCfg); err != nil {
		return Config{}, fmt.Errorf("failed to process shared auth sessionstorage config: %w", err)
	}

	return Config{
		CookieSecure:       httpCfg.CookieSecure,
		SuccessRedirectURL: httpCfg.SuccessRedirectURL,
		SessionTTL:         sessionCfg.TTL,
	}, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get auth http config: %w", err))
	}

	return config
}
