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
	UsersHost       string `env:"SUPMAP_USERS_HOST"`
	UsersPort       string `env:"SUPMAP_USERS_PORT"`
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
	UsersBaseUrl = fmt.Sprintf("http://%s:%s", config.UsersHost, config.UsersPort)
}
