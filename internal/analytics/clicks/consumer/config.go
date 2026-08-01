package consumer

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Topic           string
	BatchSize       int
	BatchTimeout    time.Duration
	GoroutinesCount int
}

type topicConfig struct {
	Topic string `envconfig:"TOPIC" required:"true"`
}

type consumerOnlyConfig struct {
	BatchSize       int           `envconfig:"BATCH_SIZE" required:"true"`
	BatchTimeout    time.Duration `envconfig:"BATCH_TIMEOUT" required:"true"`
	GoroutinesCount int           `envconfig:"GOROUTINES_COUNT" required:"true"`
}

func NewConfig() (Config, error) {
	var topic topicConfig
	if err := envconfig.Process("KAFKA", &topic); err != nil {
		return Config{}, fmt.Errorf("failed to process shared kafka config: %w", err)
	}

	var consumerCfg consumerOnlyConfig
	if err := envconfig.Process("KAFKA_CONSUMER", &consumerCfg); err != nil {
		return Config{}, fmt.Errorf("failed to process env analytics consumer config: %w", err)
	}

	return Config{
		Topic:           topic.Topic,
		BatchSize:       consumerCfg.BatchSize,
		BatchTimeout:    consumerCfg.BatchTimeout,
		GoroutinesCount: consumerCfg.GoroutinesCount,
	}, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get analytics consumer config: %w", err))
	}

	return config
}
