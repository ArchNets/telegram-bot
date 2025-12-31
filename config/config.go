package config

import (
	"strconv"
	"strings"

	"github.com/archnets/telegram-bot/internal/env"
)

type Config struct {
	BotToken        string
	BotNames        map[string]string // per-language bot names
	APIBaseURL      string
	WebAppURL       string
	DBPath          string
	BotDebug        bool
	BotTimeoutS     int    // seconds for init timeout, e.g. 5
	RequiredChannel string // channel users must join, e.g. "@Arch_Net"
	AdminEmail      string
	AdminPassword   string
}

func Load() Config {
	// BOT_DEBUG: "true"/"false" or "1"/"0"
	debugStr := strings.ToLower(env.GetString("BOT_DEBUG", "false"))
	botDebug := debugStr == "1" || debugStr == "true" || debugStr == "yes"

	// BOT_INIT_TIMEOUT_SEC: integer seconds
	timeoutStr := env.GetString("BOT_INIT_TIMEOUT_SEC", "5")
	timeoutSec, err := strconv.Atoi(timeoutStr)
	if err != nil || timeoutSec <= 0 {
		timeoutSec = 5
	}

	return Config{
		BotToken: env.GetString("TELEGRAM_BOT_TOKEN", ""),
		BotNames: map[string]string{
			"en": env.GetString("BOT_NAME_EN", "Arch Net"),
			"fa": env.GetString("BOT_NAME_FA", "آرچ نت"),
		},
		APIBaseURL:      env.GetString("API_BASE_URL", ""),
		WebAppURL:       env.GetString("WEBAPP_URL", ""),
		DBPath:          env.GetString("DB_PATH", "./data/sessions.db"),
		BotDebug:        botDebug,
		BotTimeoutS:     timeoutSec,
		RequiredChannel: env.GetString("REQUIRED_CHANNEL", ""),
		AdminEmail:      env.GetString("ADMIN_EMAIL", ""),
		AdminPassword:   env.GetString("ADMIN_PASSWORD", ""),
	}
}
