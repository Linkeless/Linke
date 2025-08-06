package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	// 标准库
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	// 新的 shared 模块
	"linke/internal/shared/config"
	"linke/internal/shared/database"
	"linke/internal/shared/framework"
	loggerPkg "linke/internal/shared/logger"
	"linke/internal/shared/queue"

	// Redis
	"github.com/go-redis/redis/v8"

	// GORM
	"gorm.io/gorm"

	// Asynq
	"github.com/hibiken/asynq"

	// 共享模块
	"linke/internal/shared/cache"
	"linke/internal/shared/events"
	"linke/internal/shared/versioning"

	// Service interface imports for event handlers
	userInterfaces "linke/internal/domains/user/usecases/interfaces"
	subscriptionInterfaces "linke/internal/domains/subscription/usecases/interfaces"
	paymentInterfaces "linke/internal/domains/payment/usecases/interfaces"
	invoiceInterfaces "linke/internal/domains/invoice/usecases/interfaces"
	serverInterfaces "linke/internal/domains/server/usecases/interfaces"

	// 业务领域模块
	authDomain "linke/internal/domains/auth"
	couponDomain "linke/internal/domains/coupon"
	invoiceDomain "linke/internal/domains/invoice"
	paymentDomain "linke/internal/domains/payment"
	referralDomain "linke/internal/domains/referral"
	serverDomain "linke/internal/domains/server"
	subscriptionDomain "linke/internal/domains/subscription"
	ticketDomain "linke/internal/domains/ticket"
	userDomain "linke/internal/domains/user"

	// 应用层
	applicationLayer "linke/internal/application"
	"linke/internal/application/server"

)

// frameworkLogger 定义 framework.Logger 接口的类型别名
type frameworkLogger = framework.Logger

// frameworkLoggerAdapter 适配器，将 loggerPkg.Logger 适配为 framework.Logger
type frameworkLoggerAdapter struct {
	loggerPkg.Logger
}

func (f *frameworkLoggerAdapter) With(fields ...zap.Field) framework.Logger {
	return &frameworkLoggerAdapter{f.Logger.With(fields...)}
}

