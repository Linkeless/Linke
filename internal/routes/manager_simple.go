package routes

import (
	"time"
	"linke/internal/response"
	"linke/internal/middleware"
	"linke/internal/modules"
	"github.com/gin-gonic/gin"
)

// SimpleManager handles all route registration using existing handlers
type SimpleManager struct {
	router  *gin.Engine
	modules *modules.SimpleManager
}

// NewSimpleManager creates a new simple route manager
func NewSimpleManager(modules *modules.SimpleManager) *SimpleManager {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	
	// Global middleware (order matters!)
	router.Use(middleware.RequestContextHandler())    // Add request context first
	router.Use(middleware.Logger())                   // Logging
	router.Use(middleware.CORS())                     // CORS handling
	router.Use(middleware.ErrorHandler())             // Panic recovery
	router.Use(middleware.DatabaseErrorHandler())     // Database error handling
	router.Use(middleware.ValidationErrorHandler())   // Validation error handling
	router.Use(middleware.SecurityErrorHandler())     // Security error handling
	router.Use(gin.Recovery())                        // Default recovery as fallback
	
	return &SimpleManager{
		router:  router,
		modules: modules,
	}
}

// SetupRoutes sets up all application routes
func (m *SimpleManager) SetupRoutes() {
	// Health check route
	m.router.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})
	
	// API version 1 routes
	v1 := m.router.Group("/api/v1")
	{
		// Basic ping endpoint
		v1.GET("/ping", func(c *gin.Context) {
			response.SuccessWithMessage(c, "pong", nil)
		})
		
		// Setup route groups
		m.setupPublicRoutes(v1)
		m.setupUserRoutes(v1)
		m.setupAdminRoutes(v1)
		m.setupServerAPIRoutes(v1)
	}
}

// setupPublicRoutes sets up public routes (no authentication required)
func (m *SimpleManager) setupPublicRoutes(v1 *gin.RouterGroup) {
	// Authentication routes
	auth := v1.Group("/auth")
	{
		// OAuth routes (legacy redirect-based)
		auth.GET("/providers", m.modules.AuthHandler.GetProviders)
		auth.GET("/telegram/widget", m.modules.AuthHandler.GetTelegramWidget)
		auth.GET("/:provider", m.modules.AuthHandler.Login)
		auth.GET("/:provider/callback", m.modules.AuthHandler.Callback)
		
		// OAuth API routes (modern JSON-based)
		auth.POST("/url", m.modules.AuthHandler.GetAuthURL)     // Get authorization URL
		auth.POST("/token", m.modules.AuthHandler.ExchangeToken) // Exchange code for token
		
		// Local authentication routes (most sensitive)
		auth.POST("/register", m.modules.AuthHandler.Register)
		auth.POST("/login", m.modules.AuthHandler.LoginLocal)
		auth.POST("/refresh", m.modules.AuthHandler.RefreshToken)
	}
	
	// Public invite code routes
	inviteCodes := v1.Group("/invite-codes")
	{
		inviteCodes.GET("/validate/:code", m.modules.UserInviteCodeHandler.ValidateInviteCode)
	}
	
	// Public subscription plan routes
	subscriptionPlans := v1.Group("/subscription-plans")
	{
		subscriptionPlans.GET("", m.modules.UserSubscriptionHandler.GetPublicSubscriptionPlans)
		subscriptionPlans.GET("/:id", m.modules.UserSubscriptionHandler.GetSubscriptionPlan)
		subscriptionPlans.GET("/code/:code", m.modules.UserSubscriptionHandler.GetSubscriptionPlanByCode)
	}
	
	// Public payment routes with strict rate limiting
	payments := v1.Group("/payments")
	payments.Use(middleware.PaymentErrorHandler()) // Add payment-specific error handling
	{
		payments.GET("/methods", m.modules.PaymentHandler.GetAvailablePaymentMethods)
		payments.GET("/configs", m.modules.PaymentHandler.GetActivePaymentConfigs)
		// Payment notifications
		payments.POST("/notify/:gateway", m.modules.PaymentHandler.PaymentNotify)
	}
	
	// Public server status routes
	v1.GET("/servers/status", func(c *gin.Context) {
		response.SuccessWithMessage(c, "Server status temporarily unavailable", map[string]interface{}{
			"status": "unavailable",
			"message": "Server management module has been removed",
		})
	})
	
	// Public referral routes
	v1.GET("/referral-campaigns", m.modules.UserReferralHandler.GetPublicReferralCampaigns)
	referral := v1.Group("/referral")
	{
		referral.POST("/track/:code", m.modules.UserReferralHandler.TrackReferralClick)
	}
	
	// Public coupon routes removed for security reasons
	// Coupon codes should be distributed through targeted marketing channels
	// such as email campaigns, SMS, or promotional materials
}

