package errors

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// ErrorContext provides additional context for errors
type ErrorContext struct {
	TraceID     string            `json:"trace_id,omitempty"`
	RequestID   string            `json:"request_id,omitempty"`
	UserID      string            `json:"user_id,omitempty"`
	Operation   string            `json:"operation,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
	StackTrace  []StackFrame      `json:"stack_trace,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
	Annotations []string          `json:"annotations,omitempty"`
}

// StackFrame represents a single frame in a stack trace
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// ContextualError wraps BusinessError with additional context
type ContextualError struct {
	*BusinessError
	Context *ErrorContext `json:"context,omitempty"`
}

// Error implements the error interface
func (e *ContextualError) Error() string {
	if e.Context != nil && e.Context.Operation != "" {
		return fmt.Sprintf("[%s] %s", e.Context.Operation, e.BusinessError.Error())
	}
	return e.BusinessError.Error()
}

// Unwrap returns the underlying error
func (e *ContextualError) Unwrap() error {
	return e.BusinessError
}

// WithContext adds context to a BusinessError
func WithContext(err *BusinessError, ctx *ErrorContext) *ContextualError {
	return &ContextualError{
		BusinessError: err,
		Context:       ctx,
	}
}

// WithContextFromCtx creates an ErrorContext from a context.Context
func WithContextFromCtx(err *BusinessError, ctx context.Context) *ContextualError {
	errorCtx := &ErrorContext{
		Timestamp: time.Now(),
	}

	// Extract common context values
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if tid, ok := traceID.(string); ok {
			errorCtx.TraceID = tid
		}
	}

	if requestID := ctx.Value("request_id"); requestID != nil {
		if rid, ok := requestID.(string); ok {
			errorCtx.RequestID = rid
		}
	}

	if userID := ctx.Value("user_id"); userID != nil {
		if uid, ok := userID.(string); ok {
			errorCtx.UserID = uid
		}
	}

	if operation := ctx.Value("operation"); operation != nil {
		if op, ok := operation.(string); ok {
			errorCtx.Operation = op
		}
	}

	return WithContext(err, errorCtx)
}

// WithStackTrace adds stack trace to error context
func (ec *ErrorContext) WithStackTrace(skip int) *ErrorContext {
	if ec == nil {
		ec = &ErrorContext{}
	}

	const maxDepth = 32
	var frames []StackFrame

	for i := skip; i < skip+maxDepth; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		fnName := "unknown"
		if fn != nil {
			fnName = fn.Name()
		}

		frames = append(frames, StackFrame{
			Function: fnName,
			File:     file,
			Line:     line,
		})
	}

	ec.StackTrace = frames
	return ec
}

// WithMetadata adds metadata to error context
func (ec *ErrorContext) WithMetadata(key string, value any) *ErrorContext {
	if ec == nil {
		ec = &ErrorContext{}
	}
	if ec.Metadata == nil {
		ec.Metadata = make(map[string]any)
	}
	ec.Metadata[key] = value
	return ec
}

// WithAnnotation adds an annotation to error context
func (ec *ErrorContext) WithAnnotation(annotation string) *ErrorContext {
	if ec == nil {
		ec = &ErrorContext{}
	}
	ec.Annotations = append(ec.Annotations, annotation)
	return ec
}

// NewErrorContext creates a new error context
func NewErrorContext() *ErrorContext {
	return &ErrorContext{
		Timestamp: time.Now(),
		Metadata:  make(map[string]any),
	}
}

// CaptureStack captures the current stack trace
func CaptureStack(skip int) []StackFrame {
	return NewErrorContext().WithStackTrace(skip + 1).StackTrace
}

// WithOperation sets the operation name
func (ec *ErrorContext) WithOperation(operation string) *ErrorContext {
	if ec == nil {
		ec = &ErrorContext{}
	}
	ec.Operation = operation
	return ec
}

// WithTraceID sets the trace ID
func (ec *ErrorContext) WithTraceID(traceID string) *ErrorContext {
	if ec == nil {
		ec = &ErrorContext{}
	}
	ec.TraceID = traceID
	return ec
}

// WithRequestID sets the request ID
func (ec *ErrorContext) WithRequestID(requestID string) *ErrorContext {
	if ec == nil {
		ec = &ErrorContext{}
	}
	ec.RequestID = requestID
	return ec
}

// WithUserID sets the user ID
func (ec *ErrorContext) WithUserID(userID string) *ErrorContext {
	if ec == nil {
		ec = &ErrorContext{}
	}
	ec.UserID = userID
	return ec
}

// Convenience functions for creating contextual errors

// NewWithContext creates a new contextual error
func NewWithContext(code ErrorCode, message string, ctx *ErrorContext) *ContextualError {
	return WithContext(New(code, message), ctx)
}

// NewfWithContext creates a new contextual error with formatted message
func NewfWithContext(code ErrorCode, ctx *ErrorContext, format string, args ...interface{}) *ContextualError {
	return WithContext(Newf(code, format, args...), ctx)
}

// WrapWithContext wraps an error with context
func WrapWithContext(err error, code ErrorCode, message string, ctx *ErrorContext) *ContextualError {
	return WithContext(Wrap(err, code, message), ctx)
}

// WrapfWithContext wraps an error with formatted message and context
func WrapfWithContext(err error, code ErrorCode, ctx *ErrorContext, format string, args ...interface{}) *ContextualError {
	return WithContext(Wrapf(err, code, format, args...), ctx)
}

// IsContextualError checks if an error is a ContextualError
func IsContextualError(err error) bool {
	_, ok := err.(*ContextualError)
	return ok
}

// AsContextualError attempts to extract a ContextualError from an error
func AsContextualError(err error) (*ContextualError, bool) {
	if ce, ok := err.(*ContextualError); ok {
		return ce, true
	}
	return nil, false
}