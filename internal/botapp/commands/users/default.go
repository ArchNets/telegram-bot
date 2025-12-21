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
	loc := i18n.Localizer(lang)

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: u.Message.Chat.ID,
		Text:   i18n.T(loc, "unknown_command"),
	})
}
