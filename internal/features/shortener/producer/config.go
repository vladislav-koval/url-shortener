package producer

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Topic        string
	BatchSize    int
	BatchTimeout time.Duration
}

type topicConfig struct {
	Topic string `envconfig:"TOPIC" required:"true"`
}

type producerOnlyConfig struct {
	BatchSize    int           `envconfig:"BATCH_SIZE" required:"true"`
	BatchTimeout time.Duration `envconfig:"BATCH_TIMEOUT" required:"true"`
}

func NewConfig() (Config, error) {
	var topic topicConfig
	if err := envconfig.Process("KAFKA", &topic); err != nil {
		return Config{}, fmt.Errorf("failed to process shared kafka config: %w", err)
	}

	var producerCfg producerOnlyConfig
	if err := envconfig.Process("KAFKA_PRODUCER", &producerCfg); err != nil {
		return Config{}, fmt.Errorf("failed to process env shortener producer config: %w", err)
	}

	return Config{
		Topic:        topic.Topic,
		BatchSize:    producerCfg.BatchSize,
		BatchTimeout: producerCfg.BatchTimeout,
	}, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get shortener producer config: %w", err))
	}

	return config
}
