package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"linke/internal/shared/logger"
)

// MockEmailProvider provides a mock implementation for development/testing
type MockEmailProvider struct {
	logger logger.Logger
}

// NewMockEmailProvider creates a new mock email provider
func NewMockEmailProvider(logger logger.Logger) *MockEmailProvider {
	return &MockEmailProvider{
		logger: logger,
	}
}

// SendEmail simulates sending an email
func (p *MockEmailProvider) SendEmail(ctx context.Context, to, subject, body string) error {
	p.logger.Info("Mock email sent",
		logger.String("to", to),
		logger.String("subject", subject),
		logger.String("body", body))
	
	// Simulate some processing time
	time.Sleep(10 * time.Millisecond)
	return nil
}

// HealthCheck simulates provider health check
func (p *MockEmailProvider) HealthCheck(ctx context.Context) error {
	p.logger.Debug("Mock email provider health check passed")
	return nil
}

// MockSMSProvider provides a mock implementation for development/testing
type MockSMSProvider struct {
	logger logger.Logger
}

// NewMockSMSProvider creates a new mock SMS provider
func NewMockSMSProvider(logger logger.Logger) *MockSMSProvider {
	return &MockSMSProvider{
		logger: logger,
	}
}

// SendSMS simulates sending an SMS
func (p *MockSMSProvider) SendSMS(ctx context.Context, to, message string) error {
	p.logger.Info("Mock SMS sent",
		logger.String("to", to),
		logger.String("message", message))
	
	// Simulate some processing time
	time.Sleep(15 * time.Millisecond)
	return nil
}

// HealthCheck simulates provider health check
func (p *MockSMSProvider) HealthCheck(ctx context.Context) error {
	p.logger.Debug("Mock SMS provider health check passed")
	return nil
}

// MockPushProvider provides a mock implementation for development/testing
type MockPushProvider struct {
	logger logger.Logger
}

// NewMockPushProvider creates a new mock push provider
func NewMockPushProvider(logger logger.Logger) *MockPushProvider {
	return &MockPushProvider{
		logger: logger,
	}
}

// SendPush simulates sending a push notification
func (p *MockPushProvider) SendPush(ctx context.Context, userID uint, title, body string) error {
	p.logger.Info("Mock push notification sent",
		logger.Uint("user_id", userID),
		logger.String("title", title),
		logger.String("body", body))
	
	// Simulate some processing time
	time.Sleep(20 * time.Millisecond)
	return nil
}

// HealthCheck simulates provider health check
func (p *MockPushProvider) HealthCheck(ctx context.Context) error {
	p.logger.Debug("Mock push provider health check passed")
	return nil
}

// LogOnlyEmailProvider logs emails instead of sending them (useful for development)
type LogOnlyEmailProvider struct {
	logger logger.Logger
}

// NewLogOnlyEmailProvider creates a new log-only email provider
func NewLogOnlyEmailProvider(logger logger.Logger) *LogOnlyEmailProvider {
	return &LogOnlyEmailProvider{
		logger: logger,
	}
}

// SendEmail logs email instead of sending
func (p *LogOnlyEmailProvider) SendEmail(ctx context.Context, to, subject, body string) error {
	p.logger.Info("Email would be sent",
		logger.String("to", to),
		logger.String("subject", subject),
		logger.String("body", body),
		logger.String("provider", "log-only"))
	return nil
}

// HealthCheck always returns healthy
func (p *LogOnlyEmailProvider) HealthCheck(ctx context.Context) error {
	return nil
}

// ConsoleEmailProvider prints emails to console (useful for development)
type ConsoleEmailProvider struct {
	logger logger.Logger
}

// NewConsoleEmailProvider creates a new console email provider
func NewConsoleEmailProvider(logger logger.Logger) *ConsoleEmailProvider {
	return &ConsoleEmailProvider{
		logger: logger,
	}
}