// NewApplication 创建并配置 Fx 应用
func NewApplication() *fx.App {
	return fx.New(
		// 配置管理
		fx.Provide(func() (*config.Config, error) {
			cfg := config.LoadConfig()
			if err := config.ValidateConfig(cfg); err != nil {
				return nil, fmt.Errorf("configuration validation failed: %w", err)
			}
			return cfg, nil
		}),

		// 数据库连接
		fx.Provide(func(cfg *config.Config) (*database.Database, error) {
			db, err := database.NewDatabase(cfg)
			if err != nil {
				return nil, fmt.Errorf("database connection failed - host: %s:%s, database: %s, error: %w",
					cfg.Database.Host, cfg.Database.Port, cfg.Database.Name, err)
			}
			return db, nil
		}),

		// 提供 GORM DB 实例
		fx.Provide(func(db *database.Database) *gorm.DB {
			return db.GetDB()
		}),

		// 日志系统
		fx.Provide(func(cfg *config.Config) (loggerPkg.Logger, error) {
			if err := loggerPkg.InitLogger(loggerPkg.LogConfig{
				Level:  cfg.Log.Level,
				Format: cfg.Log.Format,
				Output: cfg.Log.Output,
			}); err != nil {
				return nil, fmt.Errorf("logger initialization failed - level: %s, format: %s, output: %s, error: %w",
					cfg.Log.Level, cfg.Log.Format, cfg.Log.Output, err)
			}
			return loggerPkg.GetGlobalLogger(), nil
		}),

		// 提供 framework.Logger 接口适配
		fx.Provide(
			fx.Annotate(
				func(logger loggerPkg.Logger) frameworkLogger {
					return &frameworkLoggerAdapter{logger}
				},
				fx.As(new(frameworkLogger)),
			),
		),


		// 自定义 Fx 日志系统 - 统一日志输出格式
		fx.WithLogger(func(logger loggerPkg.Logger) fxevent.Logger {
			return loggerPkg.NewFxAdapter(logger)
		}),

		// Redis 客户端提供者 (从 Database 中提取)
		fx.Provide(func(db *database.Database, cfg *config.Config) (*redis.Client, error) {
			client := db.GetRedis()
			if client == nil {
				return nil, fmt.Errorf("redis client initialization failed - host: %s:%s, db: %d",
					cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.DB)
			}
			// 测试Redis连接
			ctx := context.Background()
			if err := client.Ping(ctx).Err(); err != nil {
				return nil, fmt.Errorf("redis connection test failed - host: %s:%s, db: %d, error: %w",
					cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.DB, err)
			}
			return client, nil
		}),

		// 队列系统
		fx.Provide(
			func(redisClient *redis.Client, cfg *config.Config) (*queue.TaskQueue, error) {
				taskQueue := queue.NewTaskQueue(redisClient)
				if taskQueue == nil {
					return nil, fmt.Errorf("task queue initialization failed - redis: %s:%s, db: %d",
						cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.DB)
				}
				return taskQueue, nil
			},
			func(redisClient *redis.Client, cfg *config.Config) (*queue.TaskProcessor, error) {
				taskProcessor := queue.NewTaskProcessor(redisClient)
				if taskProcessor == nil {
					return nil, fmt.Errorf("task processor initialization failed - redis: %s:%s, db: %d",
						cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.DB)
				}
				return taskProcessor, nil
			},
			// Asynq Client 提供者 - 为需要直接使用 asynq.Client 的服务提供
			func(redisClient *redis.Client, cfg *config.Config) (*asynq.Client, error) {
				asynqClient := asynq.NewClient(asynq.RedisClientOpt{
					Addr:     redisClient.Options().Addr,
					Password: redisClient.Options().Password,
					DB:       redisClient.Options().DB,
				})
				if asynqClient == nil {
					return nil, fmt.Errorf("asynq client initialization failed - redis: %s:%s, db: %d",
						cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.DB)
				}
				return asynqClient, nil
			},
		),

		// 缓存系统
		cache.Module,

		// 版本控制系统
		versioning.Module,

		// 事件系统
		fx.Provide(
			// Event bus
			fx.Annotate(
				events.NewEnhancedEventBus,
				fx.As(new(events.EventBus)),
			),
			// Event store
			func(db *gorm.DB) events.EventStore {
				return events.NewDatabaseEventStore(db)
			},
			// Async event processor
			func(taskQueue *queue.TaskQueue, eventStore events.EventStore, eventBus events.EventBus) *events.AsyncEventProcessor {
				return events.NewAsyncEventProcessor(taskQueue, eventStore, eventBus, events.DefaultRetryConfig())
			},
			// Usage monitor for real-time traffic monitoring
			func(userSubscriptionService subscriptionInterfaces.UserSubscriptionService, eventBus events.EventBus) *events.UsageMonitor {
				return events.NewUsageMonitor(userSubscriptionService, eventBus)
			},
			// EventCacheStore adapter for event handlers
			func(cacheStore cache.CacheStore) events.EventCacheStore {
				return events.NewCacheStoreAdapter(
					cacheStore.Set,
					cacheStore.Get,
					cacheStore.Delete,
					cacheStore.Exists,
					cacheStore.SetJSON,
					cacheStore.GetJSON,
					cacheStore.DeletePattern,
				)
			},
		),

		// 业务领域模块
		userDomain.Module,
		authDomain.Module,
		couponDomain.Module,
		subscriptionDomain.Module,
		paymentDomain.Module,
		serverDomain.Module,
		invoiceDomain.Module,
		ticketDomain.Module,
		referralDomain.Module,

		// 应用层模块
		applicationLayer.Module,

		// 应用处理器
		fx.Provide(
			server.NewAppHandler,
			server.NewTaskHandler,
		),

		// HTTP 服务器
		fx.Provide(server.NewHTTPServer),

		// 初始化事件系统
		fx.Invoke(func(
			eventBus events.EventBus,
			asyncProcessor *events.AsyncEventProcessor,
			taskProcessor *queue.TaskProcessor,
			taskQueue *queue.TaskQueue,
			eventCacheStore events.EventCacheStore,
			usageMonitor *events.UsageMonitor,
			logger loggerPkg.Logger,
			// All domain services needed for cross-domain event handlers
			userService userInterfaces.UserService,
			userSubscriptionService subscriptionInterfaces.UserSubscriptionService,
			subscriptionOrderService subscriptionInterfaces.SubscriptionOrderService,
			paymentService paymentInterfaces.PaymentService,
			invoiceService invoiceInterfaces.InvoiceService,
			shadowsocksServerService serverInterfaces.ShadowsocksServerService,
		) {
			// Initialize global event bus
			events.InitEventBus(eventBus)

			// Register cross-domain event handlers with all required services
			crossDomainHandlers := events.NewCrossDomainEventHandlers(
				userService,
				userSubscriptionService,
				subscriptionOrderService,
				paymentService,
				invoiceService,
				shadowsocksServerService,
				eventCacheStore,
				taskQueue,
			)
			if err := crossDomainHandlers.RegisterCrossDomainHandlers(eventBus); err != nil {
				logger.Error("Failed to register cross-domain event handlers", zap.Error(err))
			} else {
				logger.Info("Cross-domain event handlers registered successfully")
			}

			// Register notification handler
			notificationHandler := events.NewNotificationHandler()
			if err := eventBus.Subscribe(notificationHandler.EventTypes(), notificationHandler); err != nil {
				logger.Error("Failed to register notification handler", zap.Error(err))
			} else {
				logger.Info("Notification handler registered successfully")
			}

			// Register event processing handlers with the task processor
			events.RegisterEventHandlers(taskProcessor, asyncProcessor)

			logger.Info("Event system initialized successfully with full cross-domain integration",
				zap.Bool("usage_monitor_enabled", usageMonitor != nil),
			)
		}),

		// 启动服务
		fx.Invoke(StartServices),
	)
}

