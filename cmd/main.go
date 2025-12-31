package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/archnets/telegram-bot/internal/api"

	"github.com/archnets/telegram-bot/config"
	"github.com/archnets/telegram-bot/internal/auth"
	"github.com/archnets/telegram-bot/internal/botapp"
	"github.com/archnets/telegram-bot/internal/core"
	"github.com/archnets/telegram-bot/internal/db"
	"github.com/archnets/telegram-bot/internal/logger"
)

func main() {
	// Handle Ctrl+C / SIGTERM for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg := config.Load()

	// Initialize database with migrations
	// ...

	if cfg.APIBaseURL == "" {
		log.Println("WARNING: API_BASE_URL is empty. API calls will fail.")
	}

	// Fetch token from backend if not set in env (or override)
	botToken := cfg.BotToken
	if cfg.AdminEmail != "" && cfg.AdminPassword != "" {
		logger.Infof("Fetching bot token from backend...")
		token, err := bootstrapBotToken(ctx, cfg)
		if err != nil {
			logger.Errorf("Failed to fetch bot token: %v", err)
			return
		}
		if token != "" {
			botToken = token
			logger.Infof("Successfully fetched bot token from backend")
		}
	}

	if botToken == "" {
		logger.Errorf("TELEGRAM_BOT_TOKEN is not set and could not be fetched from backend")
		return
	}

	// Initialize database with migrations
	database, err := db.Open(cfg.DBPath)
	if err != nil {
		logger.Errorf("Failed to open database: %v", err)
		return
	}
	defer database.Close()
	logger.Infof("Database initialized: %s", cfg.DBPath)

	// Create session store
	sessions := auth.NewSQLiteStore(database)

	// Core services (auth, subscription) - these are legacy placeholders
	authSvc := core.NewAuthService(nil)
	subSvc := core.NewSubscriptionService(nil)

	// Dependencies for the bot layer
	deps := botapp.Dependencies{
		Auth:            authSvc,
		Subscription:    subSvc,
		WebAppURL:       cfg.WebAppURL,
		APIBaseURL:      cfg.APIBaseURL,
		BotToken:        botToken,
		BotNames:        cfg.BotNames,
		Sessions:        sessions,
		RequiredChannel: cfg.RequiredChannel,
	}

	// Bot configuration
	botCfg := botapp.Config{
		Debug:       cfg.BotDebug,
		InitTimeout: time.Duration(cfg.BotTimeoutS) * time.Second,
	}

	// Create Telegram bot
	b, err := botapp.NewBot(botToken, deps, botCfg)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	logger.Infof("Starting Telegram bot...")
	b.Start(ctx)
	logger.Infof("Bot stopped")
}

func bootstrapBotToken(ctx context.Context, cfg config.Config) (string, error) {
	client := api.NewClient(cfg.APIBaseURL, 10*time.Second)

	// Login to get admin token
	adminToken, err := client.Login(ctx, cfg.AdminEmail, cfg.AdminPassword)
	if err != nil {
		return "", fmt.Errorf("admin login failed: %w", err)
	}

	// Fetch telegram config
	botToken, err := client.GetAuthMethodConfig(ctx, adminToken, "telegram")
	if err != nil {
		return "", fmt.Errorf("fetch config failed: %w", err)
	}

	return botToken, nil
}
