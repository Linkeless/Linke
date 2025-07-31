package modules

import (
	"linke/config"
	"linke/internal/repository"
	"linke/internal/service"
	"linke/internal/handler"
	"linke/internal/handler/admin"
	adminuser "linke/internal/handler/admin/user"
	admincoupon "linke/internal/handler/admin/coupon"
	admininvoice "linke/internal/handler/admin/invoice"
	adminticket "linke/internal/handler/admin/ticket"
	// adminorder "linke/internal/handler/admin/order" // DEPRECATED - handler not created yet
	"linke/internal/handler/user"
	usercoupon "linke/internal/handler/user/coupon"
	userinvitecode "linke/internal/handler/user/invite_code"
	userprofile "linke/internal/handler/user/profile"
	userticket "linke/internal/handler/user/ticket"
	"linke/internal/queue"
)

// SimpleManager manages all application modules using existing services
type SimpleManager struct {
	// New Business Flow Services (Order → Invoice → Payment → Subscription)
	OrderService               *service.OrderService
	NewInvoiceService          *service.InvoiceService  // New business flow invoice service
	NewPaymentService          *service.PaymentService  // New business flow payment service  
	NewSubscriptionService     *service.SubscriptionService // New business flow subscription service
	
	// Existing Services
	UserService                *service.UserService
	AuthService                *service.AuthService
	JWTService                 *service.JWTService
	InviteCodeService          *service.InviteCodeService
	InviteCodeUsageService     *service.InviteCodeUsageService
	SubscriptionPlanService    *service.SubscriptionPlanService
	UserSubscriptionService    *service.UserSubscriptionService
	PaymentService             *service.PaymentService
	PaymentConfigService       *service.PaymentConfigService
	// SubscriptionOrderService   *service.SubscriptionOrderService // DEPRECATED - removed in new business flow
	TicketService              *service.TicketService
	TicketMessageService       *service.TicketMessageService
	ReferralService            *service.ReferralService
	ReferralCampaignService    *service.ReferralCampaignService
	ServerGroupService         *service.ServerGroupService
	ShadowsocksServerService   *service.ShadowsocksServerService
	CouponService              *service.CouponService
	SubscriptionExpiryService  *service.SubscriptionExpiryService
	InvoiceService             *service.InvoiceService
	
	// Handlers
	AuthHandler                *handler.AuthHandler
	TaskHandler                *handler.TaskHandler
	PaymentHandler             *handler.PaymentHandler
	// SubscriptionOrderHandler   *handler.SubscriptionOrderHandler // DEPRECATED - removed in new business flow
	ServerAPIHandler           *handler.ServerAPIHandler
	ShadowsocksServerHandler   *handler.ShadowsocksServerHandler
	
	// Admin handlers
	AdminUserHandler           *adminuser.AdminUserManager
	AdminSubscriptionHandler   *admin.AdminSubscriptionHandler
	AdminTicketHandler         *adminticket.AdminTicketManager
	AdminReferralHandler       *admin.ReferralHandler
	AdminServerGroupHandler    *admin.ServerGroupHandler
	// AdminOrderHandler          *adminorder.AdminOrderManager // DEPRECATED - handler not created yet
	AdminInvoiceHandler        *admininvoice.AdminInvoiceManager
	AdminCouponHandler         *admincoupon.AdminCouponManager
	
	// User handlers
	UserProfileHandler         *userprofile.UserProfileManager
	UserSubscriptionHandler    *user.UserSubscriptionPublicHandler
	UserTicketHandler          *userticket.UserTicketManager
	UserInviteCodeHandler      *userinvitecode.UserInviteCodeManager
	UserReferralHandler        *user.ReferralHandler
	UserCouponHandler          *usercoupon.UserCouponManager
}

