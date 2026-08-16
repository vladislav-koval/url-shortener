package service

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	BannedDomains []string `envconfig:"BANNED_DOMAINS"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("SHORTENER", &config); err != nil {
		return Config{}, fmt.Errorf("failed to process env shortener service config: %w", err)
	}
	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get shortener service config: %w", err))
	}
	return config
}
