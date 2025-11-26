package config

import "github.com/archnets/telegram-bot/internal/env"

type Config struct {
	BotToken   string
	APIBaseURL string
	WebAppURL  string
}

func Load() Config {
	return Config{
		BotToken:   env.GetString("TELEGRAM_BOT_TOKEN", ""),
		APIBaseURL: env.GetString("API_BASE_URL", ""),
		// URL for your Telegram Mini Web App (can be panel URL or dedicated webapp)
		WebAppURL: env.GetString("WEBAPP_URL", ""),
	}
}
