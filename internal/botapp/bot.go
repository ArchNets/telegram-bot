package botapp

import (
	"context"
	"time"

	"github.com/archnets/telegram-bot/internal/core"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Dependencies struct {
	Auth         *core.AuthService
	Subscription *core.SubscriptionService
	WebAppURL    string
}

func NewBot(token string, deps Dependencies) (*bot.Bot, error) {
	defaultHandler := func(ctx context.Context, b *bot.Bot, u *models.Update) {
		handleDefault(ctx, b, u, deps)
	}

	opts := []bot.Option{
		bot.WithCheckInitTimeout(5 * time.Second),
		bot.WithDefaultHandler(defaultHandler),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		return nil, err
	}

	// /start
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact,
		func(ctx context.Context, b *bot.Bot, u *models.Update) {
			handleStart(ctx, b, u, deps)
		},
	)

	// /login <token>
	b.RegisterHandler(bot.HandlerTypeMessageText, "/login", bot.MatchTypePrefix,
		func(ctx context.Context, b *bot.Bot, u *models.Update) {
			handleLogin(ctx, b, u, deps)
		},
	)

	// /status
	b.RegisterHandler(bot.HandlerTypeMessageText, "/status", bot.MatchTypeExact,
		func(ctx context.Context, b *bot.Bot, u *models.Update) {
			handleStatus(ctx, b, u, deps)
		},
	)

	// /configs
	b.RegisterHandler(bot.HandlerTypeMessageText, "/configs", bot.MatchTypeExact,
		func(ctx context.Context, b *bot.Bot, u *models.Update) {
			handleConfigs(ctx, b, u, deps)
		},
	)

	// /webapp - send Telegram Web App button (mini webapp frontend)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/webapp", bot.MatchTypeExact,
		func(ctx context.Context, b *bot.Bot, u *models.Update) {
			handleWebApp(ctx, b, u, deps)
		},
	)

	return b, nil
}
