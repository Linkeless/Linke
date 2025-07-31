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
	adminorder "linke/internal/handler/admin/order"
	"linke/internal/handler/user"
	usercoupon "linke/internal/handler/user/coupon"
	userinvitecode "linke/internal/handler/user/invite_code"
	userprofile "linke/internal/handler/user/profile"
	userticket "linke/internal/handler/user/ticket"
	"linke/internal/queue"
)

// SimpleManager manages all application modules using existing services
type SimpleManager struct {
	// Services
	UserService                *service.UserService
	AuthService                *service.AuthService
	JWTService                 *service.JWTService
	InviteCodeService          *service.InviteCodeService
	InviteCodeUsageService     *service.InviteCodeUsageService
	SubscriptionPlanService    *service.SubscriptionPlanService
	UserSubscriptionService    *service.UserSubscriptionService
	PaymentService             *service.PaymentService
	PaymentConfigService       *service.PaymentConfigService
	SubscriptionOrderService   *service.SubscriptionOrderService
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
	SubscriptionOrderHandler   *handler.SubscriptionOrderHandler
	ServerAPIHandler           *handler.ServerAPIHandler
	ShadowsocksServerHandler   *handler.ShadowsocksServerHandler
	
	// Admin handlers
	AdminUserHandler           *adminuser.AdminUserManager
	AdminSubscriptionHandler   *admin.AdminSubscriptionHandler
	AdminTicketHandler         *adminticket.AdminTicketManager
	AdminReferralHandler       *admin.ReferralHandler
	AdminServerGroupHandler    *admin.ServerGroupHandler
	AdminOrderHandler          *adminorder.AdminOrderManager
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
	
	// Initialize core services
	userService := service.NewUserService(db.DB)
	jwtService := service.NewJWTService(cfg, jwtBlacklistService)
	
	// Referral services
	referralService := service.NewReferralService(db)
	referralCampaignService := service.NewReferralCampaignService(db)
	
	inviteCodeService := service.NewInviteCodeService(db.DB, referralService)
	inviteCodeUsageService := service.NewInviteCodeUsageService(db.DB)
	subscriptionPlanService := service.NewSubscriptionPlanService(db.DB)
	userSubscriptionService := service.NewUserSubscriptionService(db.DB, subscriptionPlanService)
	paymentService := service.NewPaymentService(db.DB)
	paymentConfigService := service.NewPaymentConfigService(db.DB)
	couponService := service.NewCouponService(db.DB)
	subscriptionExpiryService := service.NewSubscriptionExpiryService(db.DB, userSubscriptionService)
	subscriptionOrderService := service.NewSubscriptionOrderService(db.DB, subscriptionPlanService, userSubscriptionService, paymentService, couponService)
	invoiceService := service.NewInvoiceService(db.DB, userService)
	
	authService := service.NewAuthService(db.DB, userService, jwtService, inviteCodeService, referralService, loginSecurityService)
	
	// Ticket services
	ticketService := service.NewTicketService(db.DB)
	ticketMessageService := service.NewTicketMessageService(db.DB)
	
	// Server services
	serverGroupService := service.NewServerGroupService(db)
	shadowsocksServerService := service.NewShadowsocksServerService(db)
	
	// Set up payment service dependencies
	paymentService.SetSubscriptionOrderService(subscriptionOrderService)
	
	// Initialize handlers
	authHandler := handler.NewAuthHandler(cfg, db, authService, jwtService)
	taskHandler := handler.NewTaskHandler(taskQueue)
	paymentHandler := handler.NewPaymentHandler(paymentService, paymentConfigService)
	subscriptionOrderHandler := handler.NewSubscriptionOrderHandler(subscriptionOrderService)
	serverAPIHandler := handler.NewServerAPIHandler(shadowsocksServerService, userService, userSubscriptionService, db, taskQueue.GetClient(), cfg)
	shadowsocksServerHandler := handler.NewShadowsocksServerHandler(shadowsocksServerService, nil, userService, userSubscriptionService)
	
	// Admin handlers
	adminUserHandler := adminuser.NewAdminUserManager(userService, authService)
	adminSubscriptionHandler := admin.NewAdminSubscriptionHandler(subscriptionPlanService, userSubscriptionService, subscriptionOrderService)
	adminTicketHandler := adminticket.NewAdminTicketManager(ticketService, ticketMessageService)
	adminReferralHandler := admin.NewReferralHandler(referralService, referralCampaignService)
	adminServerGroupHandler := admin.NewServerGroupHandler(serverGroupService)
	adminOrderHandler := adminorder.NewAdminOrderManager(subscriptionOrderService, paymentService, userService)
	adminInvoiceHandler := admininvoice.NewAdminInvoiceManager(invoiceService)
	adminCouponHandler := admincoupon.NewAdminCouponManager(couponService)
	
	// User handlers
	userProfileHandler := userprofile.NewUserProfileManager(userService)
	userSubscriptionHandler := user.NewUserSubscriptionPublicHandler(subscriptionPlanService, userSubscriptionService, subscriptionOrderService, couponService, subscriptionExpiryService)
	userTicketHandler := userticket.NewUserTicketManager(ticketService, ticketMessageService)
	userInviteCodeHandler := userinvitecode.NewUserInviteCodeManager(inviteCodeService, inviteCodeUsageService)
	userReferralHandler := user.NewReferralHandler(referralService, referralCampaignService)
	userCouponHandler := usercoupon.NewUserCouponManager(couponService)
	
	return &SimpleManager{
		// Services
		UserService:                userService,
		AuthService:                authService,
		JWTService:                 jwtService,
		InviteCodeService:          inviteCodeService,
		InviteCodeUsageService:     inviteCodeUsageService,
		SubscriptionPlanService:    subscriptionPlanService,
		UserSubscriptionService:    userSubscriptionService,
		PaymentService:             paymentService,
		PaymentConfigService:       paymentConfigService,
		SubscriptionOrderService:   subscriptionOrderService,
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
		SubscriptionOrderHandler:   subscriptionOrderHandler,
		ServerAPIHandler:           serverAPIHandler,
		ShadowsocksServerHandler:   shadowsocksServerHandler,
		
		// Admin handlers
		AdminUserHandler:           adminUserHandler,
		AdminSubscriptionHandler:   adminSubscriptionHandler,
		AdminTicketHandler:         adminTicketHandler,
		AdminReferralHandler:       adminReferralHandler,
		AdminServerGroupHandler:    adminServerGroupHandler,
		AdminOrderHandler:          adminOrderHandler,
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