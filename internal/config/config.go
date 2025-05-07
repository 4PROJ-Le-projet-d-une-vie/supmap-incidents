package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

var UsersBaseUrl string

type Config struct {
	ENV             string `env:"ENV" envDefault:"production"`
	DbUrl           string `env:"DB_URL"`
	PORT            string `env:"PORT"`
	UsersBaseUrl    string `env:"USERS_BASE_URL"`
	RedisHost       string `env:"REDIS_HOST"`
	RedisPort       string `env:"REDIS_PORT"`
	IncidentChannel string `env:"REDIS_INCIDENTS_CHANNEL" envDefault:"incidents"`
}

func New() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	exposeUrls(&cfg)
	return &cfg, nil
}

func exposeUrls(config *Config) {
	UsersBaseUrl = config.UsersBaseUrl
}
