package telegram

import (
	"context"
	"fmt"
	"strings"

	"linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/domains/user/entities"
	userInterfaces "linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/config"
	"linke/internal/shared/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot represents the Telegram bot handler
type Bot struct {
	api                     *tgbotapi.BotAPI
	userService            userInterfaces.UserService
	subscriptionService    interfaces.UserSubscriptionService
	subscriptionPlanService interfaces.SubscriptionPlanService
	cfg                    *config.Config
}

// NewBot creates a new Telegram bot instance
func NewBot(
	cfg *config.Config,
	userService userInterfaces.UserService,
	subscriptionService interfaces.UserSubscriptionService,
	subscriptionPlanService interfaces.SubscriptionPlanService,
) (*Bot, error) {
	if cfg.OAuth2.TelegramBotToken == "" {
		return nil, fmt.Errorf("telegram bot token is not configured")
	}

	bot, err := tgbotapi.NewBotAPI(cfg.OAuth2.TelegramBotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	// bot.Debug = false // Set to true for debugging

	return &Bot{
		api:                     bot,
		userService:            userService,
		subscriptionService:    subscriptionService,
		subscriptionPlanService: subscriptionPlanService,
		cfg:                    cfg,
	}, nil
}

// Start starts the bot and listens for updates
func (b *Bot) Start(ctx context.Context) error {
	logger.Info("Starting Telegram bot", logger.String("username", b.api.Self.UserName))

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping Telegram bot")
			return ctx.Err()
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			go b.handleUpdate(update)
		}
	}
}

// handleUpdate processes incoming updates
func (b *Bot) handleUpdate(update tgbotapi.Update) {
	if update.Message == nil || !update.Message.IsCommand() {
		return
	}

	msg := update.Message
	logger.Info("Received command",
		logger.String("command", msg.Command()),
		logger.String("from", msg.From.UserName),
		logger.Int64("chat_id", msg.Chat.ID))

	switch msg.Command() {
	case "start":
		b.handleStart(msg)
	case "help":
		b.handleHelp(msg)
	case "subscription", "sub":
		b.handleSubscription(msg)
	case "usage":
		b.handleUsage(msg)
	case "plans":
		b.handlePlans(msg)
	default:
		b.sendMessage(msg.Chat.ID, "未知命令。使用 /help 查看可用命令。")
	}
}

// handleStart handles the /start command
func (b *Bot) handleStart(msg *tgbotapi.Message) {
	welcome := fmt.Sprintf(
		"👋 欢迎使用 Linke 服务！\n\n"+
			"我是您的订阅管理助手。\n\n"+
			"使用 /help 查看可用命令。")
	b.sendMessage(msg.Chat.ID, welcome)
}

// handleHelp handles the /help command
func (b *Bot) handleHelp(msg *tgbotapi.Message) {
	help := `📚 *可用命令*

/start - 开始使用机器人
/help - 显示此帮助信息
/subscription - 查看订阅信息
/usage - 查看流量使用情况
/plans - 查看可用套餐

💡 *提示*
• 使用 /sub 作为 /subscription 的简写
• 所有命令都需要先绑定您的账号`

	b.sendMarkdownMessage(msg.Chat.ID, help)
}

