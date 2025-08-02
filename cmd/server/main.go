package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	// 标准库
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	// Gin 和 Swagger
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

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

	// Migration 已移至 shared/database

	// 业务领域模块
	authDomain "linke/internal/domains/auth"
	couponDomain "linke/internal/domains/coupon"
	invoiceDomain "linke/internal/domains/invoice"
	paymentDomain "linke/internal/domains/payment"
	// referralDomain "linke/internal/domains/referral"
	serverDomain "linke/internal/domains/server"
	subscriptionDomain "linke/internal/domains/subscription"
	userDomain "linke/internal/domains/user"

	// Handler 导入
	authHandlers "linke/internal/domains/auth/handlers"
	invoiceHandlers "linke/internal/domains/invoice/handlers"
	subscriptionHandlers "linke/internal/domains/subscription/handlers"
	userHandlers "linke/internal/domains/user/handlers"
	// paymentHandlers "linke/internal/domains/payment/handlers" // TODO: 待实现
	// serverHandlers "linke/internal/domains/server/handlers"   // TODO: 待实现

	// 应用层
	applicationLayer "linke/internal/application"

	// Invite code service interface (temporary import)
	referralEntities "linke/internal/domains/referral/entities"
	referralInterfaces "linke/internal/domains/referral/usecases/interfaces"

	// Swagger 文档
	_ "linke/docs"
)

// @title Linke API
// @version 1.0
// @description A comprehensive service management platform with subscription-based billing, user management, and server administration. Features include OAuth2 authentication, traffic subscription management, multi-gateway payments, referral programs, and customer support system.
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// 类型定义

// frameworkLogger 定义 framework.Logger 接口的类型别名
type frameworkLogger = framework.Logger

// frameworkLoggerAdapter 适配器，将 loggerPkg.Logger 适配为 framework.Logger
type frameworkLoggerAdapter struct {
	loggerPkg.Logger
}

func (f *frameworkLoggerAdapter) With(fields ...zap.Field) framework.Logger {
	return &frameworkLoggerAdapter{f.Logger.With(fields...)}
}

// stubInviteCodeService provides a minimal stub implementation of InviteCodeService
// This is a temporary implementation to allow the application to start without the full referral system
type stubInviteCodeService struct{}

func newStubInviteCodeService() referralInterfaces.InviteCodeService {
	return &stubInviteCodeService{}
}

func (s *stubInviteCodeService) CreateInviteCode(ctx context.Context, createdByID uint, req *referralInterfaces.CreateInviteCodeRequest) (*referralEntities.InviteCode, error) {
	return nil, fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) GetInviteCode(ctx context.Context, codeID uint) (*referralEntities.InviteCode, error) {
	return nil, fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) GetInviteCodeByCode(ctx context.Context, code string) (*referralEntities.InviteCode, error) {
	return nil, fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) UpdateInviteCode(ctx context.Context, inviteCodeID uint, req *referralInterfaces.UpdateInviteCodeRequest) (*referralEntities.InviteCode, error) {
	return nil, fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) DeleteInviteCode(ctx context.Context, inviteCodeID uint) error {
	return fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) GetInviteCodes(ctx context.Context, req *referralInterfaces.GetInviteCodesRequest) ([]*referralEntities.InviteCode, int64, error) {
	return nil, 0, fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) GetUserInviteCodes(ctx context.Context, userID uint, limit, offset int) ([]*referralEntities.InviteCode, int64, error) {
	return nil, 0, fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) ValidateInviteCode(ctx context.Context, code string) (*referralEntities.InviteCode, error) {
	// Return nil to indicate the code is not valid (but don't fail startup)
	return nil, fmt.Errorf("invite code validation not available")
}

func (s *stubInviteCodeService) UseInviteCode(ctx context.Context, code string, userID uint, ipAddress, userAgent string) (*referralEntities.InviteCodeUsage, error) {
	return nil, fmt.Errorf("invite code usage not implemented")
}

