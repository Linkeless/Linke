package logging

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GinLoggerConfig represents the configuration for Gin logging middleware
type GinLoggerConfig struct {
	Logger          Logger
	SkipPaths       []string
	EnableRequestID bool
	EnableUserID    bool
	EnableTraceID   bool
	LogRequestBody  bool
	LogResponseBody bool
	MaxBodySize     int64
}

// DefaultGinLoggerConfig returns the default Gin logger configuration
func DefaultGinLoggerConfig() *GinLoggerConfig {
	return &GinLoggerConfig{
		Logger:          GetLogger(),
		SkipPaths:       []string{"/health", "/metrics", "/ping"},
		EnableRequestID: true,
		EnableUserID:    true,
		EnableTraceID:   true,
		LogRequestBody:  false,
		LogResponseBody: false,
		MaxBodySize:     1024, // 1KB
	}
}

// GinLogger returns a Gin middleware for structured logging
func GinLogger(config *GinLoggerConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultGinLoggerConfig()
	}

	return func(c *gin.Context) {
		// Skip logging for specified paths
		path := c.Request.URL.Path
		for _, skipPath := range config.SkipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}

		// Start timer
		start := time.Now()

		// Generate request ID
		var requestID string
		if config.EnableRequestID {
			if id := c.GetHeader("X-Request-ID"); id != "" {
				requestID = id
			} else {
				requestID = uuid.New().String()
				c.Header("X-Request-ID", requestID)
			}
		}

		// Generate trace ID
		var traceID string
		if config.EnableTraceID {
			if id := c.GetHeader("X-Trace-ID"); id != "" {
				traceID = id
			} else {
				traceID = uuid.New().String()
				c.Header("X-Trace-ID", traceID)
			}
		}

		// Extract user ID (from JWT or session)
		var userID string
		if config.EnableUserID {
			if uid, exists := c.Get("user_id"); exists {
				if uidStr, ok := uid.(string); ok {
					userID = uidStr
				}
			}
		}

		// Create context with logging fields
		ctx := context.Background()
		if requestID != "" {
			ctx = context.WithValue(ctx, "request_id", requestID)
		}
		if traceID != "" {
			ctx = context.WithValue(ctx, "trace_id", traceID)
		}
		if userID != "" {
			ctx = context.WithValue(ctx, "user_id", userID)
		}

		// Store context in Gin context
		c.Set("logging_context", ctx)

		// Read request body if needed
		var requestBody []byte
		if config.LogRequestBody && c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, config.MaxBodySize))
			if err == nil {
				requestBody = bodyBytes
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// Capture response body if needed
		var responseBody []byte
		if config.LogResponseBody {
			writer := &responseBodyWriter{
				ResponseWriter: c.Writer,
				body:           bytes.NewBuffer([]byte{}),
				maxSize:        config.MaxBodySize,
			}
			c.Writer = writer
			defer func() {
				responseBody = writer.body.Bytes()
			}()
		}

		// Log request start
		fields := []Field{
			String("method", c.Request.Method),
			String("path", path),
			String("query", c.Request.URL.RawQuery),
			String("user_agent", c.Request.UserAgent()),
			String("remote_addr", c.ClientIP()),
		}

		if requestID != "" {
			fields = append(fields, String("request_id", requestID))
		}
		if traceID != "" {
			fields = append(fields, String("trace_id", traceID))
		}
		if userID != "" {
			fields = append(fields, String("user_id", userID))
		}

		if config.LogRequestBody && len(requestBody) > 0 {
			fields = append(fields, String("request_body", string(requestBody)))
		}

		config.Logger.Info(ctx, "Request started", fields...)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Prepare response fields
		responseFields := []Field{
			String("method", c.Request.Method),
			String("path", path),
			Int("status", c.Writer.Status()),
			Duration("duration", duration),
			Int("response_size", c.Writer.Size()),
		}

		if requestID != "" {
			responseFields = append(responseFields, String("request_id", requestID))
		}
		if traceID != "" {
			responseFields = append(responseFields, String("trace_id", traceID))
		}
		if userID != "" {
			responseFields = append(responseFields, String("user_id", userID))
		}

		if config.LogResponseBody && len(responseBody) > 0 {
			responseFields = append(responseFields, String("response_body", string(responseBody)))
		}

		// Add error information if any
		if len(c.Errors) > 0 {
			errors := make([]string, len(c.Errors))
			for i, err := range c.Errors {
				errors[i] = err.Error()
			}
			responseFields = append(responseFields, Any("errors", errors))
		}

		// Log based on status code
		status := c.Writer.Status()
		switch {
		case status >= 500:
			config.Logger.Error(ctx, "Request completed with server error", nil, responseFields...)
		case status >= 400:
			config.Logger.Warn(ctx, "Request completed with client error", responseFields...)
		case status >= 300:
			config.Logger.Info(ctx, "Request completed with redirect", responseFields...)
		default:
			config.Logger.Info(ctx, "Request completed successfully", responseFields...)
		}
	}
}

