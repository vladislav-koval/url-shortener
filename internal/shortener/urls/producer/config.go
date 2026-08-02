package producer

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Topic         string
	BatchSize     int
	QueueSize     int
	FlushInterval time.Duration
	WriteTimeout  time.Duration
}

type topicConfig struct {
	Topic string `envconfig:"TOPIC" required:"true"`
}

type producerOnlyConfig struct {
	BatchSize     int           `envconfig:"BATCH_SIZE" required:"true"`
	QueueSize     int           `envconfig:"QUEUE_SIZE" required:"true"`
	FlushInterval time.Duration `envconfig:"FLUSH_INTERVAL" required:"true"`
	WriteTimeout  time.Duration `envconfig:"WRITE_TIMEOUT" required:"true"`
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
		Topic:         topic.Topic,
		BatchSize:     producerCfg.BatchSize,
		QueueSize:     producerCfg.QueueSize,
		FlushInterval: producerCfg.FlushInterval,
		WriteTimeout:  producerCfg.WriteTimeout,
	}, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get shortener producer config: %w", err))
	}

	return config
}
