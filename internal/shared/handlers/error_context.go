package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"linke/internal/shared/logger"
)

// OperationContext provides structured context for operations and error logging
type OperationContext struct {
	Operation    string         `json:"operation"`
	UserID       uint           `json:"user_id"`
	AdminID      *uint          `json:"admin_id,omitempty"`
	TicketID     *uint          `json:"ticket_id,omitempty"`
	MessageID    *uint          `json:"message_id,omitempty"`
	RequestID    string         `json:"request_id,omitempty"`
	UserAgent    string         `json:"user_agent,omitempty"`
	RemoteAddr   string         `json:"remote_addr,omitempty"`
	StartTime    time.Time      `json:"start_time"`
	RequestSize  int            `json:"request_size,omitempty"`
	ResponseSize int            `json:"response_size,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// NewOperationContext creates a new operation context from gin.Context
func NewOperationContext(c *gin.Context, operation string) *OperationContext {
	ctx := &OperationContext{
		Operation:  operation,
		StartTime:  time.Now(),
		RemoteAddr: c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		Metadata:   make(map[string]any),
	}

	// Extract request ID if available
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		ctx.RequestID = requestID
	}

	// Extract content length
	if c.Request.ContentLength > 0 {
		ctx.RequestSize = int(c.Request.ContentLength)
	}

	return ctx
}

// WithUserID sets the user ID in the context
func (ctx *OperationContext) WithUserID(userID uint) *OperationContext {
	ctx.UserID = userID
	return ctx
}

// WithAdminID sets the admin ID in the context
func (ctx *OperationContext) WithAdminID(adminID uint) *OperationContext {
	ctx.AdminID = &adminID
	return ctx
}

// WithTicketID sets the ticket ID in the context
func (ctx *OperationContext) WithTicketID(ticketID uint) *OperationContext {
	ctx.TicketID = &ticketID
	return ctx
}

// WithMessageID sets the message ID in the context
func (ctx *OperationContext) WithMessageID(messageID uint) *OperationContext {
	ctx.MessageID = &messageID
	return ctx
}

// WithMetadata adds metadata to the context
func (ctx *OperationContext) WithMetadata(key string, value any) *OperationContext {
	if ctx.Metadata == nil {
		ctx.Metadata = make(map[string]any)
	}
	ctx.Metadata[key] = value
	return ctx
}

// WithResponseSize sets the response size
func (ctx *OperationContext) WithResponseSize(size int) *OperationContext {
	ctx.ResponseSize = size
	return ctx
}

// Duration returns the elapsed time since the operation started
func (ctx *OperationContext) Duration() time.Duration {
	return time.Since(ctx.StartTime)
}

// toLogFields converts the context to logger fields
func (ctx *OperationContext) toLogFields() []zap.Field {
	fields := []zap.Field{
		logger.String("operation", ctx.Operation),
		logger.Uint("user_id", ctx.UserID),
		logger.Duration("duration", ctx.Duration()),
		logger.String("remote_addr", ctx.RemoteAddr),
	}

	if ctx.AdminID != nil {
		fields = append(fields, logger.Uint("admin_id", *ctx.AdminID))
	}

	if ctx.TicketID != nil {
		fields = append(fields, logger.Uint("ticket_id", *ctx.TicketID))
	}

	if ctx.MessageID != nil {
		fields = append(fields, logger.Uint("message_id", *ctx.MessageID))
	}

	if ctx.RequestID != "" {
		fields = append(fields, logger.String("request_id", ctx.RequestID))
	}

	if ctx.UserAgent != "" {
		fields = append(fields, logger.String("user_agent", ctx.UserAgent))
	}

	if ctx.RequestSize > 0 {
		fields = append(fields, logger.Int("request_size", ctx.RequestSize))
	}

	if ctx.ResponseSize > 0 {
		fields = append(fields, logger.Int("response_size", ctx.ResponseSize))
	}

	// Add metadata fields
	for key, value := range ctx.Metadata {
		fields = append(fields, logger.Any(key, value))
	}

	return fields
}

// LogError logs an error with full context information
func (ctx *OperationContext) LogError(err error, message string, additionalFields ...zap.Field) {
	fields := ctx.toLogFields()
	fields = append(fields, logger.ErrorField(err))
	fields = append(fields, additionalFields...)

	logger.Error(message, fields...)
}

// LogWarn logs a warning with context information
func (ctx *OperationContext) LogWarn(message string, additionalFields ...zap.Field) {
	fields := ctx.toLogFields()
	fields = append(fields, additionalFields...)

	logger.Warn(message, fields...)
}

// LogInfo logs info with context information
func (ctx *OperationContext) LogInfo(message string, additionalFields ...zap.Field) {
	fields := ctx.toLogFields()
	fields = append(fields, additionalFields...)

	logger.Info(message, fields...)
}

// LogDebug logs debug information with context
func (ctx *OperationContext) LogDebug(message string, additionalFields ...zap.Field) {
	fields := ctx.toLogFields()
	fields = append(fields, additionalFields...)

	logger.Debug(message, fields...)
}

// LogSuccess logs successful operation completion
func (ctx *OperationContext) LogSuccess(message string, additionalFields ...zap.Field) {
	fields := ctx.toLogFields()
	fields = append(fields, logger.String("status", "success"))
	fields = append(fields, additionalFields...)

	logger.Info(message, fields...)
}

// LogPerformance logs performance metrics
func (ctx *OperationContext) LogPerformance(additionalFields ...zap.Field) {
	fields := ctx.toLogFields()
	fields = append(fields, logger.String("log_type", "performance"))
	fields = append(fields, additionalFields...)

	duration := ctx.Duration()
	if duration > 1*time.Second {
		logger.Warn("Slow operation detected", fields...)
	} else {
		logger.Debug("Operation performance", fields...)
	}
}

// SecurityContext provides additional security-related logging
type SecurityContext struct {
	*OperationContext
	SecurityEvent string `json:"security_event"`
	RiskLevel     string `json:"risk_level"`
	IPAddress     string `json:"ip_address"`
}

// NewSecurityContext creates a security context for security-related operations
func NewSecurityContext(ctx *OperationContext, event, riskLevel string) *SecurityContext {
	return &SecurityContext{
		OperationContext: ctx,
		SecurityEvent:    event,
		RiskLevel:        riskLevel,
		IPAddress:        ctx.RemoteAddr,
	}
}

// LogSecurityEvent logs security-related events
func (ctx *SecurityContext) LogSecurityEvent(message string, additionalFields ...zap.Field) {
	fields := ctx.toLogFields()
	fields = append(fields,
		logger.String("security_event", ctx.SecurityEvent),
		logger.String("risk_level", ctx.RiskLevel),
		logger.String("ip_address", ctx.IPAddress),
		logger.String("log_type", "security"),
	)
	fields = append(fields, additionalFields...)

	switch ctx.RiskLevel {
	case "high", "critical":
		logger.Error(message, fields...)
	case "medium":
		logger.Warn(message, fields...)
	default:
		logger.Info(message, fields...)
	}
}

// PerformanceThresholds defines performance thresholds for different operations
type PerformanceThresholds struct {
	Warning  time.Duration `json:"warning"`
	Critical time.Duration `json:"critical"`
}

// DefaultThresholds provides default performance thresholds
var DefaultThresholds = map[string]PerformanceThresholds{
	"create_ticket":    {Warning: 500 * time.Millisecond, Critical: 2 * time.Second},
	"list_tickets":     {Warning: 1 * time.Second, Critical: 3 * time.Second},
	"search_tickets":   {Warning: 1 * time.Second, Critical: 5 * time.Second},
	"get_ticket":       {Warning: 200 * time.Millisecond, Critical: 1 * time.Second},
	"update_ticket":    {Warning: 300 * time.Millisecond, Critical: 1 * time.Second},
	"assign_ticket":    {Warning: 200 * time.Millisecond, Critical: 1 * time.Second},
	"add_message":      {Warning: 300 * time.Millisecond, Critical: 1 * time.Second},
	"get_messages":     {Warning: 500 * time.Millisecond, Critical: 2 * time.Second},
	"batch_load_users": {Warning: 200 * time.Millisecond, Critical: 1 * time.Second},
}

// CheckPerformance checks performance against thresholds and logs appropriately
func (ctx *OperationContext) CheckPerformance(additionalFields ...zap.Field) {
	duration := ctx.Duration()
	thresholds, exists := DefaultThresholds[ctx.Operation]

	fields := ctx.toLogFields()
	fields = append(fields, additionalFields...)

	if exists {
		if duration >= thresholds.Critical {
			fields = append(fields, logger.String("performance_level", "critical"))
			logger.Error("Critical performance threshold exceeded", fields...)
		} else if duration >= thresholds.Warning {
			fields = append(fields, logger.String("performance_level", "warning"))
			logger.Warn("Warning performance threshold exceeded", fields...)
		} else {
			fields = append(fields, logger.String("performance_level", "normal"))
			logger.Debug("Operation completed within normal thresholds", fields...)
		}
	} else {
		// Fallback to default thresholds
		if duration >= 2*time.Second {
			fields = append(fields, logger.String("performance_level", "critical"))
			logger.Error("Operation took too long (no specific threshold defined)", fields...)
		} else if duration >= 1*time.Second {
			fields = append(fields, logger.String("performance_level", "warning"))
			logger.Warn("Operation was slow (no specific threshold defined)", fields...)
		}
	}
}
