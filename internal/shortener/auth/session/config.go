package session

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Secret string
	TTL    time.Duration
}

type sessionConfig struct {
	Secret string        `envconfig:"SECRET" required:"true"`
	TTL    time.Duration `envconfig:"TTL" required:"true"`
}

func NewConfig() (Config, error) {
	var cfg sessionConfig
	if err := envconfig.Process("AUTH_SESSION", &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to process env auth session config: %w", err)
	}

	return Config{
		Secret: cfg.Secret,
		TTL:    cfg.TTL,
	}, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get auth session config: %w", err))
	}

	return config
}
