package config

import (
	"fmt"
	"time"

	envconfig "github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	AdminToken  string `env:"ADMIN_TOKEN,required"`
	HTTPServer  HTTPServerConfig
	DatabaseURL string `env:"DATABASE_URL,required"`
	StoragePath string `env:"STORAGE_PATH" envDefault:"./data/files"`
	Redis       RedisConfig
}

type RedisConfig struct {
	Addr     string        `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	Password string        `env:"REDIS_PASSWORD" envDefault:""`
	DB       int           `env:"REDIS_DB" envDefault:"0"`
	TTL      time.Duration `env:"REDIS_TTL" envDefault:"5m"`
}

type HTTPServerConfig struct {
	Addr              string        `env:"HTTP_SERVER_ADDR,required"`
	ReadTimeout       time.Duration `env:"HTTP_SERVER_READ_TIMEOUT,required"`
	ReadHeaderTimeout time.Duration `env:"HTTP_SERVER_READ_HEADER_TIMEOUT,required"`
	WriteTimeout      time.Duration `env:"HTTP_SERVER_WRITE_TIMEOUT,required"`
	IdleTimeout       time.Duration `env:"HTTP_SERVER_IDLE_TIMEOUT,required"`
	MaxHeaderBytes    int           `env:"HTTP_SERVER_MAX_HEADERS_BYTES,required"`
}

func MustLoad() *Config {
	var cfg Config

	if err := envconfig.Parse(&cfg); err != nil {
		panic(fmt.Errorf("failed to parse env config: %w", err))
	}

	return &cfg
}