// handleSubscription handles the /subscription command
func (b *Bot) handleSubscription(msg *tgbotapi.Message) {
	ctx := context.Background()
	
	// Get user by telegram ID
	user, err := b.getUserByTelegramID(ctx, msg.From.ID)
	if err != nil {
		b.sendMessage(msg.Chat.ID, 
			"❌ 未找到绑定的账号\n\n"+
			"请先在网站上使用 Telegram 登录绑定您的账号。")
		return
	}

	// Get user's active subscriptions
	subscriptions, err := b.subscriptionService.GetUserActiveSubscriptions(ctx, user.ID)
	if err != nil {
		logger.Error("Failed to get user subscriptions",
			logger.Uint("user_id", user.ID),
			logger.Error2("error", err))
		b.sendMessage(msg.Chat.ID, "❌ 获取订阅信息失败，请稍后重试。")
		return
	}

	if len(subscriptions) == 0 {
		b.sendMessage(msg.Chat.ID, 
			"📭 您当前没有活跃的订阅\n\n"+
			"访问我们的网站选择合适的套餐开始使用服务。")
		return
	}

	// Build subscription info message
	var sb strings.Builder
	sb.WriteString("📋 *您的订阅信息*\n\n")

	for i, sub := range subscriptions {
		// Get plan details
		plan, err := b.subscriptionPlanService.GetSubscriptionPlan(ctx, sub.SubscriptionPlanID)
		if err != nil {
			logger.Error("Failed to get subscription plan",
				logger.Uint("plan_id", sub.SubscriptionPlanID),
				logger.Error2("error", err))
			continue
		}

		sb.WriteString(fmt.Sprintf("*套餐 %d: %s*\n", i+1, plan.Name))
		sb.WriteString(fmt.Sprintf("├ 状态: %s\n", b.formatStatus(sub.Status)))
		sb.WriteString(fmt.Sprintf("├ 价格: %.2f %s\n", sub.Price, sub.Currency))
		sb.WriteString(fmt.Sprintf("├ 计费周期: %s\n", b.formatBillingCycle(sub.BillingCycle)))
		
		// Add dates
		sb.WriteString(fmt.Sprintf("├ 开始日期: %s\n", sub.StartDate.Format("2006-01-02")))
		if sub.CurrentPeriodEnd != nil {
			sb.WriteString(fmt.Sprintf("├ 当前周期结束: %s\n", sub.CurrentPeriodEnd.Format("2006-01-02")))
		}
		
		// Add traffic info
		if sub.TrafficLimit > 0 {
			used := b.formatBytes(sub.TrafficUsed)
			total := b.formatBytes(sub.TrafficLimit)
			percentage := float64(sub.TrafficUsed) / float64(sub.TrafficLimit) * 100
			
			sb.WriteString(fmt.Sprintf("├ 流量使用: %s / %s (%.1f%%)\n", used, total, percentage))
			
			if sub.TrafficResetDate != nil {
				sb.WriteString(fmt.Sprintf("├ 流量重置: %s\n", sub.TrafficResetDate.Format("2006-01-02")))
			}
		} else {
			sb.WriteString("├ 流量: 无限制\n")
		}
		
		// Add auto-renew status
		if sub.AutoRenew {
			sb.WriteString("└ 自动续费: ✅ 已开启\n")
		} else {
			sb.WriteString("└ 自动续费: ❌ 已关闭\n")
		}
		
		sb.WriteString("\n")
	}

	b.sendMarkdownMessage(msg.Chat.ID, sb.String())
}

// handleUsage handles the /usage command
func (b *Bot) handleUsage(msg *tgbotapi.Message) {
	ctx := context.Background()
	
	// Get user by telegram ID
	user, err := b.getUserByTelegramID(ctx, msg.From.ID)
	if err != nil {
		b.sendMessage(msg.Chat.ID, 
			"❌ 未找到绑定的账号\n\n"+
			"请先在网站上使用 Telegram 登录绑定您的账号。")
		return
	}

	// Get user's active subscriptions
	subscriptions, err := b.subscriptionService.GetUserActiveSubscriptions(ctx, user.ID)
	if err != nil {
		logger.Error("Failed to get user subscriptions",
			logger.Uint("user_id", user.ID),
			logger.Error2("error", err))
		b.sendMessage(msg.Chat.ID, "❌ 获取使用情况失败，请稍后重试。")
		return
	}

	if len(subscriptions) == 0 {
		b.sendMessage(msg.Chat.ID, "📭 您当前没有活跃的订阅")
		return
	}

	// Build usage info message
	var sb strings.Builder
	sb.WriteString("📊 *流量使用情况*\n\n")

	totalUsed := int64(0)
	totalLimit := int64(0)

	for _, sub := range subscriptions {
		plan, err := b.subscriptionPlanService.GetSubscriptionPlan(ctx, sub.SubscriptionPlanID)
		if err != nil {
			continue
		}

		if sub.TrafficLimit > 0 {
			totalUsed += sub.TrafficUsed
			totalLimit += sub.TrafficLimit
			
			used := b.formatBytes(sub.TrafficUsed)
			total := b.formatBytes(sub.TrafficLimit)
			remaining := b.formatBytes(sub.TrafficLimit - sub.TrafficUsed)
			percentage := float64(sub.TrafficUsed) / float64(sub.TrafficLimit) * 100
			
			sb.WriteString(fmt.Sprintf("*%s*\n", plan.Name))
			sb.WriteString(fmt.Sprintf("├ 已使用: %s\n", used))
			sb.WriteString(fmt.Sprintf("├ 总流量: %s\n", total))
			sb.WriteString(fmt.Sprintf("├ 剩余: %s\n", remaining))
			sb.WriteString(fmt.Sprintf("├ 使用率: %.1f%%\n", percentage))
			sb.WriteString(b.getUsageBar(percentage))
			sb.WriteString("\n\n")
		}
	}

	if totalLimit > 0 {
		totalPercentage := float64(totalUsed) / float64(totalLimit) * 100
		sb.WriteString("*📈 总计*\n")
		sb.WriteString(fmt.Sprintf("已使用: %s / %s (%.1f%%)\n", 
			b.formatBytes(totalUsed), 
			b.formatBytes(totalLimit), 
			totalPercentage))
	}

	b.sendMarkdownMessage(msg.Chat.ID, sb.String())
}

