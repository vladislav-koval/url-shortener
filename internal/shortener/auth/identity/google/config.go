package google

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GoogleClientID     string `envconfig:"CLIENT_ID" required:"true"`
	GoogleClientSecret string `envconfig:"CLIENT_SECRET" required:"true"`
	GoogleCallbackURL  string `envconfig:"CALLBACK_URL" required:"true"`

	ExchangeTimeout time.Duration `envconfig:"EXCHANGE_TIMEOUT" default:"10s"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("GOOGLE_AUTH", &config); err != nil {
		return Config{}, fmt.Errorf("failed to process env google auth provider: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get google auth provider config: %w", err))
	}
	return config
}
