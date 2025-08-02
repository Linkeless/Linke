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

	// Gin 和 Swagger
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

	// 新的 shared 模块
	"linke/internal/shared/config"
	"linke/internal/shared/database"
	loggerPkg "linke/internal/shared/logger"
	"linke/internal/shared/queue"

	// Redis
	"github.com/go-redis/redis/v8"

	// Migration 已移至 shared/database

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
		fx.Provide(config.LoadConfig),

		// 数据库连接
		fx.Provide(database.NewDatabase),

		// 日志系统
		fx.Provide(func(cfg *config.Config) loggerPkg.Logger {
			if err := loggerPkg.InitLogger(loggerPkg.LogConfig{
				Level:  cfg.Log.Level,
				Format: cfg.Log.Format,
				Output: cfg.Log.Output,
			}); err != nil {
				panic("Failed to initialize logger: " + err.Error())
			}
			return loggerPkg.GetGlobalLogger()
		}),

		// 自定义 Fx 日志系统 - 统一日志输出格式
		fx.WithLogger(func(logger loggerPkg.Logger) fxevent.Logger {
			return loggerPkg.NewFxAdapter(logger)
		}),

		// Redis 客户端提供者 (从 Database 中提取)
		fx.Provide(func(db *database.Database) *redis.Client {
			return db.GetRedis()
		}),

		// 队列系统
		fx.Provide(
			queue.NewTaskQueue,
			queue.NewTaskProcessor,
		),

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
					loggerPkg.Error("Task processor failed", loggerPkg.ErrorField(err))
				}
			}()

			// 启动 HTTP 服务器
			go func() {
				if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
					loggerPkg.Fatal("Failed to start HTTP server", loggerPkg.ErrorField(err))
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
	cfg := config.LoadConfig()

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
	if err := migrationService.ValidateConnection(); err != nil {
		loggerPkg.Fatal("Database connection failed", loggerPkg.ErrorField(err))
	}

	// 确定要执行的命令
	command := migrateCommand
	if runMigration && command == "" {
		command = "up"
	}

	// 执行迁移命令
	if err := executeMigrationCommand(migrationService, command, migrateVersion, migrateSteps); err != nil {
		loggerPkg.Fatal("Migration command failed", loggerPkg.ErrorField(err), loggerPkg.String("command", command))
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