// NewSimpleManager creates a new simple module manager using existing services
func NewSimpleManager(cfg *config.Config, db *repository.Database, taskQueue *queue.TaskQueue) *SimpleManager {
	// Initialize security services first
	jwtBlacklistService := service.NewJWTBlacklistService(db.DB)
	loginSecurityService := service.NewLoginSecurityService(db.DB, nil) // Use default config
	
	// Initialize existing services first (dependencies)
	userService := service.NewUserService(db.DB)
	jwtService := service.NewJWTService(cfg, jwtBlacklistService)
	
	// Referral services
	referralService := service.NewReferralService(db)
	referralCampaignService := service.NewReferralCampaignService(db)
	
	inviteCodeService := service.NewInviteCodeService(db.DB, referralService)
	inviteCodeUsageService := service.NewInviteCodeUsageService(db.DB)
	subscriptionPlanService := service.NewSubscriptionPlanService(db.DB)
	userSubscriptionService := service.NewUserSubscriptionService(db.DB, subscriptionPlanService)
	couponService := service.NewCouponService(db.DB)
	subscriptionExpiryService := service.NewSubscriptionExpiryService(db.DB, userSubscriptionService)
	// subscriptionOrderService := service.NewSubscriptionOrderService(db.DB, subscriptionPlanService, userSubscriptionService, paymentService, couponService) // DEPRECATED
	
	// Initialize new business flow services (after dependencies)
	orderService := service.NewOrderService(db.DB, subscriptionPlanService, couponService)
	newInvoiceService := service.NewInvoiceService(db.DB, userService)
	newPaymentService := service.NewPaymentService(db.DB, newInvoiceService)
	newSubscriptionService := service.NewSubscriptionService(db.DB, orderService)
	
	authService := service.NewAuthService(db.DB, userService, jwtService, inviteCodeService, referralService, loginSecurityService)
	
	// Payment services
	paymentService := service.NewPaymentService(db.DB, newInvoiceService)
	paymentConfigService := service.NewPaymentConfigService(db.DB)
	invoiceService := service.NewInvoiceService(db.DB, userService)
	
	// Ticket services
	ticketService := service.NewTicketService(db.DB)
	ticketMessageService := service.NewTicketMessageService(db.DB)
	
	// Server services
	serverGroupService := service.NewServerGroupService(db)
	shadowsocksServerService := service.NewShadowsocksServerService(db)
	
	// Set up payment service dependencies
	// paymentService.SetSubscriptionOrderService(subscriptionOrderService) // DEPRECATED
	
	// Initialize handlers
	authHandler := handler.NewAuthHandler(cfg, db, authService, jwtService)
	taskHandler := handler.NewTaskHandler(taskQueue)
	paymentHandler := handler.NewPaymentHandler(paymentService, paymentConfigService)
	// subscriptionOrderHandler := handler.NewSubscriptionOrderHandler(subscriptionOrderService) // DEPRECATED
	serverAPIHandler := handler.NewServerAPIHandler(shadowsocksServerService, userService, userSubscriptionService, db, taskQueue.GetClient(), cfg)
	shadowsocksServerHandler := handler.NewShadowsocksServerHandler(shadowsocksServerService, nil, userService, userSubscriptionService)
	
	// Admin handlers
	adminUserHandler := adminuser.NewAdminUserManager(userService, authService)
	adminSubscriptionHandler := admin.NewAdminSubscriptionHandler(subscriptionPlanService, userSubscriptionService, orderService, newInvoiceService, newPaymentService, newSubscriptionService) // Updated to use new business flow services
	adminTicketHandler := adminticket.NewAdminTicketManager(ticketService, ticketMessageService)
	adminReferralHandler := admin.NewReferralHandler(referralService, referralCampaignService)
	adminServerGroupHandler := admin.NewServerGroupHandler(serverGroupService)
	// adminOrderHandler := adminorder.NewAdminOrderManager(newInvoiceService, newPaymentService, userService) // DEPRECATED - handler not created yet
	adminInvoiceHandler := admininvoice.NewAdminInvoiceManager(newInvoiceService) // Updated to use new service
	adminCouponHandler := admincoupon.NewAdminCouponManager(couponService)
	
	// User handlers
	userProfileHandler := userprofile.NewUserProfileManager(userService)
	userSubscriptionHandler := user.NewUserSubscriptionPublicHandler(subscriptionPlanService, userSubscriptionService, orderService, newInvoiceService, newPaymentService, newSubscriptionService, couponService) // Updated to use new business flow services
	userTicketHandler := userticket.NewUserTicketManager(ticketService, ticketMessageService)
	userInviteCodeHandler := userinvitecode.NewUserInviteCodeManager(inviteCodeService, inviteCodeUsageService)
	userReferralHandler := user.NewReferralHandler(referralService, referralCampaignService)
	userCouponHandler := usercoupon.NewUserCouponManager(couponService)
	
	return &SimpleManager{
		// New Business Flow Services
		OrderService:               orderService,
		NewInvoiceService:          newInvoiceService,
		NewPaymentService:          newPaymentService,
		NewSubscriptionService:     newSubscriptionService,
		
		// Existing Services
		UserService:                userService,
		AuthService:                authService,
		JWTService:                 jwtService,
		InviteCodeService:          inviteCodeService,
		InviteCodeUsageService:     inviteCodeUsageService,
		SubscriptionPlanService:    subscriptionPlanService,
		UserSubscriptionService:    userSubscriptionService,
		PaymentService:             paymentService,
		PaymentConfigService:       paymentConfigService,
		// SubscriptionOrderService:   subscriptionOrderService, // DEPRECATED
		TicketService:              ticketService,
		TicketMessageService:       ticketMessageService,
		ReferralService:            referralService,
		ReferralCampaignService:    referralCampaignService,
		ServerGroupService:         serverGroupService,
		ShadowsocksServerService:   shadowsocksServerService,
		CouponService:              couponService,
		SubscriptionExpiryService:  subscriptionExpiryService,
		InvoiceService:             invoiceService,
		
		// Handlers
		AuthHandler:                authHandler,
		TaskHandler:                taskHandler,
		PaymentHandler:             paymentHandler,
		// SubscriptionOrderHandler:   subscriptionOrderHandler, // DEPRECATED
		ServerAPIHandler:           serverAPIHandler,
		ShadowsocksServerHandler:   shadowsocksServerHandler,
		
		// Admin handlers
		AdminUserHandler:           adminUserHandler,
		AdminSubscriptionHandler:   adminSubscriptionHandler,
		AdminTicketHandler:         adminTicketHandler,
		AdminReferralHandler:       adminReferralHandler,
		AdminServerGroupHandler:    adminServerGroupHandler,
		// AdminOrderHandler:          adminOrderHandler, // DEPRECATED - handler not created yet
		AdminInvoiceHandler:        adminInvoiceHandler,
		AdminCouponHandler:         adminCouponHandler,
		
		// User handlers
		UserProfileHandler:         userProfileHandler,
		UserSubscriptionHandler:    userSubscriptionHandler,
		UserTicketHandler:          userTicketHandler,
		UserInviteCodeHandler:      userInviteCodeHandler,
		UserReferralHandler:        userReferralHandler,
		UserCouponHandler:          userCouponHandler,
	}
}