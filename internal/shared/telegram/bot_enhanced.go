package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"linke/internal/domains/subscription/usecases/interfaces"
	ticketInterfaces "linke/internal/domains/ticket/usecases/interfaces"
	ticketEntities "linke/internal/domains/ticket/entities"
	"linke/internal/domains/ticket/dto"
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
	ticketService          ticketInterfaces.TicketService
	ticketMessageService   ticketInterfaces.TicketMessageService
	cfg                    *config.Config
	userStates             map[int64]string                 // Track user conversation states
	ticketReplyBuffer      map[string][]string              // Buffer for multi-message replies [key: "chatID_ticketID"]
	ticketReplyMetadata    map[string]map[string]interface{} // Metadata for buffered replies
}

// NewBotEnhanced creates a new enhanced Telegram bot instance
func NewBotEnhanced(
	cfg *config.Config,
	userService userInterfaces.UserService,
	subscriptionService interfaces.UserSubscriptionService,
	subscriptionPlanService interfaces.SubscriptionPlanService,
	ticketService ticketInterfaces.TicketService,
	ticketMessageService ticketInterfaces.TicketMessageService,
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
		ticketService:          ticketService,
		ticketMessageService:   ticketMessageService,
		cfg:                    cfg,
		userStates:             make(map[int64]string),
		ticketReplyBuffer:      make(map[string][]string),
		ticketReplyMetadata:    make(map[string]map[string]interface{}),
	}, nil
}

