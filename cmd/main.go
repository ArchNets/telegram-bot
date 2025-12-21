package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	if cfg.BotToken == "" {
		logger.Errorf("TELEGRAM_BOT_TOKEN is not set")
		return
	}

	if cfg.APIBaseURL == "" {
		log.Println("WARNING: API_BASE_URL is empty. API calls will fail.")
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
		BotToken:        cfg.BotToken,
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
	b, err := botapp.NewBot(cfg.BotToken, deps, botCfg)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	logger.Infof("Starting Telegram bot...")
	b.Start(ctx)
	logger.Infof("Bot stopped")
}
