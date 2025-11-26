package botapp

import (
	"context"
	"time"

	"github.com/archnets/telegram-bot/internal/botapp/commands/users"
	"github.com/archnets/telegram-bot/internal/core"
	"github.com/archnets/telegram-bot/internal/logger"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Dependencies struct {
	Auth         *core.AuthService
	Subscription *core.SubscriptionService
	WebAppURL    string

	// config-like values for the bot itself
	Debug       bool
	InitTimeout time.Duration
}

func NewBot(token string, deps Dependencies) (*bot.Bot, error) {
	// fallback if not set
	initTimeout := deps.InitTimeout
	if initTimeout == 0 {
		initTimeout = 5 * time.Second
	}

	opts := []bot.Option{
		bot.WithCheckInitTimeout(initTimeout),
		bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, u *models.Update) {
			users.DefaultHandler(ctx, b, u, users.Deps{
				Auth:         deps.Auth,
				Subscription: deps.Subscription,
				WebAppURL:    deps.WebAppURL,
			})
		}),
	}

	// Only enable debug if configured
	if deps.Debug {
		opts = append(opts,
			bot.WithDebug(),
			bot.WithDebugHandler(func(format string, args ...any) {
				logger.Debugf("telegram: "+format, args...)
			}),
		)
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		return nil, err
	}

	// /start
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact,
		func(ctx context.Context, b *bot.Bot, u *models.Update) {
			users.HandleStart(ctx, b, u, users.Deps{
				Auth:         deps.Auth,
				Subscription: deps.Subscription,
				WebAppURL:    deps.WebAppURL,
			})
		},
	)

	// /status example (if you have it)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/status", bot.MatchTypeExact,
		func(ctx context.Context, b *bot.Bot, u *models.Update) {
			users.HandleStatus(ctx, b, u, users.Deps{
				Auth:         deps.Auth,
				Subscription: deps.Subscription,
				WebAppURL:    deps.WebAppURL,
			})
		},
	)

	// Admin commands later, once you really create commands/admins package
	// b.RegisterHandler(... admins.HandleAdmin ...)

	return b, nil
}
