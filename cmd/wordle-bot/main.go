package main

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/Fr11zy/WordleBot/internal/config"
	"github.com/Fr11zy/WordleBot/internal/bot"
)


func main() {
	rand.Seed(time.Now().UnixNano())

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	if cfg.BotToken == "" {
		log.Fatalf("Bot token not found in environment variables")
	}

	b, err := bot.New(cfg.BotToken)
	if err != nil {
		log.Fatalf("Failed to intialize bot: %v", err)
	}

	if err := b.Run(context.Background()); err != nil {
		log.Fatalf("Failed to run bot: %v", err)
	}
}

