package goredis

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host     string `envconfig:"HOST" required:"true"`
	Port     string `envconfig:"PORT" default:"6379"`
	Password string `envconfig:"PASSWORD" required:"true"`
	DB       int    `envconfig:"DB" default:"0"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("REDIS", &config); err != nil {
		return Config{}, fmt.Errorf("failed to process env redis config: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()

	if err != nil {
		err = fmt.Errorf("get Redis config: %w", err)
		panic(err)
	}

	return config
}
