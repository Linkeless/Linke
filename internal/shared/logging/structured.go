package logging

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogLevel represents the logging level
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
	LogLevelPanic LogLevel = "panic"
)

// LogFormat represents the logging format
type LogFormat string

const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
)

// Logger represents the standardized logger interface
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, err error, fields ...Field)
	Fatal(ctx context.Context, msg string, err error, fields ...Field)
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
}

// Field represents a logging field
type Field struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// String creates a string field
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an integer field
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates an int64 field
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 creates a float64 field
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a boolean field
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Duration creates a duration field
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value.String()}
}

// Time creates a time field
func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value.Format(time.RFC3339)}
}

// Any creates a field with any value
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Error creates an error field
func Error(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}

// LogConfig represents logger configuration
type LogConfig struct {
	Level           LogLevel  `json:"level" yaml:"level"`
	Format          LogFormat `json:"format" yaml:"format"`
	Output          string    `json:"output" yaml:"output"` // stdout, stderr, file path
	EnableCaller    bool      `json:"enable_caller" yaml:"enable_caller"`
	EnableStacktrace bool     `json:"enable_stacktrace" yaml:"enable_stacktrace"`
	TimestampFormat string    `json:"timestamp_format" yaml:"timestamp_format"`
	ServiceName     string    `json:"service_name" yaml:"service_name"`
	ServiceVersion  string    `json:"service_version" yaml:"service_version"`
	Environment     string    `json:"environment" yaml:"environment"`
}

// DefaultLogConfig returns the default logging configuration
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:           LogLevelInfo,
		Format:          LogFormatJSON,
		Output:          "stdout",
		EnableCaller:    true,
		EnableStacktrace: false,
		TimestampFormat: time.RFC3339,
		ServiceName:     "linke-service",
		ServiceVersion:  "1.0.0",
		Environment:     "development",
	}
}

// StructuredLogger implements the Logger interface using zap
type StructuredLogger struct {
	logger *zap.Logger
	config *LogConfig
	fields []Field
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(config *LogConfig) (*StructuredLogger, error) {
	if config == nil {
		config = DefaultLogConfig()
	}

	zapConfig := zap.NewProductionConfig()

	// Set log level
	level, err := zapcore.ParseLevel(string(config.Level))
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	zapConfig.Level.SetLevel(level)

	// Set encoding
	if config.Format == LogFormatText {
		zapConfig.Encoding = "console"
	} else {
		zapConfig.Encoding = "json"
	}

	// Set output paths
	if config.Output == "stdout" {
		zapConfig.OutputPaths = []string{"stdout"}
	} else if config.Output == "stderr" {
		zapConfig.OutputPaths = []string{"stderr"}
	} else {
		zapConfig.OutputPaths = []string{config.Output}
	}

	// Configure encoder
	zapConfig.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if config.TimestampFormat != "" {
		zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(config.TimestampFormat)
	}

	// Disable caller info if not needed
	if !config.EnableCaller {
		zapConfig.DisableCaller = true
	}

	// Disable stacktrace if not needed
	if !config.EnableStacktrace {
		zapConfig.DisableStacktrace = true
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	structuredLogger := &StructuredLogger{
		logger: logger,
		config: config,
		fields: []Field{
			String("service", config.ServiceName),
			String("version", config.ServiceVersion),
			String("environment", config.Environment),
		},
	}

	return structuredLogger, nil
}

// zapFieldsFromFields converts Field slice to zap.Field slice
func (sl *StructuredLogger) zapFieldsFromFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		zapFields = append(zapFields, zap.Any(field.Key, field.Value))
	}
	return zapFields
}

// fieldsFromContext extracts logging fields from context
func (sl *StructuredLogger) fieldsFromContext(ctx context.Context) []Field {
	var fields []Field

	if ctx == nil {
		return fields
	}

	// Extract common context values
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if tid, ok := traceID.(string); ok {
			fields = append(fields, String("trace_id", tid))
		}
	}

	if requestID := ctx.Value("request_id"); requestID != nil {
		if rid, ok := requestID.(string); ok {
			fields = append(fields, String("request_id", rid))
		}
	}

	if userID := ctx.Value("user_id"); userID != nil {
		if uid, ok := userID.(string); ok {
			fields = append(fields, String("user_id", uid))
		}
	}

	if operation := ctx.Value("operation"); operation != nil {
		if op, ok := operation.(string); ok {
			fields = append(fields, String("operation", op))
		}
	}

	return fields
}

