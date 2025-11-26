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
	"github.com/archnets/telegram-bot/service"
)

func main() {
	// Handle Ctrl+C / SIGTERM for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg := config.Load()

	if cfg.BotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
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
	}

	// Create Telegram bot
	b, err := botapp.NewBot(cfg.BotToken, deps)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	log.Println("Starting Telegram bot...")
	b.Start(ctx)
	log.Println("Bot stopped")
}