func (s *stubInviteCodeService) ActivateInviteCode(ctx context.Context, inviteCodeID uint) error {
	return fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) DeactivateInviteCode(ctx context.Context, inviteCodeID uint) error {
	return fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) ExpireInviteCode(ctx context.Context, inviteCodeID uint) error {
	return fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) GetInviteCodeUsage(ctx context.Context, inviteCodeID uint, limit, offset int) ([]*referralEntities.InviteCodeUsage, int64, error) {
	return nil, 0, fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) GetUserInviteCodeUsage(ctx context.Context, userID uint, limit, offset int) ([]*referralEntities.InviteCodeUsage, int64, error) {
	return nil, 0, fmt.Errorf("invite code service not implemented")
}

func (s *stubInviteCodeService) GetInviteCodeStatistics(ctx context.Context, inviteCodeID uint) (map[string]interface{}, error) {
	return map[string]interface{}{
		"total_codes": 0,
		"used_codes": 0,
		"active_codes": 0,
	}, nil
}

func (s *stubInviteCodeService) GetUserInviteCodeStatistics(ctx context.Context, userID uint) (map[string]interface{}, error) {
	return map[string]interface{}{
		"total_codes": 0,
		"used_codes": 0,
		"active_codes": 0,
	}, nil
}

func (s *stubInviteCodeService) GenerateInviteCode() (string, error) {
	return "", fmt.Errorf("invite code generation not implemented")
}

// AppHandler 简单的应用处理器
type AppHandler struct {
	logger   loggerPkg.Logger
	database *database.Database
}

// TaskHandler 简单的任务处理器
type TaskHandler struct {
	taskQueue *queue.TaskQueue
	logger    loggerPkg.Logger
}

// HTTPServer HTTP 服务器结构
type HTTPServer struct {
	*gin.Engine
	logger loggerPkg.Logger
	config *config.Config
}

// 构造函数

