package users

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/archnets/telegram-bot/internal/api"
	"github.com/archnets/telegram-bot/internal/botapp/commands"
	"github.com/archnets/telegram-bot/internal/i18n"
	"github.com/archnets/telegram-bot/internal/logger"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// HandleTraffic shows user's subscription traffic usage
func HandleTraffic(ctx context.Context, b *bot.Bot, u *models.Update, deps commands.Deps) {
	if u.Message == nil {
		return
	}

	lang := GetLanguage(ctx, u.Message.From.ID, u.Message.From.LanguageCode, deps)
	token := deps.Sessions.GetToken(u.Message.From.ID)

	log.Printf("[DEBUG] HandleTraffic: user=%d, lang=%s, hasToken=%v", u.Message.From.ID, lang, token != "")

	if token == "" {
		// Try to authenticate if no token
		log.Printf("[DEBUG] HandleTraffic: No token for user %d, attempting auth", u.Message.From.ID)
		var err error
		token, err = Authenticate(ctx, b, u.Message.From, deps, logger.ForUpdate(u))
		if err != nil {
			log.Printf("[ERROR] HandleTraffic: Initial auth failed: %v", err)
			SendError(ctx, b, u.Message.Chat.ID, lang, "auth_error")
			return
		}
	}

	// Fetch subscriptions from API
	log.Printf("[DEBUG] HandleTraffic: Calling API GetUserSubscriptions for user %d", u.Message.From.ID)
	subs, err := deps.API.GetUserSubscriptions(ctx, token)
	if err != nil {
		log.Printf("[ERROR] HandleTraffic: API error for user %d: %v", u.Message.From.ID, err)

		// Check if it's an auth error (token expired, invalid, etc.)
		if apiErr, ok := err.(*api.Error); ok && api.IsAuthError(apiErr.Code) {
			log.Printf("[DEBUG] HandleTraffic: Token expired for user %d, refreshing", u.Message.From.ID)

			// Re-authenticate
			var authErr error
			token, authErr = Authenticate(ctx, b, u.Message.From, deps, logger.ForUpdate(u))
			if authErr != nil {
				log.Printf("[ERROR] HandleTraffic: Auth refresh failed: %v", authErr)
				deps.Sessions.Delete(u.Message.From.ID) // clear invalid session
				SendError(ctx, b, u.Message.Chat.ID, lang, "session_expired")
				return
			}

			// Retry API with new token
			subs, err = deps.API.GetUserSubscriptions(ctx, token)
			if err != nil {
				log.Printf("[ERROR] HandleTraffic: Retry API failed: %v", err)
				SendError(ctx, b, u.Message.Chat.ID, lang, "traffic_error")
				return
			}
		} else {
			SendError(ctx, b, u.Message.Chat.ID, lang, "traffic_error")
			return
		}
	}
	log.Printf("[DEBUG] HandleTraffic: Got %d subscriptions for user %d", len(subs), u.Message.From.ID)

	if len(subs) == 0 {
		loc := i18n.Localizer(lang)
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    u.Message.Chat.ID,
			Text:      i18n.T(loc, "no_subscriptions"),
			ParseMode: models.ParseModeHTML,
		})
		return
	}

	// Build traffic message
	text := formatTrafficMessage(subs, lang)

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    u.Message.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})
}

func formatTrafficMessage(subs []api.UserSubscription, lang string) string {
	loc := i18n.Localizer(lang)
	title := i18n.T(loc, "traffic_title")

	msg := fmt.Sprintf("<b>%s</b>\n\n", title)

	for _, sub := range subs {
		name := sub.CustomName
		if name == "" {
			name = sub.Subscribe.Name
		}

		// Calculate usage
		used := sub.Download + sub.Upload
		total := sub.Traffic
		percent := float64(0)
		if total > 0 {
			percent = float64(used) / float64(total) * 100
		}

		// Format bytes
		usedStr := formatBytes(used)
		totalStr := formatBytes(total)

		// Format expire time from Unix milliseconds
		expireDate := formatExpireTime(sub.ExpireTime)

		msg += fmt.Sprintf("<b>📦 %s</b>\n", name)
		msg += fmt.Sprintf("├ %s / %s (%.0f%%)\n", usedStr, totalStr, percent)
		msg += fmt.Sprintf("├ ⬇️ %s ⬆️ %s\n", formatBytes(sub.Download), formatBytes(sub.Upload))
		msg += fmt.Sprintf("└ 📅 %s\n\n", expireDate)
	}

	return msg
}

func formatExpireTime(unixMs int64) string {
	if unixMs == 0 {
		return "Never"
	}
	t := time.Unix(unixMs/1000, 0)
	return t.Format("2006-01-02")
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
