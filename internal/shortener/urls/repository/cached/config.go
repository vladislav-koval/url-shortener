package cached

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	TTL time.Duration `envconfig:"CACHE_TTL" default:"24h"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("SHORTENER", &config); err != nil {
		return Config{}, fmt.Errorf("failed to process env shortener cache config: %w", err)
	}
	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get shortener cache config: %w", err))
	}
	return config
}