// NewAppHandler 创建应用处理器
func NewAppHandler(logger loggerPkg.Logger, database *database.Database) *AppHandler {
	return &AppHandler{
		logger:   logger,
		database: database,
	}
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(taskQueue *queue.TaskQueue, logger loggerPkg.Logger) *TaskHandler {
	return &TaskHandler{
		taskQueue: taskQueue,
		logger:    logger,
	}
}

// NewHTTPServer 创建 HTTP 服务器
func NewHTTPServer(
	cfg *config.Config,
	logger loggerPkg.Logger,
	appHandler *AppHandler,
	taskHandler *TaskHandler,
	// 业务领域 handlers
	authHandler *authHandlers.AuthHandler,
	userProfileHandler *userHandlers.UserProfileHandler,
	adminUserHandler *userHandlers.AdminUserHandler,
	subscriptionOrderHandler *subscriptionHandlers.SubscriptionOrderHandler,
	invoiceHandler *invoiceHandlers.InvoiceHandler,
	// TODO: 待实现的handlers
	// paymentHandler *paymentHandlers.PaymentHandler,
	// serverHandler *serverHandlers.ServerAPIHandler,
) *HTTPServer {
	// 设置 Gin 模式
	if cfg.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// 添加中间件
	router.Use(gin.Recovery())

	// 健康检查端点
	router.GET("/health", appHandler.HealthCheck)

	// API 路由组
	apiV1 := router.Group("/api/v1")

	// 应用层路由
	appGroup := apiV1.Group("/app")
	{
		appGroup.GET("/system/health", appHandler.HealthCheck)
	}

	// 任务路由
	taskGroup := apiV1.Group("/tasks")
	{
		taskGroup.POST("", taskHandler.CreateTask)
	}

	// 业务领域路由注册

	// 认证路由 (/api/v1/auth)
	authGroup := apiV1.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.LoginLocal)
		authGroup.POST("/logout", authHandler.Logout)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.GET("/callback", authHandler.Callback)
		authGroup.GET("/providers", authHandler.GetProviders)
		authGroup.GET("/profile", authHandler.GetProfile)
		authGroup.POST("/change-password", authHandler.ChangePassword)
	}

	// 用户路由 (/api/v1/user)
	userGroup := apiV1.Group("/user")
	{
		userGroup.GET("/profile", userProfileHandler.GetProfile)
		userGroup.PUT("/profile", userProfileHandler.UpdateProfile)
		userGroup.POST("/change-password", userProfileHandler.ChangePassword)
	}

	// 管理员用户路由 (/api/v1/admin/users)
	adminUserGroup := apiV1.Group("/admin/users")
	{
		adminUserGroup.POST("", adminUserHandler.CreateUser)
		adminUserGroup.GET("", adminUserHandler.ListUsers)
		adminUserGroup.GET("/:id", adminUserHandler.GetUser)
		adminUserGroup.PUT("/:id", adminUserHandler.UpdateUser)
		adminUserGroup.DELETE("/:id", adminUserHandler.SoftDeleteUser)
		adminUserGroup.POST("/:id/restore", adminUserHandler.RestoreUser)
		adminUserGroup.GET("/stats", adminUserHandler.GetUserStats)
	}

	// 订阅路由 (/api/v1/subscription)
	subscriptionGroup := apiV1.Group("/subscription")
	{
		subscriptionGroup.POST("/orders", subscriptionOrderHandler.CreateSubscriptionOrder)
	}

	// 发票路由 (/api/v1/invoice) - 使用RegisterRoutes方法
	invoiceHandler.RegisterRoutes(apiV1)

	// TODO: 以下路由需要等待handler方法实现后启用
	// 支付路由 (/api/v1/payment)
	// paymentGroup := apiV1.Group("/payment")
	// {
	//     paymentGroup.POST("/create", paymentHandler.CreatePayment)
	//     paymentGroup.GET("/:id", paymentHandler.GetPayment)
	//     paymentGroup.POST("/webhook/:gateway", paymentHandler.HandleWebhook)
	// }

	// 服务器路由 (/api/v1/server)
	// serverGroup := apiV1.Group("/server")
	// {
	//     serverGroup.GET("/groups", serverHandler.GetServerGroups)
	//     serverGroup.POST("/groups", serverHandler.CreateServerGroup)
	//     serverGroup.PUT("/groups/:id", serverHandler.UpdateServerGroup)
	//     serverGroup.DELETE("/groups/:id", serverHandler.DeleteServerGroup)
	// }

	// Swagger 文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return &HTTPServer{
		Engine: router,
		logger: logger,
		config: cfg,
	}
}

// 方法

// HealthCheck 健康检查
func (h *AppHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	health := h.database.HealthCheck(ctx)
	result := map[string]interface{}{
		"status":       "healthy",
		"database":     health,
		"architecture": "VSA + Clean Architecture",
		"framework":    "Fx Dependency Injection",
		"note":         "New architecture successfully implemented",
	}

	c.JSON(http.StatusOK, result)
}