// setupUserRoutes sets up user routes (authentication required)
func (m *SimpleManager) setupUserRoutes(v1 *gin.RouterGroup) {
	// User route group with authentication middleware
	user := v1.Group("/user")
	user.Use(middleware.AuthMiddleware(m.modules.AuthService))
	
	// User authentication routes
	auth := user.Group("/auth")
	{
		auth.POST("/logout", m.modules.AuthHandler.Logout)
		auth.POST("/change-password", m.modules.AuthHandler.ChangePassword)
		auth.GET("/profile", m.modules.AuthHandler.GetProfile)
	}
	
	// User profile routes
	{
		user.GET("/profile", m.modules.UserProfileHandler.GetProfile)
		user.PUT("/profile", m.modules.UserProfileHandler.UpdateProfile)
		user.PUT("/password", m.modules.UserProfileHandler.ChangePassword)
	}
	
	// User subscription routes
	subscriptions := user.Group("/subscriptions")
	{
		subscriptions.GET("", m.modules.UserSubscriptionHandler.GetMySubscriptions)
		subscriptions.GET("/:id", m.modules.UserSubscriptionHandler.GetMySubscription)
		subscriptions.GET("/history", m.modules.UserSubscriptionHandler.GetMySubscriptionHistory)
		subscriptions.POST("/purchase", m.modules.UserSubscriptionHandler.PurchaseSubscription)
		subscriptions.POST("/:id/cancel", m.modules.UserSubscriptionHandler.CancelMySubscription)
	}
	
	// User shadowsocks server routes
	shadowsocksServers := user.Group("/shadowsocks-servers")
	{
		shadowsocksServers.GET("", m.modules.ShadowsocksServerHandler.GetAvailableShadowsocksServers)
	}
	
	// User payment routes
	payments := user.Group("/payments")
	payments.Use(middleware.PaymentErrorHandler()) // Add payment-specific error handling
	{
		payments.POST("/orders", m.modules.PaymentHandler.CreatePaymentOrder)
		payments.GET("/orders/my", m.modules.PaymentHandler.GetMyPaymentOrders)
		payments.GET("/orders/:payment_no", m.modules.PaymentHandler.GetPaymentOrder)
	}
	
	// User referral routes
	referrals := user.Group("/referrals")
	{
		referrals.GET("", m.modules.UserReferralHandler.GetMyReferrals)
		referrals.GET("/stats", m.modules.UserReferralHandler.GetMyReferralStats)
	}
	
	// User invite code routes
	inviteCodes := user.Group("/invite-codes")
	{
		inviteCodes.POST("", m.modules.UserInviteCodeHandler.CreateInviteCode)
		inviteCodes.GET("/my", m.modules.UserInviteCodeHandler.GetMyInviteCodes)
		inviteCodes.GET("/:id", m.modules.UserInviteCodeHandler.GetInviteCode)
		inviteCodes.GET("/:id/usages", m.modules.UserInviteCodeHandler.GetInviteCodeUsages)
		inviteCodes.PUT("/:id/status", m.modules.UserInviteCodeHandler.UpdateInviteCodeStatus)
		inviteCodes.DELETE("/:id", m.modules.UserInviteCodeHandler.DeleteInviteCode)
	}
	
	// Task routes
	tasks := user.Group("/tasks")
	{
		tasks.POST("", m.modules.TaskHandler.CreateTask)
		tasks.GET("/status", m.modules.TaskHandler.GetQueueStatus)
	}
	
	// User coupon routes (authenticated)
	coupons := user.Group("/coupons")
	{
		coupons.POST("/validate", m.modules.UserCouponHandler.ValidateCoupon)
		coupons.GET("/usages", m.modules.UserCouponHandler.GetMyCouponUsages)
	}
	
	// Setup user ticket routes
	m.modules.UserTicketHandler.SetupUserTicketRoutes(user)
}

