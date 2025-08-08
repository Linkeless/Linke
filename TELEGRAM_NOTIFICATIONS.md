# Telegram Notifications Integration

This document describes how to configure and use Telegram notifications in the Linke platform.

## 🚀 Features

- ✅ **Real Telegram Bot Integration** - Uses official Telegram Bot API
- ✅ **Rich HTML Formatting** - Support for bold, italic, code blocks, and emojis
- ✅ **Mock Mode** - Development-friendly mock provider for testing
- ✅ **Template System** - Pre-built templates for common notifications
- ✅ **Health Checks** - Built-in provider health monitoring
- ✅ **Dependency Injection** - Seamless Fx integration
- ✅ **Error Handling** - Robust error handling and logging

## ⚙️ Configuration

### 1. Environment Variables

Add your Telegram bot token to your environment:

```bash
# .env file
TELEGRAM_BOT_TOKEN=your_bot_token_here
```

### 2. Creating a Telegram Bot

1. Message [@BotFather](https://t.me/BotFather) on Telegram
2. Use `/newbot` command to create a new bot
3. Follow the instructions to set bot name and username
4. Copy the bot token provided by BotFather
5. Set the token in your environment variables

### 3. Getting Chat IDs

To send messages, you need the user's chat ID. Users must first start a conversation with your bot:

1. User searches for your bot by username
2. User starts conversation with `/start` command
3. Your application can then send notifications to that user

## 📋 Usage Examples

### Basic Notification Service

```go
// The notification service is automatically injected via Fx
func SomeHandler(notificationService notification.NotificationService) {
    ctx := context.Background()
    
    // Send a simple Telegram message
    err := notificationService.SendTelegram(ctx, chatID, "Hello from Linke!")
    if err != nil {
        log.Error("Failed to send Telegram message", err)
    }
}
```

### Using Templates

```go
func SendPaymentNotification(notificationService notification.NotificationService, userID uint, chatID int64) {
    ctx := context.Background()
    
    req := &notification.NotificationRequest{
        UserID:         userID,
        TelegramChatID: chatID,
        Channels:       []notification.NotificationChannel{notification.ChannelTelegram},
        Subject:        "Payment Confirmation",
        Template:       "telegram_payment_completed",
        Variables: map[string]string{
            "user_name":  "John Doe",
            "amount":     "29.99",
            "currency":   "USD",
            "order_id":   "ORD-123456",
            "date":       time.Now().Format("2006-01-02 15:04:05"),
        },
    }
    
    results, err := notificationService.Send(ctx, req)
    if err != nil {
        log.Error("Failed to send notification", err)
        return
    }
    
    // Check results
    for _, result := range results {
        if result.Success {
            log.Info("Notification sent successfully", "channel", result.Channel, "message_id", result.MessageID)
        } else {
            log.Error("Notification failed", "channel", result.Channel, "error", result.Error)
        }
    }
}
```

### Multi-Channel Notifications

```go
func SendMultiChannelNotification(notificationService notification.NotificationService) {
    ctx := context.Background()
    
    req := &notification.NotificationRequest{
        UserID:           123,
        Email:           "user@example.com",
        TelegramChatID:  456789012,
        Channels: []notification.NotificationChannel{
            notification.ChannelEmail,
            notification.ChannelTelegram,
        },
        Subject:  "Welcome to our platform!",
        Template: "user_created", // Uses email template for email, will use telegram_user_created if available
        Variables: map[string]string{
            "user_name": "Alice",
            "username":  "alice123",
            "email":     "alice@example.com",
            "join_date": time.Now().Format("2006-01-02"),
        },
    }
    
    results, err := notificationService.Send(ctx, req)
    // Handle results...
}
```

## 📝 Available Templates

### Standard Templates
- `payment_completed` - Payment confirmation
- `subscription_expired` - Subscription expiration notice
- `invoice_overdue` - Overdue invoice reminder
- `order_paid` - Order confirmation
- `user_created` - Welcome message for new users
- `subscription_activated` - Subscription activation notice

### Telegram-Specific Templates (with rich formatting)
- `telegram_payment_completed` - 🎉 Rich payment confirmation
- `telegram_subscription_expired` - ⚠️ Subscription expiration with emojis
- `telegram_invoice_overdue` - 🔴 Urgent overdue notice
- `telegram_order_paid` - ✅ Order confirmation with formatting
- `telegram_user_created` - 🎊 Welcome message with account details
- `telegram_subscription_activated` - 🔥 Subscription activation celebration
- `telegram_ticket_created` - 🎫 Support ticket created notice
- `telegram_ticket_resolved` - ✅ Ticket resolution notice

## 🔧 Development Mode

When no `TELEGRAM_BOT_TOKEN` is provided, the system automatically uses mock providers for development:

```bash
# .env (development)
# TELEGRAM_BOT_TOKEN=  # Leave empty or comment out

# The application will log:
# INFO: Using mock Telegram provider (no bot token configured)
```

Mock providers log all notification attempts without making real API calls.

## 📊 Integration with Business Events

The notification system integrates seamlessly with the event-driven architecture:

```go
// Example: Listening for subscription events
func HandleSubscriptionActivated(event events.SubscriptionActivated) {
    // Get user's Telegram chat ID from database
    chatID := getUserTelegramChatID(event.UserID)
    if chatID == 0 {
        return // User hasn't connected Telegram
    }
    
    req := &notification.NotificationRequest{
        UserID:         event.UserID,
        TelegramChatID: chatID,
        Channels:       []notification.NotificationChannel{notification.ChannelTelegram},
        Template:       "telegram_subscription_activated",
        Variables: map[string]string{
            "user_name":   event.UserName,
            "plan_name":   event.PlanName,
            "amount":      fmt.Sprintf("%.2f", event.Amount),
            "currency":    event.Currency,
            "expiry_date": event.ExpiryDate.Format("2006-01-02"),
            "data_limit":  event.DataLimit,
        },
    }
    
    notificationService.Send(context.Background(), req)
}
```

## ⚕️ Health Monitoring

The notification service includes built-in health checks:

```go
func CheckNotificationHealth(notificationService notification.NotificationService) {
    ctx := context.Background()
    
    if err := notificationService.HealthCheck(ctx); err != nil {
        log.Error("Notification service health check failed", err)
        // Handle service degradation
    } else {
        log.Info("All notification providers are healthy")
    }
}
```

## 🚨 Error Handling

The system provides comprehensive error handling:

```go
// Provider-level errors
err := notificationService.SendTelegram(ctx, chatID, message)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "telegram bot token not configured"):
        // Handle missing configuration
    case strings.Contains(err.Error(), "telegram API error"):
        // Handle API errors (invalid chat_id, bot blocked, etc.)
    case strings.Contains(err.Error(), "failed to send telegram message"):
        // Handle network/HTTP errors
    default:
        // Handle other errors
    }
}
```

## 🔐 Security Considerations

1. **Bot Token Security**: Store bot tokens securely and never commit to version control
2. **Chat ID Validation**: Verify chat IDs belong to authenticated users
3. **Rate Limiting**: Telegram has rate limits - implement proper queuing for high-volume scenarios
4. **User Privacy**: Only send notifications to users who have opted in
5. **Content Filtering**: Sanitize user-generated content in notifications

## 📱 User Experience Flow

1. **User Registration**: User signs up on your platform
2. **Telegram Linking**: User is instructed to start conversation with your bot
3. **Chat ID Capture**: Your bot receives `/start` command and stores user's chat ID
4. **Notification Delivery**: System can now send notifications to user's Telegram

## 🚧 Limitations & Future Enhancements

### Current Limitations
- Username-based messaging requires prior chat interaction
- No support for Telegram channels/groups (only private messages)
- Basic template variable substitution

### Future Enhancements
- Telegram channels/groups support
- Interactive buttons and inline keyboards
- Advanced template engine with conditionals
- Message scheduling and delayed delivery
- Telegram-specific analytics and metrics

## 🛠️ Troubleshooting

### Common Issues

1. **"telegram bot token not configured"**
   - Solution: Set `TELEGRAM_BOT_TOKEN` environment variable

2. **"telegram API error [403]: Forbidden: bot was blocked by the user"**
   - Solution: User has blocked your bot - handle gracefully and remove chat ID

3. **"telegram API error [400]: Bad Request: chat not found"**
   - Solution: Invalid chat ID - verify the chat ID exists and is correct

4. **"failed to send telegram message: context deadline exceeded"**
   - Solution: Network timeout - check connectivity and increase timeout if needed

### Debug Logging

Enable debug logging to troubleshoot issues:

```bash
LOG_LEVEL=debug
```

This will show detailed logs of all notification attempts and API interactions.

## 📚 Related Documentation

- [Event System Documentation](./EVENTS.md)
- [Caching Best Practices](./CACHING_BEST_PRACTICES.md)
- [Architecture Overview](./CLAUDE.md)
- [Telegram Bot API Documentation](https://core.telegram.org/bots/api)