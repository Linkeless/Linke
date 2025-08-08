package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/domains/user/entities"
	userInterfaces "linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/config"
	"linke/internal/shared/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BotEnhanced represents an enhanced Telegram bot with friendly UI
type BotEnhanced struct {
	api                     *tgbotapi.BotAPI
	userService            userInterfaces.UserService
	subscriptionService    interfaces.UserSubscriptionService
	subscriptionPlanService interfaces.SubscriptionPlanService
	cfg                    *config.Config
	userStates             map[int64]string // Track user conversation states
}

// NewBotEnhanced creates a new enhanced Telegram bot instance
func NewBotEnhanced(
	cfg *config.Config,
	userService userInterfaces.UserService,
	subscriptionService interfaces.UserSubscriptionService,
	subscriptionPlanService interfaces.SubscriptionPlanService,
) (*BotEnhanced, error) {
	if cfg.OAuth2.TelegramBotToken == "" {
		return nil, fmt.Errorf("telegram bot token is not configured")
	}

	bot, err := tgbotapi.NewBotAPI(cfg.OAuth2.TelegramBotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	return &BotEnhanced{
		api:                     bot,
		userService:            userService,
		subscriptionService:    subscriptionService,
		subscriptionPlanService: subscriptionPlanService,
		cfg:                    cfg,
		userStates:             make(map[int64]string),
	}, nil
}

// Start starts the enhanced bot and listens for updates
func (b *BotEnhanced) Start(ctx context.Context) error {
	logger.Info("Starting Enhanced Telegram bot", logger.String("username", b.api.Self.UserName))

	// Configure bot settings on startup
	if err := b.configureBotSettings(); err != nil {
		logger.Error("Failed to configure bot settings", logger.Error2("error", err))
		// Continue anyway, don't fail startup
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping Enhanced Telegram bot")
			return ctx.Err()
		case update := <-updates:
			go b.handleUpdate(update)
		}
	}
}

// handleUpdate processes all types of updates
func (b *BotEnhanced) handleUpdate(update tgbotapi.Update) {
	// Handle callback queries (button clicks)
	if update.CallbackQuery != nil {
		b.handleCallbackQuery(update.CallbackQuery)
		return
	}

	// Handle messages
	if update.Message != nil {
		// Check if it's a command
		if update.Message.IsCommand() {
			b.handleCommand(update.Message)
		} else {
			// Handle natural language or conversation flow
			b.handleMessage(update.Message)
		}
	}
}

// handleCommand processes command messages
func (b *BotEnhanced) handleCommand(msg *tgbotapi.Message) {
	logger.Info("Received command",
		logger.String("command", msg.Command()),
		logger.String("from", msg.From.UserName),
		logger.Int64("chat_id", msg.Chat.ID))

	switch msg.Command() {
	case "start":
		b.showMainMenu(msg.Chat.ID)
	case "menu":
		b.showMainMenu(msg.Chat.ID)
	case "help":
		b.showHelp(msg.Chat.ID)
	default:
		b.sendMessage(msg.Chat.ID, "未知命令。请使用 /menu 查看主菜单。")
	}
}

// handleMessage processes non-command messages
func (b *BotEnhanced) handleMessage(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := strings.ToLower(msg.Text)
	
	// Check user state for conversation flow
	state, exists := b.userStates[chatID]
	if exists {
		b.handleConversationFlow(msg, state)
		return
	}

	// Natural language understanding
	switch {
	case strings.Contains(text, "订阅") || strings.Contains(text, "套餐"):
		b.showSubscriptionMenu(chatID)
	case strings.Contains(text, "流量") || strings.Contains(text, "使用"):
		b.showUsageMenu(chatID)
	case strings.Contains(text, "帮助") || strings.Contains(text, "help"):
		b.showHelp(chatID)
	case strings.Contains(text, "价格") || strings.Contains(text, "购买"):
		b.showPlansMenu(chatID)
	default:
		// Show quick reply keyboard for common actions
		b.showQuickActions(chatID)
	}
}

// handleCallbackQuery processes button clicks
func (b *BotEnhanced) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	// Answer callback to remove loading state
	callback := tgbotapi.NewCallback(query.ID, "")
	b.api.Request(callback)

	// Parse callback data
	data := query.Data
	parts := strings.Split(data, ":")
	
	if len(parts) < 1 {
		return
	}

	action := parts[0]
	
	logger.Info("Processing callback",
		logger.String("action", action),
		logger.String("from", query.From.UserName))

	switch action {
	case "main_menu":
		b.editMessageToMainMenu(query.Message.Chat.ID, query.Message.MessageID)
	case "subscription":
		b.editMessageToSubscriptionMenu(query.Message.Chat.ID, query.Message.MessageID)
	case "subscription_info":
		b.showSubscriptionInfo(query.Message.Chat.ID, query.Message.MessageID)
	case "usage_info":
		b.showUsageDetails(query.Message.Chat.ID, query.Message.MessageID)
	case "plans":
		b.editMessageToPlansMenu(query.Message.Chat.ID, query.Message.MessageID)
	case "plan_details":
		if len(parts) > 1 {
			planID, _ := strconv.Atoi(parts[1])
			b.showPlanDetails(query.Message.Chat.ID, query.Message.MessageID, uint(planID))
		}
	case "support":
		b.showSupportMenu(query.Message.Chat.ID, query.Message.MessageID)
	case "settings":
		b.showSettingsMenu(query.Message.Chat.ID, query.Message.MessageID)
	case "toggle_autorenew":
		b.toggleAutoRenew(query.Message.Chat.ID, query.Message.MessageID)
	case "back":
		if len(parts) > 1 {
			b.navigateBack(query.Message.Chat.ID, query.Message.MessageID, parts[1])
		}
	}
}

