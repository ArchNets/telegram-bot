package users

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/archnets/telegram-bot/internal/api"
	"github.com/archnets/telegram-bot/internal/botapp/commands"
	"github.com/archnets/telegram-bot/internal/i18n"
	"github.com/archnets/telegram-bot/internal/logger"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// PurchaseState tracks user's progress through purchase flow.
type PurchaseState struct {
	PlanID        int64
	PlanName      string
	Quantity      int64 // months/periods
	PaymentID     int64
	PaymentName   string
	PlanUnitPrice int64
	UnitTime      string                  // "month", "year", etc.
	Discounts     []api.SubscribeDiscount // Available quantity options
}

// In-memory purchase state (per user). In production, use Redis or DB.
var purchaseStates = make(map[int64]*PurchaseState)

// HandlePlanCallback handles plan browsing and selection callbacks.
func HandlePlanCallback(ctx context.Context, b *bot.Bot, u *models.Update, deps commands.Deps) {
	if u.CallbackQuery == nil {
		return
	}

	cb := u.CallbackQuery
	lg := logger.ForUser(cb.From.ID)
	lang := deps.Sessions.GetLang(cb.From.ID)
	if lang == "" {
		lang = "en"
	}

	token := deps.Sessions.GetToken(cb.From.ID)
	if token == "" {
		answerCallback(ctx, b, cb.ID, "Please /start first", true)
		return
	}

	// Parse callback data: "plan:action:param"
	parts := strings.Split(strings.TrimPrefix(cb.Data, "plan:"), ":")
	if len(parts) < 1 {
		return
	}

	action := parts[0]
	lg.Debugf("Plan callback: action=%s parts=%v", action, parts)

	switch action {
	case "page":
		// plan:page:N - Navigate to page N
		if len(parts) < 2 {
			return
		}
		page, _ := strconv.Atoi(parts[1])
		answerCallback(ctx, b, cb.ID, "", false)
		showPlanPage(ctx, b, cb.Message.Message, deps, token, lang, page, lg)

	case "select":
		// plan:select:planID - Select plan, go to quantity selection
		if len(parts) < 2 {
			return
		}
		planID, _ := strconv.ParseInt(parts[1], 10, 64)
		answerCallback(ctx, b, cb.ID, "", false)
		startQuantitySelection(ctx, b, cb.Message.Message, deps, token, lang, planID, cb.From.ID, lg)

	case "back":
		// plan:back - Go back to main menu
		answerCallback(ctx, b, cb.ID, "", false)
		deleteMessage(ctx, b, cb)
		sendWelcomeMessage(ctx, b, cb.From.ID, lang, deps, lg)

	default:
		lg.Warnf("Unknown plan action: %s", action)
		answerCallback(ctx, b, cb.ID, "", false)
	}
}

// HandleQuantityCallback handles quantity (duration) selection callbacks.
func HandleQuantityCallback(ctx context.Context, b *bot.Bot, u *models.Update, deps commands.Deps) {
	if u.CallbackQuery == nil {
		return
	}

	cb := u.CallbackQuery
	lg := logger.ForUser(cb.From.ID)
	lang := deps.Sessions.GetLang(cb.From.ID)
	if lang == "" {
		lang = "en"
	}

	token := deps.Sessions.GetToken(cb.From.ID)
	if token == "" {
		answerCallback(ctx, b, cb.ID, "Please /start first", true)
		return
	}

	// Parse callback data: "qty:months:M" or "qty:back"
	parts := strings.Split(strings.TrimPrefix(cb.Data, "qty:"), ":")
	if len(parts) < 1 {
		return
	}

	action := parts[0]
	lg.Debugf("Quantity callback: action=%s parts=%v", action, parts)

	switch action {
	case "months":
		if len(parts) < 2 {
			return
		}
		months, _ := strconv.ParseInt(parts[1], 10, 64)
		answerCallback(ctx, b, cb.ID, "", false)

		// Save quantity to state
		state := purchaseStates[cb.From.ID]
		if state == nil {
			lg.Errorf("No purchase state found")
			return
		}
		state.Quantity = months

		// Go to payment selection
		showPaymentSelection(ctx, b, cb.Message.Message, deps, token, lang, cb.From.ID, lg)

	case "back":
		// Go back to plan selection
		answerCallback(ctx, b, cb.ID, "", false)
		showPlanPage(ctx, b, cb.Message.Message, deps, token, lang, 0, lg)

	default:
		answerCallback(ctx, b, cb.ID, "", false)
	}
}

