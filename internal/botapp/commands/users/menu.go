package users

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/archnets/telegram-bot/internal/api"
	"github.com/archnets/telegram-bot/internal/botapp/commands"
	"github.com/archnets/telegram-bot/internal/i18n"
	"github.com/archnets/telegram-bot/internal/logger"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// HandleMenuMessage handles text messages from the reply keyboard menu.
func HandleMenuMessage(ctx context.Context, b *bot.Bot, u *models.Update, deps commands.Deps) {
	if u.Message == nil {
		return
	}

	lg := logger.ForUser(u.Message.From.ID)
	lang := deps.Sessions.GetLang(u.Message.From.ID)
	if lang == "" {
		lang = "en"
	}

	token := deps.Sessions.GetToken(u.Message.From.ID)
	if token == "" {
		return // Not authenticated
	}

	loc := i18n.Localizer(lang)
	text := u.Message.Text

	// Match against localized button labels
	switch text {
	case i18n.T(loc, "btn_my_services"):
		handleMyServices(ctx, b, u.Message.Chat.ID, deps, token, lang, lg)
	case i18n.T(loc, "btn_buy_service"):
		handleBuyService(ctx, b, u.Message.Chat.ID, deps, token, lang, lg)
	case i18n.T(loc, "btn_balance"):
		handleBalance(ctx, b, u.Message.Chat.ID, deps, token, lang, lg)
	case i18n.T(loc, "btn_invitation"):
		handleInvitation(ctx, b, u.Message.Chat.ID, deps, token, lang, lg)
	case i18n.T(loc, "btn_prices"):
		handlePrices(ctx, b, u.Message.Chat.ID, deps, token, lang, lg)
	case i18n.T(loc, "btn_support"):
		handleSupport(ctx, b, u.Message.Chat.ID, deps, lang, lg)
	case i18n.T(loc, "btn_settings"):
		handleSettings(ctx, b, u.Message.Chat.ID, deps, lang, lg)
	}
}

// HandleMenuCallback handles menu button callbacks (for inline menus in settings).
func HandleMenuCallback(ctx context.Context, b *bot.Bot, u *models.Update, deps commands.Deps) {
	if u.CallbackQuery == nil {
		return
	}

	cb := u.CallbackQuery
	lg := logger.ForUser(cb.From.ID)
	lang := deps.Sessions.GetLang(cb.From.ID)
	if lang == "" {
		lang = "en"
	}

	// Parse callback data: "menu:action"
	action := strings.TrimPrefix(cb.Data, "menu:")
	if action == cb.Data {
		return // Not a menu callback
	}

	lg.Debugf("Menu callback: %s", action)

	token := deps.Sessions.GetToken(cb.From.ID)
	if token == "" {
		answerCallback(ctx, b, cb.ID, "Please /start first", true)
		return
	}

	switch action {
	case "back":
		answerCallback(ctx, b, cb.ID, "", false)
		sendWelcomeMessage(ctx, b, cb.From.ID, lang, deps, lg)
	default:
		answerCallback(ctx, b, cb.ID, "", false)
	}
}

// HandleSettingsCallback handles settings sub-menu callbacks.
func HandleSettingsCallback(ctx context.Context, b *bot.Bot, u *models.Update, deps commands.Deps) {
	if u.CallbackQuery == nil {
		return
	}

	cb := u.CallbackQuery
	lg := logger.ForUser(cb.From.ID)
	lang := deps.Sessions.GetLang(cb.From.ID)
	if lang == "" {
		lang = "en"
	}

	// Parse callback data: "settings:action"
	action := strings.TrimPrefix(cb.Data, "settings:")
	if action == cb.Data {
		return // Not a settings callback
	}

	lg.Debugf("Settings callback: %s", action)

	switch action {
	case "lang":
		answerCallback(ctx, b, cb.ID, "", false)
		sendLanguageSelection(ctx, b, cb.From.ID, lang)
	default:
		answerCallback(ctx, b, cb.ID, "", false)
	}
}

// --- Menu Handlers ---

func handleMyServices(ctx context.Context, b *bot.Bot, chatID int64, deps commands.Deps, token, lang string, lg logger.TgLogger) {
	loc := i18n.Localizer(lang)

	subs, err := deps.API.GetUserSubscriptions(ctx, token)
	if err != nil {
		lg.Errorf("Get subscriptions error: %v", err)
		sendMenuResponse(ctx, b, chatID, i18n.T(loc, "error_loading_services"))
		return
	}

	if len(subs) == 0 {
		sendMenuResponse(ctx, b, chatID, i18n.T(loc, "no_services"))
		return
	}

	// Format subscription list
	var sb strings.Builder
	sb.WriteString(i18n.T(loc, "your_services"))
	sb.WriteString("\n\n")

	for _, sub := range subs {
		expireDate := time.Unix(sub.ExpiredAt, 0).Format("2006-01-02")
		usedGB := float64(sub.Upload+sub.Download) / (1024 * 1024 * 1024)
		totalGB := float64(sub.Traffic) / (1024 * 1024 * 1024)

		sb.WriteString(fmt.Sprintf("📦 %s\n", sub.SubscribeName))
		sb.WriteString(fmt.Sprintf("   %s: %.2f/%.2f GB\n", i18n.T(loc, "traffic"), usedGB, totalGB))
		sb.WriteString(fmt.Sprintf("   %s: %s\n\n", i18n.T(loc, "expires"), expireDate))
	}

	sendMenuResponse(ctx, b, chatID, sb.String())
}

