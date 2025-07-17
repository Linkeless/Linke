package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"linke/config"
	"linke/internal/handler"
	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/migration"
	"linke/internal/queue"
	"linke/internal/repository"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	
	_ "linke/docs"
	_ "linke/internal/handler"
)

// @title Linke API
// @version 1.0
// @description A comprehensive API server with Gin, GORM, Redis, OAuth2, and invite code system. Supports user authentication, profile management, and invite-based registration.
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	cfg := config.LoadConfig()

	if err := logger.InitLogger(logger.LogConfig{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		Output: cfg.Log.Output,
	}); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("Starting Linke server",
		logger.String("version", "1.0"),
		logger.String("log_level", cfg.Log.Level),
		logger.String("log_format", cfg.Log.Format),
	)

	db, err := repository.NewDatabase(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize database", logger.Error2("error", err))
	}
	defer db.Close()

	if err := migration.Migrate(db.DB); err != nil {
		logger.Fatal("Failed to migrate database", logger.Error2("error", err))
	}

	taskQueue := queue.NewTaskQueue(db.Redis)
	processor := queue.NewTaskProcessor(taskQueue)
	processor.RegisterHandler("email", queue.EmailTaskHandler)
	processor.RegisterHandler("notification", queue.NotificationTaskHandler)
	processor.RegisterHandler("data_processing", queue.DataProcessingTaskHandler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go processor.ProcessTasks(ctx, "default")

	userService := service.NewUserService(db.DB)
	jwtService := service.NewJWTService(cfg)
	inviteCodeService := service.NewInviteCodeService(db.DB)
	inviteCodeUsageService := service.NewInviteCodeUsageService(db.DB)
	authService := service.NewAuthService(db.DB, userService, jwtService, inviteCodeService)
	
	authHandler := handler.NewAuthHandler(cfg, db, authService, jwtService)
	taskHandler := handler.NewTaskHandler(taskQueue)
	adminUserHandler := handler.NewAdminUserHandler(userService)
	userProfileHandler := handler.NewUserProfileHandler(userService)
	inviteCodeHandler := handler.NewInviteCodeHandler(inviteCodeService, inviteCodeUsageService)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(gin.Recovery())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", func(c *gin.Context) {
			response.SuccessWithMessage(c, "pong", nil)
		})

		v1.POST("/tasks", middleware.AuthMiddleware(authService), taskHandler.CreateTask)
		v1.GET("/tasks/status", middleware.AuthMiddleware(authService), taskHandler.GetQueueStatus)
		
		// Authentication routes
		auth := v1.Group("/auth")
		{
			// OAuth routes
			auth.GET("/providers", authHandler.GetProviders)
			auth.GET("/telegram/widget", authHandler.GetTelegramWidget)
			auth.GET("/:provider", authHandler.Login)
			auth.GET("/:provider/callback", authHandler.Callback)
			
			// Local authentication routes
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.LoginLocal)
			auth.POST("/logout", middleware.AuthMiddleware(authService), authHandler.Logout)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/change-password", middleware.AuthMiddleware(authService), authHandler.ChangePassword)
			auth.GET("/profile", middleware.AuthMiddleware(authService), authHandler.GetProfile)
		}

		
		// Admin routes - require admin privileges
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(authService))
		admin.Use(middleware.RequireAdmin())
		{
			// Admin user management routes
			adminUsers := admin.Group("/users")
			{
				adminUsers.GET("", adminUserHandler.ListUsers)
				adminUsers.GET("/deleted", adminUserHandler.ListDeletedUsers)
				adminUsers.GET("/search", adminUserHandler.SearchUsers)
				adminUsers.GET("/stats", adminUserHandler.GetUserStats)
				adminUsers.GET("/provider", adminUserHandler.ListUsersByProvider)
				adminUsers.GET("/:id", adminUserHandler.GetUser)
				adminUsers.PUT("/:id", adminUserHandler.UpdateUser)
				adminUsers.PUT("/:id/role", adminUserHandler.UpdateUserRole)
				adminUsers.PUT("/:id/status", adminUserHandler.UpdateUserStatus)
				adminUsers.DELETE("/:id", adminUserHandler.SoftDeleteUser)
				adminUsers.POST("/:id/restore", adminUserHandler.RestoreUser)
				adminUsers.DELETE("/:id/hard-delete", adminUserHandler.HardDeleteUser)
				adminUsers.POST("/batch/delete", adminUserHandler.BatchDeleteUsers)
				adminUsers.POST("/batch/restore", adminUserHandler.BatchRestoreUsers)
			}

			// Admin invite code management routes
			adminInviteCodes := admin.Group("/invite-codes")
			{
				adminInviteCodes.GET("", inviteCodeHandler.ListAllInviteCodes)
				adminInviteCodes.GET("/stats", inviteCodeHandler.GetInviteCodeStats)
			}
		}

		// User routes - regular user access
		user := v1.Group("/user")
		user.Use(middleware.AuthMiddleware(authService))
		{
			// User profile management only
			user.GET("/profile", userProfileHandler.GetProfile)
			user.PUT("/profile", userProfileHandler.UpdateProfile)
			user.PUT("/password", userProfileHandler.ChangePassword)
		}

		// Invite code routes
		inviteCodes := v1.Group("/invite-codes")
		{
			// Public routes
			inviteCodes.GET("/validate/:code", inviteCodeHandler.ValidateInviteCode)
			
			// Authenticated routes
			inviteCodes.Use(middleware.AuthMiddleware(authService))
			inviteCodes.POST("", inviteCodeHandler.CreateInviteCode)
			inviteCodes.GET("/my", inviteCodeHandler.GetMyInviteCodes)
			inviteCodes.GET("/:id", inviteCodeHandler.GetInviteCode)
			inviteCodes.GET("/:id/usages", inviteCodeHandler.GetInviteCodeUsages)
			inviteCodes.PUT("/:id/status", inviteCodeHandler.UpdateInviteCodeStatus)
			inviteCodes.DELETE("/:id", inviteCodeHandler.DeleteInviteCode)
		}
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: r,
	}

	go func() {
		logger.Info("Server starting", logger.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", logger.Error2("error", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", logger.Error2("error", err))
	}

	logger.Info("Server exited")
}