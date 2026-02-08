package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port             string   `env:"PORT" envDefault:"8080"`
	DBURL            string   `env:"DB_URL,required"`
	JWTSecret        string   `env:"JWT_SECRET,required"`
	JWTRefreshSecret string   `env:"JWT_REFRESH_SECRET,required"`
	GoEnv            string   `env:"GO_ENV" envDefault:"development"`
	AllowedOrigins   []string `env:"ALLOWED_ORIGINS" envDefault:"*"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}
