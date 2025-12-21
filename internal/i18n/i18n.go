// Package i18n provides internationalization support using go-i18n.
package i18n

import (
	"embed"
	"encoding/json"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localeFS embed.FS

// Bundle holds all loaded translations.
var bundle *i18n.Bundle

// Supported language tags
var (
	Persian = language.Persian
	English = language.English
	Russian = language.Russian
	Chinese = language.Chinese
)

func init() {
	bundle = i18n.NewBundle(language.Persian) // Default language
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Load all locale files
	files := []string{"fa.json", "en.json", "ru.json", "zh.json"}
	for _, f := range files {
		_, _ = bundle.LoadMessageFileFS(localeFS, "locales/"+f)
	}
}

// Localizer creates a localizer for the given Telegram language code.
func Localizer(telegramLangCode string) *i18n.Localizer {
	tag := FromTelegram(telegramLangCode)
	return i18n.NewLocalizer(bundle, tag.String())
}

// T translates a message ID using the provided localizer.
func T(loc *i18n.Localizer, messageID string) string {
	msg, err := loc.Localize(&i18n.LocalizeConfig{MessageID: messageID})
	if err != nil {
		// Fallback: return the message ID itself
		return messageID
	}
	return msg
}

// TWithData translates a message with template data.
func TWithData(loc *i18n.Localizer, messageID string, data map[string]any) string {
	msg, err := loc.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
	})
	if err != nil {
		return messageID
	}
	return msg
}

// FromTelegram converts Telegram's language_code to a language.Tag.
func FromTelegram(code string) language.Tag {
	switch code {
	case "en":
		return English
	case "ru":
		return Russian
	case "zh", "zh-hans", "zh-hant":
		return Chinese
	case "fa":
		return Persian
	default:
		// Default to Persian for unsupported languages
		return Persian
	}
}
