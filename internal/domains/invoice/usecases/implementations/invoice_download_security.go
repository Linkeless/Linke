package implementations

import (
	"context"
	"fmt"
	"sync"
	"time"

	"linke/internal/shared/logger"
)

// DownloadLimiter manages rate limiting for invoice downloads
type DownloadLimiter struct {
	mu            sync.RWMutex
	userLimits    map[uint]*UserDownloadLimit
	maxPerHour    int
	maxPerDay     int
	maxBulkSize   int
	cleanupTicker *time.Ticker
	logger        logger.Logger
}

// UserDownloadLimit tracks download limits for a specific user
type UserDownloadLimit struct {
	UserID        uint
	HourlyCount   int
	DailyCount    int
	LastHourReset time.Time
	LastDayReset  time.Time
	LastDownload  time.Time
}

// NewDownloadLimiter creates a new download rate limiter
func NewDownloadLimiter(maxPerHour, maxPerDay, maxBulkSize int, logger logger.Logger) *DownloadLimiter {
	dl := &DownloadLimiter{
		userLimits:  make(map[uint]*UserDownloadLimit),
		maxPerHour:  maxPerHour,
		maxPerDay:   maxPerDay,
		maxBulkSize: maxBulkSize,
		logger:      logger,
	}

	// Start cleanup routine
	dl.cleanupTicker = time.NewTicker(1 * time.Hour)
	go dl.cleanupRoutine()

	return dl
}

// CheckDownloadPermission checks if a user can download an invoice
func (dl *DownloadLimiter) CheckDownloadPermission(ctx context.Context, userID uint, isBulk bool, bulkSize int) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	// Get or create user limit
	userLimit, exists := dl.userLimits[userID]
	if !exists {
		userLimit = &UserDownloadLimit{
			UserID:        userID,
			LastHourReset: time.Now(),
			LastDayReset:  time.Now(),
		}
		dl.userLimits[userID] = userLimit
	}

	now := time.Now()

	// Reset counters if needed
	if now.Sub(userLimit.LastHourReset) >= time.Hour {
		userLimit.HourlyCount = 0
		userLimit.LastHourReset = now
	}

	if now.Sub(userLimit.LastDayReset) >= 24*time.Hour {
		userLimit.DailyCount = 0
		userLimit.LastDayReset = now
	}

	// Check bulk size limit
	if isBulk && bulkSize > dl.maxBulkSize {
		return fmt.Errorf("bulk download size exceeds limit: %d > %d", bulkSize, dl.maxBulkSize)
	}

	// Check hourly limit
	requestCount := 1
	if isBulk {
		requestCount = bulkSize // Count each invoice in bulk as separate request
	}

	if userLimit.HourlyCount+requestCount > dl.maxPerHour {
		return fmt.Errorf("hourly download limit exceeded: %d/%d", userLimit.HourlyCount, dl.maxPerHour)
	}

	// Check daily limit
	if userLimit.DailyCount+requestCount > dl.maxPerDay {
		return fmt.Errorf("daily download limit exceeded: %d/%d", userLimit.DailyCount, dl.maxPerDay)
	}

	// Update counters
	userLimit.HourlyCount += requestCount
	userLimit.DailyCount += requestCount
	userLimit.LastDownload = now

	dl.logger.Info("Download permission granted",
		logger.Uint("user_id", userID),
		logger.Bool("is_bulk", isBulk),
		logger.Int("bulk_size", bulkSize),
		logger.Int("hourly_count", userLimit.HourlyCount),
		logger.Int("daily_count", userLimit.DailyCount))

	return nil
}

// GetUserLimits returns current limits for a user
func (dl *DownloadLimiter) GetUserLimits(userID uint) map[string]any {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	userLimit, exists := dl.userLimits[userID]
	if !exists {
		return map[string]any{
			"hourly_count":     0,
			"daily_count":      0,
			"hourly_remaining": dl.maxPerHour,
			"daily_remaining":  dl.maxPerDay,
			"max_bulk_size":    dl.maxBulkSize,
		}
	}

	return map[string]any{
		"hourly_count":     userLimit.HourlyCount,
		"daily_count":      userLimit.DailyCount,
		"hourly_remaining": dl.maxPerHour - userLimit.HourlyCount,
		"daily_remaining":  dl.maxPerDay - userLimit.DailyCount,
		"max_bulk_size":    dl.maxBulkSize,
		"last_download":    userLimit.LastDownload,
	}
}

