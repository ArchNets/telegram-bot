package botapp

import (
	"context"
	"time"

	"github.com/archnets/telegram-bot/internal/api"
	"github.com/archnets/telegram-bot/internal/auth"
	"github.com/archnets/telegram-bot/internal/botapp/commands"
	"github.com/archnets/telegram-bot/internal/botapp/commands/admins"
	"github.com/archnets/telegram-bot/internal/botapp/commands/users"
	"github.com/archnets/telegram-bot/internal/core"
	"github.com/archnets/telegram-bot/internal/logger"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Config holds bot configuration options.
type Config struct {
	Debug       bool
	InitTimeout time.Duration
}

// Dependencies holds external services the bot needs.
type Dependencies struct {
	Auth         *core.AuthService
	Subscription *core.SubscriptionService
	WebAppURL    string
	APIBaseURL   string
	BotToken     string
	BotNames     map[string]string
	Sessions     auth.SessionStore
}

// HandlerFunc is the standard signature for all command handlers.
type HandlerFunc func(ctx context.Context, b *bot.Bot, u *models.Update, deps commands.Deps)

// NewBot creates and configures a new Telegram bot instance.
func NewBot(token string, deps Dependencies, cfg Config) (*bot.Bot, error) {
	if cfg.InitTimeout == 0 {
		cfg.InitTimeout = 5 * time.Second
	}

	// Create shared deps
	sharedDeps := commands.Deps{
		Auth:         deps.Auth,
		Subscription: deps.Subscription,
		WebAppURL:    deps.WebAppURL,
		BotToken:     token,
		BotNames:     deps.BotNames,
		API:          api.NewClient(deps.APIBaseURL, 10*time.Second),
		AuthClient:   auth.NewClient(deps.APIBaseURL, token),
		Sessions:     deps.Sessions,
	}

	// Configure bot options
	opts := []bot.Option{
		bot.WithCheckInitTimeout(cfg.InitTimeout),
		bot.WithDefaultHandler(wrapHandler(users.DefaultHandler, sharedDeps)),
	}

	if cfg.Debug {
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

	// Clear bot UI elements (commands menu, menu button)
	clearBotUI(context.Background(), b)

	registerCommands(b, sharedDeps)
	registerCallbacks(b, sharedDeps)

	return b, nil
}

// clearBotUI removes commands list and menu button from Telegram UI.
func clearBotUI(ctx context.Context, b *bot.Bot) {
	// Delete all commands from the / menu
	_, _ = b.DeleteMyCommands(ctx, &bot.DeleteMyCommandsParams{})

	// Set menu button to "commands" type (removes WebApp button)
	_, _ = b.SetChatMenuButton(ctx, &bot.SetChatMenuButtonParams{
		MenuButton: &models.MenuButtonCommands{Type: "commands"},
	})
}

func registerCommands(b *bot.Bot, deps commands.Deps) {
	register(b, "/start", users.HandleStart, deps)
	register(b, "/status", users.HandleStatus, deps)
	register(b, "/lang", users.HandleLanguage, deps)
	register(b, "/start_admin", admins.HandleStart, deps)
}

func registerCallbacks(b *bot.Bot, deps commands.Deps) {
	b.RegisterHandler(
		bot.HandlerTypeCallbackQueryData,
		"lang:",
		bot.MatchTypePrefix,
		func(ctx context.Context, b *bot.Bot, u *models.Update) {
			users.HandleLanguageCallback(ctx, b, u, deps)
		},
	)
}

func register(b *bot.Bot, command string, handler HandlerFunc, deps commands.Deps) {
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		command,
		bot.MatchTypeExact,
		wrapHandler(handler, deps),
	)
}

func wrapHandler(handler HandlerFunc, deps commands.Deps) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, u *models.Update) {
		handler(ctx, b, u, deps)
	}
}
