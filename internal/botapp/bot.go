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
	Auth            *core.AuthService
	Subscription    *core.SubscriptionService
	WebAppURL       string
	APIBaseURL      string
	BotToken        string
	BotNames        map[string]string
	Sessions        auth.SessionStore
	RequiredChannel string
}

// NewBot creates and configures a new Telegram bot instance.
func NewBot(token string, deps Dependencies, cfg Config) (*bot.Bot, error) {
	if cfg.InitTimeout == 0 {
		cfg.InitTimeout = 5 * time.Second
	}

	// Create shared deps
	sharedDeps := commands.Deps{
		Auth:            deps.Auth,
		Subscription:    deps.Subscription,
		WebAppURL:       deps.WebAppURL,
		BotToken:        token,
		BotNames:        deps.BotNames,
		API:             api.NewClient(deps.APIBaseURL, 10*time.Second),
		AuthClient:      auth.NewClient(deps.APIBaseURL, token),
		Sessions:        deps.Sessions,
		RequiredChannel: deps.RequiredChannel,
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

	// Setup bot UI elements (WebApp menu button)
	setupBotUI(context.Background(), b, deps)

	registerCommands(b, sharedDeps)
	registerCallbacks(b, sharedDeps)

	return b, nil
}

// setupBotUI configures menu button as WebApp and clears commands.
func setupBotUI(ctx context.Context, b *bot.Bot, deps Dependencies) {
	// Delete all commands from the / menu
	_, _ = b.DeleteMyCommands(ctx, &bot.DeleteMyCommandsParams{})

	// Set menu button to WebApp type (if WebAppURL is configured)
	if deps.WebAppURL != "" {
		_, _ = b.SetChatMenuButton(ctx, &bot.SetChatMenuButtonParams{
			MenuButton: &models.MenuButtonWebApp{
				Type:   "web_app",
				Text:   deps.BotNames["en"], // Use English name for menu button
				WebApp: models.WebAppInfo{URL: deps.WebAppURL},
			},
		})
	}
}

func registerCommands(b *bot.Bot, deps commands.Deps) {
	// /start handles its own auth and channel check (special first-time user flow)
	register(b, "/start", users.HandleStart, deps)

	// Commands with authentication + channel membership middleware
	register(b, "/status", commands.WithAuthAndChannel(users.HandleStatus), deps)
	register(b, "/lang", commands.WithAuthAndChannel(users.HandleLanguage), deps)

	// Admin commands (no channel check for admins)
	register(b, "/start_admin", admins.HandleStart, deps)
}

func registerCallbacks(b *bot.Bot, deps commands.Deps) {
	b.RegisterHandler(
		bot.HandlerTypeCallbackQueryData,
		"lang:",
		bot.MatchTypePrefix,
		wrapHandler(commands.WithAuth(users.HandleLanguageCallback), deps),
	)

}

func register(b *bot.Bot, command string, handler commands.HandlerFunc, deps commands.Deps) {
	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		command,
		bot.MatchTypeExact,
		wrapHandler(handler, deps),
	)
}

func wrapHandler(handler commands.HandlerFunc, deps commands.Deps) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, u *models.Update) {
		handler(ctx, b, u, deps)
	}
}
