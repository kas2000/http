package http

import (
	log "github.com/kas2000/logger"
	"time"
)

type Config struct {
	Addr            string        `envconfig:"addr" mapstructure:"addr" default:":8080"`
	ShutdownTimeout time.Duration `envconfig:"shutdown_timeout" mapstructure:"shutdown_timeout" default:"20"`
	GracefulTimeout time.Duration `envconfig:"graceful_timeout" mapstructure:"graceful_timeout" default:"21"`
	ApiVersion      string        `envconfig:"api_version" mapstructure:"api_version" default:"v1"`
	Timeout         time.Duration `envconfig:"timeout" mapstructure:"timeout" default:"20"`
	Logger          log.Logger
}

func Listen(config Config) error {


	return nil
}