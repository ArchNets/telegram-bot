// Package commands provides command handler types, middleware, and shared dependencies.
package commands

import (
	"context"
	"time"

	"github.com/archnets/telegram-bot/internal/auth"
	"github.com/archnets/telegram-bot/internal/logger"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// HandlerFunc is the standard signature for all command handlers.
type HandlerFunc func(ctx context.Context, b *bot.Bot, u *models.Update, deps Deps)

// Middleware wraps a handler to add functionality.
type Middleware func(HandlerFunc) HandlerFunc

// WithAuth ensures the user is authenticated before the handler runs.
// If authentication fails, sends an error message and stops.
func WithAuth(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, u *models.Update, deps Deps) {
		user := getUserFromUpdate(u)
		if user == nil {
			next(ctx, b, u, deps)
			return
		}

		// Skip if already authenticated
		if deps.Sessions.GetToken(user.ID) != "" {
			next(ctx, b, u, deps)
			return
		}

		// Fetch user's profile photo URL
		lg := logger.ForUser(user.ID)
		photoURL := getUserPhotoURL(ctx, b, user.ID, deps.BotToken, lg)

		// Authenticate with backend
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
			sendAuthError(ctx, b, u, user.LanguageCode)
			return
		}

		// Store session
		deps.Sessions.Set(user.ID, &auth.Session{
			Token:     token,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		})

		lg.Infof("User authenticated")
		next(ctx, b, u, deps)
	}
}

// Chain combines multiple middleware into a single middleware.
func Chain(middlewares ...Middleware) Middleware {
	return func(final HandlerFunc) HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// --- Helpers ---

// getUserFromUpdate extracts the user from any update type.
func getUserFromUpdate(u *models.Update) *models.User {
	if u.Message != nil {
		return u.Message.From
	}
	if u.CallbackQuery != nil {
		return &u.CallbackQuery.From
	}
	return nil
}

// getChatIDFromUpdate extracts the chat ID from any update type.
func getChatIDFromUpdate(u *models.Update) int64 {
	if u.Message != nil {
		return u.Message.Chat.ID
	}
	if u.CallbackQuery != nil {
		return u.CallbackQuery.From.ID
	}
	return 0
}

// sendAuthError sends an authentication error message.
func sendAuthError(ctx context.Context, b *bot.Bot, u *models.Update, lang string) {
	chatID := getChatIDFromUpdate(u)
	if chatID == 0 {
		return
	}

	// Use a simple error message (localization can be added later)
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "❌ Authentication error. Please try again.",
	})
}

// getUserPhotoURL fetches the direct URL to the user's largest profile photo.
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
