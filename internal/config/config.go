package config

import (
	"os"
	"errors"
	"github.com/joho/godotenv"
)

type Config struct {
	BotToken string
}

func Load() (Config, error) {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return Config{}, err
		}
	}
	
	token := os.Getenv("TG_TOKEN")
	if token == "" {
		return Config{}, errors.New("TG_TOKEN environment variable is required")
	}

	return Config{BotToken: token}, nil
}