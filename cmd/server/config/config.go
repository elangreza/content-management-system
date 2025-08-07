package config

import (
	"github.com/joho/godotenv"

	kenv "github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	HTTP_PORT         string `koanf:"HTTP_PORT"`
	POSTGRES_USER     string `koanf:"POSTGRES_USER"`
	POSTGRES_PASSWORD string `koanf:"POSTGRES_PASSWORD"`
	POSTGRES_HOSTNAME string `koanf:"POSTGRES_HOSTNAME"`
	POSTGRES_PORT     string `koanf:"POSTGRES_PORT"`
	POSTGRES_DB       string `koanf:"POSTGRES_DB"`
	POSTGRES_SSL      string `koanf:"POSTGRES_SSL"`
}

func LoadConfig() (*Config, error) {
	var config Config

	_ = godotenv.Load(".env")

	k := koanf.New(".")

	k.Load(kenv.Provider("", ".", func(s string) string {
		return s
	}), nil)

	if err := k.Unmarshal("", &config); err != nil {
		return nil, err
	}

	return &config, nil
}