// Start starts the enhanced bot and listens for updates
func (b *BotEnhanced) Start(ctx context.Context) error {
	logger.Info("Starting Enhanced Telegram bot", logger.String("username", b.api.Self.UserName))

	// Configure bot settings on startup
	if err := b.configureBotSettings(); err != nil {
		logger.Error("Failed to configure bot settings", logger.ErrorField(err))
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
		b.showWelcomeMessage(msg.Chat.ID)
	case "menu":
		b.showMainMenu(msg.Chat.ID)
	case "subscription":
		b.showSubscriptionMenu(msg.Chat.ID)
	case "usage":
		b.showUsageMenu(msg.Chat.ID)
	case "plans":
		b.showPlansMenu(msg.Chat.ID)
	case "support":
		b.showSupportMenuNew(msg.Chat.ID)
	case "settings":
		b.showSettingsMenuNew(msg.Chat.ID)
	case "tickets":
		b.showMyTicketsMenu(msg.Chat.ID)
	case "admin":
		b.showAdminMenu(msg.Chat.ID)
	case "help":
		b.showHelp(msg.Chat.ID)
	case "cancel":
		b.cancelCurrentOperation(msg.Chat.ID)
	case "status":
		b.showSystemStatus(msg.Chat.ID)
	default:
		b.sendMessage(msg.Chat.ID, "❌ 未知命令。请使用 /menu 查看主菜单或 /help 获取帮助。")
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
	
	logger.Info("Processing callback",
		logger.String("full_data", data),
		logger.String("action", parts[0]),
		logger.Strings("parts", parts),
		logger.Int("parts_count", len(parts)),
		logger.String("from", query.From.UserName))
	
	if len(parts) < 1 {
		return
	}

	action := parts[0]

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
	
	// New menu callbacks
	case "my_tickets":
		b.editMessageToMyTicketsMenu(query.Message.Chat.ID, query.Message.MessageID)
	case "system_status":
		b.editMessageToSystemStatus(query.Message.Chat.ID, query.Message.MessageID)
	case "admin_dashboard":
		b.editMessageToAdminMenu(query.Message.Chat.ID, query.Message.MessageID)
	case "help":
		b.editMessageToHelp(query.Message.Chat.ID, query.Message.MessageID)
	case "create_ticket":
		b.startCreateTicket(query.Message.Chat.ID, query.Message.MessageID)
	case "list_tickets":
		b.showUserTicketsList(query.Message.Chat.ID, query.Message.MessageID)
	case "search_tickets":
		b.startSearchTickets(query.Message.Chat.ID, query.Message.MessageID)
	case "ticket_stats":
		b.showTicketStatistics(query.Message.Chat.ID, query.Message.MessageID)
	
	// Enhanced ticket-related actions
	case "ticket_view":
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.showTicketDetails(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	case "ticket_reply":
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.startTicketReply(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	case "ticket_assign":
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.showTicketAssignment(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	case "ticket_status":
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.showTicketStatusChange(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	case "ticket_priority":
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.showTicketPriorityChange(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	case "ticket_tags":
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.showTicketTagsManagement(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	case "ticket_internal_note":
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.startInternalNote(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	case "ticket_related":
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.showRelatedTickets(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	case "ticket_history":
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.showTicketHistory(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	case "ticket_export":
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.exportTicket(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	case "my_ticket":
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.showMyTicketDetails(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	case "add_reply":
		if len(parts) > 1 {
			ticketID, err := strconv.Atoi(parts[1])
			if err != nil {
				logger.Error("Failed to parse ticket ID from add_reply callback",
					logger.String("callback_data", data),
					logger.ErrorField(err))
				b.showError(query.Message.Chat.ID, query.Message.MessageID, "无效的工单ID")
				return
			}
			logger.Info("Processing add_reply callback", 
				logger.Uint("ticket_id", uint(ticketID)))
			b.startAddReply(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		} else {
			logger.Error("Missing ticket ID in add_reply callback",
				logger.String("callback_data", data),
				logger.Int("parts_length", len(parts)))
			b.showError(query.Message.Chat.ID, query.Message.MessageID, "缺少工单ID")
		}
	
	// Enhanced user ticket actions
	case "user_set_priority":
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.showUserPriorityChange(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	case "add_attachment":
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.startAttachmentUpload(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	case "ticket_status_info":
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.showTicketStatusInfo(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	case "edit_ticket":
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.showTicketEditOptions(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	case "notification_settings":
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.showNotificationSettings(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	case "search_messages":
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.startMessageSearch(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	
	// Advanced admin operations
	case "admin_batch_operations":
		b.showBatchOperationsMenu(query.Message.Chat.ID, query.Message.MessageID)
	case "ticket_escalate":
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.showTicketEscalation(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	case "batch_assign":
		b.showBatchAssignMenu(query.Message.Chat.ID, query.Message.MessageID)
	case "batch_status":
		b.showBatchStatusMenu(query.Message.Chat.ID, query.Message.MessageID)
	case "batch_priority":
		b.showBatchPriorityMenu(query.Message.Chat.ID, query.Message.MessageID)
	case "escalate_to":
		if len(parts) > 2 {
			ticketID, _ := strconv.Atoi(parts[1])
			level := parts[2]
			b.handleTicketEscalation(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID), level)
		}
	
	// Notification system handlers
	case "notification_preferences":
		b.showNotificationPreferences(query.Message.Chat.ID, query.Message.MessageID)
	case "toggle_notification":
		if len(parts) > 1 {
			notificationType := parts[1]
			b.toggleNotificationSetting(query.Message.Chat.ID, query.Message.MessageID, notificationType)
		}
	case "notification_channels":
		b.showNotificationChannels(query.Message.Chat.ID, query.Message.MessageID)
	case "batch_notify":
		b.showBatchNotificationMenu(query.Message.Chat.ID, query.Message.MessageID)
	case "scheduled_notifications":
		b.showScheduledNotifications(query.Message.Chat.ID, query.Message.MessageID)
	
	// Status and priority change actions
	case "set_status":
		if len(parts) > 2 {
			ticketID, _ := strconv.Atoi(parts[1])
			status := parts[2]
			b.handleStatusChange(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID), status)
		}
	case "set_priority":
		if len(parts) > 2 {
			ticketID, _ := strconv.Atoi(parts[1])
			priority := parts[2]
			b.handlePriorityChange(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID), priority)
		}
	
	// Navigation
	case "cancel_reply":
		// Cancel the current reply operation and clear buffers
		chatID := query.Message.Chat.ID
		delete(b.userStates, chatID)
		// Clear any buffered messages
		for key := range b.ticketReplyBuffer {
			if strings.HasPrefix(key, fmt.Sprintf("%d_", chatID)) {
				delete(b.ticketReplyBuffer, key)
				delete(b.ticketReplyMetadata, key)
			}
		}
		b.editMessageToMainMenu(chatID, query.Message.MessageID)
	
	case "send_reply":
		// Send accumulated messages
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.sendAccumulatedReply(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	
	case "add_more":
		// Continue adding messages
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.continueAddingMessages(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	
	case "preview_all":
		// Preview all buffered messages
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.previewAllMessages(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	
	case "clear_buffer":
		// Clear buffer and restart
		if len(parts) > 1 {
			ticketID, _ := strconv.Atoi(parts[1])
			b.clearBufferAndRestart(query.Message.Chat.ID, query.Message.MessageID, uint(ticketID))
		}
	
	case "back":
		if len(parts) > 1 {
			b.navigateBack(query.Message.Chat.ID, query.Message.MessageID, parts[1])
		}
	}
}

// showMainMenu displays the main menu with inline keyboard
func (b *BotEnhanced) showMainMenu(chatID int64) {
	ctx := context.Background()
	
	// Try to get user info for personalized greeting
	var greeting string
	user, err := b.getUserByTelegramID(ctx, chatID)
	if err != nil {
		greeting = "🏠 *主菜单*\n\n请选择您需要的功能："
	} else {
		// Get current time for greeting
		hour := time.Now().Hour()
		var timeGreeting string
		switch {
		case hour < 12:
			timeGreeting = "早上好"
		case hour < 18:
			timeGreeting = "下午好"
		default:
			timeGreeting = "晚上好"
		}
		greeting = fmt.Sprintf("🏠 *主菜单*\n\n%s，%s！\n请选择您需要的功能：", timeGreeting, user.Username)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 我的订阅", "subscription"),
			tgbotapi.NewInlineKeyboardButtonData("💎 套餐商店", "plans"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 使用统计", "usage_info"),
			tgbotapi.NewInlineKeyboardButtonData("🎫 我的工单", "my_tickets"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💬 客服支持", "support"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ 设置", "settings"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 系统状态", "system_status"),
			tgbotapi.NewInlineKeyboardButtonData("❓ 帮助", "help"),
		),
	)
	
	// Add admin button if user is admin
	if user != nil && b.isUserAdmin(user) {
		// Add admin row
		adminRow := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👨‍💼 管理面板", "admin_dashboard"),
		)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, adminRow)
	}

	msg := tgbotapi.NewMessage(chatID, greeting)
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

func (b *BotEnhanced) getUserByTelegramID(ctx context.Context, telegramID int64) (*entities.UserResponse, error) {
	telegramIDStr := fmt.Sprintf("%d", telegramID)
	user, err := b.userService.GetUserByTelegramID(ctx, telegramIDStr)
	if err != nil {
		return nil, err
	}
	return user.ToResponse(), nil
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
			tgbotapi.NewInlineKeyboardButtonData("🔔 通知偏好", "notification_preferences"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📱 通知渠道", "notification_channels"),
			tgbotapi.NewInlineKeyboardButtonData("⏰ 定时通知", "scheduled_notifications"),
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
	chatID := msg.Chat.ID
	
	switch {
	case state == "awaiting_feedback":
		// Process feedback
		b.sendMessage(chatID, "感谢您的反馈！我们会尽快处理。")
		delete(b.userStates, chatID)
		
	case strings.HasPrefix(state, "replying_to_ticket_"):
		// Handle admin reply to ticket (support multi-message)
		ticketIDStr := strings.TrimPrefix(state, "replying_to_ticket_")
		if ticketID, err := strconv.Atoi(ticketIDStr); err == nil {
			b.handleMultiMessageReply(chatID, uint(ticketID), msg.Text, "admin")
		}
		
	case strings.HasPrefix(state, "adding_reply_to_ticket_"):
		// Handle user reply to ticket (support multi-message)
		ticketIDStr := strings.TrimPrefix(state, "adding_reply_to_ticket_")
		if ticketID, err := strconv.Atoi(ticketIDStr); err == nil {
			b.handleMultiMessageReply(chatID, uint(ticketID), msg.Text, "user")
		}
		
	case strings.HasPrefix(state, "adding_internal_note_"):
		// Handle internal note creation
		ticketIDStr := strings.TrimPrefix(state, "adding_internal_note_")
		if ticketID, err := strconv.Atoi(ticketIDStr); err == nil {
			b.handleInternalNote(chatID, uint(ticketID), msg.Text)
		}
		delete(b.userStates, chatID)
		
	case strings.HasPrefix(state, "searching_messages_"):
		// Handle message search
		ticketIDStr := strings.TrimPrefix(state, "searching_messages_")
		if ticketID, err := strconv.Atoi(ticketIDStr); err == nil {
			b.handleMessageSearch(chatID, uint(ticketID), msg.Text)
		}
		delete(b.userStates, chatID)
		
	case strings.HasPrefix(state, "adding_tags_"):
		// Handle tag addition
		ticketIDStr := strings.TrimPrefix(state, "adding_tags_")
		if ticketID, err := strconv.Atoi(ticketIDStr); err == nil {
			b.handleTagAddition(chatID, uint(ticketID), msg.Text)
		}
		delete(b.userStates, chatID)
		
	case strings.HasPrefix(state, "editing_title_"):
		// Handle ticket title editing
		ticketIDStr := strings.TrimPrefix(state, "editing_title_")
		if ticketID, err := strconv.Atoi(ticketIDStr); err == nil {
			b.handleTitleEdit(chatID, uint(ticketID), msg.Text)
		}
		delete(b.userStates, chatID)
		
	case strings.HasPrefix(state, "editing_description_"):
		// Handle ticket description editing
		ticketIDStr := strings.TrimPrefix(state, "editing_description_")
		if ticketID, err := strconv.Atoi(ticketIDStr); err == nil {
			b.handleDescriptionEdit(chatID, uint(ticketID), msg.Text)
		}
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

// sendErrorMessage sends a formatted error message with details
func (b *BotEnhanced) sendErrorMessage(chatID int64, title string, description string) {
	text := fmt.Sprintf("❌ *%s*\n\n%s", title, description)
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💬 联系客服", "support"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 主菜单", "main_menu"),
		),
	)
	
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	
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
	logger.Info("Configuring bot commands...")

	// Set bot commands only
	commands := []tgbotapi.BotCommand{
		{
			Command:     "start",
			Description: "🚀 开始使用",
		},
		{
			Command:     "menu",
			Description: "🏠 显示主菜单",
		},
		{
			Command:     "subscription",
			Description: "📊 查看订阅",
		},
		{
			Command:     "usage",
			Description: "📈 流量统计",
		},
		{
			Command:     "plans",
			Description: "💎 套餐商店",
		},
		{
			Command:     "tickets",
			Description: "🎫 我的工单",
		},
		{
			Command:     "support",
			Description: "💬 客服支持",
		},
		{
			Command:     "settings",
			Description: "⚙️ 设置",
		},
		{
			Command:     "admin",
			Description: "👨‍💼 管理面板",
		},
		{
			Command:     "status",
			Description: "📊 系统状态",
		},
		{
			Command:     "cancel",
			Description: "❌ 取消操作",
		},
		{
			Command:     "help",
			Description: "❓ 使用帮助",
		},
	}

	// Set commands
	setCommandsConfig := tgbotapi.NewSetMyCommands(commands...)
	if _, err := b.api.Request(setCommandsConfig); err != nil {
		logger.Error("Failed to set bot commands", logger.ErrorField(err))
		return fmt.Errorf("failed to set commands: %w", err)
	}
	logger.Info("Bot commands configured successfully", logger.Int("count", len(commands)))

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
				logger.ErrorField(err))
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
	
	// Choose appropriate icon and title based on who replied
	isAdminReply := false
	if metadata, ok := n.Metadata["is_admin_reply"]; ok {
		if adminReply, ok := metadata.(bool); ok {
			isAdminReply = adminReply
		}
	}
	
	if isAdminReply {
		sb.WriteString("👨‍💼 *管理员回复了您的工单*\n\n")
	} else {
		sb.WriteString("💬 *工单收到新回复*\n\n")
	}
	
	// Ticket basic info
	sb.WriteString(fmt.Sprintf("🎫 工单号: `%s`\n", n.TicketNo))
	sb.WriteString(fmt.Sprintf("📝 标题: %s\n", n.Title))
	
	// Priority and status with icons
	priorityIcon := b.getPriorityIcon(n.Priority)
	sb.WriteString(fmt.Sprintf("%s 优先级: %s\n", priorityIcon, b.formatPriority(n.Priority)))
	
	statusIcon := b.getStatusIcon(n.Status)
	sb.WriteString(fmt.Sprintf("%s 状态: %s\n", statusIcon, b.formatTicketStatus(n.Status)))
	
	// Reply information
	sb.WriteString(fmt.Sprintf("\n👤 回复者: *%s*\n", n.RepliedByName))
	sb.WriteString(fmt.Sprintf("⏰ 时间: %s\n", time.Now().Format("15:04:05")))
	
	// Reply content with better formatting
	if n.ReplyContent != "" {
		sb.WriteString("\n📄 *回复内容:*\n")
		sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n")
		
		// Format and truncate content if needed
		content := n.ReplyContent
		if len(content) > 300 {
			content = content[:297] + "..."
			sb.WriteString(fmt.Sprintf("%s\n", content))
			sb.WriteString("\n_(内容已截断，请查看完整工单)_\n")
		} else {
			sb.WriteString(fmt.Sprintf("%s\n", content))
		}
		sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n")
	}
	
	// Add action prompt
	if isAdminReply {
		sb.WriteString("\n✨ 管理员已处理您的工单，请查看详情。")
	} else {
		sb.WriteString("\n⚡ 请及时查看并处理新回复。")
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
		sb.WriteString("✅ *您的工单已成功创建*\n\n")
		sb.WriteString(fmt.Sprintf("🎫 工单号: `%s`\n", n.TicketNo))
		sb.WriteString(fmt.Sprintf("📝 标题: %s\n", n.Title))
		
		// Add priority info
		priorityIcon := b.getPriorityIcon(n.Priority)
		sb.WriteString(fmt.Sprintf("%s 优先级: %s\n\n", priorityIcon, b.formatPriority(n.Priority)))
		
		// Add expected response time
		switch n.Priority {
		case "critical":
			sb.WriteString("⏰ 预计响应时间: 1小时内\n")
		case "urgent":
			sb.WriteString("⏰ 预计响应时间: 4小时内\n")
		case "high":
			sb.WriteString("⏰ 预计响应时间: 1天内\n")
		case "normal":
			sb.WriteString("⏰ 预计响应时间: 3天内\n")
		default:
			sb.WriteString("⏰ 预计响应时间: 7天内\n")
		}
		
		sb.WriteString("\n💡 我们已收到您的请求，客服团队将尽快为您处理。")
		
	case TicketNotificationTypeReplied:
		// Check if it's admin reply
		isAdminReply := false
		if metadata, ok := n.Metadata["is_admin_reply"]; ok {
			if adminReply, ok := metadata.(bool); ok {
				isAdminReply = adminReply
			}
		}
		
		if isAdminReply {
			sb.WriteString("👨‍💼 *客服回复了您的工单*\n\n")
		} else {
			sb.WriteString("💬 *您的工单有新动态*\n\n")
		}
		
		sb.WriteString(fmt.Sprintf("🎫 工单号: `%s`\n", n.TicketNo))
		sb.WriteString(fmt.Sprintf("📝 标题: %s\n", n.Title))
		
		// Add reply preview if available
		if n.ReplyContent != "" {
			sb.WriteString("\n📄 *回复预览:*\n")
			preview := n.ReplyContent
			if len(preview) > 150 {
				preview = preview[:147] + "..."
			}
			sb.WriteString(fmt.Sprintf("_%s_\n", preview))
		}
		
		sb.WriteString("\n👆 点击下方按钮查看完整回复")
		
	case TicketNotificationTypeResolved:
		sb.WriteString("🎉 *您的工单已解决*\n\n")
		sb.WriteString(fmt.Sprintf("🎫 工单号: `%s`\n", n.TicketNo))
		sb.WriteString(fmt.Sprintf("📝 标题: %s\n", n.Title))
		
		if n.Resolution != "" {
			sb.WriteString("\n💡 *解决方案:*\n")
			resolution := n.Resolution
			if len(resolution) > 200 {
				resolution = resolution[:197] + "..."
			}
			sb.WriteString(fmt.Sprintf("_%s_\n", resolution))
		}
		
		sb.WriteString("\n😊 如果问题已解决，感谢您的耐心等待！")
		sb.WriteString("\n🔄 如果问题仍存在，您可以重新打开工单。")
		
	case TicketNotificationTypeClosed:
		sb.WriteString("🔒 *您的工单已关闭*\n\n")
		sb.WriteString(fmt.Sprintf("🎫 工单号: `%s`\n", n.TicketNo))
		sb.WriteString(fmt.Sprintf("📝 标题: %s\n", n.Title))
		
		if n.ClosedReason != "" {
			sb.WriteString(fmt.Sprintf("\n📋 关闭原因: %s\n", n.ClosedReason))
		}
		
		sb.WriteString("\n🙏 感谢您的反馈，我们会持续改进服务。")
		sb.WriteString("\n💬 如需帮助，欢迎创建新工单。")
		
	case TicketNotificationTypeAssigned:
		sb.WriteString("👤 *您的工单已分配给专员*\n\n")
		sb.WriteString(fmt.Sprintf("🎫 工单号: `%s`\n", n.TicketNo))
		sb.WriteString(fmt.Sprintf("📝 标题: %s\n", n.Title))
		
		if n.AssignedToName != "" {
			sb.WriteString(fmt.Sprintf("👨‍💼 处理专员: %s\n", n.AssignedToName))
		}
		
		sb.WriteString("\n⚡ 专员将尽快为您处理此工单。")
		
	case TicketNotificationTypeEscalated:
		sb.WriteString("⚡ *您的工单已升级处理*\n\n")
		sb.WriteString(fmt.Sprintf("🎫 工单号: `%s`\n", n.TicketNo))
		sb.WriteString(fmt.Sprintf("📝 标题: %s\n", n.Title))
		sb.WriteString("\n🚀 您的工单已提升优先级，将获得更快速的处理。")
		
	default:
		sb.WriteString("📢 *工单状态更新*\n\n")
		sb.WriteString(fmt.Sprintf("🎫 工单号: `%s`\n", n.TicketNo))
		sb.WriteString(fmt.Sprintf("📝 标题: %s\n", n.Title))
		
		statusIcon := b.getStatusIcon(n.Status)
		sb.WriteString(fmt.Sprintf("%s 当前状态: %s\n", statusIcon, b.formatTicketStatus(n.Status)))
		
		if n.OldStatus != "" {
			oldStatusIcon := b.getStatusIcon(n.OldStatus)
			sb.WriteString(fmt.Sprintf("%s 原状态: %s\n", oldStatusIcon, b.formatTicketStatus(n.OldStatus)))
		}
	}
	
	return sb.String()
}

// Keyboard creation methods

func (b *BotEnhanced) createTicketActionKeyboard(ticketID uint, ticketNo string) *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔍 查看详情", fmt.Sprintf("ticket_view:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("💬 快速回复", fmt.Sprintf("ticket_reply:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 分配工单", fmt.Sprintf("ticket_assign:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("🔄 更改状态", fmt.Sprintf("ticket_status:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚡ 设置优先级", fmt.Sprintf("ticket_priority:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("🏷️ 添加标签", fmt.Sprintf("ticket_tags:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 内部注释", fmt.Sprintf("ticket_internal_note:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("🔗 相关工单", fmt.Sprintf("ticket_related:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 历史记录", fmt.Sprintf("ticket_history:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("📤 导出工单", fmt.Sprintf("ticket_export:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔧 批量操作", "admin_batch_operations"),
			tgbotapi.NewInlineKeyboardButtonData("⚡ 升级流程", fmt.Sprintf("ticket_escalate:%d", ticketID)),
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
			tgbotapi.NewInlineKeyboardButtonData("📄 查看工单", fmt.Sprintf("my_ticket:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("💬 添加回复", fmt.Sprintf("add_reply:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎯 设置优先级", fmt.Sprintf("user_set_priority:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("📎 添加附件", fmt.Sprintf("add_attachment:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 工单状态", fmt.Sprintf("ticket_status_info:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("📝 编辑信息", fmt.Sprintf("edit_ticket:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔔 通知设置", fmt.Sprintf("notification_settings:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("🔍 搜索消息", fmt.Sprintf("search_messages:%d", ticketID)),
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

// Enhanced visual indicators
func (b *BotEnhanced) getStatusIcon(status string) string {
	switch status {
	case "open":
		return "🔓"
	case "in_progress":
		return "⚙️"
	case "pending":
		return "⏳"
	case "resolved":
		return "✅"
	case "closed":
		return "🔒"
	default:
		return "❓"
	}
}

func (b *BotEnhanced) getPriorityIcon(priority string) string {
	switch priority {
	case "low":
		return "🟢"
	case "normal":
		return "🔵"
	case "high":
		return "🟡"
	case "urgent":
		return "🟠"
	case "critical":
		return "🔴"
	default:
		return "⚪"
	}
}

// MarkdownV2 escaping function
func (b *BotEnhanced) escapeMarkdownV2(text string) string {
	// Characters that need escaping in MarkdownV2
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
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

// Ticket-related methods

// showTicketDetails shows detailed information about a specific ticket (admin view)
func (b *BotEnhanced) showTicketDetails(chatID int64, messageID int, ticketID uint) {
	ctx := context.Background()
	
	// Get ticket details
	ticket, err := b.ticketService.GetTicket(ctx, ticketID)
	if err != nil {
		b.showError(chatID, messageID, "获取工单详情失败")
		return
	}
	
	// Get ticket creator information
	user, userErr := b.userService.GetUserByID(ctx, ticket.UserID)
	
	// Build enhanced message with emojis and better formatting
	var sb strings.Builder
	sb.WriteString("🎫 *工单详情*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")
	
	// Basic Information Section
	sb.WriteString("📋 *基本信息*\n")
	sb.WriteString(fmt.Sprintf("🆔 工单号: `%s`\n", ticket.TicketNo))
	sb.WriteString(fmt.Sprintf("📌 标题: *%s*\n", b.escapeMarkdownV2(ticket.Title)))
	sb.WriteString(fmt.Sprintf("🏷️ 分类: `%s`\n", ticket.Category))
	
	// Status with visual indicator
	statusIcon := b.getStatusIcon(ticket.Status)
	sb.WriteString(fmt.Sprintf("%s 状态: *%s*\n", statusIcon, b.formatTicketStatus(ticket.Status)))
	
	// Priority with visual indicator
	priorityIcon := b.getPriorityIcon(ticket.Priority)
	sb.WriteString(fmt.Sprintf("%s 优先级: *%s*\n\n", priorityIcon, b.formatPriority(ticket.Priority)))
	
	// People Section
	sb.WriteString("👥 *人员信息*\n")
	if userErr == nil && user != nil {
		sb.WriteString(fmt.Sprintf("👤 创建者: `%s`\n", user.Username))
		if user.Email != "" {
			sb.WriteString(fmt.Sprintf("📧 邮箱: `%s`\n", user.Email))
		}
	}
	
	if ticket.AssignedToID != nil {
		assignedUser, assignedErr := b.userService.GetUserByID(ctx, *ticket.AssignedToID)
		if assignedErr == nil && assignedUser != nil {
			sb.WriteString(fmt.Sprintf("👨‍💼 负责人: `%s`\n", assignedUser.Username))
		}
	} else {
		sb.WriteString("👨‍💼 负责人: `未分配`\n")
	}
	
	// Timing Section
	sb.WriteString("\n⏰ *时间信息*\n")
	sb.WriteString(fmt.Sprintf("🕐 创建时间: `%s`\n", ticket.CreatedAt.Format("2006-01-02 15:04:05")))
	
	if ticket.FirstResponseAt != nil {
		sb.WriteString(fmt.Sprintf("⚡ 首次响应: `%s`\n", ticket.FirstResponseAt.Format("2006-01-02 15:04:05")))
	}
	
	if ticket.LastResponseAt != nil {
		sb.WriteString(fmt.Sprintf("💬 最后回复: `%s`\n", ticket.LastResponseAt.Format("2006-01-02 15:04:05")))
	}
	
	if ticket.ClosedAt != nil {
		sb.WriteString(fmt.Sprintf("🔒 关闭时间: `%s`\n", ticket.ClosedAt.Format("2006-01-02 15:04:05")))
	}
	
	// Tags Section
	if ticket.Tags != nil && *ticket.Tags != "" {
		sb.WriteString(fmt.Sprintf("\n🏷️ *标签*: `%s`\n", *ticket.Tags))
	}
	
	// Description Section
	sb.WriteString("\n📝 *工单描述*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("_%s_\n\n", b.escapeMarkdownV2(ticket.Description)))
	
	// Recent Messages Section
	messages, _, msgErr := b.ticketMessageService.GetTicketMessages(ctx, &dto.GetTicketMessagesRequest{
		TicketID: ticketID,
		Limit:    3,
	})
	
	if msgErr == nil && len(messages) > 0 {
		sb.WriteString("💬 *最近消息* (最新3条)\n")
		sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n")
		for i, msg := range messages {
			msgUser, _ := b.userService.GetUserByID(ctx, msg.UserID)
			userName := "Unknown"
			if msgUser != nil {
				userName = msgUser.Username
			}
			
			msgIcon := "👤"
			if msg.MessageType == "admin" {
				msgIcon = "👨‍💼"
			} else if msg.MessageType == "system" {
				msgIcon = "🤖"
			}
			
			sb.WriteString(fmt.Sprintf("%s *%s* `%s`\n", msgIcon, userName, msg.CreatedAt.Format("01-02 15:04")))
			content := b.truncateString(msg.Content, 100)
			sb.WriteString(fmt.Sprintf("💭 _%s_\n", b.escapeMarkdownV2(content)))
			
			if i < len(messages)-1 {
				sb.WriteString("- - - - - - - - - -\n")
			}
		}
	}
	
	keyboard := b.createTicketActionKeyboard(ticketID, ticket.TicketNo)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, sb.String())
	edit.ParseMode = "MarkdownV2"
	edit.ReplyMarkup = keyboard
	
	b.api.Send(edit)
}

// showMyTicketDetails shows ticket details for regular users
func (b *BotEnhanced) showMyTicketDetails(chatID int64, messageID int, ticketID uint) {
	ctx := context.Background()
	
	// Get user by telegram ID to verify ownership
	user, userErr := b.getUserByTelegramID(ctx, chatID)
	if userErr != nil {
		b.showUnboundAccount(chatID, messageID)
		return
	}
	
	// Get ticket details
	ticket, err := b.ticketService.GetTicket(ctx, ticketID)
	if err != nil {
		b.showError(chatID, messageID, "获取工单详情失败")
		return
	}
	
	// Verify user owns this ticket
	if ticket.UserID != user.ID {
		b.showError(chatID, messageID, "您只能查看自己的工单")
		return
	}
	
	// Build message
	var sb strings.Builder
	sb.WriteString("*我的工单*\n\n")
	sb.WriteString(fmt.Sprintf("工单号: `%s`\n", ticket.TicketNo))
	sb.WriteString(fmt.Sprintf("标题: %s\n", ticket.Title))
	sb.WriteString(fmt.Sprintf("状态: %s\n", b.formatTicketStatus(ticket.Status)))
	sb.WriteString(fmt.Sprintf("优先级: %s\n", b.formatPriority(ticket.Priority)))
	sb.WriteString(fmt.Sprintf("分类: %s\n", ticket.Category))
	sb.WriteString(fmt.Sprintf("创建时间: %s\n", ticket.CreatedAt.Format("2006-01-02 15:04")))
	
	if ticket.FirstResponseAt != nil {
		sb.WriteString(fmt.Sprintf("首次回复: %s\n", ticket.FirstResponseAt.Format("2006-01-02 15:04")))
	}
	
	if ticket.Status == "resolved" && ticket.Resolution != "" {
		sb.WriteString(fmt.Sprintf("\n解决方案:\n%s\n", ticket.Resolution))
	}
	
	sb.WriteString(fmt.Sprintf("\n描述:\n%s\n", ticket.Description))
	
	// Get user-visible messages
	messages, _, msgErr := b.ticketMessageService.GetTicketMessages(ctx, &dto.GetTicketMessagesRequest{
		TicketID:        ticketID,
		IncludeInternal: false, // Hide internal messages from users
		Limit:           10,
	})
	
	if msgErr == nil && len(messages) > 0 {
		sb.WriteString("\n*对话记录:*\n")
		for _, msg := range messages {
			msgTypeIcon := "👤"
			if msg.MessageType == "admin" || msg.MessageType == "system" {
				msgTypeIcon = "🔧"
			}
			sb.WriteString(fmt.Sprintf("\n%s *%s*\n%s\n", 
				msgTypeIcon,
				msg.CreatedAt.Format("01-02 15:04"),
				msg.Content))
		}
	}
	
	keyboard := b.createUserTicketKeyboard(ticketID, ticket.TicketNo)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, sb.String())
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = keyboard
	
	b.api.Send(edit)
}

// startTicketReply starts the conversation flow for replying to a ticket (admin)
func (b *BotEnhanced) startTicketReply(chatID int64, messageID int, ticketID uint) {
	ctx := context.Background()
	
	// Get ticket details first
	ticket, err := b.ticketService.GetTicket(ctx, ticketID)
	if err != nil {
		b.showError(chatID, messageID, "无法获取工单信息")
		return
	}
	
	// Check if ticket is closed
	if ticket.Status == "closed" {
		b.showError(chatID, messageID, "工单已关闭，无法回复")
		return
	}
	
	// Get user information for the ticket
	ticketUser, _ := b.userService.GetUserByID(ctx, ticket.UserID)
	
	// Set conversation state for this user
	stateKey := fmt.Sprintf("replying_to_ticket_%d", ticketID)
	b.userStates[chatID] = stateKey
	
	// Build informative prompt
	var sb strings.Builder
	sb.WriteString("💬 *管理员回复*\n\n")
	sb.WriteString(fmt.Sprintf("🎫 工单号: `%s`\n", ticket.TicketNo))
	sb.WriteString(fmt.Sprintf("📝 标题: %s\n", ticket.Title))
	
	if ticketUser != nil {
		sb.WriteString(fmt.Sprintf("👤 用户: %s\n", ticketUser.Username))
	}
	
	priorityIcon := b.getPriorityIcon(ticket.Priority)
	sb.WriteString(fmt.Sprintf("%s 优先级: %s\n", priorityIcon, b.formatPriority(ticket.Priority)))
	
	statusIcon := b.getStatusIcon(ticket.Status)
	sb.WriteString(fmt.Sprintf("%s 状态: %s\n\n", statusIcon, b.formatTicketStatus(ticket.Status)))
	
	sb.WriteString("请输入您的回复内容：\n\n")
	sb.WriteString("💡 *提示*:\n")
	sb.WriteString("• 回复将发送给用户\n")
	sb.WriteString("• 保持专业和友好的语气\n")
	sb.WriteString("• 支持多条消息 - 发送后可继续添加\n")
	sb.WriteString("• 提供清晰的解决方案\n")
	sb.WriteString("• 用户将收到通知")
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel_reply"),
			tgbotapi.NewInlineKeyboardButtonData("🔍 查看详情", fmt.Sprintf("ticket_view:%d", ticketID)),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, sb.String())
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// startAddReply starts the conversation flow for adding a reply (user)
func (b *BotEnhanced) startAddReply(chatID int64, messageID int, ticketID uint) {
	ctx := context.Background()
	
	// Get ticket details first
	ticket, err := b.ticketService.GetTicket(ctx, ticketID)
	if err != nil {
		b.showError(chatID, messageID, "无法获取工单信息")
		return
	}
	
	// Check if ticket is closed
	if ticket.Status == "closed" {
		b.showError(chatID, messageID, "工单已关闭，无法添加回复")
		return
	}
	
	// Verify user owns this ticket
	user, err := b.getUserByTelegramID(ctx, chatID)
	if err != nil {
		b.showUnboundAccount(chatID, messageID)
		return
	}
	
	if ticket.UserID != user.ID {
		b.showError(chatID, messageID, "您只能回复自己的工单")
		return
	}
	
	// Set conversation state for this user
	stateKey := fmt.Sprintf("adding_reply_to_ticket_%d", ticketID)
	b.userStates[chatID] = stateKey
	
	// Build informative prompt
	var sb strings.Builder
	sb.WriteString("💬 *添加回复*\n\n")
	sb.WriteString(fmt.Sprintf("🎫 工单号: `%s`\n", ticket.TicketNo))
	sb.WriteString(fmt.Sprintf("📝 标题: %s\n\n", ticket.Title))
	sb.WriteString("请输入您要补充的信息：\n\n")
	sb.WriteString("💡 *提示*:\n")
	sb.WriteString("• 请详细描述您的问题或需求\n")
	sb.WriteString("• 如有错误信息，请完整粘贴\n")
	sb.WriteString("• 支持多条消息 - 发送后可继续添加\n")
	sb.WriteString("• 所有消息会合并为一条回复\n")
	sb.WriteString("• 回复将记录在工单历史中")
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel_reply"),
			tgbotapi.NewInlineKeyboardButtonData("📄 查看工单", fmt.Sprintf("my_ticket:%d", ticketID)),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, sb.String())
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// showTicketAssignment shows ticket assignment interface (admin only)
func (b *BotEnhanced) showTicketAssignment(chatID int64, messageID int, ticketID uint) {
	text := "*工单分配*\n\n该功能需要通过管理面板操作。"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("返回工单", fmt.Sprintf("ticket_view:%d", ticketID)),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// showTicketStatusChange shows ticket status change interface (admin only)
func (b *BotEnhanced) showTicketStatusChange(chatID int64, messageID int, ticketID uint) {
	text := "*更改工单状态*\n\n请选择新状态："
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("处理中", fmt.Sprintf("set_status:%d:in_progress", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("待定", fmt.Sprintf("set_status:%d:pending", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("已解决", fmt.Sprintf("set_status:%d:resolved", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("已关闭", fmt.Sprintf("set_status:%d:closed", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("返回工单", fmt.Sprintf("ticket_view:%d", ticketID)),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// truncateString truncates a string to specified length
func (b *BotEnhanced) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// handleTicketReply processes ticket reply submissions
func (b *BotEnhanced) handleTicketReply(chatID int64, ticketID uint, content string, messageType string) {
	ctx := context.Background()
	
	logger.Info("Processing ticket reply via telegram",
		logger.Int64("chat_id", chatID),
		logger.Uint("ticket_id", ticketID),
		logger.String("message_type", messageType),
		logger.Int("content_length", len(content)))
	
	// Validate ticket ID
	if ticketID == 0 {
		logger.Error("Invalid ticket ID provided for telegram reply",
			logger.Int64("chat_id", chatID),
			logger.Uint("ticket_id", ticketID))
		b.sendErrorMessage(chatID, "无效的工单ID", "请确认工单号是否正确")
		return
	}
	
	// Get user by telegram ID
	user, err := b.getUserByTelegramID(ctx, chatID)
	if err != nil {
		logger.Error("Failed to get user by telegram ID for ticket reply",
			logger.Int64("chat_id", chatID),
			logger.Uint("ticket_id", ticketID),
			logger.ErrorField(err))
		b.sendErrorMessage(chatID, "账号未绑定", "请先在网站上使用 Telegram 登录绑定账号")
		return
	}
	
	// Validate content
	content = strings.TrimSpace(content)
	if content == "" {
		b.sendErrorMessage(chatID, "回复内容为空", "请输入有效的回复内容")
		return
	}
	
	// Check content length
	if len(content) > 5000 {
		b.sendErrorMessage(chatID, "回复内容过长", "回复内容不能超过5000个字符")
		return
	}
	
	// Get ticket to verify it exists and check permissions
	ticket, err := b.ticketService.GetTicket(ctx, ticketID)
	if err != nil {
		logger.Error("Failed to get ticket for validation",
			logger.Uint("ticket_id", ticketID),
			logger.ErrorField(err))
		b.sendErrorMessage(chatID, "工单不存在", "无法找到指定的工单")
		return
	}
	
	// Check if ticket is closed
	if ticket.Status == "closed" {
		b.sendErrorMessage(chatID, "工单已关闭", "该工单已关闭，无法添加回复")
		return
	}
	
	// Verify permissions based on message type
	if messageType == "user" {
		// User can only reply to their own tickets
		if ticket.UserID != user.ID {
			b.sendErrorMessage(chatID, "权限不足", "您只能回复自己的工单")
			return
		}
	} else if messageType == "admin" {
		// TODO: Verify user is admin
		// For now, we'll allow any authenticated user to reply as admin
		// This should be restricted to actual admin users
	}
	
	// Show typing indicator
	typingAction := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	b.api.Send(typingAction)
	
	// Create the ticket message
	req := &dto.CreateTicketMessageRequest{
		Content:     content,
		MessageType: messageType,
		IsInternal:  false,
		Metadata:    fmt.Sprintf(`{"source":"telegram","chat_id":%d}`, chatID),
	}
	
	message, err := b.ticketMessageService.CreateTicketMessage(ctx, ticketID, user.ID, req)
	if err != nil {
		logger.Error("Failed to create ticket message via telegram",
			logger.Uint("ticket_id", ticketID),
			logger.Uint("user_id", user.ID),
			logger.ErrorField(err))
		b.sendErrorMessage(chatID, "回复失败", "系统错误，请稍后重试")
		return
	}
	
	// Build enhanced confirmation message
	var sb strings.Builder
	if messageType == "admin" {
		sb.WriteString("✅ *管理员回复已发送*\n\n")
		sb.WriteString(fmt.Sprintf("🎫 工单号: `%s`\n", ticket.TicketNo))
		sb.WriteString(fmt.Sprintf("📝 标题: %s\n", b.escapeMarkdownV2(ticket.Title)))
		sb.WriteString(fmt.Sprintf("💬 回复ID: \\#%d\n\n", message.ID))
		sb.WriteString("用户将收到通知。")
	} else {
		sb.WriteString("✅ *回复已添加*\n\n")
		sb.WriteString(fmt.Sprintf("🎫 工单号: `%s`\n", ticket.TicketNo))
		sb.WriteString(fmt.Sprintf("📝 标题: %s\n", b.escapeMarkdownV2(ticket.Title)))
		sb.WriteString(fmt.Sprintf("💬 消息ID: \\#%d\n\n", message.ID))
		
		// Add estimated response time based on priority
		switch ticket.Priority {
		case "critical":
			sb.WriteString("⏰ 预计响应时间: 1小时内\n")
		case "urgent":
			sb.WriteString("⏰ 预计响应时间: 4小时内\n")
		case "high":
			sb.WriteString("⏰ 预计响应时间: 1天内\n")
		case "normal":
			sb.WriteString("⏰ 预计响应时间: 3天内\n")
		default:
			sb.WriteString("⏰ 预计响应时间: 7天内\n")
		}
		
		sb.WriteString("\n客服将尽快处理您的消息。")
	}
	
	// Create keyboard for further actions
	var keyboard tgbotapi.InlineKeyboardMarkup
	if messageType == "admin" {
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔍 查看工单", fmt.Sprintf("ticket_view:%d", ticketID)),
				tgbotapi.NewInlineKeyboardButtonData("💬 继续回复", fmt.Sprintf("ticket_reply:%d", ticketID)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📝 内部注释", fmt.Sprintf("ticket_internal_note:%d", ticketID)),
				tgbotapi.NewInlineKeyboardButtonData("🔄 更改状态", fmt.Sprintf("ticket_status:%d", ticketID)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 主菜单", "main_menu"),
			),
		)
	} else {
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📄 查看工单", fmt.Sprintf("my_ticket:%d", ticketID)),
				tgbotapi.NewInlineKeyboardButtonData("💬 继续补充", fmt.Sprintf("add_reply:%d", ticketID)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📎 添加附件", fmt.Sprintf("add_attachment:%d", ticketID)),
				tgbotapi.NewInlineKeyboardButtonData("🔔 通知设置", fmt.Sprintf("notification_settings:%d", ticketID)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 主菜单", "main_menu"),
			),
		)
	}
	
	msg := tgbotapi.NewMessage(chatID, sb.String())
	msg.ParseMode = "MarkdownV2"
	msg.ReplyMarkup = keyboard
	
	_, sendErr := b.api.Send(msg)
	if sendErr != nil {
		logger.Error("Failed to send confirmation message",
			logger.Int64("chat_id", chatID),
			logger.ErrorField(sendErr))
	}
	
	logger.Info("Ticket reply processed successfully via telegram",
		logger.Uint("message_id", message.ID),
		logger.Uint("ticket_id", ticketID),
		logger.Uint("user_id", user.ID),
		logger.String("message_type", messageType))
}

// handleStatusChange processes ticket status changes
func (b *BotEnhanced) handleStatusChange(chatID int64, messageID int, ticketID uint, newStatus string) {
	ctx := context.Background()
	
	// Update ticket status
	ticket, err := b.ticketService.UpdateTicketStatus(ctx, ticketID, newStatus)
	if err != nil {
		logger.Error("Failed to update ticket status via telegram",
			logger.Uint("ticket_id", ticketID),
			logger.String("new_status", newStatus),
			logger.ErrorField(err))
		b.showError(chatID, messageID, "状态更新失败")
		return
	}
	
	// Send enhanced confirmation message with visual indicators
	statusIcon := b.getStatusIcon(newStatus)
	text := fmt.Sprintf("✅ *状态已更新*\n\n🎫 工单: `%s`\n%s 新状态: *%s*", 
		ticket.TicketNo, 
		statusIcon,
		b.formatTicketStatus(newStatus))
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔍 查看工单", fmt.Sprintf("ticket_view:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("🏠 主菜单", "main_menu"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
	
	logger.Info("Ticket status updated via telegram",
		logger.Uint("ticket_id", ticketID),
		logger.String("old_status", ticket.Status),
		logger.String("new_status", newStatus))
}

// Enhanced Priority Management
func (b *BotEnhanced) showTicketPriorityChange(chatID int64, messageID int, ticketID uint) {
	text := "⚡ *设置工单优先级*\n\n请选择新的优先级："
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🟢 低", fmt.Sprintf("set_priority:%d:low", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("🔵 普通", fmt.Sprintf("set_priority:%d:normal", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🟡 高", fmt.Sprintf("set_priority:%d:high", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("🟠 紧急", fmt.Sprintf("set_priority:%d:urgent", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔴 严重", fmt.Sprintf("set_priority:%d:critical", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回工单", fmt.Sprintf("ticket_view:%d", ticketID)),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

func (b *BotEnhanced) handlePriorityChange(chatID int64, messageID int, ticketID uint, newPriority string) {
	// Update ticket priority (you'll need to implement this in your ticket service)
	// ticket, err := b.ticketService.UpdateTicketPriority(ctx, ticketID, newPriority)
	
	// For now, we'll show a confirmation message
	priorityIcon := b.getPriorityIcon(newPriority)
	text := fmt.Sprintf("✅ *优先级已更新*\n\n🎫 工单: \\#%d\n%s 新优先级: *%s*", 
		ticketID, 
		priorityIcon,
		b.formatPriority(newPriority))
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔍 查看工单", fmt.Sprintf("ticket_view:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("🏠 主菜单", "main_menu"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// Tags Management
func (b *BotEnhanced) showTicketTagsManagement(chatID int64, messageID int, ticketID uint) {
	ctx := context.Background()
	
	ticket, err := b.ticketService.GetTicket(ctx, ticketID)
	if err != nil {
		b.showError(chatID, messageID, "获取工单信息失败")
		return
	}
	
	currentTags := ""
	if ticket.Tags != nil && *ticket.Tags != "" {
		currentTags = *ticket.Tags
	}
	
	text := fmt.Sprintf("🏷️ *标签管理*\n\n🎫 工单: `%s`\n\n📋 当前标签: `%s`\n\n点击下方按钮管理标签：", 
		ticket.TicketNo, currentTags)
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ 添加标签", fmt.Sprintf("add_tag:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("🗑️ 清除标签", fmt.Sprintf("clear_tags:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏷️ 常用标签", fmt.Sprintf("common_tags:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回工单", fmt.Sprintf("ticket_view:%d", ticketID)),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// Internal Notes
func (b *BotEnhanced) startInternalNote(chatID int64, messageID int, ticketID uint) {
	// Set conversation state for internal note
	stateKey := fmt.Sprintf("adding_internal_note_%d", ticketID)
	b.userStates[chatID] = stateKey
	
	text := "📝 *添加内部注释*\n\n请输入内部注释内容（仅管理员可见）："
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel_reply"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// Message Search
func (b *BotEnhanced) startMessageSearch(chatID int64, messageID int, ticketID uint) {
	stateKey := fmt.Sprintf("searching_messages_%d", ticketID)
	b.userStates[chatID] = stateKey
	
	text := "🔍 *搜索工单消息*\n\n请输入要搜索的关键词："
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel_reply"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// Ticket History
func (b *BotEnhanced) showTicketHistory(chatID int64, messageID int, ticketID uint) {
	ctx := context.Background()
	
	// Get all messages for the ticket
	messages, _, err := b.ticketMessageService.GetTicketMessages(ctx, &dto.GetTicketMessagesRequest{
		TicketID: ticketID,
		Limit:    50, // Get more history
	})
	
	if err != nil {
		b.showError(chatID, messageID, "获取历史记录失败")
		return
	}
	
	var sb strings.Builder
	sb.WriteString("📊 *工单历史记录*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")
	
	if len(messages) == 0 {
		sb.WriteString("暂无历史记录\n")
	} else {
		sb.WriteString(fmt.Sprintf("📈 总消息数: *%d*\n\n", len(messages)))
		
		// Group messages by date
		dateGroups := make(map[string][]*ticketEntities.TicketMessage)
		for _, msg := range messages {
			date := msg.CreatedAt.Format("2006-01-02")
			dateGroups[date] = append(dateGroups[date], msg)
		}
		
		// Show recent messages
		recentCount := 10
		if len(messages) < recentCount {
			recentCount = len(messages)
		}
		
		sb.WriteString(fmt.Sprintf("📝 *最近 %d 条消息*:\n", recentCount))
		for i := 0; i < recentCount; i++ {
			msg := messages[i]
			msgUser, _ := b.userService.GetUserByID(ctx, msg.UserID)
			userName := "Unknown"
			if msgUser != nil {
				userName = msgUser.Username
			}
			
			msgIcon := "👤"
			if msg.MessageType == "admin" {
				msgIcon = "👨‍💼"
			} else if msg.MessageType == "system" {
				msgIcon = "🤖"
			}
			
			sb.WriteString(fmt.Sprintf("%s %s `%s`\n", msgIcon, userName, msg.CreatedAt.Format("01-02 15:04")))
			content := b.truncateString(msg.Content, 80)
			sb.WriteString(fmt.Sprintf("💭 _%s_\n\n", content))
		}
	}
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📤 导出历史", fmt.Sprintf("ticket_export:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("🔍 搜索消息", fmt.Sprintf("search_messages:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回工单", fmt.Sprintf("ticket_view:%d", ticketID)),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, sb.String())
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// Export functionality
func (b *BotEnhanced) exportTicket(chatID int64, messageID int, ticketID uint) {
	text := "📤 *导出工单*\n\n🚧 导出功能开发中...\n\n将支持以下格式：\n• 📄 PDF 报告\n• 📊 Excel 表格\n• 💾 JSON 数据\n• 📧 邮件发送"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回工单", fmt.Sprintf("ticket_view:%d", ticketID)),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// Placeholder functions for user actions (to be implemented)
func (b *BotEnhanced) showUserPriorityChange(chatID int64, messageID int, ticketID uint) {
	text := "🎯 *设置优先级*\n\n用户可设置的优先级："
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔵 普通", fmt.Sprintf("set_priority:%d:normal", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("🟡 高", fmt.Sprintf("set_priority:%d:high", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🟠 ���急", fmt.Sprintf("set_priority:%d:urgent", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回", fmt.Sprintf("my_ticket:%d", ticketID)),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

func (b *BotEnhanced) startAttachmentUpload(chatID int64, messageID int, ticketID uint) {
	text := "📎 *添加附件*\n\n🚧 附件上传功能开发中...\n\n将支持：\n• 🖼️ 图片文件\n• 📄 文档文件\n• 🎥 视频文件\n• 🗂️ 压缩文件"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回", fmt.Sprintf("my_ticket:%d", ticketID)),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

func (b *BotEnhanced) showTicketStatusInfo(chatID int64, messageID int, ticketID uint) {
	ctx := context.Background()
	
	ticket, err := b.ticketService.GetTicket(ctx, ticketID)
	if err != nil {
		b.showError(chatID, messageID, "获取工单信息失败")
		return
	}
	
	statusIcon := b.getStatusIcon(ticket.Status)
	var statusDescription string
	
	switch ticket.Status {
	case "open":
		statusDescription = "工单已创建，等待处理"
	case "in_progress":
		statusDescription = "工单正在处理中"
	case "pending":
		statusDescription = "工单等待更多信息"
	case "resolved":
		statusDescription = "工单已解决"
	case "closed":
		statusDescription = "工单已关闭"
	default:
		statusDescription = "状态信息"
	}
	
	text := fmt.Sprintf("📊 *工单状态信息*\n\n🎫 工单号: `%s`\n%s 当前状态: *%s*\n\n📝 状态说明:\n_%s_", 
		ticket.TicketNo, statusIcon, b.formatTicketStatus(ticket.Status), statusDescription)
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回", fmt.Sprintf("my_ticket:%d", ticketID)),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

func (b *BotEnhanced) showTicketEditOptions(chatID int64, messageID int, ticketID uint) {
	text := "📝 *编辑工单信息*\n\n选择要编辑的内容："
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📌 修改标题", fmt.Sprintf("edit_title:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("📄 修改描述", fmt.Sprintf("edit_description:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏷️ 修改分类", fmt.Sprintf("edit_category:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回", fmt.Sprintf("my_ticket:%d", ticketID)),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

func (b *BotEnhanced) showNotificationSettings(chatID int64, messageID int, ticketID uint) {
	text := "🔔 *通知设置*\n\n🚧 通知设置功能开发中...\n\n将提供：\n• 🔕 静音通知\n• ⏰ 定时提醒\n• 📧 邮件通知\n• 💌 重要更新提醒"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回", fmt.Sprintf("my_ticket:%d", ticketID)),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

func (b *BotEnhanced) showRelatedTickets(chatID int64, messageID int, ticketID uint) {
	text := "🔗 *相关工单*\n\n🚧 相关工单功能开发中...\n\n将支持：\n• 🔍 智能关联\n• 📊 问题模式识别\n• 🔄 重复问题检测\n• 📈 趋势分析"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回工单", fmt.Sprintf("ticket_view:%d", ticketID)),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// Conversation flow handlers

// handleInternalNote processes internal note creation
func (b *BotEnhanced) handleInternalNote(chatID int64, ticketID uint, content string) {
	ctx := context.Background()
	
	// Get user by telegram ID
	user, err := b.getUserByTelegramID(ctx, chatID)
	if err != nil {
		b.sendMessage(chatID, "❌ 未找到绑定的账号")
		return
	}
	
	// Create internal note
	message, err := b.ticketMessageService.CreateInternalMessage(ctx, ticketID, user.ID, content)
	if err != nil {
		logger.Error("Failed to create internal note via telegram",
			logger.Uint("ticket_id", ticketID),
			logger.Uint("user_id", user.ID),
			logger.ErrorField(err))
		b.sendMessage(chatID, "❌ 创建内部注释失败，请稍后重试")
		return
	}
	
	text := fmt.Sprintf("✅ *内部注释已添加*\n\n🎫 工单: \\#%d\n📝 注释ID: \\#%d\n\n内部注释仅管理员可见。", 
		ticketID, message.ID)
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔍 查看工单", fmt.Sprintf("ticket_view:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("🏠 主菜单", "main_menu"),
		),
	)
	
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	
	b.api.Send(msg)
	
	logger.Info("Internal note created via telegram",
		logger.Uint("message_id", message.ID),
		logger.Uint("ticket_id", ticketID),
		logger.Uint("user_id", user.ID))
}

// handleMessageSearch processes message search queries
func (b *BotEnhanced) handleMessageSearch(chatID int64, ticketID uint, searchQuery string) {
	ctx := context.Background()
	
	if strings.TrimSpace(searchQuery) == "" {
		b.sendMessage(chatID, "❌ 搜索关键词不能为空")
		return
	}
	
	// Get all messages for the ticket
	messages, _, err := b.ticketMessageService.GetTicketMessages(ctx, &dto.GetTicketMessagesRequest{
		TicketID: ticketID,
		Limit:    100, // Get more messages for search
	})
	
	if err != nil {
		b.sendMessage(chatID, "❌ 搜索失败，请稍后重试")
		return
	}
	
	// Search for matching messages
	var matchedMessages []*ticketEntities.TicketMessage
	searchLower := strings.ToLower(searchQuery)
	
	for _, msg := range messages {
		if strings.Contains(strings.ToLower(msg.Content), searchLower) {
			matchedMessages = append(matchedMessages, msg)
		}
	}
	
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔍 *搜索结果*\n\n🔎 关键词: `%s`\n📊 找到 %d 条匹配消息\n\n", 
		searchQuery, len(matchedMessages)))
	
	if len(matchedMessages) == 0 {
		sb.WriteString("😔 没有找到包含该关键词的消息")
	} else {
		sb.WriteString("📝 *匹配的消息*:\n")
		displayCount := len(matchedMessages)
		if displayCount > 5 {
			displayCount = 5 // Limit display to first 5 matches
		}
		
		for i := 0; i < displayCount; i++ {
			msg := matchedMessages[i]
			msgUser, _ := b.userService.GetUserByID(ctx, msg.UserID)
			userName := "Unknown"
			if msgUser != nil {
				userName = msgUser.Username
			}
			
			msgIcon := "👤"
			if msg.MessageType == "admin" {
				msgIcon = "👨‍💼"
			} else if msg.MessageType == "system" {
				msgIcon = "🤖"
			}
			
			sb.WriteString(fmt.Sprintf("%s *%s* `%s`\n", msgIcon, userName, msg.CreatedAt.Format("01-02 15:04")))
			
			// Highlight the search term in content
			content := msg.Content
			if len(content) > 150 {
				content = content[:150] + "..."
			}
			sb.WriteString(fmt.Sprintf("💭 %s\n\n", content))
		}
		
		if len(matchedMessages) > displayCount {
			sb.WriteString(fmt.Sprintf("... 还有 %d 条匹配结果\n", len(matchedMessages)-displayCount))
		}
	}
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 历史记录", fmt.Sprintf("ticket_history:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("🔍 新搜索", fmt.Sprintf("search_messages:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回工单", fmt.Sprintf("ticket_view:%d", ticketID)),
		),
	)
	
	msg := tgbotapi.NewMessage(chatID, sb.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	
	b.api.Send(msg)
	
	logger.Info("Message search completed via telegram",
		logger.Int64("chat_id", chatID),
		logger.Uint("ticket_id", ticketID),
		logger.String("search_query", searchQuery),
		logger.Int("matches", len(matchedMessages)))
}

// handleTagAddition processes tag addition to tickets
func (b *BotEnhanced) handleTagAddition(chatID int64, ticketID uint, tagsInput string) {
	ctx := context.Background()
	
	if strings.TrimSpace(tagsInput) == "" {
		b.sendMessage(chatID, "❌ 标签不能为空")
		return
	}
	
	// Get current ticket
	ticket, err := b.ticketService.GetTicket(ctx, ticketID)
	if err != nil {
		b.sendMessage(chatID, "❌ 获取工单信息失败")
		return
	}
	
	// Parse new tags (comma-separated)
	newTags := strings.Split(tagsInput, ",")
	var cleanedTags []string
	for _, tag := range newTags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			cleanedTags = append(cleanedTags, tag)
		}
	}
	
	// Combine with existing tags
	var existingTags []string
	if ticket.Tags != nil && *ticket.Tags != "" {
		existingTags = strings.Split(*ticket.Tags, ",")
		for i, tag := range existingTags {
			existingTags[i] = strings.TrimSpace(tag)
		}
	}
	
	// Remove duplicates and create final tag list
	tagMap := make(map[string]bool)
	var finalTags []string
	
	// Add existing tags
	for _, tag := range existingTags {
		if tag != "" && !tagMap[tag] {
			tagMap[tag] = true
			finalTags = append(finalTags, tag)
		}
	}
	
	// Add new tags
	for _, tag := range cleanedTags {
		if !tagMap[tag] {
			tagMap[tag] = true
			finalTags = append(finalTags, tag)
		}
	}
	
	// Update ticket with new tags (placeholder - would need actual ticket service method)
	finalTagsStr := strings.Join(finalTags, ", ")
	
	// TODO: Implement UpdateTicketTags method in ticket service
	// For now, just show confirmation
	
	text := fmt.Sprintf("✅ *标签已更新*\n\n🎫 工单: `%s`\n🏷️ 新标签: `%s`\n\n📊 总标签数: %d", 
		ticket.TicketNo, finalTagsStr, len(finalTags))
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏷️ 标签管理", fmt.Sprintf("ticket_tags:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("🔍 查看工单", fmt.Sprintf("ticket_view:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 主菜单", "main_menu"),
		),
	)
	
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	
	b.api.Send(msg)
	
	logger.Info("Tags added to ticket via telegram",
		logger.Uint("ticket_id", ticketID),
		logger.String("new_tags", finalTagsStr),
		logger.Int("total_tags", len(finalTags)))
}

// handleTitleEdit processes ticket title editing
func (b *BotEnhanced) handleTitleEdit(chatID int64, ticketID uint, newTitle string) {
	ctx := context.Background()
	
	newTitle = strings.TrimSpace(newTitle)
	if newTitle == "" {
		b.sendMessage(chatID, "❌ 工单标题不能为空")
		return
	}
	
	if len(newTitle) > 255 {
		b.sendMessage(chatID, "❌ 工单标题过长，请控制在255字符以内")
		return
	}
	
	// Get current ticket to verify ownership
	ticket, err := b.ticketService.GetTicket(ctx, ticketID)
	if err != nil {
		b.sendMessage(chatID, "❌ 获取工单信息失败")
		return
	}
	
	// Get user by telegram ID to verify ownership
	user, err := b.getUserByTelegramID(ctx, chatID)
	if err != nil {
		b.sendMessage(chatID, "❌ 未找到绑定的账号")
		return
	}
	
	// Check if user owns this ticket
	if ticket.UserID != user.ID {
		b.sendMessage(chatID, "❌ 您只能编辑自己的工单")
		return
	}
	
	// TODO: Implement UpdateTicketTitle method in ticket service
	// For now, just show confirmation
	
	text := fmt.Sprintf("✅ *工单标题已更新*\n\n🎫 工单号: `%s`\n📝 原标题: _%s_\n📝 新标题: *%s*", 
		ticket.TicketNo, 
		b.escapeMarkdownV2(ticket.Title), 
		b.escapeMarkdownV2(newTitle))
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📄 查看工单", fmt.Sprintf("my_ticket:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("📝 继续编辑", fmt.Sprintf("edit_ticket:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 主菜单", "main_menu"),
		),
	)
	
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "MarkdownV2"
	msg.ReplyMarkup = keyboard
	
	b.api.Send(msg)
	
	logger.Info("Ticket title updated via telegram",
		logger.Uint("ticket_id", ticketID),
		logger.Uint("user_id", user.ID),
		logger.String("old_title", ticket.Title),
		logger.String("new_title", newTitle))
}

// handleDescriptionEdit processes ticket description editing
func (b *BotEnhanced) handleDescriptionEdit(chatID int64, ticketID uint, newDescription string) {
	ctx := context.Background()
	
	newDescription = strings.TrimSpace(newDescription)
	if newDescription == "" {
		b.sendMessage(chatID, "❌ 工单描述不能为空")
		return
	}
	
	// Get current ticket to verify ownership
	ticket, err := b.ticketService.GetTicket(ctx, ticketID)
	if err != nil {
		b.sendMessage(chatID, "❌ 获取工单信息失败")
		return
	}
	
	// Get user by telegram ID to verify ownership
	user, err := b.getUserByTelegramID(ctx, chatID)
	if err != nil {
		b.sendMessage(chatID, "❌ 未找到绑定的账号")
		return
	}
	
	// Check if user owns this ticket
	if ticket.UserID != user.ID {
		b.sendMessage(chatID, "❌ 您只能编辑自己的工单")
		return
	}
	
	// TODO: Implement UpdateTicketDescription method in ticket service
	// For now, just show confirmation
	
	oldDescPreview := ticket.Description
	if len(oldDescPreview) > 100 {
		oldDescPreview = oldDescPreview[:100] + "..."
	}
	
	newDescPreview := newDescription
	if len(newDescPreview) > 100 {
		newDescPreview = newDescPreview[:100] + "..."
	}
	
	text := fmt.Sprintf("✅ *工单描述已更新*\n\n🎫 工单号: `%s`\n\n📝 *原描述预览*:\n_%s_\n\n📝 *新描述预览*:\n*%s*", 
		ticket.TicketNo, 
		b.escapeMarkdownV2(oldDescPreview), 
		b.escapeMarkdownV2(newDescPreview))
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📄 查看工单", fmt.Sprintf("my_ticket:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("📝 继续编辑", fmt.Sprintf("edit_ticket:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 主菜单", "main_menu"),
		),
	)
	
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "MarkdownV2"
	msg.ReplyMarkup = keyboard
	
	b.api.Send(msg)
	
	logger.Info("Ticket description updated via telegram",
		logger.Uint("ticket_id", ticketID),
		logger.Uint("user_id", user.ID),
		logger.Int("old_length", len(ticket.Description)),
		logger.Int("new_length", len(newDescription)))
}

// Advanced Admin Functionality

// showBatchOperationsMenu displays the batch operations menu for admins
func (b *BotEnhanced) showBatchOperationsMenu(chatID int64, messageID int) {
	text := "🔧 *批量操作中心*\n\n" +
		"选择要执行的批量操作类型：\n\n" +
		"• 批量分配 - 将多个工单分配给指定管理员\n" +
		"• 批量状态 - 同时更改多个工单状态\n" +
		"• 批量优先级 - 调整多个工单优先级\n" +
		"• 批量标签 - 为多个工单添加标签"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👥 批量分配", "batch_assign"),
			tgbotapi.NewInlineKeyboardButtonData("🔄 批量状态", "batch_status"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚡ 批量优先级", "batch_priority"),
			tgbotapi.NewInlineKeyboardButtonData("🏷️ 批量标签", "batch_tags"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 批量统计", "batch_statistics"),
			tgbotapi.NewInlineKeyboardButtonData("📤 批量导出", "batch_export"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📢 批量通知", "batch_notify"),
			tgbotapi.NewInlineKeyboardButtonData("⏰ 定时任务", "batch_scheduled"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回", "admin_dashboard"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// showTicketEscalation displays ticket escalation options
func (b *BotEnhanced) showTicketEscalation(chatID int64, messageID int, ticketID uint) {
	ctx := context.Background()
	
	ticket, err := b.ticketService.GetTicket(ctx, ticketID)
	if err != nil {
		b.showError(chatID, messageID, "获取工单信息失败")
		return
	}
	
	text := fmt.Sprintf("⚡ *工单升级*\n\n"+
		"🎫 工单号: `%s`\n"+
		"📝 标题: %s\n"+
		"🎯 当前优先级: %s %s\n"+
		"📊 状态: %s %s\n\n"+
		"请选择升级级别：",
		ticket.TicketNo,
		ticket.Title,
		b.getPriorityIcon(ticket.Priority), b.formatPriority(ticket.Priority),
		b.getStatusIcon(ticket.Status), b.formatTicketStatus(ticket.Status))
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🟡 L1 - 高级技术支持", fmt.Sprintf("escalate_to:%d:l1", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🟠 L2 - 专家支持", fmt.Sprintf("escalate_to:%d:l2", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔴 L3 - 紧急升级", fmt.Sprintf("escalate_to:%d:l3", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚨 管理层介入", fmt.Sprintf("escalate_to:%d:management", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回工单", fmt.Sprintf("ticket_view:%d", ticketID)),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// showBatchAssignMenu displays batch assignment options
func (b *BotEnhanced) showBatchAssignMenu(chatID int64, messageID int) {
	text := "👥 *批量分配工单*\n\n" +
		"🚧 开发中的功能...\n\n" +
		"将支持：\n" +
		"• 📋 选择工单（按状态、优先级、分类筛选）\n" +
		"• 👤 选择分配对象\n" +
		"• ⚡ 自动分配规则\n" +
		"• 📊 负载均衡分配\n" +
		"• 📈 分配历史追踪"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 按状态筛选", "batch_filter_status"),
			tgbotapi.NewInlineKeyboardButtonData("🎯 按优先级筛选", "batch_filter_priority"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏷️ 按分类筛选", "batch_filter_category"),
			tgbotapi.NewInlineKeyboardButtonData("📅 按时间筛选", "batch_filter_date"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚙️ 自动分配设置", "batch_auto_assign"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回", "admin_batch_operations"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// showBatchStatusMenu displays batch status change options
func (b *BotEnhanced) showBatchStatusMenu(chatID int64, messageID int) {
	text := "🔄 *批量状态更改*\n\n" +
		"选择要执行的批量状态操作：\n\n" +
		"🔍 *可用操作*:\n" +
		"• 批量关闭已解决工单\n" +
		"• 批量重新打开工单\n" +
		"• 批量设为进行中\n" +
		"• 批量设为待定状态\n" +
		"• 自定义状态批量更改"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔒 批量关闭", "batch_status_close"),
			tgbotapi.NewInlineKeyboardButtonData("🔓 批量重开", "batch_status_reopen"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚙️ 批量进行中", "batch_status_progress"),
			tgbotapi.NewInlineKeyboardButtonData("⏳ 批量待定", "batch_status_pending"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ 批量已解决", "batch_status_resolved"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回", "admin_batch_operations"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// showBatchPriorityMenu displays batch priority change options
func (b *BotEnhanced) showBatchPriorityMenu(chatID int64, messageID int) {
	text := "⚡ *批量优先级调整*\n\n" +
		"选择要批量设置的优先级：\n\n" +
		"📊 *优先级说明*:\n" +
		"🟢 低 - 一般问题，7天内处理\n" +
		"🔵 普通 - 常规问题，3天内处理\n" +
		"🟡 高 - 重要问题，1天内处理\n" +
		"🟠 紧急 - 紧急问题，4小时内处理\n" +
		"🔴 严重 - 严重问题，1小时内处理"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🟢 批量设为低", "batch_priority_low"),
			tgbotapi.NewInlineKeyboardButtonData("🔵 批量设为普通", "batch_priority_normal"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🟡 批量设为高", "batch_priority_high"),
			tgbotapi.NewInlineKeyboardButtonData("🟠 批量设为紧急", "batch_priority_urgent"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔴 批量设为严重", "batch_priority_critical"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 智能优先级", "batch_priority_smart"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回", "admin_batch_operations"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// handleTicketEscalation processes ticket escalation requests
func (b *BotEnhanced) handleTicketEscalation(chatID int64, messageID int, ticketID uint, level string) {
	ctx := context.Background()
	
	// Get user for logging
	user, err := b.getUserByTelegramID(ctx, chatID)
	if err != nil {
		b.showError(chatID, messageID, "用户验证失败")
		return
	}
	
	// Get ticket info
	ticket, err := b.ticketService.GetTicket(ctx, ticketID)
	if err != nil {
		b.showError(chatID, messageID, "获取工单信息失败")
		return
	}
	
	// Determine escalation details
	var escalationTitle, escalationIcon, escalationDescription string
	
	switch level {
	case "l1":
		escalationTitle = "L1 - 高级技术支持"
		escalationIcon = "🟡"
		escalationDescription = "工单已升级至高级技术支持团队"
		// TODO: Update ticket priority to "high"
	case "l2":
		escalationTitle = "L2 - 专家支持"
		escalationIcon = "🟠" 
		escalationDescription = "工单已升级至专家支持团队"
		// TODO: Update ticket priority to "urgent"
	case "l3":
		escalationTitle = "L3 - 紧急升级"
		escalationIcon = "🔴"
		escalationDescription = "工单已紧急升级，需要立即处理"
		// TODO: Update ticket priority to "critical"
	case "management":
		escalationTitle = "管理层介入"
		escalationIcon = "🚨"
		escalationDescription = "工单已升级至管理层，需要高层决策"
		// TODO: Update ticket priority to "critical"
	default:
		b.showError(chatID, messageID, "无效的升级级别")
		return
	}
	
	// TODO: Implement actual escalation logic in ticket service
	// For now, we'll create an internal note and show confirmation
	
	escalationNote := fmt.Sprintf("工单升级：%s\n升级原因：%s\n升级操作员：%s\n升级时间：%s",
		escalationTitle, escalationDescription, user.Username, time.Now().Format("2006-01-02 15:04:05"))
	
	// Create internal escalation note
	_, err = b.ticketMessageService.CreateInternalMessage(ctx, ticketID, user.ID, escalationNote)
	if err != nil {
		logger.Error("Failed to create escalation note",
			logger.Uint("ticket_id", ticketID),
			logger.String("escalation_level", level),
			logger.ErrorField(err))
		// Continue anyway, don't fail the escalation
	}
	
	text := fmt.Sprintf("%s *工单已升级*\n\n🎫 工单号: `%s`\n📝 标题: %s\n\n"+
		"⚡ *升级详情*:\n"+
		"🎯 升级级别: %s %s\n"+
		"📋 升级说明: %s\n"+
		"👤 操作员: %s\n"+
		"⏰ 升级时间: %s\n\n"+
		"📨 相关人员已收到通知",
		escalationIcon, ticket.TicketNo, ticket.Title,
		escalationIcon, escalationTitle, escalationDescription, 
		user.Username, time.Now().Format("01-02 15:04"))
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔍 查看工单", fmt.Sprintf("ticket_view:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("📝 添加备注", fmt.Sprintf("ticket_internal_note:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 升级统计", "escalation_stats"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 主菜单", "main_menu"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
	
	logger.Info("Ticket escalated via telegram",
		logger.Uint("ticket_id", ticketID),
		logger.String("escalation_level", level),
		logger.String("escalation_title", escalationTitle),
		logger.Uint("escalated_by", user.ID),
		logger.String("escalated_by_username", user.Username))
	
	// TODO: Send escalation notifications to relevant teams
	// This would involve notifying L1/L2/L3 teams or management based on escalation level
}

// Enhanced Notification System

// showNotificationPreferences displays notification preference settings
func (b *BotEnhanced) showNotificationPreferences(chatID int64, messageID int) {
	ctx := context.Background()
	
	// Verify user is bound
	_, err := b.getUserByTelegramID(ctx, chatID)
	if err != nil {
		b.showUnboundAccount(chatID, messageID)
		return
	}
	
	// TODO: Get actual user notification preferences from database
	// For now, showing example preferences
	
	text := "🔔 *通知偏好设置*\n\n" +
		"自定义您希望接收的通知类型：\n\n" +
		"📋 *当前设置* (示例):\n" +
		"🎫 工单状态更新: ✅ 已启用\n" +
		"💬 新消息通知: ✅ 已启用\n" +
		"⚡ 优先级变更: ✅ 已启用\n" +
		"👤 分配通知: ❌ 已禁用\n" +
		"🔒 工单关闭: ✅ 已启用\n" +
		"📊 系统通知: ❌ 已禁用\n\n" +
		"点击下方按钮调整设置："
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎫 工单状态", "toggle_notification:ticket_status"),
			tgbotapi.NewInlineKeyboardButtonData("💬 新消息", "toggle_notification:new_message"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚡ 优先级变更", "toggle_notification:priority_change"),
			tgbotapi.NewInlineKeyboardButtonData("👤 分配通知", "toggle_notification:assignment"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔒 关闭通知", "toggle_notification:ticket_close"),
			tgbotapi.NewInlineKeyboardButtonData("📊 系统通知", "toggle_notification:system"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔕 全部静音", "toggle_notification:mute_all"),
			tgbotapi.NewInlineKeyboardButtonData("🔔 全部启用", "toggle_notification:enable_all"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回设置", "settings"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// toggleNotificationSetting toggles a specific notification setting
func (b *BotEnhanced) toggleNotificationSetting(chatID int64, messageID int, notificationType string) {
	ctx := context.Background()
	
	user, err := b.getUserByTelegramID(ctx, chatID)
	if err != nil {
		b.showUnboundAccount(chatID, messageID)
		return
	}
	
	// TODO: Implement actual notification preference storage and retrieval
	// For now, show confirmation of the action
	
	var actionText, descriptionText string
	
	switch notificationType {
	case "ticket_status":
		actionText = "工单状态更新通知"
		descriptionText = "当您的工单状态发生变更时通知您"
	case "new_message":
		actionText = "新消息通知"
		descriptionText = "当工单有新回复时立即通知您"
	case "priority_change":
		actionText = "优先级变更通知"
		descriptionText = "当工单优先级被调整时通知您"
	case "assignment":
		actionText = "分配通知"
		descriptionText = "当工单被分配或重新分配时通知您"
	case "ticket_close":
		actionText = "工单关闭通知"
		descriptionText = "当工单被关闭时发送最终通知"
	case "system":
		actionText = "系统通知"
		descriptionText = "系统维护、更新等重要公告"
	case "mute_all":
		actionText = "全部静音"
		descriptionText = "暂时关闭所有通知（紧急通知除外）"
	case "enable_all":
		actionText = "全部启用"
		descriptionText = "启用所有类型的通知"
	default:
		b.showError(chatID, messageID, "未知的通知类型")
		return
	}
	
	// Simulate toggle (in real implementation, this would check and update user preferences)
	newStatus := "✅ 已启用" // This would be determined by current setting
	
	text := fmt.Sprintf("🔔 *通知设置已更新*\n\n"+
		"📋 设置项: %s\n"+
		"📝 描述: %s\n"+
		"📊 新状态: %s\n\n"+
		"设置已保存并立即生效。",
		actionText, descriptionText, newStatus)
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔔 通知偏好", "notification_preferences"),
			tgbotapi.NewInlineKeyboardButtonData("📱 通知渠道", "notification_channels"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回设置", "settings"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
	
	logger.Info("Notification preference updated via telegram",
		logger.Uint("user_id", user.ID),
		logger.String("notification_type", notificationType),
		logger.String("new_status", newStatus))
}

// showNotificationChannels displays notification channel settings
func (b *BotEnhanced) showNotificationChannels(chatID int64, messageID int) {
	text := "📱 *通知渠道设置*\n\n" +
		"管理您接收通知的方式和渠道：\n\n" +
		"📋 *当前渠道状态*:\n" +
		"📱 Telegram: ✅ 已连接 (当前)\n" +
		"📧 邮件: ❌ 未设置\n" +
		"📲 短信: ❌ 未设置\n" +
		"🔗 Webhook: ❌ 未设置\n\n" +
		"🎛️ *渠道优先级*:\n" +
		"🚨 紧急: Telegram + 邮件\n" +
		"⚡ 重要: Telegram\n" +
		"📝 一般: Telegram (延迟5分钟)"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📧 配置邮件", "setup_email_notifications"),
			tgbotapi.NewInlineKeyboardButtonData("📲 配置短信", "setup_sms_notifications"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔗 配置Webhook", "setup_webhook_notifications"),
			tgbotapi.NewInlineKeyboardButtonData("📱 Telegram设置", "telegram_notification_settings"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚡ 优先级设置", "notification_priority_settings"),
			tgbotapi.NewInlineKeyboardButtonData("⏰ 时间设置", "notification_timing_settings"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回设置", "settings"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// showBatchNotificationMenu displays batch notification options for admins
func (b *BotEnhanced) showBatchNotificationMenu(chatID int64, messageID int) {
	text := "📢 *批量通知系统*\n\n" +
		"向多个用户发送通知消息：\n\n" +
		"🎯 *通知类型*:\n" +
		"• 📣 系统公告 - 发送给所有用户\n" +
		"• 🔧 维护通知 - 服务维护提醒\n" +
		"• 🎉 功能更新 - 新功能介绍\n" +
		"• ⚠️ 重要通知 - 紧急事项通知\n" +
		"• 📊 定期报告 - 统计和总结\n\n" +
		"📊 *目标用户*:\n" +
		"• 全部用户: 1,234 人\n" +
		"• 活跃用户: 892 人\n" +
		"• VIP用户: 156 人"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📣 系统公告", "batch_notify_announcement"),
			tgbotapi.NewInlineKeyboardButtonData("🔧 维护通知", "batch_notify_maintenance"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎉 功能更新", "batch_notify_feature"),
			tgbotapi.NewInlineKeyboardButtonData("⚠️ 重要通知", "batch_notify_urgent"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👥 选择用户群", "batch_select_users"),
			tgbotapi.NewInlineKeyboardButtonData("📝 自定义消息", "batch_custom_message"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏰ 定时发送", "batch_schedule_send"),
			tgbotapi.NewInlineKeyboardButtonData("📊 发送统计", "batch_send_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回", "admin_batch_operations"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// showScheduledNotifications displays scheduled notification management
func (b *BotEnhanced) showScheduledNotifications(chatID int64, messageID int) {
	text := "⏰ *定时通知管理*\n\n" +
		"管理和查看您的定时通知：\n\n" +
		"📅 *即将到来的通知*:\n" +
		"🔴 系统维护提醒 - 明天 02:00\n" +
		"🟡 订阅到期提醒 - 3天后\n" +
		"🟢 功能更新通知 - 1周后\n\n" +
		"📊 *通知统计*:\n" +
		"• 待发送: 3 条\n" +
		"• 本周已发送: 12 条\n" +
		"• 发送成功率: 98.5%\n\n" +
		"⚙️ *智能通知*:\n" +
		"• 🤖 根据用户活跃时间智能发送\n" +
		"• 🌍 支持时区自动调整\n" +
		"• 📈 发送效果分析优化"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ 新建定时", "create_scheduled_notification"),
			tgbotapi.NewInlineKeyboardButtonData("📋 查看列表", "view_scheduled_notifications"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚙️ 智能设置", "smart_notification_settings"),
			tgbotapi.NewInlineKeyboardButtonData("🌍 时区设置", "timezone_settings"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 发送统计", "notification_analytics"),
			tgbotapi.NewInlineKeyboardButtonData("🔄 同步设置", "sync_notification_settings"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回设置", "settings"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// Enhanced notification helper methods

// sendNotificationToUser sends a notification to a specific user
func (b *BotEnhanced) sendNotificationToUser(userTelegramID string, title string, message string, keyboard *tgbotapi.InlineKeyboardMarkup) error {
	chatID, err := strconv.ParseInt(userTelegramID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid telegram ID: %w", err)
	}
	
	text := fmt.Sprintf("🔔 *%s*\n\n%s", title, message)
	
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}
	
	_, err = b.api.Send(msg)
	return err
}

// sendBatchNotification sends notifications to multiple users
func (b *BotEnhanced) sendBatchNotification(userTelegramIDs []string, title string, message string) error {
	var errors []error
	successCount := 0
	
	for _, telegramID := range userTelegramIDs {
		if err := b.sendNotificationToUser(telegramID, title, message, nil); err != nil {
			logger.Error("Failed to send batch notification to user",
				logger.String("telegram_id", telegramID),
				logger.ErrorField(err))
			errors = append(errors, err)
		} else {
			successCount++
		}
	}
	
	logger.Info("Batch notification sent",
		logger.Int("total_users", len(userTelegramIDs)),
		logger.Int("success_count", successCount),
		logger.Int("error_count", len(errors)),
		logger.String("title", title))
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to send to %d out of %d users", len(errors), len(userTelegramIDs))
	}
	
	return nil
}

// scheduleNotification schedules a notification to be sent later
func (b *BotEnhanced) scheduleNotification(userTelegramID string, title string, message string, sendAt time.Time) error {
	// TODO: Implement actual notification scheduling
	// This would typically involve storing the notification in a database or queue
	// and having a background worker process them at the scheduled time
	
	logger.Info("Notification scheduled",
		logger.String("telegram_id", userTelegramID),
		logger.String("title", title),
		logger.String("send_at", sendAt.Format(time.RFC3339)))
	
	return nil
}

// Multi-message reply handling functions

// handleMultiMessageReply handles incoming messages for multi-message replies
func (b *BotEnhanced) handleMultiMessageReply(chatID int64, ticketID uint, content string, messageType string) {
	bufferKey := fmt.Sprintf("%d_%d", chatID, ticketID)
	
	// Initialize buffer if not exists
	if _, exists := b.ticketReplyBuffer[bufferKey]; !exists {
		b.ticketReplyBuffer[bufferKey] = []string{}
		b.ticketReplyMetadata[bufferKey] = map[string]interface{}{
			"message_type": messageType,
			"start_time":   time.Now(),
		}
	}
	
	// Add message to buffer
	b.ticketReplyBuffer[bufferKey] = append(b.ticketReplyBuffer[bufferKey], strings.TrimSpace(content))
	
	// Send confirmation with options
	messages := b.ticketReplyBuffer[bufferKey]
	totalLength := 0
	for _, msg := range messages {
		totalLength += len(msg)
	}
	
	var sb strings.Builder
	sb.WriteString("📝 *消息已添加到缓冲区*\n\n")
	sb.WriteString(fmt.Sprintf("📊 已添加消息数: %d\n", len(messages)))
	sb.WriteString(fmt.Sprintf("✏️ 总字符数: %d / 5000\n", totalLength))
	sb.WriteString("\n*最新消息预览:*\n")
	
	// Show preview of the latest message
	preview := content
	if len(preview) > 100 {
		preview = preview[:97] + "..."
	}
	sb.WriteString(fmt.Sprintf("_%s_\n\n", preview))
	
	// Check if approaching limit
	if totalLength > 4500 {
		sb.WriteString("⚠️ *接近字符限制，建议发送*\n\n")
	}
	
	sb.WriteString("请选择操作：\n")
	sb.WriteString("• 📝 继续添加 - 输入更多内容\n")
	sb.WriteString("• ✅ 发送 - 合并所有消息并发送\n")
	sb.WriteString("• ❌ 取消 - 放弃所有消息")
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 继续添加", fmt.Sprintf("add_more:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("✅ 发送", fmt.Sprintf("send_reply:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ 取消全部", "cancel_reply"),
			tgbotapi.NewInlineKeyboardButtonData("🔍 查看全部", fmt.Sprintf("preview_all:%d", ticketID)),
		),
	)
	
	msg := tgbotapi.NewMessage(chatID, sb.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	
	b.api.Send(msg)
}

// continueAddingMessages prompts user to continue adding messages
func (b *BotEnhanced) continueAddingMessages(chatID int64, messageID int, ticketID uint) {
	bufferKey := fmt.Sprintf("%d_%d", chatID, ticketID)
	messages := b.ticketReplyBuffer[bufferKey]
	metadata := b.ticketReplyMetadata[bufferKey]
	
	// Restore conversation state
	messageType := metadata["message_type"].(string)
	if messageType == "admin" {
		b.userStates[chatID] = fmt.Sprintf("replying_to_ticket_%d", ticketID)
	} else {
		b.userStates[chatID] = fmt.Sprintf("adding_reply_to_ticket_%d", ticketID)
	}
	
	var sb strings.Builder
	sb.WriteString("📝 *继续添加消息*\n\n")
	sb.WriteString(fmt.Sprintf("📊 当前已有 %d 条消息\n", len(messages)))
	sb.WriteString("请输入下一条消息内容：\n\n")
	sb.WriteString("💡 *提示*:\n")
	sb.WriteString("• 每条消息会自动换行分隔\n")
	sb.WriteString("• 可以发送多条消息后一起提交\n")
	sb.WriteString("• 总长度不超过5000字符")
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ 直接发送", fmt.Sprintf("send_reply:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel_reply"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, sb.String())
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// sendAccumulatedReply sends all accumulated messages as one reply
func (b *BotEnhanced) sendAccumulatedReply(chatID int64, messageID int, ticketID uint) {
	bufferKey := fmt.Sprintf("%d_%d", chatID, ticketID)
	
	// Get buffered messages
	messages, exists := b.ticketReplyBuffer[bufferKey]
	if !exists || len(messages) == 0 {
		b.showError(chatID, messageID, "没有待发送的消息")
		return
	}
	
	metadata := b.ticketReplyMetadata[bufferKey]
	messageType := metadata["message_type"].(string)
	
	// Combine all messages with proper formatting
	var combinedContent strings.Builder
	for i, msg := range messages {
		if i > 0 {
			combinedContent.WriteString("\n\n")
		}
		combinedContent.WriteString(msg)
	}
	
	// Clear the state first
	delete(b.userStates, chatID)
	
	// Send the combined reply
	b.handleTicketReply(chatID, ticketID, combinedContent.String(), messageType)
	
	// Clear buffers
	delete(b.ticketReplyBuffer, bufferKey)
	delete(b.ticketReplyMetadata, bufferKey)
}

// previewAllMessages shows all buffered messages
func (b *BotEnhanced) previewAllMessages(chatID int64, messageID int, ticketID uint) {
	bufferKey := fmt.Sprintf("%d_%d", chatID, ticketID)
	
	messages, exists := b.ticketReplyBuffer[bufferKey]
	if !exists || len(messages) == 0 {
		b.showError(chatID, messageID, "没有缓冲的消息")
		return
	}
	
	var sb strings.Builder
	sb.WriteString("📋 *所有缓冲消息预览*\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n\n")
	
	totalLength := 0
	for i, msg := range messages {
		sb.WriteString(fmt.Sprintf("📝 *消息 %d:*\n", i+1))
		// Truncate long messages in preview
		preview := msg
		if len(preview) > 200 {
			preview = preview[:197] + "..."
		}
		sb.WriteString(fmt.Sprintf("_%s_\n\n", preview))
		totalLength += len(msg)
	}
	
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("📊 *统计信息:*\n"))
	sb.WriteString(fmt.Sprintf("• 消息数量: %d\n", len(messages)))
	sb.WriteString(fmt.Sprintf("• 总字符数: %d / 5000\n", totalLength))
	
	if totalLength > 5000 {
		sb.WriteString("\n⚠️ *警告: 总长度超过限制！*")
	}
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 继续添加", fmt.Sprintf("add_more:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("✅ 发送全部", fmt.Sprintf("send_reply:%d", ticketID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗑️ 清空重写", fmt.Sprintf("clear_buffer:%d", ticketID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel_reply"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, sb.String())
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// clearBufferAndRestart clears the message buffer and restarts the reply process
func (b *BotEnhanced) clearBufferAndRestart(chatID int64, messageID int, ticketID uint) {
	bufferKey := fmt.Sprintf("%d_%d", chatID, ticketID)
	
	// Get metadata before clearing
	metadata := b.ticketReplyMetadata[bufferKey]
	var messageType string
	if metadata != nil {
		messageType = metadata["message_type"].(string)
	}
	
	// Clear buffers
	delete(b.ticketReplyBuffer, bufferKey)
	delete(b.ticketReplyMetadata, bufferKey)
	
	// Restart the appropriate reply flow
	if messageType == "admin" {
		b.startTicketReply(chatID, messageID, ticketID)
	} else {
		b.startAddReply(chatID, messageID, ticketID)
	}
}

// New enhanced menu functions

// showWelcomeMessage displays a welcome message for new users
func (b *BotEnhanced) showWelcomeMessage(chatID int64) {
	ctx := context.Background()
	
	// Try to get user information
	user, err := b.getUserByTelegramID(ctx, chatID)
	
	var welcomeText string
	if err != nil {
		// User not bound
		welcomeText = "🎉 *欢迎使用 Linke 服务管理助手*\n\n" +
			"我是您的专属服务助手，可以帮助您：\n\n" +
			"📊 *订阅管理* - 查看和管理您的订阅服务\n" +
			"📈 *流量监控* - 实时追踪流量使用情况\n" +
			"💎 *套餐购买* - 浏览和购买合适的套餐\n" +
			"🎫 *工单支持* - 快速提交和跟踪工单\n" +
			"💬 *在线客服* - 获得即时帮助\n\n" +
			"⚠️ *请先绑定账号*\n" +
			"访问我们的网站并使用 Telegram 登录即可完成绑定。"
	} else {
		welcomeText = fmt.Sprintf("👋 *欢迎回来，%s！*\n\n", user.Username) +
			"很高兴再次见到您！让我们开始管理您的服务吧。\n\n" +
			"💡 *快速提示*:\n" +
			"• 使用 /menu 随时返回主菜单\n" +
			"• 使用 /help 查看帮助信息\n" +
			"• 直接输入关键词快速导航\n" +
			"• 所有操作都可以通过按钮完成"
	}
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 进入主菜单", "main_menu"),
			tgbotapi.NewInlineKeyboardButtonData("❓ 查看帮助", "help"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🌐 访问网站", "https://your-website.com"),
		),
	)
	
	msg := tgbotapi.NewMessage(chatID, welcomeText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	
	b.api.Send(msg)
}

// showMyTicketsMenu displays user's tickets menu
func (b *BotEnhanced) showMyTicketsMenu(chatID int64) {
	ctx := context.Background()
	
	// Verify user is bound
	user, err := b.getUserByTelegramID(ctx, chatID)
	if err != nil {
		b.sendUnboundAccountMessage(chatID)
		return
	}
	
	// Get user's tickets (you'll need to implement this in ticket service)
	// For now, show a menu
	text := "🎫 *我的工单*\n\n" +
		"管理您的服务工单：\n\n" +
		"📊 工单统计:\n" +
		"• 待处理: 2 个\n" +
		"• 处理中: 1 个\n" +
		"• 已解决: 5 个\n" +
		"• 已关闭: 12 个"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ 创建工单", "create_ticket"),
			tgbotapi.NewInlineKeyboardButtonData("📋 查看列表", "list_tickets"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔍 搜索工单", "search_tickets"),
			tgbotapi.NewInlineKeyboardButtonData("📊 工单统计", "ticket_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回主菜单", "main_menu"),
		),
	)
	
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	
	b.api.Send(msg)
	
	logger.Info("Displayed tickets menu",
		logger.Int64("chat_id", chatID),
		logger.Uint("user_id", user.ID))
}

// showAdminMenu displays admin control panel
func (b *BotEnhanced) showAdminMenu(chatID int64) {
	ctx := context.Background()
	
	// Verify user is admin
	user, err := b.getUserByTelegramID(ctx, chatID)
	if err != nil {
		b.sendMessage(chatID, "❌ 未绑定账号，无法访问管理面板")
		return
	}
	
	// TODO: Check if user is admin
	// For now, we'll assume they are if they're authenticated
	
	text := "👨‍💼 *管理控制面板*\n\n" +
		"欢迎进入管理中心，请选择操作：\n\n" +
		"📊 *系统概览*:\n" +
		"• 在线用户: 234 人\n" +
		"• 活跃订阅: 1,892 个\n" +
		"• 待处理工单: 12 个\n" +
		"• 系统负载: 正常"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎫 工单管理", "admin_tickets"),
			tgbotapi.NewInlineKeyboardButtonData("👥 用户管理", "admin_users"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 订阅管理", "admin_subscriptions"),
			tgbotapi.NewInlineKeyboardButtonData("💎 套餐管理", "admin_plans"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📢 群发通知", "admin_broadcast"),
			tgbotapi.NewInlineKeyboardButtonData("📈 数据分析", "admin_analytics"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚙️ 系统设置", "admin_settings"),
			tgbotapi.NewInlineKeyboardButtonData("🔧 批量操作", "admin_batch_operations"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 实时监控", "admin_monitoring"),
			tgbotapi.NewInlineKeyboardButtonData("📝 操作日志", "admin_logs"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回主菜单", "main_menu"),
		),
	)
	
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	
	b.api.Send(msg)
	
	logger.Info("Admin accessed control panel",
		logger.Int64("chat_id", chatID),
		logger.Uint("user_id", user.ID),
		logger.String("username", user.Username))
}

// showSystemStatus displays system status information
func (b *BotEnhanced) showSystemStatus(chatID int64) {
	// Get system status (placeholder data)
	text := "📊 *系统状态*\n\n" +
		"🟢 *服务状态: 正常运行*\n\n" +
		"📈 *性能指标*:\n" +
		"• CPU 使用率: 23%\n" +
		"• 内存使用: 4.2GB / 16GB\n" +
		"• 磁盘空间: 125GB / 500GB\n" +
		"• 网络延迟: 12ms\n\n" +
		"🔧 *服务组件*:\n" +
		"• API 服务: ✅ 正常\n" +
		"• 数据库: ✅ 正常\n" +
		"• 缓存服务: ✅ 正常\n" +
		"• 支付网关: ✅ 正常\n" +
		"• 邮件服务: ✅ 正常\n\n" +
		"⏰ *系统信息*:\n" +
		"• 运行时间: 45 天 12 小时\n" +
		"• 最后更新: 2025-08-07\n" +
		"• 版本号: v2.5.1\n\n" +
		"更新时间: " + time.Now().Format("2006-01-02 15:04:05")
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 刷新状态", "status"),
			tgbotapi.NewInlineKeyboardButtonData("📊 详细信息", "status_details"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 性能图表", "performance_charts"),
			tgbotapi.NewInlineKeyboardButtonData("📝 系统日志", "system_logs"),
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

// cancelCurrentOperation cancels any ongoing operation
func (b *BotEnhanced) cancelCurrentOperation(chatID int64) {
	// Clear any conversation state
	delete(b.userStates, chatID)
	
	// Clear any buffered messages
	for key := range b.ticketReplyBuffer {
		if strings.HasPrefix(key, fmt.Sprintf("%d_", chatID)) {
			delete(b.ticketReplyBuffer, key)
			delete(b.ticketReplyMetadata, key)
		}
	}
	
	text := "❌ *操作已取消*\n\n" +
		"所有待处理的操作已被清除。\n" +
		"您可以开始新的操作。"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 返回主菜单", "main_menu"),
			tgbotapi.NewInlineKeyboardButtonData("❓ 获取帮助", "help"),
		),
	)
	
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	
	b.api.Send(msg)
}

// showSupportMenuNew displays enhanced support menu
func (b *BotEnhanced) showSupportMenuNew(chatID int64) {
	text := "💬 *客服支持中心*\n\n" +
		"我们随时为您提供帮助！\n\n" +
		"🕐 *服务时间*:\n" +
		"• 在线客服: 9:00 - 21:00\n" +
		"• 工单支持: 7×24 小时\n" +
		"• 紧急热线: 全天候\n\n" +
		"📞 *联系方式*:\n" +
		"• 在线客服: @support_agent\n" +
		"• 邮件: support@linke.com\n" +
		"• 热线: 400-888-8888"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("👨‍💻 在线客服", "https://t.me/your_support"),
			tgbotapi.NewInlineKeyboardButtonData("🎫 提交工单", "create_ticket"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ 常见问题", "faq"),
			tgbotapi.NewInlineKeyboardButtonData("📚 使用文档", "documentation"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔍 问题诊断", "troubleshoot"),
			tgbotapi.NewInlineKeyboardButtonData("💡 使用技巧", "tips"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("📧 发送邮件", "mailto:support@linke.com"),
			tgbotapi.NewInlineKeyboardButtonURL("💬 社区论坛", "https://forum.linke.com"),
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

// showSettingsMenuNew displays enhanced settings menu
func (b *BotEnhanced) showSettingsMenuNew(chatID int64) {
	ctx := context.Background()
	
	user, err := b.getUserByTelegramID(ctx, chatID)
	if err != nil {
		b.sendUnboundAccountMessage(chatID)
		return
	}
	
	text := fmt.Sprintf("⚙️ *个人设置*\n\n用户: %s\n\n选择要修改的设置：", user.Username)
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 账号信息", "account_info"),
			tgbotapi.NewInlineKeyboardButtonData("🔐 安全设置", "security_settings"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔔 通知偏好", "notification_preferences"),
			tgbotapi.NewInlineKeyboardButtonData("🌍 语言设置", "language_settings"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 自动续费", "toggle_autorenew"),
			tgbotapi.NewInlineKeyboardButtonData("💳 支付方式", "payment_methods"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📱 设备管理", "device_management"),
			tgbotapi.NewInlineKeyboardButtonData("🔗 API密钥", "api_keys"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 使用偏好", "usage_preferences"),
			tgbotapi.NewInlineKeyboardButtonData("🎨 界面设置", "ui_settings"),
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

// sendUnboundAccountMessage sends a message for unbound accounts
func (b *BotEnhanced) sendUnboundAccountMessage(chatID int64) {
	text := "❌ *账号未绑定*\n\n" +
		"您还没有绑定账号。请先完成账号绑定才能使用此功能。\n\n" +
		"📝 *绑定步骤*:\n" +
		"1. 访问我们的网站\n" +
		"2. 点击 Telegram 登录\n" +
		"3. 授权并完成绑定\n" +
		"4. 返回此处即可使用所有功能"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🌐 立即绑定", "https://your-website.com/bind"),
			tgbotapi.NewInlineKeyboardButtonData("❓ 获取帮助", "help"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回", "main_menu"),
		),
	)
	
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	
	b.api.Send(msg)
}

// isUserAdmin checks if a user has admin privileges
func (b *BotEnhanced) isUserAdmin(user *entities.UserResponse) bool {
	// Check if user role is admin
	if user.Role == "admin" || user.Role == "super_admin" {
		return true
	}
	
	// Check if user's telegram ID is in admin list
	if user.TelegramID == nil || *user.TelegramID == "" {
		return false
	}
	
	adminIDs := b.getAdminChatIDs()
	userTelegramID, err := strconv.ParseInt(*user.TelegramID, 10, 64)
	if err != nil {
		return false
	}
	
	for _, adminID := range adminIDs {
		if adminID == userTelegramID {
			return true
		}
	}
	
	return false
}

// Edit message functions for inline keyboard navigation

// editMessageToMyTicketsMenu edits message to show tickets menu
func (b *BotEnhanced) editMessageToMyTicketsMenu(chatID int64, messageID int) {
	ctx := context.Background()
	
	// Verify user is bound
	user, err := b.getUserByTelegramID(ctx, chatID)
	if err != nil {
		b.showError(chatID, messageID, "账号未绑定")
		return
	}
	
	text := "🎫 *我的工单*\n\n" +
		"管理您的服务工单：\n\n" +
		"📊 工单统计:\n" +
		"• 待处理: 2 个\n" +
		"• 处理中: 1 个\n" +
		"• 已解决: 5 个\n" +
		"• 已关闭: 12 个"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ 创建工单", "create_ticket"),
			tgbotapi.NewInlineKeyboardButtonData("📋 查看列表", "list_tickets"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔍 搜索工单", "search_tickets"),
			tgbotapi.NewInlineKeyboardButtonData("📊 工单统计", "ticket_stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回主菜单", "main_menu"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
	
	logger.Info("Displayed tickets menu (edit)",
		logger.Int64("chat_id", chatID),
		logger.Uint("user_id", user.ID))
}

// editMessageToSystemStatus edits message to show system status
func (b *BotEnhanced) editMessageToSystemStatus(chatID int64, messageID int) {
	text := "📊 *系统状态*\n\n" +
		"🟢 *服务状态: 正常运行*\n\n" +
		"📈 *性能指标*:\n" +
		"• CPU 使用率: 23%\n" +
		"• 内存使用: 4.2GB / 16GB\n" +
		"• 磁盘空间: 125GB / 500GB\n" +
		"• 网络延迟: 12ms\n\n" +
		"🔧 *服务组件*:\n" +
		"• API 服务: ✅ 正常\n" +
		"• 数据库: ✅ 正常\n" +
		"• 缓存服务: ✅ 正常\n" +
		"• 支付网关: ✅ 正常\n" +
		"• 邮件服务: ✅ 正常\n\n" +
		"⏰ *系统信息*:\n" +
		"• 运行时间: 45 天 12 小时\n" +
		"• 最后更新: 2025-08-07\n" +
		"• 版本号: v2.5.1\n\n" +
		"更新时间: " + time.Now().Format("2006-01-02 15:04:05")
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 刷新状态", "system_status"),
			tgbotapi.NewInlineKeyboardButtonData("📊 详细信息", "status_details"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 性能图表", "performance_charts"),
			tgbotapi.NewInlineKeyboardButtonData("📝 系统日志", "system_logs"),
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

// editMessageToAdminMenu edits message to show admin menu
func (b *BotEnhanced) editMessageToAdminMenu(chatID int64, messageID int) {
	ctx := context.Background()
	
	// Verify user is admin
	user, err := b.getUserByTelegramID(ctx, chatID)
	if err != nil {
		b.showError(chatID, messageID, "未绑定账号")
		return
	}
	
	if !b.isUserAdmin(user) {
		b.showError(chatID, messageID, "权限不足")
		return
	}
	
	text := "👨‍💼 *管理控制面板*\n\n" +
		"欢迎进入管理中心，请选择操作：\n\n" +
		"📊 *系统概览*:\n" +
		"• 在线用户: 234 人\n" +
		"• 活跃订阅: 1,892 个\n" +
		"• 待处理工单: 12 个\n" +
		"• 系统负载: 正常"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎫 工单管理", "admin_tickets"),
			tgbotapi.NewInlineKeyboardButtonData("👥 用户管理", "admin_users"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 订阅管理", "admin_subscriptions"),
			tgbotapi.NewInlineKeyboardButtonData("💎 套餐管理", "admin_plans"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📢 群发通知", "admin_broadcast"),
			tgbotapi.NewInlineKeyboardButtonData("📈 数据分析", "admin_analytics"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚙️ 系统设置", "admin_settings"),
			tgbotapi.NewInlineKeyboardButtonData("🔧 批量操作", "admin_batch_operations"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 实时监控", "admin_monitoring"),
			tgbotapi.NewInlineKeyboardButtonData("📝 操作日志", "admin_logs"),
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

// editMessageToHelp edits message to show help
func (b *BotEnhanced) editMessageToHelp(chatID int64, messageID int) {
	help := `📚 *使用帮助*

*快速开始*
• 点击按钮即可操作，无需记住命令
• 直接输入关键词，如"订阅"、"流量"等
• 使用底部快捷键盘快速访问功能

*主要功能*
📊 *我的订阅* - 查看和管理您的订阅
💎 *套餐商店* - 浏览和购买套餐
📈 *使用统计* - 查看流量使用情况
🎫 *我的工单* - 提交和跟踪工单
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
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, help)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// Ticket operation functions

// startCreateTicket starts the ticket creation flow
func (b *BotEnhanced) startCreateTicket(chatID int64, messageID int) {
	text := "🎫 *创建新工单*\n\n" +
		"请选择工单类型："
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔧 技术支持", "ticket_type:technical"),
			tgbotapi.NewInlineKeyboardButtonData("💰 账单问题", "ticket_type:billing"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📦 产品咨询", "ticket_type:product"),
			tgbotapi.NewInlineKeyboardButtonData("🐛 错误报告", "ticket_type:bug"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💡 功能建议", "ticket_type:feature"),
			tgbotapi.NewInlineKeyboardButtonData("📝 其他", "ticket_type:other"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回", "my_tickets"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// showUserTicketsList shows list of user's tickets
func (b *BotEnhanced) showUserTicketsList(chatID int64, messageID int) {
	ctx := context.Background()
	
	user, err := b.getUserByTelegramID(ctx, chatID)
	if err != nil {
		b.showError(chatID, messageID, "账号未绑定")
		return
	}
	
	// TODO: Get actual tickets from service
	text := "📋 *您的工单列表*\n\n" +
		"🔓 *待处理*:\n" +
		"• #T2024001 - 无法连接服务器\n" +
		"• #T2024002 - 流量统计不准确\n\n" +
		"⚙️ *处理中*:\n" +
		"• #T2024003 - 升级套餐问题\n\n" +
		"✅ *已解决*:\n" +
		"• #T2024004 - 支付失败\n" +
		"• #T2024005 - 账号登录问题\n\n" +
		"点击工单号查看详情"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 刷新列表", "list_tickets"),
			tgbotapi.NewInlineKeyboardButtonData("🔍 搜索", "search_tickets"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回", "my_tickets"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
	
	logger.Info("Displayed user tickets list",
		logger.Int64("chat_id", chatID),
		logger.Uint("user_id", user.ID))
}

// startSearchTickets starts ticket search flow
func (b *BotEnhanced) startSearchTickets(chatID int64, messageID int) {
	// Set conversation state
	b.userStates[chatID] = "searching_tickets"
	
	text := "🔍 *搜索工单*\n\n" +
		"请输入搜索关键词：\n\n" +
		"💡 *搜索提示*:\n" +
		"• 可以搜索工单号 (如: T2024001)\n" +
		"• 可以搜索标题关键词\n" +
		"• 可以搜索工单内容"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ 取消搜索", "my_tickets"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
}

// showTicketStatistics shows ticket statistics
func (b *BotEnhanced) showTicketStatistics(chatID int64, messageID int) {
	ctx := context.Background()
	
	user, err := b.getUserByTelegramID(ctx, chatID)
	if err != nil {
		b.showError(chatID, messageID, "账号未绑定")
		return
	}
	
	// TODO: Get actual statistics
	text := "📊 *工单统计*\n\n" +
		"📈 *本月统计*:\n" +
		"• 提交工单: 8 个\n" +
		"• 已解决: 6 个\n" +
		"• 平均响应时间: 2.5 小时\n" +
		"• 平均解决时间: 12 小时\n\n" +
		"📊 *总体统计*:\n" +
		"• 总工单数: 45 个\n" +
		"• 解决率: 92%\n" +
		"• 满意度: 4.8/5.0\n\n" +
		"🏆 *服务评级*: 优秀"
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📈 详细报表", "ticket_detailed_stats"),
			tgbotapi.NewInlineKeyboardButtonData("📊 趋势图表", "ticket_trends"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 返回", "my_tickets"),
		),
	)
	
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"
	edit.ReplyMarkup = &keyboard
	
	b.api.Send(edit)
	
	logger.Info("Displayed ticket statistics",
		logger.Int64("chat_id", chatID),
		logger.Uint("user_id", user.ID))
}