// StartServices 启动服务器
func StartServices(
	lc fx.Lifecycle,
	httpServer *server.HTTPServer,
	taskProcessor *queue.TaskProcessor,
	logger loggerPkg.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// 注册任务处理器
			taskProcessor.RegisterHandler("email", queue.EmailTaskHandler)
			taskProcessor.RegisterHandler("notification", queue.NotificationTaskHandler)
			taskProcessor.RegisterHandler("data_processing", queue.DataProcessingTaskHandler)

			// 启动任务处理器
			go func() {
				logger.Info("Starting task processor")
				if err := taskProcessor.Start(ctx); err != nil {
					logger.Error("Task processor startup failed",
						zap.Error(err),
						zap.String("component", "task_processor"),
						zap.String("action", "start"))
					fmt.Printf("❌ Task processor failed to start: %v\n", err)
					os.Exit(1)
				}
			}()

			// 启动 HTTP 服务器
			go func() {
				logger.Info("Starting HTTP server")
				if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
					logger.Fatal("HTTP server startup failed",
						zap.Error(err),
						zap.String("component", "http_server"))
					fmt.Printf("❌ HTTP server failed to start: %v\n", err)
					fmt.Printf("   Common causes:\n")
					fmt.Printf("   - Port already in use\n")
					fmt.Printf("   - Insufficient permissions to bind to port\n")
					fmt.Printf("   - Invalid port number\n")
					os.Exit(1)
				}
			}()

			logger.Info("Application started successfully with new VSA + Clean Architecture")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down application")

			// 停止任务处理器
			taskProcessor.Stop()

			logger.Info("Application shutdown complete")
			return nil
		},
	})

	// 监听中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("Received shutdown signal")
	}()
}
