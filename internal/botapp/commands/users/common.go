package users

import (
	"context"
	"time"

	"github.com/archnets/telegram-bot/internal/api"
	"github.com/archnets/telegram-bot/internal/auth"
	"github.com/archnets/telegram-bot/internal/botapp/commands"
	"github.com/archnets/telegram-bot/internal/i18n"
	"github.com/archnets/telegram-bot/internal/logger"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Authenticate authenticates the user with the backend API.
// It forces a token refresh and preserves the existing language setting.
func Authenticate(ctx context.Context, b *bot.Bot, user *models.User, deps commands.Deps, lg logger.TgLogger) (string, error) {
	// Preserve existing language
	existingLang := deps.Sessions.GetLang(user.ID)

	// Fetch user's profile photo URL
	photoURL := getUserPhotoURL(ctx, b, user.ID, deps.BotToken, lg)

	token, err := deps.AuthClient.Authenticate(auth.TelegramUser{
		ID:           user.ID,
		Username:     user.Username,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		LanguageCode: user.LanguageCode,
		PhotoURL:     photoURL,
	})
	if err != nil {
		lg.Errorf("Auth failed: %v", err)
		return "", err
	}

	// Save session with new token but preserve language
	session := &auth.Session{
		Token:     token,
		Lang:      existingLang,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	deps.Sessions.Set(user.ID, session)

	lg.Infof("User authenticated (forced refresh)")
	return token, nil
}

// getUserPhotoURL fetches the direct URL to the user's largest profile photo.
// Returns empty string if the user has no photo or if fetching fails.
func getUserPhotoURL(ctx context.Context, b *bot.Bot, userID int64, botToken string, lg logger.TgLogger) string {
	photos, err := b.GetUserProfilePhotos(ctx, &bot.GetUserProfilePhotosParams{
		UserID: userID,
		Limit:  1,
	})
	if err != nil {
		lg.Debugf("Failed to get profile photos: %v", err)
		return ""
	}

	if photos.TotalCount == 0 || len(photos.Photos) == 0 {
		return ""
	}

	// Get the largest size from the first photo
	photoSizes := photos.Photos[0]
	if len(photoSizes) == 0 {
		return ""
	}
	largestPhoto := photoSizes[len(photoSizes)-1]

	// Get direct download URL
	file, err := b.GetFile(ctx, &bot.GetFileParams{
		FileID: largestPhoto.FileID,
	})
	if err != nil {
		lg.Debugf("Failed to get file info: %v", err)
		return ""
	}

	if file.FilePath == "" {
		return ""
	}

	// Build direct URL: https://api.telegram.org/file/bot<token>/<file_path>
	return "https://api.telegram.org/file/bot" + botToken + "/" + file.FilePath
}

// GetLanguage retrieves the user's language preference.
// It checks the session cache first, then the API (if token exists).
// Falls back to fallback language or "en".
func GetLanguage(ctx context.Context, userID int64, fallback string, deps commands.Deps) string {
	// Check cache
	if lang := deps.Sessions.GetLang(userID); lang != "" {
		return lang
	}

	// Fetch from API
	token := deps.Sessions.GetToken(userID)
	if token == "" {
		if fallback == "" {
			return "en"
		}
		return fallback
	}

	info, err := deps.API.GetUserInfo(ctx, token)
	if err != nil || info.Lang == "" {
		if fallback == "" {
			return "en"
		}
		return fallback
	}

	deps.Sessions.SetLang(userID, info.Lang)
	return info.Lang
}

// SendError sends a localized error message to the user.
func SendError(ctx context.Context, b *bot.Bot, chatID int64, lang, key string) {
	loc := i18n.Localizer(lang)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      i18n.T(loc, key),
		ParseMode: models.ParseModeHTML,
	})
}

// ExecuteWithAuth executes an API action with automatic token refresh.
// It handles token retrieval, initial execution, and retry on auth error.
func ExecuteWithAuth(ctx context.Context, b *bot.Bot, u *models.Update, deps commands.Deps, action func(token string) error) {
	if u.Message == nil {
		return
	}
	user := u.Message.From
	lang := GetLanguage(ctx, user.ID, user.LanguageCode, deps)
	token := deps.Sessions.GetToken(user.ID)
	lg := logger.ForUpdate(u)

	// If no token, try initial auth
	if token == "" {
		lg.Debugf("No token found, authenticating...")
		var err error
		token, err = Authenticate(ctx, b, user, deps, lg)
		if err != nil {
			SendError(ctx, b, u.Message.Chat.ID, lang, "auth_error")
			return
		}
	}

	// Execute action
	err := action(token)
	if err == nil {
		return
	}

	// Check if it's an auth error
	if apiErr, ok := err.(*api.Error); ok && api.IsAuthError(apiErr.Code) {
		lg.Infof("Token expired, refreshing...")

		// Refresh token
		newToken, authErr := Authenticate(ctx, b, user, deps, lg)
		if authErr != nil {
			deps.Sessions.Delete(user.ID)
			SendError(ctx, b, u.Message.Chat.ID, lang, "session_expired")
			return
		}

		// Retry action with new token
		if err := action(newToken); err != nil {
			lg.Errorf("Retry failed: %v", err)
			SendError(ctx, b, u.Message.Chat.ID, lang, "traffic_error") // Generic error
		}
		return
	}

	// Other errors
	lg.Errorf("Action failed: %v", err)
	SendError(ctx, b, u.Message.Chat.ID, lang, "traffic_error") // Generic error
}
