package errors

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strings"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// ConvertServiceError converts common service errors to BusinessError
func ConvertServiceError(err error, service, resource string) error {
	if err == nil {
		return nil
	}

	// Already a business error
	if IsBusinessError(err) {
		return err
	}

	ctx := NewErrorContext().
		WithOperation(fmt.Sprintf("%s.%s", service, resource)).
		WithMetadata("service", service).
		WithMetadata("resource", resource).
		WithStackTrace(1)

	// Database errors (GORM)
	if dbErr := convertGORMError(err, ctx); dbErr != nil {
		return dbErr
	}

	// Redis errors
	if redisErr := convertRedisError(err, ctx); redisErr != nil {
		return redisErr
	}

	// Network errors
	if netErr := convertNetworkError(err, ctx); netErr != nil {
		return netErr
	}

	// Context errors
	if ctxErr := convertContextError(err, ctx); ctxErr != nil {
		return ctxErr
	}

	// System errors
	if sysErr := convertSystemError(err, ctx); sysErr != nil {
		return sysErr
	}

	// Default to internal error
	return WrapWithContext(err, ErrCodeInternal, "Internal service error", ctx)
}

// ConvertServiceErrorWithID converts service errors with typed ID
func ConvertServiceErrorWithID(err error, service, resource string, id interface{}) error {
	if err == nil {
		return nil
	}

	var idStr string
	switch v := id.(type) {
	case string:
		idStr = v
	case uint:
		idStr = fmt.Sprintf("%d", v)
	case int:
		idStr = fmt.Sprintf("%d", v)
	case uint64:
		idStr = fmt.Sprintf("%d", v)
	case int64:
		idStr = fmt.Sprintf("%d", v)
	default:
		idStr = fmt.Sprintf("%v", v)
	}

	if converted := ConvertServiceError(err, service, resource); converted != nil {
		if ce, ok := converted.(*ContextualError); ok {
			if ce.Context != nil {
				ce.Context.WithMetadata(resource+"_id", idStr)
			}
			return ce
		}
		return converted
	}

	return err
}

// convertGORMError converts GORM database errors
func convertGORMError(err error, ctx *ErrorContext) *ContextualError {
	switch {
	case err == gorm.ErrRecordNotFound:
		return WrapWithContext(err, ErrCodeNotFound, "Record not found", ctx)
	case err == gorm.ErrInvalidTransaction:
		return WrapWithContext(err, ErrCodeInvalidState, "Invalid database transaction", ctx)
	case err == gorm.ErrNotImplemented:
		return WrapWithContext(err, ErrCodeInternal, "Database operation not implemented", ctx)
	case err == gorm.ErrMissingWhereClause:
		return WrapWithContext(err, ErrCodeInvalidInput, "Missing WHERE clause in query", ctx)
	case err == gorm.ErrUnsupportedRelation:
		return WrapWithContext(err, ErrCodeInternal, "Unsupported database relation", ctx)
	case err == gorm.ErrPrimaryKeyRequired:
		return WrapWithContext(err, ErrCodeInvalidInput, "Primary key required for operation", ctx)
	case err == gorm.ErrModelValueRequired:
		return WrapWithContext(err, ErrCodeInvalidInput, "Model value required for operation", ctx)
	case err == gorm.ErrInvalidData:
		return WrapWithContext(err, ErrCodeInvalidInput, "Invalid data provided", ctx)
	}

	// Check for SQL errors
	if err == sql.ErrNoRows {
		return WrapWithContext(err, ErrCodeNotFound, "No rows found", ctx)
	}

	if err == sql.ErrTxDone {
		return WrapWithContext(err, ErrCodeInvalidState, "Transaction already committed or rolled back", ctx)
	}

	if err == sql.ErrConnDone {
		return WrapWithContext(err, ErrCodeUnavailable, "Database connection closed", ctx)
	}

	// Check for common database constraint violations
	errMsg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errMsg, "duplicate") || strings.Contains(errMsg, "unique constraint"):
		return WrapWithContext(err, ErrCodeAlreadyExists, "Resource already exists (duplicate)", ctx)
	case strings.Contains(errMsg, "foreign key constraint"):
		return WrapWithContext(err, ErrCodeConflict, "Foreign key constraint violation", ctx)
	case strings.Contains(errMsg, "check constraint"):
		return WrapWithContext(err, ErrCodeInvalidInput, "Check constraint violation", ctx)
	case strings.Contains(errMsg, "not null constraint"):
		return WrapWithContext(err, ErrCodeInvalidInput, "Required field is null", ctx)
	case strings.Contains(errMsg, "connection refused"):
		return WrapWithContext(err, ErrCodeUnavailable, "Database connection refused", ctx)
	case strings.Contains(errMsg, "timeout"):
		return WrapWithContext(err, ErrCodeTimeout, "Database operation timeout", ctx)
	}

	return nil
}

