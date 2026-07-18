package segmentio

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Brokers []string `envconfig:"BROKERS" required:"true"`
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("KAFKA", &config); err != nil {
		return Config{}, fmt.Errorf("failed to process env kafka config: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get kafka config: %w", err))
	}

	return config
}
