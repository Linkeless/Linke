package notification

import (
	"context"
	"fmt"
	"strings"
	"time"

	"linke/internal/shared/logger"
)

// BaseNotificationService provides a basic implementation of NotificationService
type BaseNotificationService struct {
	emailProvider    EmailProvider
	smsProvider      SMSProvider
	pushProvider     PushProvider
	telegramProvider TelegramProvider
	templates        map[string]string
	logger           logger.Logger
}

// NewBaseNotificationService creates a new base notification service
func NewBaseNotificationService(logger logger.Logger) *BaseNotificationService {
	return &BaseNotificationService{
		logger: logger,
		templates: map[string]string{
			// Email/SMS templates
			"payment_completed": "Dear {{user_name}}, your payment of {{amount}} {{currency}} has been successfully processed. Order ID: {{order_id}}.",
			"subscription_expired": "Dear {{user_name}}, your subscription has expired. Please renew to continue using our services.",
			"invoice_overdue": "Dear {{user_name}}, your invoice {{invoice_id}} is now overdue. Please make payment to avoid service interruption.",
			"order_paid": "Dear {{user_name}}, thank you for your order! Your order {{order_id}} has been confirmed and is being processed.",
			"user_created": "Welcome {{user_name}}! Your account has been successfully created. Start exploring our services now.",
			"subscription_activated": "Great news {{user_name}}! Your {{plan_name}} subscription is now active and ready to use.",
			
			// Telegram HTML templates with rich formatting
			"telegram_payment_completed": "🎉 <b>Payment Successful!</b>\n\nHi {{user_name}},\n\nYour payment has been processed successfully!\n\n💰 <b>Amount:</b> {{amount}} {{currency}}\n📋 <b>Order ID:</b> <code>{{order_id}}</code>\n📅 <b>Date:</b> {{date}}\n\nThank you for your business! 🙏",
			
			"telegram_subscription_expired": "⚠️ <b>Subscription Expired</b>\n\nHi {{user_name}},\n\nYour subscription has expired and needs renewal.\n\n📦 <b>Plan:</b> {{plan_name}}\n📅 <b>Expired:</b> {{expiry_date}}\n\n<i>Renew now to continue enjoying our services!</i> ⚡",
			
			"telegram_invoice_overdue": "🔴 <b>Invoice Overdue</b>\n\nHi {{user_name}},\n\nYour invoice is now overdue and requires immediate attention.\n\n🧾 <b>Invoice ID:</b> <code>{{invoice_id}}</code>\n💵 <b>Amount Due:</b> {{amount}} {{currency}}\n📅 <b>Due Date:</b> {{due_date}}\n\n<b>Please make payment to avoid service interruption.</b>",
			
			"telegram_order_paid": "✅ <b>Order Confirmed!</b>\n\nHi {{user_name}},\n\nThank you for your order! We're processing it now.\n\n📋 <b>Order ID:</b> <code>{{order_id}}</code>\n💰 <b>Amount:</b> {{amount}} {{currency}}\n📦 <b>Items:</b> {{item_count}} items\n\n<i>You'll receive updates as we process your order.</i> 📱",
			
			"telegram_user_created": "🎊 <b>Welcome to our platform!</b>\n\nHi {{user_name}}!\n\nYour account has been successfully created. You're now part of our community!\n\n👤 <b>Username:</b> {{username}}\n📧 <b>Email:</b> {{email}}\n📅 <b>Joined:</b> {{join_date}}\n\n<i>Start exploring our services and enjoy the experience!</i> 🚀",
			
			"telegram_subscription_activated": "🔥 <b>Subscription Activated!</b>\n\nHi {{user_name}}!\n\nGreat news! Your subscription is now active and ready to use.\n\n📦 <b>Plan:</b> {{plan_name}}\n💰 <b>Price:</b> {{amount}} {{currency}}\n📅 <b>Valid Until:</b> {{expiry_date}}\n💾 <b>Data Limit:</b> {{data_limit}}\n\n<b>Enjoy your premium experience!</b> ⭐",
			
			"telegram_ticket_created": "🎫 <b>Support Ticket Created</b>\n\nHi {{user_name}},\n\nYour support ticket has been created successfully.\n\n🎟️ <b>Ticket ID:</b> <code>{{ticket_id}}</code>\n📝 <b>Subject:</b> {{subject}}\n⚡ <b>Priority:</b> {{priority}}\n📊 <b>Status:</b> {{status}}\n\n<i>We'll get back to you soon!</i> 💪",
			
			"telegram_ticket_resolved": "✅ <b>Ticket Resolved</b>\n\nHi {{user_name}},\n\nYour support ticket has been resolved!\n\n🎟️ <b>Ticket ID:</b> <code>{{ticket_id}}</code>\n👨‍💻 <b>Resolved by:</b> {{resolved_by}}\n📝 <b>Resolution:</b>\n<i>{{resolution}}</i>\n\n<b>Thanks for your patience!</b> 🎉",
		},
	}
}

