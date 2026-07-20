package recorder

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Topic        string        `envconfig:"TOPIC" required:"true"`
	BatchSize    int           `envconfig:"BATCH_SIZE" required:"true"`
	BatchTimeout time.Duration `envconfig:"BATCH_TIMEOUT" required:"true"`
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("KAFKA", &config); err != nil {
		return Config{}, fmt.Errorf("failed to process env kafka recorder config: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get kafka recorder config: %w", err))
	}

	return config
}
