package shared

import (
	"go.uber.org/fx"
	"go.uber.org/zap"

	"linke/internal/shared/config"
	"linke/internal/shared/database"
	"linke/internal/shared/framework"
	"linke/internal/shared/logger"
	"linke/internal/shared/queue"
)

// frameworkLoggerAdapter adapts logger.Logger to framework.Logger interface
type frameworkLoggerAdapter struct {
	logger.Logger
}

func (f *frameworkLoggerAdapter) With(fields ...zap.Field) framework.Logger {
	return &frameworkLoggerAdapter{f.Logger.With(fields...)}
}

// Module 共享基础设施模块
// 提供配置管理、数据库连接、日志系统、队列系统等基础设施服务
var Module = fx.Module("shared",
	// 配置管理
	fx.Provide(config.LoadConfig),

	// 数据库连接
	fx.Provide(database.NewDatabase),

	// 日志系统
	fx.Provide(
		func(cfg *config.Config) logger.Logger {
			// 初始化日志
			if err := logger.InitLogger(logger.LogConfig{
				Level:  cfg.Log.Level,
				Format: cfg.Log.Format,
				Output: cfg.Log.Output,
			}); err != nil {
				panic("Failed to initialize logger: " + err.Error())
			}
			return logger.GetGlobalLogger()
		},
		fx.Annotate(
			func(zapLogger logger.Logger) framework.Logger {
				// 将 logger.Logger 包装为 framework.Logger
				return &frameworkLoggerAdapter{zapLogger}
			},
			fx.As(new(framework.Logger)),
		),
	),

	// 队列系统
	fx.Provide(
		queue.NewTaskQueue,
		queue.NewTaskProcessor,
	),

	// 共享基础设施初始化
	fx.Invoke(func(cfg *config.Config, loggerSvc logger.Logger) {
		loggerSvc.Info("Shared infrastructure module initialized",
			logger.String("log_level", cfg.Log.Level),
			logger.String("log_format", cfg.Log.Format),
		)
	}),
)