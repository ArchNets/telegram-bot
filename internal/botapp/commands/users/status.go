package users

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func HandleStatus(ctx context.Context, b *bot.Bot, u *models.Update, deps Deps) {
	if u.Message == nil {
		return
	}

	chatID := u.Message.Chat.ID

	status, err := deps.Subscription.GetStatusText(ctx, chatID)
	if err != nil {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Failed to get status. Please try again later.",
		})
		return
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   status,
	})
}