// HandlePaymentCallback handles payment method selection callbacks.
func HandlePaymentCallback(ctx context.Context, b *bot.Bot, u *models.Update, deps commands.Deps) {
	if u.CallbackQuery == nil {
		return
	}

	cb := u.CallbackQuery
	lg := logger.ForUser(cb.From.ID)
	lang := deps.Sessions.GetLang(cb.From.ID)
	if lang == "" {
		lang = "en"
	}

	token := deps.Sessions.GetToken(cb.From.ID)
	if token == "" {
		answerCallback(ctx, b, cb.ID, "Please /start first", true)
		return
	}

	// Parse callback data: "pay:select:ID" or "pay:back"
	parts := strings.Split(strings.TrimPrefix(cb.Data, "pay:"), ":")
	if len(parts) < 1 {
		return
	}

	action := parts[0]
	lg.Debugf("Payment callback: action=%s parts=%v", action, parts)

	switch action {
	case "select":
		if len(parts) < 3 {
			return
		}
		paymentID, _ := strconv.ParseInt(parts[1], 10, 64)
		// parts[2] is payment name (URL encoded or just passed)
		paymentName := strings.Join(parts[2:], ":")
		answerCallback(ctx, b, cb.ID, "", false)

		// Save payment to state
		state := purchaseStates[cb.From.ID]
		if state == nil {
			lg.Errorf("No purchase state found")
			return
		}
		state.PaymentID = paymentID
		state.PaymentName = paymentName

		// Show order confirmation
		showOrderConfirmation(ctx, b, cb.Message.Message, deps, token, lang, cb.From.ID, lg)

	case "back":
		// Go back to quantity selection
		answerCallback(ctx, b, cb.ID, "", false)
		state := purchaseStates[cb.From.ID]
		if state != nil {
			showQuantitySelection(ctx, b, cb.Message.Message, lang, state.PlanName, state.Discounts, state.UnitTime, cb.From.ID)
		}

	default:
		answerCallback(ctx, b, cb.ID, "", false)
	}
}

// HandleOrderCallback handles order confirmation and cancellation.
func HandleOrderCallback(ctx context.Context, b *bot.Bot, u *models.Update, deps commands.Deps) {
	if u.CallbackQuery == nil {
		return
	}

	cb := u.CallbackQuery
	lg := logger.ForUser(cb.From.ID)
	lang := deps.Sessions.GetLang(cb.From.ID)
	if lang == "" {
		lang = "en"
	}

	token := deps.Sessions.GetToken(cb.From.ID)
	if token == "" {
		answerCallback(ctx, b, cb.ID, "Please /start first", true)
		return
	}

	// Parse callback data: "order:confirm" or "order:cancel" or "order:back"
	action := strings.TrimPrefix(cb.Data, "order:")
	lg.Debugf("Order callback: action=%s", action)

	loc := i18n.Localizer(lang)

	switch action {
	case "confirm":
		answerCallback(ctx, b, cb.ID, "", false)
		confirmAndCreateOrder(ctx, b, cb.Message.Message, deps, token, lang, cb.From.ID, lg)

	case "cancel":
		answerCallback(ctx, b, cb.ID, "", false)
		// Clear state
		delete(purchaseStates, cb.From.ID)
		// Edit message to show cancelled
		editMessageText(ctx, b, cb.Message.Message.Chat.ID, cb.Message.Message.ID,
			i18n.T(loc, "order_cancelled"), nil)

	case "back":
		// Go back to payment selection
		answerCallback(ctx, b, cb.ID, "", false)
		state := purchaseStates[cb.From.ID]
		if state == nil {
			// State expired or order already confirmed - show error and plan selection
			lg.Warnf("No state found on order:back, redirecting to plans")
			showPlanPage(ctx, b, cb.Message.Message, deps, token, lang, 0, lg)
			return
		}
		showPaymentSelection(ctx, b, cb.Message.Message, deps, token, lang, cb.From.ID, lg)

	default:
		answerCallback(ctx, b, cb.ID, "", false)
	}
}

