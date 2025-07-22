package config

import (
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	BotToken string
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil {
		return Config{}, err
	}

	return Config{BotToken: os.Getenv("TOKEN")}, nil
}