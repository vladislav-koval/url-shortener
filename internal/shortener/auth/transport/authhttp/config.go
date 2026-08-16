package authhttp

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	FrontendURL string
	SessionTTL  time.Duration
}

type frontendConfig struct {
	URL string `envconfig:"FRONTEND_URL" required:"true"`
}

type sessionTTLConfig struct {
	TTL time.Duration `envconfig:"TTL" required:"true"`
}

func NewConfig() (Config, error) {
	var frontendCfg frontendConfig
	if err := envconfig.Process("", &frontendCfg); err != nil {
		return Config{}, fmt.Errorf("failed to process env frontend config: %w", err)
	}

	var sessionCfg sessionTTLConfig
	if err := envconfig.Process("AUTH_SESSION", &sessionCfg); err != nil {
		return Config{}, fmt.Errorf("failed to process shared auth sessionstorage config: %w", err)
	}

	return Config{
		FrontendURL: frontendCfg.URL,
		SessionTTL:  sessionCfg.TTL,
	}, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get auth http config: %w", err))
	}

	return config
}
