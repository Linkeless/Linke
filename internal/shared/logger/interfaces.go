package logger

import (
	"go.uber.org/zap"
)

// Logger interface defines the logging contract
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
	Sync() error
}

// ZapLogger wraps zap.Logger to implement our Logger interface
type ZapLogger struct {
	logger *zap.Logger
}

// NewZapLogger creates a new ZapLogger
func NewZapLogger(logger *zap.Logger) *ZapLogger {
	return &ZapLogger{
		logger: logger,
	}
}

func (l *ZapLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

func (l *ZapLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

func (l *ZapLogger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

func (l *ZapLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

func (l *ZapLogger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

func (l *ZapLogger) With(fields ...zap.Field) Logger {
	return &ZapLogger{
		logger: l.logger.With(fields...),
	}
}

func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}

// ContextualLogger provides contextual logging capabilities
type ContextualLogger interface {
	Logger
	WithContext(key string, value any) Logger
	WithRequestID(requestID string) Logger
	WithUserID(userID string) Logger
}

// ContextualZapLogger extends ZapLogger with contextual logging
type ContextualZapLogger struct {
	*ZapLogger
}

// NewContextualZapLogger creates a new contextual logger
func NewContextualZapLogger(logger *zap.Logger) *ContextualZapLogger {
	return &ContextualZapLogger{
		ZapLogger: NewZapLogger(logger),
	}
}

func (l *ContextualZapLogger) WithContext(key string, value any) Logger {
	return &ContextualZapLogger{
		ZapLogger: NewZapLogger(l.logger.With(zap.Any(key, value))),
	}
}

func (l *ContextualZapLogger) WithRequestID(requestID string) Logger {
	return &ContextualZapLogger{
		ZapLogger: NewZapLogger(l.logger.With(zap.String("request_id", requestID))),
	}
}

func (l *ContextualZapLogger) WithUserID(userID string) Logger {
	return &ContextualZapLogger{
		ZapLogger: NewZapLogger(l.logger.With(zap.String("user_id", userID))),
	}
}

func (l *ContextualZapLogger) With(fields ...zap.Field) Logger {
	return &ContextualZapLogger{
		ZapLogger: NewZapLogger(l.logger.With(fields...)),
	}
}

// NopLogger is a no-operation logger for testing
type NopLogger struct{}

func NewNopLogger() *NopLogger {
	return &NopLogger{}
}

func (l *NopLogger) Debug(msg string, fields ...zap.Field) {}
func (l *NopLogger) Info(msg string, fields ...zap.Field)  {}
func (l *NopLogger) Warn(msg string, fields ...zap.Field)  {}
func (l *NopLogger) Error(msg string, fields ...zap.Field) {}
func (l *NopLogger) Fatal(msg string, fields ...zap.Field) {}
func (l *NopLogger) With(fields ...zap.Field) Logger       { return l }
func (l *NopLogger) Sync() error                           { return nil }

// GetGlobalLogger returns the global logger instance as our Logger interface
func GetGlobalLogger() Logger {
	return NewZapLogger(GetLogger())
}

// GetContextualLogger returns a contextual logger
func GetContextualLogger() ContextualLogger {
	return NewContextualZapLogger(GetLogger())
}
