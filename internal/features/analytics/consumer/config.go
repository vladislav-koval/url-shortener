package consumer

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Topic          string        `envconfig:"TOPIC" required:"true"`
	BatchSize      int           `envconfig:"BATCH_SIZE" required:"true"`
	BatchTimeout   time.Duration `envconfig:"BATCH_TIMEOUT" required:"true"`
	GorutinesCount int           `envconfig:"GORUTINES_COUNT" required:"true"`
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("KAFKA_CONSUMER", &config); err != nil {
		return Config{}, fmt.Errorf("failed to process env analytics consumer config: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get analytics consumer config: %w", err))
	}

	return config
}
