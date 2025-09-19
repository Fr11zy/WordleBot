package bot

import (
	"context"
	"log"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type Bot struct {
	bot		*telego.Bot
	handler	*th.BotHandler
}

func New(token string) (*Bot, error) {
	bot, err := telego.NewBot(token, telego.WithDefaultDebugLogger())
	if err != nil {
		return nil, err
	}

	updates, err := bot.UpdatesViaLongPolling(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	bothandler, err := th.NewBotHandler(bot, updates)
	if err != nil {
		return nil, err
	}

	return &Bot{
		bot: bot,
		handler: bothandler,
	}, nil
}

func (b* Bot) Run(ctx context.Context) error {
	b.handler.Handle(handleSolve, th.CommandEqual("solve"))
	b.handler.Handle(handleStart, th.CommandEqual("start"))
	b.handler.Handle(handleHelp, th.CommandEqual("help"))
	b.handler.Handle(handlePLay, th.CommandEqual("play"))
	b.handler.Handle(handleFeedBack)

	defer func() { 
		if err := b.handler.Stop(); err != nil {
			log.Printf("Failed to stop bot handler: %v", err)
		}
	}()

	return b.handler.Start()
}