// convertRedisError converts Redis errors
func convertRedisError(err error, ctx *ErrorContext) *ContextualError {
	switch {
	case err == redis.Nil:
		return WrapWithContext(err, ErrCodeNotFound, "Cache key not found", ctx)
	case err == redis.TxFailedErr:
		return WrapWithContext(err, ErrCodeConflict, "Redis transaction failed", ctx)
	}

	errMsg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errMsg, "connection refused"):
		return WrapWithContext(err, ErrCodeUnavailable, "Redis connection refused", ctx)
	case strings.Contains(errMsg, "timeout"):
		return WrapWithContext(err, ErrCodeTimeout, "Redis operation timeout", ctx)
	case strings.Contains(errMsg, "connection reset"):
		return WrapWithContext(err, ErrCodeUnavailable, "Redis connection reset", ctx)
	case strings.Contains(errMsg, "broken pipe"):
		return WrapWithContext(err, ErrCodeUnavailable, "Redis connection broken", ctx)
	case strings.Contains(errMsg, "readonly"):
		return WrapWithContext(err, ErrCodeForbidden, "Redis is in read-only mode", ctx)
	case strings.Contains(errMsg, "noauth"):
		return WrapWithContext(err, ErrCodeUnauthorized, "Redis authentication required", ctx)
	}

	return nil
}

// convertNetworkError converts network errors
func convertNetworkError(err error, ctx *ErrorContext) *ContextualError {
	if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() {
			return WrapWithContext(err, ErrCodeTimeout, "Network operation timeout", ctx)
		}
		if netErr.Temporary() {
			return WrapWithContext(err, ErrCodeUnavailable, "Temporary network error", ctx)
		}
	}

	if opErr, ok := err.(*net.OpError); ok {
		switch {
		case opErr.Op == "dial":
			return WrapWithContext(err, ErrCodeUnavailable, "Failed to establish connection", ctx)
		case opErr.Op == "read":
			return WrapWithContext(err, ErrCodeUnavailable, "Network read error", ctx)
		case opErr.Op == "write":
			return WrapWithContext(err, ErrCodeUnavailable, "Network write error", ctx)
		}
	}

	if dnsErr, ok := err.(*net.DNSError); ok {
		if dnsErr.IsNotFound {
			return WrapWithContext(err, ErrCodeNotFound, "DNS name not found", ctx)
		}
		if dnsErr.IsTimeout {
			return WrapWithContext(err, ErrCodeTimeout, "DNS lookup timeout", ctx)
		}
		return WrapWithContext(err, ErrCodeUnavailable, "DNS error", ctx)
	}

	return nil
}

// convertContextError converts context errors
func convertContextError(err error, ctx *ErrorContext) *ContextualError {
	switch err {
	case context.Canceled:
		return WrapWithContext(err, ErrCodeTimeout, "Operation was canceled", ctx)
	case context.DeadlineExceeded:
		return WrapWithContext(err, ErrCodeTimeout, "Operation deadline exceeded", ctx)
	}

	return nil
}