// SetEmailProvider sets the email provider
func (s *BaseNotificationService) SetEmailProvider(provider EmailProvider) {
	s.emailProvider = provider
}

// SetSMSProvider sets the SMS provider
func (s *BaseNotificationService) SetSMSProvider(provider SMSProvider) {
	s.smsProvider = provider
}

// SetPushProvider sets the push notification provider
func (s *BaseNotificationService) SetPushProvider(provider PushProvider) {
	s.pushProvider = provider
}

// SetTelegramProvider sets the Telegram notification provider
func (s *BaseNotificationService) SetTelegramProvider(provider TelegramProvider) {
	s.telegramProvider = provider
}

// Send sends notifications through multiple channels
func (s *BaseNotificationService) Send(ctx context.Context, req *NotificationRequest) ([]*NotificationResult, error) {
	var results []*NotificationResult

	s.logger.Info("Sending notification",
		logger.Uint("user_id", req.UserID),
		logger.String("event_type", req.EventType),
		logger.String("subject", req.Subject),
		logger.Int("channels_count", len(req.Channels)))

	// Process message content
	content := s.processTemplate(req.Body, req.Variables)
	if req.Template != "" {
		if template, err := s.GetTemplate(req.Template); err == nil {
			content = s.processTemplate(template, req.Variables)
		}
	}

	// Send through each requested channel
	for _, channel := range req.Channels {
		result := &NotificationResult{
			Channel: channel,
			SentAt:  time.Now().Format(time.RFC3339),
		}

		switch channel {
		case ChannelEmail:
			if req.Email == "" {
				result.Success = false
				result.Error = "email address not provided"
			} else {
				err := s.SendEmail(ctx, req.Email, req.Subject, content)
				result.Success = err == nil
				if err != nil {
					result.Error = err.Error()
				} else {
					result.MessageID = fmt.Sprintf("email_%d_%s", req.UserID, time.Now().Format("20060102150405"))
				}
			}

		case ChannelSMS:
			if req.Phone == "" {
				result.Success = false
				result.Error = "phone number not provided"
			} else {
				// For SMS, combine subject and body
				smsMessage := fmt.Sprintf("%s: %s", req.Subject, content)
				err := s.SendSMS(ctx, req.Phone, smsMessage)
				result.Success = err == nil
				if err != nil {
					result.Error = err.Error()
				} else {
					result.MessageID = fmt.Sprintf("sms_%d_%s", req.UserID, time.Now().Format("20060102150405"))
				}
			}

		case ChannelPush:
			err := s.SendPush(ctx, req.UserID, req.Subject, content)
			result.Success = err == nil
			if err != nil {
				result.Error = err.Error()
			} else {
				result.MessageID = fmt.Sprintf("push_%d_%s", req.UserID, time.Now().Format("20060102150405"))
			}

		case ChannelTelegram:
			// Try to send using chat ID first, then fallback to username
			var err error
			if req.TelegramChatID != 0 {
				telegramMessage := fmt.Sprintf("<b>%s</b>\n\n%s", req.Subject, content)
				err = s.SendTelegram(ctx, req.TelegramChatID, telegramMessage)
			} else if req.TelegramUsername != "" {
				telegramMessage := fmt.Sprintf("%s: %s", req.Subject, content)
				err = s.SendTelegramByUsername(ctx, req.TelegramUsername, telegramMessage)
			} else {
				err = fmt.Errorf("telegram chat ID or username not provided")
			}
			
			result.Success = err == nil
			if err != nil {
				result.Error = err.Error()
			} else {
				result.MessageID = fmt.Sprintf("telegram_%d_%s", req.UserID, time.Now().Format("20060102150405"))
			}

		default:
			result.Success = false
			result.Error = fmt.Sprintf("unsupported channel: %s", channel)
		}

		results = append(results, result)
	}

	// Log results
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	s.logger.Info("Notification sending completed",
		logger.Uint("user_id", req.UserID),
		logger.String("event_type", req.EventType),
		logger.Int("total_channels", len(req.Channels)),
		logger.Int("successful_sends", successCount))

	return results, nil
}

