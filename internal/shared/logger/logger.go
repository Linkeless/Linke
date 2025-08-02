package logger

import (
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var GlobalLogger *zap.Logger

type LogConfig struct {
	Level  string
	Format string
	Output string
}

func InitLogger(config LogConfig) error {
	var zapConfig zap.Config

	switch config.Format {
	case "json":
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.CallerKey = ""     // Disable caller in JSON format
		zapConfig.EncoderConfig.StacktraceKey = "" // Disable stacktrace in JSON format
	case "text", "console":
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapConfig.EncoderConfig.CallerKey = ""     // Disable caller in text format
		zapConfig.EncoderConfig.StacktraceKey = "" // Disable stacktrace in text format
		zapConfig.EncoderConfig.ConsoleSeparator = " "
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	default:
		zapConfig = zap.NewProductionConfig()
		zapConfig.Encoding = "console" // Use console encoder for default format
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapConfig.EncoderConfig.CallerKey = ""     // Disable caller in default format
		zapConfig.EncoderConfig.StacktraceKey = "" // Disable stacktrace in default format
		// Configure console encoder for simple text output
		zapConfig.EncoderConfig.ConsoleSeparator = " "
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	if config.Output != "" && config.Output != "stdout" {
		zapConfig.OutputPaths = []string{config.Output}
		zapConfig.ErrorOutputPaths = []string{config.Output}
	} else {
		zapConfig.OutputPaths = []string{"stdout"}
		zapConfig.ErrorOutputPaths = []string{"stderr"}
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return err
	}

	GlobalLogger = logger
	zap.ReplaceGlobals(logger)

	return nil
}

func GetLogger() *zap.Logger {
	if GlobalLogger == nil {
		config := LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		}
		if err := InitLogger(config); err != nil {
			panic("Failed to initialize logger: " + err.Error())
		}
	}
	return GlobalLogger
}

func Sync() error {
	if GlobalLogger != nil {
		return GlobalLogger.Sync()
	}
	return nil
}

func Info(message string, fields ...zap.Field) {
	GetLogger().Info(message, fields...)
}

func Error(message string, fields ...zap.Field) {
	GetLogger().Error(message, fields...)
}

func Warn(message string, fields ...zap.Field) {
	GetLogger().Warn(message, fields...)
}

func Debug(message string, fields ...zap.Field) {
	GetLogger().Debug(message, fields...)
}

func Fatal(message string, fields ...zap.Field) {
	GetLogger().Fatal(message, fields...)
}

func WithFields(fields ...zap.Field) *zap.Logger {
	return GetLogger().With(fields...)
}

func String(key, val string) zap.Field {
	return zap.String(key, val)
}

func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

func Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

func Uint(key string, val uint) zap.Field {
	return zap.Uint(key, val)
}

func Duration(key string, val interface{}) zap.Field {
	if d, ok := val.(interface{ String() string }); ok {
		return zap.String(key, d.String())
	}
	return zap.Any(key, val)
}

func Error2(key string, err error) zap.Field {
	return zap.Error(err)
}

func ErrorField(err error) zap.Field {
	return zap.Error(err)
}

func Any(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}

func SetLogLevel(levelStr string) error {
	level, err := zapcore.ParseLevel(levelStr)
	if err != nil {
		return err
	}

	if GlobalLogger != nil {
		atomicLevel := zap.NewAtomicLevelAt(level)
		GlobalLogger = GlobalLogger.WithOptions(zap.IncreaseLevel(atomicLevel))
	}

	return nil
}

func GetEnvLogLevel() string {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		return "info"
	}
	return level
}

func GetEnvLogFormat() string {
	format := os.Getenv("LOG_FORMAT")
	if format == "" {
		return "text"
	}
	return format
}

func GetEnvLogOutput() string {
	output := os.Getenv("LOG_OUTPUT")
	if output == "" {
		return "stdout"
	}
	return output
}

// AsynqLogger implements the asynq.Logger interface using our project's logger
type AsynqLogger struct{}

// NewAsynqLogger creates a new logger adapter for asynq
func NewAsynqLogger() *AsynqLogger {
	return &AsynqLogger{}
}

// Debug logs a message at Debug level
func (l *AsynqLogger) Debug(args ...interface{}) {
	message := fmt.Sprint(args...)
	Debug(message, String("component", "asynq"))
}

// Info logs a message at Info level
func (l *AsynqLogger) Info(args ...interface{}) {
	message := fmt.Sprint(args...)
	Info(message, String("component", "asynq"))
}

// Warn logs a message at Warning level
func (l *AsynqLogger) Warn(args ...interface{}) {
	message := fmt.Sprint(args...)
	Warn(message, String("component", "asynq"))
}

// Error logs a message at Error level
func (l *AsynqLogger) Error(args ...interface{}) {
	message := fmt.Sprint(args...)
	Error(message, String("component", "asynq"))
}

// Fatal logs a message at Fatal level and process will exit with status set to 1
func (l *AsynqLogger) Fatal(args ...interface{}) {
	message := fmt.Sprint(args...)
	Fatal(message, String("component", "asynq"))
}

// RateLimiter provides log rate limiting functionality
type RateLimiter struct {
	mu       sync.RWMutex
	counters map[string]*logCounter
	interval time.Duration
	maxLogs  int
}

type logCounter struct {
	count     int
	lastReset time.Time
	lastLog   time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(interval time.Duration, maxLogs int) *RateLimiter {
	return &RateLimiter{
		counters: make(map[string]*logCounter),
		interval: interval,
		maxLogs:  maxLogs,
	}
}

// ShouldLog checks if a log with the given key should be logged
func (rl *RateLimiter) ShouldLog(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	counter, exists := rl.counters[key]

	if !exists {
		rl.counters[key] = &logCounter{
			count:     1,
			lastReset: now,
			lastLog:   now,
		}
		return true
	}

	// Reset counter if interval has passed
	if now.Sub(counter.lastReset) >= rl.interval {
		counter.count = 1
		counter.lastReset = now
		counter.lastLog = now
		return true
	}

	// Check if we've exceeded the limit
	if counter.count >= rl.maxLogs {
		return false
	}

	counter.count++
	counter.lastLog = now
	return true
}

// GetSuppressedCount returns the number of suppressed logs for a key
func (rl *RateLimiter) GetSuppressedCount(key string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	counter, exists := rl.counters[key]
	if !exists {
		return 0
	}

	if counter.count > rl.maxLogs {
		return counter.count - rl.maxLogs
	}
	return 0
}

// Global rate limiter instance
var globalRateLimiter = NewRateLimiter(1*time.Minute, 3)

// ErrorWithRateLimit logs an error with rate limiting
func ErrorWithRateLimit(key, message string, fields ...zap.Field) {
	if globalRateLimiter.ShouldLog(key) {
		suppressed := globalRateLimiter.GetSuppressedCount(key)
		if suppressed > 0 {
			fields = append(fields, Int("suppressed_logs", suppressed))
		}
		Error(message, fields...)
	}
}

// WarnWithRateLimit logs a warning with rate limiting
func WarnWithRateLimit(key, message string, fields ...zap.Field) {
	if globalRateLimiter.ShouldLog(key) {
		suppressed := globalRateLimiter.GetSuppressedCount(key)
		if suppressed > 0 {
			fields = append(fields, Int("suppressed_logs", suppressed))
		}
		Warn(message, fields...)
	}
}

// SetRateLimiter allows customizing the global rate limiter
func SetRateLimiter(interval time.Duration, maxLogs int) {
	globalRateLimiter = NewRateLimiter(interval, maxLogs)
}
