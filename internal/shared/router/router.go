package router

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	"linke/internal/shared/cache"
	"linke/internal/shared/config"
	loggerPkg "linke/internal/shared/logger"
	"linke/internal/shared/middleware"
	"linke/internal/shared/versioning"

	// Handler imports
	authHandlers "linke/internal/domains/auth/handlers"
	authInterfaces "linke/internal/domains/auth/usecases/interfaces"
	couponHandlers "linke/internal/domains/coupon/handlers"
	invoiceHandlers "linke/internal/domains/invoice/handlers"
	paymentHandlers "linke/internal/domains/payment/handlers"
	referralHandlers "linke/internal/domains/referral/handlers"
	serverHandlers "linke/internal/domains/server/handlers"
	subscriptionHandlers "linke/internal/domains/subscription/handlers"
	ticketHandlers "linke/internal/domains/ticket/handlers"
	userHandlers "linke/internal/domains/user/handlers"
)

// authServiceAdapter adapts the domain AuthService to middleware AuthService interface
type authServiceAdapter struct {
	authService authInterfaces.AuthService
}

func newAuthServiceAdapter(authService authInterfaces.AuthService) middleware.AuthService {
	return &authServiceAdapter{authService: authService}
}

func (a *authServiceAdapter) ValidateToken(token string) (any, error) {
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
	adminAuthHandler *authHandlers.AdminAuthHandler,
	userProfileHandler *userHandlers.UserProfileHandler,
	adminUserHandler *userHandlers.AdminUserHandler,
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
	// App and task handlers
	appHealthCheckFunc func(*gin.Context),
	taskCreateFunc func(*gin.Context),
) {
	logger.Debug("Registering HTTP routes...")

	// Health check endpoint
	router.GET("/health", appHealthCheckFunc)
	logger.Debug("Registered health check route", zap.String("route", "/health"))

	// API versioning routes - version info endpoints
	router.GET("/api/version", versionMiddleware.VersionInfo())
	router.GET("/api/health", versionMiddleware.HealthCheck())
	logger.Debug("Registered versioning routes",
		zap.String("route1", "/api/version"),
		zap.String("route2", "/api/health"))

	// API route group - Apply versioning middleware
	api := router.Group("/api")
	api.Use(versionMiddleware.Middleware())

	// Version-specific route groups
	apiV1 := api.Group("/v1")
	// apiV2 := api.Group("/v2") // Reserved for future v2 endpoints
	logger.Info("Created API route groups",
		zap.String("base", "/api"),
		zap.String("v1", "/api/v1"))

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

	// Admin auth routes (/api/v1/admin/auth) - Authentication security management
	adminAuthGroup := adminGroup.Group("/auth")
	{
		// JWT Token Management
		jwtGroup := adminAuthGroup.Group("/jwt")
		{
			jwtGroup.GET("/tokens", adminAuthHandler.ListActiveTokens)
			jwtGroup.GET("/blacklist", adminAuthHandler.ListJWTBlacklist)
			jwtGroup.POST("/revoke", adminAuthHandler.RevokeJWTToken)
			jwtGroup.GET("/analytics", adminAuthHandler.GetJWTAnalytics)
		}

		// Login Security Monitoring
		securityGroup := adminAuthGroup.Group("/security")
		{
			securityGroup.GET("/login-attempts", adminAuthHandler.ListLoginAttempts)
			securityGroup.GET("/failed-logins", adminAuthHandler.GetFailedLoginAnalysis)
		}

		// Account Security Management
		accountsGroup := adminAuthGroup.Group("/accounts")
		{
			accountsGroup.POST("/unlock", adminAuthHandler.UnlockAccount)
			accountsGroup.POST("/reset-password", adminAuthHandler.ForcePasswordReset)
			accountsGroup.GET("/:user_id/security-status", adminAuthHandler.GetAccountSecurityStatus)
		}

		// OAuth Provider Management
		oauthGroup := adminAuthGroup.Group("/oauth")
		{
			oauthGroup.GET("/providers", adminAuthHandler.GetOAuthProviderStats)
			oauthGroup.GET("/incidents", adminAuthHandler.ListOAuthSecurityEvents)
		}

		// Security Analytics and Reporting
		analyticsGroup := adminAuthGroup.Group("/analytics")
		{
			analyticsGroup.GET("/statistics", adminAuthHandler.GetSecurityStatistics)
			analyticsGroup.GET("/patterns", adminAuthHandler.GetSecurityPatterns)
			analyticsGroup.GET("/security-score", adminAuthHandler.GetSecurityScore)
		}

		// Bulk Security Operations
		bulkGroup := adminAuthGroup.Group("/bulk")
		{
			bulkGroup.POST("/revoke-tokens", adminAuthHandler.BulkRevokeTokens)
			bulkGroup.POST("/unlock-accounts", adminAuthHandler.BulkUnlockAccounts)
			bulkGroup.POST("/reset-passwords", adminAuthHandler.BulkResetPasswords)
			bulkGroup.POST("/notifications", adminAuthHandler.SendSecurityNotifications)
		}
	}

	// Admin cache routes (/api/v1/admin/cache) - using RegisterRoutes method
	logger.Debug("Registering cache monitoring routes", zap.String("prefix", "/api/v1/admin"))
	if cacheMonitoringHandler == nil {
		logger.Error("Cache monitoring handler is nil - routes will not be registered")
	} else {
		cacheMonitoringHandler.RegisterRoutes(adminGroup)
		logger.Debug("Successfully registered cache monitoring routes")
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
	logger.Debug("Registered admin payment routes", zap.String("prefix", "/api/v1/admin/payment"))

	// Admin server routes (/api/v1/admin/servers)
	adminServerGroup := adminGroup.Group("/servers")
	{
		adminServerGroup.POST("", adminServerHandler.CreateServer)
		adminServerGroup.GET("", adminServerHandler.ListServers)
		adminServerGroup.GET("/:id", adminServerHandler.GetServer)
		adminServerGroup.PUT("/:id", adminServerHandler.UpdateServer)
		adminServerGroup.PATCH("/:id", adminServerHandler.PatchServer)
		adminServerGroup.DELETE("/:id", adminServerHandler.DeleteServer)
		adminServerGroup.PUT("/:id/status", adminServerHandler.UpdateServerStatus)
		adminServerGroup.GET("/:id/statistics", adminServerHandler.GetServerStatistics)
		adminServerGroup.GET("/:id/health", adminServerHandler.CheckServerHealth)
		adminServerGroup.GET("/group/:group_id", adminServerHandler.GetServersByGroup)
		adminServerGroup.POST("/bulk/update", adminServerHandler.BulkUpdateServers)
	}
	logger.Debug("Registered admin server routes", zap.String("prefix", "/api/v1/admin/servers"))

	// Admin server group routes (/api/v1/admin/server-groups)
	adminServerGroupGroup := adminGroup.Group("/server-groups")
	{
		adminServerGroupGroup.POST("", adminServerGroupHandler.CreateGroup)
		adminServerGroupGroup.GET("", adminServerGroupHandler.ListGroups)
		adminServerGroupGroup.GET("/:id", adminServerGroupHandler.GetGroup)
		adminServerGroupGroup.PUT("/:id", adminServerGroupHandler.UpdateGroup)
		adminServerGroupGroup.PATCH("/:id", adminServerGroupHandler.PatchGroup)
		adminServerGroupGroup.DELETE("/:id", adminServerGroupHandler.DeleteGroup)
		adminServerGroupGroup.GET("/:id/servers", adminServerGroupHandler.GetGroupServers)
		adminServerGroupGroup.GET("/:id/statistics", adminServerGroupHandler.GetGroupStatistics)
		adminServerGroupGroup.GET("/statistics", adminServerGroupHandler.GetAllGroupStatistics)
	}
	logger.Debug("Registered admin server group routes", zap.String("prefix", "/api/v1/admin/server-groups"))

	// Admin subscription routes (/api/v1/admin/subscriptions)
	adminSubscriptionGroup := adminGroup.Group("/subscriptions")
	{
		// Subscription Plans Management
		plansGroup := adminSubscriptionGroup.Group("/plans")
		{
			plansGroup.POST("", adminSubscriptionHandler.CreateSubscriptionPlan)
			plansGroup.GET("", adminSubscriptionHandler.ListSubscriptionPlans)
			plansGroup.GET("/:id", adminSubscriptionHandler.GetSubscriptionPlan)
			plansGroup.PUT("/:id", adminSubscriptionHandler.UpdateSubscriptionPlan)
			plansGroup.DELETE("/:id", adminSubscriptionHandler.DeleteSubscriptionPlan)
			plansGroup.PUT("/:id/toggle-status", adminSubscriptionHandler.ToggleSubscriptionPlanStatus)
		}

		// User Subscriptions Management
		usersGroup := adminSubscriptionGroup.Group("/users")
		{
			usersGroup.POST("", adminSubscriptionHandler.CreateUserSubscription)
			usersGroup.GET("", adminSubscriptionHandler.ListUserSubscriptions)
			usersGroup.GET("/:id", adminSubscriptionHandler.GetUserSubscription)
			usersGroup.PUT("/:id", adminSubscriptionHandler.UpdateUserSubscription)
			usersGroup.POST("/:id/pause", adminSubscriptionHandler.PauseUserSubscription)
			usersGroup.POST("/:id/resume", adminSubscriptionHandler.ResumeUserSubscription)
			usersGroup.POST("/:id/upgrade", adminSubscriptionHandler.UpgradeSubscription)
			usersGroup.POST("/:id/downgrade", adminSubscriptionHandler.DowngradeSubscription)
			usersGroup.POST("/:id/extend", adminSubscriptionHandler.ExtendUserSubscription)
			usersGroup.POST("/:id/cancel", adminSubscriptionHandler.CancelUserSubscription)
			usersGroup.POST("/:id/reset-traffic", adminSubscriptionHandler.ResetTrafficUsage)
		}

		// Subscription Orders Management
		ordersGroup := adminSubscriptionGroup.Group("/orders")
		{
			ordersGroup.GET("", adminSubscriptionHandler.ListSubscriptionOrders)
			ordersGroup.GET("/:id", adminSubscriptionHandler.GetSubscriptionOrder)
			ordersGroup.POST("/:id/cancel", adminSubscriptionHandler.CancelSubscriptionOrder)
		}

		// Usage Management
		usageGroup := adminSubscriptionGroup.Group("/usage")
		{
			usageGroup.GET("/:id/statistics", adminSubscriptionHandler.GetUsageStatistics)
			usageGroup.GET("/:id/current", adminSubscriptionHandler.GetCurrentUsage)
		}

		// Analytics
		analyticsGroup := adminSubscriptionGroup.Group("/analytics")
		{
			analyticsGroup.GET("/statistics", adminSubscriptionHandler.GetSubscriptionStatistics)
			analyticsGroup.GET("/orders", adminSubscriptionHandler.GetOrderStatistics)
		}

		// Usage Alerts Management
		alertsGroup := adminSubscriptionGroup.Group("/alerts")
		{
			alertsGroup.GET("", adminSubscriptionHandler.GetUsageAlerts)
			alertsGroup.GET("/statistics", adminSubscriptionHandler.GetAlertStatistics)
			alertsGroup.POST("/bulk/resolve", adminSubscriptionHandler.BulkResolveAlerts)
		}

		// Bulk Operations
		bulkGroup := adminSubscriptionGroup.Group("/bulk")
		{
			bulkGroup.POST("/action", adminSubscriptionHandler.BulkSubscriptionAction)
		}
	}
	logger.Debug("Registered admin subscription routes", zap.String("prefix", "/api/v1/admin/subscriptions"))

	// Admin invoice routes (/api/v1/admin/invoices)
	adminInvoiceGroup := adminGroup.Group("/invoices")
	{
		// Invoice Management
		adminInvoiceGroup.POST("", adminInvoiceHandler.CreateInvoice)
		adminInvoiceGroup.GET("", adminInvoiceHandler.ListInvoices)
		adminInvoiceGroup.GET("/:id", adminInvoiceHandler.GetInvoice)
		adminInvoiceGroup.PUT("/:id", adminInvoiceHandler.UpdateInvoice)
		adminInvoiceGroup.DELETE("/:id", adminInvoiceHandler.DeleteInvoice)

		// Invoice Status Operations
		adminInvoiceGroup.POST("/:id/void", adminInvoiceHandler.VoidInvoice)
		adminInvoiceGroup.POST("/:id/mark-paid", adminInvoiceHandler.MarkInvoiceAsPaid)

		// PDF and Email Operations
		adminInvoiceGroup.POST("/:id/regenerate-pdf", adminInvoiceHandler.RegenerateInvoicePDF)
		adminInvoiceGroup.POST("/:id/resend-email", adminInvoiceHandler.ResendInvoice)

		// Search and Analytics
		adminInvoiceGroup.GET("/search", adminInvoiceHandler.SearchInvoices)
		adminInvoiceGroup.GET("/statistics", adminInvoiceHandler.GetInvoiceStatistics)
		adminInvoiceGroup.GET("/analytics", adminInvoiceHandler.GetInvoiceAnalytics)
		adminInvoiceGroup.GET("/overdue", adminInvoiceHandler.GetOverdueInvoices)

		// Bulk Operations
		bulkGroup := adminInvoiceGroup.Group("/bulk")
		{
			bulkGroup.POST("/void", adminInvoiceHandler.BulkVoidInvoices)
			bulkGroup.POST("/mark-paid", adminInvoiceHandler.BulkMarkPaid)
			bulkGroup.POST("/resend", adminInvoiceHandler.BulkResendInvoices)
			bulkGroup.POST("/regenerate-pdf", adminInvoiceHandler.BulkRegeneratePDF)
		}

		// Templates and Configuration
		adminInvoiceGroup.GET("/templates", adminInvoiceHandler.GetAvailableTemplates)
		adminInvoiceGroup.GET("/languages", adminInvoiceHandler.GetAvailableLanguages)
	}
	logger.Debug("Registered admin invoice routes", zap.String("prefix", "/api/v1/admin/invoices"))

	// Admin ticket routes (/api/v1/admin/tickets)
	adminTicketGroup := adminGroup.Group("/tickets")
	{
		// Ticket Management
		adminTicketGroup.POST("", adminTicketHandler.CreateTicket)
		adminTicketGroup.GET("", adminTicketHandler.ListTickets)
		adminTicketGroup.GET("/:id", adminTicketHandler.GetTicket)
		adminTicketGroup.PUT("/:id", adminTicketHandler.UpdateTicket)
		adminTicketGroup.DELETE("/:id", adminTicketHandler.DeleteTicket)

		// Ticket Assignment and Status Operations
		adminTicketGroup.POST("/:id/assign", adminTicketHandler.AssignTicket)
		adminTicketGroup.POST("/:id/escalate", adminTicketHandler.EscalateTicket)
		adminTicketGroup.POST("/:id/close", adminTicketHandler.CloseTicket)
		adminTicketGroup.POST("/:id/reopen", adminTicketHandler.ReopenTicket)

		// Message Management
		adminTicketGroup.GET("/:id/messages", adminTicketHandler.GetTicketMessages)
		adminTicketGroup.POST("/:id/messages", adminTicketHandler.AddMessage)
		adminTicketGroup.GET("/:id/messages/:msg_id", adminTicketHandler.GetMessage)
		adminTicketGroup.PUT("/:id/messages/:msg_id", adminTicketHandler.UpdateMessage)
		adminTicketGroup.DELETE("/:id/messages/:msg_id", adminTicketHandler.DeleteMessage)

		// Internal Notes
		adminTicketGroup.POST("/:id/notes", adminTicketHandler.AddInternalNote)

		// Search and Analytics
		adminTicketGroup.GET("/search", adminTicketHandler.SearchTickets)
		adminTicketGroup.GET("/statistics", adminTicketHandler.GetStatistics)
		adminTicketGroup.GET("/analytics", adminTicketHandler.GetAnalytics)

		// Agent Management
		adminTicketGroup.GET("/agents", adminTicketHandler.GetAgents)

		// Bulk Operations
		bulkGroup := adminTicketGroup.Group("/bulk")
		{
			bulkGroup.POST("/assign", adminTicketHandler.BulkAssignTickets)
			bulkGroup.POST("/status", adminTicketHandler.BulkUpdateStatus)
			bulkGroup.POST("/close", adminTicketHandler.BulkCloseTickets)
		}
	}
	logger.Debug("Registered admin ticket routes", zap.String("prefix", "/api/v1/admin/tickets"))

	// User ticket routes (/api/v1/tickets) - requires authentication
	userTicketGroup := apiV1.Group("/tickets")
	userTicketGroup.Use(middleware.AuthMiddleware(newAuthServiceAdapter(authService)))
	{
		// Ticket Management
		userTicketGroup.POST("", userTicketHandler.CreateTicket)
		userTicketGroup.GET("/my", userTicketHandler.GetMyTickets)
		userTicketGroup.GET("/:id", userTicketHandler.GetTicket)
		userTicketGroup.PUT("/:id/close", userTicketHandler.CloseTicket)

		// Message Management
		userTicketGroup.GET("/:id/messages", userTicketHandler.GetTicketMessages)
		userTicketGroup.POST("/:id/messages", userTicketHandler.AddMessage)
	}
	logger.Debug("Registered user ticket routes", zap.String("prefix", "/api/v1/tickets"))

	// Admin coupon routes (/api/v1/admin/coupons)
	adminCouponGroup := adminGroup.Group("/coupons")
	{
		// Coupon CRUD
		adminCouponGroup.POST("", adminCouponHandler.CreateCoupon)
		adminCouponGroup.GET("", adminCouponHandler.ListCoupons)
		adminCouponGroup.GET("/:id", adminCouponHandler.GetCoupon)
		adminCouponGroup.PUT("/:id", adminCouponHandler.UpdateCoupon)
		adminCouponGroup.DELETE("/:id", adminCouponHandler.DeleteCoupon)

		// Coupon Management
		adminCouponGroup.PUT("/:id/toggle-status", adminCouponHandler.ToggleCouponStatus)
		adminCouponGroup.PUT("/:id/extend", adminCouponHandler.ExtendCouponExpiry)
		adminCouponGroup.GET("/:id/usage", adminCouponHandler.GetCouponUsage)

		// Search and Analytics
		adminCouponGroup.GET("/search", adminCouponHandler.SearchCoupons)
		adminCouponGroup.GET("/statistics", adminCouponHandler.GetCouponStatistics)
		adminCouponGroup.GET("/analytics", adminCouponHandler.GetCouponAnalytics)

		// Bulk Operations
		bulkCouponGroup := adminCouponGroup.Group("/bulk")
		{
			bulkCouponGroup.POST("/create", adminCouponHandler.BulkCreateCoupons)
			bulkCouponGroup.POST("/update", adminCouponHandler.BulkUpdateCoupons)
			bulkCouponGroup.POST("/deactivate", adminCouponHandler.BulkDeactivateCoupons)
		}
	}
	logger.Debug("Registered admin coupon routes", zap.String("prefix", "/api/v1/admin/coupons"))

	// User coupon routes (/api/v1/coupons) - using RegisterRoutes method
	logger.Debug("Registering user coupon routes")
	if userCouponHandler == nil {
		logger.Error("User coupon handler is nil - routes will not be registered")
	} else {
		userCouponHandler.RegisterRoutes(apiV1)
		logger.Debug("Successfully registered user coupon routes")
	}

	// Admin referral routes (/api/v1/admin/referrals)
	adminReferralGroup := adminGroup.Group("/referrals")
	{
		// Referral CRUD
		adminReferralGroup.POST("", adminReferralHandler.CreateReferral)
		adminReferralGroup.GET("", adminReferralHandler.ListReferrals)
		adminReferralGroup.GET("/:id", adminReferralHandler.GetReferral)
		adminReferralGroup.PUT("/:id", adminReferralHandler.UpdateReferral)

		// Referral Management
		adminReferralGroup.POST("/:id/approve", adminReferralHandler.ApproveReferral)
		adminReferralGroup.POST("/:id/reject", adminReferralHandler.RejectReferral)
		adminReferralGroup.POST("/:id/payout", adminReferralHandler.ProcessReferralPayout)

		// Search and Analytics
		adminReferralGroup.GET("/search", adminReferralHandler.SearchReferrals)
		adminReferralGroup.GET("/statistics", adminReferralHandler.GetReferralStatistics)
		adminReferralGroup.GET("/analytics", adminReferralHandler.GetReferralAnalytics)

		// Campaign Management
		campaignGroup := adminReferralGroup.Group("/campaigns")
		{
			campaignGroup.GET("", adminReferralHandler.ListCampaigns)
			campaignGroup.POST("", adminReferralHandler.CreateCampaign)
		}

		// Invite Code Management
		inviteCodeGroup := adminReferralGroup.Group("/invite-codes")
		{
			inviteCodeGroup.GET("", adminReferralHandler.ListInviteCodes)
		}

		// Bulk Operations
		bulkReferralGroup := adminReferralGroup.Group("/bulk")
		{
			bulkReferralGroup.POST("/approve", adminReferralHandler.BulkApproveReferrals)
			bulkReferralGroup.POST("/payout", adminReferralHandler.BulkProcessPayouts)
		}
	}
	logger.Debug("Registered admin referral routes", zap.String("prefix", "/api/v1/admin/referrals"))

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

	// Subscription plan routes (/api/v1/subscription/plans) - using RegisterRoutes method
	logger.Debug("Registering subscription plan routes")
	if subscriptionPlanHandler == nil {
		logger.Error("Subscription plan handler is nil - routes will not be registered")
	} else {
		subscriptionPlanHandler.RegisterRoutes(apiV1)
		logger.Debug("Successfully registered subscription plan routes")
	}

	// User subscription management routes (/api/v1/subscriptions) - using RegisterRoutes method
	logger.Debug("Registering user subscription routes")
	if userSubscriptionHandler == nil {
		logger.Error("User subscription handler is nil - routes will not be registered")
	} else {
		userSubscriptionHandler.RegisterRoutes(apiV1)
		logger.Debug("Successfully registered user subscription routes")
	}

	// Usage tracking routes (/api/v1/usage) - using RegisterRoutes method
	logger.Debug("Registering usage tracking routes")
	if usageHandler == nil {
		logger.Error("Usage handler is nil - routes will not be registered")
	} else {
		usageHandler.RegisterRoutes(apiV1)
		logger.Debug("Successfully registered usage tracking routes")
	}

	// Usage alert routes (/api/v1/usage-alerts) - using RegisterRoutes method
	logger.Debug("Registering usage alert routes")
	if usageAlertHandler == nil {
		logger.Error("Usage alert handler is nil - routes will not be registered")
	} else {
		usageAlertHandler.RegisterRoutes(apiV1)
		logger.Debug("Successfully registered usage alert routes")
	}

	// Invoice routes (/api/v1/invoice) - using RegisterRoutes method
	logger.Debug("Registering invoice routes")
	if invoiceHandler == nil {
		logger.Error("Invoice handler is nil - routes will not be registered")
	} else {
		invoiceHandler.RegisterRoutes(apiV1)
		logger.Debug("Successfully registered invoice routes")
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

	// Server routes (/api/v1/server) - UniProxy endpoints use token parameter authentication, not Bearer
	serverGroup := apiV1.Group("/server")
	{
		// UniProxy endpoints - use token parameter authentication
		serverGroup.GET("/UniProxy/health", serverHandler.Health)
		serverGroup.GET("/UniProxy/config", serverHandler.UniProxyConfig)
		serverGroup.GET("/UniProxy/user", serverHandler.UniProxyUsers)
		serverGroup.POST("/UniProxy/push", serverHandler.UniProxyPush)
	}

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	logger.Debug("Registered Swagger documentation route", zap.String("route", "/swagger/*any"))

	// Log all registered routes for debugging
	routes := router.Routes()
	logger.Info("HTTP route registration completed successfully",
		zap.Int("total_routes", len(routes)))

	// Log important routes for verification
	importantRoutes := []string{
		"/api/v1/admin/auth/jwt/tokens",
		"/api/v1/admin/auth/jwt/blacklist",
		"/api/v1/admin/auth/jwt/analytics",
		"/api/v1/admin/auth/security/login-attempts",
		"/api/v1/admin/auth/security/failed-logins",
		"/api/v1/admin/auth/accounts/unlock",
		"/api/v1/admin/auth/accounts/reset-password",
		"/api/v1/admin/auth/oauth/providers",
		"/api/v1/admin/auth/oauth/incidents",
		"/api/v1/admin/auth/analytics/statistics",
		"/api/v1/admin/auth/analytics/patterns",
		"/api/v1/admin/auth/analytics/security-score",
		"/api/v1/admin/auth/bulk/*",
		"/api/v1/admin/cache/metrics",
		"/api/v1/admin/cache/statistics",
		"/api/v1/admin/payment/configs",
		"/api/v1/admin/payment/retries",
		"/api/v1/admin/servers",
		"/api/v1/admin/servers/*/statistics",
		"/api/v1/admin/servers/*/health",
		"/api/v1/admin/server-groups",
		"/api/v1/admin/server-groups/*/servers",
		"/api/v1/admin/server-groups/statistics",
		"/api/v1/admin/subscriptions/plans",
		"/api/v1/admin/subscriptions/users",
		"/api/v1/admin/subscriptions/orders",
		"/api/v1/admin/subscriptions/analytics/*",
		"/api/v1/admin/subscriptions/alerts",
		"/api/v1/admin/subscriptions/bulk/*",
		"/api/v1/admin/invoices",
		"/api/v1/admin/invoices/*/void",
		"/api/v1/admin/invoices/*/mark-paid",
		"/api/v1/admin/invoices/*/regenerate-pdf",
		"/api/v1/admin/invoices/search",
		"/api/v1/admin/invoices/statistics",
		"/api/v1/admin/invoices/analytics",
		"/api/v1/admin/invoices/overdue",
		"/api/v1/admin/invoices/bulk/*",
		"/api/v1/admin/invoices/templates",
		"/api/v1/admin/tickets",
		"/api/v1/admin/tickets/*/assign",
		"/api/v1/admin/tickets/*/escalate",
		"/api/v1/admin/tickets/*/close",
		"/api/v1/admin/tickets/*/reopen",
		"/api/v1/admin/tickets/*/messages",
		"/api/v1/admin/tickets/*/notes",
		"/api/v1/admin/tickets/search",
		"/api/v1/admin/tickets/statistics",
		"/api/v1/admin/tickets/analytics",
		"/api/v1/admin/tickets/agents",
		"/api/v1/admin/tickets/bulk/*",
		"/api/v1/tickets",
		"/api/v1/tickets/my",
		"/api/v1/tickets/*/messages",
		"/api/v1/admin/coupons",
		"/api/v1/admin/coupons/*/toggle-status",
		"/api/v1/admin/coupons/*/extend",
		"/api/v1/admin/coupons/*/usage",
		"/api/v1/admin/coupons/search",
		"/api/v1/admin/coupons/statistics",
		"/api/v1/admin/coupons/analytics",
		"/api/v1/admin/coupons/bulk/*",
		"/api/v1/coupons",
		"/api/v1/coupons/validate",
		"/api/v1/coupons/my-usage",
		"/api/v1/admin/referrals",
		"/api/v1/admin/referrals/*/approve",
		"/api/v1/admin/referrals/*/reject",
		"/api/v1/admin/referrals/*/payout",
		"/api/v1/admin/referrals/search",
		"/api/v1/admin/referrals/statistics",
		"/api/v1/admin/referrals/analytics",
		"/api/v1/admin/referrals/campaigns",
		"/api/v1/admin/referrals/invite-codes",
		"/api/v1/admin/referrals/bulk/*",
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

	// Collect all verified important routes for consolidated logging
	verifiedRoutes := make([]string, 0)
	for _, route := range routes {
		for _, important := range importantRoutes {
			if route.Path == important ||
				(important[len(important)-1] == '*' &&
					len(route.Path) >= len(important)-1 &&
					route.Path[:len(important)-1] == important[:len(important)-1]) {
				verifiedRoutes = append(verifiedRoutes, route.Method+" "+route.Path)
				break
			}
		}
	}
	
	// Log consolidated verification result at debug level to reduce noise
	logger.Debug("Important routes verification completed",
		zap.Int("verified_count", len(verifiedRoutes)),
		zap.Int("expected_count", len(importantRoutes)),
		zap.Strings("verified_routes", verifiedRoutes))
}