// SendEmail sends an email notification
func (s *BaseNotificationService) SendEmail(ctx context.Context, to, subject, body string) error {
	if s.emailProvider == nil {
		s.logger.Warn("Email provider not configured, skipping email notification",
			logger.String("to", to),
			logger.String("subject", subject))
		return fmt.Errorf("email provider not configured")
	}

	return s.emailProvider.SendEmail(ctx, to, subject, body)
}

// SendSMS sends an SMS notification
func (s *BaseNotificationService) SendSMS(ctx context.Context, to, message string) error {
	if s.smsProvider == nil {
		s.logger.Warn("SMS provider not configured, skipping SMS notification",
			logger.String("to", to),
			logger.String("message", message))
		return fmt.Errorf("SMS provider not configured")
	}

	return s.smsProvider.SendSMS(ctx, to, message)
}

// SendPush sends a push notification
func (s *BaseNotificationService) SendPush(ctx context.Context, userID uint, title, body string) error {
	if s.pushProvider == nil {
		s.logger.Warn("Push provider not configured, skipping push notification",
			logger.Uint("user_id", userID),
			logger.String("title", title))
		return fmt.Errorf("push provider not configured")
	}

	return s.pushProvider.SendPush(ctx, userID, title, body)
}

// SendTelegram sends a Telegram notification
func (s *BaseNotificationService) SendTelegram(ctx context.Context, chatID int64, message string) error {
	if s.telegramProvider == nil {
		preview := message
		if len(message) > 50 {
			preview = message[:50] + "..."
		}
		s.logger.Warn("Telegram provider not configured, skipping telegram notification",
			logger.Int64("chat_id", chatID),
			logger.String("message_preview", preview))
		return fmt.Errorf("telegram provider not configured")
	}

	return s.telegramProvider.SendMessage(ctx, chatID, message)
}

// SendTelegramByUsername sends a Telegram notification by username
func (s *BaseNotificationService) SendTelegramByUsername(ctx context.Context, username, message string) error {
	if s.telegramProvider == nil {
		preview := message
		if len(message) > 50 {
			preview = message[:50] + "..."
		}
		s.logger.Warn("Telegram provider not configured, skipping telegram notification",
			logger.String("username", username),
			logger.String("message_preview", preview))
		return fmt.Errorf("telegram provider not configured")
	}

	return s.telegramProvider.SendMessageByUsername(ctx, username, message)
}

// GetTemplate retrieves a notification template
func (s *BaseNotificationService) GetTemplate(templateName string) (string, error) {
	template, exists := s.templates[templateName]
	if !exists {
		return "", fmt.Errorf("template not found: %s", templateName)
	}
	return template, nil
}

// processTemplate processes template variables
func (s *BaseNotificationService) processTemplate(template string, variables map[string]string) string {
	content := template
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		content = strings.ReplaceAll(content, placeholder, value)
	}
	return content
}

// HealthCheck checks the health of all notification providers
func (s *BaseNotificationService) HealthCheck(ctx context.Context) error {
	var errors []string

	// Check email provider
	if s.emailProvider != nil {
		if err := s.emailProvider.HealthCheck(ctx); err != nil {
			errors = append(errors, fmt.Sprintf("email provider: %s", err.Error()))
		}
	}

	// Check SMS provider
	if s.smsProvider != nil {
		if err := s.smsProvider.HealthCheck(ctx); err != nil {
			errors = append(errors, fmt.Sprintf("SMS provider: %s", err.Error()))
		}
	}

	// Check push provider
	if s.pushProvider != nil {
		if err := s.pushProvider.HealthCheck(ctx); err != nil {
			errors = append(errors, fmt.Sprintf("push provider: %s", err.Error()))
		}
	}

	// Check Telegram provider
	if s.telegramProvider != nil {
		if err := s.telegramProvider.HealthCheck(ctx); err != nil {
			errors = append(errors, fmt.Sprintf("telegram provider: %s", err.Error()))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("notification service health check failed: %s", strings.Join(errors, ", "))
	}

	return nil
}