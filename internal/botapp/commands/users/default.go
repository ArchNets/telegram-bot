package users

import (
	"context"

	"github.com/archnets/telegram-bot/internal/botapp/commands"
	"github.com/archnets/telegram-bot/internal/i18n"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// DefaultHandler handles any unrecognized messages.
func DefaultHandler(ctx context.Context, b *bot.Bot, u *models.Update, deps commands.Deps) {
	if u.Message == nil {
		return
	}

	// Use saved language if available, otherwise Telegram's
	lang := getLanguage(ctx, u.Message.From.ID, u.Message.From.LanguageCode, deps)

	// Check if this is a menu button press
	if isMenuButton(u.Message.Text, lang) {
		HandleMenuMessage(ctx, b, u, deps)
		return
	}

	// Unknown message
	loc := i18n.Localizer(lang)
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: u.Message.Chat.ID,
		Text:   i18n.T(loc, "unknown_command"),
	})
}

// isMenuButton checks if the text matches a menu button label.
func isMenuButton(text, lang string) bool {
	loc := i18n.Localizer(lang)
	buttons := []string{
		i18n.T(loc, "btn_my_services"),
		i18n.T(loc, "btn_buy_service"),
		i18n.T(loc, "btn_balance"),
		i18n.T(loc, "btn_invitation"),
		i18n.T(loc, "btn_prices"),
		i18n.T(loc, "btn_support"),
		i18n.T(loc, "btn_settings"),
	}
	for _, btn := range buttons {
		if text == btn {
			return true
		}
	}
	return false
}
