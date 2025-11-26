package botapp

import (
	"context"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func handleStart(ctx context.Context, b *bot.Bot, u *models.Update, _ Dependencies) {
	if u.Message == nil {
		return
	}

	text := "Welcome to the VPN bot 👋\n\n" +
		"This bot is just a frontend for your VPN panel.\n\n" +
		"Commands:\n" +
		"/login <token>  - link your account\n" +
		"/status         - show subscription status\n" +
		"/configs        - show your configs/links\n" +
		"/webapp         - open the mini web app"

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: u.Message.Chat.ID,
		Text:   text,
	})
}

func handleDefault(ctx context.Context, b *bot.Bot, u *models.Update, _ Dependencies) {
	if u.Message == nil {
		return
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: u.Message.Chat.ID,
		Text:   "Unknown command. Try /start",
	})
}

func handleLogin(ctx context.Context, b *bot.Bot, u *models.Update, deps Dependencies) {
	if u.Message == nil {
		return
	}

	chatID := u.Message.Chat.ID
	text := strings.TrimSpace(u.Message.Text)

	parts := strings.SplitN(text, " ", 2)
	if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Usage: /login <token>",
		})
		return
	}

	token := strings.TrimSpace(parts[1])

	if err := deps.Auth.LoginWithToken(ctx, chatID, token); err != nil {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Login failed. Please try again.\n\n(Backend error placeholder, check logs.)",
		})
		return
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "You are now logged in ✅",
	})
}

func handleStatus(ctx context.Context, b *bot.Bot, u *models.Update, deps Dependencies) {
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

func handleConfigs(ctx context.Context, b *bot.Bot, u *models.Update, deps Dependencies) {
	if u.Message == nil {
		return
	}

	chatID := u.Message.Chat.ID

	configs, err := deps.Subscription.GetConfigsText(ctx, chatID)
	if err != nil {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Failed to get configs. Please try again later.",
		})
		return
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   configs,
	})
}

func handleWebApp(ctx context.Context, b *bot.Bot, u *models.Update, deps Dependencies) {
	if u.Message == nil {
		return
	}

	chatID := u.Message.Chat.ID

	if deps.WebAppURL == "" {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "WEBAPP_URL is not configured on the bot.",
		})
		return
	}

	// Telegram Mini Web App button
	markup := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text: "Open Web App",
					WebApp: &models.WebAppInfo{
						URL: deps.WebAppURL,
					},
				},
			},
		},
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        "Tap the button below to open the web app:",
		ReplyMarkup: markup,
	})
}