// CreateTask 创建任务
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req struct {
		Type    string                 `json:"type" binding:"required"`
		Payload map[string]interface{} `json:"payload" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := &queue.Task{
		ID:       fmt.Sprintf("task-%d", os.Getpid()),
		Type:     req.Type,
		Payload:  req.Payload,
		MaxRetry: 3,
	}

	if err := h.taskQueue.Enqueue(c.Request.Context(), "default", task); err != nil {
		h.logger.Error("Failed to enqueue task", loggerPkg.ErrorField(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue task"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Task enqueued successfully",
		"task_id": task.ID,
	})
}

// Start 启动 HTTP 服务器
func (s *HTTPServer) Start() error {
	addr := fmt.Sprintf(":%s", s.config.Server.Port)
	s.logger.Info("Starting HTTP server", loggerPkg.String("addr", addr))

	return http.ListenAndServe(addr, s.Engine)
}

func main() {
	// Parse command line flags
	var (
		runMigration    = flag.Bool("migrate", false, "Run database migrations and exit (alias for -migrate-command=up)")
		migrateCommand  = flag.String("migrate-command", "", "Migration command (up, down, reset, status, force, goto, steps, list, fix-dirty)")
		migrateVersion  = flag.String("migrate-version", "", "Target version for goto/force commands")
		migrateSteps    = flag.String("migrate-steps", "", "Number of steps for steps command")
		showMigrateHelp = flag.Bool("migrate-help", false, "Show migration help and exit")
	)
	flag.Parse()

	// Show migration help if requested
	if *showMigrateHelp {
		showMigrationHelp()
		return
	}

	// 如果是迁移命令，直接处理后退出
	if *runMigration || *migrateCommand != "" {
		handleMigrationCommand(*runMigration, *migrateCommand, *migrateVersion, *migrateSteps)
		return
	}

	// 创建 Fx 应用
	app := fx.New(
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

		// 提供临时的邀请码服务实现
		fx.Provide(
			fx.Annotate(
				newStubInviteCodeService,
				fx.As(new(referralInterfaces.InviteCodeService)),
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
		),

		// 业务领域模块
		userDomain.Module,
		authDomain.Module,
		couponDomain.Module,
		// TODO: Fix referral service interface implementation
		// referralDomain.Module,
		subscriptionDomain.Module,
		paymentDomain.Module,
		serverDomain.Module,
		invoiceDomain.Module,

		// 应用层模块
		applicationLayer.Module,

		// 应用处理器
		fx.Provide(
			NewAppHandler,
			NewTaskHandler,
		),

		// HTTP 服务器
		fx.Provide(NewHTTPServer),

		// 启动服务
		fx.Invoke(startServer),
	)

	// 运行应用
	if err := app.Err(); err != nil {
		fmt.Printf("❌ Application startup failed\n")
		fmt.Printf("   Error: %v\n", err)
		fmt.Printf("   This error occurred during dependency injection initialization.\n")
		fmt.Printf("   Common causes:\n")
		fmt.Printf("   - Database connection failure (check DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME)\n")
		fmt.Printf("   - Redis connection failure (check REDIS_HOST, REDIS_PORT, REDIS_PASSWORD, REDIS_DB)\n")
		fmt.Printf("   - Invalid JWT_SECRET (must be 32+ characters)\n")
		fmt.Printf("   - Missing required environment variables\n")
		fmt.Printf("   - Configuration validation errors\n")
		fmt.Printf("\n   For detailed troubleshooting, run: make security-check\n")
		os.Exit(1)
	}
	app.Run()
}

// startServer 启动服务器
func startServer(
	lc fx.Lifecycle,
	httpServer *HTTPServer,
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
				loggerPkg.Info("Starting task processor")
				if err := taskProcessor.Start(ctx); err != nil {
					loggerPkg.Error("Task processor startup failed", 
						loggerPkg.ErrorField(err),
						loggerPkg.String("component", "task_processor"),
						loggerPkg.String("action", "start"))
					fmt.Printf("❌ Task processor failed to start: %v\n", err)
					os.Exit(1)
				}
			}()

			// 启动 HTTP 服务器
			go func() {
				addr := fmt.Sprintf(":%s", httpServer.config.Server.Port)
				loggerPkg.Info("Starting HTTP server", 
					loggerPkg.String("port", httpServer.config.Server.Port),
					loggerPkg.String("address", addr))
				if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
					loggerPkg.Fatal("HTTP server startup failed", 
						loggerPkg.ErrorField(err),
						loggerPkg.String("address", addr),
						loggerPkg.String("component", "http_server"))
					fmt.Printf("❌ HTTP server failed to start on %s: %v\n", addr, err)
					fmt.Printf("   Common causes:\n")
					fmt.Printf("   - Port %s already in use\n", httpServer.config.Server.Port)
					fmt.Printf("   - Insufficient permissions to bind to port\n")
					fmt.Printf("   - Invalid port number\n")
					os.Exit(1)
				}
			}()

			loggerPkg.Info("Application started successfully with new VSA + Clean Architecture")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			loggerPkg.Info("Shutting down application")

			// 停止任务处理器
			taskProcessor.Stop()

			loggerPkg.Info("Application shutdown complete")
			return nil
		},
	})

	// 监听中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		loggerPkg.Info("Received shutdown signal")
	}()
}

// handleMigrationCommand 处理迁移命令
func handleMigrationCommand(runMigration bool, migrateCommand, migrateVersion, migrateSteps string) {
	// 加载配置
	fmt.Println("Loading configuration for migration...")
	cfg := config.LoadConfig()
	if err := config.ValidateConfig(cfg); err != nil {
		fmt.Printf("❌ Configuration validation failed: %v\n", err)
		fmt.Printf("   Please check your environment variables and .env file\n")
		os.Exit(1)
	}

	// 初始化日志
	if err := loggerPkg.InitLogger(loggerPkg.LogConfig{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		Output: cfg.Log.Output,
	}); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer loggerPkg.Sync()

	loggerPkg.Info("Running migration command")

	// 创建迁移服务
	migrationService := database.NewMigrationService(
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
	)

	// 验证数据库连接
	loggerPkg.Info("Validating database connection", 
		loggerPkg.String("host", cfg.Database.Host),
		loggerPkg.String("port", cfg.Database.Port),
		loggerPkg.String("database", cfg.Database.Name))
	if err := migrationService.ValidateConnection(); err != nil {
		loggerPkg.Fatal("Database connection failed", 
			loggerPkg.ErrorField(err),
			loggerPkg.String("host", cfg.Database.Host),
			loggerPkg.String("port", cfg.Database.Port),
			loggerPkg.String("database", cfg.Database.Name))
		fmt.Printf("❌ Database connection failed\n")
		fmt.Printf("   Host: %s:%s\n", cfg.Database.Host, cfg.Database.Port)
		fmt.Printf("   Database: %s\n", cfg.Database.Name)
		fmt.Printf("   User: %s\n", cfg.Database.User)
		fmt.Printf("   Error: %v\n", err)
		fmt.Printf("   Common causes:\n")
		fmt.Printf("   - Database server not running\n")
		fmt.Printf("   - Incorrect credentials (DB_USER, DB_PASSWORD)\n")
		fmt.Printf("   - Network connectivity issues\n")
		fmt.Printf("   - Database does not exist\n")
		os.Exit(1)
	}

	// 确定要执行的命令
	command := migrateCommand
	if runMigration && command == "" {
		command = "up"
	}

	// 执行迁移命令
	if err := executeMigrationCommand(migrationService, command, migrateVersion, migrateSteps); err != nil {
		loggerPkg.Fatal("Migration command failed", 
			loggerPkg.ErrorField(err), 
			loggerPkg.String("command", command),
			loggerPkg.String("version", migrateVersion),
			loggerPkg.String("steps", migrateSteps))
		fmt.Printf("❌ Migration command '%s' failed: %v\n", command, err)
		fmt.Printf("   For help with migration commands, run: go run cmd/server/main.go -migrate-help\n")
		os.Exit(1)
	}

	loggerPkg.Info("Migration command completed, exiting...", loggerPkg.String("command", command))
}

// 工具函数
func showMigrationHelp() {
	fmt.Println("Database Migration Commands")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/server/main.go [migration-options]")
	fmt.Println("")
	fmt.Println("Migration Options:")
	fmt.Println("  -migrate                    Run database migrations and exit (same as -migrate-command=up)")
	fmt.Println("  -migrate-command=<command>  Execute specific migration command")
	fmt.Println("  -migrate-version=<N>        Target version for goto/force commands")
	fmt.Println("  -migrate-steps=<N>          Number of steps for steps command")
	fmt.Println("  -migrate-help               Show this help and exit")
	fmt.Println("")
	fmt.Println("Migration Commands:")
	fmt.Println("  up       - Run all pending migrations")
	fmt.Println("  down     - Rollback one migration")
	fmt.Println("  reset    - Drop all tables and re-run migrations (DANGEROUS!)")
	fmt.Println("  status   - Show current migration version")
	fmt.Println("  list     - List all applied migrations")
	fmt.Println("  force    - Force set migration version (use with caution)")
	fmt.Println("  goto     - Migrate to specific version")
	fmt.Println("  steps    - Run specific number of migration steps")
	fmt.Println("  fix-dirty - Fix dirty migration state (requires -migrate-version)")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/server/main.go -migrate")
	fmt.Println("  go run cmd/server/main.go -migrate-command=status")
	fmt.Println("  go run cmd/server/main.go -migrate-command=fix-dirty -migrate-version=9")
	fmt.Println("  go run cmd/server/main.go -migrate-command=goto -migrate-version=5")
	fmt.Println("  go run cmd/server/main.go -migrate-command=steps -migrate-steps=2")
	fmt.Println("")
	fmt.Println("Fixing Dirty Migration State:")
	fmt.Println("  If you see 'Dirty database version X' error:")
	fmt.Println("  go run cmd/server/main.go -migrate-command=fix-dirty -migrate-version=X")
}

func executeMigrationCommand(migrationService *database.MigrationService, command, version, steps string) error {
	switch command {
	case "up":
		loggerPkg.Info("Running database migrations...")
		return migrationService.Up()

	case "down":
		loggerPkg.Info("Rolling back one migration...")
		return migrationService.Down()

	case "reset":
		fmt.Print("WARNING: This will drop all tables and re-run migrations. Are you sure? (y/N): ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm == "y" || confirm == "Y" {
			loggerPkg.Info("Resetting database...")
			return migrationService.Reset()
		} else {
			loggerPkg.Info("Reset cancelled")
			return nil
		}

	case "status":
		status, err := migrationService.Status()
		if err != nil {
			return err
		}
		fmt.Println(status)
		return nil

	case "list":
		versions, err := migrationService.GetAppliedMigrations()
		if err != nil {
			return err
		}
		if len(versions) == 0 {
			fmt.Println("No migrations have been applied")
		} else {
			fmt.Println("Applied migrations:")
			for _, v := range versions {
				fmt.Printf("  - Version %d\n", v)
			}
		}
		return nil

	case "force":
		if version == "" {
			return fmt.Errorf("version is required for force command")
		}
		v, err := strconv.Atoi(version)
		if err != nil {
			return fmt.Errorf("invalid version number: %w", err)
		}
		loggerPkg.Warn("Forcing migration version", loggerPkg.Int("version", v))
		return migrationService.Force(v)

	case "goto":
		if version == "" {
			return fmt.Errorf("version is required for goto command")
		}
		v, err := strconv.ParseUint(version, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid version number: %w", err)
		}
		loggerPkg.Info("Migrating to specific version", loggerPkg.Uint("version", uint(v)))
		return migrationService.Goto(uint(v))

	case "steps":
		if steps == "" {
			return fmt.Errorf("steps is required for steps command")
		}
		s, err := strconv.Atoi(steps)
		if err != nil {
			return fmt.Errorf("invalid steps number: %w", err)
		}
		direction := "up"
		if s < 0 {
			direction = "down"
		}
		loggerPkg.Info("Running migration steps", loggerPkg.Int("steps", s), loggerPkg.String("direction", direction))
		return migrationService.Steps(s)

	case "fix-dirty":
		if version == "" {
			return fmt.Errorf("version is required for fix-dirty command. Use the version shown in the dirty database error")
		}
		v, err := strconv.Atoi(version)
		if err != nil {
			return fmt.Errorf("invalid version number: %w", err)
		}

		fmt.Printf("WARNING: This will force the migration version to %d without running the migration.\n", v)
		fmt.Print("This should only be used to fix a dirty migration state. Continue? (y/N): ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm == "y" || confirm == "Y" {
			loggerPkg.Warn("Fixing dirty migration state", loggerPkg.Int("version", v))
			if err := migrationService.Force(v); err != nil {
				return fmt.Errorf("failed to fix dirty migration: %w", err)
			}
			loggerPkg.Info("Dirty migration state fixed", loggerPkg.Int("version", v))
			fmt.Println("Migration state fixed. You can now run migrations again.")
			return nil
		} else {
			loggerPkg.Info("Fix dirty operation cancelled")
			return nil
		}

	case "":
		return fmt.Errorf("migration command is required")

	default:
		return fmt.Errorf("unknown migration command: %s", command)
	}
}
