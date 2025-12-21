package admins

import (
	"context"

	"github.com/archnets/telegram-bot/internal/botapp/commands"
	"github.com/archnets/telegram-bot/internal/i18n"
	"github.com/archnets/telegram-bot/internal/logger"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// HandleStart handles the /start_admin command for admins.
func HandleStart(ctx context.Context, b *bot.Bot, u *models.Update, deps commands.Deps) {
	if u.Message == nil {
		return
	}
	lg := logger.ForUpdate(u)
	loc := i18n.Localizer(u.Message.From.LanguageCode)

	// Check if user is an admin
	if !deps.Auth.IsAdmin(u.Message.From.ID) {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: u.Message.Chat.ID,
			Text:   i18n.T(loc, "access_denied"),
		})
		return
	}

	text := i18n.T(loc, "admin_welcome") + "\n\n" + i18n.T(loc, "admin_commands")

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: u.Message.Chat.ID,
		Text:   text,
	})

	lg.Infof("Admin start command handled")
}