func handleBuyService(ctx context.Context, b *bot.Bot, chatID int64, deps commands.Deps, token, lang string, lg logger.TgLogger) {
	loc := i18n.Localizer(lang)

	plans, err := deps.API.GetSubscribePlans(ctx, token, lang)
	if err != nil {
		lg.Errorf("Get plans error: %v", err)
		sendMenuResponse(ctx, b, chatID, i18n.T(loc, "error_loading_plans"))
		return
	}

	if len(plans) == 0 {
		sendMenuResponse(ctx, b, chatID, i18n.T(loc, "no_plans_available"))
		return
	}

	// Show first plan with navigation
	plan := plans[0]
	trafficGB := float64(plan.Traffic) / (1024 * 1024 * 1024)

	text := i18n.TWithData(loc, "plan_details", map[string]any{
		"Name":     plan.Name,
		"Price":    plan.UnitPrice,
		"Traffic":  fmt.Sprintf("%.0f", trafficGB),
		"Duration": fmt.Sprintf("1 %s", plan.UnitTime),
		"Devices":  plan.DeviceLimit,
	})

	// Add page indicator
	text += "\n\n" + i18n.TWithData(loc, "plan_page", map[string]any{
		"Current": 1,
		"Total":   len(plans),
	})

	// Build navigation keyboard
	var navRow []models.InlineKeyboardButton
	if len(plans) > 1 {
		navRow = append(navRow, models.InlineKeyboardButton{
			Text:         i18n.T(loc, "btn_next"),
			CallbackData: "plan:page:1",
		})
	}

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:         i18n.T(loc, "btn_select_plan"),
					CallbackData: fmt.Sprintf("plan:select:%d", plan.ID),
				},
			},
		},
	}

	if len(navRow) > 0 {
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, navRow)
	}

	// Add back button
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []models.InlineKeyboardButton{
		{Text: i18n.T(loc, "btn_back"), CallbackData: "plan:back"},
	})

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: keyboard,
	})
}

func handleBalance(ctx context.Context, b *bot.Bot, chatID int64, deps commands.Deps, token, lang string, lg logger.TgLogger) {
	loc := i18n.Localizer(lang)

	userInfo, err := deps.API.GetUserInfo(ctx, token)
	if err != nil {
		lg.Errorf("Get user info error: %v", err)
		sendMenuResponse(ctx, b, chatID, i18n.T(loc, "error_loading_balance"))
		return
	}

	text := i18n.TWithData(loc, "balance_info", map[string]any{
		"Balance": userInfo.Balance,
	})

	sendMenuResponse(ctx, b, chatID, text)
}

func handleInvitation(ctx context.Context, b *bot.Bot, chatID int64, deps commands.Deps, token, lang string, lg logger.TgLogger) {
	loc := i18n.Localizer(lang)

	// Get user info for referral code
	userInfo, err := deps.API.GetUserInfo(ctx, token)
	if err != nil {
		lg.Errorf("Get user info error: %v", err)
		sendMenuResponse(ctx, b, chatID, i18n.T(loc, "error_loading_invite"))
		return
	}

	// Get affiliate stats
	affiliate, err := deps.API.GetAffiliateCount(ctx, token)
	if err != nil {
		lg.Warnf("Get affiliate error: %v", err)
		affiliate = &api.AffiliateCount{}
	}

	text := i18n.TWithData(loc, "invite_info", map[string]any{
		"ReferCode":       userInfo.ReferCode,
		"Registers":       affiliate.Registers,
		"TotalCommission": affiliate.TotalCommission,
	})

	sendMenuResponse(ctx, b, chatID, text)
}

func handlePrices(ctx context.Context, b *bot.Bot, chatID int64, deps commands.Deps, token, lang string, lg logger.TgLogger) {
	handleBuyService(ctx, b, chatID, deps, token, lang, lg)
}

func handleSupport(ctx context.Context, b *bot.Bot, chatID int64, deps commands.Deps, lang string, lg logger.TgLogger) {
	loc := i18n.Localizer(lang)
	text := i18n.T(loc, "support_info")
	sendMenuResponse(ctx, b, chatID, text)
}

func handleSettings(ctx context.Context, b *bot.Bot, chatID int64, deps commands.Deps, lang string, lg logger.TgLogger) {
	loc := i18n.Localizer(lang)

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: i18n.T(loc, "btn_change_language"), CallbackData: "settings:lang"},
			},
		},
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        i18n.T(loc, "settings_menu"),
		ReplyMarkup: keyboard,
	})
}

// --- Helpers ---

func sendMenuResponse(ctx context.Context, b *bot.Bot, chatID int64, text string) {
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
}

// buildMainMenuKeyboard creates the main menu reply keyboard.
func buildMainMenuKeyboard(lang string) *models.ReplyKeyboardMarkup {
	loc := i18n.Localizer(lang)

	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{
				{Text: i18n.T(loc, "btn_my_services")},
				{Text: i18n.T(loc, "btn_buy_service")},
			},
			{
				{Text: i18n.T(loc, "btn_balance")},
				{Text: i18n.T(loc, "btn_invitation")},
				{Text: i18n.T(loc, "btn_prices")},
			},
			{
				{Text: i18n.T(loc, "btn_support")},
				{Text: i18n.T(loc, "btn_settings")},
			},
		},
		ResizeKeyboard: true,
	}
}
