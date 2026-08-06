package authorization

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	CookieSecure bool `envconfig:"COOKIE_SECURE" required:"true"`
}

func NewConfig() (Config, error) {
	var cfg Config
	if err := envconfig.Process("AUTHORIZATION", &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to process env authorization config: %w", err)
	}

	return cfg, nil
}

func NewConfigMust() Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get authorization config: %w", err))
	}

	return cfg
}
