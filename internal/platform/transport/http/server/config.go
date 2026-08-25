package server

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr           string   `envconfig:"ADDR" required:"true"`
	AllowedOrigins []string `envconfig:"ALLOWED_ORIGINS" required:"true"`

	ReadHeaderTimeout time.Duration `envconfig:"READ_HEADER_TIMEOUT" default:"5s"`
	ReadTimeout       time.Duration `envconfig:"READ_TIMEOUT" default:"15s"`
	WriteTimeout      time.Duration `envconfig:"WRITE_TIMEOUT" default:"30s"`
	IdleTimeout       time.Duration `envconfig:"IDLE_TIMEOUT" default:"120s"`
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("HTTP", &config); err != nil {
		return Config{}, fmt.Errorf("failed to process env server config: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()

	if err != nil {
		err = fmt.Errorf("get HTTPServer config: %w", err)
		panic(err)
	}

	return config
}