// responseBodyWriter captures response body for logging
type responseBodyWriter struct {
	gin.ResponseWriter
	body    *bytes.Buffer
	maxSize int64
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	// Write to original response
	n, err := w.ResponseWriter.Write(b)

	// Capture for logging (with size limit)
	if w.body.Len() < int(w.maxSize) {
		remaining := int(w.maxSize) - w.body.Len()
		if len(b) <= remaining {
			w.body.Write(b)
		} else {
			w.body.Write(b[:remaining])
		}
	}

	return n, err
}

// ContextLogger returns a logger with context from Gin
func ContextLogger(c *gin.Context) Logger {
	if ctx, exists := c.Get("logging_context"); exists {
		if context, ok := ctx.(context.Context); ok {
			return GetLogger().WithContext(context)
		}
	}
	return GetLogger()
}

// LoggerFromContext extracts logger from context
func LoggerFromContext(ctx context.Context) Logger {
	return GetLogger().WithContext(ctx)
}

// WithRequestContext adds request context to logger
func WithRequestContext(logger Logger, c *gin.Context) Logger {
	fields := []Field{}

	if method := c.Request.Method; method != "" {
		fields = append(fields, String("method", method))
	}

	if path := c.Request.URL.Path; path != "" {
		fields = append(fields, String("path", path))
	}

	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		fields = append(fields, String("request_id", requestID))
	}

	if traceID := c.GetHeader("X-Trace-ID"); traceID != "" {
		fields = append(fields, String("trace_id", traceID))
	}

	if userID, exists := c.Get("user_id"); exists {
		if uidStr, ok := userID.(string); ok {
			fields = append(fields, String("user_id", uidStr))
		}
	}

	return logger.With(fields...)
}

// Operation logging helpers

// LogOperation logs the start and end of an operation
func LogOperation(ctx context.Context, operation string, fn func() error) error {
	logger := GetLogger().WithContext(ctx).With(String("operation", operation))
	
	start := time.Now()
	logger.Info(ctx, "Operation started")

	err := fn()
	duration := time.Since(start)

	if err != nil {
		logger.Error(ctx, "Operation failed", err, Duration("duration", duration))
		return err
	}

	logger.Info(ctx, "Operation completed", Duration("duration", duration))
	return nil
}

// LogAsyncOperation logs an asynchronous operation
func LogAsyncOperation(ctx context.Context, operation string, fn func() error) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				GetLogger().Error(ctx, "Async operation panicked", 
					fmt.Errorf("panic: %v", r), 
					String("operation", operation))
			}
		}()

		err := LogOperation(ctx, operation, fn)
		if err != nil {
			GetLogger().Error(ctx, "Async operation failed", err, String("operation", operation))
		}
	}()
}

// Database operation logging

// LogDatabaseOperation logs database operations with timing
func LogDatabaseOperation(ctx context.Context, operation, table string, fn func() error) error {
	logger := GetLogger().WithContext(ctx).With(
		String("operation", "db."+operation),
		String("table", table),
	)

	start := time.Now()
	logger.Debug(ctx, "Database operation started")

	err := fn()
	duration := time.Since(start)

	if err != nil {
		logger.Error(ctx, "Database operation failed", err, Duration("duration", duration))
		return err
	}

	logger.Debug(ctx, "Database operation completed", Duration("duration", duration))
	return nil
}

// Cache operation logging

// LogCacheOperation logs cache operations
func LogCacheOperation(ctx context.Context, operation, key string, hit bool) {
	logger := GetLogger().WithContext(ctx).With(
		String("operation", "cache."+operation),
		String("key", key),
		Bool("hit", hit),
	)

	if hit {
		logger.Debug(ctx, "Cache hit")
	} else {
		logger.Debug(ctx, "Cache miss")
	}
}

// External service logging

// LogExternalCall logs external service calls
func LogExternalCall(ctx context.Context, service, method string, fn func() error) error {
	logger := GetLogger().WithContext(ctx).With(
		String("operation", "external."+service),
		String("method", method),
	)

	start := time.Now()
	logger.Info(ctx, "External call started")

	err := fn()
	duration := time.Since(start)

	if err != nil {
		logger.Error(ctx, "External call failed", err, Duration("duration", duration))
		return err
	}

	logger.Info(ctx, "External call completed", Duration("duration", duration))
	return nil
}