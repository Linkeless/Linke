package middleware

import (
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// ErrorHandler middleware provides comprehensive error handling
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic with stack trace
				stack := make([]byte, 1024*8)
				stack = stack[:runtime.Stack(stack, false)]
				
				logger.Error("Recovered from panic",
					logger.String("error", toString(r)),
					logger.String("path", c.Request.URL.Path),
					logger.String("method", c.Request.Method),
					logger.String("client_ip", c.ClientIP()),
					logger.String("user_agent", c.Request.UserAgent()),
					logger.String("stack", string(stack)))

				// Return appropriate error response
				if c.IsAborted() {
					return
				}

				// Check if response has already been written
				if c.Writer.Written() {
					return
				}

				// Return generic error to avoid information disclosure
				response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Internal server error occurred")
			}
		}()

		c.Next()
	}
}

// DatabaseErrorHandler middleware handles database-related errors
func DatabaseErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check for database-related errors in the context
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			logger.Error("Database error occurred",
				logger.String("error", err.Error()),
				logger.String("path", c.Request.URL.Path),
				logger.String("method", c.Request.Method),
				logger.String("client_ip", c.ClientIP()))

			// Don't overwrite existing response
			if c.Writer.Written() {
				return
			}

			// Categorize database errors and return appropriate responses
			errMsg := strings.ToLower(err.Error())
			
			switch {
			case strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "unique"):
				response.Error(c, http.StatusConflict, http.StatusConflict, "Resource already exists")
			case strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "no rows"):
				response.Error(c, http.StatusNotFound, http.StatusNotFound, "Resource not found")
			case strings.Contains(errMsg, "foreign key") || strings.Contains(errMsg, "constraint"):
				response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid operation: constraint violation")
			case strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline"):
				response.Error(c, http.StatusRequestTimeout, http.StatusRequestTimeout, "Request timeout")
			case strings.Contains(errMsg, "connection") || strings.Contains(errMsg, "network"):
				response.Error(c, http.StatusServiceUnavailable, http.StatusServiceUnavailable, "Service temporarily unavailable")
			default:
				response.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "Database operation failed")
			}
		}
	}
}

// ValidationErrorHandler middleware handles validation errors
func ValidationErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle validation errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			// Skip if response already written
			if c.Writer.Written() {
				return
			}

			errMsg := err.Error()
			
			// Check for common validation error patterns
			switch {
			case strings.Contains(errMsg, "required"):
				response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Required field missing")
			case strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "format"):
				response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid input format")
			case strings.Contains(errMsg, "length") || strings.Contains(errMsg, "size"):
				response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Input length validation failed")
			case strings.Contains(errMsg, "email"):
				response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Invalid email format")
			case strings.Contains(errMsg, "password"):
				response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Password requirements not met")
			default:
				response.Error(c, http.StatusBadRequest, http.StatusBadRequest, "Validation failed")
			}
		}
	}
}

// SecurityErrorHandler middleware handles security-related errors
func SecurityErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle security errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			// Skip if response already written
			if c.Writer.Written() {
				return
			}

			errMsg := strings.ToLower(err.Error())
			
			// Log security-related errors with additional context
			if isSecurityError(errMsg) {
				logger.Warn("Security error detected",
					logger.String("error", err.Error()),
					logger.String("path", c.Request.URL.Path),
					logger.String("method", c.Request.Method),
					logger.String("client_ip", c.ClientIP()),
					logger.String("user_agent", c.Request.UserAgent()),
					logger.String("referer", c.Request.Referer()))
			}

			switch {
			case strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "authentication"):
				response.Error(c, http.StatusUnauthorized, http.StatusUnauthorized, "Authentication required")
			case strings.Contains(errMsg, "forbidden") || strings.Contains(errMsg, "permission"):
				response.Error(c, http.StatusForbidden, http.StatusForbidden, "Access denied")
			case strings.Contains(errMsg, "token") && strings.Contains(errMsg, "invalid"):
				response.Error(c, http.StatusUnauthorized, http.StatusUnauthorized, "Invalid token")
			case strings.Contains(errMsg, "token") && strings.Contains(errMsg, "expired"):
				response.Error(c, http.StatusUnauthorized, http.StatusUnauthorized, "Token expired")
			case strings.Contains(errMsg, "rate limit"):
				response.Error(c, http.StatusTooManyRequests, http.StatusTooManyRequests, "Rate limit exceeded")
			case strings.Contains(errMsg, "csrf"):
				response.Error(c, http.StatusForbidden, http.StatusForbidden, "CSRF validation failed")
			}
		}
	}
}

// PaymentErrorHandler middleware handles payment-specific errors
func PaymentErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle payment errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			// Skip if response already written
			if c.Writer.Written() {
				return
			}

			errMsg := strings.ToLower(err.Error())
			
			// Log payment errors for audit
			logger.Error("Payment error occurred",
				logger.String("error", err.Error()),
				logger.String("path", c.Request.URL.Path),
				logger.String("method", c.Request.Method),
				logger.String("client_ip", c.ClientIP()))

			switch {
			case strings.Contains(errMsg, "insufficient"):
				response.Error(c, http.StatusPaymentRequired, http.StatusPaymentRequired, "Insufficient funds")
			case strings.Contains(errMsg, "payment") && strings.Contains(errMsg, "failed"):
				response.Error(c, http.StatusPaymentRequired, http.StatusPaymentRequired, "Payment processing failed")
			case strings.Contains(errMsg, "gateway"):
				response.Error(c, http.StatusServiceUnavailable, http.StatusServiceUnavailable, "Payment gateway unavailable")
			case strings.Contains(errMsg, "duplicate") && strings.Contains(errMsg, "payment"):
				response.Error(c, http.StatusConflict, http.StatusConflict, "Payment already processed")
			case strings.Contains(errMsg, "expired") && strings.Contains(errMsg, "order"):
				response.Error(c, http.StatusGone, http.StatusGone, "Payment order expired")
			}
		}
	}
}

// Helper functions

func toString(v interface{}) string {
	switch s := v.(type) {
	case string:
		return s
	case error:
		return s.Error()
	default:
		return "Unknown error"
	}
}

func isSecurityError(errMsg string) bool {
	securityKeywords := []string{
		"unauthorized", "forbidden", "permission", "token", "authentication",
		"csrf", "xss", "injection", "rate limit", "suspicious", "blocked",
	}
	
	for _, keyword := range securityKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}
	return false
}

// RequestContextHandler adds request context information for error handling
func RequestContextHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add request ID for tracing
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// Add request start time for performance monitoring
		c.Set("request_start", time.Now().Unix())

		c.Next()
	}
}

func generateRequestID() string {
	// Simple request ID generation (in production, use a proper UUID library)
	return strconv.FormatInt(int64(hash([]byte("request"))), 36)
}

func hash(data []byte) uint32 {
	hash := uint32(0)
	for _, b := range data {
		hash = hash*31 + uint32(b)
	}
	return hash
}