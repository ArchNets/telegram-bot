package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/archnets/telegram-bot/internal/env"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithCheckInitTimeout(5 * time.Second),
		bot.WithDefaultHandler(handler),
	}

	token := env.GetString("TOKEN", "fallback") 

	b, err := bot.New(token, opts...)
	if err != nil {
		panic(err)
	}

	b.Start(ctx)
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   update.Message.Text,
	})
}