// showMainMenu displays the main menu with inline keyboard
func (b *BotEnhanced) showMainMenu(chatID int64) {
	text := "🏠 *主菜单*\n\n请选择您需要的功能："

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 我的订阅", "subscription"),
			tgbotapi.NewInlineKeyboardButtonData("💎 套餐商店", "plans"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 使用统计", "usage_info"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ 设置", "settings"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💬 客服支持", "support"),
			tgbotapi.NewInlineKeyboardButtonData("❓ 帮助", "help"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	
	b.api.Send(msg)
}

// editMessageToMainMenu edits existing message to show main menu
func (b *BotEnhanced) editMessageToMainMenu(chatID int64, messageID int) {
	text := "🏠 *主菜单*\n\n请选择您需要的功能："

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 我的订阅", "subscription"),
			tgbotapi.NewInlineKeyboardButtonData("💎 套餐商店", "plans"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 使用统计", "usage_info"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ 设置", "settings"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💬 客服支持", "support"),
			tgbotapi.NewInlineKeyboardButtonData("❓ 帮助", "help"),
		),
	)

	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// showSubscriptionMenu displays subscription management menu
func (b *BotEnhanced) showSubscriptionMenu(chatID int64) {
	text := "📊 *订阅管理*\n\n请选择操作："

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 查看详情", "subscription_info"),
			tgbotapi.NewInlineKeyboardButtonData("📈 流量使用", "usage_info"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 续费设置", "toggle_autorenew"),
			tgbotapi.NewInlineKeyboardButtonData("⏸️ 暂停订阅", "pause_subscription"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬆️ 升级套餐", "upgrade_plan"),
			tgbotapi.NewInlineKeyboardButtonData("📝 订阅历史", "subscription_history"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回主菜单", "main_menu"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	
	b.api.Send(msg)
}

// editMessageToSubscriptionMenu edits message to show subscription menu
func (b *BotEnhanced) editMessageToSubscriptionMenu(chatID int64, messageID int) {
	text := "📊 *订阅管理*\n\n请选择操作："

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 查看详情", "subscription_info"),
			tgbotapi.NewInlineKeyboardButtonData("📈 流量使用", "usage_info"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 续费设置", "toggle_autorenew"),
			tgbotapi.NewInlineKeyboardButtonData("⏸️ 暂停订阅", "pause_subscription"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬆️ 升级套餐", "upgrade_plan"),
			tgbotapi.NewInlineKeyboardButtonData("📝 订阅历史", "subscription_history"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回主菜单", "main_menu"),
		),
	)

	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// showQuickActions displays quick action buttons as reply keyboard
func (b *BotEnhanced) showQuickActions(chatID int64) {
	text := "您好！我可以帮您做什么？"

	// Create custom keyboard with common actions
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📊 查看订阅"),
			tgbotapi.NewKeyboardButton("📈 流量使用"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("💎 套餐价格"),
			tgbotapi.NewKeyboardButton("💬 联系客服"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🏠 主菜单"),
		),
	)
	keyboard.ResizeKeyboard = true

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	
	b.api.Send(msg)
}

// showSubscriptionInfo displays detailed subscription information
func (b *BotEnhanced) showSubscriptionInfo(chatID int64, messageID int) {
	ctx := context.Background()
	
	// Get user by telegram ID
	user, err := b.getUserByTelegramID(ctx, chatID)
	if err != nil {
		b.showUnboundAccount(chatID, messageID)
		return
	}

	// Get user's active subscriptions
	subscriptions, err := b.subscriptionService.GetUserActiveSubscriptions(ctx, user.ID)
	if err != nil {
		b.showError(chatID, messageID, "获取订阅信息失败")
		return
	}

	if len(subscriptions) == 0 {
		b.showNoSubscription(chatID, messageID)
		return
	}

	// Build subscription info message
	var sb strings.Builder
	sb.WriteString("📋 *您的订阅详情*\n\n")

	for i, sub := range subscriptions {
		plan, _ := b.subscriptionPlanService.GetSubscriptionPlan(ctx, sub.SubscriptionPlanID)
		if plan == nil {
			continue
		}

		if i > 0 {
			sb.WriteString("\n➖➖➖➖➖➖➖➖➖\n\n")
		}

		sb.WriteString(fmt.Sprintf("*%s*\n", plan.Name))
		sb.WriteString(fmt.Sprintf("📌 状态: %s\n", b.formatStatus(sub.Status)))
		sb.WriteString(fmt.Sprintf("💰 价格: %.2f %s/%s\n", 
			sub.Price, 
			sub.Currency,
			b.formatBillingCycleShort(sub.BillingCycle)))
		
		// Traffic info with progress bar
		if sub.TrafficLimit > 0 {
			percentage := float64(sub.TrafficUsed) / float64(sub.TrafficLimit) * 100
			sb.WriteString(fmt.Sprintf("📊 流量: %s / %s\n", 
				b.formatBytes(sub.TrafficUsed), 
				b.formatBytes(sub.TrafficLimit)))
			sb.WriteString(fmt.Sprintf("%s %.1f%%\n", 
				b.getProgressBar(percentage), 
				percentage))
		}
		
		// Date info
		if sub.CurrentPeriodEnd != nil {
			now := time.Now()
			daysLeft := int(sub.CurrentPeriodEnd.Sub(now).Hours() / 24)
			if daysLeft > 0 {
				sb.WriteString(fmt.Sprintf("📅 剩余天数: %d 天\n", daysLeft))
			}
		}
		
		// Auto-renew status
		if sub.AutoRenew {
			sb.WriteString("🔄 自动续费: ✅ 已开启\n")
		} else {
			sb.WriteString("🔄 自动续费: ❌ 已关闭\n")
		}
	}

	// Add action buttons
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 查看使用详情", "usage_info"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ 管理订阅", "subscription"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回", "back:subscription"),
		),
	)

	edit := tgbotapi.NewEditMessageText(chatID, messageID, sb.String())
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// showUsageDetails displays detailed usage information
func (b *BotEnhanced) showUsageDetails(chatID int64, messageID int) {
	ctx := context.Background()
	
	user, err := b.getUserByTelegramID(ctx, chatID)
	if err != nil {
		b.showUnboundAccount(chatID, messageID)
		return
	}

	subscriptions, err := b.subscriptionService.GetUserActiveSubscriptions(ctx, user.ID)
	if err != nil || len(subscriptions) == 0 {
		b.showNoSubscription(chatID, messageID)
		return
	}

	var sb strings.Builder
	sb.WriteString("📈 *流量使用详情*\n\n")

	for _, sub := range subscriptions {
		plan, _ := b.subscriptionPlanService.GetSubscriptionPlan(ctx, sub.SubscriptionPlanID)
		if plan == nil || sub.TrafficLimit == 0 {
			continue
		}

		used := sub.TrafficUsed
		total := sub.TrafficLimit
		remaining := total - used
		percentage := float64(used) / float64(total) * 100

		sb.WriteString(fmt.Sprintf("*%s*\n", plan.Name))
		sb.WriteString(fmt.Sprintf("已使用: %s\n", b.formatBytes(used)))
		sb.WriteString(fmt.Sprintf("总流量: %s\n", b.formatBytes(total)))
		sb.WriteString(fmt.Sprintf("剩余: %s\n", b.formatBytes(remaining)))
		sb.WriteString(fmt.Sprintf("%s %.1f%%\n\n", b.getProgressBar(percentage), percentage))

		// Add usage tips based on percentage
		if percentage > 90 {
			sb.WriteString("⚠️ *流量即将用尽，建议升级套餐*\n")
		} else if percentage > 70 {
			sb.WriteString("💡 流量使用较多，请注意控制\n")
		}
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬆️ 升级套餐", "upgrade_plan"),
			tgbotapi.NewInlineKeyboardButtonData("📊 订阅详情", "subscription_info"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回", "back:subscription"),
		),
	)

	edit := tgbotapi.NewEditMessageText(chatID, messageID, sb.String())
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// showUsageMenu displays usage information menu (new message)
func (b *BotEnhanced) showUsageMenu(chatID int64) {
	text := "📈 *流量使用*\n\n查看您的流量使用详情："

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 查看详情", "usage_info"),
			tgbotapi.NewInlineKeyboardButtonData("📋 订阅信息", "subscription_info"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回主菜单", "main_menu"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	
	b.api.Send(msg)
}

// showPlansMenu displays available plans (new message)
func (b *BotEnhanced) showPlansMenu(chatID int64) {
	ctx := context.Background()
	
	plans, _, err := b.subscriptionPlanService.GetSubscriptionPlans(ctx, &interfaces.GetSubscriptionPlansRequest{
		Status: "active",
		Limit:  10,
	})
	
	if err != nil || len(plans) == 0 {
		b.sendMessage(chatID, "❌ 获取套餐信息失败，请稍后重试。")
		return
	}

	var sb strings.Builder
	sb.WriteString("💎 *套餐商店*\n\n选择套餐查看详情：\n\n")

	var buttons [][]tgbotapi.InlineKeyboardButton
	
	for _, plan := range plans {
		if !plan.IsVisible {
			continue
		}
		
		buttonText := fmt.Sprintf("%s - %.2f %s/%s", 
			plan.Name, 
			plan.Price, 
			plan.Currency,
			b.formatBillingCycleShort(plan.BillingCycle))
		
		if plan.IsRecommended {
			buttonText = "⭐ " + buttonText
		}
		
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(buttonText, fmt.Sprintf("plan_details:%d", plan.ID)),
		))
	}
	
	// Add back button
	buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 返回主菜单", "main_menu"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg := tgbotapi.NewMessage(chatID, sb.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	
	b.api.Send(msg)
}

// editMessageToPlansMenu edits existing message to show plans menu
func (b *BotEnhanced) editMessageToPlansMenu(chatID int64, messageID int) {
	ctx := context.Background()
	
	plans, _, err := b.subscriptionPlanService.GetSubscriptionPlans(ctx, &interfaces.GetSubscriptionPlansRequest{
		Status: "active",
		Limit:  10,
	})
	
	if err != nil || len(plans) == 0 {
		b.showError(chatID, messageID, "获取套餐信息失败")
		return
	}

	var sb strings.Builder
	sb.WriteString("💎 *套餐商店*\n\n选择套餐查看详情：\n\n")

	var buttons [][]tgbotapi.InlineKeyboardButton
	
	for _, plan := range plans {
		if !plan.IsVisible {
			continue
		}
		
		buttonText := fmt.Sprintf("%s - %.2f %s/%s", 
			plan.Name, 
			plan.Price, 
			plan.Currency,
			b.formatBillingCycleShort(plan.BillingCycle))
		
		if plan.IsRecommended {
			buttonText = "⭐ " + buttonText
		}
		
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(buttonText, fmt.Sprintf("plan_details:%d", plan.ID)),
		))
	}
	
	// Add back button
	buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 返回主菜单", "main_menu"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	edit := tgbotapi.NewEditMessageText(chatID, messageID, sb.String())
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// showPlanDetails shows detailed information about a specific plan
func (b *BotEnhanced) showPlanDetails(chatID int64, messageID int, planID uint) {
	ctx := context.Background()
	
	plan, err := b.subscriptionPlanService.GetSubscriptionPlan(ctx, planID)
	if err != nil {
		b.showError(chatID, messageID, "获取套餐详情失败")
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("💎 *%s*\n\n", plan.Name))
	
	if plan.Description != "" {
		sb.WriteString(fmt.Sprintf("📝 %s\n\n", plan.Description))
	}
	
	sb.WriteString(fmt.Sprintf("💰 *价格:* %.2f %s/%s\n", 
		plan.Price, 
		plan.Currency,
		b.formatBillingCycle(plan.BillingCycle)))
	
	if plan.TrafficLimit > 0 {
		sb.WriteString(fmt.Sprintf("📊 *流量:* %s/月\n", b.formatBytes(plan.TrafficLimit)))
	} else {
		sb.WriteString("📊 *流量:* 无限制\n")
	}
	
	if plan.TrialPeriodDays > 0 {
		sb.WriteString(fmt.Sprintf("🎁 *试用期:* %d 天免费试用\n", plan.TrialPeriodDays))
	}
	
	// Add features if available
	sb.WriteString("\n*套餐特色:*\n")
	sb.WriteString("✅ 高速稳定连接\n")
	sb.WriteString("✅ 全球节点覆盖\n")
	sb.WriteString("✅ 7×24 技术支持\n")
	
	if plan.IsRecommended {
		sb.WriteString("\n⭐ *推荐套餐* - 性价比最高的选择！\n")
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🛒 立即购买", fmt.Sprintf("purchase:%d", planID)),
			tgbotapi.NewInlineKeyboardButtonData("💬 咨询客服", "support"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回套餐列表", "plans"),
		),
	)

	edit := tgbotapi.NewEditMessageText(chatID, messageID, sb.String())
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// Helper functions

func (b *BotEnhanced) getUserByTelegramID(ctx context.Context, telegramID int64) (*entities.User, error) {
	telegramIDStr := fmt.Sprintf("%d", telegramID)
	return b.userService.GetUserByTelegramID(ctx, telegramIDStr)
}

func (b *BotEnhanced) showUnboundAccount(chatID int64, messageID int) {
	text := "❌ *账号未绑定*\n\n" +
		"您还没有绑定账号。请先在网站上使用 Telegram 登录以绑定您的账号。\n\n" +
		"绑定步骤：\n" +
		"1. 访问我们的网站\n" +
		"2. 选择 Telegram 登录\n" +
		"3. 授权登录即可完成绑定"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🌐 访问网站", "https://your-website.com"),
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回", "main_menu"),
		),
	)

	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

func (b *BotEnhanced) showNoSubscription(chatID int64, messageID int) {
	text := "📭 *没有活跃订阅*\n\n" +
		"您当前没有活跃的订阅。\n" +
		"立即选择合适的套餐开始使用我们的服务！"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💎 查看套餐", "plans"),
			tgbotapi.NewInlineKeyboardButtonData("💬 联系客服", "support"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回主菜单", "main_menu"),
		),
	)

	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

func (b *BotEnhanced) showError(chatID int64, messageID int, errorMsg string) {
	text := fmt.Sprintf("❌ *错误*\n\n%s\n\n请稍后重试或联系客服。", errorMsg)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 重试", "main_menu"),
			tgbotapi.NewInlineKeyboardButtonData("💬 联系客服", "support"),
		),
	)

	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

func (b *BotEnhanced) showHelp(chatID int64) {
	help := `📚 *使用帮助*

*快速开始*
• 点击按钮即可操作，无需记住命令
• 直接输入关键词，如"订阅"、"流量"等
• 使用底部快捷键盘快速访问功能

*主要功能*
📊 *我的订阅* - 查看和管理您的订阅
💎 *套餐商店* - 浏览和购买套餐
📈 *使用统计* - 查看流量使用情况
⚙️ *设置* - 管理账号设置
💬 *客服支持* - 获取帮助和支持

*常见问题*
Q: 如何绑定账号？
A: 在网站使用 Telegram 登录即可自动绑定

Q: 如何查看剩余流量？
A: 点击"使用统计"查看详细信息

Q: 如何升级套餐？
A: 在"套餐商店"选择新套餐购买

需要更多帮助？点击"客服支持"联系我们！`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 返回主菜单", "main_menu"),
			tgbotapi.NewInlineKeyboardButtonData("💬 联系客服", "support"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, help)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	
	b.api.Send(msg)
}

func (b *BotEnhanced) showSupportMenu(chatID int64, messageID int) {
	text := "💬 *客服支持*\n\n" +
		"请选择您需要的帮助类型：\n\n" +
		"工作时间：周一至周日 9:00-21:00\n" +
		"紧急问题请直接联系在线客服"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("👨‍💻 在线客服", "https://t.me/your_support"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ 常见问题", "faq"),
			tgbotapi.NewInlineKeyboardButtonData("📝 提交工单", "ticket"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("📧 发送邮件", "mailto:support@your-domain.com"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回主菜单", "main_menu"),
		),
	)

	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

func (b *BotEnhanced) showSettingsMenu(chatID int64, messageID int) {
	text := "⚙️ *设置*\n\n选择要修改的设置："

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 自动续费设置", "toggle_autorenew"),
			tgbotapi.NewInlineKeyboardButtonData("🔔 通知设置", "notification_settings"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🌍 语言设置", "language_settings"),
			tgbotapi.NewInlineKeyboardButtonData("🔐 安全设置", "security_settings"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回主菜单", "main_menu"),
		),
	)

	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

func (b *BotEnhanced) toggleAutoRenew(chatID int64, messageID int) {
	// This would toggle auto-renewal setting
	// For now, just show a message
	text := "🔄 *自动续费设置*\n\n" +
		"当前状态：✅ 已开启\n\n" +
		"自动续费可以确保您的服务不会中断。"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ 关闭自动续费", "disable_autorenew"),
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回设置", "settings"),
		),
	)

	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

func (b *BotEnhanced) handleConversationFlow(msg *tgbotapi.Message, state string) {
	// Handle conversation states
	// This can be expanded for multi-step interactions
	chatID := msg.Chat.ID
	
	switch state {
	case "awaiting_feedback":
		// Process feedback
		b.sendMessage(chatID, "感谢您的反馈！我们会尽快处理。")
		delete(b.userStates, chatID)
	default:
		delete(b.userStates, chatID)
	}
}

func (b *BotEnhanced) navigateBack(chatID int64, messageID int, destination string) {
	switch destination {
	case "main":
		b.editMessageToMainMenu(chatID, messageID)
	case "subscription":
		b.editMessageToSubscriptionMenu(chatID, messageID)
	case "plans":
		b.editMessageToPlansMenu(chatID, messageID)
	default:
		b.editMessageToMainMenu(chatID, messageID)
	}
}

func (b *BotEnhanced) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	b.api.Send(msg)
}

func (b *BotEnhanced) formatStatus(status string) string {
	switch status {
	case "active":
		return "✅ 活跃"
	case "paused":
		return "⏸️ 暂停"
	case "cancelled":
		return "❌ 已取消"
	case "expired":
		return "⏰ 已过期"
	case "trial":
		return "🎁 试用中"
	default:
		return status
	}
}

func (b *BotEnhanced) formatBillingCycle(cycle string) string {
	switch cycle {
	case "monthly":
		return "月"
	case "yearly":
		return "年"
	case "lifetime":
		return "终身"
	default:
		return cycle
	}
}

func (b *BotEnhanced) formatBillingCycleShort(cycle string) string {
	return b.formatBillingCycle(cycle)
}

func (b *BotEnhanced) formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func (b *BotEnhanced) getProgressBar(percentage float64) string {
	const barLength = 10
	filled := int(percentage / 10)
	if filled > barLength {
		filled = barLength
	}
	
	var bar strings.Builder
	for i := 0; i < barLength; i++ {
		if i < filled {
			bar.WriteString("█")
		} else {
			bar.WriteString("░")
		}
	}
	
	return bar.String()
}

// configureBotSettings configures bot settings through Telegram API
func (b *BotEnhanced) configureBotSettings() error {
	logger.Info("Configuring bot settings...")

	// Set bot commands
	commands := []tgbotapi.BotCommand{
		{
			Command:     "start",
			Description: "开始使用",
		},
		{
			Command:     "menu",
			Description: "显示主菜单",
		},
		{
			Command:     "subscription",
			Description: "查看订阅",
		},
		{
			Command:     "usage",
			Description: "流量统计",
		},
		{
			Command:     "plans",
			Description: "套餐商店",
		},
		{
			Command:     "support",
			Description: "客服支持",
		},
		{
			Command:     "help",
			Description: "使用帮助",
		},
	}

	// Set commands
	setCommandsConfig := tgbotapi.NewSetMyCommands(commands...)
	if _, err := b.api.Request(setCommandsConfig); err != nil {
		logger.Error("Failed to set bot commands", logger.Error2("error", err))
		return fmt.Errorf("failed to set commands: %w", err)
	}
	logger.Info("Bot commands configured successfully")

	// Set bot description using direct API call
	descriptionParams := tgbotapi.Params{
		"description": "Linke 订阅管理助手\n\n" +
			"为您提供便捷的订阅管理服务：\n" +
			"• 实时查看订阅状态\n" +
			"• 监控流量使用情况\n" +
			"• 浏览和购买套餐\n" +
			"• 获取客服支持\n\n" +
			"点击下方按钮开始使用",
	}
	if _, err := b.api.MakeRequest("setMyDescription", descriptionParams); err != nil {
		logger.Error("Failed to set bot description", logger.Error2("error", err))
	}

	// Set bot short description using direct API call
	shortDescParams := tgbotapi.Params{
		"short_description": "Linke 订阅管理助手 - 轻松管理您的订阅服务",
	}
	if _, err := b.api.MakeRequest("setMyShortDescription", shortDescParams); err != nil {
		logger.Error("Failed to set bot short description", logger.Error2("error", err))
	}

	// Set bot name using direct API call
	nameParams := tgbotapi.Params{
		"name": "Linke 订阅助手",
	}
	if _, err := b.api.MakeRequest("setMyName", nameParams); err != nil {
		logger.Error("Failed to set bot name", logger.Error2("error", err))
	}

	// Note: Admin rights and menu button configuration are optional
	// These features may not be available in all versions of the telegram-bot-api library
	// They can be enabled when the library supports these features
	
	// TODO: Configure default admin rights when supported by library
	// adminRights := tgbotapi.ChatAdministratorRights{...}
	// setAdminRightsConfig := tgbotapi.NewSetMyDefaultAdministratorRights(adminRights, false)
	
	// TODO: Set chat menu button when supported by library
	// menuButton := tgbotapi.MenuButton{Type: "default"}
	// setMenuConfig := tgbotapi.NewSetChatMenuButton(0, &menuButton)

	logger.Info("Bot configuration completed successfully")
	return nil
}

// Ticket Notification Methods

// SendTicketNotification sends a formatted ticket notification to specified chat
func (b *BotEnhanced) SendTicketNotification(chatID int64, notification *TicketNotification) error {
	// Format the message based on notification type
	var message string
	var keyboard *tgbotapi.InlineKeyboardMarkup
	
	switch notification.Type {
	case TicketNotificationTypeCreated:
		message = b.formatTicketCreatedMessage(notification)
		keyboard = b.createTicketActionKeyboard(notification.TicketID, notification.TicketNo)
	case TicketNotificationTypeAssigned:
		message = b.formatTicketAssignedMessage(notification)
		keyboard = b.createTicketActionKeyboard(notification.TicketID, notification.TicketNo)
	case TicketNotificationTypeStatusChanged:
		message = b.formatTicketStatusChangedMessage(notification)
		keyboard = b.createTicketActionKeyboard(notification.TicketID, notification.TicketNo)
	case TicketNotificationTypeReplied:
		message = b.formatTicketRepliedMessage(notification)
		keyboard = b.createTicketActionKeyboard(notification.TicketID, notification.TicketNo)
	case TicketNotificationTypeResolved:
		message = b.formatTicketResolvedMessage(notification)
		keyboard = b.createTicketViewKeyboard(notification.TicketID, notification.TicketNo)
	case TicketNotificationTypeClosed:
		message = b.formatTicketClosedMessage(notification)
		keyboard = b.createTicketViewKeyboard(notification.TicketID, notification.TicketNo)
	default:
		message = b.formatGenericTicketMessage(notification)
		keyboard = b.createTicketActionKeyboard(notification.TicketID, notification.TicketNo)
	}
	
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}
	
	_, err := b.api.Send(msg)
	return err
}

// SendTicketNotificationToAdmins sends ticket notification to all admin users
func (b *BotEnhanced) SendTicketNotificationToAdmins(notification *TicketNotification) error {
	// Get admin chat IDs from configuration
	adminChatIDs := b.getAdminChatIDs()
	
	var errors []error
	for _, chatID := range adminChatIDs {
		if err := b.SendTicketNotification(chatID, notification); err != nil {
			logger.Error("Failed to send ticket notification to admin",
				logger.Int64("chat_id", chatID),
				logger.Error2("error", err))
			errors = append(errors, err)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to send to %d admins", len(errors))
	}
	
	return nil
}

// SendTicketNotificationToUser sends ticket notification to the ticket creator
func (b *BotEnhanced) SendTicketNotificationToUser(userTelegramID string, notification *TicketNotification) error {
	chatID, err := strconv.ParseInt(userTelegramID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid telegram ID: %w", err)
	}
	
	// Format user-friendly message (hide internal details)
	message := b.formatUserTicketMessage(notification)
	
	keyboard := b.createUserTicketKeyboard(notification.TicketID, notification.TicketNo)
	
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	
	_, err = b.api.Send(msg)
	return err
}

// Format methods for different notification types

func (b *BotEnhanced) formatTicketCreatedMessage(n *TicketNotification) string {
	var sb strings.Builder
	sb.WriteString("*新工单创建*\n\n")
	sb.WriteString(fmt.Sprintf("工单号: `%s`\n", n.TicketNo))
	sb.WriteString(fmt.Sprintf("标题: %s\n", n.Title))
	sb.WriteString(fmt.Sprintf("优先级: %s\n", b.formatPriority(n.Priority)))
	sb.WriteString(fmt.Sprintf("分类: %s\n", n.Category))
	sb.WriteString(fmt.Sprintf("创建者: %s\n", n.UserName))
	sb.WriteString(fmt.Sprintf("\n描述:\n%s\n", n.Description))
	
	if n.Priority == "urgent" || n.Priority == "critical" {
		sb.WriteString("\n⚠️ *需要立即处理*")
	}
	
	return sb.String()
}

func (b *BotEnhanced) formatTicketAssignedMessage(n *TicketNotification) string {
	var sb strings.Builder
	sb.WriteString("*工单已分配*\n\n")
	sb.WriteString(fmt.Sprintf("工单号: `%s`\n", n.TicketNo))
	sb.WriteString(fmt.Sprintf("标题: %s\n", n.Title))
	sb.WriteString(fmt.Sprintf("分配给: %s\n", n.AssignedToName))
	sb.WriteString(fmt.Sprintf("优先级: %s\n", b.formatPriority(n.Priority)))
	return sb.String()
}

func (b *BotEnhanced) formatTicketStatusChangedMessage(n *TicketNotification) string {
	var sb strings.Builder
	sb.WriteString("*工单状态更新*\n\n")
	sb.WriteString(fmt.Sprintf("工单号: `%s`\n", n.TicketNo))
	sb.WriteString(fmt.Sprintf("标题: %s\n", n.Title))
	sb.WriteString(fmt.Sprintf("新状态: %s\n", b.formatTicketStatus(n.Status)))
	if n.OldStatus != "" {
		sb.WriteString(fmt.Sprintf("原状态: %s\n", b.formatTicketStatus(n.OldStatus)))
	}
	return sb.String()
}

func (b *BotEnhanced) formatTicketRepliedMessage(n *TicketNotification) string {
	var sb strings.Builder
	sb.WriteString("*工单新回复*\n\n")
	sb.WriteString(fmt.Sprintf("工单号: `%s`\n", n.TicketNo))
	sb.WriteString(fmt.Sprintf("标题: %s\n", n.Title))
	sb.WriteString(fmt.Sprintf("回复者: %s\n", n.RepliedByName))
	if n.ReplyContent != "" && len(n.ReplyContent) > 100 {
		sb.WriteString(fmt.Sprintf("\n回复内容:\n%s...\n", n.ReplyContent[:100]))
	} else if n.ReplyContent != "" {
		sb.WriteString(fmt.Sprintf("\n回复内容:\n%s\n", n.ReplyContent))
	}
	return sb.String()
}

func (b *BotEnhanced) formatTicketResolvedMessage(n *TicketNotification) string {
	var sb strings.Builder
	sb.WriteString("*工单已解决*\n\n")
	sb.WriteString(fmt.Sprintf("工单号: `%s`\n", n.TicketNo))
	sb.WriteString(fmt.Sprintf("标题: %s\n", n.Title))
	if n.Resolution != "" {
		sb.WriteString(fmt.Sprintf("\n解决方案:\n%s\n", n.Resolution))
	}
	return sb.String()
}

func (b *BotEnhanced) formatTicketClosedMessage(n *TicketNotification) string {
	var sb strings.Builder
	sb.WriteString("*工单已关闭*\n\n")
	sb.WriteString(fmt.Sprintf("工单号: `%s`\n", n.TicketNo))
	sb.WriteString(fmt.Sprintf("标题: %s\n", n.Title))
	if n.ClosedReason != "" {
		sb.WriteString(fmt.Sprintf("关闭原因: %s\n", n.ClosedReason))
	}
	return sb.String()
}

func (b *BotEnhanced) formatGenericTicketMessage(n *TicketNotification) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*工单通知 - %s*\n\n", n.Type))
	sb.WriteString(fmt.Sprintf("工单号: `%s`\n", n.TicketNo))
	sb.WriteString(fmt.Sprintf("标题: %s\n", n.Title))
	sb.WriteString(fmt.Sprintf("状态: %s\n", b.formatTicketStatus(n.Status)))
	return sb.String()
}

func (b *BotEnhanced) formatUserTicketMessage(n *TicketNotification) string {
	var sb strings.Builder
	
	switch n.Type {
	case TicketNotificationTypeCreated:
		sb.WriteString("*您的工单已创建*\n\n")
		sb.WriteString(fmt.Sprintf("工单号: `%s`\n", n.TicketNo))
		sb.WriteString("我们已收到您的请求，将尽快处理。")
	case TicketNotificationTypeReplied:
		sb.WriteString("*您的工单有新回复*\n\n")
		sb.WriteString(fmt.Sprintf("工单号: `%s`\n", n.TicketNo))
		sb.WriteString(fmt.Sprintf("标题: %s\n", n.Title))
		sb.WriteString("\n请查看工单了解详情。")
	case TicketNotificationTypeResolved:
		sb.WriteString("*您的工单已解决*\n\n")
		sb.WriteString(fmt.Sprintf("工单号: `%s`\n", n.TicketNo))
		sb.WriteString(fmt.Sprintf("标题: %s\n", n.Title))
		sb.WriteString("\n如果问题仍未解决，请重新打开工单。")
	case TicketNotificationTypeClosed:
		sb.WriteString("*您的工单已关闭*\n\n")
		sb.WriteString(fmt.Sprintf("工单号: `%s`\n", n.TicketNo))
		sb.WriteString("感谢您的反馈。")
	default:
		sb.WriteString("*工单状态更新*\n\n")
		sb.WriteString(fmt.Sprintf("工单号: `%s`\n", n.TicketNo))
		sb.WriteString(fmt.Sprintf("当前状态: %s\n", b.formatTicketStatus(n.Status)))
	}
	
	return sb.String()
}

// Keyboard creation methods

func (b *BotEnhanced) createTicketActionKeyboard(ticketID uint, ticketNo string) *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("查看详情", fmt.Sprintf("ticket_view:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("快速回复", fmt.Sprintf("ticket_reply:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("分配工单", fmt.Sprintf("ticket_assign:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("更改状态", fmt.Sprintf("ticket_status:%d", ticketID)),
		),
	)
	return &keyboard
}

func (b *BotEnhanced) createTicketViewKeyboard(ticketID uint, ticketNo string) *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("查看详情", fmt.Sprintf("ticket_view:%d", ticketID)),
		),
	)
	return &keyboard
}

func (b *BotEnhanced) createUserTicketKeyboard(ticketID uint, ticketNo string) *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("查看工单", fmt.Sprintf("my_ticket:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("添加回复", fmt.Sprintf("add_reply:%d", ticketID)),
		),
	)
	return &keyboard
}

// Helper methods

func (b *BotEnhanced) formatPriority(priority string) string {
	switch priority {
	case "low":
		return "低"
	case "normal":
		return "普通"
	case "high":
		return "高"
	case "urgent":
		return "紧急"
	case "critical":
		return "严重"
	default:
		return priority
	}
}

func (b *BotEnhanced) formatTicketStatus(status string) string {
	switch status {
	case "open":
		return "待处理"
	case "in_progress":
		return "处理中"
	case "pending":
		return "待定"
	case "resolved":
		return "已解决"
	case "closed":
		return "已关闭"
	default:
		return status
	}
}

func (b *BotEnhanced) getAdminChatIDs() []int64 {
	// Get admin chat IDs from configuration
	// This can be configured in environment variables or config file
	adminIDsStr := b.cfg.Telegram.AdminChatIDs
	if adminIDsStr == "" {
		return []int64{}
	}
	
	var adminIDs []int64
	for _, idStr := range strings.Split(adminIDsStr, ",") {
		idStr = strings.TrimSpace(idStr)
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			adminIDs = append(adminIDs, id)
		}
	}
	
	return adminIDs
}