package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/archnets/telegram-bot/config"
	"github.com/archnets/telegram-bot/internal/botapp"
	"github.com/archnets/telegram-bot/internal/core"
	"github.com/archnets/telegram-bot/internal/logger"
	"github.com/archnets/telegram-bot/service"
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
		log.Println("WARNING: API_BASE_URL is empty. API calls are placeholders until you wire them.")
	}

	// HTTP client to your backend API (placeholders now)
	apiClient := service.NewAPIClient(cfg.APIBaseURL, 10*time.Second)

	// Core services (auth, subscription) use the API client
	authSvc := core.NewAuthService(apiClient)
	subSvc := core.NewSubscriptionService(apiClient)

	// Dependencies for the bot layer
	deps := botapp.Dependencies{
		Auth:         authSvc,
		Subscription: subSvc,
		WebAppURL:    cfg.WebAppURL,

		Debug:       cfg.BotDebug,
		InitTimeout: time.Duration(cfg.BotTimeoutS) * time.Second,
	}

	// Create Telegram bot
	b, err := botapp.NewBot(cfg.BotToken, deps)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}
	
	logger.Infof("Starting Telegram bot...")
	b.Start(ctx)
	logger.Infof("Bot stopped")
}
