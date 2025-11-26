package users

import (
	"context"

	"github.com/archnets/telegram-bot/internal/core"
	"github.com/archnets/telegram-bot/internal/logger"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Deps is the dependencies that user command handlers need.
// It is separate from botapp.Dependencies to avoid an import cycle.
type Deps struct {
	Auth         *core.AuthService
	Subscription *core.SubscriptionService
	WebAppURL    string
}

func HandleStart(ctx context.Context, b *bot.Bot, u *models.Update, deps Deps) {
	if u.Message == nil {
		return
	}
	lg := logger.ForUpdate(u)
	text := "Welcome! 👋\n\n" +
		"/login <token>\n/status\n/configs\n/webapp"

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: u.Message.Chat.ID,
		Text:   text,
	})
	lg.Infof("Start command handled")
	lg.Errorf("This is a Error message")
	lg.Debugf("This is a Debug message")
	lg.Warnf("This is a Warning message")

	_ = deps // just to show we can use it later (webapp URL, etc.)
}
