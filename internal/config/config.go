package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"time"
)

type Config struct {
	Server      Server
	Postgres    Postgres
	Redis       Redis
	WatchPeriod time.Duration `env:"WATCH_PERIOD" env-default:"5m"`
	LogLevel    string        `env:"LOG_LEVEL"`
}

type Server struct {
	Addr  string `env:"SERVER_ADDR"`
	Admin struct {
		Username string `env:"SERVER_ADMIN_USERNAME" env-default:"admin"`
		Password string `env:"SERVER_ADMIN_PASSWORD" env-default:"admin"`
	}
}

type Postgres struct {
	Host     string `env:"POSTGRES_HOST"`
	Port     string `env:"POSTGRES_PORT"`
	User     string `env:"POSTGRES_USER"`
	Password string `env:"POSTGRES_PASSWORD"`
	DB       string `env:"POSTGRES_DB"`
}

type Redis struct {
	Addr string `env:"REDIS_ADDR"`
}

func New() Config {
	var config Config
	err := cleanenv.ReadEnv(&config)
	if err != nil {
		log.Fatal(err)
	}

	return config
}
