package geo

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	FilePath string `envconfig:"FILE_PATH" required:"true"`
	Required bool   `envconfig:"REQUIRED" default:"false"`
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("GEO", &config); err != nil {
		return Config{}, fmt.Errorf("failed to process env geo config: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get geo config: %w", err))
	}

	return config
}