// setupAdminRoutes sets up admin routes (authentication + admin role required)
func (m *SimpleManager) setupAdminRoutes(v1 *gin.RouterGroup) {
	// Admin route group with authentication and admin middleware
	admin := v1.Group("/admin")
	admin.Use(middleware.AuthMiddleware(m.modules.AuthService))
	admin.Use(middleware.RequireAdmin())
	
	// Admin user management routes
	adminUsers := admin.Group("/users")
	{
		adminUsers.POST("", m.modules.AdminUserHandler.CreateUser)
		adminUsers.GET("", m.modules.AdminUserHandler.ListUsers)
		adminUsers.GET("/deleted", m.modules.AdminUserHandler.ListDeletedUsers)
		adminUsers.GET("/search", m.modules.AdminUserHandler.SearchUsers)
		adminUsers.GET("/stats", m.modules.AdminUserHandler.GetUserStats)
		adminUsers.GET("/provider", m.modules.AdminUserHandler.ListUsersByProvider)
		adminUsers.GET("/:id", m.modules.AdminUserHandler.GetUser)
		adminUsers.PUT("/:id", m.modules.AdminUserHandler.UpdateUser)
		adminUsers.PATCH("/:id", m.modules.AdminUserHandler.PatchUser)
		adminUsers.PUT("/:id/role", m.modules.AdminUserHandler.UpdateUserRole)
		adminUsers.PUT("/:id/status", m.modules.AdminUserHandler.UpdateUserStatus)
		adminUsers.POST("/:id/reset-password", m.modules.AdminUserHandler.ResetUserPassword)
		adminUsers.DELETE("/:id", m.modules.AdminUserHandler.SoftDeleteUser)
		adminUsers.POST("/:id/restore", m.modules.AdminUserHandler.RestoreUser)
		adminUsers.DELETE("/:id/hard-delete", m.modules.AdminUserHandler.HardDeleteUser)
		adminUsers.POST("/batch/delete", m.modules.AdminUserHandler.BatchDeleteUsers)
		adminUsers.POST("/batch/restore", m.modules.AdminUserHandler.BatchRestoreUsers)
	}
	
	// Admin invite code management routes
	adminInviteCodes := admin.Group("/invite-codes")
	{
		adminInviteCodes.GET("", m.modules.UserInviteCodeHandler.ListAllInviteCodes)
		adminInviteCodes.GET("/stats", m.modules.UserInviteCodeHandler.GetInviteCodeStats)
	}
	
	// Admin referral management routes
	adminReferrals := admin.Group("/referrals")
	{
		adminReferrals.GET("", m.modules.AdminReferralHandler.ListAllReferrals)
		adminReferrals.GET("/:id", m.modules.AdminReferralHandler.GetReferral)
	}
	
	// Admin referral campaign management routes
	adminReferralCampaigns := admin.Group("/referral-campaigns")
	{
		adminReferralCampaigns.POST("", m.modules.AdminReferralHandler.CreateReferralCampaign)
		adminReferralCampaigns.GET("", m.modules.AdminReferralHandler.ListReferralCampaigns)
		adminReferralCampaigns.GET("/:id", m.modules.AdminReferralHandler.GetReferralCampaign)
		adminReferralCampaigns.PUT("/:id", m.modules.AdminReferralHandler.UpdateReferralCampaign)
		adminReferralCampaigns.DELETE("/:id", m.modules.AdminReferralHandler.DeleteReferralCampaign)
		adminReferralCampaigns.GET("/:id/stats", m.modules.AdminReferralHandler.GetReferralCampaignStats)
	}
	
	// Admin server group management routes
	adminServerGroups := admin.Group("/server-groups")
	{
		adminServerGroups.POST("", m.modules.AdminServerGroupHandler.CreateServerGroup)
		adminServerGroups.GET("", m.modules.AdminServerGroupHandler.ListServerGroups)
		adminServerGroups.GET("/all", m.modules.AdminServerGroupHandler.GetAllServerGroups)
		adminServerGroups.GET("/:id", m.modules.AdminServerGroupHandler.GetServerGroup)
		adminServerGroups.PUT("/:id", m.modules.AdminServerGroupHandler.UpdateServerGroup)
		adminServerGroups.DELETE("/:id", m.modules.AdminServerGroupHandler.DeleteServerGroup)
	}

	// Admin shadowsocks server management routes
	adminShadowsocksServers := admin.Group("/shadowsocks-servers")
	{
		adminShadowsocksServers.POST("", m.modules.ShadowsocksServerHandler.CreateShadowsocksServer)
		adminShadowsocksServers.GET("", m.modules.ShadowsocksServerHandler.GetShadowsocksServers)
		adminShadowsocksServers.GET("/:id", m.modules.ShadowsocksServerHandler.GetShadowsocksServerByID)
		adminShadowsocksServers.PUT("/:id", m.modules.ShadowsocksServerHandler.UpdateShadowsocksServer)
		adminShadowsocksServers.PATCH("/:id", m.modules.ShadowsocksServerHandler.PatchShadowsocksServer)
		adminShadowsocksServers.DELETE("/:id", m.modules.ShadowsocksServerHandler.DeleteShadowsocksServer)
	}

	
	// Admin payment config management routes
	adminPayments := admin.Group("/payments")
	{
		adminPayments.POST("/configs", m.modules.PaymentHandler.CreatePaymentConfig)
		adminPayments.GET("/configs", m.modules.PaymentHandler.GetPaymentConfigs)
		adminPayments.PUT("/configs/:id", m.modules.PaymentHandler.UpdatePaymentConfig)
		adminPayments.DELETE("/configs/:id", m.modules.PaymentHandler.DeletePaymentConfig)
	}
	
	// Admin subscription management routes
	adminSubscriptions := admin.Group("/subscriptions")
	{
		// Subscription plan management
		adminSubscriptions.POST("/plans", m.modules.AdminSubscriptionHandler.CreateSubscriptionPlan)
		adminSubscriptions.GET("/plans", m.modules.AdminSubscriptionHandler.ListSubscriptionPlans)
		adminSubscriptions.GET("/plans/:id", m.modules.AdminSubscriptionHandler.GetSubscriptionPlan)
		adminSubscriptions.PUT("/plans/:id", m.modules.AdminSubscriptionHandler.UpdateSubscriptionPlan)
		adminSubscriptions.PATCH("/plans/:id", m.modules.AdminSubscriptionHandler.PatchSubscriptionPlan)
		adminSubscriptions.DELETE("/plans/:id", m.modules.AdminSubscriptionHandler.DeleteSubscriptionPlan)
		
		// User subscription management
		adminSubscriptions.POST("/users", m.modules.AdminSubscriptionHandler.CreateUserSubscription)
		adminSubscriptions.GET("/users", m.modules.AdminSubscriptionHandler.ListUserSubscriptions)
		adminSubscriptions.GET("/users/:id", m.modules.AdminSubscriptionHandler.GetUserSubscription)
		adminSubscriptions.PUT("/users/:id", m.modules.AdminSubscriptionHandler.UpdateUserSubscription)
		adminSubscriptions.PATCH("/users/:id", m.modules.AdminSubscriptionHandler.PatchUserSubscription)
		adminSubscriptions.POST("/users/:id/renew", m.modules.AdminSubscriptionHandler.RenewUserSubscription)
		adminSubscriptions.DELETE("/users/:id", m.modules.AdminSubscriptionHandler.DeleteUserSubscription)
	}
	
	// Admin order management routes
	adminOrders := admin.Group("/orders")
	{
		adminOrders.GET("", m.modules.AdminOrderHandler.ListOrders)
		adminOrders.GET("/:id", m.modules.AdminOrderHandler.GetOrder)
		adminOrders.PATCH("/:id/status", m.modules.AdminOrderHandler.UpdateOrderStatus)
		adminOrders.POST("/:id/refund", m.modules.AdminOrderHandler.ProcessRefund)
		adminOrders.GET("/stats", m.modules.AdminOrderHandler.GetOrderStats)
		adminOrders.POST("/bulk", m.modules.AdminOrderHandler.BulkUpdate)
	}
	
	// Admin invoice management routes
	adminInvoices := admin.Group("/invoices")
	{
		adminInvoices.POST("", m.modules.AdminInvoiceHandler.CreateInvoice)
		adminInvoices.POST("/from-order", m.modules.AdminInvoiceHandler.CreateInvoiceFromOrder)
		adminInvoices.GET("", m.modules.AdminInvoiceHandler.ListInvoices)
		adminInvoices.GET("/:id", m.modules.AdminInvoiceHandler.GetInvoice)
		adminInvoices.PUT("/:id", m.modules.AdminInvoiceHandler.UpdateInvoice)
		adminInvoices.POST("/:id/send", m.modules.AdminInvoiceHandler.SendInvoice)
		adminInvoices.POST("/:id/mark-paid", m.modules.AdminInvoiceHandler.MarkInvoiceAsPaid)
		adminInvoices.POST("/:id/void", m.modules.AdminInvoiceHandler.VoidInvoice)
		adminInvoices.DELETE("/:id", m.modules.AdminInvoiceHandler.DeleteInvoice)
	}
	
	// Admin coupon management routes
	adminCoupons := admin.Group("/coupons")
	{
		adminCoupons.POST("", m.modules.AdminCouponHandler.CreateCoupon)
		adminCoupons.GET("", m.modules.AdminCouponHandler.GetCoupons)
		adminCoupons.GET("/:id", m.modules.AdminCouponHandler.GetCoupon)
		adminCoupons.PUT("/:id", m.modules.AdminCouponHandler.UpdateCoupon)
		adminCoupons.DELETE("/:id", m.modules.AdminCouponHandler.DeleteCoupon)
		adminCoupons.GET("/:id/usages", m.modules.AdminCouponHandler.GetCouponUsages)
		adminCoupons.GET("/code/:code", m.modules.AdminCouponHandler.GetCouponByCode)
	}
	
	// Set up admin ticket routes
	m.modules.AdminTicketHandler.SetupAdminTicketRoutes(admin)
}