// SendEmail prints email to console
func (p *ConsoleEmailProvider) SendEmail(ctx context.Context, to, subject, body string) error {
	fmt.Printf("\n=== EMAIL NOTIFICATION ===\n")
	fmt.Printf("To: %s\n", to)
	fmt.Printf("Subject: %s\n", subject)
	fmt.Printf("Body:\n%s\n", body)
	fmt.Printf("Sent at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("===========================\n\n")

	p.logger.Info("Email printed to console",
		logger.String("to", to),
		logger.String("subject", subject))
	return nil
}

// HealthCheck always returns healthy
func (p *ConsoleEmailProvider) HealthCheck(ctx context.Context) error {
	return nil
}

// MockTelegramProvider provides a mock implementation for development/testing
type MockTelegramProvider struct {
	logger logger.Logger
}

// NewMockTelegramProvider creates a new mock Telegram provider
func NewMockTelegramProvider(logger logger.Logger) *MockTelegramProvider {
	return &MockTelegramProvider{
		logger: logger,
	}
}

// SendMessage simulates sending a Telegram message
func (p *MockTelegramProvider) SendMessage(ctx context.Context, chatID int64, message string) error {
	p.logger.Info("Mock Telegram message sent",
		logger.Int64("chat_id", chatID),
		logger.String("message", message))
	
	// Simulate some processing time
	time.Sleep(25 * time.Millisecond)
	return nil
}

// SendMessageByUsername simulates sending a message by username
func (p *MockTelegramProvider) SendMessageByUsername(ctx context.Context, username, message string) error {
	p.logger.Info("Mock Telegram message sent by username",
		logger.String("username", username),
		logger.String("message", message))
	
	time.Sleep(25 * time.Millisecond)
	return nil
}

// GetChatID simulates getting chat ID by username
func (p *MockTelegramProvider) GetChatID(ctx context.Context, username string) (int64, error) {
	p.logger.Info("Mock getting Telegram chat ID",
		logger.String("username", username))
	
	// Return a mock chat ID based on username hash
	chatID := int64(12345678) // Mock chat ID
	return chatID, nil
}

// HealthCheck simulates provider health check
func (p *MockTelegramProvider) HealthCheck(ctx context.Context) error {
	p.logger.Debug("Mock Telegram provider health check passed")
	return nil
}

// TelegramBotProvider provides real Telegram Bot API integration
type TelegramBotProvider struct {
	botToken   string
	httpClient *http.Client
	logger     logger.Logger
	baseURL    string
}

// TelegramMessage represents a Telegram message payload
type TelegramMessage struct {
	ChatID    int64  `json:"chat_id,omitempty"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// TelegramResponse represents a Telegram API response
type TelegramResponse struct {
	OK          bool   `json:"ok"`
	Result      interface{} `json:"result,omitempty"`
	Description string `json:"description,omitempty"`
	ErrorCode   int    `json:"error_code,omitempty"`
}

// TelegramChat represents a Telegram chat
type TelegramChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
	Username string `json:"username,omitempty"`
}

// NewTelegramBotProvider creates a new real Telegram Bot provider
func NewTelegramBotProvider(botToken string, logger logger.Logger) *TelegramBotProvider {
	return &TelegramBotProvider{
		botToken: botToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:  logger,
		baseURL: "https://api.telegram.org/bot",
	}
}

// SendMessage sends a message to a specific chat ID
func (p *TelegramBotProvider) SendMessage(ctx context.Context, chatID int64, message string) error {
	if p.botToken == "" {
		return fmt.Errorf("telegram bot token not configured")
	}

	// Prepare message payload
	telegramMsg := TelegramMessage{
		ChatID:    chatID,
		Text:      message,
		ParseMode: "HTML", // Support HTML formatting
	}

	// Marshal to JSON
	payload, err := json.Marshal(telegramMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram message: %w", err)
	}

	// Make API request
	url := fmt.Sprintf("%s%s/sendMessage", p.baseURL, p.botToken)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create telegram request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var telegramResp TelegramResponse
	if err := json.NewDecoder(resp.Body).Decode(&telegramResp); err != nil {
		return fmt.Errorf("failed to decode telegram response: %w", err)
	}

	if !telegramResp.OK {
		return fmt.Errorf("telegram API error [%d]: %s", telegramResp.ErrorCode, telegramResp.Description)
	}

	p.logger.Info("Telegram message sent successfully",
		logger.Int64("chat_id", chatID),
		logger.String("message_preview", message[:min(50, len(message))]))

	return nil
}

// SendMessageByUsername attempts to send a message by username (requires prior chat)
func (p *TelegramBotProvider) SendMessageByUsername(ctx context.Context, username, message string) error {
	// Note: Telegram Bot API doesn't support sending messages directly by username
	// without having the chat_id. This would require a database mapping of username -> chat_id
	// or using the updates mechanism to track user interactions
	
	p.logger.Warn("SendMessageByUsername called but not fully supported by Telegram Bot API",
		logger.String("username", username),
		logger.String("suggestion", "Use GetChatID first or maintain a username->chat_id mapping"))

	return fmt.Errorf("sending by username requires prior chat interaction to obtain chat_id")
}

// GetChatID attempts to resolve username to chat ID (limited functionality)
func (p *TelegramBotProvider) GetChatID(ctx context.Context, username string) (int64, error) {
	// Note: Telegram Bot API doesn't provide a direct way to resolve username to chat_id
	// This would require maintaining a database mapping from previous interactions
	// or using webhook updates to track user messages
	
	p.logger.Warn("GetChatID called but not directly supported by Telegram Bot API",
		logger.String("username", username),
		logger.String("suggestion", "Maintain username->chat_id mapping from user interactions"))

	return 0, fmt.Errorf("resolving username to chat_id requires prior user interaction tracking")
}

// HealthCheck verifies the Telegram Bot API connection
func (p *TelegramBotProvider) HealthCheck(ctx context.Context) error {
	if p.botToken == "" {
		return fmt.Errorf("telegram bot token not configured")
	}

	// Call getMe API to verify bot token
	url := fmt.Sprintf("%s%s/getMe", p.baseURL, p.botToken)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create telegram health check request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram health check request failed: %w", err)
	}
	defer resp.Body.Close()

	var telegramResp TelegramResponse
	if err := json.NewDecoder(resp.Body).Decode(&telegramResp); err != nil {
		return fmt.Errorf("failed to decode telegram health check response: %w", err)
	}

	if !telegramResp.OK {
		return fmt.Errorf("telegram bot token invalid: %s", telegramResp.Description)
	}

	p.logger.Debug("Telegram provider health check passed")
	return nil
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}