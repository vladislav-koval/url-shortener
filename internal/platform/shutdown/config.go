package shutdown

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Timeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"15s"`
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("", &config); err != nil {
		return Config{}, fmt.Errorf("failed to process env shutdown config: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get shutdown config: %w", err))
	}

	return config
}
