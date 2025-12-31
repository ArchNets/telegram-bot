package users

import (
	"context"

	"github.com/archnets/telegram-bot/internal/botapp/commands"
	"github.com/archnets/telegram-bot/internal/logger"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// HandleStart handles the /start command.
func HandleStart(ctx context.Context, b *bot.Bot, u *models.Update, deps commands.Deps) {
	if u.Message == nil {
		return
	}
	lg := logger.ForUpdate(u)
	user := u.Message.From

	// Authenticate (creates account if new)
	if _, err := Authenticate(ctx, b, user, deps, lg); err != nil {
		sendError(ctx, b, u.Message.Chat.ID, user.LanguageCode, "auth_error")
		return
	}

	// Get saved language (empty for first-time users)
	savedLang := deps.Sessions.GetLang(user.ID)

	// First-time users: only show language selection
	if savedLang == "" {
		sendLanguageSelection(ctx, b, u.Message.Chat.ID, user.LanguageCode)
		lg.Infof("Start command handled (first-time user)")
		return
	}

	// Returning users: check channel membership first
	if deps.RequiredChannel != "" {
		isMember, err := isChannelMember(ctx, b, deps.RequiredChannel, user.ID)
		if err != nil {
			lg.Warnf("Channel check failed: %v", err)
			// Fail open - show welcome
		} else if !isMember {
			// Not a member - send join prompt
			sendJoinChannelPrompt(ctx, b, u.Message.Chat.ID, savedLang, deps)
			lg.Infof("Start command handled (pending channel join)")
			return
		}
	}

	// Member (or no channel required) - show welcome message
	sendWelcomeMessage(ctx, b, u.Message.Chat.ID, savedLang, deps, lg)
	lg.Infof("Start command handled")
}
