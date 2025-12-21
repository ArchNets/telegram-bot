package users

import (
	"context"
	"os"
	"time"

	"github.com/archnets/telegram-bot/internal/auth"
	"github.com/archnets/telegram-bot/internal/botapp/commands"
	"github.com/archnets/telegram-bot/internal/i18n"
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
	if err := authenticate(ctx, user, deps, lg); err != nil {
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

	// Returning users: show welcome message
	loc := i18n.Localizer(savedLang)
	botName := deps.BotNames[savedLang]
	if botName == "" {
		botName = deps.BotNames["en"]
	}
	welcome := i18n.TWithData(loc, "welcome", map[string]any{"BotName": botName})
	caption := welcome + "\n\n" + i18n.T(loc, "select_option")

	// Send welcome image with caption
	photo, err := os.Open("assets/welcome.jpg")
	if err != nil {
		lg.Errorf("Failed to open welcome image: %v", err)
		// Fallback to text message
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      u.Message.Chat.ID,
			Text:        caption,
			ReplyMarkup: &models.ReplyKeyboardRemove{RemoveKeyboard: true},
		})
	} else {
		defer photo.Close()
		_, _ = b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID:      u.Message.Chat.ID,
			Photo:       &models.InputFileUpload{Filename: "welcome.jpg", Data: photo},
			Caption:     caption,
			ReplyMarkup: &models.ReplyKeyboardRemove{RemoveKeyboard: true},
		})
	}

	lg.Infof("Start command handled")
}

// --- Private ---

func authenticate(ctx context.Context, user *models.User, deps commands.Deps, lg logger.TgLogger) error {
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
		lg.Errorf("Auth failed: %v", err)
		return err
	}

	deps.Sessions.Set(user.ID, &auth.Session{
		Token:     token,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})

	lg.Infof("User authenticated")
	return nil
}

func getLanguage(ctx context.Context, userID int64, fallback string, deps commands.Deps) string {
	// Check cache
	if lang := deps.Sessions.GetLang(userID); lang != "" {
		return lang
	}

	// Fetch from API
	token := deps.Sessions.GetToken(userID)
	if token == "" {
		return fallback
	}

	info, err := deps.API.GetUserInfo(ctx, token)
	if err != nil || info.Lang == "" {
		return fallback
	}

	deps.Sessions.SetLang(userID, info.Lang)
	return info.Lang
}