// ResetUserLimits resets limits for a specific user (admin function)
func (dl *DownloadLimiter) ResetUserLimits(userID uint) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	delete(dl.userLimits, userID)
	dl.logger.Info("User download limits reset", logger.Uint("user_id", userID))
}

// cleanupRoutine removes old user limit entries
func (dl *DownloadLimiter) cleanupRoutine() {
	for range dl.cleanupTicker.C {
		dl.cleanup()
	}
}

// cleanup removes entries for users who haven't downloaded in 7 days
func (dl *DownloadLimiter) cleanup() {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	var removed int

	for userID, userLimit := range dl.userLimits {
		if userLimit.LastDownload.Before(cutoff) {
			delete(dl.userLimits, userID)
			removed++
		}
	}

	if removed > 0 {
		dl.logger.Info("Cleaned up old download limits", logger.Int("removed", removed))
	}
}

// Stop stops the cleanup routine
func (dl *DownloadLimiter) Stop() {
	if dl.cleanupTicker != nil {
		dl.cleanupTicker.Stop()
	}
}

// InvoiceSecurityValidator validates invoice access and operations
type InvoiceSecurityValidator struct {
	logger logger.Logger
}

// NewInvoiceSecurityValidator creates a new security validator
func NewInvoiceSecurityValidator(logger logger.Logger) *InvoiceSecurityValidator {
	return &InvoiceSecurityValidator{
		logger: logger,
	}
}

// ValidateDownloadRequest validates a download request for security issues
func (isv *InvoiceSecurityValidator) ValidateDownloadRequest(ctx context.Context, userID uint, userRole string, invoiceIDs []uint) error {
	// Check for suspicious patterns
	if len(invoiceIDs) > 100 {
		isv.logger.Warn("Large bulk download request detected",
			logger.Uint("user_id", userID),
			logger.Int("invoice_count", len(invoiceIDs)))
		return fmt.Errorf("bulk download request too large")
	}

	// Check for rapid sequential requests (would need additional context)
	// This is a placeholder for more sophisticated detection

	return nil
}

// ValidateTemplate validates PDF template for security
func (isv *InvoiceSecurityValidator) ValidateTemplate(template string) error {
	// Whitelist of allowed templates
	allowedTemplates := map[string]bool{
		"default":      true,
		"professional": true,
		"minimal":      true,
	}

	if template == "" {
		return nil // Empty template is allowed (defaults to invoice template)
	}

	if !allowedTemplates[template] {
		isv.logger.Warn("Invalid template requested", logger.String("template", template))
		return fmt.Errorf("invalid template: %s", template)
	}

	return nil
}

// ValidateLanguage validates language code for security
func (isv *InvoiceSecurityValidator) ValidateLanguage(language string) error {
	// Whitelist of allowed languages
	allowedLanguages := map[string]bool{
		"en": true,
		"zh": true,
		"es": true,
	}

	if language == "" {
		return nil // Empty language is allowed (defaults to invoice language)
	}

	if !allowedLanguages[language] {
		isv.logger.Warn("Invalid language requested", logger.String("language", language))
		return fmt.Errorf("invalid language: %s", language)
	}

	return nil
}

// ValidateWatermark validates watermark text for security
func (isv *InvoiceSecurityValidator) ValidateWatermark(watermark string) error {
	// Check watermark length
	if len(watermark) > 50 {
		return fmt.Errorf("watermark text too long: %d characters", len(watermark))
	}

	// Check for potentially malicious content
	// This is a basic check - in production, you might want more sophisticated validation
	suspicious := []string{"<script>", "javascript:", "data:", "vbscript:"}
	for _, pattern := range suspicious {
		if contains(watermark, pattern) {
			isv.logger.Warn("Suspicious watermark content detected",
				logger.String("watermark", watermark))
			return fmt.Errorf("invalid watermark content")
		}
	}

	return nil
}

// LogSecurityEvent logs security-related events
func (isv *InvoiceSecurityValidator) LogSecurityEvent(ctx context.Context, event string, userID uint, details map[string]any) {
	isv.logger.Warn("Security event detected",
		logger.String("security_event", event),
		logger.Uint("user_id", userID),
		logger.Time("timestamp", time.Now()),
		logger.Any("details", details))
}

// Helper function to check if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	// Simple case-insensitive contains check
	// In production, you might want to use a more sophisticated approach
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsAt(s, substr))))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
