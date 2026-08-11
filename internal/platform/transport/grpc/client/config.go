package grpcclient

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr       string        `envconfig:"ADDR" required:"true"`
	MaxRetries uint          `envconfig:"MAX_RETRIES" default:"3"`
	PerTimeout time.Duration `envconfig:"PER_TIMEOUT" default:"1s"`
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("GRPC_CLIENT", &config); err != nil {
		return Config{}, fmt.Errorf("failed to process env grpc client config: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()

	if err != nil {
		err = fmt.Errorf("get GRPC client config: %w", err)
		panic(err)
	}

	return config
}
