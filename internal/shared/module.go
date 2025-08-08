package shared

import (
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"linke/internal/shared/cache"
	"linke/internal/shared/config"
	"linke/internal/shared/database"
	"linke/internal/shared/events"
	"linke/internal/shared/framework"
	"linke/internal/shared/logger"
	"linke/internal/shared/queue"

	"gorm.io/gorm"
)

// frameworkLoggerAdapter adapts logger.Logger to framework.Logger interface
type frameworkLoggerAdapter struct {
	logger.Logger
}

func (f *frameworkLoggerAdapter) With(fields ...zap.Field) framework.Logger {
	return &frameworkLoggerAdapter{f.Logger.With(fields...)}
}

// Module 共享基础设施模块
// 提供配置管理、数据库连接、日志系统、队列系统、缓存系统等基础设施服务
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

	// 事件系统
	fx.Provide(
		// Event store
		func(db *gorm.DB) events.EventStore {
			return events.NewDatabaseEventStore(db)
		},
		// Event bus
		fx.Annotate(
			events.NewEnhancedEventBus,
			fx.As(new(events.EventBus)),
		),
		// Async event processor
		func(taskQueue *queue.TaskQueue, eventStore events.EventStore, eventBus events.EventBus) *events.AsyncEventProcessor {
			return events.NewAsyncEventProcessor(taskQueue, eventStore, eventBus, events.DefaultRetryConfig())
		},
		// Async event bus
		func(eventBus events.EventBus, asyncProcessor *events.AsyncEventProcessor) *events.AsyncEventBus {
			return events.NewAsyncEventBus(eventBus, asyncProcessor)
		},
		// Cross-domain event handlers - skip for now, will be created via dependency injection
		func() *events.CrossDomainEventHandlers {
			return nil // NewCrossDomainEventHandlers requires service dependencies
		},
		// Notification handler
		func() *events.NotificationHandler {
			mockNotificationService := &events.MockNotificationService{}
			return events.NewNotificationHandler(mockNotificationService)
		},
	),

	// 缓存系统
	cache.Module,

	// 事件系统初始化
	fx.Invoke(func(
		eventBus events.EventBus,
		crossDomainHandlers *events.CrossDomainEventHandlers,
		notificationHandler *events.NotificationHandler,
		asyncProcessor *events.AsyncEventProcessor,
		taskProcessor *queue.TaskProcessor,
		loggerSvc logger.Logger,
	) {
		// Initialize global event bus
		events.InitEventBus(eventBus)

		// Register cross-domain event handlers
		if err := crossDomainHandlers.RegisterCrossDomainHandlers(eventBus); err != nil {
			loggerSvc.Error("Failed to register cross-domain event handlers", logger.ErrorField(err))
		}

		// Register notification handler
		if err := eventBus.Subscribe(notificationHandler.EventTypes(), notificationHandler); err != nil {
			loggerSvc.Error("Failed to register notification handler", logger.ErrorField(err))
		}

		// Register event processing handlers with the task processor
		events.RegisterEventHandlers(taskProcessor, asyncProcessor)

		loggerSvc.Info("Event system initialized successfully")
	}),

	// 共享基础设施初始化
	fx.Invoke(func(cfg *config.Config, loggerSvc logger.Logger, cacheManager cache.CacheManager) {
		loggerSvc.Info("Shared infrastructure module initialized",
			logger.String("log_level", cfg.Log.Level),
			logger.String("log_format", cfg.Log.Format),
			logger.String("cache_metrics", fmt.Sprintf("%v", cfg.Cache.EnableMetrics)),
		)
	}),
)
