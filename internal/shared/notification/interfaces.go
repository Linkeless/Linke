package notification

import (
	"context"
	"time"
)

// NotificationChannel represents different notification delivery methods
type NotificationChannel string

const (
	ChannelEmail    NotificationChannel = "email"
	ChannelSMS      NotificationChannel = "sms"
	ChannelPush     NotificationChannel = "push"
	ChannelSlack    NotificationChannel = "slack"
	ChannelTelegram NotificationChannel = "telegram"
)

// NotificationPriority represents the priority level of notifications
type NotificationPriority int

const (
	PriorityLow    NotificationPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityUrgent
)

func (p NotificationPriority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityNormal:
		return "normal"
	case PriorityHigh:
		return "high"
	case PriorityUrgent:
		return "urgent"
	default:
		return "normal"
	}
}

// NotificationRequest represents a request to send a notification
type NotificationRequest struct {
	// Recipient information
	UserID       uint     `json:"user_id"`
	Email        string   `json:"email,omitempty"`
	Phone        string   `json:"phone,omitempty"`
	TelegramChatID int64  `json:"telegram_chat_id,omitempty"`
	TelegramUsername string `json:"telegram_username,omitempty"`
	Channels     []NotificationChannel `json:"channels"`
	
	// Message content
	Subject   string            `json:"subject"`
	Body      string            `json:"body"`
	Template  string            `json:"template,omitempty"`
	Variables map[string]string `json:"variables,omitempty"`
	
	// Metadata and tracking
	Priority  NotificationPriority   `json:"priority"`
	EventType string                 `json:"event_type"`
	EventID   string                 `json:"event_id"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Tags      []string               `json:"tags,omitempty"`
}

// NotificationResult represents the result of sending a notification
type NotificationResult struct {
	Channel   NotificationChannel    `json:"channel"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	MessageID string                 `json:"message_id,omitempty"`
	SentAt    string                 `json:"sent_at,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NotificationService defines the interface for sending notifications
type NotificationService interface {
	// Send a notification to specified channels
	Send(ctx context.Context, req *NotificationRequest) ([]*NotificationResult, error)
	
	// Send email notification
	SendEmail(ctx context.Context, to, subject, body string) error
	
	// Send SMS notification  
	SendSMS(ctx context.Context, to, message string) error
	
	// Send push notification
	SendPush(ctx context.Context, userID uint, title, body string) error
	
	// Send Telegram notification
	SendTelegram(ctx context.Context, chatID int64, message string) error
	SendTelegramByUsername(ctx context.Context, username, message string) error
	
	// Get notification templates
	GetTemplate(templateName string) (string, error)
	
	// Health check
	HealthCheck(ctx context.Context) error
}

// EmailProvider defines the interface for email providers
type EmailProvider interface {
	SendEmail(ctx context.Context, to, subject, body string) error
	HealthCheck(ctx context.Context) error
}

// SMSProvider defines the interface for SMS providers
type SMSProvider interface {
	SendSMS(ctx context.Context, to, message string) error
	HealthCheck(ctx context.Context) error
}

// PushProvider defines the interface for push notification providers
type PushProvider interface {
	SendPush(ctx context.Context, userID uint, title, body string) error
	HealthCheck(ctx context.Context) error
}

// TelegramProvider defines the interface for Telegram notification providers
type TelegramProvider interface {
	SendMessage(ctx context.Context, chatID int64, message string) error
	SendMessageByUsername(ctx context.Context, username, message string) error
	GetChatID(ctx context.Context, username string) (int64, error)
	HealthCheck(ctx context.Context) error
}