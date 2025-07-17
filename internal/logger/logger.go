package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

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
	case "text", "console":
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	default:
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
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

	logger, err := zapConfig.Build(zap.AddCallerSkip(1))
	if err != nil {
		return err
	}

	Logger = logger
	zap.ReplaceGlobals(logger)

	return nil
}

func GetLogger() *zap.Logger {
	if Logger == nil {
		config := LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		}
		if err := InitLogger(config); err != nil {
			panic("Failed to initialize logger: " + err.Error())
		}
	}
	return Logger
}

func Sync() error {
	if Logger != nil {
		return Logger.Sync()
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

func Any(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}

func SetLogLevel(levelStr string) error {
	level, err := zapcore.ParseLevel(levelStr)
	if err != nil {
		return err
	}

	if Logger != nil {
		atomicLevel := zap.NewAtomicLevelAt(level)
		Logger = Logger.WithOptions(zap.IncreaseLevel(atomicLevel))
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