// setupServerAPIRoutes sets up server node API routes
func (m *SimpleManager) setupServerAPIRoutes(v1 *gin.RouterGroup) {
	// Server API routes
	serverAPI := v1.Group("/server")
	
	// UniProxy Server API - Support both original and V2 formats
	uniProxy := serverAPI.Group("/UniProxy")
	{
		// Health check endpoint
		uniProxy.GET("/health", m.modules.ServerAPIHandler.Health)
		// Config endpoint
		uniProxy.GET("/config", m.modules.ServerAPIHandler.UniProxyConfig)
		// Users endpoint
		uniProxy.GET("/user", m.modules.ServerAPIHandler.UniProxyUsers)
		// Push endpoint for node data
		uniProxy.POST("/push", m.modules.ServerAPIHandler.UniProxyPush)
	}
	
	// Subscription order routes (authenticated)
	subscriptionOrders := v1.Group("/subscription-orders")
	subscriptionOrders.Use(middleware.AuthMiddleware(m.modules.AuthService))
	{
		subscriptionOrders.POST("", m.modules.SubscriptionOrderHandler.CreateSubscriptionOrder)
		subscriptionOrders.GET("/my", m.modules.SubscriptionOrderHandler.GetMySubscriptionOrders)
		subscriptionOrders.GET("/:id", m.modules.SubscriptionOrderHandler.GetSubscriptionOrder)
	}
}

// GetRouter returns the gin router instance
func (m *SimpleManager) GetRouter() *gin.Engine {
	return m.router
}