// convertSystemError converts system call errors
func convertSystemError(err error, ctx *ErrorContext) *ContextualError {
	if errno, ok := err.(syscall.Errno); ok {
		switch errno {
		case syscall.ECONNREFUSED:
			return WrapWithContext(err, ErrCodeUnavailable, "Connection refused", ctx)
		case syscall.ETIMEDOUT:
			return WrapWithContext(err, ErrCodeTimeout, "Operation timed out", ctx)
		case syscall.ECONNRESET:
			return WrapWithContext(err, ErrCodeUnavailable, "Connection reset by peer", ctx)
		case syscall.EPIPE:
			return WrapWithContext(err, ErrCodeUnavailable, "Broken pipe", ctx)
		case syscall.ENOENT:
			return WrapWithContext(err, ErrCodeNotFound, "File or directory not found", ctx)
		case syscall.EACCES:
			return WrapWithContext(err, ErrCodeForbidden, "Permission denied", ctx)
		case syscall.EEXIST:
			return WrapWithContext(err, ErrCodeAlreadyExists, "File or directory already exists", ctx)
		case syscall.ENOSPC:
			return WrapWithContext(err, ErrCodeUnavailable, "No space left on device", ctx)
		case syscall.EMFILE:
			return WrapWithContext(err, ErrCodeUnavailable, "Too many open files", ctx)
		}
	}

	return nil
}

// Domain-specific error converters

// ConvertTicketError converts ticket service errors with domain context
func ConvertTicketError(err error, ticketID string) error {
	if err == nil {
		return nil
	}

	if converted := ConvertServiceError(err, "ticket", "ticket"); converted != nil {
		if ce, ok := converted.(*ContextualError); ok {
			if ce.Context != nil {
				ce.Context.WithMetadata("ticket_id", ticketID)
			}
			return ce
		}
		return converted
	}

	return err
}

// ConvertTicketErrorUint converts ticket service errors with uint ID
func ConvertTicketErrorUint(err error, ticketID uint) error {
	return ConvertTicketError(err, fmt.Sprintf("%d", ticketID))
}

// ConvertUserError converts user service errors with domain context
func ConvertUserError(err error, userID string) error {
	if err == nil {
		return nil
	}

	if converted := ConvertServiceError(err, "user", "user"); converted != nil {
		if ce, ok := converted.(*ContextualError); ok {
			if ce.Context != nil {
				ce.Context.WithMetadata("user_id", userID)
			}
			return ce
		}
		return converted
	}

	return err
}

// ConvertUserErrorUint converts user service errors with uint ID
func ConvertUserErrorUint(err error, userID uint) error {
	return ConvertUserError(err, fmt.Sprintf("%d", userID))
}

// ConvertPaymentError converts payment service errors with domain context
func ConvertPaymentError(err error, paymentID string) error {
	if err == nil {
		return nil
	}

	if converted := ConvertServiceError(err, "payment", "payment"); converted != nil {
		if ce, ok := converted.(*ContextualError); ok {
			if ce.Context != nil {
				ce.Context.WithMetadata("payment_id", paymentID)
			}
			return ce
		}
		return converted
	}

	return err
}

// ConvertSubscriptionError converts subscription service errors with domain context
func ConvertSubscriptionError(err error, subscriptionID string) error {
	if err == nil {
		return nil
	}

	if converted := ConvertServiceError(err, "subscription", "subscription"); converted != nil {
		if ce, ok := converted.(*ContextualError); ok {
			if ce.Context != nil {
				ce.Context.WithMetadata("subscription_id", subscriptionID)
			}
			return ce
		}
		return converted
	}

	return err
}

// Timeout helpers

// WithTimeout wraps an operation with a timeout context
func WithTimeout(timeout time.Duration, operation func(ctx context.Context) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- operation(ctx)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return WrapWithContext(ctx.Err(), ErrCodeTimeout, "Operation timeout", 
			NewErrorContext().WithMetadata("timeout", timeout.String()))
	}
}

// RetryWithBackoff executes an operation with retry and exponential backoff
func RetryWithBackoff(maxAttempts int, baseDelay time.Duration, operation func() error) error {
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry non-retryable errors
		if IsBusinessError(err) {
			if be, ok := AsBusinessError(err); ok {
				switch be.Code {
				case ErrCodeForbidden, ErrCodeUnauthorized, ErrCodeInvalidInput, ErrCodeNotFound:
					return err // Don't retry these
				}
			}
		}

		if attempt == maxAttempts {
			break
		}

		// Exponential backoff
		multiplier := 1 << uint(attempt-1)
		delay := time.Duration(float64(baseDelay) * float64(multiplier))
		time.Sleep(delay)
	}

	return WrapWithContext(lastErr, ErrCodeInternal, 
		fmt.Sprintf("Operation failed after %d attempts", maxAttempts),
		NewErrorContext().WithMetadata("attempts", maxAttempts))
}