// --- Helper Functions ---

// showPlanPage displays a single plan with navigation buttons.
func showPlanPage(ctx context.Context, b *bot.Bot, msg *models.Message, deps commands.Deps, token, lang string, page int, lg logger.TgLogger) {
	loc := i18n.Localizer(lang)

	plans, err := deps.API.GetSubscribePlans(ctx, token, lang)
	if err != nil {
		lg.Errorf("Get plans error: %v", err)
		editMessageText(ctx, b, msg.Chat.ID, msg.ID, i18n.T(loc, "error_loading_plans"), nil)
		return
	}

	if len(plans) == 0 {
		editMessageText(ctx, b, msg.Chat.ID, msg.ID, i18n.T(loc, "no_plans_available"), nil)
		return
	}

	// Ensure page is in bounds
	if page < 0 {
		page = 0
	}
	if page >= len(plans) {
		page = len(plans) - 1
	}

	plan := plans[page]
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
		"Current": page + 1,
		"Total":   len(plans),
	})

	// Build navigation keyboard
	var navRow []models.InlineKeyboardButton

	if page > 0 {
		navRow = append(navRow, models.InlineKeyboardButton{
			Text:         i18n.T(loc, "btn_prev"),
			CallbackData: fmt.Sprintf("plan:page:%d", page-1),
		})
	}

	if page < len(plans)-1 {
		navRow = append(navRow, models.InlineKeyboardButton{
			Text:         i18n.T(loc, "btn_next"),
			CallbackData: fmt.Sprintf("plan:page:%d", page+1),
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

	editMessageText(ctx, b, msg.Chat.ID, msg.ID, text, keyboard)
}

// startQuantitySelection initializes and shows quantity selection.
func startQuantitySelection(ctx context.Context, b *bot.Bot, msg *models.Message, deps commands.Deps, token, lang string, planID int64, userID int64, lg logger.TgLogger) {
	loc := i18n.Localizer(lang)

	// Fetch plan details
	plans, err := deps.API.GetSubscribePlans(ctx, token, lang)
	if err != nil {
		lg.Errorf("Get plans error: %v", err)
		editMessageText(ctx, b, msg.Chat.ID, msg.ID, i18n.T(loc, "error_loading_plans"), nil)
		return
	}

	var selectedPlan *api.SubscribePlan
	for _, p := range plans {
		if p.ID == planID {
			selectedPlan = &p
			break
		}
	}

	if selectedPlan == nil {
		lg.Errorf("Plan not found: %d", planID)
		return
	}

	// Initialize purchase state with user ID as key
	purchaseStates[userID] = &PurchaseState{
		PlanID:        selectedPlan.ID,
		PlanName:      selectedPlan.Name,
		PlanUnitPrice: selectedPlan.UnitPrice,
		UnitTime:      selectedPlan.UnitTime,
		Discounts:     selectedPlan.Discount,
	}

	showQuantitySelection(ctx, b, msg, lang, selectedPlan.Name, selectedPlan.Discount, selectedPlan.UnitTime, userID)
}

// showQuantitySelection displays duration/quantity selection.
func showQuantitySelection(ctx context.Context, b *bot.Bot, msg *models.Message, lang, planName string, discounts []api.SubscribeDiscount, unitTime string, userID int64) {
	loc := i18n.Localizer(lang)

	text := i18n.TWithData(loc, "select_quantity", map[string]any{
		"PlanName": planName,
	})

	// Build quantity buttons: always include 1 (base unit), then additional quantities from discount data
	var allQuantities []api.SubscribeDiscount

	// Add base unit (quantity 1) first - no discount
	allQuantities = append(allQuantities, api.SubscribeDiscount{Quantity: 1, Discount: 0})

	// Add quantities from API discount array
	for _, d := range discounts {
		// Skip if quantity is 1 (already added)
		if d.Quantity != 1 {
			allQuantities = append(allQuantities, d)
		}
	}

	// Build buttons
	var rows [][]models.InlineKeyboardButton
	var currentRow []models.InlineKeyboardButton

	for i, d := range allQuantities {
		var btnText string
		if d.Quantity == 1 {
			btnText = i18n.TWithData(loc, "btn_month", map[string]any{"Months": d.Quantity})
		} else {
			btnText = i18n.TWithData(loc, "btn_months", map[string]any{"Months": d.Quantity})
		}
		// Add discount indicator if applicable
		if d.Discount > 0 {
			btnText = fmt.Sprintf("%s (-%d%%)", btnText, d.Discount)
		}
		currentRow = append(currentRow, models.InlineKeyboardButton{
			Text:         btnText,
			CallbackData: fmt.Sprintf("qty:months:%d", d.Quantity),
		})
		// Two buttons per row
		if len(currentRow) == 2 || i == len(allQuantities)-1 {
			rows = append(rows, currentRow)
			currentRow = nil
		}
	}

	// Add back button
	rows = append(rows, []models.InlineKeyboardButton{
		{Text: i18n.T(loc, "btn_back"), CallbackData: "qty:back"},
	})

	keyboard := &models.InlineKeyboardMarkup{InlineKeyboard: rows}
	editMessageText(ctx, b, msg.Chat.ID, msg.ID, text, keyboard)
}

// showPaymentSelection displays available payment methods.
func showPaymentSelection(ctx context.Context, b *bot.Bot, msg *models.Message, deps commands.Deps, token, lang string, userID int64, lg logger.TgLogger) {
	loc := i18n.Localizer(lang)

	state := purchaseStates[userID]
	if state == nil {
		lg.Errorf("No purchase state found for user %d", userID)
		return
	}

	payments, err := deps.API.GetPaymentMethods(ctx, token)
	if err != nil {
		lg.Errorf("Get payment methods error: %v", err)
		editMessageText(ctx, b, msg.Chat.ID, msg.ID, i18n.T(loc, "error_loading_payments"), nil)
		return
	}

	text := i18n.TWithData(loc, "select_payment", map[string]any{
		"PlanName": state.PlanName,
		"Months":   state.Quantity,
	})

	// Build payment buttons
	var rows [][]models.InlineKeyboardButton
	for _, pm := range payments {
		rows = append(rows, []models.InlineKeyboardButton{
			{
				Text:         pm.Name,
				CallbackData: fmt.Sprintf("pay:select:%d:%s", pm.ID, pm.Name),
			},
		})
	}

	// Add back button
	rows = append(rows, []models.InlineKeyboardButton{
		{Text: i18n.T(loc, "btn_back"), CallbackData: "pay:back"},
	})

	keyboard := &models.InlineKeyboardMarkup{InlineKeyboard: rows}
	editMessageText(ctx, b, msg.Chat.ID, msg.ID, text, keyboard)
}

// showOrderConfirmation displays the order summary for confirmation.
func showOrderConfirmation(ctx context.Context, b *bot.Bot, msg *models.Message, deps commands.Deps, token, lang string, userID int64, lg logger.TgLogger) {
	loc := i18n.Localizer(lang)

	state := purchaseStates[userID]
	if state == nil {
		lg.Errorf("No purchase state found for user %d", userID)
		return
	}

	// Get price preview
	preOrder, err := deps.API.PreOrder(ctx, token, api.PreOrderRequest{
		SubscribeID: state.PlanID,
		Quantity:    state.Quantity,
		Payment:     state.PaymentID,
	})

	var totalAmount int64
	if err != nil {
		lg.Warnf("PreOrder error: %v, using calculated price", err)
		totalAmount = state.PlanUnitPrice * state.Quantity
	} else {
		totalAmount = preOrder.Amount
	}

	text := i18n.TWithData(loc, "confirm_order", map[string]any{
		"PlanName":    state.PlanName,
		"Months":      state.Quantity,
		"PaymentName": state.PaymentName,
		"Total":       totalAmount,
	})

	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: i18n.T(loc, "btn_confirm"), CallbackData: "order:confirm"},
			},
			{
				{Text: i18n.T(loc, "btn_back"), CallbackData: "order:back"},
				{Text: i18n.T(loc, "btn_cancel"), CallbackData: "order:cancel"},
			},
		},
	}

	editMessageText(ctx, b, msg.Chat.ID, msg.ID, text, keyboard)
}

