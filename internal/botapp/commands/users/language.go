package users

import (
	"context"
	"os"
	"strings"

	"github.com/archnets/telegram-bot/internal/botapp/commands"
	"github.com/archnets/telegram-bot/internal/i18n"
	"github.com/archnets/telegram-bot/internal/logger"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// HandleLanguage shows language selection buttons.
// Note: Authentication is handled by middleware.
func HandleLanguage(ctx context.Context, b *bot.Bot, u *models.Update, deps commands.Deps) {
	if u.Message == nil {
		return
	}

	sendLanguageSelection(ctx, b, u.Message.Chat.ID, u.Message.From.LanguageCode)
}

// HandleLanguageCallback handles inline button callback for language selection.
// Note: Authentication is handled by middleware.
func HandleLanguageCallback(ctx context.Context, b *bot.Bot, u *models.Update, deps commands.Deps) {
	if u.CallbackQuery == nil {
		return
	}

	cb := u.CallbackQuery
	lg := logger.ForUser(cb.From.ID)

	// Parse "lang:fa" -> "fa"
	lang := strings.TrimPrefix(cb.Data, "lang:")
	if lang == cb.Data {
		return // Not a language callback
	}

	lg.Debugf("Language callback: %s", lang)

	// Get token from session (middleware ensures we're authenticated)
	token := deps.Sessions.GetToken(cb.From.ID)
	if token == "" {
		lg.Errorf("No token in session after middleware auth")
		answerCallback(ctx, b, cb.ID, "Authentication error", true)
		return
	}

	// Update language via API
	if err := deps.API.UpdateUserLanguage(ctx, token, lang); err != nil {
		lg.Errorf("Update lang error: %v", err)
		answerCallback(ctx, b, cb.ID, "Failed to save", true)
		return
	}

	// Update cache
	deps.Sessions.SetLang(cb.From.ID, lang)

	// Answer callback (removes loading state)
	answerCallback(ctx, b, cb.ID, "", false)

	// Delete the language selection message
	deleteMessage(ctx, b, cb)

	// Check channel membership before showing welcome
	if deps.RequiredChannel != "" {
		isMember, err := isChannelMember(ctx, b, deps.RequiredChannel, cb.From.ID)
		if err != nil {
			lg.Warnf("Channel check failed: %v", err)
			// Fail open - show welcome
		} else if !isMember {
			// Not a member - send join prompt
			sendJoinChannelPrompt(ctx, b, cb.From.ID, lang, deps)
			lg.Infof("Language changed to %s (pending channel join)", lang)
			return
		}
	}

	// Member (or no channel required) - send welcome message
	sendWelcomeMessage(ctx, b, cb.From.ID, lang, deps, lg)
	lg.Infof("Language changed to %s", lang)
}

// sendJoinChannelPrompt sends a localized message asking user to join the channel.
func sendJoinChannelPrompt(ctx context.Context, b *bot.Bot, chatID int64, lang string, deps commands.Deps) {
	loc := i18n.Localizer(lang)
	channelURL := "https://t.me/" + deps.RequiredChannel[1:] // Remove @ prefix

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: i18n.T(loc, "join_channel_button"), URL: channelURL},
			},
		},
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        i18n.T(loc, "join_channel_required"),
		ReplyMarkup: keyboard,
	})
}

// isChannelMember checks if a user is a member of the specified channel.
func isChannelMember(ctx context.Context, b *bot.Bot, channelUsername string, userID int64) (bool, error) {
	member, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
		ChatID: channelUsername,
		UserID: userID,
	})
	if err != nil {
		return false, err
	}

	switch member.Type {
	case "member", "administrator", "creator":
		return true, nil
	default:
		return false, nil
	}
}

// sendWelcomeMessage sends the welcome message with image and menu keyboard.
func sendWelcomeMessage(ctx context.Context, b *bot.Bot, chatID int64, lang string, deps commands.Deps, lg logger.TgLogger) {
	loc := i18n.Localizer(lang)
	botName := deps.BotNames[lang]
	if botName == "" {
		botName = deps.BotNames["en"]
	}
	welcome := i18n.TWithData(loc, "welcome", map[string]any{"BotName": botName})
	caption := welcome + "\n\n" + i18n.T(loc, "select_option")
	keyboard := buildMainMenuKeyboard(lang)

	photo, err := os.Open("assets/welcome.jpg")
	if err != nil {
		lg.Errorf("Failed to open welcome image: %v", err)
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        caption,
			ReplyMarkup: keyboard,
		})
	} else {
		defer photo.Close()
		_, _ = b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID:      chatID,
			Photo:       &models.InputFileUpload{Filename: "welcome.jpg", Data: photo},
			Caption:     caption,
			ReplyMarkup: keyboard,
		})
	}
}

// --- Helpers ---

func sendLanguageSelection(ctx context.Context, b *bot.Bot, chatID int64, lang string) {
	loc := i18n.Localizer(lang)

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "🇮🇷 فارسی", CallbackData: "lang:fa"},
				{Text: "🇬🇧 English", CallbackData: "lang:en"},
			},
			{
				{Text: "🇷🇺 Русский", CallbackData: "lang:ru"},
				{Text: "🇨🇳 中文", CallbackData: "lang:zh"},
			},
		},
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        i18n.T(loc, "choose_language"),
		ReplyMarkup: keyboard,
	})
}

func sendError(ctx context.Context, b *bot.Bot, chatID int64, lang, key string) {
	loc := i18n.Localizer(lang)
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   i18n.T(loc, key),
	})
}

func answerCallback(ctx context.Context, b *bot.Bot, id, text string, alert bool) {
	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: id,
		Text:            text,
		ShowAlert:       alert,
	})
}

func editMessage(ctx context.Context, b *bot.Bot, cb *models.CallbackQuery, text string) {
	if cb.Message.Message != nil {
		_, _ = b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    cb.Message.Message.Chat.ID,
			MessageID: cb.Message.Message.ID,
			Text:      text,
		})
	}
}

func deleteMessage(ctx context.Context, b *bot.Bot, cb *models.CallbackQuery) {
	if cb.Message.Message != nil {
		_, _ = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    cb.Message.Message.Chat.ID,
			MessageID: cb.Message.Message.ID,
		})
	}
}
