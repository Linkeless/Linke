package services

import (
	"context"

	"linke/internal/shared/logger"
	"linke/internal/shared/database"
	"linke/internal/domains/user"
	"linke/internal/domains/auth"
	"linke/internal/domains/subscription"
	"linke/internal/domains/payment"
	"linke/internal/domains/coupon"
	"linke/internal/domains/invoice"
	"linke/internal/domains/ticket"
	"linke/internal/domains/server"
	"linke/internal/domains/referral"
)

// ApplicationService 应用层服务
// 作为所有领域模块的协调器，提供统一的应用程序接口
type ApplicationService struct {
	logger   logger.Logger
	database *database.Database
	
	// 领域服务提供者
	userService         *user.ServiceProvider
	authService         *auth.ServiceProvider
	subscriptionService *subscription.ServiceProvider
	paymentService      *payment.ServiceProvider
	couponService       *coupon.ServiceProvider
	invoiceService      *invoice.ServiceProvider
	ticketService       *ticket.ServiceProvider
	serverService       *server.ServiceProvider
	referralService     *referral.ServiceProvider
}

// NewApplicationService 创建应用层服务
func NewApplicationService(
	logger logger.Logger,
	database *database.Database,
	userService *user.ServiceProvider,
	authService *auth.ServiceProvider,
	subscriptionService *subscription.ServiceProvider,
	paymentService *payment.ServiceProvider,
	couponService *coupon.ServiceProvider,
	invoiceService *invoice.ServiceProvider,
	ticketService *ticket.ServiceProvider,
	serverService *server.ServiceProvider,
	referralService *referral.ServiceProvider,
) *ApplicationService {
	return &ApplicationService{
		logger:              logger,
		database:            database,
		userService:         userService,
		authService:         authService,
		subscriptionService: subscriptionService,
		paymentService:      paymentService,
		couponService:       couponService,
		invoiceService:      invoiceService,
		ticketService:       ticketService,
		serverService:       serverService,
		referralService:     referralService,
	}
}

// HealthCheck 系统健康检查
func (s *ApplicationService) HealthCheck(ctx context.Context) map[string]interface{} {
	result := make(map[string]interface{})
	
	// 数据库健康检查
	dbHealth := s.database.HealthCheck(ctx)
	result["database"] = dbHealth
	
	// 添加应用层状态
	result["application"] = map[string]interface{}{
		"status": "healthy",
		"modules": map[string]string{
			"user":         "active",
			"auth":         "active",
			"subscription": "active", 
			"payment":      "active",
			"coupon":       "active",
			"invoice":      "active",
			"ticket":       "active",
			"server":       "active",
			"referral":     "active",
		},
	}
	
	return result
}

// GetUserService 获取用户服务
func (s *ApplicationService) GetUserService() *user.ServiceProvider {
	return s.userService
}

// GetAuthService 获取认证服务
func (s *ApplicationService) GetAuthService() *auth.ServiceProvider {
	return s.authService
}

// GetSubscriptionService 获取订阅服务
func (s *ApplicationService) GetSubscriptionService() *subscription.ServiceProvider {
	return s.subscriptionService
}

// GetPaymentService 获取支付服务
func (s *ApplicationService) GetPaymentService() *payment.ServiceProvider {
	return s.paymentService
}

// GetCouponService 获取优惠券服务
func (s *ApplicationService) GetCouponService() *coupon.ServiceProvider {
	return s.couponService
}

// GetInvoiceService 获取发票服务
func (s *ApplicationService) GetInvoiceService() *invoice.ServiceProvider {
	return s.invoiceService
}

// GetTicketService 获取工单服务
func (s *ApplicationService) GetTicketService() *ticket.ServiceProvider {
	return s.ticketService
}

// GetServerService 获取服务器服务
func (s *ApplicationService) GetServerService() *server.ServiceProvider {
	return s.serverService
}

// GetReferralService 获取推荐服务
func (s *ApplicationService) GetReferralService() *referral.ServiceProvider {
	return s.referralService
}