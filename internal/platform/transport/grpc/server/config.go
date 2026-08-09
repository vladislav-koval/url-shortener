package server

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr string `envconfig:"ADDR" required:"true"`
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("GRPC", &config); err != nil {
		return Config{}, fmt.Errorf("failed to process env grpc server config: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()

	if err != nil {
		err = fmt.Errorf("get GRPC server config: %w", err)
		panic(err)
	}

	return config
}