// confirmAndCreateOrder creates the order and shows payment link.
func confirmAndCreateOrder(ctx context.Context, b *bot.Bot, msg *models.Message, deps commands.Deps, token, lang string, userID int64, lg logger.TgLogger) {
	loc := i18n.Localizer(lang)

	state := purchaseStates[userID]
	if state == nil {
		lg.Errorf("No purchase state found for user %d", userID)
		editMessageText(ctx, b, msg.Chat.ID, msg.ID, i18n.T(loc, "error_creating_order"), nil)
		return
	}

	// Create order
	orderResp, err := deps.API.PurchaseOrder(ctx, token, api.PurchaseOrderRequest{
		SubscribeID: state.PlanID,
		Quantity:    state.Quantity,
		Payment:     state.PaymentID,
	})

	if err != nil {
		lg.Errorf("PurchaseOrder error: %v", err)
		editMessageText(ctx, b, msg.Chat.ID, msg.ID, i18n.T(loc, "error_creating_order"), nil)
		return
	}

	// Clear purchase state BEFORE editing message to prevent race conditions
	delete(purchaseStates, userID)

	// Process payment via checkout endpoint
	// This is the step that actually deducts balance for balance payments
	checkoutResp, err := deps.API.CheckoutOrder(ctx, token, api.CheckoutOrderRequest{
		OrderNo: orderResp.OrderNo,
	})
	if err != nil {
		lg.Warnf("CheckoutOrder error: %v", err)
	} else {
		lg.Debugf("Checkout completed: order_no=%s type=%s checkout_url=%s", orderResp.OrderNo, checkoutResp.Type, checkoutResp.CheckoutURL)
	}

	// Get order details for status check
	orderDetail, err := deps.API.GetOrderDetail(ctx, token, orderResp.OrderNo)
	if err != nil {
		lg.Warnf("GetOrderDetail error: %v", err)
	} else {
		lg.Debugf("Order status: order_no=%s status=%d", orderResp.OrderNo, orderDetail.Status)
	}

	// Choose message based on order status and payment type
	// See api/codes.go for Order Status constants
	var text string
	var keyboard *models.InlineKeyboardMarkup

	if orderDetail != nil && orderDetail.Status == api.OrderStatusFinished {
		// Order is paid (balance payment completed)
		text = i18n.TWithData(loc, "order_created_balance", map[string]any{
			"OrderNo": orderResp.OrderNo,
		})
		keyboard = &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{},
		}
	} else if checkoutResp != nil && checkoutResp.CheckoutURL != "" {
		// Order pending, external payment (show Pay Now button from checkout response)
		text = i18n.TWithData(loc, "order_created", map[string]any{
			"OrderNo": orderResp.OrderNo,
		})
		keyboard = &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: i18n.T(loc, "btn_pay_now"), URL: checkoutResp.CheckoutURL},
				},
			},
		}
	} else {
		// Fallback - order created but payment not completed
		text = i18n.TWithData(loc, "order_pending", map[string]any{
			"OrderNo": orderResp.OrderNo,
		})
		keyboard = &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{},
		}
	}

	editMessageTextPlain(ctx, b, msg.Chat.ID, msg.ID, text, keyboard)
}

// --- UI Helpers ---

func editMessageText(ctx context.Context, b *bot.Bot, chatID int64, msgID int, text string, keyboard *models.InlineKeyboardMarkup) {
	params := &bot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: msgID,
		Text:      text,
		ParseMode: models.ParseModeMarkdown,
	}
	if keyboard != nil {
		params.ReplyMarkup = keyboard
	}
	_, _ = b.EditMessageText(ctx, params)
}

// editMessageTextPlain edits message without markdown parsing (for text with special chars like order numbers)
func editMessageTextPlain(ctx context.Context, b *bot.Bot, chatID int64, msgID int, text string, keyboard *models.InlineKeyboardMarkup) {
	params := &bot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: msgID,
		Text:      text,
	}
	if keyboard != nil {
		params.ReplyMarkup = keyboard
	}
	_, _ = b.EditMessageText(ctx, params)
}
