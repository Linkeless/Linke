package router

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

	"linke/internal/shared/cache"
	"linke/internal/shared/config"
	loggerPkg "linke/internal/shared/logger"
	"linke/internal/shared/middleware"
	"linke/internal/shared/versioning"

	// Handler imports
	authHandlers "linke/internal/domains/auth/handlers"
	authInterfaces "linke/internal/domains/auth/usecases/interfaces"
	invoiceHandlers "linke/internal/domains/invoice/handlers"
	paymentHandlers "linke/internal/domains/payment/handlers"
	serverHandlers "linke/internal/domains/server/handlers"
	subscriptionHandlers "linke/internal/domains/subscription/handlers"
	userHandlers "linke/internal/domains/user/handlers"
)

// authServiceAdapter adapts the domain AuthService to middleware AuthService interface
type authServiceAdapter struct {
	authService authInterfaces.AuthService
}

func newAuthServiceAdapter(authService authInterfaces.AuthService) middleware.AuthService {
	return &authServiceAdapter{authService: authService}
}

func (a *authServiceAdapter) ValidateToken(token string) (interface{}, error) {
	user, err := a.authService.ValidateToken(token)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// SetupRoutes configures all HTTP routes for the application
func SetupRoutes(
	router *gin.Engine,
	cfg *config.Config,
	logger loggerPkg.Logger,
	// Business domain handlers
	authHandler *authHandlers.AuthHandler,
	userProfileHandler *userHandlers.UserProfileHandler,
	adminUserHandler *userHandlers.AdminUserHandler,
	subscriptionOrderHandler *subscriptionHandlers.SubscriptionOrderHandler,
	userSubscriptionHandler *subscriptionHandlers.UserSubscriptionHandler,
	quickPurchaseHandler *subscriptionHandlers.QuickPurchaseHandler,
	usageHandler *subscriptionHandlers.UsageHandler,
	usageAlertHandler *subscriptionHandlers.UsageAlertHandler,
	invoiceHandler *invoiceHandlers.InvoiceHandler,
	paymentHandler *paymentHandlers.PaymentHandler,
	paymentMethodHandler *paymentHandlers.PaymentMethodHandler,
	serverHandler *serverHandlers.ServerAPIHandler,
	// Cache monitoring handlers
	cacheMonitoringHandler *cache.CacheMonitoringHandler,
	// Versioning middleware
	versionMiddleware *versioning.VersionMiddleware,
	// Auth service for middleware
	authService authInterfaces.AuthService,
	// App and task handlers
	appHealthCheckFunc func(*gin.Context),
	taskCreateFunc func(*gin.Context),
) {
	logger.Info("Registering HTTP routes...")

	// Health check endpoint
	router.GET("/health", appHealthCheckFunc)
	logger.Info("Registered health check route", loggerPkg.String("route", "/health"))

	// API versioning routes - version info endpoints
	router.GET("/api/version", versionMiddleware.VersionInfo())
	router.GET("/api/health", versionMiddleware.HealthCheck())
	logger.Info("Registered versioning routes",
		loggerPkg.String("route1", "/api/version"),
		loggerPkg.String("route2", "/api/health"))

	// API route group - Apply versioning middleware
	api := router.Group("/api")
	api.Use(versionMiddleware.Middleware())

	// Version-specific route groups
	apiV1 := api.Group("/v1")
	// apiV2 := api.Group("/v2") // Reserved for future v2 endpoints
	logger.Info("Created API route groups",
		loggerPkg.String("base", "/api"),
		loggerPkg.String("v1", "/api/v1"))

	// Application routes
	appGroup := apiV1.Group("/app")
	{
		appGroup.GET("/system/health", appHealthCheckFunc)
	}

	// Task routes
	taskGroup := apiV1.Group("/tasks")
	{
		taskGroup.POST("", taskCreateFunc)
	}

	// Business domain routes registration

	// Auth routes (/api/v1/auth)
	authGroup := apiV1.Group("/auth")
	{
		// Local authentication
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.LoginLocal)
		authGroup.POST("/logout", authHandler.Logout)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.POST("/change-password", middleware.AuthMiddleware(newAuthServiceAdapter(authService)), authHandler.ChangePassword)

		// OAuth provider routes
		authGroup.GET("/providers", authHandler.GetProviders)

		// OAuth authentication start routes - using dynamic parameters to match handler expectations
		authGroup.GET("/:provider", authHandler.Login)

		// OAuth callback routes - using dynamic parameters to match handler expectations
		authGroup.GET("/:provider/callback", authHandler.Callback)

		// Generic callback route (compatibility)
		authGroup.GET("/callback", authHandler.Callback)

		// Additional OAuth endpoints
		authGroup.POST("/url", authHandler.GetAuthURL)
		authGroup.POST("/token", authHandler.ExchangeToken)
		authGroup.GET("/telegram/widget", authHandler.GetTelegramWidget)
	}

	// User routes (/api/v1/user) - requires authentication
	userGroup := apiV1.Group("/user")
	userGroup.Use(middleware.AuthMiddleware(newAuthServiceAdapter(authService)))
	{
		userGroup.GET("/profile", userProfileHandler.GetProfile)
		userGroup.PUT("/profile", userProfileHandler.UpdateProfile)
		// Password change functionality is handled by /api/v1/auth/change-password
	}

	// Admin route group (/api/v1/admin) - requires authentication and admin privileges
	adminGroup := apiV1.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware(newAuthServiceAdapter(authService)))
	adminGroup.Use(middleware.RequireAdmin())

	// Admin user routes (/api/v1/admin/users)
	adminUserGroup := adminGroup.Group("/users")
	{
		adminUserGroup.POST("", adminUserHandler.CreateUser)
		adminUserGroup.GET("", adminUserHandler.ListUsers)
		adminUserGroup.GET("/:id", adminUserHandler.GetUser)
		adminUserGroup.PUT("/:id", adminUserHandler.UpdateUser)
		adminUserGroup.PATCH("/:id", adminUserHandler.PatchUser)
		adminUserGroup.DELETE("/:id", adminUserHandler.SoftDeleteUser)
		adminUserGroup.POST("/:id/restore", adminUserHandler.RestoreUser)
		adminUserGroup.DELETE("/:id/hard-delete", adminUserHandler.HardDeleteUser)
		adminUserGroup.PUT("/:id/role", adminUserHandler.UpdateUserRole)
		adminUserGroup.PUT("/:id/status", adminUserHandler.UpdateUserStatus)
		adminUserGroup.POST("/:id/reset-password", adminUserHandler.ResetUserPassword)
		adminUserGroup.GET("/statistics", adminUserHandler.GetUserStats)
		adminUserGroup.GET("/provider", adminUserHandler.ListUsersByProvider)
		adminUserGroup.GET("/deleted", adminUserHandler.ListDeletedUsers)
		adminUserGroup.GET("/search", adminUserHandler.SearchUsers)
		adminUserGroup.POST("/bulk/delete", adminUserHandler.BatchDeleteUsers)
		adminUserGroup.POST("/bulk/restore", adminUserHandler.BatchRestoreUsers)
	}

	// Admin cache routes (/api/v1/admin/cache) - using RegisterRoutes method
	logger.Info("Registering cache monitoring routes", loggerPkg.String("prefix", "/api/v1/admin"))
	if cacheMonitoringHandler == nil {
		logger.Error("Cache monitoring handler is nil - routes will not be registered")
	} else {
		cacheMonitoringHandler.RegisterRoutes(adminGroup)
		logger.Info("Successfully registered cache monitoring routes")
	}

	// Admin payment routes (/api/v1/admin/payment)
	adminPaymentGroup := adminGroup.Group("/payment")
	{
		// Payment configuration management routes
		adminPaymentGroup.POST("/configs", paymentHandler.CreatePaymentConfig)
		adminPaymentGroup.GET("/configs", paymentHandler.GetPaymentConfigs)
		adminPaymentGroup.PUT("/configs/:id", paymentHandler.UpdatePaymentConfig)
		adminPaymentGroup.DELETE("/configs/:id", paymentHandler.DeletePaymentConfig)

		// Payment retry management routes
		adminPaymentGroup.GET("/retries", paymentHandler.GetPaymentRetries)
		adminPaymentGroup.GET("/retries/:id", paymentHandler.GetPaymentRetry)
		adminPaymentGroup.POST("/retries/:id/cancel", paymentHandler.CancelPaymentRetry)
		adminPaymentGroup.POST("/retries/:id/reset", paymentHandler.ResetPaymentRetry)
		adminPaymentGroup.POST("/retries/bulk/cancel", paymentHandler.BulkCancelPaymentRetries)
		adminPaymentGroup.POST("/retries/bulk/reset", paymentHandler.BulkResetPaymentRetries)
		adminPaymentGroup.GET("/retries/statistics", paymentHandler.GetRetryStatistics)
		adminPaymentGroup.GET("/retries/health", paymentHandler.GetRetryHealthMetrics)
	}
	logger.Info("Registered admin payment routes", loggerPkg.String("prefix", "/api/v1/admin/payment"))

	// Admin subscription routes (/api/v1/admin/subscriptions)
	// TODO: Add subscription pause/resume routes when methods are implemented
	// adminSubscriptionGroup := adminGroup.Group("/subscriptions")
	// {
	//     adminSubscriptionGroup.POST("/:id/pause", userSubscriptionHandler.PauseSubscription)
	//     adminSubscriptionGroup.POST("/:id/resume", userSubscriptionHandler.ResumeSubscription)
	// }
	logger.Info("Admin subscription routes will be added when pause/resume methods are implemented")

	// Subscription routes (/api/v1/subscription) - most require authentication
	subscriptionGroup := apiV1.Group("/subscription")
	{
		// Quick purchase - requires authentication
		subscriptionGroup.POST("/quick-purchase", middleware.AuthMiddleware(newAuthServiceAdapter(authService)), quickPurchaseHandler.QuickPurchase)

		// Order related - requires authentication
		subscriptionGroup.POST("/orders", middleware.AuthMiddleware(newAuthServiceAdapter(authService)), subscriptionOrderHandler.CreateSubscriptionOrder)
		subscriptionGroup.GET("/orders/my", middleware.AuthMiddleware(newAuthServiceAdapter(authService)), subscriptionOrderHandler.GetMySubscriptionOrders)
		subscriptionGroup.GET("/orders/:id", middleware.AuthMiddleware(newAuthServiceAdapter(authService)), subscriptionOrderHandler.GetSubscriptionOrder)
	}

	// User subscription management routes (/api/v1/subscriptions) - using RegisterRoutes method
	logger.Info("Registering user subscription routes")
	if userSubscriptionHandler == nil {
		logger.Error("User subscription handler is nil - routes will not be registered")
	} else {
		userSubscriptionHandler.RegisterRoutes(apiV1)
		logger.Info("Successfully registered user subscription routes")
	}

	// Usage tracking routes (/api/v1/usage) - using RegisterRoutes method
	logger.Info("Registering usage tracking routes")
	if usageHandler == nil {
		logger.Error("Usage handler is nil - routes will not be registered")
	} else {
		usageHandler.RegisterRoutes(apiV1)
		logger.Info("Successfully registered usage tracking routes")
	}

	// Usage alert routes (/api/v1/usage-alerts) - using RegisterRoutes method
	logger.Info("Registering usage alert routes")
	if usageAlertHandler == nil {
		logger.Error("Usage alert handler is nil - routes will not be registered")
	} else {
		usageAlertHandler.RegisterRoutes(apiV1)
		logger.Info("Successfully registered usage alert routes")
	}

	// Invoice routes (/api/v1/invoice) - using RegisterRoutes method
	logger.Info("Registering invoice routes")
	if invoiceHandler == nil {
		logger.Error("Invoice handler is nil - routes will not be registered")
	} else {
		invoiceHandler.RegisterRoutes(apiV1)
		logger.Info("Successfully registered invoice routes")
	}

	// Payment routes (/api/v1/payment)
	paymentGroup := apiV1.Group("/payment")
	{
		paymentGroup.POST("/orders", middleware.AuthMiddleware(newAuthServiceAdapter(authService)), paymentHandler.CreatePaymentOrder)
		paymentGroup.GET("/orders/:payment_no", middleware.AuthMiddleware(newAuthServiceAdapter(authService)), paymentHandler.GetPaymentOrder)
		paymentGroup.GET("/orders/my", middleware.AuthMiddleware(newAuthServiceAdapter(authService)), paymentHandler.GetMyPaymentOrders)
		// Public endpoint - no auth needed
		paymentGroup.GET("/methods", paymentHandler.GetAvailablePaymentMethods)
		paymentGroup.GET("/configs", paymentHandler.GetActivePaymentConfigs)
		// Webhook endpoint - no auth needed (uses signature validation)
		paymentGroup.POST("/notify/:gateway", paymentHandler.PaymentNotify)
	}

	// Payment methods management routes (/api/v1/payment-methods) - requires authentication
	paymentMethodsGroup := apiV1.Group("/payment-methods")
	paymentMethodsGroup.Use(middleware.AuthMiddleware(newAuthServiceAdapter(authService)))
	{
		paymentMethodsGroup.POST("", paymentMethodHandler.CreatePaymentMethod)
		paymentMethodsGroup.GET("/:id", paymentMethodHandler.GetPaymentMethod)
		paymentMethodsGroup.GET("", paymentMethodHandler.ListPaymentMethods)
		paymentMethodsGroup.PUT("/:id", paymentMethodHandler.UpdatePaymentMethod)
		paymentMethodsGroup.DELETE("/:id", paymentMethodHandler.DeletePaymentMethod)
		paymentMethodsGroup.POST("/:id/validate", paymentMethodHandler.ValidatePaymentMethod)
		paymentMethodsGroup.GET("/default", paymentMethodHandler.GetDefaultPaymentMethod)
		paymentMethodsGroup.PUT("/:id/default", paymentMethodHandler.SetDefaultPaymentMethod)
		paymentMethodsGroup.GET("/:id/statistics", paymentMethodHandler.GetPaymentMethodUsageStats)
	}

	// Server routes (/api/v1/server)
	serverGroup := apiV1.Group("/server")
	{
		serverGroup.GET("/health", serverHandler.Health)
		serverGroup.GET("/uni-proxy/config", serverHandler.UniProxyConfig)
		serverGroup.GET("/uni-proxy/users", serverHandler.UniProxyUsers)
		serverGroup.POST("/uni-proxy/push", serverHandler.UniProxyPush)
	}

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	logger.Info("Registered Swagger documentation route", loggerPkg.String("route", "/swagger/*any"))

	// Log all registered routes for debugging
	routes := router.Routes()
	logger.Info("HTTP route registration completed successfully",
		loggerPkg.Int("total_routes", len(routes)))

	// Log important routes for verification
	importantRoutes := []string{
		"/api/v1/admin/cache/metrics",
		"/api/v1/admin/cache/statistics",
		"/api/v1/admin/payment/configs",
		"/api/v1/admin/payment/retries",
		"/api/v1/payment-methods",
		"/api/v1/payment-methods/*/validate",
		"/api/v1/payment/orders",
		"/api/v1/subscriptions/my",
		"/api/v1/subscription/orders",
		"/api/v1/usage/current/*",
		"/api/v1/usage-alerts",
		"/api/v1/auth/providers",
		"/api/v1/invoice/*/pdf",
	}

	for _, route := range routes {
		for _, important := range importantRoutes {
			if route.Path == important ||
				(important[len(important)-1] == '*' &&
					len(route.Path) >= len(important)-1 &&
					route.Path[:len(important)-1] == important[:len(important)-1]) {
				logger.Info("Verified important route",
					loggerPkg.String("method", route.Method),
					loggerPkg.String("path", route.Path))
				break
			}
		}
	}
}
