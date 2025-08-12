package server

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"linke/internal/shared/cache"
	"linke/internal/shared/config"
	"linke/internal/shared/database"
	loggerPkg "linke/internal/shared/logger"
	"linke/internal/shared/middleware"
	"linke/internal/shared/queue"
	routerPkg "linke/internal/shared/router"
	"linke/internal/shared/versioning"

	// Handler imports
	authHandlers "linke/internal/domains/auth/adapters/handlers"
	authInterfaces "linke/internal/domains/auth/usecases/interfaces"
	couponHandlers "linke/internal/domains/coupon/adapters/handlers"
	invoiceHandlers "linke/internal/domains/invoice/adapters/handlers"
	paymentHandlers "linke/internal/domains/payment/adapters/handlers"
	referralHandlers "linke/internal/domains/referral/adapters/handlers"
	serverHandlers "linke/internal/domains/server/adapters/handlers"
	subscriptionHandlers "linke/internal/domains/subscription/adapters/handlers"
	ticketHandlers "linke/internal/domains/ticket/adapters/handlers"
	userHandlers "linke/internal/domains/user/adapters/handlers"
)

// HTTPServer HTTP 服务器结构
type HTTPServer struct {
	*gin.Engine
	logger loggerPkg.Logger
	config *config.Config
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

// NewAppHandler 创建应用处理器
func NewAppHandler(logger loggerPkg.Logger, db *database.Database) *AppHandler {
	return &AppHandler{
		logger:   logger,
		database: db,
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
	adminAuthHandler *authHandlers.AdminAuthHandler,
	userProfileHandler *userHandlers.UserProfileHandler,
	adminUserHandler *userHandlers.AdminUserHandler,
	userAccountBindingHandler *userHandlers.UserAccountBindingHandler,
	subscriptionPlanHandler *subscriptionHandlers.SubscriptionPlanHandler,
	subscriptionOrderHandler *subscriptionHandlers.SubscriptionOrderHandler,
	userSubscriptionHandler *subscriptionHandlers.UserSubscriptionHandler,
	quickPurchaseHandler *subscriptionHandlers.QuickPurchaseHandler,
	usageHandler *subscriptionHandlers.UsageHandler,
	usageAlertHandler *subscriptionHandlers.UsageAlertHandler,
	adminSubscriptionHandler *subscriptionHandlers.AdminSubscriptionHandler,
	invoiceHandler *invoiceHandlers.InvoiceHandler,
	adminInvoiceHandler *invoiceHandlers.AdminInvoiceHandler,
	adminTicketHandler *ticketHandlers.AdminTicketHandler,
	userTicketHandler *ticketHandlers.UserTicketHandler,
	paymentHandler *paymentHandlers.PaymentHandler,
	paymentMethodHandler *paymentHandlers.PaymentMethodHandler,
	serverHandler *serverHandlers.ServerAPIHandler,
	// Admin server handlers
	adminServerHandler *serverHandlers.AdminServerHandler,
	adminServerGroupHandler *serverHandlers.AdminServerGroupHandler,
	// Admin coupon and referral handlers
	adminCouponHandler *couponHandlers.AdminCouponHandler,
	userCouponHandler *couponHandlers.UserCouponHandler,
	adminReferralHandler *referralHandlers.AdminReferralHandler,
	// Cache monitoring handlers
	cacheMonitoringHandler *cache.CacheMonitoringHandler,
	// Versioning middleware
	versionMiddleware *versioning.VersionMiddleware,
	// Auth service for middleware
	authService authInterfaces.AuthService,
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

	// 添加 CORS 中间件
	router.Use(middleware.CORSFromConfig(cfg))

	// 设置路由
	routerPkg.SetupRoutes(
		router,
		cfg,
		logger,
		authHandler,
		adminAuthHandler,
		userProfileHandler,
		adminUserHandler,
		userAccountBindingHandler,
		subscriptionPlanHandler,
		subscriptionOrderHandler,
		userSubscriptionHandler,
		quickPurchaseHandler,
		usageHandler,
		usageAlertHandler,
		adminSubscriptionHandler,
		invoiceHandler,
		adminInvoiceHandler,
		adminTicketHandler,
		userTicketHandler,
		paymentHandler,
		paymentMethodHandler,
		serverHandler,
		adminServerHandler,
		adminServerGroupHandler,
		adminCouponHandler,
		userCouponHandler,
		adminReferralHandler,
		cacheMonitoringHandler,
		versionMiddleware,
		authService,
		appHandler.HealthCheck,
		taskHandler.CreateTask,
	)

	return &HTTPServer{
		Engine: router,
		logger: logger,
		config: cfg,
	}
}

// Start 启动 HTTP 服务器
func (s *HTTPServer) Start() error {
	addr := fmt.Sprintf(":%s", s.config.Server.Port)
	// 移动到debug级别，避免与bootstrap中的启动日志重复
	s.logger.Debug("HTTP server starting", zap.String("addr", addr))

	return http.ListenAndServe(addr, s.Engine)
}

// HealthCheck 健康检查
func (h *AppHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	health := h.database.HealthCheck(ctx)
	result := map[string]any{
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
		Type    string         `json:"type" binding:"required"`
		Payload map[string]any `json:"payload" binding:"required"`
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
		h.logger.Error("Failed to enqueue task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue task"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Task enqueued successfully",
		"task_id": task.ID,
	})
}