// handlePlans handles the /plans command
func (b *Bot) handlePlans(msg *tgbotapi.Message) {
	ctx := context.Background()
	
	// Get available plans
	plans, _, err := b.subscriptionPlanService.GetSubscriptionPlans(ctx, &interfaces.GetSubscriptionPlansRequest{
		Status: "active",
		Limit:  20,
	})
	if err != nil {
		logger.Error("Failed to get subscription plans", logger.Error2("error", err))
		b.sendMessage(msg.Chat.ID, "❌ 获取套餐信息失败，请稍后重试。")
		return
	}

	if len(plans) == 0 {
		b.sendMessage(msg.Chat.ID, "📭 当前没有可用的套餐")
		return
	}

	// Build plans info message
	var sb strings.Builder
	sb.WriteString("💎 *可用套餐*\n\n")

	for _, plan := range plans {
		if !plan.IsVisible {
			continue
		}

		sb.WriteString(fmt.Sprintf("*%s*\n", plan.Name))
		if plan.Description != "" {
			sb.WriteString(fmt.Sprintf("📝 %s\n", plan.Description))
		}
		sb.WriteString(fmt.Sprintf("💰 价格: %.2f %s/%s\n", 
			plan.Price, 
			plan.Currency,
			b.formatBillingCycleShort(plan.BillingCycle)))
		
		if plan.TrafficLimit > 0 {
			sb.WriteString(fmt.Sprintf("📊 流量: %s/月\n", b.formatBytes(plan.TrafficLimit)))
		} else {
			sb.WriteString("📊 流量: 无限制\n")
		}
		
		if plan.TrialPeriodDays > 0 {
			sb.WriteString(fmt.Sprintf("🎁 试用期: %d 天\n", plan.TrialPeriodDays))
		}
		
		if plan.IsRecommended {
			sb.WriteString("⭐ *推荐套餐*\n")
		}
		
		sb.WriteString("\n")
	}

	sb.WriteString("💡 访问我们的网站购买或升级套餐")

	b.sendMarkdownMessage(msg.Chat.ID, sb.String())
}

// Helper functions

func (b *Bot) getUserByTelegramID(ctx context.Context, telegramID int64) (*entities.User, error) {
	// Convert telegram ID to string
	telegramIDStr := fmt.Sprintf("%d", telegramID)
	
	user, err := b.userService.GetUserByTelegramID(ctx, telegramIDStr)
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		logger.Error("Failed to send telegram message",
			logger.Int64("chat_id", chatID),
			logger.Error2("error", err))
	}
}

func (b *Bot) sendMarkdownMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	if _, err := b.api.Send(msg); err != nil {
		logger.Error("Failed to send telegram markdown message",
			logger.Int64("chat_id", chatID),
			logger.Error2("error", err))
	}
}

func (b *Bot) formatStatus(status string) string {
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

func (b *Bot) formatBillingCycle(cycle string) string {
	switch cycle {
	case "monthly":
		return "月付"
	case "yearly":
		return "年付"
	case "lifetime":
		return "终身"
	default:
		return cycle
	}
}

func (b *Bot) formatBillingCycleShort(cycle string) string {
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

func (b *Bot) formatBytes(bytes int64) string {
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

func (b *Bot) getUsageBar(percentage float64) string {
	const barLength = 10
	filled := int(percentage / 10)
	if filled > barLength {
		filled = barLength
	}
	
	bar := "["
	for i := 0; i < barLength; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	bar += "]"
	
	return bar
}