package users

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/archnets/telegram-bot/internal/auth"
	"github.com/archnets/telegram-bot/internal/botapp/commands"
	"github.com/archnets/telegram-bot/internal/i18n"
	"github.com/archnets/telegram-bot/internal/logger"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// HandleLanguage shows language selection buttons.
func HandleLanguage(ctx context.Context, b *bot.Bot, u *models.Update, deps commands.Deps) {
	if u.Message == nil {
		return
	}
	lg := logger.ForUpdate(u)

	if err := ensureAuthenticated(ctx, u.Message.From, deps); err != nil {
		lg.Errorf("Auth error: %v", err)
		sendError(ctx, b, u.Message.Chat.ID, u.Message.From.LanguageCode, "auth_error")
		return
	}

	sendLanguageSelection(ctx, b, u.Message.Chat.ID, u.Message.From.LanguageCode)
}

// HandleLanguageCallback handles inline button callback for language selection.
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

	// Ensure authenticated
	token := deps.Sessions.GetToken(cb.From.ID)
	lg.Debugf("Token from session: len=%d", len(token))

	if token == "" {
		lg.Debugf("No token, authenticating...")
		if err := ensureAuthenticated(ctx, &cb.From, deps); err != nil {
			lg.Errorf("Auth error: %v", err)
			answerCallback(ctx, b, cb.ID, "Authentication error", true)
			return
		}
		token = deps.Sessions.GetToken(cb.From.ID)
		lg.Debugf("Token after auth: len=%d", len(token))
	}

	if token == "" {
		lg.Errorf("Token still empty after auth!")
		answerCallback(ctx, b, cb.ID, "Authentication failed", true)
		return
	}

	// Update language via API
	lg.Debugf("Calling UpdateUserLanguage...")
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

	// Send welcome message with image
	loc := i18n.Localizer(lang)
	botName := deps.BotNames[lang]
	if botName == "" {
		botName = deps.BotNames["en"] // fallback
	}
	welcome := i18n.TWithData(loc, "welcome", map[string]any{"BotName": botName})
	caption := welcome + "\n\n" + i18n.T(loc, "select_option")

	photo, err := os.Open("assets/welcome.jpg")
	if err != nil {
		lg.Errorf("Failed to open welcome image: %v", err)
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: cb.From.ID,
			Text:   caption,
		})
	} else {
		defer photo.Close()
		_, _ = b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID:  cb.From.ID,
			Photo:   &models.InputFileUpload{Filename: "welcome.jpg", Data: photo},
			Caption: caption,
		})
	}

	lg.Infof("Language changed to %s", lang)
}

// --- Helpers ---

func ensureAuthenticated(ctx context.Context, user *models.User, deps commands.Deps) error {
	if deps.Sessions.GetToken(user.ID) != "" {
		return nil
	}

	token, err := deps.AuthClient.Authenticate(auth.TelegramUser{
		ID:           user.ID,
		Username:     user.Username,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		LanguageCode: user.LanguageCode,
	})
	if err != nil {
		return err
	}

	deps.Sessions.Set(user.ID, &auth.Session{
		Token:     token,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})
	return nil
}

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