// addStackTrace adds stack trace information
func (sl *StructuredLogger) addStackTrace(fields []Field, skip int) []Field {
	if !sl.config.EnableStacktrace {
		return fields
	}

	const maxDepth = 10
	var stack []string

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

		stack = append(stack, fmt.Sprintf("%s (%s:%d)", fnName, file, line))
	}

	if len(stack) > 0 {
		fields = append(fields, Any("stack_trace", stack))
	}

	return fields
}

// combineFields combines multiple field slices
func (sl *StructuredLogger) combineFields(fieldSlices ...[]Field) []Field {
	var combined []Field
	for _, fields := range fieldSlices {
		combined = append(combined, fields...)
	}
	return combined
}

// Debug logs a debug message
func (sl *StructuredLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	combinedFields := sl.combineFields(sl.fields, sl.fieldsFromContext(ctx), fields)
	sl.logger.Debug(msg, sl.zapFieldsFromFields(combinedFields)...)
}

// Info logs an info message
func (sl *StructuredLogger) Info(ctx context.Context, msg string, fields ...Field) {
	combinedFields := sl.combineFields(sl.fields, sl.fieldsFromContext(ctx), fields)
	sl.logger.Info(msg, sl.zapFieldsFromFields(combinedFields)...)
}

// Warn logs a warning message
func (sl *StructuredLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	combinedFields := sl.combineFields(sl.fields, sl.fieldsFromContext(ctx), fields)
	sl.logger.Warn(msg, sl.zapFieldsFromFields(combinedFields)...)
}

// Error logs an error message
func (sl *StructuredLogger) Error(ctx context.Context, msg string, err error, fields ...Field) {
	combinedFields := sl.combineFields(sl.fields, sl.fieldsFromContext(ctx), fields)
	if err != nil {
		combinedFields = append(combinedFields, Error(err))
		combinedFields = sl.addStackTrace(combinedFields, 2)
	}
	sl.logger.Error(msg, sl.zapFieldsFromFields(combinedFields)...)
}

// Fatal logs a fatal message and exits
func (sl *StructuredLogger) Fatal(ctx context.Context, msg string, err error, fields ...Field) {
	combinedFields := sl.combineFields(sl.fields, sl.fieldsFromContext(ctx), fields)
	if err != nil {
		combinedFields = append(combinedFields, Error(err))
		combinedFields = sl.addStackTrace(combinedFields, 2)
	}
	sl.logger.Fatal(msg, sl.zapFieldsFromFields(combinedFields)...)
}

// With returns a new logger with additional fields
func (sl *StructuredLogger) With(fields ...Field) Logger {
	return &StructuredLogger{
		logger: sl.logger,
		config: sl.config,
		fields: sl.combineFields(sl.fields, fields),
	}
}

// WithContext returns a new logger with context fields
func (sl *StructuredLogger) WithContext(ctx context.Context) Logger {
	contextFields := sl.fieldsFromContext(ctx)
	return &StructuredLogger{
		logger: sl.logger,
		config: sl.config,
		fields: sl.combineFields(sl.fields, contextFields),
	}
}

// Note: LegacyLogrusAdapter removed to avoid additional dependencies
// The StructuredLogger using zap is the recommended approach

// Global logger instance
var globalLogger Logger

// InitializeLogger initializes the global logger
func InitializeLogger(config *LogConfig) error {
	logger, err := NewStructuredLogger(config)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// GetLogger returns the global logger instance
func GetLogger() Logger {
	if globalLogger == nil {
		// Initialize with default config if not set
		logger, _ := NewStructuredLogger(DefaultLogConfig())
		globalLogger = logger
	}
	return globalLogger
}

// Convenience functions for global logger

// Debug logs a debug message using the global logger
func Debug(ctx context.Context, msg string, fields ...Field) {
	GetLogger().Debug(ctx, msg, fields...)
}

// Info logs an info message using the global logger
func Info(ctx context.Context, msg string, fields ...Field) {
	GetLogger().Info(ctx, msg, fields...)
}

// Warn logs a warning message using the global logger
func Warn(ctx context.Context, msg string, fields ...Field) {
	GetLogger().Warn(ctx, msg, fields...)
}

// LogError logs an error message using the global logger
func LogError(ctx context.Context, msg string, err error, fields ...Field) {
	GetLogger().Error(ctx, msg, err, fields...)
}

// Fatal logs a fatal message using the global logger
func Fatal(ctx context.Context, msg string, err error, fields ...Field) {
	GetLogger().Fatal(ctx, msg, err, fields...)
}